package tts

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"strconv"
	"time"

	"github.com/cockroachdb/errors"
	"github.com/mopemope/quicknews/config"
)

type VoiceVox struct {
	Config  *config.Config
	Speaker int
	Style   int
}

type voiceVoxParams struct {
	AccentPhrases      []voiceVoxAccentPhrases `json:"accent_phrases"`
	SpeedScale         float64                 `json:"speedScale"`
	PitchScale         float64                 `json:"pitchScale"`
	IntonationScale    float64                 `json:"intonationScale"`
	VolumeScale        float64                 `json:"volumeScale"`
	PrePhonemeLength   float64                 `json:"prePhonemeLength"`
	PostPhonemeLength  float64                 `json:"postPhonemeLength"`
	OutputSamplingRate int                     `json:"outputSamplingRate"`
	OutputStereo       bool                    `json:"outputStereo"`
	Kana               string                  `json:"kana"`
}

type voiceVoxMora struct {
	Text            string   `json:"text"`
	Consonant       *string  `json:"consonant"`
	ConsonantLength *float64 `json:"consonant_length"`
	Vowel           string   `json:"vowel"`
	VowelLength     float64  `json:"vowel_length"`
	Pitch           float64  `json:"pitch"`
}

type voiceVoxAccentPhrases struct {
	Moras           []voiceVoxMora `json:"moras"`
	Accent          int            `json:"accent"`
	PauseMora       *voiceVoxMora  `json:"pause_mora"`
	IsInterrogative bool           `json:"is_interrogative"`
}

type voiceVoxSpeakers []struct {
	Name        string           `json:"name"`
	SpeakerUUID string           `json:"speaker_uuid"`
	Styles      []voiceVoxStyles `json:"styles"`
	Version     string           `json:"version"`
}

type voiceVoxStyles struct {
	ID   int    `json:"id"`
	Name string `json:"name"`
}

type voicevoxConfig struct {
	endpoint   string
	client     *http.Client
	speaker    int
	style      int
	speed      float64
	intonation float64
	volume     float64
	pitch      float64
}

func (cfg voicevoxConfig) httpClient() *http.Client {
	if cfg.client != nil {
		return cfg.client
	}
	return &http.Client{Timeout: 30 * time.Second}
}

func getSpeakers(ctx context.Context, cfg voicevoxConfig) (voiceVoxSpeakers, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, cfg.endpoint+"/speakers", nil)
	if err != nil {
		return nil, errors.Wrap(err, "failed to create speakers request")
	}
	resp, err := cfg.httpClient().Do(req)
	if err != nil {
		return nil, errors.Wrap(err, "failed to get speakers")
	}
	defer func() {
		if err := resp.Body.Close(); err != nil {
			slog.Warn("failed to close response body", "error", err)
		}
	}()
	if err := ensureVoiceVoxStatus(resp, "speakers"); err != nil {
		return nil, err
	}
	var speakers voiceVoxSpeakers
	if err := json.NewDecoder(resp.Body).Decode(&speakers); err != nil {
		return nil, errors.Wrap(err, "failed to decode speakers")
	}
	return speakers, nil
}

func getQuery(ctx context.Context, cfg voicevoxConfig, id int, text string) (*voiceVoxParams, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, cfg.endpoint+"/audio_query", nil)
	if err != nil {
		return nil, err
	}
	q := req.URL.Query()
	q.Add("speaker", strconv.Itoa(id))
	q.Add("text", text)
	req.URL.RawQuery = q.Encode()
	resp, err := cfg.httpClient().Do(req)
	if err != nil {
		return nil, errors.Wrap(err, " failed to get audio query")
	}
	defer func() {
		if err := resp.Body.Close(); err != nil {
			slog.Warn("failed to close response body", "error", err)
		}
	}()
	if err := ensureVoiceVoxStatus(resp, "audio_query"); err != nil {
		return nil, err
	}
	var params *voiceVoxParams
	if err := json.NewDecoder(resp.Body).Decode(&params); err != nil {
		return nil, errors.Wrap(err, "failed to decode params")
	}
	return params, nil
}

