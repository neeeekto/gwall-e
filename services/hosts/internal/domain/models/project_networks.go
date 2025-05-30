package models

// VlanScheme represents VLAN allocation scheme
type VlanScheme string

const (
	VlanSchemeStatic             VlanScheme = "STATIC"
	VlanSchemeMTN                VlanScheme = "MTN"
	VlanSchemeMTNHostID          VlanScheme = "MTN_HOSTID"
	VlanSchemeCloud              VlanScheme = "CLOUD"
	VlanSchemeMock               VlanScheme = "MOCK"
	VlanSchemeMTNWithoutFastBone VlanScheme = "MTN_WITHOUT_FASTBONE"
)

type ProjectNetwork struct {
	OwnedVlans []int      `bson:"owned_vlans"`
	VlanScheme VlanScheme `bson:"vlan_scheme"`
	NativeVlan int        `bson:"native_vlan"`
	ExtraVlans []int      `bson:"extra_vlans"`
	DNSDomain  string     `bson:"dns_domain"`
}
