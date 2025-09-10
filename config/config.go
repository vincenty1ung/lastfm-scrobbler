package config

import (
	"github.com/spf13/viper"
)

var ConfigObj = &Config{}

type Config struct {
	Lastfm     ScrobblerConfig  `yaml:"lastfm"`
	Musixmatch MusixmatchConfig `yaml:"musixmatch"`
	Log        LogConfig        `yaml:"log"`
	Database   DatabaseConfig   `yaml:"database"`
	HTTP       HTTPConfig       `yaml:"http"`
	Telemetry  TelemetryConfig  `yaml:"telemetry"`
	Scrobblers []string         `yaml:"scrobblers"`
}

type ScrobblerConfig struct {
	ApplicationName string `yaml:"applicationName"`
	ApiKey          string `yaml:"apiKey"`
	SharedSecret    string `yaml:"sharedSecret"`
	RegisteredTo    string `yaml:"registeredTo"`
	UserLoginToken  string `yaml:"userLoginToken"`
	UserUsername    string `yaml:"userUsername"`
	UserPassword    string `yaml:"userPassword"`
}

type LogConfig struct {
	Path  string `yaml:"path"`
	Level string `yaml:"level"`
}

type MusixmatchConfig struct {
	ApiKey string `yaml:"apiKey"`
}

type DatabaseConfig struct {
	Path string `yaml:"path"`
}

type HTTPConfig struct {
	Port string `yaml:"port"`
}

type TelemetryConfig struct {
	Name     string  `yaml:"name,optional"`
	Endpoint string  `yaml:",optional"`
	Sampler  float64 `yaml:",default=1.0"`
	Batcher  string  `yaml:",default=jaeger,options=jaeger|zipkin|otlpgrpc|otlphttp|file"`
	// OtlpHeaders represents the headers for OTLP gRPC or HTTP transport.
	// For example:
	//  uptrace-dsn: 'http://project2_secret_token@localhost:14317/2'
	OtlpHeaders map[string]string `yaml:",optional"`
	// OtlpHttpPath represents the path for OTLP HTTP transport.
	// For example
	// /v1/traces
	OtlpHttpPath string `yaml:",optional"`
	// OtlpHttpSecure represents the scheme to use for OTLP HTTP transport.
	OtlpHttpSecure bool `yaml:",optional"`
	// Disabled indicates whether StartAgent starts the agent.
	Disabled bool `yaml:",optional"`
}

func InitConfig(filePath string) {
	viper.SetConfigFile(filePath)
	viper.SetConfigType("yaml")
	if err := viper.ReadInConfig(); err != nil {
		panic(err)
	}
	if err := viper.Unmarshal(ConfigObj); err != nil {
		panic(err)
	}
}
