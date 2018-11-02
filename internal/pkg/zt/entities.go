package zt

import (
	"github.com/nalej/derrors"
	"github.com/nalej/dhttp"
	"github.com/nalej/network-manager/internal/pkg/entities"
	"time"
)

type ZTNetwork struct {
	// 16-digit ZeroTier network ID [ro]
	ID string `json:"id,omitempty"`
	// 16-digit ZeroTier network ID (for backward compatibility) [ro]
	Nwid string `json:"nwid,omitempty"`
	// Current clock, ms since epoch [ro]
	Clock int `json:"clock,omitempty"`
	// Short name of network [rw]
	Name string `json:"name"`
	// Object type on controller ("network") [ro]
	ObjType string `json:"objtype,omitempty"`
	// If true, certificate access control is enabled [rw]
	Private *bool `json:"private,omitempty"`
	// Ethernet ff:ff:ff:ff:ff:ff allowed? [rw].
	EnableBroadcast *bool `json:"enableBroadcast,omitempty"`
	// Allow any member to bridge (very experimental) [rw]
	AllowPassiveBridging *bool `json:"allowPassiveBridging,omitempty"`
	// IPv4 management and assign options (see below) [rw]
	V4AssignMode *V4AssignMode `json:"v4AssignMode,omitempty"`
	// IPv6 management and assign options (see below) [rw]
	V6AssignMode *V6AssignMode `json:"v6AssignMode,omitempty"`
	// Maximum recipients for a multicast packet [rw]
	MulticastLimit *int `json:"multicastLimit,omitempty"`
	// Time network was first created [ro]
	CreationTime int `json:"creationTime,omitempty"`
	// Network config revision counter [ro]
	Revision *int `json:"revision,omitempty"`
	// Time config was last modified [ro]
	LastModified int `json:"lastModified,omitempty"`
	// Number of authorized members (for private nets) [ro]
	AuthorizedMemberCount *int `json:"authorizedMemberCount,omitempty"`
	// Number of members that appear to be online [ro]
	ActiveMemberCount *int `json:"activeMemberCount,omitempty"`
	// Total known members of this network [ro]
	TotalMemberCount *int `json:"totalMemberCount,omitempty"`
	// IPv4 and IPv6 routes; see below [rw]
	Routes []Route `json:"routes,omitempty"`
	// IP auto-assign ranges; see below [rw]
	IpAssignmentPools []IpAssignmentPool `json:"ipAssignmentPools,omitempty"`
	// Traffic rules; see below [rw]
	Rules []Rule `json:"rules,omitempty"`
}

func (n *ZTNetwork) ToNetwork (organizationId string) entities.Network {
	return entities.Network {
		OrganizationId: organizationId,
		NetworkId: n.ID,
		NetworkName: n.Name,
		CreationTimestamp: time.Now().Unix(),
	}
}


type PeerNC struct {
	client dhttp.Client
}

func (p *PeerNC) GetStatus() (*PeerStatus, derrors.Error) {
	return nil, nil
}

type V4AssignMode struct {
	// IPv4 addresses to be auto-assigned from ipAssignmentPools
	Zt bool `json:"zt"`
}

type V6AssignMode struct {
	// IPv6 addresses to be auto-assigned from ipAssignmentPools
	Zt bool `json:"zt"`
	// 6plane gives every member a /80 within a /40 network
	SixPlane bool `json:"6plane"`
	// rfc4193 mode gives every member a /128 on a /88 network
	Rfc4193 bool `json:"rfc4193"`
}

type Route struct {
	Target string `json:"target"`
	Via string `json:"via,omitempty"`
}

// IP assignment pool object format
type IpAssignmentPool struct {
	// Starting IP address in range
	IpRangeStart string `json:"ipRangeStart"`
	// Ending IP address in range (inclusive)
	IpRangeEnd string `json:"ipRangeEnd"`
}

// Rule object format
type Rule struct {
	// Common fields for each rule
	// Entry type (all caps, case sensitive)
	Type string `json:"type"`
	// If true, MATCHes match if they don't match
	Not bool `json:"not"`
	// If true, match is ORed with previous match result state
	Or bool `json:"or"`
}

// Peer status, /status
type PeerStatus struct {
	// 10-digit (40 bit) ZeroTier address of this node
	Address string `json:"address"`
	// Current system clock at time of this request
	Clock int `json:"clock"`
	// Cluster status if clustering is enabled (usually null)
	Cluster interface{} `json:"cluster"`
	// Contents of local.conf configuration file (see section 4.2)
	Config interface{} `json:"config"`
	// True if node can communicate with at least one root
	Online bool `json:"online"`
	// World ID of current planet (always 149604618 except in testing scenarios)
	PlanetWorldId int `json:"planetWorldId"`
	// Timestamp of current planet
	PlanetWorldTimestamp int `json:"planetWorldTimestamp"`
	// Public identity of this node (address and public key)
	PublicIdentity string `json:"publicIdentity"`
	// If true, node is tunneling through a ZeroTier TCP relay (slow)
	TcpFallbackActive bool `json:"tcpFallbackActive"`
	// ZeroTier One version
	Version string `json:"version"`
	// Build version
	VersionBuild int `json:"versionBuild"`
	// Major version number
	VersionMajor int `json:"versionMajor"`
	// Minor version number
	VersionMinor int `json:"versionMinor"`
	// Revision portion of version number
	VersionRev int `json:"versionRev"`
}

type Config struct {
	// Use ZeroTier Central on https://my.zerotier.com. This also implies a
	// different API
	Central bool

	// Address of Network Controller - ignored when Central is true
	Address string

	// Either API token for ZeroTier Central, or authentication token for
	// network controller
	Token string
}

// Pre-defined values to be used for pointers
func True() *bool {
	val := true
	return &val
}
func False() *bool {
	val := false
	return &val
}
func Zero() *int {
	return Int(0)
}
func Int(val int) *int {
	return &val
}