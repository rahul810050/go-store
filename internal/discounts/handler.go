package discounts

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	repo "github.com/rahul810050/go-store/internal/adapters/postgresql/sqlc"
	"github.com/rahul810050/go-store/internal/jsonwriter"
	"github.com/rahul810050/go-store/internal/middleware"
)

type handler struct {
	service Service
}

func NewHandler(service Service) *handler {
	return &handler{
		service: service,
	}
}

func isValidId(s string) bool {
	if len(s) <= 0 {
		return false
	}
	for i := 0; i < len(s); i++ {
		if s[i] < '0' || s[i] > '9' {
			return false
		}
	}
	return true
}

func (h *handler) CreateCampaign(w http.ResponseWriter, r *http.Request) {
	var input CreateCampaignInput

	err := json.NewDecoder(r.Body).Decode(&input)
	if err != nil {
		http.Error(w, "Invalid JSON payload", http.StatusBadRequest)
		return
	}
	defer r.Body.Close()

	campaign, err := h.service.CreateCampaign(r.Context(), input)
	if err != nil {
		writeServiceError(w, err)
		return
	}

	jsonwriter.Write(w, http.StatusCreated, campaign)
}

func (h *handler) GetCampaignByID(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")

	if !isValidId(id) {
		http.Error(w, "Invalid campaign ID", http.StatusBadRequest)
		return
	}

	campaignID, err := strconv.ParseInt(id, 10, 64)
	if err != nil || campaignID <= 0 {
		http.Error(w, "Invalid campaign ID", http.StatusBadRequest)
		return
	}

	role, ok := middleware.UserRoleFromContext(r.Context())
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	var campaign repo.DiscountCampaign

	switch role {
	case "admin":
		campaign, err = h.service.GetCampaignByID(r.Context(), campaignID)

	case "user":
		campaign, err = h.service.GetEligibleCampaignByID(r.Context(), campaignID)

	default:
		http.Error(w, "Forbidden", http.StatusForbidden)
		return
	}

	if err != nil {
		writeServiceError(w, err)
		return
	}

	jsonwriter.Write(w, http.StatusOK, campaign)
}

func (h *handler) ListCampaigns(w http.ResponseWriter, r *http.Request) {
	role, ok := middleware.UserRoleFromContext(r.Context())
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	status := r.URL.Query().Get("status")

	var campaigns []repo.DiscountCampaign
	var err error

	switch role {
	case "admin":
		if status == "" {
			campaigns, err = h.service.ListCampaigns(r.Context())
		} else {
			campaigns, err = h.service.ListCampaignsByStatus(r.Context(), status)
		}

	case "user":
		if status != "" {
			http.Error(w, "Status filtering is restricted to admins", http.StatusForbidden)
			return
		}

		campaigns, err = h.service.ListEligibleCampaigns(r.Context())

	default:
		http.Error(w, "Forbidden", http.StatusForbidden)
		return
	}

	if err != nil {
		writeServiceError(w, err)
		return
	}

	jsonwriter.Write(w, http.StatusOK, campaigns)
}

func (h *handler) UpdateCampaign(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")

	if !isValidId(id) {
		http.Error(w, "Invalid campaign ID", http.StatusBadRequest)
		return
	}

	campaignID, err := strconv.ParseInt(id, 10, 64)
	if err != nil || campaignID <= 0 {
		http.Error(w, "Invalid campaign ID", http.StatusBadRequest)
		return
	}

	var input UpdateCampaignInput
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		http.Error(w, "Invalid JSON payload", http.StatusBadRequest)
		return
	}
	defer r.Body.Close()

	campaign, err := h.service.UpdateCampaign(r.Context(), campaignID, input)
	if err != nil {
		writeServiceError(w, err)
		return
	}

	jsonwriter.Write(w, http.StatusOK, campaign)
}

