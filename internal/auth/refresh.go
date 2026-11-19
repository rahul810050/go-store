package auth

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5/pgtype"
	repo "github.com/rahul810050/go-store/internal/adapters/postgresql/sqlc"
)

// creates a cryptographically secure token
func generateRefreshToken() (string, error) {
	randomBytes := make([]byte, 32)

	if _, err := rand.Read(randomBytes); err != nil {
		return "", err
	}
	token := base64.RawURLEncoding.EncodeToString(randomBytes)

	return token, nil
}

// generate a SHA-256 hash of a refresh token
func hashRefreshToken(token string) string {
	hash := sha256.Sum256([]byte(token))
	return hex.EncodeToString(hash[:])
}

// generates a UUID v4
func generateTokenFamilyID() ([16]byte, error) {
	var id [16]byte
	if _, err := rand.Read(id[:]); err != nil {
		return id, err
	}

	id[6] = (id[6] & 0x0f) | 0x40
	id[8] = (id[8] & 0x3f) | 0x80

	return id, nil
}

func newTokenFamilyID() (pgtype.UUID, error) {
	id, err := generateTokenFamilyID()
	if err != nil {
		return pgtype.UUID{}, err
	}
	return pgtype.UUID{
		Bytes: id,
		Valid: true,
	}, nil
}


func (s *service) saveRefreshToken(ctx context.Context, userID int64, familyID pgtype.UUID) (string, error) {
	// generate the token
	token, err := generateRefreshToken()
	if err != nil {
		return "", fmt.Errorf("generate refresh token: %w", err)
	}
	// hash token
	hashedToken := hashRefreshToken(token)

	expiresAt := time.Now().Add(7 * 24 * time.Hour)

	_, err = s.repo.CreateRefreshToken(ctx, repo.CreateRefreshTokenParams{
		UserID: userID,
		TokenHash: hashedToken,
		FamilyID: familyID,
		ExpiresAt: pgtype.Timestamptz{
			Time: expiresAt,
			Valid: true,
		},
	})
	if err != nil {
		return "", fmt.Errorf("save refresh token: %w", err)
	}

	// returning the original token not the hashed token stored in the db
	return token, nil
}