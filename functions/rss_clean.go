package functions

import (
	"bytes"
	"database/sql"
	"errors"
	"fmt"
	"io"
	"log"
	"log/slog"
	"os"
	"strings"

	"github.com/cl3mcg/speakrine/databases"
	"github.com/gage-technologies/mistral-go"
	gowebly "github.com/gowebly/helpers"
	"golang.org/x/net/html"
)

// cleanHTMLContent parses the HTML, removes undesirable attributes, adjusts <a> tag attributes, and removes or replaces surrounding <html>, <head>, and <body> tags.
// It returns the cleaned HTML content as a string and an error if any occurs during the process.
func cleanHTMLContent(rawContent string) (cleanedContent string, err error) {
	rawContent = processBaseCleaning(rawContent)

	// Parse the HTML string
	doc, err := html.Parse(bytes.NewReader([]byte(rawContent)))
	if err != nil {
		return "", err
	}

	// Define a list of HTML elements to remove if they are empty
	emptyElementsToRemove := []string{"div", "p", "span", "article", "section", "template"}

	// Define a list of HTML elements to remove along with their contents and child elements
	undesirableElementsToRemove := []string{"hr", "figure", "figcaptions", "caption", "video", "audio", "img", "script", "style", "meta", "title", "head", "nav", "aside", "form", "input", "button", "select", "textarea", "label", "option", "optgroup", "progress", "meter", "fieldset", "legend", "details", "summary", "dialog", "menu", "menuitem", "command", "keygen", "source", "track", "map", "area", "embed", "object", "param", "video", "audio", "canvas", "svg", "math", "iframe", "frame", "frameset", "noframes", "noscript", "applet", "basefont", "big", "blink", "center", "font", "marquee", "nobr", "spacer", "strike", "tt", "xmp"}

	// Process various functions to further clean the HTML document
	removeUndesirableElements(doc, undesirableElementsToRemove)
	removeEmptyElements(doc, emptyElementsToRemove)
	removeUndesirableAttributes(doc)
	ensureAnchorAttributes(doc)
	removeComments(doc)

	// Find the body tag or fall back to the root node
	var bodyNode *html.Node
	var findBody func(*html.Node)
	findBody = func(n *html.Node) {
		if n.Type == html.ElementNode && n.Data == "body" {
			bodyNode = n
			return
		}
		for c := n.FirstChild; c != nil; c = c.NextSibling {
			findBody(c)
		}
	}
	findBody(doc)

	// Render only the body’s content, or the entire document if no <body> is found
	var buf bytes.Buffer
	w := io.Writer(&buf)
	if bodyNode != nil {
		for c := bodyNode.FirstChild; c != nil; c = c.NextSibling {
			if err := html.Render(w, c); err != nil {
				return "", err
			}
		}
	} else {
		if err := html.Render(w, doc); err != nil {
			return "", err
		}
	}

	processedContent := buf.String()
	processedContent = html.UnescapeString(processedContent)

	cleanedContent = processBaseCleaning(processedContent)

	return cleanedContent, nil
}

