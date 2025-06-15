package services

import (
	"context"
	"strings"

	"github.com/google/uuid"

	"github.com/MayaCris/stock-info-app/internal/application/dto/request"
	"github.com/MayaCris/stock-info-app/internal/application/dto/response"
	"github.com/MayaCris/stock-info-app/internal/application/services/interfaces"
	"github.com/MayaCris/stock-info-app/internal/domain/entities"
	repoInterfaces "github.com/MayaCris/stock-info-app/internal/domain/repositories/interfaces"
	"github.com/MayaCris/stock-info-app/internal/infrastructure/logger"
)

// brokerageService implements the BrokerageService interface
type brokerageService struct {
	brokerageRepo repoInterfaces.BrokerageRepository
	logger        logger.Logger
}

// NewBrokerageService creates a new brokerage service
func NewBrokerageService(
	brokerageRepo repoInterfaces.BrokerageRepository,
	logger logger.Logger,
) interfaces.BrokerageService {
	return &brokerageService{
		brokerageRepo: brokerageRepo,
		logger:        logger,
	}
}

// CreateBrokerage creates a new brokerage
func (s *brokerageService) CreateBrokerage(ctx context.Context, req *request.CreateBrokerageRequest) (*response.BrokerageResponse, error) { // Check if brokerage already exists by name
	exists, err := s.brokerageRepo.Exists(ctx, req.Name)
	if err != nil {
		s.logger.Error(ctx, "Failed to check brokerage existence", err,
			logger.String("name", req.Name))
		return nil, response.InternalServerError("Failed to check brokerage existence")
	}

	if exists {
		return nil, response.Conflict("Brokerage with name already exists")
	}

	// Create brokerage entity
	brokerage := &entities.Brokerage{
		ID:       uuid.New(),
		Name:     strings.TrimSpace(req.Name),
		Website:  strings.TrimSpace(req.Website),
		IsActive: true,
	}

	// Save to repository
	if err := s.brokerageRepo.Create(ctx, brokerage); err != nil {
		s.logger.Error(ctx, "Failed to create brokerage", err,
			logger.String("name", req.Name))
		return nil, response.InternalServerError("Failed to create brokerage")
	}

	s.logger.Info(ctx, "Brokerage created successfully",
		logger.String("brokerage_id", brokerage.ID.String()),
		logger.String("name", brokerage.Name))

	return s.convertToBrokerageResponse(brokerage), nil
}

// GetBrokerageByID retrieves a brokerage by ID
func (s *brokerageService) GetBrokerageByID(ctx context.Context, id uuid.UUID) (*response.BrokerageResponse, error) {
	brokerage, err := s.brokerageRepo.GetByID(ctx, id)
	if err != nil {
		s.logger.Error(ctx, "Failed to get brokerage by ID", err,
			logger.String("brokerage_id", id.String()))
		return nil, response.NotFound("Brokerage")
	}

	return s.convertToBrokerageResponse(brokerage), nil
}

// UpdateBrokerage updates an existing brokerage
func (s *brokerageService) UpdateBrokerage(ctx context.Context, id uuid.UUID, req *request.UpdateBrokerageRequest) (*response.BrokerageResponse, error) {
	// Get existing brokerage
	brokerage, err := s.brokerageRepo.GetByID(ctx, id)
	if err != nil {
		return nil, response.NotFound("Brokerage")
	}

	// Update fields if provided
	if req.Name != nil {
		brokerage.Name = strings.TrimSpace(*req.Name)
	}
	if req.Website != nil {
		brokerage.Website = strings.TrimSpace(*req.Website)
	}
	if req.IsActive != nil {
		brokerage.IsActive = *req.IsActive
	}

	// Save changes
	if err := s.brokerageRepo.Update(ctx, brokerage); err != nil {
		s.logger.Error(ctx, "Failed to update brokerage", err,
			logger.String("brokerage_id", id.String()))
		return nil, response.InternalServerError("Failed to update brokerage")
	}

	s.logger.Info(ctx, "Brokerage updated successfully",
		logger.String("brokerage_id", brokerage.ID.String()),
		logger.String("name", brokerage.Name))

	return s.convertToBrokerageResponse(brokerage), nil
}

// DeleteBrokerage deletes a brokerage
func (s *brokerageService) DeleteBrokerage(ctx context.Context, id uuid.UUID) error {
	// Check if exists
	_, err := s.brokerageRepo.GetByID(ctx, id)
	if err != nil {
		return response.NotFound("Brokerage")
	}

	if err := s.brokerageRepo.Delete(ctx, id); err != nil {
		s.logger.Error(ctx, "Failed to delete brokerage", err,
			logger.String("brokerage_id", id.String()))
		return response.InternalServerError("Failed to delete brokerage")
	}

	s.logger.Info(ctx, "Brokerage deleted successfully",
		logger.String("brokerage_id", id.String()))
	return nil
}

