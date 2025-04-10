package tts

import (
	"bytes"
	"context"
	"io"
	"sync"
	"time"

	"github.com/cockroachdb/errors"
	"github.com/gopxl/beep"
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
	switch SpeachOpt.Engine {
	case "google":
		return NewGoogleTTS(config)
	case "voicevox":
		return NewVoiceVox(config)
	default:
		return NewGoogleTTS(config)
	}
}

func PlayMP3Audio(audioData []byte) error {
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
	defer func() {
		_ = streamer.Close()
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
		return errors.Wrap(err, "failed to decode mp3 data")
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