// processBaseCleaning processes the standard cleaning of the HTML document with string operations.
// It returns the cleaned HTML content as a string.
func processBaseCleaning(content string) (cleanedContent string) {
	var processedContent string

	// Remove all trailing and leading whitespace
	processedContent = strings.Trim(content, " ")

	// Remove all trailing and leading carriage returns
	processedContent = strings.Trim(content, "\n")
	processedContent = strings.Trim(content, "\r")

	// Remove all trailing and leading elements that are not tags
	processedContent = strings.TrimLeftFunc(processedContent, func(r rune) bool { return r != '<' })
	processedContent = strings.TrimRightFunc(processedContent, func(r rune) bool { return r != '>' })

	// Remove any leading "<!DOCTYPE html>" tag
	processedContent = strings.TrimPrefix(processedContent, "<!DOCTYPE html>")

	// Remove all leading and trailing '<html>', '<head>' and <body> tags
	processedContent = strings.TrimPrefix(processedContent, "<html>")
	processedContent = strings.TrimSuffix(processedContent, "</html>")
	processedContent = strings.TrimPrefix(processedContent, "<head>")
	processedContent = strings.TrimSuffix(processedContent, "</head>")
	processedContent = strings.TrimPrefix(processedContent, "<body>")
	processedContent = strings.TrimSuffix(processedContent, "</body>")

	// Remove any leading and trailing '<section>' tags if there is only one surrounding the HTML content
	if strings.HasPrefix(processedContent, "<section>") && strings.Count(processedContent, "<section>") == 1 {
		processedContent = strings.TrimPrefix(processedContent, "<section>")
		processedContent = strings.TrimSuffix(processedContent, "</section>")
	}

	// Remove any leading and trailing '<article>' tags if there is only one surrounding the HTML content
	if strings.HasPrefix(processedContent, "<article>") && strings.Count(processedContent, "<article>") == 1 {
		processedContent = strings.TrimPrefix(processedContent, "<article>")
		processedContent = strings.TrimSuffix(processedContent, "</article>")
	}

	// Remove any leading and trailing '<div>' tags if there is only one surrounding the HTML content
	if strings.HasPrefix(processedContent, "<div>") && strings.Count(processedContent, "<div>") == 1 {
		processedContent = strings.TrimPrefix(processedContent, "<div>")
		processedContent = strings.TrimSuffix(processedContent, "</div>")
	}

	return processedContent
}

// isElementEmpty checks if an HTML element is empty or contains only whitespace.
// It returns true if the element is empty or contains only whitespace, otherwise it returns false.
func isElementEmpty(n *html.Node) bool {
	if n.FirstChild == nil {
		return true
	}
	for c := n.FirstChild; c != nil; c = c.NextSibling {
		if c.Type == html.TextNode && strings.TrimSpace(c.Data) != "" {
			return false
		}
		if c.Type == html.ElementNode && !isElementEmpty(c) {
			return false
		}
	}
	return true
}

// removeEmptyElements removes specified HTML elements if they are empty or contain only whitespace.
// It takes a pointer to an HTML node and a slice of strings representing the elements to remove.
func removeEmptyElements(n *html.Node, emptyElements []string) {
	var nodesToRemove []*html.Node

	// Traverse the tree and collect nodes to remove
	var traverse func(n *html.Node)
	traverse = func(n *html.Node) {
		for a := n.FirstChild; a != nil; a = a.NextSibling {
			traverse(a)
		}

		if n.Type == html.ElementNode {
			for _, element := range emptyElements {
				if n.Data == element {
					// Check if the element is empty or contains only whitespace
					if isElementEmpty(n) {
						nodesToRemove = append(nodesToRemove, n)
						break
					}
				}
			}
		}
	}

	traverse(n)

	// Remove the collected nodes
	for _, node := range nodesToRemove {
		if node.Parent != nil {
			node.Parent.RemoveChild(node)
		}
	}
}

// removeUndesirableAttributes traverses the HTML nodes and removes class, id, style, and JavaScript interactivity attributes.
// It takes a pointer to an HTML node.
func removeUndesirableAttributes(n *html.Node) {
	for a := n.FirstChild; a != nil; a = a.NextSibling {
		removeUndesirableAttributes(a)
	}

	if n.Type == html.ElementNode {
		// Define a list of attributes to remove
		attributesToRemove := []string{"class", "id", "style", "onclick", "onload", "onmouseover", "onmouseout", "onmousedown", "onmouseup", "onmousemove", "onkeypress", "onkeydown", "onkeyup"}
		for _, attr := range attributesToRemove {
			n.Attr = removeAttribute(n.Attr, attr)
		}
	}
}

// removeUndesirableElements removes specified HTML elements along with their contents and child elements.
// It takes a pointer to an HTML node and a slice of strings representing the elements to remove.
func removeUndesirableElements(n *html.Node, undesirableElements []string) {
	var nodesToRemove []*html.Node

	// Traverse the tree and collect nodes to remove
	var traverse func(n *html.Node)
	traverse = func(n *html.Node) {
		for a := n.FirstChild; a != nil; a = a.NextSibling {
			traverse(a)
		}

		if n.Type == html.ElementNode {
			for _, element := range undesirableElements {
				if n.Data == element {
					nodesToRemove = append(nodesToRemove, n)
					break
				}
			}
		}
	}

	traverse(n)

	// Remove the collected nodes
	for _, node := range nodesToRemove {
		if node.Parent != nil {
			node.Parent.RemoveChild(node)
		}
	}
}

