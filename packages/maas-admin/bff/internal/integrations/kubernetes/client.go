package kubernetes

import (
	"context"

	"github.com/opendatahub-io/gen-ai/internal/integrations"
	corev1 "k8s.io/api/core/v1"
)

const ComponenetLabelValue = "maas-admin"

type KubernetesClientInterface interface {
	// Namespace access
	GetNamespaces(ctx context.Context, identity *integrations.RequestIdentity) ([]corev1.Namespace, error)
	CanListNamespaces(ctx context.Context, identity *integrations.RequestIdentity) (bool, error)

	// Meta
	IsClusterAdmin(ctx context.Context, identity *integrations.RequestIdentity) (bool, error)
	BearerToken() (string, error)

	// Identity
	GetUser(ctx context.Context, identity *integrations.RequestIdentity) (string, error)
}
