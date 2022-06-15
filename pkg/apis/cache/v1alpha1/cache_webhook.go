package v1alpha1

import (
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/webhook"
)

// log is for logging in this package.
var cachelog = logf.Log.WithName("cache-resource")

func (c *Cache) SetupWebhookWithManager(mgr ctrl.Manager) error {
	return ctrl.NewWebhookManagedBy(mgr).
		For(c).
		Complete()
}

// TODO(user): EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!

//+kubebuilder:webhook:path=/mutate-engytita-org-v1alpha1-cache,mutating=true,failurePolicy=fail,sideEffects=None,groups=engytita.org,resources=caches,verbs=create;update,versions=v1alpha1,name=mcache.kb.io,admissionReviewVersions=v1

var _ webhook.Defaulter = &Cache{}

// Default implements webhook.Defaulter so a webhook will be registered for the type
func (c *Cache) Default() {
	cachelog.Info("default", "name", c.Name)

	// TODO default to Infinispan or Redis spec
}

// TODO(user): change verbs to "verbs=create;update;delete" if you want to enable deletion validation.
//+kubebuilder:webhook:path=/validate-engytita-org-v1alpha1-cache,mutating=false,failurePolicy=fail,sideEffects=None,groups=engytita.org,resources=caches,verbs=create;update,versions=v1alpha1,name=vcache.kb.io,admissionReviewVersions=v1

var _ webhook.Validator = &Cache{}

// ValidateCreate implements webhook.Validator so a webhook will be registered for the type
func (c *Cache) ValidateCreate() error {
	cachelog.Info("validate create", "name", c.Name)
	return c.validate()
}

// ValidateUpdate implements webhook.Validator so a webhook will be registered for the type
func (c *Cache) ValidateUpdate(old runtime.Object) error {
	cachelog.Info("validate update", "name", c.Name)
	return c.validate()
}

// ValidateDelete implements webhook.Validator so a webhook will be registered for the type
func (c *Cache) ValidateDelete() error {
	cachelog.Info("validate delete", "name", c.Name)

	// TODO(user): fill in your validation logic upon object deletion.
	return nil
}

func (c *Cache) validate() error {
	// TODO add validation to ensure that only Infinispan or Redis spec defined
	return nil
}
