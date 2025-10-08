package utils

import (
	"crypto/sha256"
	"encoding/hex"
)

func HashUrl(normUrl string) string {
	h := sha256.Sum256([]byte(normUrl))
	hash := hex.EncodeToString(h[:])
	return hash
}