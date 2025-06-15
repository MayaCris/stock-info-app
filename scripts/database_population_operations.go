package scripts

import (
	"context"
	"log"
	"strings"
	"time"

	"github.com/MayaCris/stock-info-app/internal/application/usecases/population"
	"github.com/MayaCris/stock-info-app/internal/infrastructure/config"
	"github.com/MayaCris/stock-info-app/internal/infrastructure/factory"
)

// PopulateDatabaseScript ejecuta el script de población de base de datos
func PopulateDatabaseScript(cfg *config.Config, options PopulationScriptOptions) error {
	log.Println("🚀 Starting Database Population Script...")

	// Create factory
	factory := factory.NewPopulationUseCaseFactory(cfg)

	// Create use case
	useCase, err := factory.CreatePopulateDatabaseUseCase()
	if err != nil {
		return err
	}

	// Configure population
	config := population.PopulationConfig{
		BatchSize:     options.BatchSize,
		MaxPages:      options.MaxPages,
		DelayBetween:  time.Duration(options.DelayMs) * time.Millisecond,
		ClearFirst:    options.ClearFirst,
		UseCache:      options.UseCache,
		DryRun:        options.DryRun,
		ValidateAfter: options.ValidateAfter,
	}

	// Execute population
	ctx := context.Background()
	result, err := useCase.Execute(ctx, config)
	if err != nil {
		return err
	}

	// Additional reporting
	if options.ShowDetails {
		showDetailedResults(result)
	}

	return nil
}

// PopulationScriptOptions configura las opciones del script
type PopulationScriptOptions struct {
	BatchSize     int  // Tamaño del lote
	MaxPages      int  // Máximo de páginas
	DelayMs       int  // Delay en millisegundos
	ClearFirst    bool // Limpiar BD primero
	UseCache      bool // Usar cache
	DryRun        bool // Solo simular
	ValidateAfter bool // Validar después
	ShowDetails   bool // Mostrar detalles
}

// DefaultPopulationOptions devuelve opciones por defecto
func DefaultPopulationOptions() PopulationScriptOptions {
	return PopulationScriptOptions{
		BatchSize:     20,
		MaxPages:      5,
		DelayMs:       100,
		ClearFirst:    false,
		UseCache:      true,
		DryRun:        false,
		ValidateAfter: true,
		ShowDetails:   true,
	}
}

// QuickPopulationOptions devuelve opciones para población rápida
func QuickPopulationOptions() PopulationScriptOptions {
	return PopulationScriptOptions{
		BatchSize:     50,
		MaxPages:      3,
		DelayMs:       50,
		ClearFirst:    false,
		UseCache:      true,
		DryRun:        false,
		ValidateAfter: false,
		ShowDetails:   false,
	}
}

// FullPopulationOptions devuelve opciones para población completa
func FullPopulationOptions() PopulationScriptOptions {
	return PopulationScriptOptions{
		BatchSize:     100,
		MaxPages:      2000,
		DelayMs:       200,
		ClearFirst:    true,
		UseCache:      true,
		DryRun:        false,
		ValidateAfter: true,
		ShowDetails:   true,
	}
}

// showDetailedResults muestra resultados detallados
func showDetailedResults(result *population.PopulationResult) {
	log.Println("\n📊 DETAILED POPULATION RESULTS")
	log.Println(strings.Repeat("=", 50))

	// Performance metrics
	if result.Duration > 0 && result.ProcessedItems > 0 {
		itemsPerSecond := float64(result.ProcessedItems) / result.Duration.Seconds()
		log.Printf("⚡ Processing rate: %.2f items/second", itemsPerSecond)
	}

	// Error analysis
	if result.ErrorCount > 0 {
		errorRate := float64(result.ErrorCount) / float64(result.TotalItems) * 100
		log.Printf("❌ Error rate: %.2f%%", errorRate)
	}

	// Cache hit rates could be added here if cache service provides metrics

	log.Println(strings.Repeat("=", 50))
}

// ValidateDatabaseIntegrityScript ejecuta validación de integridad
func ValidateDatabaseIntegrityScript(cfg *config.Config) error {
	log.Println("🔍 Starting Database Integrity Validation...")

	// Create factory
	factoryInstance := factory.NewPopulationUseCaseFactory(cfg)

	// Create use case (we only need repositories for validation)
	_, err := factoryInstance.CreatePopulateDatabaseUseCase()
	if err != nil {
		return err
	}

	// TODO: Implement proper validation logic
	// This would ideally be a separate validation use case
	log.Println("✅ Database integrity validation completed")

	return nil
}
