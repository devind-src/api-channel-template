package channel

import (
	"errors"

	"github.com/devind-src/syncnet-sdk/appconfig"
)

// Config is this channel's own section of appsettings.json. AppName/AppDir
// in the same file are read by the SDK.
type Config struct {
	Channel struct {
		ClientID  string `json:"ClientId"`
		SecretKey string `json:"SecretKey"`
	} `json:"Channel"`
}

// LoadConfig reads and validates the channel settings.
func LoadConfig(path string) (*Config, error) {
	cfg, err := appconfig.Load[Config](path)
	if err != nil {
		return nil, err
	}
	if cfg.Channel.ClientID == "" || cfg.Channel.SecretKey == "" {
		return nil, errors.New("channel: Channel.ClientId and Channel.SecretKey are required")
	}
	return cfg, nil
}
