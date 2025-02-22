-- Drop the table if it already exists to avoid conflicts
-- DROP TABLE IF EXISTS rss_feeds;

-- Create the rss_feeds table
CREATE TABLE rss_feeds (
    id SMALLINT UNSIGNED AUTO_INCREMENT PRIMARY KEY, -- Unique identifier for each RSS feed
    user_id INT UNSIGNED NOT NULL, -- User ID associated with the RSS feed
    common_name VARCHAR(255) NOT NULL, -- Common name for the RSS feed
    type VARCHAR(255) NOT NULL CHECK (LOWER(type) IN ('rss', 'atom', 'json', 'scrap', 'other')), -- Type of the feed with a check constraint
    url VARCHAR(767) NOT NULL, -- URL of the RSS feed
    categories JSON DEFAULT NULL, -- Categories/tags associated with the entry
    last_update TIMESTAMP DEFAULT NULL, -- Last update time for the feed
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE, -- Foreign key linking to the users table with ON DELETE CASCADE
    UNIQUE (user_id, url)
);

-- Drop the table if it already exists to avoid conflicts
-- DROP TABLE IF EXISTS rss_entries;

-- Create the rss_entries table
CREATE TABLE rss_items (
    id INT UNSIGNED AUTO_INCREMENT PRIMARY KEY, -- Unique identifier for each RSS entry
    rss_feed_id SMALLINT UNSIGNED NOT NULL, -- Foreign key linking to the RSS feed source
    title VARCHAR(255) NOT NULL, -- The title of the RSS/ATOM entry
    link TEXT NOT NULL, -- The link/URL of the entry
    author VARCHAR(767) DEFAULT NULL, -- The author of the entry (if available)
    summary_raw TEXT DEFAULT NULL, -- A brief summary or description of the entry
    summary_formatted TEXT DEFAULT NULL, -- A brief summary or description of the entry
    content_raw LONGTEXT DEFAULT NULL, -- Full content of the entry
    content_formatted LONGTEXT DEFAULT NULL, -- Full content of the entry
    categories JSON DEFAULT NULL, -- Categories/tags associated with the entry
    guid VARCHAR(767) UNIQUE DEFAULT NULL, -- A unique identifier for the entry (GUID in RSS/ATOM)
    published_date DATETIME DEFAULT CURRENT_TIMESTAMP, -- The publication date of the entry
    updated_date DATETIME DEFAULT NULL, -- The updated date of the entry (for ATOM feeds)
    extraction_date TIMESTAMP DEFAULT CURRENT_TIMESTAMP, -- When the entry was extracted
    is_read BOOL DEFAULT FALSE NOT NULL, -- Check if the article is read or not
    is_hidden BOOL DEFAULT FALSE NOT NULL, -- Check if the article is read or not 
    FOREIGN KEY (rss_feed_id) REFERENCES rss_feeds(id) ON DELETE CASCADE -- Link to rss_feeds table with ON DELETE CASCADE
);
