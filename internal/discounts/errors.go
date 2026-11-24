package discounts

import (
	"errors"
	"net/http"
)

var (
	ErrCampaignNotFound     = errors.New("discount campaign not found")
	ErrInvalidCampaign      = errors.New("invalid discount campaign")
	ErrInvalidDiscountType  = errors.New("invalid discount type")
	ErrInvalidDiscountValue = errors.New("invalid discount value")
	ErrInvalidCampaignDates = errors.New("invalid campaign dates")
	ErrInvalidCampaignState = errors.New("invalid campaign state")
	ErrInvalidCampaignID    = errors.New("invalid campaign ID")
	ErrInvalidProductID     = errors.New("invalid product ID")
	ErrCampaignNoProducts = errors.New("campaign must contain at least one product")
)

func writeServiceError(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, ErrCampaignNotFound):
		http.Error(w, "Campaign not found", http.StatusNotFound)

	case errors.Is(err, ErrInvalidCampaignID),
		errors.Is(err, ErrInvalidProductID),
		errors.Is(err, ErrInvalidCampaign),
		errors.Is(err, ErrInvalidDiscountType),
		errors.Is(err, ErrInvalidDiscountValue),
		errors.Is(err, ErrInvalidCampaignDates),
		errors.Is(err, ErrInvalidCampaignState),
		errors.Is(err, ErrCampaignNoProducts):
		http.Error(w, err.Error(), http.StatusBadRequest)

	default:
		http.Error(w, "Internal server error", http.StatusInternalServerError)
	}
}
