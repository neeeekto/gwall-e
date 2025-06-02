package entities

type Secret struct {
	// TODO: Add fields when structure is known
}

type Deploying struct {
	Config  string   `bson:"config"`
	Tags    []string `bson:"tags"`
	Network string   `bson:"network"`

	// Policy defines deployment policy (possible values: PASSTHROUGH, DISKMANAGER, SHARED,
	// SHARED_CRYPTSETUP, SHARED_STRIPED_NVME, SHAREDLVM, SHAREDLVM_EXTENDED_PLACE,
	// YT_DEDICATED, YT_DEDICATED_TEST, YT_SHARED, YT_MASTERS, YT_MASTERS_K8S, YT_STORAGE,
	// YT_DEFAULT_SHARED, YT_DEFAULT_STORAGE, YT_SACRIFICE, MDS_DEDICATED)
	Policy            string   `bson:"policy"`
	Secrets           []Secret `bson:"secrets"`
	DeployCertificate bool     `bson:"deploy_certificate"`
}