func (h *handler) ActivateCampaign(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")

	if !isValidId(id) {
		http.Error(w, "Invalid campaign ID", http.StatusBadRequest)
		return
	}

	campaignID, err := strconv.ParseInt(id, 10, 64)
	if err != nil || campaignID <= 0 {
		http.Error(w, "Invalid campaign ID", http.StatusBadRequest)
		return
	}

	campaign, err := h.service.ActivateCampaign(r.Context(), campaignID)
	if err != nil {
		writeServiceError(w, err)
		return
	}

	jsonwriter.Write(w, http.StatusOK, campaign)
}

func (h *handler) CancelCampaign(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")

	if !isValidId(id) {
		http.Error(w, "Invalid campaign ID", http.StatusBadRequest)
		return
	}

	campaignID, err := strconv.ParseInt(id, 10, 64)
	if err != nil || campaignID <= 0 {
		http.Error(w, "Invalid campaign ID", http.StatusBadRequest)
		return
	}

	campaign, err := h.service.CancelCampaign(r.Context(), campaignID)
	if err != nil {
		writeServiceError(w, err)
		return
	}

	jsonwriter.Write(w, http.StatusOK, campaign)
}

func (h *handler) AddProductToCampaign(w http.ResponseWriter, r *http.Request) {
	campaignIDStr := chi.URLParam(r, "id")
	productIDStr := chi.URLParam(r, "productID")

	if !isValidId(campaignIDStr) || !isValidId(productIDStr) {
		http.Error(w, "Invalid campaign or product ID", http.StatusBadRequest)
		return
	}

	campaignID, err := strconv.ParseInt(campaignIDStr, 10, 64)
	if err != nil || campaignID <= 0 {
		http.Error(w, "Invalid campaign ID", http.StatusBadRequest)
		return
	}

	productID, err := strconv.ParseInt(productIDStr, 10, 64)
	if err != nil || productID <= 0 {
		http.Error(w, "Invalid product ID", http.StatusBadRequest)
		return
	}

	err = h.service.AddProductToCampaign(r.Context(), campaignID, productID)
	if err != nil {
		writeServiceError(w, err)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (h *handler) RemoveProductFromCampaign(w http.ResponseWriter, r *http.Request) {
	campaignIDStr := chi.URLParam(r, "id")
	productIDStr := chi.URLParam(r, "productID")

	if !isValidId(campaignIDStr) || !isValidId(productIDStr) {
		http.Error(w, "Invalid campaign or product ID", http.StatusBadRequest)
		return
	}

	campaignID, err := strconv.ParseInt(campaignIDStr, 10, 64)
	if err != nil || campaignID <= 0 {
		http.Error(w, "Invalid campaign ID", http.StatusBadRequest)
		return
	}

	productID, err := strconv.ParseInt(productIDStr, 10, 64)
	if err != nil || productID <= 0 {
		http.Error(w, "Invalid product ID", http.StatusBadRequest)
		return
	}

	err = h.service.RemoveProductFromCampaign(r.Context(), campaignID, productID)
	if err != nil {
		writeServiceError(w, err)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (h *handler) ListProductsByCampaign(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")

	if !isValidId(id) {
		http.Error(w, "Invalid campaign ID", http.StatusBadRequest)
		return
	}

	campaignID, err := strconv.ParseInt(id, 10, 64)
	if err != nil || campaignID <= 0 {
		http.Error(w, "Invalid campaign ID", http.StatusBadRequest)
		return
	}

	role, ok := middleware.UserRoleFromContext(r.Context())
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	var products []repo.Product

	switch role {
	case "admin":
		products, err = h.service.ListProductsByCampaign(r.Context(), campaignID)

	case "user":
		products, err = h.service.ListProductsByEligibleCampaign(r.Context(), campaignID)

	default:
		http.Error(w, "Forbidden", http.StatusForbidden)
		return
	}

	if err != nil {
		writeServiceError(w, err)
		return
	}

	jsonwriter.Write(w, http.StatusOK, products)
}
