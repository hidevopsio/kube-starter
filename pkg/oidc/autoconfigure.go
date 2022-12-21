package oidc

import (
	"strings"
	"time"

	"github.com/hidevopsio/hiboot/pkg/app"
	"github.com/hidevopsio/hiboot/pkg/app/web/context"
	"github.com/hidevopsio/hiboot/pkg/at"
	"github.com/hidevopsio/hiboot/pkg/log"
	"k8s.io/apimachinery/pkg/api/errors"
)

const (
	Profile = "oidc"
)

type configuration struct {
	at.AutoConfiguration
}

func newConfiguration() *configuration {
	return &configuration{}
}

func init() {
	app.Register(newConfiguration)
}

// Token
type Token struct {
	at.ContextAware

	Context context.Context `json:"context"`
	Data    string          `json:"data"`
	Claims  *Claims         `json:"claims"`
}

// Token instantiate bearer token to object
func (c *configuration) Token(ctx context.Context) (token *Token, err error) {
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
		err = errors.NewUnauthorized("Unauthorized")
		log.Errorf("%v -> %v", pe, err)
		return  // fixes the nil pointer issue
	}
	if token.Claims.Expiry.Before(time.Now()) {
		err = errors.NewUnauthorized("Expired")
		log.Errorf("%v", err)
	}
	return
}
