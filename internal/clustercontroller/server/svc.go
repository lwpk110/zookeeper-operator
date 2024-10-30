package server

import (
	"context"

	zkv1alpha1 "github.com/zncdatadev/zookeeper-operator/api/v1alpha1"
	"github.com/zncdatadev/zookeeper-operator/internal/security"
	corev1 "k8s.io/api/core/v1"
	ctrlclient "sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/zncdatadev/operator-go/pkg/builder"
	"github.com/zncdatadev/operator-go/pkg/client"
	"github.com/zncdatadev/operator-go/pkg/constants"
	"github.com/zncdatadev/operator-go/pkg/reconciler"
)

var _ builder.ServiceBuilder = &ServiceBuilder{}

type ServiceBuilder struct {
	builder.BaseServiceBuilder
}

func (b *ServiceBuilder) Build(_ context.Context) (ctrlclient.Object, error) {
	obj := b.GetObject()

	// !!! WARNING: !!! This is a workaround for the fact that wether the statefulset headless service return endpoints until the pod is ready
	// Set publishNotReadyAddresses to true to allow the service to return endpoints before the pod is ready
	// So statefulset headless service to to propagate SRV DNS records for its Pods for the purpose of peer discovery
	obj.Spec.PublishNotReadyAddresses = true

	return obj, nil
}

func NewServiceReconciler(
	client *client.Client,
	option *reconciler.RoleGroupInfo,
	listenerClass string,
	zkSecurity *security.ZookeeperSecurity,
) *reconciler.Service {
	ports := []corev1.ContainerPort{
		{
			Name:          zkv1alpha1.ClientPortName,
			ContainerPort: int32(zkSecurity.ClientPort()),
		},
		{
			Name:          zkv1alpha1.MetricsPortName,
			ContainerPort: int32(zkv1alpha1.MetricsPort),
		},
	}

	// :bug: should fix this issue in operator-go
	// when service type is nodeport, cluster ip should not be "None", so when listener class is external-unstable, headless should be false
	headless := true
	if listenerClass == string(constants.ExternalUnstable) {
		headless = false
	}

	svcBuilder := &ServiceBuilder{
		BaseServiceBuilder: *builder.NewServiceBuilder(
			client,
			option.GetFullName(),
			ports,
			func(s *builder.ServiceBuilderOption) {
				s.Labels = option.GetLabels()
				s.Annotations = option.GetAnnotations()
				s.Headless = headless
				s.ListenerClass = constants.ListenerClass(listenerClass)
			},
		),
	}

	return &reconciler.Service{
		GenericResourceReconciler: *reconciler.NewGenericResourceReconciler[builder.ServiceBuilder](
			client,
			option.GetFullName(),
			svcBuilder,
		),
	}
}