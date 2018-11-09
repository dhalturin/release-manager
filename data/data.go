package data

// Version string
var Version string

// SlackOAuth string
var SlackOAuth string

// SlackVerification string
var SlackVerification string

const (
	// Project const
	Project = "release-manager"
	// ContentType const
	ContentType = "application/vnd.api+json"
	// Command const
	Command = "/rm"
	// RepoTTL min
	RepoTTL = 15
)

type argumentsType struct {
	Version    bool
	ConfigPath string
	ConfigFile string
}

type configType struct {
	Listen string `json:"listen"`
	Port   int    `json:"port"`
	DbHost string `json:"db_host"`
	DbPort int    `json:"db_port"`
	DbName string `json:"db_name"`
	DbUser string `json:"db_user"`
	DbPass string `json:"db_pass"`
}

// Config variable
var Config configType

// Arg variable
var Arg argumentsType
