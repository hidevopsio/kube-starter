package admin

import (
	"context"
	"github.com/hidevopsio/hiboot/pkg/at"
	"github.com/hidevopsio/kube-starter/pkg/kube"
	"github.com/hidevopsio/kube-starter/pkg/operator"
	"os"

	"github.com/go-logr/logr"
	"github.com/hidevopsio/hiboot/pkg/app"
	adminv1alpha1 "github.com/hidevopsio/kube-starter/examples/demo-operator/apis/admin/v1alpha1"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// UserReconciler reconciles a User object
type UserReconciler struct {
	kube.Controller
	at.Scope `value:"prototype"`
	client.Client
	log    logr.Logger
	Scheme *runtime.Scheme
}

func newUserReconciler(manager *operator.Manager, scheme *runtime.Scheme) *UserReconciler {
	log := ctrl.Log.WithName("controllers").WithName("admin").WithName("User")
	reconciler := &UserReconciler{
		Client: manager.GetClient(),
		log:    ctrl.Log.WithName("controllers").WithName("admin").WithName("User"),
		Scheme: scheme,
	}
	err := reconciler.SetupWithManager(manager)
	if err != nil {
		log.Error(err, "unable to create controller", "controller", "Project")
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
func (r *UserReconciler) Reconcile(ctx context.Context, req ctrl.Request) (result ctrl.Result, err error) {
	log := r.log.WithValues("user", req.NamespacedName)
	// your logic here
	log.Info("Reconcile() is called")

	var user adminv1alpha1.User
	err = r.Get(ctx, req.NamespacedName, &user)
	if err == nil {
		log.Info("[]", "[user]", user)
	}

	return ctrl.Result{}, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *UserReconciler) SetupWithManager(mgr *operator.Manager) error {
	return ctrl.NewControllerManagedBy(mgr.Manager).
		For(&adminv1alpha1.User{}).
		Complete(r)
}
