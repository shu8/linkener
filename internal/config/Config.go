package config

type configStructure struct {
	StoreType      string `json:"store_type"`
	StoreLocation  string `json:"store_location,omitempty"`
	PrivateAPI     bool   `json:"private_api"`
	Port           int    `json:"port"`
	AuthDBLocation string `json:"auth_db_location"`
}

// Config is the global config for the URL shortener, with the default values as follows
var Config = configStructure{
	StoreType:      "json",
	StoreLocation:  "/var/lib/url-shortener/urls.json",
	PrivateAPI:     false,
	Port:           3000,
	AuthDBLocation: "/var/lib/url-shortener/auth.db",
}