// ensureAnchorAttributes ensures that all <a> tags with href attributes have target="_blank" and rel="noopener noreferrer".
// It takes a pointer to an HTML node.
func ensureAnchorAttributes(n *html.Node) {
	for a := n.FirstChild; a != nil; a = a.NextSibling {
		ensureAnchorAttributes(a)
	}

	if n.Type == html.ElementNode && n.Data == "a" {
		hasHref := false
		hasTargetBlank := false
		hasRelNoopenerNoreferrer := false

		for i := 0; i < len(n.Attr); i++ {
			if n.Attr[i].Key == "href" {
				hasHref = true
			}
			if n.Attr[i].Key == "target" && n.Attr[i].Val == "_blank" {
				hasTargetBlank = true
			}
			if n.Attr[i].Key == "rel" && n.Attr[i].Val == "noopener noreferrer" {
				hasRelNoopenerNoreferrer = true
			}
		}

		if hasHref {
			if !hasTargetBlank || !hasRelNoopenerNoreferrer {
				// Remove any existing target and rel attributes
				n.Attr = removeAttribute(n.Attr, "target")
				n.Attr = removeAttribute(n.Attr, "rel")

				// Add the desired target and rel attributes
				n.Attr = append(n.Attr, html.Attribute{Key: "target", Val: "_blank"})
				n.Attr = append(n.Attr, html.Attribute{Key: "rel", Val: "noopener noreferrer"})
			}
		}
	}
}

// removeAttribute removes an attribute with the given key from the list of attributes.
// It takes a slice of html.Attribute and a string representing the key of the attribute to remove.
// It returns the updated slice of html.Attribute.
func removeAttribute(attrs []html.Attribute, key string) []html.Attribute {
	var result []html.Attribute
	for _, attr := range attrs {
		if attr.Key != key {
			result = append(result, attr)
		}
	}
	return result
}

// removeComments traverses the HTML nodes and removes all HTML comments.
// It takes a pointer to an HTML node.
func removeComments(n *html.Node) {
	for a := n.FirstChild; a != nil; a = a.NextSibling {
		removeComments(a)
	}

	if n.Type == html.CommentNode {
		if n.Parent != nil {
			n.Parent.RemoveChild(n)
		}
	}
}

