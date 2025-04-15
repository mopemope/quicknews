package rss

import (
	"encoding/xml"
	"os"
	"time"

	"github.com/cockroachdb/errors"
	"github.com/mopemope/quicknews/config"
)

type RSS struct {
	XMLName            xml.Name `xml:"rss"`
	Version            string   `xml:"version,attr"`
	XMLNamespaceItunes string   `xml:"xmlns:itunes,attr"`
	Channel            Channel  `xml:"channel"`
}

type Channel struct {
	Title          string         `xml:"title"`
	Link           string         `xml:"link"`
	Description    string         `xml:"description"`     // Podcast description
	Language       string         `xml:"language"`        // Language code (e.g., ja, en)
	Copyright      string         `xml:"copyright"`       // Copyright information
	PubDate        string         `xml:"pubDate"`         // Feed last updated date (RFC1123Z format)
	ItunesAuthor   string         `xml:"itunes:author"`   // Author
	ItunesSubtitle string         `xml:"itunes:subtitle"` // Short description
	ItunesSummary  string         `xml:"itunes:summary"`  // Detailed description (can be the same as description)
	ItunesOwner    ItunesOwner    `xml:"itunes:owner"`    // Owner information
	ItunesImage    ItunesImage    `xml:"itunes:image"`    // Artwork image
	ItunesCategory ItunesCategory `xml:"itunes:category"` // Category
	ItunesExplicit string         `xml:"itunes:explicit"` // Explicit content (yes/no)
	Items          []Item         `xml:"item"`            // List of episodes
}

type ItunesOwner struct {
	ItunesName  string `xml:"itunes:name"`  // Owner name
	ItunesEmail string `xml:"itunes:email"` // Owner email address
}

// iTunes artwork image
type ItunesImage struct {
	Href string `xml:"href,attr"` // Image URL
}

// iTunes category
// Nested categories are possible, but here only one is specified for simplicity.
type ItunesCategory struct {
	Text string `xml:"text,attr"` // Category name
}

// Item (episode) information
type Item struct {
	Title          string      `xml:"title"`           // Episode title
	Link           string      `xml:"link"`            // Episode webpage URL (optional)
	Guid           string      `xml:"guid"`            // Unique identifier (usually audio file URL, etc.)
	PubDate        string      `xml:"pubDate"`         // Episode publication date (RFC1123Z format)
	Description    string      `xml:"description"`     // Episode description (CDATA allows HTML tags)
	Enclosure      Enclosure   `xml:"enclosure"`       // Audio file information
	ItunesAuthor   string      `xml:"itunes:author"`   // Episode author (can be the same as the channel)
	ItunesSubtitle string      `xml:"itunes:subtitle"` // Episode short description
	ItunesSummary  string      `xml:"itunes:summary"`  // Episode detailed description (can be the same as description)
	ItunesDuration string      `xml:"itunes:duration"` // Duration (seconds or HH:MM:SS)
	ItunesImage    ItunesImage `xml:"itunes:image"`    // Episode-specific artwork (optional)
	ItunesExplicit string      `xml:"itunes:explicit"` // Explicit content (yes/no)
}

// Enclosure (media file information)
type Enclosure struct {
	URL    string `xml:"url,attr"`
	Length string `xml:"length,attr"`
	Type   string `xml:"type,attr"`
}

func NewRSS(config *config.Podcast) *RSS {
	pubDate := time.Now().Format(time.RFC1123Z)
	// create
	return &RSS{
		Version:            "2.0",
		XMLNamespaceItunes: "http://www.itunes.com/dtds/podcast-1.0.dtd",
		Channel: Channel{
			Title:          config.ChannelTitle,
			Link:           config.ChannelLink,
			Description:    config.ChannelDesc,
			Language:       "ja",
			PubDate:        pubDate,
			ItunesAuthor:   "",
			ItunesSubtitle: "",
			ItunesSummary:  "",
			ItunesOwner: ItunesOwner{
				ItunesName:  "",
				ItunesEmail: "",
			},
			ItunesImage: ItunesImage{
				Href: "",
			},
			ItunesCategory: ItunesCategory{
				Text: "Technology",
			},
			ItunesExplicit: "no",
			Items:          []Item{},
		},
	}
}

type RSSItem struct {
	Title       string
	Link        string
	Guid        string
	PubDate     string
	Description string
	AudioURL    string
	Length      string
	MimeType    string
}

func (r *RSS) AddItem(rssIem RSSItem) {
	item := Item{
		Title:       rssIem.Title,
		Link:        rssIem.Link,
		Guid:        rssIem.Guid,
		PubDate:     rssIem.PubDate,
		Description: rssIem.Description,
		Enclosure: Enclosure{
			URL:    rssIem.AudioURL,
			Length: rssIem.Length,
			Type:   rssIem.MimeType,
		},
		ItunesAuthor:   r.Channel.ItunesAuthor,
		ItunesSubtitle: r.Channel.ItunesSubtitle,
		ItunesSummary:  r.Channel.ItunesSummary,
		ItunesDuration: "00:00", // default value
		ItunesImage:    r.Channel.ItunesImage,
		ItunesExplicit: r.Channel.ItunesExplicit,
	}
	r.Channel.Items = append(r.Channel.Items, item)
}

func (r *RSS) WriteToFile(filePath string) error {
	xmlOutput := []byte(xml.Header)

	xmlBytes, err := xml.MarshalIndent(r, "", "  ")
	if err != nil {
		return errors.Wrap(err, "failed to marshal RSS to XML")
	}

	// Combine with XML data
	xmlOutput = append(xmlOutput, xmlBytes...)
	err = os.WriteFile(filePath, xmlOutput, 0644)
	if err != nil {
		return errors.Wrap(err, "failed to write RSS to file")
	}
	return nil
}
