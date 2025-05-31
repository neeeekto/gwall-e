package models

// DeploymentPolicy defines deployment policies for projects
type DeploymentPolicy string

const (
	DeploymentPolicyPassthrough            DeploymentPolicy = "PASSTHROUGH"
	DeploymentPolicyDiskmanager            DeploymentPolicy = "DISKMANAGER"
	DeploymentPolicyShared                 DeploymentPolicy = "SHARED"
	DeploymentPolicySharedCryptsetup       DeploymentPolicy = "SHARED_CRYPTSETUP"
	DeploymentPolicySharedStripedNvme      DeploymentPolicy = "SHARED_STRIPED_NVME"
	DeploymentPolicySharedLVM              DeploymentPolicy = "SHAREDLVM"
	DeploymentPolicySharedLVMExtendedPlace DeploymentPolicy = "SHAREDLVM_EXTENDED_PLACE"
	DeploymentPolicyYTDedicated            DeploymentPolicy = "YT_DEDICATED"
	DeploymentPolicyYTDedicatedTest        DeploymentPolicy = "YT_DEDICATED_TEST"
	DeploymentPolicyYTShared               DeploymentPolicy = "YT_SHARED"
	DeploymentPolicyYTMasters              DeploymentPolicy = "YT_MASTERS"
	DeploymentPolicyYTMastersK8s           DeploymentPolicy = "YT_MASTERS_K8S"
	DeploymentPolicyYTStorage              DeploymentPolicy = "YT_STORAGE"
	DeploymentPolicyYTDefaultShared        DeploymentPolicy = "YT_DEFAULT_SHARED"
	DeploymentPolicyYTDefaultStorage       DeploymentPolicy = "YT_DEFAULT_STORAGE"
	DeploymentPolicyYTSacrifice            DeploymentPolicy = "YT_SACRIFICE"
	DeploymentPolicyMDSDedicated           DeploymentPolicy = "MDS_DEDICATED"
)

// Secret represents a deployment secret (structure to be defined later)
type Secret struct {
	// TODO: Add fields when structure is known
}

// ProjectDeploying contains deployment configuration for a project
type ProjectDeploying struct {
	// Config holds deployment configuration data
	Config interface{} `bson:"config"`

	// Tags contains deployment tags
	Tags []string `bson:"tags"`

	// Network specifies network configuration
	Network string `bson:"network"`

	// Policy defines deployment policy
	Policy DeploymentPolicy `bson:"policy"`

	// Secrets contains deployment secrets
	Secrets []Secret `bson:"secrets"`
}