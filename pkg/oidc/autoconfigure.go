package oidc

import (
	"github.com/hidevopsio/hiboot/pkg/app"
	"github.com/hidevopsio/hiboot/pkg/app/web/context"
	"github.com/hidevopsio/hiboot/pkg/at"
	"github.com/hidevopsio/hiboot/pkg/log"
	"k8s.io/apimachinery/pkg/api/errors"
	"strings"
)

const (
	Profile = "oidc"
)

type configuration struct {
	at.AutoConfiguration

	prop *Properties
}

func newConfiguration(prop *Properties) *configuration {
	return &configuration{prop: prop}
}

func init() {
	app.Register(newConfiguration, new(Properties))
}

// Token is the token object
type Token struct {
	at.Scope `value:"request"`

	Context context.Context `json:"context"`
	Data    string          `json:"data"`
	Claims  *Claims         `json:"claims"`
}

func (c *configuration) IDTokenVerifier() (verifier *IDTokenVerifier, err error) {
	return newOIDCTokenVerifier(c.prop)
}

// Token instantiate bearer token to object
func (c *configuration) Token(ctx context.Context, verifier *IDTokenVerifier) (token *Token, err error) {
	token = new(Token)
	if ctx == nil {
		err = errors.NewBadRequest("unknown context")
		log.Error(err)
		return
	}
	bearerToken := ctx.GetHeader("Authorization")
	if bearerToken == "" {
		bearerToken = ctx.URLParam("token")
	}
	token.Data = strings.Replace(bearerToken, "Bearer ", "", -1)
	token.Claims, err = DecodeWithoutVerify(token.Data)
	if err != nil {
		pe := err
		err = errors.NewUnauthorized(err.Error())
		log.Errorf("%v -> %v", pe, err)
		return // fixes the nil pointer issue
	}
	if c.prop.Verify {
		err = verifyOIDCToken(verifier, token.Data)
		if err != nil {
			err = errors.NewUnauthorized(err.Error())
		}
	}

	return
}
