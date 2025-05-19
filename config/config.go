package config

import (
	"os"
	"path/filepath"

	"github.com/BurntSushi/toml"
	"github.com/caarlos0/env/v11"
	"github.com/cockroachdb/errors"
)

type Config struct {
	DB                           string  `toml:"db" env:"DB"`
	GoogleApplicationCredentials string  `toml:"google_application_credentials" env:"GOOGLE_APPLICATION_CREDENTIALS"`
	GeminiApiKey                 string  `toml:"gemini_api_key" env:"GEMINI_API_KEY"`
	ExportOrg                    string  `toml:"export_org" env:"EXPORT_ORG"`
	AudioPath                    *string `toml:"audio" env:"AUDIO"`
	EnableEnvOverride            bool    `toml:"enable_env_override" env:"ENABLE_ENV_OVERRIDE"`
	SpeakingRate                 float64 `toml:"speaking_rate" env:"SPEAKING_RATE"`
	RequireConfirm               bool    `toml:"require_confirm" env:"REQUIRE_CONFIRM"`
	SaveAudioData                bool    `toml:"save_audio_data" env:"SAVE_AUDIO_DATA"`
	VoiceVox                     *VoiceVox
	Prompt                       *Prompt
	Cloudflare                   *Cloudflare
	Podcast                      *Podcast
}

type Podcast struct {
	ChannelTitle string `toml:"channel_title" env:"PODCAST_CHANNEL_TITLE"`
	ChannelLink  string `toml:"channel_link" env:"PODCAST_CHANNEL_LINK"`
	ChannelDesc  string `toml:"channel_desc" env:"PODCAST_CHANNEL_DESC"`
	Author       string `toml:"author" env:"PODCAST_AUTHOR"`
	PublishURL   string `toml:"publish_url" env:"PODCAST_PUBLISH_URL"`
}

type Cloudflare struct {
	AccessKeyID     string `toml:"access_key_id" env:"CLOUDFLARE_ACCESS_KEY_ID"`
	SecretAccessKey string `toml:"secret_access_key" env:"CLOUDFLARE_SECRET_ACCESS_KEY"`
	BucketName      string `toml:"bucket_name" env:"CLOUDFLARE_BUCKET_NAME"`
	EndpointURL     string `toml:"endpoint_url" env:"CLOUDFLARE_ENDPOINT_URL"` // e.g. https://<ACCOUNT_ID>.r2.cloudflarestorage.com
}

type VoiceVox struct {
	Speaker int `toml:"speaker" env:"VOICEVOX_SPEAKER"`
	Style   int `toml:"style" env:"VOICEVOX_STYLE"`
}

type Prompt struct {
	Summary *string `toml:"summary" env:"PROMPT_SUMMARY"`
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
	if config.SpeakingRate == 0 {
		config.SpeakingRate = 1.3
	}
	if config.DB == "" {
		home, err := os.UserHomeDir()
		if err != nil {
			return nil, errors.Wrap(err, "failed to get user home directory")
		}
		config.DB = filepath.Join(home, "quicknews.db")
	}
	return &config, nil
}
