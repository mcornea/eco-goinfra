package ibgu

import (
	"context"
	"fmt"

	"github.com/golang/glog"
	"github.com/openshift-kni/eco-goinfra/pkg/clients"
	"github.com/openshift-kni/eco-goinfra/pkg/msg"
	"github.com/openshift-kni/eco-goinfra/pkg/schemes/imagebasedgroupupgrades/v1alpha1"
	lcav1 "github.com/openshift-kni/lifecycle-agent/api/imagebasedupgrade/v1"
	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	goclient "sigs.k8s.io/controller-runtime/pkg/client"
)

// IbguBuilder provides struct for the ibgu object containing connection to
// the cluster and the ibgu definitions.
type IbguBuilder struct {
	// ibgu Definition, used to create the ibgu object.
	Definition *v1alpha1.ImageBasedGroupUpgrade
	// created ibgu object.
	Object *v1alpha1.ImageBasedGroupUpgrade
	// api client to interact with the cluster.
	apiClient goclient.Client
	// used to store latest error message upon defining or mutating application definition.
	errorMsg string
}

// NewIbguBuilder creates a new instance of IbguBuilder.
func NewIbguBuilder(
	apiClient *clients.Settings,
	name string,
	nsname string) *IbguBuilder {
	glog.V(100).Infof(
		"Initializing new ibgu structure with the following params: name: %s, nsname: %s", name, nsname)

	if apiClient == nil {
		glog.V(100).Info("The apiClient for the ibgu is nil")

		return nil
	}

	err := apiClient.AttachScheme(v1alpha1.AddToScheme)
	if err != nil {
		glog.V(100).Infof("Failed to add ibgu v1alpha1 scheme to client schemes")

		return nil
	}

	builder := &IbguBuilder{
		apiClient: apiClient.Client,
		Definition: &v1alpha1.ImageBasedGroupUpgrade{
			ObjectMeta: metav1.ObjectMeta{
				Name:      name,
				Namespace: nsname,
			},
			Spec: v1alpha1.ImageBasedGroupUpgradeSpec{
				IBUSpec: lcav1.ImageBasedUpgradeSpec{},
				Plan:    []v1alpha1.PlanItem{},
			},
		},
	}

	if name == "" {
		glog.V(100).Infof("The name of the ibgu is empty")

		builder.errorMsg = "ibgu 'name' cannot be empty"

		return builder
	}

	if nsname == "" {
		glog.V(100).Infof("The namespace of the ibgu is empty")

		builder.errorMsg = "ibgu 'nsname' cannot be empty"

		return builder
	}

	return builder
}

// WithClusterLabelSelectors appends labels to the ibgu clusterLabelSelectors.
func (builder *IbguBuilder) WithClusterLabelSelectors(labels map[string]string) *IbguBuilder {
	if valid, _ := builder.validate(); !valid {
		return builder
	}

	glog.V(100).Infof("Creating IBGU with %v cluster label selector", labels)

	if len(labels) == 0 {
		glog.V(100).Infof("The 'labels' of the IBGU is empty")

		builder.errorMsg = "can not apply empty cluster label selectors to the IBGU"

		return builder
	}

	labelSelectors := []metav1.LabelSelector{
		{
			MatchLabels: labels,
		},
	}

	builder.Definition.Spec.ClusterLabelSelectors = labelSelectors

	return builder
}

// WithSeedImageRef appends the SeedImageRef to the ibuSpec.
func (builder *IbguBuilder) WithSeedImageRef(seedImage string, seedVersion string) *IbguBuilder {
	if valid, _ := builder.validate(); !valid {
		return builder
	}

	glog.V(100).Infof("Creating IBGU with %s seed image and %s seed version", seedImage, seedVersion)

	if seedImage == "" {
		glog.V(100).Info("The 'seedImage' parameter is empty")

		builder.errorMsg = "seedImage cannot be empty"

		return builder
	}

	if seedVersion == "" {
		glog.V(100).Info("The 'seedVersion' parameter is empty")

		builder.errorMsg = "seedVersion cannot be empty"

		return builder
	}

	ibuSpec := lcav1.ImageBasedUpgradeSpec{
		SeedImageRef: lcav1.SeedImageRef{
			Image:   seedImage,
			Version: seedVersion,
		},
	}

	builder.Definition.Spec.IBUSpec = ibuSpec

	return builder
}

// WithOadpContent appends the oadpContent to the ibuSpec.
func (builder *IbguBuilder) WithOadpContent(name string, namespace string) *IbguBuilder {
	if valid, _ := builder.validate(); !valid {
		return builder
	}

	glog.V(100).Infof("Creating IBGU with OADP configmap %s in namespace %s", name, namespace)

	if name == "" {
		glog.V(100).Info("The 'name' parameter for OADP content is empty")

		builder.errorMsg = "oadp content name cannot be empty"

		return builder
	}

	if namespace == "" {
		glog.V(100).Info("The 'namespace' parameter for OADP content is empty")

		builder.errorMsg = "oadp content namespace cannot be empty"

		return builder
	}

	oadpContent := lcav1.ConfigMapRef{
		Name:      name,
		Namespace: namespace,
	}

	builder.Definition.Spec.IBUSpec.OADPContent = append(builder.Definition.Spec.IBUSpec.OADPContent, oadpContent)

	return builder
}

