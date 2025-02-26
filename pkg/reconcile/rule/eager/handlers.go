package eager

import (
	"fmt"

	"github.com/gingersnap-project/operator/api/v1alpha1"
	bindingv1 "github.com/gingersnap-project/operator/pkg/applyconfigurations/servicebinding/v1beta1"
	"github.com/gingersnap-project/operator/pkg/images"
	"github.com/gingersnap-project/operator/pkg/kubernetes/client"
	"github.com/gingersnap-project/operator/pkg/reconcile/meta"
	"github.com/gingersnap-project/operator/pkg/reconcile/rule"
	apiappsv1 "k8s.io/api/apps/v1"
	apicorev1 "k8s.io/api/core/v1"
	apimetav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	appsv1 "k8s.io/client-go/applyconfigurations/apps/v1"
	corev1 "k8s.io/client-go/applyconfigurations/core/v1"
	metav1 "k8s.io/client-go/applyconfigurations/meta/v1"
	runtimeClient "sigs.k8s.io/controller-runtime/pkg/client"
)

func LoadCache(r *v1alpha1.EagerCacheRule, ctx *rule.Context) {
	cacheRef := r.CacheService()
	cache := &v1alpha1.Cache{}
	err := ctx.Client().
		WithNamespace(cacheRef.Namespace).
		Load(cacheRef.Name, cache)

	if err != nil {
		msg := fmt.Sprintf("unable to load Cache CR '%s'", cacheRef)
		r.SetCondition(
			v1alpha1.EagerCacheRuleCondition{
				Type:    v1alpha1.EagerCacheRuleConditionReady,
				Status:  apimetav1.ConditionFalse,
				Message: msg,
			},
		)
		if err := ctx.Client().UpdateStatus(r); err != nil {
			ctx.Requeue(fmt.Errorf("unable to update Ready condition on LoadCache failure: %w", err))
			return
		}
		ctx.Requeue(fmt.Errorf("%s: %w", msg, err))
	}
	ctx.Cache = cache
}

func ApplyDBServiceBinding(_ *v1alpha1.EagerCacheRule, ctx *rule.Context) {
	cache := ctx.Cache
	labels := meta.GingersnapLabels("db-syncer", meta.ComponentDBSyncer, cache.Name)

	var serviceRef *bindingv1.ServiceBindingServiceReferenceApplyConfiguration
	ds := cache.Spec.DataSource
	if service := ds.ServiceProviderRef; service != nil {
		serviceRef = bindingv1.ServiceBindingServiceReference().
			WithAPIVersion(service.ApiVersion).
			WithKind(service.Kind).
			WithName(service.Name)
	} else {
		serviceRef = bindingv1.ServiceBindingServiceReference().
			WithAPIVersion(apicorev1.SchemeGroupVersion.String()).
			WithKind("Secret").
			WithName(ds.SecretRef.Name)
	}

	sb := bindingv1.ServiceBinding(cache.CacheService().DBSyncerDataServiceBinding(), cache.Namespace).
		WithLabels(labels).
		WithOwnerReferences(ctx.Client().OwnerReference()).
		WithSpec(
			bindingv1.ServiceBindingSpec().
				WithService(serviceRef).
				WithType(ds.DbType.ServiceBinding()).
				WithWorkload(
					bindingv1.ServiceBindingWorkloadReference().
						WithAPIVersion(apiappsv1.SchemeGroupVersion.String()).
						WithKind("Deployment").
						WithSelector(
							apimetav1.LabelSelector{
								MatchLabels: labels,
							},
						),
				),
		)

	if err := ctx.Client().Apply(sb); err != nil {
		ctx.Requeue(fmt.Errorf("unable to apply DB ServiceBinding: %w", err))
	}
}

