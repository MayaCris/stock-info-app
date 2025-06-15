package routes

import (
	"github.com/gin-gonic/gin"

	"github.com/MayaCris/stock-info-app/internal/presentation/rest/handlers"
)

// AnalysisRoutes encapsula la configuración de rutas de analysis
type AnalysisRoutes struct {
	middlewareManager *MiddlewareManager
}

// NewAnalysisRoutes crea una nueva instancia del configurador de rutas de analysis
func NewAnalysisRoutes(middlewareManager *MiddlewareManager) *AnalysisRoutes {
	return &AnalysisRoutes{
		middlewareManager: middlewareManager,
	}
}

// SetupAnalysisRoutes configura todas las rutas relacionadas con análisis
// Esta función configura análisis por empresa, mercado, sector, tendencias y recomendaciones
func (ar *AnalysisRoutes) SetupAnalysisRoutes(routerGroup *gin.RouterGroup, analysisHandler *handlers.AnalysisHandler) {
	// Verificar que el handler existe
	if analysisHandler == nil {
		return
	}

	// Configurar el grupo de rutas de analysis
	analysis := routerGroup.Group("/analysis")
	{
		// Company analysis routes
		ar.setupCompanyAnalysisRoutes(analysis, analysisHandler)

		// Market analysis routes
		ar.setupMarketAnalysisRoutes(analysis, analysisHandler)

		// Sector analysis routes
		ar.setupSectorAnalysisRoutes(analysis, analysisHandler)

		// Trends analysis routes
		ar.setupTrendsAnalysisRoutes(analysis, analysisHandler)

		// Recommendations routes
		ar.setupRecommendationsRoutes(analysis, analysisHandler)
	}
}

// setupCompanyAnalysisRoutes configura las rutas de análisis por empresa
func (ar *AnalysisRoutes) setupCompanyAnalysisRoutes(analysis *gin.RouterGroup, analysisHandler *handlers.AnalysisHandler) {
	companies := analysis.Group("/companies")
	{
		// Análisis individual por empresa
		companies.GET("/:id", analysisHandler.GetCompanyAnalysis)
		companies.GET("/ticker/:ticker", analysisHandler.GetCompanyAnalysisByTicker)

		// Rankings y comparaciones
		companies.GET("/top-rated", analysisHandler.GetTopRatedCompanies)

		// Futuras rutas de análisis de empresa
		// companies.GET("/:id/performance", analysisHandler.GetCompanyPerformance)
		// companies.GET("/:id/comparison", analysisHandler.CompareCompany)
	}
}

// setupMarketAnalysisRoutes configura las rutas de análisis de mercado
func (ar *AnalysisRoutes) setupMarketAnalysisRoutes(analysis *gin.RouterGroup, analysisHandler *handlers.AnalysisHandler) {
	market := analysis.Group("/market")
	{
		// Overview general del mercado
		market.GET("/overview", analysisHandler.GetMarketOverview)

		// Futuras rutas de análisis de mercado
		// market.GET("/sentiment", analysisHandler.GetMarketSentiment)
		// market.GET("/volatility", analysisHandler.GetMarketVolatility)
		// market.GET("/volume", analysisHandler.GetMarketVolume)
	}
}

// setupSectorAnalysisRoutes configura las rutas de análisis por sector
func (ar *AnalysisRoutes) setupSectorAnalysisRoutes(analysis *gin.RouterGroup, analysisHandler *handlers.AnalysisHandler) {
	sectors := analysis.Group("/sectors")
	{
		// Análisis por sector específico
		sectors.GET("/:sector", analysisHandler.GetSectorAnalysis)

		// Futuras rutas de análisis de sector
		// sectors.GET("/", analysisHandler.GetAllSectorsAnalysis)
		// sectors.GET("/:sector/leaders", analysisHandler.GetSectorLeaders)
		// sectors.GET("/:sector/performance", analysisHandler.GetSectorPerformance)
	}
}

// setupTrendsAnalysisRoutes configura las rutas de análisis de tendencias
func (ar *AnalysisRoutes) setupTrendsAnalysisRoutes(analysis *gin.RouterGroup, analysisHandler *handlers.AnalysisHandler) {
	trends := analysis.Group("/trends")
	{
		// Tendencias de ratings
		trends.GET("/ratings", analysisHandler.GetRatingTrends)

		// Actividad de brokerages
		trends.GET("/brokerages", analysisHandler.GetBrokerageActivity)

		// Futuras rutas de tendencias
		// trends.GET("/sectors", analysisHandler.GetSectorTrends)
		// trends.GET("/volume", analysisHandler.GetVolumeTrends)
		// trends.GET("/sentiment", analysisHandler.GetSentimentTrends)
	}
}

// setupRecommendationsRoutes configura las rutas de recomendaciones
func (ar *AnalysisRoutes) setupRecommendationsRoutes(analysis *gin.RouterGroup, analysisHandler *handlers.AnalysisHandler) {
	recommendations := analysis.Group("/recommendations")
	{
		// Recomendaciones por empresa
		recommendations.GET("/companies/:id", analysisHandler.GenerateRecommendation)

		// Recomendaciones por rating
		recommendations.GET("/rating/:rating", analysisHandler.GetRecommendationsByRating)

		// Futuras rutas de recomendaciones
		// recommendations.GET("/sector/:sector", analysisHandler.GetSectorRecommendations)
		// recommendations.GET("/portfolio", analysisHandler.GetPortfolioRecommendations)
	}
}

// GetAnalysisRoutesInfo retorna información sobre las rutas de analysis disponibles
func (ar *AnalysisRoutes) GetAnalysisRoutesInfo() map[string]interface{} {
	return map[string]interface{}{
		"entity":    "analysis",
		"base_path": "/analysis",
		"operations": map[string][]string{
			"company_analysis": {
				"GET /analysis/companies/:id",
				"GET /analysis/companies/ticker/:ticker",
				"GET /analysis/companies/top-rated",
			},
			"market_analysis": {
				"GET /analysis/market/overview",
			},
			"sector_analysis": {
				"GET /analysis/sectors/:sector",
			},
			"trends_analysis": {
				"GET /analysis/trends/ratings",
				"GET /analysis/trends/brokerages",
			},
			"recommendations": {
				"GET /analysis/recommendations/companies/:id",
				"GET /analysis/recommendations/rating/:rating",
			},
		},
	}
}
