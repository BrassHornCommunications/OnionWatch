package main

// CoreConf describes the configuration that dictates how the server works
type CoreConf struct {
	DbPath      string `json:"dbpath"`
	Debug       bool   `json:"debug"`
	TLS         bool   `json:"tls_enabled"`
	TLSKey      string `json:"tls_private_key"`
	TLSCert     string `json:"tls_certificate"`
	FQDN        string `json:fqdn"`
	ListenIP    string `json:"listen_ip"`
	ListenIPv6  string `json:"listen_ipv6"`
	ListenPort  int64  `json:"listen_port"`
	SOCKSConfig string `json:"socks_config"`
	BTCAddr     string `json:"btc_rpc_address"`
	BTCPort     int    `json:"btc_rpc_port"`
	BTCUser     string `json:"btc_rpc_username"`
	BTCPass     string `json:"btc_rpc_password"`
}

// Relay describe a Tor relay (Exit or Middle relay)
type Relay struct {
	Fingerprint string   `json:"fingerprint"`
	NickName    string   `json:"nickname"`
	OrAddresses []string `json:"or_addresses"`
	DirAddress  string   `json:"dir_address"`
	LastSeen    string   `json:"last_seen"`
	LastChanged string   `json:"last_changed_address_or_port"`
	FirstSeen   string   `json:"first_seen"`
	Running     bool     `json:"running"`
	Flags       []string `json:"flags"`
	ASNumber    string   `json:"as_number"`
	ASName      string   `json:"as_number"`
	Contact     string   `json:"contact"`
}

// RelayWatch describes the complete configuration for watching a relay
type RelayWatch struct {
	Fingerprint        string `json:"fingerprint"`
	ContactEmail       string `json:"email"`
	GPGKey             string `json:"gpg"`
	EmailVerified      bool   `json:"email_verified"`
	VerificationString string `json:"verification_string"`
	LastChecked        string `json:"last_checked"`
	LastSeen           string `json:"last_seen"`
	RelayDetails       Relay  `json:"relay_details"`
}

// RelayStateTemplate helps inform the user what has changed about the relay they are monitoring
type RelayStateTemplate struct {
	anythingChanged       bool
	HasRunningChanged     bool `json:"running"`
	HasNickNameChanged    bool `json:"nickname"`
	HasOrAddressesChanged bool `json:"or_addresses"`
	HasDirAddressChanged  bool `json:"dir_address"`
	HasFlagsChanged       bool `json:"flags"`
	HasASNChanged         bool `json:"asn"`
}

// HiddenService describes a Tor hidden service
type HiddenService struct {
	HSAddr      string `json:"url"`
	IsReachable bool   `json:"is_reachable"`
}

// HSWatch describes the complete configuration required for watching a hidden service
type HSWatch struct {
	URL           string        `json:"url"`
	ContactEmail  string        `json:"email"`
	GPGKey        string        `json:"gpg"`
	EmailVerified bool          `json:"email_verified"`
	LastChecked   string        `json:"last_checked"`
	LastSeen      string        `json:"last_seen"`
	HSDetails     HiddenService `json:"hs_details"`
}

//OnionooResponse describes a response from the Onionoo API
type OnionooResponse struct {
	Version         string  `json:"versions"`
	RelaysPublished string  `json:"relays_published"`
	Relays          []Relay `json:"relays"`
}
