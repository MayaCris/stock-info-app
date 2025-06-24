package scripts

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/MayaCris/stock-info-app/internal/infrastructure/config"
	"github.com/MayaCris/stock-info-app/internal/infrastructure/logger"
)

// TestResult representa el resultado de una prueba
type TestResult struct {
	TestName    string        `json:"test_name"`
	Endpoint    string        `json:"endpoint"`
	Method      string        `json:"method"`
	StatusCode  int           `json:"status_code"`
	Success     bool          `json:"success"`
	Duration    time.Duration `json:"duration"`
	Error       string        `json:"error,omitempty"`
	ResponseLen int           `json:"response_length"`
	Timestamp   time.Time     `json:"timestamp"`
}

// TestSuite representa una suite completa de pruebas
type TestSuite struct {
	Name        string        `json:"name"`
	Version     string        `json:"version"`
	StartTime   time.Time     `json:"start_time"`
	EndTime     time.Time     `json:"end_time"`
	Duration    time.Duration `json:"duration"`
	Tests       []TestResult  `json:"tests"`
	TotalTests  int           `json:"total_tests"`
	Passed      int           `json:"passed"`
	Failed      int           `json:"failed"`
	SuccessRate float64       `json:"success_rate"`
}

// EndpointTester maneja las pruebas de los endpoints REST
type EndpointTester struct {
	baseURL    string
	client     *http.Client
	logger     logger.Logger
	testLogger logger.Logger
	config     *config.Config
	suite      *TestSuite
}

// NewEndpointTester crea una nueva instancia del tester
func NewEndpointTester() (*EndpointTester, error) {
	// Cargar configuración
	cfg, err := config.Load()
	if err != nil {
		return nil, fmt.Errorf("failed to load config: %w", err)
	}

	// Inicializar logger general
	appLogger, err := logger.InitializeGlobalLogger()
	if err != nil {
		return nil, fmt.Errorf("failed to initialize logger: %w", err)
	}

	// Crear logger específico para tests
	testLogger, err := createTestLogger()
	if err != nil {
		return nil, fmt.Errorf("failed to create test logger: %w", err)
	}

	// Cliente HTTP con timeout
	client := &http.Client{
		Timeout: 30 * time.Second,
	}

	return &EndpointTester{
		baseURL:    "http://localhost:8080/api/v1",
		client:     client,
		logger:     appLogger,
		testLogger: testLogger,
		config:     cfg,
		suite: &TestSuite{
			Name:      "Finnhub API Integration Tests",
			Version:   cfg.App.Version,
			StartTime: time.Now(),
			Tests:     make([]TestResult, 0),
		},
	}, nil
}

// createTestLogger crea un logger específico para las pruebas
func createTestLogger() (logger.Logger, error) {
	// Asegurar que existe el directorio de logs
	logsDir := "logs"
	if err := os.MkdirAll(logsDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create logs directory: %w", err)
	}
	// Crear nombre de archivo con timestamp
	timestamp := time.Now().Format("20060102_150405")

	// Configuración específica para el logger de tests de Finnhub
	logConfig := &logger.LogConfig{
		Level:      logger.DebugLevel,
		Format:     "json",
		TimeFormat: time.RFC3339,

		// Archivo
		EnableFile:     true,
		LogDir:         logsDir,
		LogFileName:    fmt.Sprintf("finnhub_tests_%s.log", timestamp),
		MaxSize:        10, // 10MB
		MaxBackups:     5,
		MaxAge:         30,
		Compress:       true,
		EnableRotation: true,

		// Consola
		EnableConsole: false,
		ColorOutput:   false,
	}

	return logger.NewFileLogger(logConfig)
}