func synth(ctx context.Context, cfg voicevoxConfig, id int, params *voiceVoxParams) ([]byte, error) {
	b, err := json.MarshalIndent(params, "", "  ")
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, cfg.endpoint+"/synthesis", bytes.NewReader(b))
	if err != nil {
		return nil, errors.Wrap(err, "failed to create request")
	}
	req.Header.Add("Accept", "audio/wav")
	req.Header.Add("Content-Type", "application/json")
	q := req.URL.Query()
	q.Add("speaker", strconv.Itoa(id))
	req.URL.RawQuery = q.Encode()

	resp, err := cfg.httpClient().Do(req)
	if err != nil {
		return nil, errors.Wrap(err, "failed to synthesize")
	}
	defer func() {
		if err := resp.Body.Close(); err != nil {
			slog.Warn("failed to close response body", "error", err)
		}
	}()
	if err := ensureVoiceVoxStatus(resp, "synthesis"); err != nil {
		return nil, err
	}
	buff := bytes.NewBuffer(nil)
	if _, err := io.Copy(buff, resp.Body); err != nil {
		return nil, errors.Wrap(err, "failed to copy response body")
	}

	slog.Info("synth", slog.Any("size", buff.Len()))

	return buff.Bytes(), nil
}

func ensureVoiceVoxStatus(resp *http.Response, operation string) error {
	if resp.StatusCode >= 200 && resp.StatusCode < 300 {
		return nil
	}
	return errors.Newf("voicevox %s failed with status %d", operation, resp.StatusCode)
}

func NewVoiceVox(config *config.Config) *VoiceVox {
	speaker := SpeachOpt.Speaker
	style := 0
	if config.VoiceVox != nil {
		speaker = config.VoiceVox.Speaker
		style = config.VoiceVox.Style
	}
	return &VoiceVox{
		Config:  config,
		Speaker: speaker,
		Style:   style,
	}
}

func (v *VoiceVox) SynthesizeText(ctx context.Context, text string) ([]byte, error) {

	cfg := voicevoxConfig{
		endpoint:   config.DefaultVoiceVoxEndpoint,
		speaker:    v.Speaker,
		style:      v.Style,
		speed:      SpeachOpt.SpeakingRate,
		intonation: 1.0,
		volume:     1.0,
		pitch:      0,
	}
	if v.Config != nil && v.Config.VoiceVox != nil && v.Config.VoiceVox.Endpoint != "" {
		cfg.endpoint = v.Config.VoiceVox.Endpoint
	}

	speakers, err := getSpeakers(ctx, cfg)
	if err != nil {
		return nil, err
	}
	if cfg.speaker < 0 {
		return nil, fmt.Errorf("speaker not found: %d", cfg.speaker)
	}
	if cfg.speaker >= len(speakers) {
		return nil, fmt.Errorf("speaker not found: %d", cfg.speaker)
	}
	spk := speakers[cfg.speaker]
	if cfg.style < 0 {
		return nil, fmt.Errorf("style not found: %d", cfg.style)
	}
	if cfg.style >= len(spk.Styles) {
		return nil, fmt.Errorf("style not found: %d", cfg.style)
	}

	spkID := spk.Styles[cfg.style].ID
	slog.Info("VoiceVox", slog.Any("name", spk.Name), slog.Any("styles", spk.Styles[cfg.style].Name), slog.Any("speaker", spkID))

	params, err := getQuery(ctx, cfg, spkID, text)
	if err != nil {
		return nil, errors.Wrap(err, "failed to get query")
	}
	params.SpeedScale = cfg.speed
	params.PitchScale = cfg.pitch
	params.IntonationScale = cfg.intonation
	params.VolumeScale = cfg.volume

	return synth(ctx, cfg, spkID, params)
}

func (v *VoiceVox) PlayAudioData(audioData []byte) error {
	if len(audioData) == 0 {
		return errors.New("audio data cannot be empty")
	}

	if err := PlayWavAudio(audioData); err != nil {
		return err
	}
	return nil
}
