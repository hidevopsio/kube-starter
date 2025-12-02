// Package jwt provides JWT manipulations.
// See https://tools.ietf.org/html/rfc7519#section-4.1.3
package oidc

import (
	"bytes"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"k8s.io/apimachinery/pkg/util/validation"
	"strings"
	"time"
)

func IsNameValidate(name string) bool {
	errs := validation.IsDNS1123Label(name)
	return len(errs) == 0
}

func FormatName(name string) string {
	// Handle empty or whitespace-only input
	name = strings.TrimSpace(name)
	if len(name) == 0 {
		// Return a default name for empty input (should not happen in normal flow)
		return "user-unknown"
	}

	// If the name is already DNS-1123 compliant, return as-is
	if IsNameValidate(name) {
		return name
	}

	// Generate a consistent 4-character suffix from hash of ORIGINAL name (before lowercasing)
	// This prevents collisions when different usernames format to the same string
	originalHash := sha256.Sum256([]byte(name))
	suffix := hex.EncodeToString(originalHash[:])[:4]

	// Convert to lowercase
	lowered := strings.ToLower(name)

	// Replace non-DNS-1123 characters with '-'
	result := make([]byte, len(lowered))
	for i := range lowered {
		if (lowered[i] >= '0' && lowered[i] <= '9') || (lowered[i] <= 'z' && lowered[i] >= 'a') {
			result[i] = lowered[i]
			continue
		}
		result[i] = '-'
	}

	formatted := string(result)

	// Trim leading/trailing hyphens (DNS-1123 requirement)
	formatted = strings.Trim(formatted, "-")

	// If after trimming we have an empty string, use a default prefix
	if len(formatted) == 0 {
		formatted = "user"
	}

	return formatted + "-" + suffix
}

// DecodeWithoutVerify decodes the JWT string and returns the claims.
// Note that this method does not verify the signature and always trust it.
func DecodeWithoutVerify(s string) (c *Claims, err error) {
	payload, err := DecodePayloadAsRawJSON(s)
	if err != nil {
		return nil, fmt.Errorf("could not decode the payload: %w", err)
	}
	var claims struct {
		Issuer    string `json:"iss,omitempty"`
		Subject   string `json:"sub,omitempty"`
		Name      string `json:"name,omitempty"`
		Username  string `json:"preferred_username,omitempty"`
		Email     string `json:"email,omitempty"`
		ExpiresAt int64  `json:"exp,omitempty"`
	}
	if err := json.NewDecoder(bytes.NewReader(payload)).Decode(&claims); err != nil {
		return nil, fmt.Errorf("could not decode the json of token: %w", err)
	}

	var prettyJson bytes.Buffer
	if err := json.Indent(&prettyJson, payload, "", "  "); err != nil {
		return nil, fmt.Errorf("could not indent the json of token: %w", err)
	}

	// Fill username as the value of name if it is empty (do this BEFORE formatting)
	username := claims.Username
	if username == "" {
		username = claims.Name
	}

	cls := &Claims{
		Issuer:            claims.Issuer,
		Subject:           claims.Subject,
		Name:              claims.Name,
		Username:          username,
		FormattedUsername: FormatName(username),
		Email:             claims.Email,
		Expiry:            time.Unix(claims.ExpiresAt, 0),
		Pretty:            prettyJson.String(),
	}

	return cls, nil
}

// DecodePayloadAsPrettyJSON decodes the JWT string and returns the pretty JSON string.
func DecodePayloadAsPrettyJSON(s string) (string, error) {
	payload, err := DecodePayloadAsRawJSON(s)
	if err != nil {
		return "", fmt.Errorf("could not decode the payload: %w", err)
	}
	var prettyJson bytes.Buffer
	if err := json.Indent(&prettyJson, payload, "", "  "); err != nil {
		return "", fmt.Errorf("could not indent the json of token: %w", err)
	}
	return prettyJson.String(), nil
}

// DecodePayloadAsRawJSON extracts the payload and returns the raw JSON.
func DecodePayloadAsRawJSON(s string) ([]byte, error) {
	parts := strings.SplitN(s, ".", 3)
	if len(parts) != 3 {
		return nil, fmt.Errorf("wants %d segments but got %d segments", 3, len(parts))
	}
	payloadJSON, err := decodePayload(parts[1])
	if err != nil {
		return nil, fmt.Errorf("could not decode the payload: %w", err)
	}
	return payloadJSON, nil
}

func decodePayload(payload string) ([]byte, error) {
	b, err := base64.URLEncoding.WithPadding(base64.NoPadding).DecodeString(payload)
	if err != nil {
		return nil, fmt.Errorf("invalid base64: %w", err)
	}
	return b, nil
}
