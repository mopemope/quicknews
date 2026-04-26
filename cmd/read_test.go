package cmd

import (
	"testing"

	"github.com/mopemope/quicknews/config"
	"github.com/mopemope/quicknews/tts"
	"github.com/stretchr/testify/require"
)

func TestValidateReadMode_NonInteractiveWithoutTTY(t *testing.T) {
	err := validateReadMode(true, false)
	require.NoError(t, err)
}

func TestValidateReadMode_TTYRequired(t *testing.T) {
	err := validateReadMode(false, false)
	require.Error(t, err)
}

func TestValidateReadMode_InteractiveTTY(t *testing.T) {
	err := validateReadMode(false, true)
	require.NoError(t, err)
}

func TestReadCmd_ApplyTTSConfig_UsesVoicevoxFlag(t *testing.T) {
	original := *tts.SpeachOpt
	defer func() {
		*tts.SpeachOpt = original
	}()

	cmd := &ReadCmd{
		Voicevox: true,
		Speaker:  7,
	}
	cfg := &config.Config{SpeakingRate: 1.4}

	cmd.applyTTSConfig(cfg)

	require.Equal(t, "voicevox", tts.SpeachOpt.Engine)
	require.Equal(t, 7, tts.SpeachOpt.Speaker)
	require.NotNil(t, cfg.VoiceVox)
	require.Equal(t, 7, cfg.VoiceVox.Speaker)
	require.Equal(t, 1.4, tts.SpeachOpt.SpeakingRate)
}
