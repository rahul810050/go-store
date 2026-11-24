package discounts

import (
	"context"
	"errors"
	"strings"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
	repo "github.com/rahul810050/go-store/internal/adapters/postgresql/sqlc"
)

type Service interface {
	CreateCampaign(ctx context.Context, input CreateCampaignInput) (repo.DiscountCampaign, error)
	GetCampaignByID(ctx context.Context, id int64) (repo.DiscountCampaign, error)
	ListCampaigns(ctx context.Context) ([]repo.DiscountCampaign, error)
	ListCampaignsByStatus(ctx context.Context, status string) ([]repo.DiscountCampaign, error)
	UpdateCampaign(ctx context.Context, id int64, input UpdateCampaignInput) (repo.DiscountCampaign, error)
	ActivateCampaign(ctx context.Context, id int64) (repo.DiscountCampaign, error)
	CancelCampaign(ctx context.Context, id int64) (repo.DiscountCampaign, error)
	AddProductToCampaign(ctx context.Context, campaignID int64, productID int64) error
	RemoveProductFromCampaign(ctx context.Context, campaignID int64, productID int64) error
	ListProductsByCampaign(ctx context.Context, campaignID int64) ([]repo.Product, error)
	ListEligibleCampaigns(ctx context.Context) ([]repo.DiscountCampaign, error)
	GetEligibleCampaignByID(ctx context.Context, id int64) (repo.DiscountCampaign, error)
	ListProductsByEligibleCampaign(ctx context.Context, campaignID int64) ([]repo.Product, error)
}

type service struct {
	repo repo.Querier
	db   *pgxpool.Pool
}

func NewService(db *pgxpool.Pool) Service {
	return &service{
		repo: repo.New(db),
		db:   db,
	}
}

func (s *service) handleCampaignMutationError(ctx context.Context, id int64, err error) error {
	if !errors.Is(err, pgx.ErrNoRows) {
		return err
	}

	_, lookupErr := s.repo.GetDiscountCampaignByID(ctx, id)
	if lookupErr != nil {
		if errors.Is(lookupErr, pgx.ErrNoRows) {
			return ErrCampaignNotFound
		}
		return lookupErr
	}

	return ErrInvalidCampaignState
}

func (s *service) CreateCampaign(ctx context.Context, input CreateCampaignInput) (repo.DiscountCampaign, error) {
	input.Name = strings.TrimSpace(input.Name)
	if input.Name == "" {
		return repo.DiscountCampaign{}, ErrInvalidCampaign
	}

	if input.DiscountType != "percentage" && input.DiscountType != "fixed" {
		return repo.DiscountCampaign{}, ErrInvalidDiscountType
	}

	if input.DiscountValue <= 0 {
		return repo.DiscountCampaign{}, ErrInvalidDiscountValue
	}

	if input.DiscountType == "percentage" && input.DiscountValue > 100 {
		return repo.DiscountCampaign{}, ErrInvalidDiscountValue
	}

	if !input.EndAt.After(input.StartAt) {
		return repo.DiscountCampaign{}, ErrInvalidCampaignDates
	}

	var description pgtype.Text
	if input.Description != nil {
		trimmed := strings.TrimSpace(*input.Description)
		if trimmed != "" {
			description = pgtype.Text{
				String: trimmed,
				Valid:  true,
			}
		}
	}

	campaign, err := s.repo.CreateDiscountCampaign(ctx, repo.CreateDiscountCampaignParams{
		Name:          input.Name,
		Description:   description,
		DiscountType:  string(input.DiscountType),
		DiscountValue: input.DiscountValue,
		StartAt: pgtype.Timestamptz{
			Time:  input.StartAt,
			Valid: true,
		},
		EndAt: pgtype.Timestamptz{
			Time:  input.EndAt,
			Valid: true,
		},
	})
	if err != nil {
		return repo.DiscountCampaign{}, err
	}

	return campaign, nil
}

func (s *service) GetCampaignByID(ctx context.Context, id int64) (repo.DiscountCampaign, error) {
	if id <= 0 {
		return repo.DiscountCampaign{}, ErrInvalidCampaignID
	}

	campaign, err := s.repo.GetDiscountCampaignByID(ctx, id)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return repo.DiscountCampaign{}, ErrCampaignNotFound
		}
		return repo.DiscountCampaign{}, err
	}

	return campaign, nil
}