// checkServerHealth verifica si el servidor está funcionando
func (ft *EndpointTester) checkServerHealth() error {
	ctx := context.Background()
	ft.logger.Info(ctx, "Verificando estado del servidor...")

	resp, err := ft.client.Get(ft.baseURL[:strings.LastIndex(ft.baseURL, "/api")])
	if err != nil {
		return fmt.Errorf("servidor no responde: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("servidor retornó status %d", resp.StatusCode)
	}

	ft.logger.Info(ctx, "✅ Servidor respondiendo correctamente")
	return nil
}

// makeRequest realiza una petición HTTP y registra el resultado
func (ft *EndpointTester) makeRequest(testName, endpoint, method string) TestResult {
	ctx := context.Background()
	startTime := time.Now()

	result := TestResult{
		TestName:  testName,
		Endpoint:  endpoint,
		Method:    method,
		Timestamp: startTime,
	}

	ft.testLogger.Info(ctx, "Iniciando prueba",
		logger.String("test", testName),
		logger.String("endpoint", endpoint),
		logger.String("method", method),
	)

	// Realizar petición
	resp, err := ft.client.Get(endpoint)
	result.Duration = time.Since(startTime)

	if err != nil {
		result.Error = err.Error()
		result.Success = false

		ft.testLogger.Error(ctx, "Error en petición HTTP", err,
			logger.String("test", testName),
			logger.String("endpoint", endpoint),
			logger.Duration("duration", result.Duration),
		)
		return result
	}
	defer resp.Body.Close()

	result.StatusCode = resp.StatusCode

	// Leer respuesta
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		result.Error = fmt.Sprintf("Error leyendo respuesta: %v", err)
		result.Success = false

		ft.testLogger.Error(ctx, "Error leyendo respuesta", err,
			logger.String("test", testName),
		)
		return result
	}

	result.ResponseLen = len(body)

	// Verificar status code
	if resp.StatusCode >= 200 && resp.StatusCode < 300 {
		result.Success = true

		// Validar que la respuesta sea JSON válido
		var jsonData interface{}
		if err := json.Unmarshal(body, &jsonData); err != nil {
			result.Error = fmt.Sprintf("Respuesta no es JSON válido: %v", err)
			result.Success = false

			ft.testLogger.Error(ctx, "Respuesta no es JSON válido", err,
				logger.String("test", testName),
				logger.String("response_preview", string(body[:min(len(body), 200)])),
			)
		} else {
			ft.testLogger.Info(ctx, "✅ Prueba exitosa",
				logger.String("test", testName),
				logger.Int("status_code", result.StatusCode),
				logger.Duration("duration", result.Duration),
				logger.Int("response_size", result.ResponseLen),
			)
		}
	} else {
		result.Success = false
		result.Error = fmt.Sprintf("HTTP %d: %s", resp.StatusCode, string(body))

		ft.testLogger.Warn(ctx, "Prueba falló",
			logger.String("test", testName),
			logger.Int("status_code", result.StatusCode),
			logger.String("error", result.Error),
		)
	}

	return result
}

// addTest añade un resultado de prueba a la suite
func (ft *EndpointTester) addTest(result TestResult) {
	ft.suite.Tests = append(ft.suite.Tests, result)
	ft.suite.TotalTests++

	if result.Success {
		ft.suite.Passed++
		fmt.Printf("✅ %s - %s (%d ms)\n",
			result.TestName,
			result.Endpoint,
			result.Duration.Milliseconds())
	} else {
		ft.suite.Failed++
		fmt.Printf("❌ %s - %s (Error: %s)\n",
			result.TestName,
			result.Endpoint,
			result.Error)
	}
}

// finalizeSuite completa la suite de pruebas y genera reporte
func (ft *EndpointTester) finalizeSuite() {
	ctx := context.Background()
	ft.suite.EndTime = time.Now()
	ft.suite.Duration = ft.suite.EndTime.Sub(ft.suite.StartTime)

	if ft.suite.TotalTests > 0 {
		ft.suite.SuccessRate = float64(ft.suite.Passed) / float64(ft.suite.TotalTests) * 100
	}

	// Log resumen
	ft.logger.Info(ctx, "🏁 Suite de pruebas completada",
		logger.Int("total_tests", ft.suite.TotalTests),
		logger.Int("passed", ft.suite.Passed),
		logger.Int("failed", ft.suite.Failed),
		logger.Float64("success_rate", ft.suite.SuccessRate),
		logger.Duration("duration", ft.suite.Duration),
	)

	ft.testLogger.Info(ctx, "=== RESUMEN DE PRUEBAS ===",
		logger.Int("total_tests", ft.suite.TotalTests),
		logger.Int("passed", ft.suite.Passed),
		logger.Int("failed", ft.suite.Failed),
		logger.Float64("success_rate", ft.suite.SuccessRate),
		logger.Duration("total_duration", ft.suite.Duration),
	)

	// Generar reporte JSON
	ft.generateJSONReport()

	// Mostrar resumen en consola
	ft.printSummary()
}

