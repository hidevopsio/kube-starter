package kubeclient

import (
	"github.com/hidevopsio/hiboot/pkg/app"
	"github.com/hidevopsio/hiboot/pkg/app/web/context"
	"github.com/hidevopsio/hiboot/pkg/at"
	"github.com/hidevopsio/hiboot/pkg/log"
	"github.com/hidevopsio/hiboot/pkg/utils/cmap"
	"github.com/hidevopsio/kube-starter/pkg/kubeconfig"
	"github.com/hidevopsio/kube-starter/pkg/oidc"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/rest"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

const (
	Profile = "kubeclient"
)

type clientCache struct {
	client client.Client
	uid    string
	token  string
}

type KubeRuntimeClients struct {
	clients cmap.ConcurrentMap
}

func (c *KubeRuntimeClients) Get(uid string) (client client.Client, ok bool) {
	var cachedClient interface{}
	cachedClient, ok = c.clients.Get(uid)
	if ok {
		cc := cachedClient.(clientCache)
		log.Infof("%v reuse cached runtime client", uid)
		client = cc.client
	}
	return
}

func (c *KubeRuntimeClients) Set(uid string, client client.Client) {
	c.clients.Set(uid, clientCache{client: client, uid: uid})
	return
}

type configuration struct {
	at.AutoConfiguration

	Properties *Properties

	kubeRuntimeClients *KubeRuntimeClients
}

func newConfiguration(kubeClients *KubeRuntimeClients) *configuration {
	kubeClients.clients = cmap.New()
	return &configuration{kubeRuntimeClients: kubeClients}
}

func init() {
	app.Register(new(KubeRuntimeClients), newConfiguration)
}

func (c *configuration) RestConfig() (cfg *rest.Config, err error) {
	cfg, err = kubeconfig.Kubeconfig(c.Properties.DefaultInCluster)
	cfg.QPS = c.Properties.QPS
	cfg.Burst = c.Properties.Burst
	cfg.Timeout = c.Properties.Timeout
	return
}

// Client is the encapsulation of the default kube client
type Client struct {
	//at.ContextAware

	client.Client
}

func (c *configuration) Client(scheme *runtime.Scheme, cfg *rest.Config) (cli *Client, err error) {

	cli, err = KubeClient(scheme, cfg)

	return
}

// ImpersonateClient is the client impersonate kube client
type ImpersonateClient struct {
	at.ContextAware

	client.Client

	Context context.Context `json:"context"`
}

func (c *configuration) ImpersonateClient(ctx context.Context, scheme *runtime.Scheme, token *oidc.Token, cfg *rest.Config) (cli *ImpersonateClient) {
	cli = new(ImpersonateClient)

	newCli, _ := RuntimeKubeClient(scheme, token, false, c.Properties)

	cli = &ImpersonateClient{
		Context: ctx,
		Client:  newCli,
	}
	return
}

// TokenizeClient is the client tokenize kube client
type TokenizeClient struct {
	at.ContextAware

	client.Client

	Context context.Context `json:"context"`
}

func (c *configuration) TokenizeClient(ctx context.Context, scheme *runtime.Scheme, token *oidc.Token, cfg *rest.Config) (cli *TokenizeClient) {
	cli = new(TokenizeClient)

	newCli, _ := RuntimeKubeClient(scheme, token, true, c.Properties)

	cli = &TokenizeClient{
		Context: ctx,
		Client:  newCli,
	}
	return
}

// RuntimeClient is the client the runtime kube client
type RuntimeClient struct {
	at.ContextAware

	client.Client

	Context context.Context `json:"context"`
}

func (c *configuration) RuntimeClient(ctx context.Context, scheme *runtime.Scheme, token *oidc.Token, cfg *rest.Config) (cli *RuntimeClient, err error) {
	cli = new(RuntimeClient)
	var newClient client.Client
	var ok bool

	uid := token.Claims.Username
	newClient, ok = c.kubeRuntimeClients.Get(uid)
	if !ok {
		newClient, err = RuntimeKubeClient(scheme, token, true, c.Properties)
		if err != nil {
			log.Error(err)
			return
		}
		log.Infof("%v create new runtime client", uid)
		c.kubeRuntimeClients.Set(uid, newClient)
	}

	cli = &RuntimeClient{
		Context: ctx,
		Client:  newClient,
	}
	return
}
