package database

import (
	"context"
	"fmt"
	"golang-url-shortener/internal/database/model"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"log"
	"os"
	"strconv"
	"time"
	
	_ "github.com/go-sql-driver/mysql"
	_ "github.com/joho/godotenv/autoload"
)

// Service represents a service that interacts with a database.
type Service interface {
	// Health returns a map of health status information.
	// The keys and values in the map are service-specific.
	Health() map[string]string
	
	// Close terminates the database connection.
	// It returns an error if the connection cannot be closed.
	Close() error
	
	// Query executes a query that returns rows, typically a SELECT statement.
	// It returns the resulting rows and an error if the query fails.
	Query(query string, args ...interface{}) (*gorm.DB, error)
	
	// Return Gorm DB
	ToGormDB() *gorm.DB
	
	// Migrate the database
	Migrate() error
}

type service struct {
	db *gorm.DB
}

var (
	dbname     = os.Getenv("BLUEPRINT_DB_DATABASE")
	password   = os.Getenv("BLUEPRINT_DB_PASSWORD")
	username   = os.Getenv("BLUEPRINT_DB_USERNAME")
	port       = os.Getenv("BLUEPRINT_DB_PORT")
	host       = os.Getenv("BLUEPRINT_DB_HOST")
	dbInstance *service
)

func New() Service {
	// Reuse Connection
	if dbInstance != nil {
		return dbInstance
	}
	
	// Opening a driver typically will not attempt to connect to the database.
	dsn := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?parseTime=true", username, password, host, port, dbname)
	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{})
	//db, err := sql.Open("mysql", fmt.Sprintf("%s:%s@tcp(%s:%s)/%s", username, password, host, port, dbname))
	if err != nil {
		// This will not be a connection error, but a DSN parse error or
		// another initialization error.
		log.Fatal(err)
	}
	
	sqlDB, err := db.DB()
	if err != nil {
		log.Fatal(err)
	}
	
	sqlDB.SetConnMaxLifetime(0)
	sqlDB.SetMaxIdleConns(50)
	sqlDB.SetMaxOpenConns(50)
	
	dbInstance = &service{
		db: db,
	}
	
	return dbInstance
}

// Health checks the health of the database connection by pinging the database.
// It returns a map with keys indicating various health statistics.
func (s *service) Health() map[string]string {
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()
	
	stats := make(map[string]string)
	
	sqlDB, err := s.db.DB()
	if err != nil {
		log.Fatal(err)
	}
	
	// Ping the database
	err = sqlDB.PingContext(ctx)
	if err != nil {
		stats["status"] = "down"
		stats["error"] = fmt.Sprintf("db down: %v", err)
		log.Fatalf("db down: %v", err) // Log the error and terminate the program
		return stats
	}
	
	// Database is up, add more statistics
	stats["status"] = "up"
	stats["message"] = "It's healthy"
	
	// Get database stats (like open connections, in use, idle, etc.)
	dbStats := sqlDB.Stats()
	stats["open_connections"] = strconv.Itoa(dbStats.OpenConnections)
	stats["in_use"] = strconv.Itoa(dbStats.InUse)
	stats["idle"] = strconv.Itoa(dbStats.Idle)
	stats["wait_count"] = strconv.FormatInt(dbStats.WaitCount, 10)
	stats["wait_duration"] = dbStats.WaitDuration.String()
	stats["max_idle_closed"] = strconv.FormatInt(dbStats.MaxIdleClosed, 10)
	stats["max_lifetime_closed"] = strconv.FormatInt(dbStats.MaxLifetimeClosed, 10)
	
	// Evaluate stats to provide a health message
	if dbStats.OpenConnections > 40 { // Assuming 50 is the max for this example
		stats["message"] = "The database is experiencing heavy load."
	}
	if dbStats.WaitCount > 1000 {
		stats["message"] = "The database has a high number of wait events, indicating potential bottlenecks."
	}
	
	if dbStats.MaxIdleClosed > int64(dbStats.OpenConnections)/2 {
		stats["message"] = "Many idle connections are being closed, consider revising the connection pool settings."
	}
	
	if dbStats.MaxLifetimeClosed > int64(dbStats.OpenConnections)/2 {
		stats["message"] = "Many connections are being closed due to max lifetime, consider increasing max lifetime or revising the connection usage pattern."
	}
	
	return stats
}

// Close closes the database connection.
// It logs a message indicating the disconnection from the specific database.
// If the connection is successfully closed, it returns nil.
// If an error occurs while closing the connection, it returns the error.
func (s *service) Close() error {
	log.Printf("Disconnected from database: %s", dbname)
	sqlDB, err := s.db.DB()
	if err != nil {
		log.Fatal(err)
	}
	return sqlDB.Close()
}

// Query executes a query that returns rows, typically a SELECT statement.
// It returns the resulting rows and an error if the query fails.
func (s *service) Query(query string, args ...interface{}) (*gorm.DB, error) {
	// Execute a raw SQL query using GORM's Raw function.
	result := s.db.Raw(query, args...)
	if result.Error != nil {
		return nil, result.Error
	}
	return result, nil
}

func (s *service) Migrate() error {
	err := s.db.AutoMigrate(&model.Users{}, &model.Shortens{})
	if err != nil {
		log.Println("Database migration failed:", err)
		return err
	}
	log.Println("Database Migration Completed!")
	return nil
}

func (s *service) ToGormDB() *gorm.DB {
	return s.db
}