// generateJSONReport genera un reporte detallado en JSON
func (ft *EndpointTester) generateJSONReport() {
	timestamp := time.Now().Format("20060102_150405")
	reportFile := filepath.Join("logs", fmt.Sprintf("finnhub_test_report_%s.json", timestamp))

	data, err := json.MarshalIndent(ft.suite, "", "  ")
	if err != nil {
		ft.logger.Error(context.Background(), "Error generando reporte JSON", err)
		return
	}

	if err := os.WriteFile(reportFile, data, 0644); err != nil {
		ft.logger.Error(context.Background(), "Error escribiendo reporte", err)
		return
	}

	ft.logger.Info(context.Background(), "📄 Reporte JSON generado",
		logger.String("file", reportFile),
	)
}

// printSummary muestra un resumen en la consola
func (ft *EndpointTester) printSummary() {
	fmt.Print("\n" + strings.Repeat("=", 60) + "\n")
	fmt.Printf("🧪 RESUMEN DE PRUEBAS FINNHUB API\n")
	fmt.Print(strings.Repeat("=", 60) + "\n")
	fmt.Printf("📊 Total de pruebas: %d\n", ft.suite.TotalTests)
	fmt.Printf("✅ Exitosas: %d\n", ft.suite.Passed)
	fmt.Printf("❌ Fallidas: %d\n", ft.suite.Failed)
	fmt.Printf("📈 Tasa de éxito: %.1f%%\n", ft.suite.SuccessRate)
	fmt.Printf("⏱️  Duración total: %v\n", ft.suite.Duration.Round(time.Millisecond))
	fmt.Print(strings.Repeat("=", 60) + "\n")

	if ft.suite.Failed > 0 {
		fmt.Printf("\n❌ PRUEBAS FALLIDAS:\n")
		for _, test := range ft.suite.Tests {
			if !test.Success {
				fmt.Printf("  • %s: %s\n", test.TestName, test.Error)
			}
		}
		fmt.Printf("\n")
	}

	if ft.suite.SuccessRate >= 80 {
		fmt.Printf("🎉 ¡INTEGRACIÓN FINNHUB EXITOSA!\n")
	} else {
		fmt.Printf("⚠️  INTEGRACIÓN FINNHUB NECESITA REVISIÓN\n")
	}
	fmt.Printf("\n📁 Logs generados en: logs/\n")
	fmt.Printf("📄 Reporte detallado: logs/finnhub_test_report_*.json\n")
	fmt.Print(strings.Repeat("=", 60) + "\n")
}

