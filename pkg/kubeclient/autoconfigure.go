package kubeclient

import (
	"github.com/hidevopsio/hiboot/pkg/app"
	"github.com/hidevopsio/hiboot/pkg/app/web/context"
	"github.com/hidevopsio/hiboot/pkg/at"
	"github.com/hidevopsio/hiboot/pkg/factory/instantiate"
	"github.com/hidevopsio/hiboot/pkg/log"
	"github.com/hidevopsio/kube-starter/pkg/kubeconfig"
	"github.com/hidevopsio/kube-starter/pkg/oidc"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/rest"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

const (
	Profile = "kubeclient"
)

type RestConfig struct {
	at.Scope `value:"prototype"`

	*rest.Config
}

// Client is the encapsulation of the default kube client
type Client struct {
	at.Scope `value:"prototype"`

	client.Client
}

type FnClient func(params ...interface{}) *Client
type FnRuntimeClient func(params ...interface{}) *RuntimeClient

type configuration struct {
	at.AutoConfiguration

	Properties *Properties

	clientFactory        *instantiate.ScopedInstanceFactory[*Client]
	runtimeClientFactory *instantiate.ScopedInstanceFactory[*RuntimeClient]
	//clientConfigFactory        *instantiate.ScopedInstanceFactory[*kubeconfig.ClusterConfig]
	//runtimeClientConfigFactory *instantiate.ScopedInstanceFactory[*kubeconfig.RuntimeClusterConfig]
}

func newConfiguration(prop *Properties) *configuration {

	return &configuration{Properties: prop}
}

func init() {
	app.Register(
		newConfiguration,
		new(Properties),
		new(instantiate.ScopedInstanceFactory[*Client]),
		new(instantiate.ScopedInstanceFactory[*RuntimeClient]),
		//new(instantiate.ScopedInstanceFactory[*kubeconfig.ClusterConfig]),
		//new(instantiate.ScopedInstanceFactory[*kubeconfig.RuntimeClusterConfig]),
	)
}

func (c *configuration) ClusterConfig(prop *Properties) (cluster *kubeconfig.ClusterConfig, err error) {

	cluster = &kubeconfig.ClusterConfig{
		ClusterInfo: kubeconfig.ClusterInfo{
			Name:      "default",
			Config:    "", // if config is empty, will use the default $HOME/.kube/config
			InCluster: c.inCluster(prop),
		},
	}
	log.Infof("ClusterConfig: %+v", cluster)
	return
}

func (c *configuration) RuntimeClusterConfig(ctx context.Context, token *oidc.Token, prop *Properties) (cluster *kubeconfig.RuntimeClusterConfig, err error) {

	clusterName := ctx.GetHeader("cluster")
	if clusterName == "" {
		clusterName = "default"
	}
	cluster = &kubeconfig.RuntimeClusterConfig{
		ClusterInfo: kubeconfig.ClusterInfo{
			Name:      clusterName,
			Config:    ctx.GetHeader("config"),
			InCluster: c.inCluster(prop),
		},
		Username: token.Claims.Username,
	}
	log.Infof("RuntimeClusterConfig: %+v", cluster)
	return
}

func (c *configuration) ClientFunc(clientFactory *instantiate.ScopedInstanceFactory[*Client]) (clientFunc FnClient, err error) {
	clientFunc = func(params ...interface{}) *Client {
		// TODO: check if param is a string which specify the cluster name
		//cc := c.clientConfigFactory.GetInstance()
		//params = append(params, cc)
		return clientFactory.GetInstance(params...)
	}
	return
}

func (c *configuration) RuntimeClientFunc(clientFactory *instantiate.ScopedInstanceFactory[*RuntimeClient]) (clientFunc FnRuntimeClient, err error) {
	clientFunc = func(params ...interface{}) *RuntimeClient {
		// TODO: check if param is a string which specify the cluster name and username
		//cc := c.runtimeClientConfigFactory.GetInstance(params...)
		//params = append(params, cc)
		rc := clientFactory.GetInstance(params...)
		return rc
	}
	return
}

func (c *configuration) RestConfig(cluster *kubeconfig.ClusterConfig) (restConfig *RestConfig, err error) {
	restConfig = new(RestConfig)

	restConfig.Config, err = kubeconfig.Kubeconfig(&cluster.ClusterInfo)
	restConfig.Config.QPS = c.Properties.QPS
	restConfig.Config.Burst = c.Properties.Burst
	restConfig.Config.Timeout = c.Properties.Timeout
	return
}

func (c *configuration) Client(scheme *runtime.Scheme, cfg *RestConfig) (cli *Client, err error) {

	cli = &Client{}
	cli.Client, err = KubeClient(scheme, cfg)

	return
}

// RuntimeClient is the client the runtime kube client
type RuntimeClient struct {
	at.Scope `value:"prototype"`

	client.Client

	Context context.Context `json:"context"`
}

func (c *configuration) RuntimeClient(ctx context.Context, scheme *runtime.Scheme, token *oidc.Token, cluster *kubeconfig.RuntimeClusterConfig) (cli *RuntimeClient, err error) {
	cli = new(RuntimeClient)
	var newClient client.Client

	uid := token.Claims.Username
	cluster.Username = uid

	newClient, err = RuntimeKubeClient(scheme, token, true, c.Properties, cluster)
	if err != nil {
		log.Error(err)
		return
	}

	cli = &RuntimeClient{
		Context: ctx,
		Client:  newClient,
	}
	return
}

func (c *configuration) inCluster(prop *Properties) bool {
	var inCluster bool
	if prop == nil {
		inCluster = false
	} else if prop.DefaultInCluster != nil {
		inCluster = *prop.DefaultInCluster
	}
	return inCluster
}
