// Package oidc provides JWT token manipulations.
// See https://tools.ietf.org/html/rfc7519#section-4.1.3
package oidc

import (
	"bytes"
	"context"
	"crypto"
	"crypto/rsa"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"github.com/coreos/go-oidc/v3/oidc"
	"github.com/golang-jwt/jwt"
	"github.com/hidevopsio/hiboot/pkg/log"
	"strings"
	"time"
)

type IDTokenVerifier struct {
	Verifiers []*oidc.IDTokenVerifier
}

// getOIDCTokenVerifier verifies an OIDC token using the issuer's JWK set
func newOIDCTokenVerifier(properties *Properties) (verifier *IDTokenVerifier, err error) {
	// Validate the token using OIDC
	verifier = new(IDTokenVerifier)
	for _, publicKey := range properties.PublicKeys {
		var verifyKey *rsa.PublicKey
		var pk []byte
		pk, err = base64.URLEncoding.WithPadding(base64.StdPadding).DecodeString(publicKey)
		if verifyKey, err = jwt.ParseRSAPublicKeyFromPEM(pk); err == nil {
			keySet := &oidc.StaticKeySet{PublicKeys: []crypto.PublicKey{verifyKey}}
			v := oidc.NewVerifier("", keySet, &oidc.Config{
				SkipClientIDCheck: true,
				SkipIssuerCheck:   true,
			})
			verifier.Verifiers = append(verifier.Verifiers, v)
		}
	}
	return
}

// verifyOIDCToken verifies an OIDC token using the issuer's JWK set
func verifyOIDCToken(verifier *IDTokenVerifier, tokenString string) (err error) {
	for _, v := range verifier.Verifiers {
		_, err = v.Verify(context.Background(), tokenString)
		if err == nil {
			log.Infof("OIDC Token verification succeeded!")
			return
		}
	}

	log.Errorf("OIDC Token verification failed: %v", err)
	return
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
		Issuer:   claims.Issuer,
		Subject:  claims.Subject,
		Name:     claims.Name,
		Username: claims.Username,
		Email:    claims.Email,
		Expiry:   time.Unix(claims.ExpiresAt, 0),
		Pretty:   prettyJson.String(),
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

func decodePayload(payload string) (b []byte, err error) {
	b, err = base64.URLEncoding.WithPadding(base64.NoPadding).DecodeString(payload)
	if err != nil {
		return nil, fmt.Errorf("invalid base64: %w", err)
	}
	return b, nil
}
