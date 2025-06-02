package entities

type CMSAuth struct {
	Type  string `bson:"type"`
	Value string `bson:"value"`
}

type CMS struct {
	Enabled      bool    `bson:"enabled"`
	Version      string  `bson:"version"`
	MaxBusyHosts int     `bson:"max_busy_hosts"`
	Auth         CMSAuth `bson:"auth"`
}