func ApplyCacheServiceBinding(_ *v1alpha1.EagerCacheRule, ctx *rule.Context) {
	cache := ctx.Cache
	labels := meta.GingersnapLabels("db-syncer", meta.ComponentDBSyncer, cache.Name)

	sb := bindingv1.ServiceBinding(cache.CacheService().DBSyncerCacheServiceBinding(), cache.Namespace).
		WithLabels(labels).
		WithOwnerReferences(ctx.Client().OwnerReference()).
		WithSpec(
			bindingv1.ServiceBindingSpec().
				WithService(
					bindingv1.ServiceBindingServiceReference().
						WithAPIVersion(apicorev1.SchemeGroupVersion.String()).
						WithKind("Secret").
						WithName(cache.CacheService().DBSyncerCacheServiceBindingSecret()),
				).
				WithType("gingersnap").
				WithWorkload(
					bindingv1.ServiceBindingWorkloadReference().
						WithAPIVersion(apiappsv1.SchemeGroupVersion.String()).
						WithKind("Deployment").
						WithSelector(
							apimetav1.LabelSelector{
								MatchLabels: labels,
							},
						),
				),
		)

	if err := ctx.Client().Apply(sb); err != nil {
		ctx.Requeue(fmt.Errorf("unable to apply Cache ServiceBinding: %w", err))
	}
}

func ApplyDBSyncer(r *v1alpha1.EagerCacheRule, ctx *rule.Context) {
	cache := ctx.Cache
	labels := meta.GingersnapLabels("db-syncer", meta.ComponentDBSyncer, cache.Name)

	cacheService := cache.CacheService()
	name := cacheService.DBSyncerName()
	deployment := appsv1.Deployment(name, cache.Namespace).
		WithLabels(labels).
		WithOwnerReferences(ctx.Client().OwnerReference()).
		WithSpec(appsv1.DeploymentSpec().
			WithSelector(
				metav1.LabelSelector().WithMatchLabels(labels),
			).
			WithTemplate(corev1.PodTemplateSpec().
				WithName("db-syncer").
				WithLabels(labels).
				WithSpec(corev1.PodSpec().
					WithServiceAccountName(cache.Name).
					WithContainers(
						corev1.Container().
							WithName("db-syncer").
							WithImage(images.DBSyncer).
							WithEnv(
								corev1.EnvVar().WithName("GINGERSNAP_DYNAMIC_MEMBERSHIP").WithValue("true"),
								corev1.EnvVar().WithName("GINGERSNAP_K8S_NAMESPACE").WithValue(cacheService.Namespace),
								corev1.EnvVar().WithName("GINGERSNAP_K8S_RULE_CONFIG_MAP").WithValue(cacheService.EagerCacheConfigMap()),
								corev1.EnvVar().WithName("QUARKUS_LOG_CATEGORY__IO_QUARKUS_KUBERNETES_SERVICE_BINDING__LEVEL").WithValue("DEBUG"),
							).
							WithResources(
								corev1.ResourceRequirements().
									WithLimits(cache.DBSyncerLimits()).
									WithRequests(cache.DBSyncerRequests()),
							).
							WithVolumeMounts(
								corev1.VolumeMount().WithName("eager-rules").WithMountPath("/rules/eager").WithReadOnly(true),
							),
					).
					WithVolumes(
						corev1.Volume().
							WithName("eager-rules").
							WithConfigMap(
								corev1.ConfigMapVolumeSource().WithName(r.ConfigMap()).WithOptional(true),
							),
					),
				),
			),
		)

	if err := ctx.Client().Apply(deployment); err != nil {
		ctx.Requeue(fmt.Errorf("unable to apply DB-Syncer Deployment: %w", err))
	}
}

func RemoveDBSyncer(r *v1alpha1.EagerCacheRule, ctx *rule.Context) {
	cacheService := r.CacheService()
	labels := cacheService.LabelSelector()
	eagerCaches := &v1alpha1.EagerCacheRuleList{}
	if err := ctx.Client().List(labels, eagerCaches, client.ClusterScoped); err != nil {
		ctx.Requeue(fmt.Errorf("unable to list all EagerCacheRules to determine db-syncer lifecycle: %w", err))
		return
	}

	if len(eagerCaches.Items) == 1 && eagerCaches.Items[0].UID == r.UID {
		// Remove the db-syncer deployment as no other dependent EagerCacheRules exist
		if err := ctx.Client().Delete(cacheService.DBSyncerName(), &apiappsv1.Deployment{}); runtimeClient.IgnoreNotFound(err) != nil {
			ctx.Requeue(fmt.Errorf("unable to remove db-syncer: %w", err))
		}
	}
}
