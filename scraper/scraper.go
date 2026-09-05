package scraper

import (
	"context"
	"fmt"
	"mime"
	"net/url"
	"strings"
	"time"

	"github.com/PuerkitoBio/goquery"
	"github.com/gocolly/colly/v2"
	"github.com/saintfish/chardet"
	"golang.org/x/net/html"
	"golang.org/x/text/encoding/htmlindex"
)

const (
	// defaultFetchTimeout bounds a single page fetch.
	defaultFetchTimeout = 30 * time.Second
	// MaxContentRunes is the maximum number of runes kept from a page body.
	MaxContentRunes = 20000
	// DefaultUserAgent is sent with page requests.
	DefaultUserAgent = "Mozilla/5.0 (compatible; quicknews)"
)

// PageContent holds the title and main text extracted from a web page.
type PageContent struct {
	URL     string
	Title   string
	Content string
}

// HTTPStatusError is returned when the target server answers with an error
// status. Callers can inspect StatusCode to decide whether a retry makes
// sense (e.g. 404 vs 503).
type HTTPStatusError struct {
	StatusCode int
	URL        string
	Err        error
}

func (e *HTTPStatusError) Error() string {
	return fmt.Sprintf("request to %s failed: status %d: %v", e.URL, e.StatusCode, e.Err)
}

func (e *HTTPStatusError) Unwrap() error { return e.Err }

// PermanentError marks failures that must not be retried (e.g. unsupported
// content types, malformed pages).
type PermanentError struct {
	Err error
}

func (e *PermanentError) Error() string { return e.Err.Error() }

func (e *PermanentError) Unwrap() error { return e.Err }

// NonRetryable marks the error as non-retryable for summarizer.SummarizeWithRetry.
func (e *PermanentError) NonRetryable() bool { return true }

// GetTitle fetches the HTML title from the given URL.
func GetTitle(ctx context.Context, targetURL string) (string, error) {
	page, err := GetPageContent(ctx, targetURL)
	if err != nil {
		return "", err
	}
	if page.Title == "" {
		return "", &PermanentError{Err: fmt.Errorf("could not find title tag on %s", targetURL)}
	}
	return page.Title, nil
}

// GetPageContent fetches the page and extracts its title and main text.
// The text is truncated to MaxContentRunes runes. Network failures surface
// as *HTTPStatusError when the server responded, *PermanentError for
// structurally unusable responses.
func GetPageContent(ctx context.Context, targetURL string) (*PageContent, error) {
	if _, err := url.ParseRequestURI(targetURL); err != nil {
		return nil, &PermanentError{Err: fmt.Errorf("invalid URL %s: %w", targetURL, err)}
	}

	c := colly.NewCollector(
		colly.AllowedDomains(),
		colly.MaxDepth(1),
		colly.UserAgent(DefaultUserAgent),
	)
	c.SetRequestTimeout(defaultFetchTimeout)
	c.Context = ctx

	var page *PageContent
	var visitError error

	c.OnResponse(func(r *colly.Response) {
		if !isHTMLContentType(r.Headers.Get("Content-Type")) {
			visitError = &PermanentError{Err: fmt.Errorf("unsupported content type %q for %s", r.Headers.Get("Content-Type"), targetURL)}
			return
		}
		text := decodeBytes(r.Body)
		title, body := parsePageHTML(text)
		page = &PageContent{
			URL:     targetURL,
			Title:   title,
			Content: truncateRunes(body, MaxContentRunes),
		}
	})

	c.OnError(func(r *colly.Response, err error) {
		visitError = &HTTPStatusError{
			StatusCode: r.StatusCode,
			URL:        r.Request.URL.String(),
			Err:        err,
		}
	})

	if err := c.Visit(targetURL); err != nil {
		if visitError != nil {
			return nil, visitError
		}
		return nil, fmt.Errorf("failed to visit %s: %w", targetURL, err)
	}
	if visitError != nil {
		return nil, visitError
	}
	if page == nil || (page.Title == "" && page.Content == "") {
		return nil, &PermanentError{Err: fmt.Errorf("could not extract page content from %s", targetURL)}
	}
	return page, nil
}

// isHTMLContentType reports whether the Content-Type header indicates an
// HTML document. An empty header (no Content-Type sent) is allowed so the
// charset detector can still inspect the body.
func isHTMLContentType(contentType string) bool {
	if strings.TrimSpace(contentType) == "" {
		return true
	}
	mediatype, _, err := mime.ParseMediaType(contentType)
	if err != nil {
		return false
	}
	return mediatype == "text/html" || mediatype == "application/xhtml+xml"
}

// decodeBytes converts a response body to UTF-8 when a non-UTF-8 charset is detected.
func decodeBytes(body []byte) string {
	if len(body) == 0 {
		return ""
	}
	res, err := chardet.NewHtmlDetector().DetectBest(body)
	if err != nil || res == nil {
		return string(body)
	}
	if strings.EqualFold(res.Charset, "utf-8") || strings.EqualFold(res.Charset, "ascii") {
		return string(body)
	}
	enc, err := htmlindex.Get(res.Charset)
	if err != nil {
		return string(body)
	}
	decoded, err := enc.NewDecoder().Bytes(body)
	if err != nil {
		return string(body)
	}
	return string(decoded)
}

// parsePageHTML extracts the title tag text and the visible body text.
// Script, style and other non-content elements are removed and whitespace
// runs are collapsed.
func parsePageHTML(data string) (string, string) {
	doc, err := goquery.NewDocumentFromReader(strings.NewReader(data))
	if err != nil {
		return "", ""
	}
	title := strings.TrimSpace(doc.Find("title").First().Text())

	body := doc.Find("body").Clone()
	body.Find("script,style,noscript,template,nav,footer,header,aside,form,iframe,svg,button,select").Remove()
	// Insert separators between block elements so adjacent texts do not merge.
	body.Find(blockSelector).Each(func(_ int, s *goquery.Selection) {
		for _, n := range s.Nodes {
			if n.Parent == nil {
				continue
			}
			n.Parent.InsertBefore(&html.Node{Type: html.TextNode, Data: " "}, n)
		}
	})
	text := strings.Join(strings.Fields(body.Text()), " ")
	return title, text
}

// blockSelector lists elements rendered as visual blocks in browsers.
const blockSelector = "p,div,section,article,li,tr,h1,h2,h3,h4,h5,h6,br,pre,blockquote,ul,ol,table,figure"

// truncateRunes cuts s to at most max runes.
func truncateRunes(s string, max int) string {
	if max <= 0 {
		return ""
	}
	runes := []rune(s)
	if len(runes) <= max {
		return s
	}
	return string(runes[:max])
}
