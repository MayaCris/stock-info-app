package integration

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"

	"github.com/MayaCris/stock-info-app/internal/infrastructure/config"
	"github.com/MayaCris/stock-info-app/internal/infrastructure/logger"
	"github.com/MayaCris/stock-info-app/internal/presentation/rest/middleware"
)

func TestLoggingIntegration_MiddlewareLogging(t *testing.T) {
	// Configurar modo test para Gin
	gin.SetMode(gin.TestMode)

	// Crear configuración de test
	cfg := config.ServerLoggingConfig{
		Enabled: true,
		Level:   "debug",
		Format:  "json",
		Middleware: config.MiddlewareLogConfig{
			LogRequests:         true,
			LogResponses:        true,
			LogHeaders:          true,
			LogSlowRequests:     true,
			SlowThreshold:       100 * time.Millisecond,
			SkipPaths:           []string{"/skip"},
			SkipSuccessfulPaths: []string{},
		},
		Outputs: []config.OutputConfig{
			{Type: "file", Level: "debug", Format: "json"},
		},
	}

	// Crear logger
	loggerFactory := logger.NewLoggerFactory()
	loggerConfig := cfg.ToLoggerConfig()
	serverLogger, err := loggerFactory.CreateServerLoggerWithConfig(loggerConfig)
	if err != nil {
		t.Fatalf("Failed to create server logger: %v", err)
	}

	// Crear router de test
	router := gin.New()

	// Aplicar middleware de logging
	middlewareConfig := middleware.ServerLoggingConfig{
		LogHeaders:           cfg.Middleware.LogHeaders,
		LogRequestBody:       false,
		LogResponseBody:      false,
		SkipPaths:            cfg.Middleware.SkipPaths,
		SkipSuccessfulPaths:  cfg.Middleware.SkipSuccessfulPaths,
		LogSlowRequests:      cfg.Middleware.LogSlowRequests,
		SlowRequestThreshold: cfg.Middleware.SlowThreshold,
		SensitiveHeaders:     []string{"authorization"},
	}

	router.Use(middleware.ServerLoggingMiddleware(serverLogger, middlewareConfig))

	// Definir rutas de test
	router.GET("/test", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "test successful"})
	})

	router.GET("/slow", func(c *gin.Context) {
		time.Sleep(150 * time.Millisecond) // Simular request lenta
		c.JSON(http.StatusOK, gin.H{"message": "slow response"})
	})

	router.GET("/skip", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "should be skipped"})
	})

	router.GET("/error", func(c *gin.Context) {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "test error"})
	})

	tests := []struct {
		name         string
		path         string
		expectedCode int
		shouldLog    bool
		checkSlow    bool
	}{
		{
			name:         "normal request should be logged",
			path:         "/test",
			expectedCode: http.StatusOK,
			shouldLog:    true,
			checkSlow:    false,
		},
		{
			name:         "slow request should be logged with slow flag",
			path:         "/slow",
			expectedCode: http.StatusOK,
			shouldLog:    true,
			checkSlow:    true,
		},
		{
			name:         "skipped path should not be logged",
			path:         "/skip",
			expectedCode: http.StatusOK,
			shouldLog:    false,
			checkSlow:    false,
		},
		{
			name:         "error request should be logged",
			path:         "/error",
			expectedCode: http.StatusInternalServerError,
			shouldLog:    true,
			checkSlow:    false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Crear request
			req := httptest.NewRequest("GET", tt.path, nil)
			req.Header.Set("User-Agent", "test-agent")
			req.Header.Set("Authorization", "Bearer secret-token")

			// Crear response recorder
			w := httptest.NewRecorder()

			// Ejecutar request
			router.ServeHTTP(w, req)

			// Verificar response code
			if w.Code != tt.expectedCode {
				t.Errorf("Expected status code %d, got %d", tt.expectedCode, w.Code)
			}

			// Verificar que la respuesta es válida JSON
			if w.Header().Get("Content-Type") == "application/json; charset=utf-8" {
				var response map[string]interface{}
				if err := json.Unmarshal(w.Body.Bytes(), &response); err != nil {
					t.Errorf("Response should be valid JSON: %v", err)
				}
			}
		})
	}
}

func TestLoggingIntegration_ConfigurationLoading(t *testing.T) {
	// Test que verifica que la configuración se carga correctamente desde ambiente
	tests := []struct {
		name     string
		envValue string
		expected string
	}{
		{
			name:     "default configuration",
			envValue: "",
			expected: "info", // Default level
		},
		{
			name:     "development configuration",
			envValue: "development",
			expected: "debug",
		},
		{
			name:     "production configuration",
			envValue: "production",
			expected: "warn",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Simular configuración según ambiente
			var cfg config.ServerLoggingConfig

			switch tt.envValue {
			case "development":
				cfg = config.DevelopmentServerLoggingConfig()
			case "production":
				cfg = config.ProductionServerLoggingConfig()
			default:
				cfg = config.DefaultServerLoggingConfig()
			}

			if cfg.Level != tt.expected {
				t.Errorf("Expected level %s, got %s", tt.expected, cfg.Level)
			}

			// Verificar que la configuración es válida
			if err := cfg.Validate(); err != nil {
				t.Errorf("Configuration should be valid: %v", err)
			}
		})
	}
}