func (s *service) ListCampaigns(ctx context.Context) ([]repo.DiscountCampaign, error) {
	campaigns, err := s.repo.ListDiscountCampaigns(ctx)
	if err != nil {
		return []repo.DiscountCampaign{}, err
	}

	return campaigns, nil
}

func (s *service) ListCampaignsByStatus(ctx context.Context, status string) ([]repo.DiscountCampaign, error) {
	status = strings.TrimSpace(status)

	if status != "draft" && status != "active" && status != "cancelled" {
		return []repo.DiscountCampaign{}, ErrInvalidCampaignState
	}

	campaigns, err := s.repo.ListDiscountCampaignsByStatus(ctx, status)
	if err != nil {
		return []repo.DiscountCampaign{}, err
	}

	return campaigns, nil
}

func (s *service) UpdateCampaign(ctx context.Context, id int64, input UpdateCampaignInput) (repo.DiscountCampaign, error) {
	if id <= 0 {
		return repo.DiscountCampaign{}, ErrInvalidCampaignID
	}

	input.Name = strings.TrimSpace(input.Name)
	if input.Name == "" {
		return repo.DiscountCampaign{}, ErrInvalidCampaign
	}

	if input.DiscountType != "percentage" && input.DiscountType != "fixed" {
		return repo.DiscountCampaign{}, ErrInvalidDiscountType
	}

	if input.DiscountValue <= 0 {
		return repo.DiscountCampaign{}, ErrInvalidDiscountValue
	}

	if input.DiscountType == "percentage" && input.DiscountValue > 100 {
		return repo.DiscountCampaign{}, ErrInvalidDiscountValue
	}

	if !input.EndAt.After(input.StartAt) {
		return repo.DiscountCampaign{}, ErrInvalidCampaignDates
	}

	var description pgtype.Text
	if input.Description != nil {
		trimmed := strings.TrimSpace(*input.Description)
		if trimmed != "" {
			description = pgtype.Text{
				String: trimmed,
				Valid:  true,
			}
		}
	}

	campaign, err := s.repo.UpdateDiscountCampaign(ctx, repo.UpdateDiscountCampaignParams{
		ID:            id,
		Name:          input.Name,
		Description:   description,
		DiscountType:  string(input.DiscountType),
		DiscountValue: input.DiscountValue,
		StartAt: pgtype.Timestamptz{
			Time:  input.StartAt,
			Valid: true,
		},
		EndAt: pgtype.Timestamptz{
			Time:  input.EndAt,
			Valid: true,
		},
	})
	if err != nil {
		return repo.DiscountCampaign{}, s.handleCampaignMutationError(ctx, id, err)
	}

	return campaign, nil
}

func (s *service) ActivateCampaign(ctx context.Context, id int64) (repo.DiscountCampaign, error) {
	if id <= 0 {
		return repo.DiscountCampaign{}, ErrInvalidCampaignID
	}

	tx, err := s.db.Begin(ctx)
	if err != nil {
		return repo.DiscountCampaign{}, err
	}
	defer tx.Rollback(ctx)

	txRepo := repo.New(tx)

	campaign, err := txRepo.GetDiscountCampaignForUpdate(ctx, id)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return repo.DiscountCampaign{}, ErrCampaignNotFound
		}
		return repo.DiscountCampaign{}, err
	}

	if campaign.Status != "draft" {
		return repo.DiscountCampaign{}, ErrInvalidCampaignState
	}

	productCount, err := txRepo.CountProductsInDiscountCampaign(ctx, id)
	if err != nil {
		return repo.DiscountCampaign{}, err
	}

	if productCount == 0 {
		return repo.DiscountCampaign{}, ErrCampaignNoProducts
	}

	campaign, err = txRepo.ActivateDiscountCampaign(ctx, id)
	if err != nil {
		return repo.DiscountCampaign{}, err
	}

	if err = tx.Commit(ctx); err != nil {
		return repo.DiscountCampaign{}, err
	}

	return campaign, nil
}

