package routes

import (
	"github.com/gin-gonic/gin"

	"github.com/MayaCris/stock-info-app/internal/presentation/rest/handlers"
)

// CompanyRoutes encapsula la configuración de rutas de companies
type CompanyRoutes struct {
	middlewareManager *MiddlewareManager
}

// NewCompanyRoutes crea una nueva instancia del configurador de rutas de companies
func NewCompanyRoutes(middlewareManager *MiddlewareManager) *CompanyRoutes {
	return &CompanyRoutes{
		middlewareManager: middlewareManager,
	}
}

// SetupCompanyRoutes configura todas las rutas relacionadas con companies
// Esta función configura operaciones CRUD, gestión de estado y búsquedas
func (cr *CompanyRoutes) SetupCompanyRoutes(routerGroup *gin.RouterGroup, companyHandler *handlers.CompanyHandler) {
	// Verificar que el handler existe
	if companyHandler == nil {
		return
	}

	// Configurar el grupo de rutas de companies
	companies := routerGroup.Group("/companies")
	{
		// CRUD operations
		cr.setupCRUDRoutes(companies, companyHandler)

		// State management operations
		cr.setupStateRoutes(companies, companyHandler)

		// Search and filter operations
		cr.setupSearchRoutes(companies, companyHandler)
	}
}

// setupCRUDRoutes configura las operaciones básicas CRUD
func (cr *CompanyRoutes) setupCRUDRoutes(companies *gin.RouterGroup, companyHandler *handlers.CompanyHandler) {
	// Grupo para operaciones de escritura (CREATE, UPDATE, DELETE)
	writeOps := companies.Group("")
	if cr.middlewareManager != nil {
		cr.middlewareManager.ApplyWriteMiddlewares(writeOps)
	}
	{
		// Create - Crear una nueva company
		writeOps.POST("/", companyHandler.CreateCompany)

		// Update - Actualizar company completa
		writeOps.PUT("/:id", companyHandler.UpdateCompany)

		// Delete - Eliminar company
		writeOps.DELETE("/:id", companyHandler.DeleteCompany)
	}

	// Grupo para operaciones de lectura (READ, LIST)
	readOps := companies.Group("")
	if cr.middlewareManager != nil {
		cr.middlewareManager.ApplyReadOnlyMiddlewares(readOps)
	}
	{
		// Read operations
		readOps.GET("/:id", companyHandler.GetCompanyByID)
		readOps.GET("/ticker/:ticker", companyHandler.GetCompanyByTicker)

		// List operations
		readOps.GET("/", companyHandler.ListCompanies)
		readOps.GET("/active", companyHandler.ListActiveCompanies)
	}
}

// setupStateRoutes configura las rutas de gestión de estado
func (cr *CompanyRoutes) setupStateRoutes(companies *gin.RouterGroup, companyHandler *handlers.CompanyHandler) {
	// Grupo para operaciones de administración (requieren permisos especiales)
	adminOps := companies.Group("")
	if cr.middlewareManager != nil {
		cr.middlewareManager.ApplyAdminMiddlewares(adminOps)
	}
	{
		// Activación y desactivación
		adminOps.PATCH("/:id/activate", companyHandler.ActivateCompany)
		adminOps.PATCH("/:id/deactivate", companyHandler.DeactivateCompany)

		// Actualización de market cap
		adminOps.PATCH("/:id/market-cap", companyHandler.UpdateMarketCap)

		// Futuras operaciones de estado se pueden agregar aquí
		// adminOps.PATCH("/:id/suspend", companyHandler.SuspendCompany)
		// adminOps.PATCH("/:id/verify", companyHandler.VerifyCompany)
	}
}

// setupSearchRoutes configura las rutas de búsqueda y filtrado
func (cr *CompanyRoutes) setupSearchRoutes(companies *gin.RouterGroup, companyHandler *handlers.CompanyHandler) {
	// Grupo para operaciones de búsqueda
	searchOps := companies.Group("")
	if cr.middlewareManager != nil {
		cr.middlewareManager.ApplySearchMiddlewares(searchOps)
	}
	{
		// Búsqueda por nombre
		searchOps.GET("/search", companyHandler.SearchCompaniesByName)

		// Filtrado por sector
		searchOps.GET("/sector/:sector", companyHandler.GetCompaniesBySector)

		// Futuras búsquedas se pueden agregar aquí
		// searchOps.GET("/market-cap/range", companyHandler.GetCompaniesByMarketCapRange)
		// searchOps.GET("/country/:country", companyHandler.GetCompaniesByCountry)
	}
}

// GetCompanyRoutesInfo retorna información sobre las rutas de companies disponibles
func (cr *CompanyRoutes) GetCompanyRoutesInfo() map[string]interface{} {
	return map[string]interface{}{
		"entity":    "companies",
		"base_path": "/companies",
		"operations": map[string][]string{
			"crud": {
				"POST /companies",
				"GET /companies/:id",
				"GET /companies/ticker/:ticker",
				"PUT /companies/:id",
				"DELETE /companies/:id",
				"GET /companies",
				"GET /companies/active",
			},
			"state_management": {
				"PATCH /companies/:id/activate",
				"PATCH /companies/:id/deactivate",
				"PATCH /companies/:id/market-cap",
			},
			"search": {
				"GET /companies/search",
				"GET /companies/sector/:sector",
			},
		},
	}
}