// CleanAllRSSData cleans the RSS feeds by sending raw content to Mistral AI for processing.
// It updates the database with the cleaned content.
// It returns an error if any occurs during the process.
func CleanAllRSSData() (err error) {
	// Get the database connection
	db := databases.GetDB()

	// Define the SQL query to fetch the necessary feed details
	query := `
		SELECT
			rss_items.id,
			rss_items.summary_raw,
			rss_items.content_raw
		FROM
			rss_items
		JOIN
			rss_feeds ON rss_items.rss_feed_id = rss_feeds.id
		WHERE
				rss_items.content_formatted IS NULL
			OR
				LENGTH(rss_items.content_formatted) <= 10
	`

	// Execute the SQL query
	rows, err := db.Query(query)
	if err != nil {
		// Log and return any errors encountered while executing the query
		slog.Error("Error executing query", "error", err)
		return err
	}
	// Ensure that rows are closed after processing is complete
	defer func(rows *sql.Rows) {
		err := rows.Close()
		if err != nil {
			// Log any errors encountered while closing the rows
			slog.Error("Error closing the rows after query", "error", err)
		}
	}(rows)

	// Check if there are any rows returned
	if !rows.Next() {
		// No rows returned
		slog.Info("No new RSS items to clean: No rows returned from the query")
		return nil
	}

	// Retrieve the Mistral AI API key from the environment variables
	mistralApiKey := gowebly.Getenv("MISTRAL_API_KEY", "")
	if mistralApiKey == "" {
		// Log an error if the Mistral AI API key is not set
		slog.Error("Missing required environment variable", "details", "Mistral AI API key environment variable is missing")
		return errors.New("mistral API key is not set as an environment variable")
	}

	// Retrieve the Mistral AI model name from the environment variables
	mistralModel := gowebly.Getenv("MISTRAL_MODEL_TINY", "")
	if mistralModel == "" {
		// Log an error if the Mistral AI model name is not set
		slog.Error("Missing required environment variable", "details", "Mistral AI model name environment variable is missing")
		return errors.New("mistral model is not set as an environment variable")
	}

	// Retrieve the Mistral AI text prompt used to instruct the model
	// Define the file path
	filePath := "assets/prompt/speakrine_prompt_clean_article_content_v1.txt"

	// Read the file content
	content, err := os.ReadFile(filePath)
	if err != nil {
		log.Fatalf("Failed to read file: %v", err)
	}

	// Convert the content to a string
	mistralPrompt := string(content)

	// Process the rows and convert them into RssFeed objects
	for rows.Next() {
		var articleId int
		var articleRawSummary sql.NullString
		var articleRawContent sql.NullString

		err := rows.Scan(&articleId, &articleRawSummary, &articleRawContent)
		if err != nil {
			slog.Error("Error scanning row", "error", err)
			return err
		}

		// Check if both articleCleanSummary and articleCleanContent are not NULL and not empty
		if !articleRawContent.Valid || len(articleRawContent.String) <= 10 {
			slog.Info("Bypassing the cleaning of article with an empty content", "rss_article.id", articleId)
			continue
		}

		// Here call the cleanHTMLContent function to clean the articleRawContent
		// The output of the cleanHTMLContent function is the input of the Mistral AI model
		cleanedContent, err := cleanHTMLContent(articleRawContent.String)
		if err != nil {
			slog.Error("Error cleaning HTML content", "error", err)
			return err
		}

		// Define the Mistral AI client
		client := mistral.NewMistralClientDefault(mistralApiKey)

		// Using Chat Completions
		mistralQuery, err := client.Chat(mistralModel, []mistral.ChatMessage{{Content: fmt.Sprintf("%v \n %v", mistralPrompt, cleanedContent), Role: mistral.RoleUser}}, nil)
		if err != nil {
			slog.Error("Error retrieving the chat completion from Mistral AI", "error", err)
			return err
		}

		mistralResponse := mistralQuery.Choices[0].Message.Content

		mistralOutputCleaned, err := cleanHTMLContent(mistralResponse)
		if err != nil {
			slog.Error("Error cleaning HTML content", "error", err)
			return err
		}

		// Update the database with the cleaned content
		updateQuery := `
			UPDATE rss_items
			SET content_formatted = ?
			WHERE id = ?
		`

		_, err = db.Exec(updateQuery, mistralOutputCleaned, articleId)
		if err != nil {
			slog.Error("Error updating the database with cleaned content", "error", err)
			return err
		}

		slog.Info("An article content has been cleaned", "rss_article.id", articleId)
	}

	// Check for any errors encountered during iteration
	if err = rows.Err(); err != nil {
		slog.Error("Error iterating over rows", "error", err)
		return err
	}

	return nil
}