func (s *service) CancelCampaign(ctx context.Context, id int64) (repo.DiscountCampaign, error) {
	if id <= 0 {
		return repo.DiscountCampaign{}, ErrInvalidCampaignID
	}

	campaign, err := s.repo.CancelDiscountCampaign(ctx, id)
	if err != nil {
		return repo.DiscountCampaign{}, s.handleCampaignMutationError(ctx, id, err)
	}

	return campaign, nil
}

func (s *service) AddProductToCampaign(ctx context.Context, campaignID int64, productID int64) error {
	if campaignID <= 0 {
		return ErrInvalidCampaignID
	}

	if productID <= 0 {
		return ErrInvalidProductID
	}

	tx, err := s.db.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	txRepo := repo.New(tx)

	campaign, err := txRepo.GetDiscountCampaignForUpdate(ctx, campaignID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return ErrCampaignNotFound
		}
		return err
	}

	if campaign.Status != "draft" {
		return ErrInvalidCampaignState
	}

	err = txRepo.AddProductToDiscountCampaign(ctx, repo.AddProductToDiscountCampaignParams{
		CampaignID: campaignID,
		ProductID:  productID,
	})
	if err != nil {
		return err
	}

	if err = tx.Commit(ctx); err != nil {
		return err
	}

	return nil
}

func (s *service) RemoveProductFromCampaign(ctx context.Context, campaignID int64, productID int64) error {
	if campaignID <= 0 {
		return ErrInvalidCampaignID
	}

	if productID <= 0 {
		return ErrInvalidProductID
	}

	tx, err := s.db.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	txRepo := repo.New(tx)

	campaign, err := txRepo.GetDiscountCampaignForUpdate(ctx, campaignID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return ErrCampaignNotFound
		}
		return err
	}

	if campaign.Status != "draft" {
		return ErrInvalidCampaignState
	}

	_, err = txRepo.RemoveProductFromDiscountCampaign(
		ctx,
		repo.RemoveProductFromDiscountCampaignParams{
			CampaignID: campaignID,
			ProductID:  productID,
		},
	)
	if err != nil {
		return err
	}

	if err = tx.Commit(ctx); err != nil {
		return err
	}

	return nil
}

func (s *service) ListProductsByCampaign(ctx context.Context, campaignID int64) ([]repo.Product, error) {
	if campaignID <= 0 {
		return []repo.Product{}, ErrInvalidCampaignID
	}

	_, err := s.repo.GetDiscountCampaignByID(ctx, campaignID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return []repo.Product{}, ErrCampaignNotFound
		}
		return []repo.Product{}, err
	}

	products, err := s.repo.ListProductsByDiscountCampaign(ctx, campaignID)
	if err != nil {
		return []repo.Product{}, err
	}

	return products, nil
}

func (s *service) ListEligibleCampaigns(ctx context.Context) ([]repo.DiscountCampaign, error) {
	campaigns, err := s.repo.ListEligibleDiscountCampaigns(ctx)
	if err != nil {
		return []repo.DiscountCampaign{}, err
	}

	return campaigns, nil
}

func (s *service) GetEligibleCampaignByID(ctx context.Context, id int64) (repo.DiscountCampaign, error) {
	if id <= 0 {
		return repo.DiscountCampaign{}, ErrInvalidCampaignID
	}

	campaign, err := s.repo.GetEligibleDiscountCampaignByID(ctx, id)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return repo.DiscountCampaign{}, ErrCampaignNotFound
		}
		return repo.DiscountCampaign{}, err
	}

	return campaign, nil
}

func (s *service) ListProductsByEligibleCampaign(ctx context.Context, campaignID int64) ([]repo.Product, error) {
	if campaignID <= 0 {
		return []repo.Product{}, ErrInvalidCampaignID
	}

	// Verify that the campaign is currently eligible.
	_, err := s.repo.GetEligibleDiscountCampaignByID(ctx, campaignID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return []repo.Product{}, ErrCampaignNotFound
		}
		return []repo.Product{}, err
	}

	products, err := s.repo.ListProductsByDiscountCampaign(ctx, campaignID)
	if err != nil {
		return []repo.Product{}, err
	}

	return products, nil
}