// runAllTests ejecuta todas las pruebas de la suite
func (ft *EndpointTester) runAllTests() error {
	ctx := context.Background()

	ft.logger.Info(ctx, "🧪 Iniciando suite de pruebas de Finnhub API",
		logger.String("base_url", ft.baseURL),
		logger.Time("start_time", ft.suite.StartTime),
	)

	ft.testLogger.Info(ctx, "=== INICIANDO SUITE DE PRUEBAS FINNHUB ===",
		logger.String("version", ft.suite.Version),
		logger.String("base_url", ft.baseURL),
	)

	// Verificar servidor antes de empezar
	if err := ft.checkServerHealth(); err != nil {
		return fmt.Errorf("servidor no está disponible: %w", err)
	}

	// Símbolos de prueba
	testSymbols := []string{"AAPL", "MSFT", "GOOGL"}

	// Test 1: Health check
	ft.addTest(ft.makeRequest(
		"Server Health Check",
		ft.baseURL[:strings.LastIndex(ft.baseURL, "/api")]+"/health",
		"GET",
	))

	// Test 2: Cotizaciones en tiempo real (Finnhub)
	for _, symbol := range testSymbols {
		ft.addTest(ft.makeRequest(
			fmt.Sprintf("Real-time Quote - %s", symbol),
			fmt.Sprintf("%s/market-data/quote/%s", ft.baseURL, symbol),
			"GET",
		))
	}

	// Test 3: Perfiles de empresa (Finnhub)
	for _, symbol := range testSymbols[:2] { // Solo 2 para no sobrecargar
		ft.addTest(ft.makeRequest(
			fmt.Sprintf("Company Profile - %s", symbol),
			fmt.Sprintf("%s/market-data/profile/%s", ft.baseURL, symbol),
			"GET",
		))
	}

	// Test 4: Noticias de empresa (Finnhub)
	for _, symbol := range testSymbols[:1] { // Solo 1 para pruebas
		ft.addTest(ft.makeRequest(
			fmt.Sprintf("Company News - %s", symbol),
			fmt.Sprintf("%s/market-data/news/%s?days=7", ft.baseURL, symbol),
			"GET",
		))
	}

	// Test 5: Métricas financieras (Finnhub)
	for _, symbol := range testSymbols[:2] {
		ft.addTest(ft.makeRequest(
			fmt.Sprintf("Basic Financials - %s", symbol),
			fmt.Sprintf("%s/market-data/financials/%s", ft.baseURL, symbol),
			"GET",
		))
	}
	// Test 6: Market Overview (Finnhub)
	ft.addTest(ft.makeRequest(
		"Market Overview",
		fmt.Sprintf("%s/market-data/overview", ft.baseURL),
		"GET",
	))

	// Test 7: Recommendation Trends (Finnhub)
	ft.addTest(ft.makeRequest(
		"Finnhub Recommendation Trends - AAPL",
		fmt.Sprintf("%s/market-data/recommendations/%s", ft.baseURL, testSymbols[0]),
		"GET",
	))

	// Test 8: Earnings Calendar (Finnhub)
	ft.addTest(ft.makeRequest(
		"Finnhub Earnings Calendar",
		fmt.Sprintf("%s/market-data/earnings-calendar", ft.baseURL),
		"GET",
	))

	// Test 9: Market Status (Finnhub)
	ft.addTest(ft.makeRequest(
		"Finnhub Market Status",
		fmt.Sprintf("%s/market-data/market-status", ft.baseURL),
		"GET",
	))

	// Test 10: Endpoints que deberían fallar (para probar manejo de errores)
	ft.addTest(ft.makeRequest(
		"Invalid Symbol Test",
		fmt.Sprintf("%s/market-data/quote/INVALID_SYMBOL_TEST", ft.baseURL),
		"GET",
	))

	ft.addTest(ft.makeRequest(
		"Non-existent Endpoint Test",
		fmt.Sprintf("%s/market-data/nonexistent", ft.baseURL),
		"GET"))
	// Finalizar suite
	ft.finalizeSuite()

	return nil
}

