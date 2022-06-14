package controllers

import (
	"context"
	"fmt"

	"github.com/engytita/engytita-operator/api/v1alpha1"
	"github.com/engytita/engytita-operator/pkg/kubernetes/client"
	"github.com/engytita/engytita-operator/pkg/reconcile"
	"github.com/engytita/engytita-operator/pkg/reconcile/cache"
	"github.com/engytita/engytita-operator/pkg/reconcile/pipeline"
	"gopkg.in/yaml.v2"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/serializer"
	v1 "k8s.io/client-go/applyconfigurations/core/v1"
	k8s "k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/record"
	ctrl "sigs.k8s.io/controller-runtime"
	runtimeClient "sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"
)

// CacheReconciler reconciles a Cache object
type CacheReconciler struct {
	runtimeClient.Client
	Scheme *runtime.Scheme
	record.EventRecorder
}

//+kubebuilder:rbac:groups=engytita.org,resources=caches,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=engytita.org,resources=caches/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=engytita.org,resources=caches/finalizers,verbs=update

// +kubebuilder:rbac:groups=apps,namespace=engytita-operator-system,resources=daemonsets,verbs=get;list;watch;create;delete;deletecollection;update;patch
// +kubebuilder:rbac:groups=core,namespace=engytita-operator-system,resources=services;configmaps,verbs=get;list;watch;create;delete;deletecollection;update;patch

// Reconcile the Cache resource
func (r *CacheReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	reqLogger := log.FromContext(ctx)

	instance := &v1alpha1.Cache{}
	if err := r.Get(ctx, req.NamespacedName, instance); err != nil {
		if errors.IsNotFound(err) {
			reqLogger.Info("Cache CR not found")
			return ctrl.Result{}, nil
		}
		// Error reading the object - requeue the request.
		return ctrl.Result{}, fmt.Errorf("unable to fetch Cache CR %w", err)
	}

	// Don't reconcile Infinispan CRs marked for deletion
	if instance.GetDeletionTimestamp() != nil {
		reqLogger.Info(fmt.Sprintf("Ignoring Cache CR '%s:%s' marked for deletion", instance.Namespace, instance.Name))
		return ctrl.Result{}, nil
	}

	ctxProvider := reconcile.ContextProviderFunc(func(i interface{}) (reconcile.Context, error) {
		return pipeline.NewContext(ctx, reqLogger, &client.Runtime{
			Client:        r.Client,
			Ctx:           ctx,
			EventRecorder: r.EventRecorder,
			Namespace:     instance.Namespace,
			Owner:         instance,
			Scheme:        r.Scheme,
		}), nil
	})

	retry, delay, err := cache.PipelineBuilder(instance).
		WithContextProvider(ctxProvider).
		Build().
		Process(instance)

	reqLogger.Info("Done", "requeue", retry, "requeueAfter", delay, "error", err)
	return ctrl.Result{Requeue: retry, RequeueAfter: delay}, err
}

// SetupWithManager sets up the controller with the Manager.
func (r *CacheReconciler) SetupWithManager(mgr ctrl.Manager) error {
	r.Scheme = mgr.GetScheme()
	r.Client = mgr.GetClient()

	config := mgr.GetConfig()
	config.GroupVersion = &corev1.SchemeGroupVersion
	config.APIPath = "/api"
	config.NegotiatedSerializer = serializer.WithoutConversionCodecFactory{CodecFactory: serializer.NewCodecFactory(r.Scheme)}
	config.UserAgent = rest.DefaultKubernetesUserAgent()

	restClient, err := rest.RESTClientFor(mgr.GetConfig())
	if err != nil {
		panic(err.Error())
	}

	spec := v1.ServiceSpec().
		WithType(corev1.ServiceTypeClusterIP).
		WithClusterIP(corev1.ClusterIPNone).
		WithSelector(map[string]string{}).
		WithPorts(
			v1.ServicePort().WithName("infinispan").WithPort(11222),
		)

	serviceConfig := v1.Service("test", "engytita-operator-system").WithSpec(spec)
	s, err := k8s.New(restClient).
		CoreV1().
		Services("engytita-operator-system").
		Apply(context.TODO(), serviceConfig, metav1.ApplyOptions{FieldManager: "mycontroller", Force: true})

	fmt.Printf("Err=%s\n", err)
	var y string
	if s != nil {
		yBytes, err := yaml.Marshal(s)
		y = string(yBytes)
		fmt.Println(err)
	}
	fmt.Printf("Yaml=%s\nErr=%s", y, err)

	r.EventRecorder = mgr.GetEventRecorderFor("cache")
	return ctrl.NewControllerManagedBy(mgr).
		For(&v1alpha1.Cache{}).
		Complete(r)
}
