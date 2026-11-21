package webhooks

import (
	"context"
	"encoding/json"
	"errors"
	"log"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
	repo "github.com/rahul810050/go-store/internal/adapters/postgresql/sqlc"
)


type worker struct {
	repo repo.Querier
	db *pgxpool.Pool
}


func NewWorker(db *pgxpool.Pool) *worker {
	return &worker{
		repo: repo.New(db),
		db: db,
	}
}

func decodeWebhookEvent(payload []byte) (WebhookEvent, error) {
	var event WebhookEvent
	err := json.Unmarshal(payload, &event)
	if err != nil {
		return WebhookEvent{}, err
	}
	return event, nil
}

func validateCapturedPayment(payment PaymentEntity) error {
	if payment.ID == "" {
		return errors.New("missing razorpay payment ID")
	}

	if payment.OrderID == "" {
		return errors.New("missing razorpay order ID")
	}

	if payment.Amount <= 0 {
		return errors.New("invalid payment amount")
	}

	if payment.Currency == "" {
		return errors.New("missing payment currency")
	}

	if payment.Status != "captured" || !payment.Captured {
		return errors.New("payment is not captured")
	}

	return nil
}

func (w *worker) ClaimNextEvent(ctx context.Context) (repo.WebhookEvent, error) {
	event, err := w.repo.ClaimPendingWebhookEvent(ctx)
	if err != nil {
		return repo.WebhookEvent{}, err
	}
	return event, nil
}

func (w *worker) ProcessEvent(ctx context.Context, event WebhookEvent) error {
	switch event.Event {
	case EventPaymentCaptured:
		return w.HandlePaymentCaptured(ctx, event)

	case EventPaymentFailed:
		return errors.New("payment.failed handler not implemented")

	default:
		return errors.New("unsupported webhook event: " + event.Event)
	}
}

func (w *worker) ProcessClaimedEvent(ctx context.Context, dbEvent repo.WebhookEvent) error {
	event, err := decodeWebhookEvent(dbEvent.Payload)
	if err != nil {
		if markErr := w.MarkEventFailed(ctx, dbEvent.ID, err); markErr != nil {
			return errors.Join(err, markErr)
		}
		return err
	}

	if err := w.ProcessEvent(ctx, event); err != nil {
		if markErr := w.MarkEventFailed(ctx, dbEvent.ID, err); markErr != nil {
			return errors.Join(err, markErr)
		}
		return err
	}

	return w.MarkEventProcessed(ctx, dbEvent.ID)
}

func (w *worker) HandlePaymentCaptured(ctx context.Context, event WebhookEvent) error {
	payment := event.Payload.Payment.Entity
	if err := validateCapturedPayment(payment); err != nil {
		return err
	}
	localPayment, err := w.FindPaymentByPaymentOrderID(ctx, payment.OrderID)
	if err != nil {
		return err
	}
	// validation checks
	if localPayment.Amount != payment.Amount {
		return errors.New("payment amount mismatch")
	}

	if localPayment.Currency != payment.Currency {
		return errors.New("payment currency mismatch")
	}

	if localPayment.Status == "success" {
		return nil
	}

	if localPayment.Status != "created" && localPayment.Status != "pending" {
		return errors.New("payment is not in a valid state")
	}
	// create a transaction
	tx, err := w.db.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	qtx := repo.New(tx)

	// lock payment row
	lockedPayment, err := qtx.GetPaymentByIDForUpdate(ctx, localPayment.ID)
	if err != nil {
		return err
	}
	// validation checks
	if lockedPayment.Status == "success" {
		return nil
	}

	if lockedPayment.Status != "created" && lockedPayment.Status != "pending" {
		return errors.New("payment is not in a valid state")
	}

	if lockedPayment.Amount != payment.Amount {
		return errors.New("payment amount mismatch")
	}

	if lockedPayment.Currency != payment.Currency {
		return errors.New("payment currency mismatch")
	}
	// update payment status
	_, err = qtx.UpdatePaymentStatus(ctx, repo.UpdatePaymentStatusParams{
		ID:                 lockedPayment.ID,
		Status:             "success",
		ProviderPaymentID:  pgtype.Text{
			String: payment.ID,
			Valid:  true,
		},
	})
	if err != nil {
		return err
	}
	// update order status
	_, err = qtx.UpdateOrderStatus(ctx, repo.UpdateOrderStatusParams{
		ID:     lockedPayment.OrderID,
		Status: "confirmed",
	})
	if err != nil {
		return err
	}

	return tx.Commit(ctx)
}

func (w *worker) FindPaymentByPaymentOrderID(ctx context.Context, providerOrderID string) (repo.Payment, error) {
	payment, err := w.repo.GetPaymentByProviderOrderID(ctx, pgtype.Text{
		String: providerOrderID,
		Valid: true,
	})
	if err != nil {
		return repo.Payment{}, err
	}

	return payment, nil
}

func (w *worker) MarkEventProcessed(ctx context.Context, eventID int64) error {
	return w.repo.MarkWebhookEventAsProcessed(ctx, eventID)
}

func (w *worker) MarkEventFailed(ctx context.Context, eventID int64, err error) error {
	return w.repo.MarkWebhookEventAsFailed(ctx, repo.MarkWebhookEventAsFailedParams{
		ID:        eventID,
		LastError: pgtype.Text{
			String: err.Error(),
			Valid: true,
		},
	})
}

func (w *worker) RecoverStaleEvents(ctx context.Context) error {
	return w.repo.RecoverStaleWebhookEvents(ctx)
}

func (w *worker) RunOnce(ctx context.Context) error {
	if err := w.RecoverStaleEvents(ctx); err != nil {
		return err
	}

	const maxBatchSize = 100

	for i := 0; i < maxBatchSize; i++ {
		event, err := w.ClaimNextEvent(ctx)

		if errors.Is(err, pgx.ErrNoRows) {
			return nil // Queue is empty
		}
		if err != nil {
			return err
		}
		if err := w.ProcessClaimedEvent(ctx, event); err != nil {
			log.Printf("failed to process webhook event %d: %v", event.ID, err)
			continue
		}
	}

	return nil
}


func (w *worker) Run(ctx context.Context, interval time.Duration) {
	// process any existing events immediately on startup.
	if err := w.RunOnce(ctx); err != nil {
		log.Printf("webhook worker error: %v", err)
	}

	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			log.Println("webhook worker stopped")
			return

		case <-ticker.C:
			if err := w.RunOnce(ctx); err != nil {
				log.Printf("webhook worker error: %v", err)
			}
		}
	}
}