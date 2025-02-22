package types

import (
	"fmt"
	"sort"
	"time"
)

// RssItem represents a single RSS feed item with its details.
type RssItem struct {
	Id               int       // Unique identifier for the RSS item.
	RssFeedId        int       // Identifier for the RSS feed to which this item belongs.
	Title            string    // Title of the RSS item.
	Link             string    // URL link to the full content of the RSS item.
	Author           string    // Author of the RSS item.
	SummaryRaw       string    // Short raw summary of the RSS item.
	SummaryFormatted string    // Short formatted summary of the RSS item.
	ContentRaw       string    // Full raw content of the RSS item.
	ContentFormatted string    // Full formatted content of the RSS item.
	PublishedDate    time.Time // Date and time when the RSS item was published.
	UpdatedDate      time.Time // Date and time when the RSS item was last updated.
	ExtractionDate   time.Time // Date and time when the RSS item was extracted.
	Categories       []string  // List of categories associated with the RSS item.
	IsRead           bool      // Flag indicating whether the RSS item has been read.
	IsHidden         bool      // Flag indicating whether the RSS item is marked as 'hidden'.
	PrevItemId       int       // Identifier for the previous RSS item in the feed.
	NextItemId       int       // Identifier for the next RSS item in the feed.
}

// RssFeed represents an RSS feed, containing metadata and its list of entries.
type RssFeed struct {
	Id              int       // Unique identifier for the RSS feed.
	CommonName      string    // Common name of the RSS feed.
	Description     string    // Description of the RSS feed.
	FeedType        string    // Type of feed (e.g., RSS, Atom).
	Url             string    // URL of the RSS feed.
	LastPublication time.Time // Last time an article was published in the RSS feed.
	LastUpdate      time.Time // Last time the RSS feed was updated.
	Categories      []string  // List of categories associated with the RSS feed.
	EntriesUnread   int       // Number of unread entries in the RSS feed.
	Entries         []RssItem // List of RSS feed items (entries).
}

// OrderEntries sorts the Entries of an RssFeed in reverse chronological order.
//
// The Entries slice is reordered in-place, with the most recent RssItem (by PublishedDate)
// appearing at index 0 and the oldest at the last index. If two entries have the same
// PublishedDate, they are ordered by ExtractionDate. If both PublishedDate and ExtractionDate
// are the same, they are ordered alphabetically by Title.
func (f *RssFeed) OrderEntries() {
	// Sort the Entries slice based on the specified criteria.
	sort.Slice(f.Entries, func(i, j int) bool {
		// Compare the PublishedDate of two entries.
		if !f.Entries[i].PublishedDate.Equal(f.Entries[j].PublishedDate) {
			return f.Entries[i].PublishedDate.After(f.Entries[j].PublishedDate)
		}
		// If PublishedDate is the same, compare the ExtractionDate.
		if !f.Entries[i].ExtractionDate.Equal(f.Entries[j].ExtractionDate) {
			return f.Entries[i].ExtractionDate.After(f.Entries[j].ExtractionDate)
		}
		// If both PublishedDate and ExtractionDate are the same, compare the Title alphabetically.
		return f.Entries[i].Title < f.Entries[j].Title
	})
}

// WhenWasLastUpdate returns a human-readable string indicating when the last update was.
func (f *RssFeed) WhenWasLastUpdate() (lastUpdate string) {
	now := time.Now()
	duration := now.Sub(f.LastUpdate)

	switch {
	case duration > 365*24*time.Hour:
		return "More than a year ago"
	case duration > 6*30*24*time.Hour:
		return "More than 6 months ago"
	case duration > 5*30*24*time.Hour:
		return "More than 5 months ago"
	case duration > 4*30*24*time.Hour:
		return "More than 4 months ago"
	case duration > 3*30*24*time.Hour:
		return "More than 3 months ago"
	case duration > 2*30*24*time.Hour:
		return "More than 2 months ago"
	case duration > 1*30*24*time.Hour:
		return "More than a month ago"
	case duration > 3*7*24*time.Hour:
		return "More than 3 weeks ago"
	case duration > 2*7*24*time.Hour:
		return "More than 2 weeks ago"
	case duration > 1*7*24*time.Hour:
		return "More than a week ago"
	case duration > 48*time.Hour:
		return fmt.Sprintf("%d days ago", int(duration.Hours()/24))
	case duration > 24*time.Hour:
		return "Yesterday"
	case duration > 2*time.Hour:
		return fmt.Sprintf("%d hours ago", int(duration.Hours()))
	case duration > time.Hour:
		return "Less than an hour ago"
	default:
		return "Just now"
	}
}

// WhenWasLastPublication returns a human-readable string indicating when the last article publication occurred in a given RSS feed.
func (f *RssFeed) WhenWasLastPublication() (lastPublication string) {
	now := time.Now()
	duration := now.Sub(f.LastPublication)

	switch {
	case duration > 365*24*time.Hour:
		return "More than a year ago"
	case duration > 6*30*24*time.Hour:
		return "More than 6 months ago"
	case duration > 5*30*24*time.Hour:
		return "More than 5 months ago"
	case duration > 4*30*24*time.Hour:
		return "More than 4 months ago"
	case duration > 3*30*24*time.Hour:
		return "More than 3 months ago"
	case duration > 2*30*24*time.Hour:
		return "More than 2 months ago"
	case duration > 1*30*24*time.Hour:
		return "More than a month ago"
	case duration > 3*7*24*time.Hour:
		return "More than 3 weeks ago"
	case duration > 2*7*24*time.Hour:
		return "More than 2 weeks ago"
	case duration > 1*7*24*time.Hour:
		return "More than a week ago"
	case duration > 48*time.Hour:
		return fmt.Sprintf("%d days ago", int(duration.Hours()/24))
	case duration > 24*time.Hour:
		return "Yesterday"
	case duration > 2*time.Hour:
		return fmt.Sprintf("%d hours ago", int(duration.Hours()))
	case duration > time.Hour:
		return "Less than an hour ago"
	default:
		return "Just now"
	}
}

// NumberOfUnreadItems returns the count of Entries that has a IsRead value of false
func (f *RssFeed) NumberOfUnreadItems() (count int) {

	// If the EntriesUnread field is set, return it directly.
	if f.EntriesUnread > 0 {
		return f.EntriesUnread
	}

	// Otherwise, count the number of unread entries from the Entries field.
	for _, entry := range f.Entries {
		if !entry.IsRead && !entry.IsHidden {
			count++
		}
	}

	return count
}
