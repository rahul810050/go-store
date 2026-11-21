package webhooks

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
)

func verifyWebhookSignature(body []byte, signatue string, secret string) bool {
	if len(body) == 0 || signatue == "" || secret == "" {
		return false
	}

	signatureBytes, err := hex.DecodeString(signatue)
	if err != nil {
		return false
	}

	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write(body)

	expectedSignature := mac.Sum(nil)

	return hmac.Equal(expectedSignature, signatureBytes)
}