package config

import (
	"github.com/BurntSushi/toml"
	"github.com/caarlos0/env/v11"
	"github.com/pkg/errors"
)

type Config struct {
	GoogleApplicationCredentials string `toml:"google_application_credentials" env:"GOOGLE_APPLICATION_CREDENTIALS"`
	GeminiApiKey                 string `toml:"gemini_api_key" env:"GEMINI_API_KEY"`
	ExportOrg                    string `toml:"export_org" env:"EXPORT_ORG"`
	EnableEnvOverride            bool   `toml:"enable_env_override" env:"ENABLE_ENV_OVERRIDE"`
	VoiceVox                     VoiceVox
}

type VoiceVox struct {
	Speaker int `toml:"speaker" env:"VOICEVOX_SPEAKER"`
	Style   int `toml:"style" env:"VOICEVOX_STYLE"`
}

func LoadConfig(path string) (*Config, error) {
	var config Config
	if _, err := toml.DecodeFile(path, &config); err != nil {
		return nil, errors.Wrap(err, "failed to decode config file")
	}

	if config.EnableEnvOverride {
		// overwrite config with environment variables
		if err := env.Parse(&config); err != nil {
			return nil, errors.Wrap(err, "failed to parse game config")
		}
	}
	return &config, nil
}
