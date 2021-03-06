package config

type configStructure struct {
	StoreType           string `json:"store_type"`
	PrivateAPI          bool   `json:"private_api"`
	Port                int    `json:"port"`
	AuthDBLocation      string `json:"auth_db_location"`
	AuthEnabled         bool   `json:"auth_enabled"`
	RegistrationEnabled bool   `json:"registration_enabled"`
	APIRoot             string `json:"api_root"`
	RedirectRoot        string `json:"redirect_root"`
	JSONStoreLocation   string `json:"json_store_location,omitempty"`
	SQLiteStoreLocation string `json:"sqlite_store_location,omitempty"`
}

// Config is the global config for the URL shortener, with the default values as follows
var Config = configStructure{
	StoreType:           "json",
	PrivateAPI:          false,
	Port:                3000,
	AuthDBLocation:      "/var/lib/linkener/auth.db",
	AuthEnabled:         true,
	RegistrationEnabled: true,
	APIRoot:             "api",
	RedirectRoot:        "",
	JSONStoreLocation:   "/var/lib/linkener/urls.json",
	SQLiteStoreLocation: "/var/lib/linkener/urls.db",
}
