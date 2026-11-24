package discounts

import "time"

type DiscountType string

const (
	DiscountTypePercentage DiscountType = "percentage"
	DiscountTypeFixed      DiscountType = "fixed"
)

type CampaignStatus string

const (
	CampaignStatusDraft     CampaignStatus = "draft"
	CampaignStatusActive    CampaignStatus = "active"
	CampaignStatusCancelled CampaignStatus = "cancelled"
)

type Campaign struct {
	ID            int64          `json:"id"`
	Name          string         `json:"name"`
	Description   *string        `json:"description,omitempty"`
	DiscountType  DiscountType   `json:"discount_type"`
	DiscountValue int32          `json:"discount_value"`
	StartAt       time.Time      `json:"start_at"`
	EndAt         time.Time      `json:"end_at"`
	Status        CampaignStatus `json:"status"`
	CreatedAt     time.Time      `json:"created_at"`
	UpdatedAt     time.Time      `json:"updated_at"`
}

type CreateCampaignInput struct {
	Name          string       `json:"name"`
	Description   *string      `json:"description"`
	DiscountType  DiscountType `json:"discount_type"`
	DiscountValue int32        `json:"discount_value"`
	StartAt       time.Time    `json:"start_at"`
	EndAt         time.Time    `json:"end_at"`
}

type UpdateCampaignInput struct {
	Name          string       `json:"name"`
	Description   *string      `json:"description"`
	DiscountType  DiscountType `json:"discount_type"`
	DiscountValue int32        `json:"discount_value"`
	StartAt       time.Time    `json:"start_at"`
	EndAt         time.Time    `json:"end_at"`
}
