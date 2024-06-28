package interfaces

import "context"

type KubeHelper interface {
	ClusterNodesUpdated(ctx context.Context, nodes []*NodeInfo) error
}
