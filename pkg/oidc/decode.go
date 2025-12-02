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
	// If the name is already DNS-1123 compliant, return as-is
	if IsNameValidate(name) {
		return name
	}

	// Convert to lowercase
	name = strings.ToLower(name)

	// Replace non-DNS-1123 characters with '-'
	result := make([]byte, len(name))
	for i := range name {
		if (name[i] >= '0' && name[i] <= '9') || (name[i] <= 'z' && name[i] >= 'a') {
			result[i] = name[i]
			continue
		}
		result[i] = '-'
	}

	// Generate a consistent 4-character suffix from hash of original name
	// This prevents collisions when different usernames format to the same string
	hash := sha256.Sum256([]byte(name))
	suffix := hex.EncodeToString(hash[:])[:4]

	return string(result) + "-" + suffix
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
	cls := &Claims{
		Issuer:            claims.Issuer,
		Subject:           claims.Subject,
		Name:              claims.Name,
		Username:          claims.Username,
		FormattedUsername: FormatName(claims.Username),
		Email:             claims.Email,
		Expiry:            time.Unix(claims.ExpiresAt, 0),
		Pretty:            prettyJson.String(),
	}

	// fill username as the value of name if it is empty
	if cls.Username == "" {
		cls.Username = cls.Name
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
