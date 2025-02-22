package databases

import (
	"database/sql"
	"fmt"
	"log/slog"
	"os"

	_ "github.com/go-sql-driver/mysql"
	"github.com/gowebly/helpers"
	_ "github.com/joho/godotenv/autoload"
)

// The db variable holds the database connection pool.
var db *sql.DB

// init initializes the database connection using environment variables.
//
// This function ensures that all required environment variables are present and then
// establishes a connection to the MySQL database. If any step fails, the
// program logs the error and terminates.
func init() {
	// Load environment variables for database connection.
	dbHost := gowebly.Getenv("DB_HOST", "")
	dbPort := gowebly.Getenv("DB_PORT", "")
	dbName := gowebly.Getenv("DB_DATABASE", "")
	dbUser := gowebly.Getenv("DB_USERNAME", "")
	dbPass := gowebly.Getenv("DB_PASSWORD", "")

	// Ensure all required environment variables are provided.
	if dbHost == "" || dbPort == "" || dbName == "" || dbUser == "" || dbPass == "" {
		slog.Error("Missing required environment variable", "details", "One or more required database environment variables are missing.")
		os.Exit(1)
	}

	// Construct the Data Source CommonName (DSN) for connecting to the database.
	dsn := fmt.Sprintf("%v:%v@tcp(%v:%v)/%v", dbUser, dbPass, dbHost, dbPort, dbName)

	// Open a connection to the database.
	var err error
	db, err = sql.Open("mysql", dsn)
	if err != nil {
		slog.Error("Failed to connect to the database", "details", err.Error())
		os.Exit(1)
	}

	// Test the database connection by pinging it.
	if err := db.Ping(); err != nil {
		slog.Error("Failed to ping the database", "details", err.Error())
		os.Exit(1)
	}

	// Log success once the database connection is successful.
	slog.Info("Successfully connected to the database")
}

// GetDB returns a pointer to the sql.DB instance representing the database connection.
// This function allows other packages to use the established database connection.
//
// Returns:
//   - *sql.DB - A pointer to the database connection instance.
func GetDB() *sql.DB {
	return db
}
