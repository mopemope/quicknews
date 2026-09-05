package scraper

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGetPageContent_StatusErrors(t *testing.T) {
	tests := []struct {
		name       string
		status     int
		wantStatus int
	}{
		{name: "404", status: http.StatusNotFound, wantStatus: http.StatusNotFound},
		{name: "403", status: http.StatusForbidden, wantStatus: http.StatusForbidden},
		{name: "500", status: http.StatusInternalServerError, wantStatus: http.StatusInternalServerError},
		{name: "503", status: http.StatusServiceUnavailable, wantStatus: http.StatusServiceUnavailable},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(tt.status)
			}))
			t.Cleanup(server.Close)

			_, err := GetPageContent(context.Background(), server.URL)
			require.Error(t, err)

			var statusErr *HTTPStatusError
			require.ErrorAs(t, err, &statusErr)
			assert.Equal(t, tt.wantStatus, statusErr.StatusCode)
		})
	}
}

func TestGetPageContent_UnsupportedContentType(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/pdf")
		_, _ = w.Write([]byte("%PDF-1.4 fake"))
	}))
	t.Cleanup(server.Close)

	_, err := GetPageContent(context.Background(), server.URL)
	require.Error(t, err)

	var permErr *PermanentError
	require.ErrorAs(t, err, &permErr)
	assert.True(t, permErr.NonRetryable())
}

func TestGetPageContent_Success(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		_, _ = w.Write([]byte(`<html><head><title>OK</title></head><body><p>本文です。十分な長さのテキストを返します。</p></body></html>`))
	}))
	t.Cleanup(server.Close)

	page, err := GetPageContent(context.Background(), server.URL)
	require.NoError(t, err)
	assert.Equal(t, "OK", page.Title)
	assert.NotEmpty(t, page.Content)
}

func TestGetPageContent_InvalidURL(t *testing.T) {
	_, err := GetPageContent(context.Background(), "not-a-url")
	require.Error(t, err)

	var permErr *PermanentError
	require.ErrorAs(t, err, &permErr)
}

func TestErrorTypes(t *testing.T) {
	statusErr := &HTTPStatusError{StatusCode: 404, URL: "https://example.com", Err: assert.AnError}
	assert.Contains(t, statusErr.Error(), "404")
	assert.ErrorIs(t, statusErr, assert.AnError)

	permErr := &PermanentError{Err: assert.AnError}
	assert.ErrorIs(t, permErr, assert.AnError)
	assert.True(t, permErr.NonRetryable())
}
