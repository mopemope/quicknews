package tts

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"log/slog"
	"net/http"
	"strconv"

	"github.com/cockroachdb/errors"
)

type VoiceVox struct {
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

type config struct {
	endpoint   string
	speaker    int
	style      int
	speed      float64
	intonation float64
	volume     float64
	pitch      float64
}

func getSpeakers(cfg config) (voiceVoxSpeakers, error) {
	resp, err := http.Get(cfg.endpoint + "/speakers")
	if err != nil {
		return nil, errors.Wrap(err, "failed to get speakers")
	}
	defer resp.Body.Close()
	var speakers voiceVoxSpeakers
	if err := json.NewDecoder(resp.Body).Decode(&speakers); err != nil {
		return nil, errors.Wrap(err, "failed to decode speakers")
	}
	return speakers, nil
}

func getQuery(cfg config, id int, text string) (*voiceVoxParams, error) {
	req, err := http.NewRequest("POST", cfg.endpoint+"/audio_query", nil)
	if err != nil {
		return nil, err
	}
	q := req.URL.Query()
	q.Add("speaker", strconv.Itoa(id))
	q.Add("text", text)
	req.URL.RawQuery = q.Encode()
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, errors.Wrap(err, " failed to get audio query")
	}
	defer resp.Body.Close()
	var params *voiceVoxParams
	if err := json.NewDecoder(resp.Body).Decode(&params); err != nil {
		return nil, errors.Wrap(err, "failed to decode params")
	}
	return params, nil
}

func synth(cfg config, id int, params *voiceVoxParams) ([]byte, error) {
	b, err := json.MarshalIndent(params, "", "  ")
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest("POST", cfg.endpoint+"/synthesis", bytes.NewReader(b))
	if err != nil {
		return nil, errors.Wrap(err, "failed to create request")
	}
	req.Header.Add("Accept", "audio/wav")
	req.Header.Add("Content-Type", "application/json")
	q := req.URL.Query()
	q.Add("speaker", strconv.Itoa(id))
	req.URL.RawQuery = q.Encode()

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, errors.Wrap(err, "failed to synthesize")
	}
	defer resp.Body.Close()
	buff := bytes.NewBuffer(nil)
	if _, err := io.Copy(buff, resp.Body); err != nil {
		return nil, errors.Wrap(err, "failed to copy response body")
	}

	slog.Info("synth", slog.Any("size", buff.Len()))

	return buff.Bytes(), nil
}

func NewVoiceVox(speaker int, style int) *VoiceVox {
	return &VoiceVox{
		Speaker: speaker,
		Style:   style,
	}
}

func (v *VoiceVox) SynthesizeText(ctx context.Context, text string) ([]byte, error) {

	cfg := config{
		endpoint:   "http://localhost:50021",
		speaker:    v.Speaker,
		style:      v.Style,
		speed:      SpeachOpt.SpeakingRate,
		intonation: 1.0,
		volume:     1.0,
		pitch:      0,
	}

	speakers, err := getSpeakers(cfg)
	if err != nil {
		return nil, err
	}
	if cfg.speaker >= len(speakers) {
		return nil, errors.New("speaker not found")
	}
	spk := speakers[cfg.speaker]
	if cfg.style >= len(spk.Styles) {
		return nil, errors.New("style not found")
	}

	spkID := spk.Styles[cfg.style].ID
	slog.Info("VoiceVox", slog.Any("name", spk.Name), slog.Any("styles", spk.Styles[cfg.style].Name), slog.Any("speaker", spkID))

	params, err := getQuery(cfg, spkID, text)
	if err != nil {
		return nil, errors.Wrap(err, "failed to get query")
	}
	params.SpeedScale = cfg.speed
	params.PitchScale = cfg.pitch
	params.IntonationScale = cfg.intonation
	params.VolumeScale = cfg.volume

	return synth(cfg, spkID, params)
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
