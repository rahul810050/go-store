package auth

import (
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"net/mail"
	"strings"

	"github.com/rahul810050/go-store/internal/jsonwriter"
)

type handler struct {
	service Service
}

type SignupInput struct {
	Name         string `json:"name"`
	Email        string `json:"email"`
	Password string `json:"password"`
}

type SigninInput struct {
	Email        string `json:"email"`
	Password string `json:"password"`
}

type RefreshInput struct {
	RefreshToken string `json:"refresh_token"`
}

func NewHandler(service Service) *handler {
	return &handler{
		service: service,
	}
}

// signup
func (h *handler) Signup(w http.ResponseWriter, r *http.Request) {
	var input SignupInput

	r.Body = http.MaxBytesReader(w, r.Body, 1<<20)

	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()
	defer r.Body.Close()

	if err := decoder.Decode(&input); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	input.Name = strings.TrimSpace(input.Name)
	input.Email = strings.TrimSpace(input.Email)
	// input empty check
	if input.Name == "" || input.Email == "" || input.Password == "" {
		http.Error(w, "Name, email and password are required", http.StatusBadRequest)
		return
	}
	// size check
	if len(input.Name) > 255 || len(input.Email) > 254 {
		http.Error(w, "Name or email is too long", http.StatusBadRequest)
		return
	}
	// email pattern check
	email, err := mail.ParseAddress(input.Email)
	if err != nil || email.Address != input.Email {
		http.Error(w, "Invalid email address", http.StatusBadRequest)
        return
	}
	// password check
	if len(input.Password) < 6 || len(input.Password) > 128 {
        http.Error(w, "Password must be between 12 and 128 bytes", http.StatusBadRequest)
        return
    }
	// call signup service
	user, err := h.service.Signup(r.Context(), SignupInput{
		Name: input.Name,
		Email: input.Email,
		Password: input.Password,
	})
	if err != nil {
		if errors.Is(err, ErrEmailAlreadyExists) {
			http.Error(w, "Email already registered", http.StatusConflict)
			return
		}
		log.Println("signup failed", err)
		http.Error(w, "Something went wrong", http.StatusInternalServerError)
		return
	}
	jsonwriter.Write(w, http.StatusCreated, user)
}

// signin
func (h *handler) Signin(w http.ResponseWriter, r *http.Request) {
	var input SigninInput

	r.Body = http.MaxBytesReader(w, r.Body, 1<<20)
	defer r.Body.Close()

	decoded := json.NewDecoder(r.Body)
	decoded.DisallowUnknownFields()

	
	if err := decoded.Decode(&input); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}
	input.Email = strings.TrimSpace(input.Email)

	if input.Email == "" || input.Password == "" {
		http.Error(w, "Email and password are required", http.StatusBadRequest)
		return
	}

	result, err := h.service.Signin(r.Context(), input)
	if err != nil {
		if errors.Is(err, ErrInvalidCredentials) {
			http.Error(w, "Invalid email or password", http.StatusUnauthorized)
			return
		}
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	res := map[string]any{
		"message": "signin successful",
		"user": result.User,
		"access_token": result.AccessToken,
		"refresh_token": result.RefreshToken,
	}
	jsonwriter.Write(w, http.StatusOK, res)
}

// refresh token
func (h *handler) Refresh(w http.ResponseWriter, r *http.Request) {
	var input RefreshInput
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}
	defer r.Body.Close()

	if input.RefreshToken == "" {
		http.Error(w, "refresh token is required", http.StatusBadRequest)
		return
	}
	result, err := h.service.Refresh(r.Context(), input)
	if err != nil {
		http.Error(w, "Invalid or expired refresh token", http.StatusUnauthorized)
		return
	}
	jsonwriter.Write(w, http.StatusOK, result)
}

// logout
func (h *handler) Logout(w http.ResponseWriter, r *http.Request) {
	var input RefreshInput
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if input.RefreshToken == "" {
		http.Error(w, "Refresh token is required", http.StatusBadRequest)
		return
	}

	err := h.service.Logout(r.Context(), input)
	if err != nil {
		http.Error(w, "Invalid or expired refresh token", http.StatusUnauthorized)
		return
	}
	jsonwriter.Write(w, http.StatusOK, map[string]string{
		"message": "Logout successful",
	})
}