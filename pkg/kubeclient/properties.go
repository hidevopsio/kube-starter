package kubeclient

import (
	"time"

	"github.com/hidevopsio/hiboot/pkg/at"
)

// Properties the operator properties
type Properties struct {
	at.ConfigurationProperties `value:"kubeclient"`
	at.AutoWired

	// use DefaultInCluster as default
	DefaultInCluster *bool `json:"defaultInCluster"`

	//OIDC Scope Impersonate
	OIDCScope string `json:"oidcScope" default:"profile"`

	QPS float32 `json:"qps"`

	// Maximum burst for throttle.
	// If it's zero, the created RESTClient will use DefaultBurst: 10.
	Burst int `json:"burst"`

	// The maximum length of time to wait before giving up on a server request. A value of zero means no timeout.
	Timeout time.Duration `json:"timeout"`
}
