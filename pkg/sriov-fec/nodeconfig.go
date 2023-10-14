package sriovfec

import (
	"github.com/golang/glog"
	"github.com/openshift-kni/eco-goinfra/pkg/clients"
	sriovfecV1 "github.com/smart-edge-open/sriov-fec-operator/api/v1"
)

// NodeConfigBuilder provides struct for SriovFecNodeConfig object which contains connection to cluster and
// SriovFecNodeConfig definitions.
type NodeConfigBuilder struct {
	// Dynamically discovered SriovFecNodeConfig object.
	Objects *sriovfecV1.SriovFecNodeConfig
	// apiClient opens api connection to the cluster.
	apiClient *clients.Settings
	// nodeName defines on what node SriovFecNodeConfig resource should be queried.
	nodeName string
	// nsName defines SriovFec operator namespace.
	nsName string
	// errorMsg used in discovery function before sending api request to cluster.
	errorMsg string
}

// NewNodeConfigBuilder creates new instance of NetworkNodeStateBuilder.
func NewNodeConfigBuilder(apiClient *clients.Settings, nodeName, nsname string) *NodeConfigBuilder {
	glog.V(100).Infof(
		"Initializing new NodeConfigBuilder structure with the following params: %s, %s",
		nodeName, nsname)

	builder := &NodeConfigBuilder{
		apiClient: apiClient,
		nodeName:  nodeName,
		nsName:    nsname,
	}

	if nodeName == "" {
		glog.V(100).Infof("The name of the nodeName is empty")

		builder.errorMsg = "SriovFecNodeConfig 'nodeName' is empty"
	}

	if nsname == "" {
		glog.V(100).Infof("The namespace of the SriovFecNodeConfig is empty")

		builder.errorMsg = "SriovFecNodeConfig 'nsname' is empty"
	}

	return builder
}
