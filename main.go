package main

import (
	"database/sql"
	"flag"
	"fmt"
	"log"

	_ "github.com/go-sql-driver/mysql"
)

func main() {
	configPath := flag.String("config", "config.json", "Path to the configuration file")
	flag.Parse()

	// 1. Configuration Phase
	config, err := LoadConfig(*configPath)
	if err != nil {
		log.Fatalf("Error loading config: %v", err)
	}

	if config.Port == 0 {
		config.Port = 3306
	}

	dsn := fmt.Sprintf("%s:%s@tcp(%s:%d)/%s",
		config.Username, config.Password, config.Host, config.Port, config.Database)

	// 2. Database Connection Phase
	db, err := sql.Open("mysql", dsn)
	if err != nil {
		log.Fatalf("Error opening database: %v", err)
	}
	defer db.Close()

	if err := db.Ping(); err != nil {
		log.Fatalf("Error connecting to database: %v", err)
	}
	log.Println("Successfully connected to MySQL database.")

	// 3. Querying Phase (defined in query.go)
	// We extract PivotData, which contains massive in-memory structures (defined in models.go)
	pivotData, err := QueryPivotData(db, config)
	if err != nil {
		log.Fatalf("Error querying pivot data: %v", err)
	}

	// 4. Serialization Phase (defined in serializer.go)
	// Here we define the CSV serializer, but it implements a generic Serializer interface.
	serializer := &CSVSerializer{
		LogbooksFile:       "logbooks.csv",
		LogbookEntriesFile: "logbook_entries.csv",
		IdsToFields:        config.IdsToFields,
	}

	log.Println("Serializing to CSV...")
	if err := serializer.Serialize(*pivotData); err != nil {
		log.Fatalf("Error during serialization: %v", err)
	}

	log.Println("Successfully exported data.")
}