// ListBrokerages lists brokerages with filters and pagination
func (s *brokerageService) ListBrokerages(ctx context.Context, filter *request.BrokerageFilterRequest, pagination *response.PaginationRequest) (*response.PaginatedResponse[*response.BrokerageResponse], error) {
	// Validate pagination
	if err := pagination.Validate(); err != nil {
		return nil, response.BadRequest("Invalid pagination parameters")
	}

	var brokerages []*entities.Brokerage
	var total int64
	var err error

	// Apply filters
	if filter != nil {
		if filter.IsActive != nil && *filter.IsActive {
			brokerages, err = s.brokerageRepo.GetAllActive(ctx)
		} else {
			brokerages, err = s.brokerageRepo.GetAll(ctx)
		}
	} else {
		brokerages, err = s.brokerageRepo.GetAll(ctx)
	}

	if err != nil {
		s.logger.Error(ctx, "Failed to get brokerages", err)
		return nil, response.InternalServerError("Failed to get brokerages")
	}

	// Apply name filter if provided
	if filter != nil && filter.Name != "" {
		nameFilter := strings.ToLower(filter.Name)
		var filteredBrokerages []*entities.Brokerage
		for _, brokerage := range brokerages {
			if strings.Contains(strings.ToLower(brokerage.Name), nameFilter) {
				filteredBrokerages = append(filteredBrokerages, brokerage)
			}
		}
		brokerages = filteredBrokerages
	}

	total = int64(len(brokerages))

	// Apply pagination manually (in production, implement pagination in repository)
	start := pagination.GetOffset()
	end := start + pagination.GetLimit()
	if start > len(brokerages) {
		start = len(brokerages)
	}
	if end > len(brokerages) {
		end = len(brokerages)
	}
	paginatedBrokerages := brokerages[start:end]

	// Convert to responses
	responses := make([]*response.BrokerageResponse, len(paginatedBrokerages))
	for i, brokerage := range paginatedBrokerages {
		responses[i] = s.convertToBrokerageResponse(brokerage)
	}

	return response.NewPaginatedResponse(responses, pagination.Page, pagination.PerPage, int(total)), nil
}

// ListActiveBrokerages lists only active brokerages
func (s *brokerageService) ListActiveBrokerages(ctx context.Context, pagination *response.PaginationRequest) (*response.PaginatedResponse[*response.BrokerageResponse], error) {
	// Validate pagination
	if err := pagination.Validate(); err != nil {
		return nil, response.BadRequest("Invalid pagination parameters")
	}

	brokerages, err := s.brokerageRepo.GetAllActive(ctx)
	if err != nil {
		s.logger.Error(ctx, "Failed to get active brokerages", err)
		return nil, response.InternalServerError("Failed to get brokerages")
	}

	total := len(brokerages)
	start := pagination.GetOffset()
	end := start + pagination.GetLimit()
	if start > total {
		start = total
	}
	if end > total {
		end = total
	}
	paginatedBrokerages := brokerages[start:end]

	// Convert to responses
	responses := make([]*response.BrokerageResponse, len(paginatedBrokerages))
	for i, brokerage := range paginatedBrokerages {
		responses[i] = s.convertToBrokerageResponse(brokerage)
	}

	return response.NewPaginatedResponse(responses, pagination.Page, pagination.PerPage, total), nil
}

// ActivateBrokerage activates a brokerage
func (s *brokerageService) ActivateBrokerage(ctx context.Context, id uuid.UUID) error {
	if err := s.brokerageRepo.Activate(ctx, id); err != nil {
		s.logger.Error(ctx, "Failed to activate brokerage", err,
			logger.String("brokerage_id", id.String()))
		return response.InternalServerError("Failed to activate brokerage")
	}

	s.logger.Info(ctx, "Brokerage activated successfully",
		logger.String("brokerage_id", id.String()))
	return nil
}

// DeactivateBrokerage deactivates a brokerage
func (s *brokerageService) DeactivateBrokerage(ctx context.Context, id uuid.UUID) error {
	if err := s.brokerageRepo.Deactivate(ctx, id); err != nil {
		s.logger.Error(ctx, "Failed to deactivate brokerage", err,
			logger.String("brokerage_id", id.String()))
		return response.InternalServerError("Failed to deactivate brokerage")
	}

	s.logger.Info(ctx, "Brokerage deactivated successfully",
		logger.String("brokerage_id", id.String()))
	return nil
}

// SearchBrokeragesByName searches brokerages by name
func (s *brokerageService) SearchBrokeragesByName(ctx context.Context, name string, pagination *response.PaginationRequest) (*response.PaginatedResponse[*response.BrokerageResponse], error) {
	// Validate pagination
	if err := pagination.Validate(); err != nil {
		return nil, response.BadRequest("Invalid pagination parameters")
	}

	// Get all brokerages and filter manually
	brokerages, err := s.brokerageRepo.GetAll(ctx)
	if err != nil {
		s.logger.Error(ctx, "Failed to search brokerages", err)
		return nil, response.InternalServerError("Failed to search brokerages")
	}

	// Simple text search
	query := strings.ToLower(name)
	var filteredBrokerages []*entities.Brokerage
	for _, brokerage := range brokerages {
		if strings.Contains(strings.ToLower(brokerage.Name), query) {
			filteredBrokerages = append(filteredBrokerages, brokerage)
		}
	}

	// Apply pagination
	total := len(filteredBrokerages)
	start := pagination.GetOffset()
	end := start + pagination.GetLimit()
	if start > total {
		start = total
	}
	if end > total {
		end = total
	}
	paginatedBrokerages := filteredBrokerages[start:end]

	// Convert to responses
	responses := make([]*response.BrokerageResponse, len(paginatedBrokerages))
	for i, brokerage := range paginatedBrokerages {
		responses[i] = s.convertToBrokerageResponse(brokerage)
	}

	return response.NewPaginatedResponse(responses, pagination.Page, pagination.PerPage, total), nil
}

// Helper methods

func (s *brokerageService) convertToBrokerageResponse(brokerage *entities.Brokerage) *response.BrokerageResponse {
	return &response.BrokerageResponse{
		ID:        brokerage.ID,
		Name:      brokerage.Name,
		Website:   brokerage.Website,
		IsActive:  brokerage.IsActive,
		CreatedAt: brokerage.CreatedAt,
		UpdatedAt: brokerage.UpdatedAt,
	}
}
