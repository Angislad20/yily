package crypto

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"errors"
	"io"
)

const (
	sessionTokenBytes = 32
	patTokenBytes     = 32
	PATPrefix         = "yily_"
)

var (
	ErrEmptyToken = errors.New("token cannot be empty")
)

// GenerateSessionToken generates a random 256-bit token and its SHA-256 hex hash.
func GenerateSessionToken() (rawToken string, tokenHash string, err error) {
	buf := make([]byte, sessionTokenBytes)
	if _, err := io.ReadFull(rand.Reader, buf); err != nil {
		return "", "", err
	}

	rawToken = base64.RawURLEncoding.EncodeToString(buf)
	tokenHash = HashToken(rawToken)
	return rawToken, tokenHash, nil
}

// GeneratePAT generates a random 256-bit PAT with the "yily_" prefix and its SHA-256 hex hash.
func GeneratePAT() (rawToken string, tokenHash string, err error) {
	buf := make([]byte, patTokenBytes)
	if _, err := io.ReadFull(rand.Reader, buf); err != nil {
		return "", "", err
	}

	rawToken = PATPrefix + base64.RawURLEncoding.EncodeToString(buf)
	tokenHash = HashToken(rawToken)
	return rawToken, tokenHash, nil
}

// HashToken returns the SHA-256 hex digest of a raw token.
func HashToken(rawToken string) string {
	sum := sha256.Sum256([]byte(rawToken))
	return hex.EncodeToString(sum[:])
}
