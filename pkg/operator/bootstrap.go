package operator

import (
	"github.com/hidevopsio/hiboot/pkg/log"
	"os"

	"github.com/hidevopsio/hiboot/pkg/app"
	"github.com/hidevopsio/hiboot/pkg/at"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/healthz"
	"sigs.k8s.io/controller-runtime/pkg/manager"
)

type bootstrap struct {
	at.EnableScheduling

	manager manager.Manager
}

func newBootstrap(manager manager.Manager) *bootstrap {
	return &bootstrap{
		manager: manager,
	}
}

func init() {
	app.Register(newBootstrap)
}

func (b *bootstrap) Run(_ struct {
	at.Scheduled `limit:"1"`
}) {
	log.Debugf("bootstrap.Run()")
	//var setupLog = ctrl.Log.WithName("setup")
	if err := b.manager.AddHealthzCheck("healthz", healthz.Ping); err != nil {
		log.Error(err, "unable to set up health check")
		//setupLog.Error(err, "unable to set up health check")
		os.Exit(1)
	}
	if err := b.manager.AddReadyzCheck("readyz", healthz.Ping); err != nil {
		//setupLog.Error(err, "unable to set up ready check")
		log.Error(err, "unable to set up ready check")
		os.Exit(1)
	}

	//setupLog.Info("starting operator manager")
	log.Info("starting operator manager")
	if err := b.manager.Start(ctrl.SetupSignalHandler()); err != nil {
		//setupLog.Error(err, "problem running manager")
		log.Error(err, "problem running manager")
		os.Exit(1)
	}
}
