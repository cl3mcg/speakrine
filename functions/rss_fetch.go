package functions

import (
	"database/sql"
	"errors"
	"log/slog"
	"strings"

	"github.com/cl3mcg/speakrine/databases"
	"github.com/mmcdole/gofeed"
)

// fetchRSSFeed fetches the RSS feed from the provided URL.
// It returns a parsed *gofeed.Feed and an error if any occurs during the fetching or parsing process.
//
// Parameters:
//   - url: The URL of the RSS feed to fetch.
//
// Returns:
//   - *gofeed.Feed: The parsed RSS feed.
//   - error: An error, if any, occurred while fetching or parsing the feed.
func fetchRSSFeed(url string) (*gofeed.Feed, error) {
	fp := gofeed.NewParser()
	feed, err := fp.ParseURL(url)
	if err != nil {
		return nil, err
	}
	return feed, nil
}

// isItemInDatabase checks if an RSS item with the given GUID already exists in the database.
// It returns true if the item exists, false if it doesn't, and an error if there was an issue querying the database.
//
// Parameters:
//   - db: The database connection instance.
//   - guid: The GUID of the RSS item to check.
//
// Returns:
//   - bool: A boolean indicating if the item exists in the database (true if it exists, false if it doesn't).
//   - error: An error, if any, occurred during the database query.
func isItemInDatabase(db *sql.DB, guid string) (bool, error) {
	query := "SELECT 1 FROM rss_items WHERE guid = ?"
	var exists bool
	err := db.QueryRow(query, guid).Scan(&exists)
	if errors.Is(err, sql.ErrNoRows) {
		return false, nil
	} else if err != nil {
		return false, err
	}
	return true, nil
}

// insertRSSItem inserts a new RSS item into the database.
//
// Parameters:
//   - db: The database connection instance.
//   - feedId: The ID of the associated feed.
//   - item: The *gofeed.Item representing the RSS item to insert.
//
// Returns:
//   - error: An error, if any, occurred during the insertion process.
func insertRSSItem(db *sql.DB, feedId int, item *gofeed.Item) error {
	query := `
		INSERT INTO rss_items (rss_feed_id, title, link, author, summary_raw, content_raw, guid, published_date, updated_date, extraction_date)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, NOW())
	`

	// Handle nil values and concatenate authors' names
	title := item.Title
	link := item.Link
	author := ""
	if len(item.Authors) > 0 {
		authorNames := make([]string, len(item.Authors))
		for i, author := range item.Authors {
			authorNames[i] = author.Name
		}
		author = strings.Join(authorNames, ", ")
	}
	description := item.Description
	content := ""
	if item.Content == "" {
		content = item.Description
	} else {
		content = item.Content
	}
	guid := item.GUID
	publishedDate := item.PublishedParsed
	updatedDate := item.UpdatedParsed

	_, err := db.Exec(
		query,
		feedId,
		title,
		link,
		author,
		description,
		content,
		guid,
		publishedDate,
		updatedDate,
	)
	return err
}

// updateFeedLastUpdate updates the last_update timestamp for the given feed in the database.
//
// Parameters:
//   - db: The database connection instance.
//   - feedId: The ID of the feed to update.
//
// Returns:
//   - error: An error, if any, occurred while updating the feed's last update timestamp.
func updateFeedLastUpdate(db *sql.DB, feedId int) error {
	query := "UPDATE rss_feeds SET last_update = NOW() WHERE id = ?"
	_, err := db.Exec(query, feedId)
	return err
}

// FetchAllRSSData retrieves and processes RSS feeds for all the database.
// It queries the database for feeds that need updating, fetches their items, and inserts them into the database if they don't already exist.
// It also updates the feed's last update timestamp after processing.
//
// Returns:
//   - error: An error if any occurs during the fetching, processing, or database operations.
func FetchAllRSSData() error {
	// Get the database connection instance.
	db := databases.GetDB()

	// Define the SQL query to find feeds that need updating.
	query := `
		SELECT id, url
		FROM rss_feeds
		WHERE last_update IS NULL OR last_update < NOW() - INTERVAL 15 MINUTE
	`

	// Execute the query to get the feeds.
	rows, err := db.Query(query)
	if err != nil {
		slog.Error("Error querying rss_feeds", "Error", err)
		return err
	}
	defer func(rows *sql.Rows) {
		err := rows.Close()
		if err != nil {
			slog.Error("Error closing rows", "Error", err)
		}
	}(rows)

	// Check if there are any rows returned
	if !rows.Next() {
		// No rows returned
		slog.Info("No RSS feeds to fetch: No rows returned from the query")
		return nil
	}

	// Iterate over the feeds.
	for rows.Next() {
		var feedId int
		var feedUrl string
		if err := rows.Scan(&feedId, &feedUrl); err != nil {
			slog.Error("Error scanning rss_feeds row", "Error", err)
			return err
		}

		// Fetch the RSS feed.
		feed, err := fetchRSSFeed(feedUrl)
		if err != nil {
			slog.Error("Error fetching RSS feed", "Feed URL", feedUrl, "Error", err)
			continue
		}

		// Process the RSS items.
		for _, item := range feed.Items {
			// Check if the item is already in the database.
			if exists, err := isItemInDatabase(db, item.GUID); err != nil {
				slog.Error("Error checking if item exists in database", "GUID", item.GUID, "Error", err)
				continue
			} else if exists {
				continue
			}

			// Insert the new item into the database.
			if err := insertRSSItem(db, feedId, item); err != nil {
				slog.Error("Error inserting RSS item", "GUID", item.GUID, "Error", err)
				continue
			}
		}

		// Update the last_update timestamp for the feed.
		if err := updateFeedLastUpdate(db, feedId); err != nil {
			slog.Error("Error updating feed last_update", "Feed ID", feedId, "Error", err)
			continue
		}
	}

	return nil
}
