package tts

import (
	"context"
	"os"
	"testing"

	"github.com/mopemope/quicknews/config"
	"github.com/mopemope/quicknews/log"
	"github.com/stretchr/testify/require"
)

func TestSynthesizeText(t *testing.T) {
	err := log.InitializeLogger("", true)
	require.NoError(t, err, "Failed to initialize logger")
	text := `
Gemini API は、ネイティブのテキスト読み上げ（TTS）生成機能を使用して、テキスト入力を 1 人のスピーカーまたは複数のスピーカーの音声に変換できます。
テキスト読み上げ（TTS）の生成は制御可能です。つまり、自然言語を使用してインタラクションを構造化し、音声のスタイル、アクセント、ペース、トーンをガイドできます。
TTS 機能は、インタラクティブで非構造化の音声、マルチモーダルの入力と出力用に設計された Live API によって提供される音声生成とは異なります。
Live API は動的会話コンテキストに優れていますが、Gemini API による TTS は、ポッドキャストやオーディオブックの生成など、スタイルと音声をきめ細かく制御して正確なテキストを朗読する必要があるシナリオに適しています。
このガイドでは、テキストから単一スピーカーと複数スピーカーの音声を生成する方法について説明します。
`
	apiKey := os.Getenv("GEMINI_API_KEY")
	if apiKey == "" {
		t.Skip("Skipping test: GEMINI_API_KEY environment variable not set")
	}

	cfg := &config.Config{
		GeminiApiKey: apiKey,
	}
	ctx := context.Background()
	engine := NewGeminiTTS(cfg)

	audioData, err := engine.SynthesizeText(ctx, text)
	require.NoError(t, err, "Failed to synthesize text")

	err = engine.PlayAudioData(audioData)
	require.NoError(t, err, "Failed to play audio data")
}
