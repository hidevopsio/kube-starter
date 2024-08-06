package main

import (
	"github.com/hidevopsio/hiboot/pkg/app"
	"github.com/hidevopsio/hiboot/pkg/app/web"
	"github.com/hidevopsio/hiboot/pkg/starter/actuator"
	"github.com/hidevopsio/hiboot/pkg/starter/logging"
	examplev1 "github.com/hidevopsio/kube-starter/examples/hiboot-operator/api/v1" // Ensure this import path is correct
	_ "github.com/hidevopsio/kube-starter/examples/hiboot-operator/controllers"    // Ensure this import path is correct
	"github.com/hidevopsio/kube-starter/pkg/kubeclient"
	"github.com/hidevopsio/kube-starter/pkg/operator"
	"k8s.io/apimachinery/pkg/runtime"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	clientgoscheme "k8s.io/client-go/kubernetes/scheme"
)

func init() {
	scheme := runtime.NewScheme()
	utilruntime.Must(clientgoscheme.AddToScheme(scheme))
	utilruntime.Must(examplev1.AddToScheme(scheme))
	app.Register(scheme)
}

func main() {
	web.NewApplication().
		SetProperty(logging.Level, logging.LevelDebug).
		SetProperty(app.ProfilesInclude,
			actuator.Profile,
			logging.Profile,
			operator.Profile,
			kubeclient.Profile,
		).
		Run()
}