// runFinnhubTests ejecuta solo las pruebas específicas de Finnhub
func (ft *EndpointTester) runFinnhubTests() error {
	ctx := context.Background()

	ft.logger.Info(ctx, "🧪 Iniciando suite de pruebas específicas de Finnhub API",
		logger.String("base_url", ft.baseURL),
		logger.Time("start_time", ft.suite.StartTime),
	)

	ft.testLogger.Info(ctx, "=== INICIANDO SUITE DE PRUEBAS FINNHUB ===",
		logger.String("version", ft.suite.Version),
		logger.String("base_url", ft.baseURL),
	)

	// Verificar servidor antes de empezar
	if err := ft.checkServerHealth(); err != nil {
		return fmt.Errorf("servidor no está disponible: %w", err)
	}

	// Símbolos de prueba para Finnhub
	testSymbols := []string{"AAPL", "MSFT", "GOOGL", "TSLA"}

	// Test 1: Health check
	ft.addTest(ft.makeRequest(
		"Server Health Check",
		ft.baseURL[:strings.LastIndex(ft.baseURL, "/api")]+"/health",
		"GET",
	))

	// Test 2: Cotizaciones en tiempo real (Finnhub)
	fmt.Println("🔍 Testing Finnhub Real-time Quotes...")
	for _, symbol := range testSymbols {
		ft.addTest(ft.makeRequest(
			fmt.Sprintf("Finnhub Real-time Quote - %s", symbol),
			fmt.Sprintf("%s/market-data/quote/%s", ft.baseURL, symbol),
			"GET",
		))
	}

	// Test 3: Perfiles de empresa (Finnhub)
	fmt.Println("🏢 Testing Finnhub Company Profiles...")
	for _, symbol := range testSymbols[:2] { // Solo 2 para no sobrecargar
		ft.addTest(ft.makeRequest(
			fmt.Sprintf("Finnhub Company Profile - %s", symbol),
			fmt.Sprintf("%s/market-data/profile/%s", ft.baseURL, symbol),
			"GET",
		))
	}

	// Test 4: Noticias de empresa (Finnhub)
	fmt.Println("📰 Testing Finnhub Company News...")
	for _, symbol := range testSymbols[:2] {
		ft.addTest(ft.makeRequest(
			fmt.Sprintf("Finnhub Company News - %s", symbol),
			fmt.Sprintf("%s/market-data/news/%s?days=7", ft.baseURL, symbol),
			"GET",
		))
	}

	// Test 5: Métricas financieras (Finnhub)
	fmt.Println("📊 Testing Finnhub Financial Metrics...")
	for _, symbol := range testSymbols[:2] {
		ft.addTest(ft.makeRequest(
			fmt.Sprintf("Finnhub Basic Financials - %s", symbol),
			fmt.Sprintf("%s/market-data/financials/%s", ft.baseURL, symbol),
			"GET",
		))
	}

	// Test 6: Market Overview (Finnhub)
	fmt.Println("🌍 Testing Finnhub Market Overview...")
	ft.addTest(ft.makeRequest(
		"Finnhub Market Overview",
		fmt.Sprintf("%s/market-data/overview", ft.baseURL),
		"GET",
	))

	// Test 7: Recommendation Trends (Finnhub)
	fmt.Println("📈 Testing Finnhub Recommendation Trends...")
	ft.addTest(ft.makeRequest(
		"Finnhub Recommendation Trends - AAPL",
		fmt.Sprintf("%s/market-data/recommendations/%s", ft.baseURL, testSymbols[0]),
		"GET",
	))

	// Test 8: Earnings Calendar (Finnhub)
	fmt.Println("📅 Testing Finnhub Earnings Calendar...")
	ft.addTest(ft.makeRequest(
		"Finnhub Earnings Calendar",
		fmt.Sprintf("%s/market-data/earnings-calendar", ft.baseURL),
		"GET",
	))

	// Test 9: Market Status (Finnhub)
	fmt.Println("⏰ Testing Finnhub Market Status...")
	ft.addTest(ft.makeRequest(
		"Finnhub Market Status",
		fmt.Sprintf("%s/market-data/market-status", ft.baseURL),
		"GET",
	))

	// Test 10: Pruebas de manejo de errores
	fmt.Println("🚫 Testing Error Handling...")
	ft.addTest(ft.makeRequest(
		"Finnhub Invalid Symbol Test",
		fmt.Sprintf("%s/market-data/quote/INVALID_SYMBOL_TEST", ft.baseURL),
		"GET",
	))
	// Finalizar suite
	ft.finalizeSuite()

	return nil
}

// RunFinnhubIntegrationTest ejecuta las pruebas de integración de Finnhub específicamente
func RunFinnhubIntegrationTest() error {
	fmt.Printf("🚀 Finnhub Integration Tester\n")
	fmt.Printf("=============================\n\n")

	// Crear tester
	tester, err := NewEndpointTester()
	if err != nil {
		fmt.Printf("❌ Error inicializando tester: %v\n", err)
		return err
	}
	// Ejecutar pruebas específicas de Finnhub
	if err := tester.runAllTests(); err != nil {
		fmt.Printf("❌ Error ejecutando pruebas: %v\n", err)
		return err
	}

	// Cerrar loggers
	if err := tester.logger.Close(); err != nil {
		fmt.Printf("⚠️  Advertencia: Error cerrando logger: %v\n", err)
	}

	if err := tester.testLogger.Close(); err != nil {
		fmt.Printf("⚠️  Advertencia: Error cerrando test logger: %v\n", err)
	}

	// Retornar error si el éxito es bajo
	if tester.suite.SuccessRate < 80 {
		return fmt.Errorf("finnhub test success rate too low: %.1f%%", tester.suite.SuccessRate)
	}

	return nil
}
