package controllers

import (
	"context"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	examplev1 "my.domain/api/v1"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/log"
)

// MyAppReconciler reconciles a MyApp object
type MyAppReconciler struct {
	client.Client
	Scheme *runtime.Scheme
}

// +kubebuilder:rbac:groups=example.com,resources=myapps,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=apps,resources=deployments,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=core,resources=pods,verbs=get;list;watch

func (r *MyAppReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	log := log.FromContext(ctx)
	log.Info("Reconciling MyApp", " namespace: ", req.Namespace, " name: ", req.Name)

	var myApp examplev1.MyApp
	if err := r.Get(ctx, req.NamespacedName, &myApp); err != nil {
		if client.IgnoreNotFound(err) != nil {
			log.Error(err, "unable to fetch MyApp", " namespace: ", req.Namespace, " name: ", req.Name)
			return ctrl.Result{}, err
		}
		log.Info("MyApp resource not found. Ignoring since object must be deleted", " namespace: ", req.Namespace, " name: ", req.Name)
		return ctrl.Result{}, nil
	}

	log.Info("Fetched MyApp resource", " namespace: ", myApp.Namespace, " name: ", myApp.Name, "replicas", *myApp.Spec.Replicas, "image", *myApp.Spec.Image)

	deployment := &appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name:      myApp.Name,
			Namespace: myApp.Namespace,
		},
	}

	_, err := controllerutil.CreateOrUpdate(ctx, r.Client, deployment, func() error {
		replicas := *myApp.Spec.Replicas
		image := *myApp.Spec.Image
		deployment.Spec = appsv1.DeploymentSpec{
			Replicas: &replicas,
			Selector: &metav1.LabelSelector{
				MatchLabels: map[string]string{"app": myApp.Name},
			},
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: map[string]string{"app": myApp.Name},
				},
				Spec: corev1.PodSpec{
					Containers: []corev1.Container{{
						Name:  "myapp",
						Image: image,
					}},
				},
			},
		}

		return nil
	})
	if err != nil {
		log.Error(err, "unable to create or update Deployment", " namespace: ", myApp.Namespace, " name: ", myApp.Name)
		return ctrl.Result{}, err
	}

	log.Info("Successfully reconciled MyApp ", " namespace: ", myApp.Namespace, " name: ", myApp.Name)

	return ctrl.Result{}, nil
}

func (r *MyAppReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&examplev1.MyApp{}).
		Owns(&appsv1.Deployment{}).
		Complete(r)
}
