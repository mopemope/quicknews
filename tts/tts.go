package tts

import (
	"bytes"
	"context"
	"io"
	"log/slog"
	"sync"
	"time"

	"github.com/cockroachdb/errors"
	"github.com/gopxl/beep/v2"
	"github.com/gopxl/beep/v2/mp3"
	"github.com/gopxl/beep/v2/speaker"
	"github.com/gopxl/beep/v2/wav"
	"github.com/mopemope/quicknews/config"
)

var (
	ErrNoCredentials   = errors.New("google application credentials not configured") // Updated error message
	ErrEmptyAudioData  = errors.New("audio data cannot be empty")
	mutex              = &sync.Mutex{}
	speakerInitialized = false

	SpeachOpt = &SpeechOptions{
		Engine:       "google",
		Speaker:      0,
		Pitch:        1.3,
		SpeakingRate: 1.3,
	}
)

type TTSEngine interface {
	SynthesizeText(ctx context.Context, text string) ([]byte, error)
	PlayAudioData(audioData []byte) error
}

type SpeechOptions struct {
	Engine       string
	Speaker      int
	Pitch        float64
	SpeakingRate float64
}

func (s *SpeechOptions) UpPitch() {
	s.Pitch += 0.1
}

func (s *SpeechOptions) DownPitch() {
	s.Pitch -= 0.1
}

func (s *SpeechOptions) UpSpeakingRate() {
	s.SpeakingRate += 0.1
}

func (s *SpeechOptions) DownSpeakingRate() {
	s.SpeakingRate -= 0.1
}

func NewTTSEngine(config *config.Config) TTSEngine {
	if config == nil {
		slog.Error("Config is nil, using default Google TTS")
		return NewGoogleTTS(nil)
	}

	if config.UseGeminiTTS {
		// override the engine to Gemini if UseGeminiTTS is true
		SpeachOpt.Engine = "gemini"
		slog.Debug("Using Gemini TTS engine")
	}

	var engine TTSEngine
	switch SpeachOpt.Engine {
	case "gemini":
		engine = NewGeminiTTS(config)
	case "google":
		engine = NewGoogleTTS(config)
	case "voicevox":
		engine = NewVoiceVox(config)
	default:
		engine = NewGoogleTTS(config)
	}

	if engine == nil {
		slog.Error("Failed to create TTS engine, falling back to Google TTS")
		return NewGoogleTTS(config)
	}

	return engine
}

func PlayMP3Audio(audioData []byte) error {
	if len(audioData) == 0 {
		return ErrEmptyAudioData
	}

	mutex.Lock()
	defer mutex.Unlock()

	reader := bytes.NewReader(audioData)
	if reader == nil {
		return errors.New("failed to create bytes reader")
	}

	// Wrap the reader with io.NopCloser to satisfy the io.ReadCloser interface
	streamer, format, err := mp3.Decode(io.NopCloser(reader))
	if err != nil {
		return errors.Wrap(err, "failed to decode mp3 data")
	}
	if streamer == nil {
		return errors.New("mp3 decoder returned nil streamer")
	}
	defer func() {
		if closeErr := streamer.Close(); closeErr != nil {
			slog.Warn("Failed to close mp3 streamer", "error", closeErr)
		}
	}()

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

func PlayWavAudio(audioData []byte) error {
	mutex.Lock()
	defer mutex.Unlock()

	if len(audioData) == 0 {
		return ErrEmptyAudioData
	}

	reader := bytes.NewReader(audioData)
	streamer, format, err := wav.Decode(io.NopCloser(reader))
	if err != nil {
		return errors.Wrap(err, "failed to decode wave data")
	}
	defer func() {
		_ = streamer.Close()
	}()

	if !speakerInitialized {
		err = speaker.Init(format.SampleRate, format.SampleRate.N(time.Second/10))
		if err != nil {
			return errors.Wrap(err, "failed to initialize speaker")
		}
		speakerInitialized = true
	}

	done := make(chan struct{})

	speaker.Play(beep.Seq(streamer, beep.Callback(func() {
		close(done)
	})))

	<-done

	speaker.Clear()
	return nil
}