// WithPlan appends the plan to the ibgu.
func (builder *IbguBuilder) WithPlan(actions []string, maxConcurrency int, timeout int) *IbguBuilder {
	if valid, _ := builder.validate(); !valid {
		return builder
	}

	glog.V(100).Infof(
		"Creating IBGU with plan actions %v, maxConcurrency %d and timeout %d",
		actions,
		maxConcurrency,
		timeout,
	)

	if len(actions) == 0 {
		glog.V(100).Info("The 'actions' slice is empty")

		builder.errorMsg = "plan actions cannot be empty"

		return builder
	}

	if maxConcurrency <= 0 {
		glog.V(100).Infof("Invalid maxConcurrency value: %d", maxConcurrency)

		builder.errorMsg = "maxConcurrency must be greater than 0"

		return builder
	}

	if timeout <= 0 {
		glog.V(100).Infof("Invalid timeout value: %d", timeout)

		builder.errorMsg = "timeout must be greater than 0"

		return builder
	}

	plan := v1alpha1.PlanItem{
		Actions: actions,
		RolloutStrategy: v1alpha1.RolloutStrategy{
			MaxConcurrency: maxConcurrency,
			Timeout:        timeout,
		},
	}

	builder.Definition.Spec.Plan = append(builder.Definition.Spec.Plan, plan)

	return builder
}

// Get returns imagebasedgroupupgrade object if found.
func (builder *IbguBuilder) Get() (*v1alpha1.ImageBasedGroupUpgrade, error) {
	if valid, err := builder.validate(); !valid {
		return nil, err
	}

	glog.V(100).Infof("Getting imagebasedgroupupgrade %s",
		builder.Definition.Name)

	imagebasedgroupupgrade := &v1alpha1.ImageBasedGroupUpgrade{}

	err := builder.apiClient.Get(context.TODO(), goclient.ObjectKey{
		Name:      builder.Definition.Name,
		Namespace: builder.Definition.Namespace,
	}, imagebasedgroupupgrade)

	if err != nil {
		return nil, err
	}

	return imagebasedgroupupgrade, err
}

// Exists checks whether the given imagebasedgroupupgrade exists.
func (builder *IbguBuilder) Exists() bool {
	if valid, _ := builder.validate(); !valid {
		return false
	}

	glog.V(100).Infof("Checking if imagebasedgroupupgrade %s exists",
		builder.Definition.Name)

	var err error
	builder.Object, err = builder.Get()

	return err == nil || !k8serrors.IsNotFound(err)
}

// Create makes an IBGU in the cluster and stores the created object in struct.
func (builder *IbguBuilder) Create() (*IbguBuilder, error) {
	if valid, err := builder.validate(); !valid {
		return builder, err
	}

	glog.V(100).Infof("Creating the imageasedgroupupgrade %s in namespace %s",
		builder.Definition.Name, builder.Definition.Namespace)

	var err error
	if !builder.Exists() {
		err = builder.apiClient.Create(context.TODO(), builder.Definition)
		if err == nil {
			builder.Object = builder.Definition
		}
	}

	return builder, err
}

// Delete removes the IBGU from the cluster.
func (builder *IbguBuilder) Delete() error {
	if valid, err := builder.validate(); !valid {
		return err
	}

	glog.V(100).Infof("Deleting the ImageBasedGroupUpgrade %s in namespace %s",
		builder.Definition.Name, builder.Definition.Namespace)

	if !builder.Exists() {
		builder.Object = nil

		return nil
	}

	err := builder.apiClient.Delete(context.TODO(), builder.Object)

	if err != nil {
		return err
	}

	builder.Object = nil

	return nil
}

// validate will check that the builder and builder definition are properly initialized before
// accessing any member fields.
func (builder *IbguBuilder) validate() (bool, error) {
	resourceCRD := "ibgu"

	if builder == nil {
		glog.V(100).Infof("The %s builder is uninitialized", resourceCRD)

		return false, fmt.Errorf("error: received nil %s builder", resourceCRD)
	}

	if builder.Definition == nil {
		glog.V(100).Infof("The %s is undefined", resourceCRD)

		return false, fmt.Errorf(msg.UndefinedCrdObjectErrString(resourceCRD))
	}

	if builder.apiClient == nil {
		glog.V(100).Infof("The %s builder apiclient is nil", resourceCRD)

		return false, fmt.Errorf("%s builder cannot have nil apiClient", resourceCRD)
	}

	if builder.errorMsg != "" {
		glog.V(100).Infof("The %s builder has error message: %s", resourceCRD, builder.errorMsg)

		return false, fmt.Errorf(builder.errorMsg)
	}

	return true, nil
}
