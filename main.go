package main

import (
	"log/slog"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/cl3mcg/speakrine/functions"
	gowebly "github.com/gowebly/helpers"
	_ "github.com/joho/godotenv/autoload"
)

func main() {
	// Get environment variables
	strInterval := gowebly.Getenv("FETCH_INTERVAL", "")
	if strInterval == "" {
		slog.Error("The environment variable FETCH_INTERVAL is not set")
		os.Exit(1)
	}

	// Convert the interval to an integer
	interval, err := strconv.Atoi(strings.TrimSpace(strInterval))
	if err != nil {
		slog.Error("The environment variable FETCH_INTERVAL cannot be casted to an int value", "error", err)
		os.Exit(1)
	}

	// Check if the interval is greater than 0
	if interval < 1 {
		slog.Error("The environment variable FETCH_INTERVAL must be greater than 0")
		os.Exit(1)
	}

	// Create a ticker that ticks every interval
	ticker := time.NewTicker(time.Duration(interval) * time.Minute)

	// Initial run
	slog.Info("Starting the rss feed fetching process")
	err = functions.FetchAllRSSData()
	if err != nil {
		slog.Error("The process used to fetch RSS article content has failed", "error", err)
	}
	slog.Info("Ending the rss feed fetching process")
	slog.Info("Starting the rss article cleaning process")
	err = functions.CleanAllRSSData()
	if err != nil {
		slog.Error("The process used to clean RSS article content has failed", "error", err)
	}
	slog.Info("Ending the rss cleaning process")

	// Periodic run
	for range ticker.C {
		slog.Info("Starting the rss feed fetching process")
		err := functions.FetchAllRSSData()
		if err != nil {
			slog.Error("The process used to fetch RSS article content has failed", "error", err)
		}
		slog.Info("Ending the rss feed fetching process")
		slog.Info("Starting the rss article cleaning process")
		err = functions.CleanAllRSSData()
		if err != nil {
			slog.Error("The process used to clean RSS article content has failed", "error", err)
		}
		slog.Info("Ending the rss cleaning process")
	}

}
