package tts

import (
	"context"
	"fmt"
	"log/slog"

	texttospeech "cloud.google.com/go/texttospeech/apiv1"
	"cloud.google.com/go/texttospeech/apiv1/texttospeechpb"
	"github.com/cockroachdb/errors"
	"github.com/mopemope/quicknews/config"

	"google.golang.org/api/option"
)

type GoogleTTS struct {
	config *config.Config
}

func NewGoogleTTS(config *config.Config) TTSEngine {
	return &GoogleTTS{
		config: config,
	}
}

func (g *GoogleTTS) SynthesizeText(ctx context.Context, text string) ([]byte, error) {
	cred := g.config.GoogleApplicationCredentials
	if cred == "" {
		return nil, ErrNoCredentials
	}

	opts := make([]option.ClientOption, 0)
	opts = append(opts, option.WithCredentialsFile(cred))

	client, err := NewClient(ctx, opts...)
	if err != nil {
		return nil, errors.Wrap(err, "failed to create Google TTS client")
	}

	defer func() {
		if err := client.Close(); err != nil {
			slog.Error("failed closing connection to google tts", "error", err)
		}
	}()

	audioContent, err := client.SynthesizeSpeech(ctx, text)
	if err != nil {
		return nil, errors.Wrap(err, "failed to synthesize speech")
	}

	return audioContent, nil
}
func (g *GoogleTTS) PlayAudioData(audioData []byte) error {
	if len(audioData) == 0 {
		return ErrEmptyAudioData
	}
	if err := PlayMP3Audio(audioData); err != nil {
		return err
	}
	return nil
}

// Client wraps the Google Cloud Text-to-Speech client.
type Client struct {
	client *texttospeech.Client
}

// NewClient creates a new Google Cloud Text-to-Speech client.
// It expects the GOOGLE_APPLICATION_CREDENTIALS environment variable to be set.
func NewClient(ctx context.Context, opts ...option.ClientOption) (*Client, error) {
	c, err := texttospeech.NewClient(ctx, opts...)
	if err != nil {
		return nil, errors.Wrap(err, "failed to create texttospeech client")
	}
	return &Client{client: c}, nil
}

// Close closes the underlying client connection.
func (c *Client) Close() error {
	return c.client.Close()
}

// SynthesizeSpeech synthesizes speech from the given text.
func (c *Client) SynthesizeSpeech(ctx context.Context, text string) ([]byte, error) {
	if text == "" {
		return nil, errors.New("input text cannot be empty")
	}

	req := &texttospeechpb.SynthesizeSpeechRequest{
		Input: &texttospeechpb.SynthesisInput{
			InputSource: &texttospeechpb.SynthesisInput_Text{Text: text},
		},
		// Build the voice request, select the language code ("en-US") and the SSML
		// voice gender ("neutral").
		Voice: &texttospeechpb.VoiceSelectionParams{
			LanguageCode: "ja-JP", // Japanese language code
			SsmlGender:   texttospeechpb.SsmlVoiceGender_FEMALE,
		},
		// Select the type of audio file you want returned.
		AudioConfig: &texttospeechpb.AudioConfig{
			AudioEncoding:    texttospeechpb.AudioEncoding_MP3,
			EffectsProfileId: []string{"small-bluetooth-speaker-class-device"},
			SpeakingRate:     SpeachOpt.SpeakingRate,
			Pitch:            SpeachOpt.Pitch,
		},
	}

	resp, err := c.client.SynthesizeSpeech(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("failed to synthesize speech: %w", err)
	}

	return resp.AudioContent, nil
}
