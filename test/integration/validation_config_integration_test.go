package integration

import (
	"context"
	"os"
	"testing"
	"time"

	"github.com/MayaCris/stock-info-app/internal/domain/repositories/implementation"
	"github.com/MayaCris/stock-info-app/internal/domain/services"
	"github.com/MayaCris/stock-info-app/internal/infrastructure/config"
	"github.com/MayaCris/stock-info-app/internal/infrastructure/database/cockroachdb"
	"github.com/MayaCris/stock-info-app/internal/infrastructure/factory"
	"github.com/MayaCris/stock-info-app/internal/infrastructure/logger"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestValidationConfigurationIntegration tests the validation system with real data
func TestValidationConfigurationIntegration(t *testing.T) {
	// Configurar variables de entorno básicas para el test si no existen
	if os.Getenv("APP_ENV") == "" {
		os.Setenv("APP_ENV", "development")
	}

	// Cargar configuración real
	cfg, err := config.Load()
	if err != nil {
		t.Skipf("Skipping integration test: Failed to load configuration: %v", err)
		return
	}

	// Conectar a la base de datos real
	db, err := cockroachdb.NewConnection(cfg)
	if err != nil {
		t.Skipf("Skipping integration test: Failed to connect to database: %v", err)
		return
	}
	defer db.Close()

	// Crear repositorios
	companyRepo := implementation.NewTransactionalCompanyRepository(db.DB)
	brokerageRepo := implementation.NewTransactionalBrokerageRepository(db.DB)
	stockRatingRepo := implementation.NewTransactionalStockRatingRepository(db.DB)

	// Crear logger base
	baseLogger, err := logger.InitializeGlobalLogger()
	require.NoError(t, err, "Failed to initialize logger")

	ctx := context.Background()

	t.Run("Test_Default_Configuration_Works", func(t *testing.T) {
		// Crear servicio con configuración por defecto
		integrityService := services.NewIntegrityValidationServiceWithDefaults(
			companyRepo,
			brokerageRepo,
			stockRatingRepo,
			logger.NewIntegrityLogger(baseLogger, &logger.LogConfig{}),
		)

		// Ejecutar validación con configuración por defecto
		// Esto usa la nueva configuración en lugar de valores hardcodeados
		report, err := integrityService.ValidateFullIntegrity(ctx)

		assert.NoError(t, err, "Validation should not fail")
		assert.NotNil(t, report, "Report should not be nil")

		// Verificar que la validación retorna resultados estructurados
		t.Logf("✅ Integrity validation completed successfully")
		t.Logf("📊 Overall Status: %s", report.OverallStatus)
		t.Logf("🔍 Total Issues: %d", report.TotalIssues)
		t.Logf("⚠️ Warning Issues: %d", report.WarningIssues)
		t.Logf("❌ Critical Issues: %d", report.CriticalIssues)

		// Verificar estructura básica del reporte
		assert.Contains(t, []string{"GOOD", "WARNING", "CRITICAL"}, report.OverallStatus, "Status should be valid")
		assert.GreaterOrEqual(t, report.TotalIssues, 0, "Total issues should not be negative")
		assert.Equal(t, report.TotalIssues, report.WarningIssues+report.CriticalIssues, "Total should equal sum of warning and critical")
	})

	t.Run("Test_Custom_Strict_Configuration", func(t *testing.T) {
		// Crear servicio con configuración más estricta (como producción)
		integrityService := factory.CreateIntegrityValidationServiceWithCustomConfig(
			companyRepo,
			brokerageRepo,
			stockRatingRepo,
			baseLogger,
			true, // isProduction = true (más estricto)
		)

		// Ejecutar validación con configuración estricta
		report, err := integrityService.ValidateFullIntegrity(ctx)

		assert.NoError(t, err, "Validation should not fail with strict config")
		assert.NotNil(t, report, "Report should not be nil")

		t.Logf("✅ Strict validation completed successfully")
		t.Logf("📊 Overall Status: %s", report.OverallStatus)
		t.Logf("🔍 Total Issues: %d", report.TotalIssues)

		// Con configuración más estricta, podríamos esperar más issues
		assert.Contains(t, []string{"GOOD", "WARNING", "CRITICAL"}, report.OverallStatus, "Status should be valid")
	})

	t.Run("Test_Custom_Lenient_Configuration", func(t *testing.T) {
		// Crear servicio con configuración más permisiva (como desarrollo)
		integrityService := factory.CreateIntegrityValidationServiceWithCustomConfig(
			companyRepo,
			brokerageRepo,
			stockRatingRepo,
			baseLogger,
			false, // isProduction = false (más permisivo)
		)

		// Ejecutar validación con configuración permisiva
		report, err := integrityService.ValidateFullIntegrity(ctx)

		assert.NoError(t, err, "Validation should not fail with lenient config")
		assert.NotNil(t, report, "Report should not be nil")

		t.Logf("✅ Lenient validation completed successfully")
		t.Logf("📊 Overall Status: %s", report.OverallStatus)
		t.Logf("🔍 Total Issues: %d", report.TotalIssues)

		// Con configuración más permisiva, podríamos esperar menos issues
		assert.Contains(t, []string{"GOOD", "WARNING", "CRITICAL"}, report.OverallStatus, "Status should be valid")
	})

	t.Run("Test_Configuration_Consistency", func(t *testing.T) {
		// Verificar que la configuración por defecto es consistente
		defaultConfig := services.DefaultValidationConfig()

		assert.NotNil(t, defaultConfig, "Default config should not be nil")
		assert.NotNil(t, defaultConfig.Rules, "Default rules should not be nil")
		assert.NotNil(t, defaultConfig.Rules.Company, "Company rules should not be nil")
		assert.NotNil(t, defaultConfig.Rules.Brokerage, "Brokerage rules should not be nil")
		assert.NotNil(t, defaultConfig.Rules.StockRating, "StockRating rules should not be nil")
		assert.NotNil(t, defaultConfig.Thresholds, "Thresholds should not be nil")
		// Verificar valores razonables
		assert.Greater(t, defaultConfig.Rules.Company.ViolationsForCritical, 0, "Critical violations threshold should be positive")
		assert.Greater(t, defaultConfig.Rules.Company.NameMaxLength, 0, "Max name length should be positive")
		assert.Greater(t, defaultConfig.Rules.StockRating.MaxAgeYearsBusiness, 0, "Max age should be positive")

		t.Logf("✅ Configuration consistency verified")
		t.Logf("📋 Company Critical Violations: %d", defaultConfig.Rules.Company.ViolationsForCritical)
		t.Logf("📋 Company Max Name Length: %d", defaultConfig.Rules.Company.NameMaxLength)
		t.Logf("📋 Stock Rating Max Age: %d years", defaultConfig.Rules.StockRating.MaxAgeYearsBusiness)
	})

	t.Run("Test_Performance_With_Real_Data", func(t *testing.T) {
		// Crear servicio con configuración por defecto
		integrityService := services.NewIntegrityValidationServiceWithDefaults(
			companyRepo,
			brokerageRepo,
			stockRatingRepo,
			logger.NewIntegrityLogger(baseLogger, &logger.LogConfig{}),
		)

		// Medir el tiempo de ejecución
		start := time.Now()
		report, err := integrityService.ValidateFullIntegrity(ctx)
		duration := time.Since(start)

		assert.NoError(t, err, "Performance test should not fail")
		assert.NotNil(t, report, "Report should not be nil")

		t.Logf("⏱️ Validation completed in: %v", duration)
		t.Logf("📊 Performance Result - Status: %s, Issues: %d", report.OverallStatus, report.TotalIssues)

		// Verificar que la validación no toma demasiado tiempo (ajustar según necesidades)
		assert.Less(t, duration, 30*time.Second, "Validation should complete within reasonable time")
	})

	t.Run("Test_Backward_Compatibility", func(t *testing.T) {
		// Verificar que el constructor por defecto funciona igual que antes
		integrityService1 := services.NewIntegrityValidationServiceWithDefaults(
			companyRepo,
			brokerageRepo,
			stockRatingRepo,
			logger.NewIntegrityLogger(baseLogger, &logger.LogConfig{}),
		)

		// Comparar con constructor con configuración explícita por defecto
		defaultConfig := services.DefaultValidationConfig()
		integrityService2 := services.NewIntegrityValidationService(
			companyRepo,
			brokerageRepo,
			stockRatingRepo,
			logger.NewIntegrityLogger(baseLogger, &logger.LogConfig{}),
			defaultConfig,
		)

		// Ambos deberían dar resultados idénticos
		report1, err1 := integrityService1.ValidateFullIntegrity(ctx)
		report2, err2 := integrityService2.ValidateFullIntegrity(ctx)

		assert.NoError(t, err1, "First validation should not fail")
		assert.NoError(t, err2, "Second validation should not fail")
		assert.NotNil(t, report1, "First report should not be nil")
		assert.NotNil(t, report2, "Second report should not be nil")

		// Los reportes deberían ser idénticos
		assert.Equal(t, report1.OverallStatus, report2.OverallStatus, "Overall status should match")
		assert.Equal(t, report1.TotalIssues, report2.TotalIssues, "Total issues should match")
		assert.Equal(t, report1.WarningIssues, report2.WarningIssues, "Warning issues should match")
		assert.Equal(t, report1.CriticalIssues, report2.CriticalIssues, "Critical issues should match")

		t.Logf("✅ Backward compatibility verified")
		t.Logf("📊 Both services returned identical results: Status=%s, Issues=%d",
			report1.OverallStatus, report1.TotalIssues)
	})
}

// TestValidationConfigurationUnitScenarios tests specific configuration scenarios
func TestValidationConfigurationUnitScenarios(t *testing.T) {
	t.Run("Test_Configuration_Customization", func(t *testing.T) {
		// Crear configuración personalizada
		customConfig := services.DefaultValidationConfig()

		// Modificar algunos valores
		originalCritical := customConfig.Rules.Company.ViolationsForCritical
		customConfig.Rules.Company.ViolationsForCritical = 1
		customConfig.Rules.StockRating.MaxAgeYearsBusiness = 10
		customConfig.Thresholds.BusinessRulesWarningLimit = 5

		// Verificar que los cambios se aplicaron
		assert.Equal(t, 1, customConfig.Rules.Company.ViolationsForCritical, "Custom critical violations should be set")
		assert.Equal(t, 10, customConfig.Rules.StockRating.MaxAgeYearsBusiness, "Custom max age should be set")
		assert.Equal(t, 5, customConfig.Thresholds.BusinessRulesWarningLimit, "Custom warning limit should be set")

		// Verificar que los valores por defecto no cambiaron
		defaultConfig := services.DefaultValidationConfig()
		assert.Equal(t, originalCritical, defaultConfig.Rules.Company.ViolationsForCritical, "Default should not be modified")

		t.Logf("✅ Configuration customization works correctly")
	})

	t.Run("Test_Helper_Functions_Exist", func(t *testing.T) {
		// Verificar que las funciones helper están disponibles (esto es más una prueba de compilación)
		config := services.DefaultValidationConfig()

		// Estos helpers deberían estar disponibles aunque sean internos
		// El test asegura que la refactorización no rompió la estructura
		assert.NotNil(t, config, "Config should be available")
		assert.NotEmpty(t, config.Rules.Company, "Company rules should not be empty")
		assert.NotEmpty(t, config.Rules.Brokerage, "Brokerage rules should not be empty")
		assert.NotEmpty(t, config.Rules.StockRating, "StockRating rules should not be empty")
		t.Logf("✅ All helper structures are accessible")
	})
}
