package infinispan

import (
	"fmt"

	"github.com/engytita/engytita-operator/api/v1alpha1"
	"github.com/engytita/engytita-operator/pkg/kubernetes"
	"github.com/engytita/engytita-operator/pkg/kubernetes/client"
	"github.com/engytita/engytita-operator/pkg/reconcile"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const (
	containerName = "infinispan"
)

var (
	selectorLabels = map[string]string{
		"app": "infinispan",
	}
)

func Service(c *v1alpha1.Cache, ctx reconcile.Context) {
	svc := &corev1.Service{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "v1",
			Kind:       "Service",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      c.Name,
			Namespace: c.Namespace,
		},
	}

	mutateFn := func() error {
		svc.Spec.Type = corev1.ServiceTypeClusterIP
		svc.Spec.ClusterIP = corev1.ClusterIPNone
		svc.Spec.Selector = selectorLabels
		// We must utilise the existing ServicePort values if updating the service, to prevent the created ports being overwritten
		if svc.CreationTimestamp.IsZero() {
			svc.Spec.Ports = []corev1.ServicePort{{}}
		}
		servicePort := &svc.Spec.Ports[0]
		servicePort.Name = "infinispan"
		servicePort.Port = 11222
		return nil
	}
	res, err := ctx.Client().CreateOrPatch(svc, mutateFn, client.SetControllerRef)
	ctx.Log().Info(fmt.Sprintf("Res=%s,Err=%s", res, err))
	if err != nil {
		ctx.Requeue(fmt.Errorf("unable to CreateOrPatch Infinispan Service: %w", err))
	}
}

func ConfigMap(c *v1alpha1.Cache, ctx reconcile.Context) {
	cm := &corev1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Name:      c.Name,
			Namespace: c.Namespace,
		},
	}

	config := `
    <infinispan
      xmlns:xsi="http://www.w3.org/2001/XMLSchema-instance"
      xsi:schemaLocation="urn:infinispan:config:13.0 https://infinispan.org/schemas/infinispan-config-13.0.xsd
                            urn:infinispan:server:13.0 https://infinispan.org/schemas/infinispan-server-13.0.xsd"
      xmlns="urn:infinispan:config:13.0"
      xmlns:server="urn:infinispan:server:13.0">

      <cache-container name="default" statistics="true">
          <local-cache name="airports" />
          <transport cluster="${infinispan.cluster.name:cluster}" stack="${infinispan.cluster.stack:tcp}" node-name="${infinispan.node.name:}"/>
          <security>
            <authorization/>
          </security>
      </cache-container>

      <server xmlns="urn:infinispan:server:13.0">
          <interfaces>
            <interface name="public">
                <inet-address value="${infinispan.bind.address:127.0.0.1}"/>
            </interface>
          </interfaces>

          <socket-bindings default-interface="public" port-offset="${infinispan.socket.binding.port-offset:0}">
            <socket-binding name="default" port="${infinispan.bind.port:11222}"/>
            <socket-binding name="memcached" port="11221"/>
          </socket-bindings>

          <security>
            <credential-stores>
                <credential-store name="credentials" path="credentials.pfx">
                  <clear-text-credential clear-text="secret"/>
                </credential-store>
            </credential-stores>
            <security-realms>
                <security-realm name="default">
                  <!-- Uncomment to enable TLS on the realm -->
                  <!-- server-identities>
                      <ssl>
                        <keystore path="application.keystore"
                                  password="password" alias="server"
                                  generate-self-signed-certificate-host="localhost"/>
                      </ssl>
                  </server-identities-->
                  <properties-realm groups-attribute="Roles">
                      <user-properties path="users.properties"/>
                      <group-properties path="groups.properties"/>
                  </properties-realm>
                </security-realm>
            </security-realms>
          </security>

          <endpoints socket-binding="default" security-realm="default" />
      </server>
    </infinispan>
`

	mutateFn := func() error {
		cm.Data = map[string]string{
			"infinispan.xml": config,
		}
		return nil
	}

	_, err := ctx.Client().CreateOrPatch(cm, mutateFn, client.SetControllerRef)
	if err != nil {
		ctx.Requeue(fmt.Errorf("unable to CreateOrPatch Infinispan ConfigMap: %w", err))
	}
}

func DaemonSet(c *v1alpha1.Cache, ctx reconcile.Context) {
	ds := &appsv1.DaemonSet{
		ObjectMeta: metav1.ObjectMeta{
			Name:      c.Name,
			Namespace: c.Namespace,
		},
	}

	mutateFn := func() error {
		ds.Spec.Selector.MatchLabels = selectorLabels
		ds.Spec.Template.ObjectMeta.Labels = selectorLabels

		createDs := ds.CreationTimestamp.IsZero()

		// Configure container
		var container *corev1.Container
		if createDs {
			container = &corev1.Container{}
		} else {
			kubernetes.GetContainer(containerName, &ds.Spec.Template.Spec)
		}

		container.Name = containerName
		container.Image = "quay.io/infinispan/server:13.0"
		container.Args = []string{"-c", "/config/infinispan.xml"}
		container.Ports = []corev1.ContainerPort{{ContainerPort: 11222}}
		container.Env = []corev1.EnvVar{
			{Name: "USER", Value: "admin"},
			{Name: "PASS", Value: "password"},
		}
		container.VolumeMounts = []corev1.VolumeMount{{
			Name:      "config",
			MountPath: "/config",
			ReadOnly:  true,
		}}
		if createDs {
			ds.Spec.Template.Spec.Containers = []corev1.Container{*container}
		}

		// Configure Volumes
		ds.Spec.Template.Spec.Volumes = []corev1.Volume{{
			Name: "config",
			VolumeSource: corev1.VolumeSource{
				ConfigMap: &corev1.ConfigMapVolumeSource{
					LocalObjectReference: corev1.LocalObjectReference{
						Name: c.Name,
					},
				},
			},
		}}
		return nil
	}

	_, err := ctx.Client().CreateOrPatch(ds, mutateFn, client.SetControllerRef)
	if err != nil {
		ctx.Requeue(fmt.Errorf("unable to CreateOrPatch Infinispan DaemonSet: %w", err))
	}
}
