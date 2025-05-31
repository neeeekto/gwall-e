package entities

// Secret represents a deployment secret (structure to be defined later)
type Secret struct {
	// TODO: Add fields when structure is known
}

// ProjectDeploying contains deployment configuration for a project
type Deploying struct {
	// Config holds deployment configuration data
	Config interface{} `bson:"config"`

	// Tags contains deployment tags
	Tags []string `bson:"tags"`

	// Network specifies network configuration
	Network string `bson:"network"`

	// Policy defines deployment policy (possible values: PASSTHROUGH, DISKMANAGER, SHARED,
	// SHARED_CRYPTSETUP, SHARED_STRIPED_NVME, SHAREDLVM, SHAREDLVM_EXTENDED_PLACE,
	// YT_DEDICATED, YT_DEDICATED_TEST, YT_SHARED, YT_MASTERS, YT_MASTERS_K8S, YT_STORAGE,
	// YT_DEFAULT_SHARED, YT_DEFAULT_STORAGE, YT_SACRIFICE, MDS_DEDICATED)
	Policy string `bson:"policy"`

	// Secrets contains deployment secrets
	Secrets []Secret `bson:"secrets"`
}