// CleanRSSData cleans the RSS feeds for a specific user by sending raw content to Mistral AI for processing.
// It updates the database with the cleaned content.
// It returns an error if any occurs during the process.
func CleanRSSData(accountId int) (err error) {
	// Validate the function parameters
	if accountId <= 0 {
		return errors.New("invalid account ID")
	}

	// Get the database connection
	db := databases.GetDB()

	// Define the SQL query to fetch the necessary feed details
	query := `
		SELECT
			rss_items.id,
			rss_items.summary_raw,
			rss_items.content_raw
		FROM
			rss_items
		JOIN
			rss_feeds ON rss_items.rss_feed_id = rss_feeds.id
		WHERE
			rss_feeds.user_id = ?
			AND (rss_items.content_formatted IS NULL OR LENGTH(rss_items.content_formatted) <= 10)
	`

	// Execute the SQL query
	rows, err := db.Query(query)
	if err != nil {
		// Log and return any errors encountered while executing the query
		slog.Error("Error executing query", "error", err)
		return err
	}
	// Ensure that rows are closed after processing is complete
	defer func(rows *sql.Rows) {
		err := rows.Close()
		if err != nil {
			// Log any errors encountered while closing the rows
			slog.Error("Error closing the rows after query", "error", err)
		}
	}(rows)

	// Check if there are any rows returned
	if !rows.Next() {
		// No rows returned
		slog.Info("No new RSS items to clean: No rows returned from the query")
		return nil
	}

	// Retrieve the Mistral AI API key from the environment variables
	mistralApiKey := gowebly.Getenv("MISTRAL_API_KEY", "")
	if mistralApiKey == "" {
		// Log an error if the Mistral AI API key is not set
		slog.Error("Missing required environment variable", "details", "Mistral AI API key environment variable is missing")
		return errors.New("mistral API key is not set as an environment variable")
	}

	// Retrieve the Mistral AI model name from the environment variables
	mistralModel := gowebly.Getenv("MISTRAL_MODEL_TINY", "")
	if mistralModel == "" {
		// Log an error if the Mistral AI model name is not set
		slog.Error("Missing required environment variable", "details", "Mistral AI model name environment variable is missing")
		return errors.New("mistral model is not set as an environment variable")
	}

	// Retrieve the Mistral AI text prompt used to instruct the model
	// Define the file path
	filePath := "assets/prompt/speakrine_prompt_clean_article_content_v1.txt"

	// Read the file content
	content, err := os.ReadFile(filePath)
	if err != nil {
		log.Fatalf("Failed to read file: %v", err)
	}

	// Convert the content to a string
	mistralPrompt := string(content)

	// Process the rows and convert them into RssFeed objects
	for rows.Next() {
		var articleId int
		var articleRawSummary sql.NullString
		var articleRawContent sql.NullString

		err := rows.Scan(&articleId, &articleRawSummary, &articleRawContent)
		if err != nil {
			slog.Error("Error scanning row", "error", err)
			return err
		}

		// Check if both articleCleanSummary and articleCleanContent are not NULL and not empty
		if !articleRawContent.Valid || len(articleRawContent.String) <= 10 {
			slog.Info("Bypassing the cleaning of article with an empty content", "rss_article.id", articleId)
			continue
		}

		// Here call the cleanHTMLContent function to clean the articleRawContent
		// The output of the cleanHTMLContent function is the input of the Mistral AI model
		cleanedContent, err := cleanHTMLContent(articleRawContent.String)
		if err != nil {
			slog.Error("Error cleaning HTML content", "error", err)
			return err
		}

		// Define the Mistral AI client
		client := mistral.NewMistralClientDefault(mistralApiKey)

		// Using Chat Completions
		mistralQuery, err := client.Chat(mistralModel, []mistral.ChatMessage{{Content: fmt.Sprintf("%v \n %v", mistralPrompt, cleanedContent), Role: mistral.RoleUser}}, nil)
		if err != nil {
			slog.Error("Error retrieving the chat completion from Mistral AI", "error", err)
			return err
		}

		mistralResponse := mistralQuery.Choices[0].Message.Content

		mistralOutputCleaned, err := cleanHTMLContent(mistralResponse)
		if err != nil {
			slog.Error("Error cleaning HTML content", "error", err)
			return err
		}

		// Update the database with the cleaned content
		updateQuery := `
			UPDATE rss_items
			SET content_formatted = ?
			WHERE id = ?
		`

		_, err = db.Exec(updateQuery, mistralOutputCleaned, articleId)
		if err != nil {
			slog.Error("Error updating the database with cleaned content", "error", err)
			return err
		}

		slog.Info("An article content has been cleaned", "rss_article.id", articleId)
	}

	// Check for any errors encountered during iteration
	if err = rows.Err(); err != nil {
		slog.Error("Error iterating over rows", "error", err)
		return err
	}

	return nil
}
