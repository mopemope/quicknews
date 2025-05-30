package tts

import (
	"bytes"
	"context"
	"os/exec"

	"github.com/cockroachdb/errors"
	"github.com/mopemope/quicknews/config"
	"google.golang.org/genai"
)

var ModelName = "gemini-2.5-flash-preview-tts"
var VoiceName = "Aoede"
var FFmpegBin = "/usr/bin/ffmpeg"

type GeminiTTS struct {
	config *config.Config
}

func NewGeminiTTS(config *config.Config) TTSEngine {
	return &GeminiTTS{
		config: config,
	}
}

func runFFmpeg(data []byte) ([]byte, error) {
	cmd := exec.Command(FFmpegBin, "-f", "s16le", "-ar", "24k", "-ac", "1", "-i", "-", "-f", "mp3", "-")
	cmd.Stdin = (bytes.NewReader(data))
	b, err := cmd.Output()
	if err != nil {
		return nil, errors.Wrap(err, "failed to run ffmpeg")
	}
	return b, nil
}

func (g *GeminiTTS) SynthesizeText(ctx context.Context, text string) ([]byte, error) {
	apiKey := g.config.GeminiApiKey
	if apiKey == "" {
		return nil, errors.New("GEMINI_API_KEY environment variable not set")
	}

	client, err := genai.NewClient(ctx, &genai.ClientConfig{
		APIKey:  apiKey,
		Backend: genai.BackendGeminiAPI,
	})
	if err != nil {
		return nil, errors.Wrap(err, "failed to create gemini client")
	}

	res, err := client.Models.GenerateContent(ctx,
		ModelName,
		genai.Text(text),
		&genai.GenerateContentConfig{
			ResponseModalities: []string{string(genai.MediaModalityAudio)},
			SpeechConfig: &genai.SpeechConfig{
				VoiceConfig: &genai.VoiceConfig{
					PrebuiltVoiceConfig: &genai.PrebuiltVoiceConfig{
						VoiceName: VoiceName,
					},
				},
			},
		})
	if err != nil {
		return nil, errors.Wrap(err, "failed to generate content")
	}

	for _, can := range res.Candidates {
		for _, part := range can.Content.Parts {
			// Check if the part is of type audio
			data := part.InlineData.Data
			if data != nil {
				return runFFmpeg(data)
			}
		}
	}
	return nil, err
}

func (g *GeminiTTS) PlayAudioData(audioData []byte) error {
	if len(audioData) == 0 {
		return errors.New("audio data cannot be empty")
	}

	if err := PlayMP3Audio(audioData); err != nil {
		return err
	}
	return nil
}
