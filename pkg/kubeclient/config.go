package kubeclient

import (
	"github.com/hidevopsio/hiboot/pkg/log"
	"github.com/hidevopsio/kube-starter/pkg/kubeconfig"
	"github.com/hidevopsio/kube-starter/pkg/oidc"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/rest"
	"os"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"time"
)

const (
	Subject = "subject"
)

// KubeClient new kube client
func KubeClient(scheme *runtime.Scheme, cfg *rest.Config) (cli *Client, err error) {
	log.Info("Creating kube client")
	var k8sClient client.Client
	cli = &Client{}
	go func() {
		var count int
		for cli.Client == nil {
			count++
			log.Info("Creating kube client")
			k8sClient, err = client.New(cfg, client.Options{Scheme: scheme})
			log.Infof("created kube client by retried %v times: %v", k8sClient, count)
			if err == nil && k8sClient != nil {
				cli.Client = k8sClient
				log.Infof("Got kube client by retry %v times: %v", k8sClient, count)
				break
			}
			time.Sleep(time.Second)
		}
	}()

	return
}

// RuntimeKubeClient new runtime kube client
func RuntimeKubeClient(scheme *runtime.Scheme, token *oidc.Token, useToken bool, properties *Properties) (cli client.Client, err error) {
	var cfg *rest.Config
	cfg, err = kubeconfig.Kubeconfig(properties.DefaultInCluster)
	if err != nil {
		log.Warn(err)
		return
	}
	cfg.QPS = properties.QPS
	cfg.Burst = properties.Burst
	cfg.Timeout = properties.Timeout

	if token != nil && token.Claims != nil && token.Data != "" {
		kubeServiceHost := os.Getenv("KUBERNETES_SERVICE_HOST")
		if kubeServiceHost == "" && useToken {
			cfg.BearerToken = token.Data
			cfg.BearerTokenFile = ""
		} else {
			switch properties.OIDCScope {
			case "email":
				cfg.Impersonate.UserName = token.Claims.Email
			case "profile":
				cfg.Impersonate.UserName = token.Claims.Username
			case "openid":
				cfg.Impersonate.UserName = token.Claims.Issuer + "#" + token.Claims.Subject
			default:
				cfg.Impersonate.UserName = token.Claims.Email
			}
		}
	} else {
		log.Warn("Unauthorized")
		err = errors.NewUnauthorized("Unauthorized")
		return
	}

	cli, err = client.New(cfg, client.Options{Scheme: scheme})
	if err != nil {
		log.Warn(err)
	}

	log.Infof("created runtime client with qps: %v, burst: %v", cfg.QPS, cfg.Burst)
	return
}
