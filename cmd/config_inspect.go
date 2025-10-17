package cmd

import (
	"encoding/json"
	"fmt"
	"os"
	"sort"
	"strings"
	"text/tabwriter"

	"github.com/cockroachdb/errors"
	"github.com/mopemope/quicknews/config"
)

// ConfigCmd outputs the currently loaded configuration in a human friendly format.
type ConfigCmd struct {
	Format      string `help:"Output format. Supported values: table, json." enum:"table,json" default:"table"`
	ShowSecrets bool   `help:"Display sensitive values such as API keys."`
}

// Run executes the config command.
func (c *ConfigCmd) Run(cfg *config.Config) error {
	if cfg == nil {
		return errors.New("config is not loaded")
	}

	entries := sanitizeConfigEntries(cfg, c.ShowSecrets)

	switch c.Format {
	case "json":
		return printConfigJSON(entries)
	case "table":
		printConfigTable(entries)
		return nil
	default:
		return errors.Newf("unsupported format: %s", c.Format)
	}
}

type configEntry struct {
	key   string
	value any
}

func sanitizeConfigEntries(cfg *config.Config, showSecrets bool) []configEntry {
	entries := make([]configEntry, 0, 24)
	add := func(key string, value any) {
		entries = append(entries, configEntry{key: key, value: value})
	}

	add("config.path", cfg.SourcePath)
	add("db", cfg.DB)
	add("export_org", cfg.ExportOrg)
	add("audio", derefString(cfg.AudioPath))
	add("use_gemini_tts", cfg.UseGeminiTTS)
	add("enable_env_override", cfg.EnableEnvOverride)
	add("speaking_rate", cfg.SpeakingRate)
	add("require_confirm", cfg.RequireConfirm)
	add("save_audio_data", cfg.SaveAudioData)
	add("gemini_api_key", maskIfNeeded("gemini_api_key", cfg.GeminiApiKey, showSecrets))
	add("google_application_credentials", maskIfNeeded("google_application_credentials", cfg.GoogleApplicationCredentials, showSecrets))

	if cfg.VoiceVox != nil {
		add("voicevox.speaker", cfg.VoiceVox.Speaker)
		add("voicevox.style", cfg.VoiceVox.Style)
	} else {
		add("voicevox", nil)
	}

	if cfg.Prompt != nil {
		add("prompt.summary", derefString(cfg.Prompt.Summary))
	} else {
		add("prompt", nil)
	}

	if cfg.Cloudflare != nil {
		add("cloudflare.access_key_id", maskIfNeeded("cloudflare_access_key_id", cfg.Cloudflare.AccessKeyID, showSecrets))
		add("cloudflare.secret_access_key", maskIfNeeded("cloudflare_secret_access_key", cfg.Cloudflare.SecretAccessKey, showSecrets))
		add("cloudflare.bucket_name", cfg.Cloudflare.BucketName)
		add("cloudflare.endpoint_url", cfg.Cloudflare.EndpointURL)
	} else {
		add("cloudflare", nil)
	}

	if cfg.Podcast != nil {
		add("podcast.channel_title", cfg.Podcast.ChannelTitle)
		add("podcast.channel_link", cfg.Podcast.ChannelLink)
		add("podcast.channel_desc", cfg.Podcast.ChannelDesc)
		add("podcast.author", cfg.Podcast.Author)
		add("podcast.publish_url", cfg.Podcast.PublishURL)
	} else {
		add("podcast", nil)
	}

	sort.Slice(entries, func(i, j int) bool {
		return entries[i].key < entries[j].key
	})

	return entries
}

func printConfigTable(entries []configEntry) {
	tw := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	fmt.Fprintln(tw, "KEY\tVALUE")
	for _, entry := range entries {
		fmt.Fprintf(tw, "%s\t%s\n", entry.key, formatValue(entry.value))
	}
	_ = tw.Flush()
}

func printConfigJSON(entries []configEntry) error {
	values := make(map[string]any, len(entries))
	for _, entry := range entries {
		values[entry.key] = entry.value
	}
	encoded, err := json.MarshalIndent(values, "", "  ")
	if err != nil {
		return errors.Wrap(err, "failed to marshal config")
	}
	fmt.Println(string(encoded))
	return nil
}

func derefString(value *string) any {
	if value == nil {
		return nil
	}
	return *value
}

func maskIfNeeded(field string, value string, showSecrets bool) any {
	if showSecrets || value == "" {
		return value
	}
	lower := strings.ToLower(field)
	if !(strings.Contains(lower, "key") || strings.Contains(lower, "secret") || strings.Contains(lower, "credential")) {
		return value
	}
	if len(value) <= 4 {
		return "****"
	}
	return fmt.Sprintf("%s…%s", value[:3], value[len(value)-2:])
}

func formatValue(value any) string {
	switch v := value.(type) {
	case nil:
		return "<nil>"
	case string:
		if v == "" {
			return ""
		}
		return v
	default:
		return fmt.Sprintf("%v", v)
	}
}
