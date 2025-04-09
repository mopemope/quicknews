package scraper

import (
	"fmt"
	"net/url"

	"github.com/gocolly/colly/v2"
)

// GetTitle fetches the HTML title from the given URL.
func GetTitle(targetURL string) (string, error) {
	// Validate the URL
	_, err := url.ParseRequestURI(targetURL)
	if err != nil {
		return "", fmt.Errorf("invalid URL %s: %w", targetURL, err)
	}

	c := colly.NewCollector(
		// Allow visiting only the domain of the target URL
		colly.AllowedDomains(), // TODO: Consider if specific domains should be allowed instead
	)

	var pageTitle string
	var visitError error

	// Find and extract the title element
	c.OnHTML("title", func(e *colly.HTMLElement) {
		pageTitle = e.Text
	})

	// Handle request errors
	c.OnError(func(r *colly.Response, err error) {
		visitError = fmt.Errorf("request to %s failed: status %d, error: %w", r.Request.URL, r.StatusCode, err)
	})

	// Start scraping
	err = c.Visit(targetURL)
	if err != nil {
		// Handle errors that occur before or during the visit initiation
		// (e.g., network issues, DNS resolution errors)
		// Prioritize the OnError callback's error if it exists
		if visitError != nil {
			return "", visitError
		}
		return "", fmt.Errorf("failed to visit %s: %w", targetURL, err)
	}

	// If OnError was triggered, return that error
	if visitError != nil {
		return "", visitError
	}

	// If the title wasn't found but no other error occurred
	if pageTitle == "" && visitError == nil {
		return "", fmt.Errorf("could not find title tag on %s", targetURL)
	}

	return pageTitle, nil
}
