package tts

import (
	"context"
	"io"
	"net/http"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

type roundTripFunc func(*http.Request) (*http.Response, error)

func (f roundTripFunc) RoundTrip(req *http.Request) (*http.Response, error) {
	return f(req)
}

func TestVoiceVoxRequestsUseContext(t *testing.T) {
	type contextKey string
	ctx := context.WithValue(context.Background(), contextKey("request-id"), "voicevox-test")

	cfg := voicevoxConfig{
		endpoint: "http://voicevox.local",
		client: &http.Client{
			Transport: roundTripFunc(func(req *http.Request) (*http.Response, error) {
				require.Equal(t, "voicevox-test", req.Context().Value(contextKey("request-id")))
				return voiceVoxResponse(http.StatusOK, `[]`), nil
			}),
		},
	}

	_, err := getSpeakers(ctx, cfg)
	require.NoError(t, err)
}

func TestVoiceVoxChecksHTTPStatus(t *testing.T) {
	cfg := voicevoxConfig{
		endpoint: "http://voicevox.local",
		client: &http.Client{
			Transport: roundTripFunc(func(*http.Request) (*http.Response, error) {
				return voiceVoxResponse(http.StatusInternalServerError, `server error`), nil
			}),
		},
	}

	_, err := getSpeakers(context.Background(), cfg)
	require.Error(t, err)
	require.ErrorContains(t, err, "status 500")
}

func TestVoiceVoxGetQueryChecksHTTPStatus(t *testing.T) {
	cfg := voicevoxConfig{
		endpoint: "http://voicevox.local",
		client: &http.Client{
			Transport: roundTripFunc(func(*http.Request) (*http.Response, error) {
				return voiceVoxResponse(http.StatusBadRequest, `bad request`), nil
			}),
		},
	}

	_, err := getQuery(context.Background(), cfg, 1, "text")
	require.Error(t, err)
	require.ErrorContains(t, err, "status 400")
}

func TestVoiceVoxSynthChecksHTTPStatus(t *testing.T) {
	cfg := voicevoxConfig{
		endpoint: "http://voicevox.local",
		client: &http.Client{
			Transport: roundTripFunc(func(*http.Request) (*http.Response, error) {
				return voiceVoxResponse(http.StatusBadGateway, `bad gateway`), nil
			}),
		},
	}

	_, err := synth(context.Background(), cfg, 1, &voiceVoxParams{})
	require.Error(t, err)
	require.ErrorContains(t, err, "status 502")
}

func voiceVoxResponse(status int, body string) *http.Response {
	return &http.Response{
		StatusCode: status,
		Body:       io.NopCloser(strings.NewReader(body)),
		Header:     make(http.Header),
	}
}
