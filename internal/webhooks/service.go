package webhooks

import (
	"context"

	"github.com/jackc/pgx/v5/pgxpool"
	repo "github.com/rahul810050/go-store/internal/adapters/postgresql/sqlc"
)

type Service interface {
	SaveWebhookEvent(ctx context.Context, eventID string, eventType string, payload []byte) error
}

type service struct {
	repo repo.Querier
	db *pgxpool.Pool
}


func NewService(db *pgxpool.Pool) Service {
	return &service{
		repo: repo.New(db),
		db: db,
	}
}


func (s *service) SaveWebhookEvent(ctx context.Context, eventID string, eventType string, payload []byte) error {
	return s.repo.CreateWebhookEvent(ctx, repo.CreateWebhookEventParams{
		EventID: eventID,
		EventType: eventType,
		Payload: payload,
	})
}