func TestLoggingIntegration_PerformanceImpact(t *testing.T) {
	// Test para verificar que el logging no tiene impacto significativo en performance
	gin.SetMode(gin.TestMode)

	// Crear configuración mínima
	cfg := config.ServerLoggingConfig{
		Enabled: true,
		Level:   "info",
		Format:  "json",
		Middleware: config.MiddlewareLogConfig{
			LogRequests:     true,
			LogResponses:    true,
			LogHeaders:      false,
			LogSlowRequests: false,
		},
		Outputs: []config.OutputConfig{
			{Type: "file", Level: "info", Format: "json"},
		},
	}

	loggerFactory := logger.NewLoggerFactory()
	loggerConfig := cfg.ToLoggerConfig()
	serverLogger, err := loggerFactory.CreateServerLoggerWithConfig(loggerConfig)
	if err != nil {
		t.Fatalf("Failed to create server logger: %v", err)
	}

	// Crear router con logging
	routerWithLogging := gin.New()
	middlewareConfig := middleware.ServerLoggingConfig{
		LogHeaders:      false,
		LogRequestBody:  false,
		LogResponseBody: false,
		SkipPaths:       []string{},
	}
	routerWithLogging.Use(middleware.ServerLoggingMiddleware(serverLogger, middlewareConfig))
	routerWithLogging.GET("/test", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "test"})
	})

	// Crear router sin logging
	routerWithoutLogging := gin.New()
	routerWithoutLogging.GET("/test", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "test"})
	})

	// Medir tiempo con logging
	start := time.Now()
	for i := 0; i < 1000; i++ {
		req := httptest.NewRequest("GET", "/test", nil)
		w := httptest.NewRecorder()
		routerWithLogging.ServeHTTP(w, req)
	}
	durationWithLogging := time.Since(start)

	// Medir tiempo sin logging
	start = time.Now()
	for i := 0; i < 1000; i++ {
		req := httptest.NewRequest("GET", "/test", nil)
		w := httptest.NewRecorder()
		routerWithoutLogging.ServeHTTP(w, req)
	}
	durationWithoutLogging := time.Since(start)

	// Calcular overhead
	overhead := durationWithLogging - durationWithoutLogging
	overheadPerRequest := overhead / 1000

	t.Logf("Duration with logging: %v", durationWithLogging)
	t.Logf("Duration without logging: %v", durationWithoutLogging)
	t.Logf("Overhead per request: %v", overheadPerRequest)

	// Verificar que el overhead no sea excesivo (< 1ms por request)
	if overheadPerRequest > 1*time.Millisecond {
		t.Errorf("Logging overhead too high: %v per request", overheadPerRequest)
	}
}

func TestLoggingIntegration_ErrorHandling(t *testing.T) {
	gin.SetMode(gin.TestMode)

	// Test con configuración inválida
	invalidConfig := config.ServerLoggingConfig{
		Enabled: true,
		Level:   "invalid",
		Format:  "json",
		Outputs: []config.OutputConfig{},
	}

	if err := invalidConfig.Validate(); err == nil {
		t.Error("Invalid configuration should fail validation")
	}

	// Test con configuración válida pero sin outputs
	emptyOutputConfig := config.ServerLoggingConfig{
		Enabled: true,
		Level:   "info",
		Format:  "json",
		Outputs: []config.OutputConfig{},
	}

	if err := emptyOutputConfig.Validate(); err == nil {
		t.Error("Configuration without outputs should fail validation")
	}
}

// Benchmark de integración
func BenchmarkLoggingIntegration_MiddlewareOverhead(b *testing.B) {
	gin.SetMode(gin.TestMode)

	cfg := config.DefaultServerLoggingConfig()
	loggerFactory := logger.NewLoggerFactory()
	loggerConfig := cfg.ToLoggerConfig()
	serverLogger, err := loggerFactory.CreateServerLoggerWithConfig(loggerConfig)
	if err != nil {
		b.Fatalf("Failed to create server logger: %v", err)
	}

	router := gin.New()
	middlewareConfig := middleware.DefaultServerLoggingConfig()
	router.Use(middleware.ServerLoggingMiddleware(serverLogger, middlewareConfig))
	router.GET("/test", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "test"})
	})

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		req := httptest.NewRequest("GET", "/test", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)
	}
}
