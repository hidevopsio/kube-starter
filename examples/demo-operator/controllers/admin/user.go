package admin

import (
	"context"
	"os"

	"github.com/go-logr/logr"
	"github.com/hidevopsio/hiboot/pkg/app"
	"github.com/hidevopsio/kube-starter/pkg/operator"
	adminv1alpha1 "github.com/icloudnative-net/demo-operator/apis/admin/v1alpha1"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// UserReconciler reconciles a User object
type UserReconciler struct {
	client.Client
	Log    logr.Logger
	Scheme *runtime.Scheme
}

func newUserReconciler(manager *operator.Manager) *UserReconciler {
	reconciler := &UserReconciler{
		Client: manager.GetClient(),
		Log: ctrl.Log.WithName("controllers").WithName("admin").WithName("User"),
		Scheme: manager.GetScheme(manager.Manager),
	}
	err := reconciler.SetupWithManager(manager)
	if err != nil {
		setupLog.Error(err, "unable to create controller", "controller", "Project")
		os.Exit(1)
	}

	return reconciler
}

func init() {
	app.Register(newUserReconciler)
}

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
// TODO(user): Modify the Reconcile function to compare the state specified by
// the User object against the actual cluster state, and then
// perform operations to make the cluster state reflect the state specified by
// the user.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.7.0/pkg/reconcile
func (r *UserReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	_ = r.Log.WithValues("user", req.NamespacedName)

	// your logic here

	return ctrl.Result{}, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *UserReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&adminv1alpha1.User{}).
		Complete(r)
}

