# Speakrine project 👱‍♀️

_Speakrine_ is a Go application designed to fetch, clean, and manage data from RSS feeds. It periodically fetches RSS feed data, processes it to remove unwanted HTML content, and stores the cleaned data in a MySQL database.

## Table of Contents

- [Features](#features)
- [Configuration](#configuration)
- [Usage](#usage)
- [Project Structure](#project-structure)
- [License](#license)

## Features

- **Periodic RSS Feed Fetching**: Fetches RSS feeds at regular intervals (every 20 minutes).
- **HTML Content Cleaning**: Cleans HTML content from RSS feed items to ensure only plain text is stored.
- **Database Integration**: Stores fetched and cleaned RSS data in a MySQL database.
- **Environment Configuration**: Uses environment variables for configuration, including database credentials and Mistral AI API key.

## Configuration

The application uses environment variables for configuration. Create a `.env` file based on the `.env.example` file and fill in the required values.
All environment variables are required and should be provided as strings.

```env
DB_HOST=your_db_host
DB_PORT=your_db_port
DB_DATABASE=your_db_name
DB_USERNAME=your_db_username
DB_PASSWORD=your_db_password
MISTRAL_API_KEY=your_mistral_api_key
MISTRAL_MODEL_TINY=your_mistral_model_tiny
MISTRAL_MODEL_MEDIUM=your_mistral_model_medium
FETCH_INTERVAL=rss_feed_fetch_time_interval_in_minutes
```

## Usage

The application runs in two main processes:

1. **Fetching RSS Feeds**: Periodically fetches RSS feed data and stores it in the database.
2. **Cleaning RSS Data**: Cleans the HTML content from the fetched RSS data and updates the database with the cleaned content.

The main process is defined in `main.go`, which initializes a ticker to run the fetching and cleaning processes every 20 minutes.

## Project Structure

### `root` directory

- **`main.go`**: Main entry point of the application.
- **`go.mod`**: Go module file listing dependencies.
- **`.env.example`**: Example environment variables file. The actual environment variables should be stored in a `.env` file and be based on this example.
- **`README.md`**: Contains a brief overview of the project.

### `databases` directory

- **`databases/dbconnect.go`**: Contains functions to initialize and manage the database connection.

### `functions` directory

- **`functions/rss_clean.go`**: Contains functions to clean HTML content from RSS feed items.
- **`functions/rss_fetch.go`**: Contains functions to fetch RSS feed data and store it in the database.

### `assets` directory

The assets directory contains any static files or resources used by the application such as LLM prompts used for cleaning RSS data or the project logo etc.

## License

This project is licensed under the Creative Commons BY-NC-SA 4.0 License - see the [LICENSE](LICENSE) file for details.

Attribution Clément CHAIB-GALLI - https://github.com/cl3mcg
