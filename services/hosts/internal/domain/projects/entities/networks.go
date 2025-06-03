package entities

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

type Network struct {
	OwnedVlans        []int      `bson:"owned_vlans"`
	VlanScheme        VlanScheme `bson:"vlan_scheme"`
	NativeVlan        int        `bson:"native_vlan"`
	ExtraVlans        []int      `bson:"extra_vlans"`
	DNSDomain         string     `bson:"dns_domain"`
	ShortnameTemplate string     `bson:"shortname_template"`
	YcDNSZoneID       string     `bson:"yc_dns_zone_id"`
	YcIAMFolderID     string     `bson:"yc_iam_folder_id"`
	HbfProjectID      int        `bson:"hbf_project_id"`
}
