package auth

import (
	"fmt"
	"os"
	"strconv"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

type Claims struct {
	Role string `json:"role"`
	jwt.RegisteredClaims
}

func generateAccessToken(userID int64, role string) (string, error) {
	secret := os.Getenv("JWT_SECRET")
	if len(secret) < 32 {
		return "", fmt.Errorf("JWT_SECRET must be atleast 32 bytes")
	}
	now := time.Now()
	claims := Claims{
		Role: role,
		RegisteredClaims: jwt.RegisteredClaims{
			Subject: strconv.FormatInt(userID, 10),
			Issuer: "ecom-api",
			IssuedAt: jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(now.Add(15 * time.Minute)),
		},
	}
	// creates jwt object with the algorithm and the info that we want to store
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	// token is signed with the secret 
	return token.SignedString([]byte(secret))
}