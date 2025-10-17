package gemini

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type parseTestCase struct {
	name    string
	input   string
	want    *PageSummary
	wantErr bool
}

func TestParseResponse(t *testing.T) {
	tests := []parseTestCase{
		{
			name:  "well formed content",
			input: "Example Title\n-----\nLine one\n\nLine two\n",
			want: &PageSummary{
				Title:   "Example Title",
				Summary: "Line one\nLine two",
			},
		},
		{
			name:    "missing delimiter",
			input:   "No delimiter present",
			wantErr: true,
		},
		{
			name:  "cleans known noise",
			input: "**記事のタイトル**#了解しました。 Awesome Title *\n-----\nParagraph one.\n\nParagraph two.\n",
			want: &PageSummary{
				Title:   "Awesome Title",
				Summary: "Paragraph one.\nParagraph two.",
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got, err := parseResponse(tc.input)

			if tc.wantErr {
				require.Error(t, err)
				assert.Nil(t, got)
				return
			}

			require.NoError(t, err)
			require.NotNil(t, got)

			if tc.want.Title != "" {
				assert.Equal(t, tc.want.Title, got.Title)
			}
			if tc.want.Summary != "" {
				assert.Equal(t, tc.want.Summary, got.Summary)
			}
		})
	}
}
