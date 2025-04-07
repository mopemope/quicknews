package tts

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"log/slog"
	"os"
	"sync"
	"time"

	texttospeech "cloud.google.com/go/texttospeech/apiv1"
	"cloud.google.com/go/texttospeech/apiv1/texttospeechpb"
	"github.com/cockroachdb/errors"
	"github.com/gopxl/beep/v2" // Add beep package import
	"github.com/gopxl/beep/v2/mp3"
	"github.com/gopxl/beep/v2/speaker"
	"google.golang.org/api/option"
)

var (
	ErrNoCredentials   = errors.New("GOOGLE_APPLICATION_CREDENTIALS environment variable is not set")
	ErrEmptyAudioData  = errors.New("audio data cannot be empty")
	mutex              = &sync.Mutex{}
	speakerInitialized = false
)

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
			SpeakingRate:     1.5,
			Pitch:            1.3,
		},
	}

	resp, err := c.client.SynthesizeSpeech(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("failed to synthesize speech: %w", err)
	}

	return resp.AudioContent, nil
}

func SynthesizeText(ctx context.Context, text string) ([]byte, error) {
	if os.Getenv("GOOGLE_APPLICATION_CREDENTIALS") == "" {
		return nil, ErrNoCredentials
	}
	client, err := NewClient(ctx)
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

// PlayAudioData plays the given MP3 audio data using beep.
func PlayAudioData(audioData []byte) error {
	mutex.Lock()
	defer mutex.Unlock()

	if len(audioData) == 0 {
		return ErrEmptyAudioData
	}

	reader := bytes.NewReader(audioData)
	// Wrap the reader with io.NopCloser to satisfy the io.ReadCloser interface
	streamer, format, err := mp3.Decode(io.NopCloser(reader))
	if err != nil {
		return errors.Wrap(err, "failed to decode mp3 data")
	}
	defer streamer.Close()

	if !speakerInitialized {
		// Initialize the speaker with the format retrieved from the decoder.
		// Use a buffer size that provides reasonable latency.
		err = speaker.Init(format.SampleRate, format.SampleRate.N(time.Second/10))
		if err != nil {
			return errors.Wrap(err, "failed to initialize speaker")
		}
		speakerInitialized = true
	}

	done := make(chan struct{})
	// Use beep.Seq to play the streamer and then call the callback
	speaker.Play(beep.Seq(streamer, beep.Callback(func() {
		close(done)
	})))

	<-done // Wait until playback is finished.

	// It's good practice to clear the speaker buffer after playing.
	speaker.Clear()
	return nil
}
