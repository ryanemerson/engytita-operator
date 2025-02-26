package v1alpha1

import (
	"errors"
	"time"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
)

var _ = Describe("Cache Webhooks", func() {

	const timeout = time.Second * 30
	const interval = time.Second * 1

	key := types.NamespacedName{
		Name:      "cache-envtest",
		Namespace: "default",
	}

	AfterEach(func() {
		// Delete created resources
		By("Expecting to delete successfully")
		Eventually(func() error {
			f := &Cache{}
			if err := k8sClient.Get(ctx, key, f); err != nil {
				var statusError *apierrors.StatusError
				if !errors.As(err, &statusError) {
					return err
				}
				// If the resource does not exist, do nothing
				if statusError.ErrStatus.Code == 404 {
					return nil
				}
			}
			return k8sClient.Delete(ctx, f)
		}, timeout, interval).Should(Succeed())

		By("Expecting to delete finish")
		Eventually(func() error {
			f := &Cache{}
			return k8sClient.Get(ctx, key, f)
		}, timeout, interval).ShouldNot(Succeed())
	})

	It("should correctly set Local Cache defaults", func() {

		created := &Cache{
			ObjectMeta: metav1.ObjectMeta{
				Name:      key.Name,
				Namespace: key.Namespace,
			},
			Spec: CacheSpec{
				DataSource: &DataSourceSpec{
					DbType: DBType_POSTGRES_14.Enum(),
					SecretRef: &LocalObjectReference{
						Name: "some-secret",
					},
				},
			},
		}

		Expect(k8sClient.Create(ctx, created)).Should(Succeed())
		Expect(k8sClient.Get(ctx, key, created)).Should(Succeed())
		Expect(created.Spec.Deployment.Type).Should(Equal(CacheDeploymentType_LOCAL))
		Expect(created.Spec.Deployment.Replicas).Should(Equal(int32(0)))
	})

	It("should correctly set Cluster Cache defaults", func() {

		created := &Cache{
			ObjectMeta: metav1.ObjectMeta{
				Name:      key.Name,
				Namespace: key.Namespace,
			},
			Spec: CacheSpec{
				Deployment: &CacheDeploymentSpec{
					Type: CacheDeploymentType_CLUSTER,
				},
				DataSource: &DataSourceSpec{
					DbType: DBType_POSTGRES_14.Enum(),
					SecretRef: &LocalObjectReference{
						Name: "some-secret",
					},
				},
			},
		}

		Expect(k8sClient.Create(ctx, created)).Should(Succeed())
		Expect(k8sClient.Get(ctx, key, created)).Should(Succeed())
		Expect(created.Spec.Deployment.Type).Should(Equal(CacheDeploymentType_CLUSTER))
		Expect(created.Spec.Deployment.Replicas).Should(Equal(int32(1)))
	})

	It("should reject missing datasource", func() {
		invalid := &Cache{
			ObjectMeta: metav1.ObjectMeta{
				Name:      key.Name,
				Namespace: key.Namespace,
			},
			Spec: CacheSpec{},
		}
		ExpectInvalidErrStatus(
			k8sClient.Create(ctx, invalid),
			statusDetailCause{metav1.CauseTypeFieldValueRequired, "spec.dataSource", "A dataSource must be defined"},
		)
	})

	It("should reject missing datasource.dbType", func() {
		invalid := &Cache{
			ObjectMeta: metav1.ObjectMeta{
				Name:      key.Name,
				Namespace: key.Namespace,
			},
			Spec: CacheSpec{
				DataSource: &DataSourceSpec{
					SecretRef: &LocalObjectReference{
						Name: "some-secret",
					},
				},
			},
		}
		ExpectInvalidErrStatus(
			k8sClient.Create(ctx, invalid),
			statusDetailCause{metav1.CauseTypeFieldValueRequired, "spec.dataSource.dbType", "A dataSource dbType must be defined"},
		)
	})

	It("should reject missing datasource secret or service ref", func() {
		invalid := &Cache{
			ObjectMeta: metav1.ObjectMeta{
				Name:      key.Name,
				Namespace: key.Namespace,
			},
			Spec: CacheSpec{
				DataSource: &DataSourceSpec{
					DbType: DBType_POSTGRES_14.Enum(),
				},
			},
		}
		ExpectInvalidErrStatus(
			k8sClient.Create(ctx, invalid),
			statusDetailCause{metav1.CauseTypeFieldValueRequired, "spec.dataSource", "'secretRef' OR 'serviceProviderRef' must be supplied"},
		)
	})

	It("should reject duplicate datasource secret and service ref", func() {
		invalid := &Cache{
			ObjectMeta: metav1.ObjectMeta{
				Name:      key.Name,
				Namespace: key.Namespace,
			},
			Spec: CacheSpec{
				DataSource: &DataSourceSpec{
					DbType: DBType_POSTGRES_14.Enum(),
					SecretRef: &LocalObjectReference{
						Name: "some-secret",
					},
					ServiceProviderRef: &ServiceRef{
						ApiVersion: "acme.org/v1alpha1",
						Kind:       "Example",
						Name:       "example-instance",
					},
				},
			},
		}
		ExpectInvalidErrStatus(
			k8sClient.Create(ctx, invalid),
			statusDetailCause{metav1.CauseTypeFieldValueDuplicate, "spec.dataSource", "At most one of ['secretRef', 'serviceProviderRef'] must be configured"},
		)
	})

	It("should reject empty datasource secretRef fields", func() {
		invalid := &Cache{
			ObjectMeta: metav1.ObjectMeta{
				Name:      key.Name,
				Namespace: key.Namespace,
			},
			Spec: CacheSpec{
				DataSource: &DataSourceSpec{
					DbType: DBType_POSTGRES_14.Enum(),
					SecretRef: &LocalObjectReference{
						Name: "",
					},
				},
			},
		}
		ExpectInvalidErrStatus(
			k8sClient.Create(ctx, invalid),
			statusDetailCause{metav1.CauseTypeFieldValueRequired, "spec.dataSource.secretRef.name", "'name' field must not be empty"},
		)
	})

	It("should reject empty datasource serviceProviderRef fields", func() {
		invalid := &Cache{
			ObjectMeta: metav1.ObjectMeta{
				Name:      key.Name,
				Namespace: key.Namespace,
			},
			Spec: CacheSpec{
				DataSource: &DataSourceSpec{
					DbType: DBType_POSTGRES_14.Enum(),
					ServiceProviderRef: &ServiceRef{
						ApiVersion: "",
						Kind:       "",
						Name:       "",
					},
				},
			},
		}
		ExpectInvalidErrStatus(
			k8sClient.Create(ctx, invalid),
			statusDetailCause{metav1.CauseTypeFieldValueRequired, "spec.dataSource.serviceProviderRef.apiVersion", "'apiVersion' field must not be empty"},
			statusDetailCause{metav1.CauseTypeFieldValueRequired, "spec.dataSource.serviceProviderRef.kind", "'kind' field must not be empty"},
			statusDetailCause{metav1.CauseTypeFieldValueRequired, "spec.dataSource.serviceProviderRef.name", "'name' field must not be empty"},
		)
	})

	It("should reject invalid resource quantities", func() {

		valid := &Cache{
			ObjectMeta: metav1.ObjectMeta{
				Name:      key.Name,
				Namespace: key.Namespace,
			},
			Spec: CacheSpec{
				Deployment: &CacheDeploymentSpec{
					Resources: &Resources{
						Requests: &ResourceQuantity{
							Cpu:    "0.1",
							Memory: "512Mi",
						},
						Limits: &ResourceQuantity{
							Cpu:    "1",
							Memory: "512Mi",
						},
					},
				},
				DbSyncer: &DBSyncerDeploymentSpec{
					Resources: &Resources{
						Requests: &ResourceQuantity{
							Cpu:    "0.1",
							Memory: "512Mi",
						},
						Limits: &ResourceQuantity{
							Cpu:    "1",
							Memory: "512Mi",
						},
					},
				},
				DataSource: &DataSourceSpec{
					DbType: DBType_POSTGRES_14.Enum(),
					SecretRef: &LocalObjectReference{
						Name: "some-secret",
					},
				},
			},
		}

		Expect(k8sClient.Create(ctx, valid)).Should(Succeed())

		invalid := valid.DeepCopy()
		invalid.Spec.Deployment.Resources.Requests.Cpu = "regex fail"
		invalid.Spec.Deployment.Resources.Requests.Memory = "512mi"
		invalid.Spec.Deployment.Resources.Limits.Cpu = "regex fail"
		invalid.Spec.Deployment.Resources.Limits.Memory = "1a"
		invalid.Spec.DbSyncer.Resources.Requests.Cpu = "regex fail"
		invalid.Spec.DbSyncer.Resources.Requests.Memory = "512mi"
		invalid.Spec.DbSyncer.Resources.Limits.Cpu = "regex fail"
		invalid.Spec.DbSyncer.Resources.Limits.Memory = "1a"

		ExpectInvalidErrStatus(k8sClient.Create(ctx, invalid),
			statusDetailCause{metav1.CauseTypeFieldValueInvalid, "spec.deployment.resources.requests.cpu", "quantities must match the regular expression"},
			statusDetailCause{metav1.CauseTypeFieldValueInvalid, "spec.deployment.resources.requests.memory", "unable to parse quantity's suffix"},
			statusDetailCause{metav1.CauseTypeFieldValueInvalid, "spec.deployment.resources.limits.cpu", "quantities must match the regular expression"},
			statusDetailCause{metav1.CauseTypeFieldValueInvalid, "spec.deployment.resources.limits.memory", "quantities must match the regular expression"},
			statusDetailCause{metav1.CauseTypeFieldValueInvalid, "spec.dbSyncer.resources.requests.cpu", "quantities must match the regular expression"},
			statusDetailCause{metav1.CauseTypeFieldValueInvalid, "spec.dbSyncer.resources.requests.memory", "unable to parse quantity's suffix"},
			statusDetailCause{metav1.CauseTypeFieldValueInvalid, "spec.dbSyncer.resources.limits.cpu", "quantities must match the regular expression"},
			statusDetailCause{metav1.CauseTypeFieldValueInvalid, "spec.dbSyncer.resources.limits.memory", "quantities must match the regular expression"},
		)
	})
})
