package oidc

import "github.com/hidevopsio/hiboot/pkg/at"

type Properties struct {
	at.ConfigurationProperties `value:"oidc"`
	Verify   bool   `json:"verify"`
	Issuer   string `json:"issuer"`
	ClientID string `json:"clientId"`
}
