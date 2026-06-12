package main

import (
	"database/sql"
	"flag"
	"fmt"
	"log"

	_ "github.com/go-sql-driver/mysql"
	"github.com/schollz/progressbar/v3"
)

func main() {
	configPath := flag.String("config", "config.json", "Path to the configuration file")
	progressFlag := flag.Bool("progress", false, "Enable progress bars for querying and serialization")
	helpFlag := flag.Bool("h", false, "Print help info and exit")
	helpFlagLong := flag.Bool("help", false, "Print help info and exit")

	flag.Usage = func() {
		fmt.Println("eWHALES v1 Exporter Tool")
		fmt.Println("Usage: exporter [options]")
		fmt.Println("\nOptions:")
		flag.PrintDefaults()
	}

	flag.Parse()

	if *helpFlag || *helpFlagLong {
		flag.Usage()
		return
	}

	// 1. Configuration Phase
	fmt.Println("Step 1: Configuration Phase")
	config, err := LoadConfig(*configPath)
	if err != nil {
		log.Fatalf("Error loading config: %v", err)
	}

	if config.Port == 0 {
		config.Port = 3306
	}

	fmt.Printf("  - Database Host : %s\n", config.Host)
	fmt.Printf("  - Database Port : %d\n", config.Port)
	fmt.Printf("  - Database Name : %s\n", config.Database)
	fmt.Printf("  - Database User : %s\n", config.Username)
	fmt.Printf("  - Target Table  : %s\n", config.Table)
	fmt.Printf("  - Logbooks CSV  : logbooks_%s\n", config.CSVBaseName)
	fmt.Printf("  - Entries CSV   : %s\n", config.CSVBaseName)

	dsn := fmt.Sprintf("%s:%s@tcp(%s:%d)/%s",
		config.Username, config.Password, config.Host, config.Port, config.Database)

	// 2. Database Connection Phase
	fmt.Println("\nStep 2: Database Connection Phase")
	db, err := sql.Open("mysql", dsn)
	if err != nil {
		log.Fatalf("Error opening database: %v", err)
	}
	defer db.Close()

	if err := db.Ping(); err != nil {
		log.Fatalf("Error connecting to database: %v", err)
	}
	fmt.Println("Successfully connected to MySQL database.")

	// 3. Querying Phase
	fmt.Println("\nStep 3: Querying Phase")
	var queryProgressCallback func(int, int)
	if *progressFlag {
		var queryBar *progressbar.ProgressBar
		queryProgressCallback = func(processed int, total int) {
			if queryBar == nil {
				queryBar = progressbar.Default(int64(total), "Querying database")
			}
			queryBar.Set(processed)
		}
	}
	pivotData, err := QueryPivotData(db, config, queryProgressCallback)
	if err != nil {
		log.Fatalf("Error querying pivot data: %v", err)
	}

	// 4. Serialization Phase
	fmt.Println("\nStep 4: Serialization Phase")
	serializer := &CSVSerializer{
		LogbooksFile:       "logbooks_" + config.CSVBaseName,
		LogbookEntriesFile: config.CSVBaseName,
		IdsToFields:        config.IdsToFields,
	}

	var serializeProgressCallback func(int, int)
	if *progressFlag {
		var serializeBar *progressbar.ProgressBar
		serializeProgressCallback = func(processed int, total int) {
			if serializeBar == nil {
				serializeBar = progressbar.Default(int64(total), "Writing to CSV   ")
			}
			serializeBar.Set(processed)
		}
	}
	err = serializer.Serialize(*pivotData, serializeProgressCallback)
	if err != nil {
		log.Fatalf("Error during serialization: %v", err)
	}

	fmt.Printf("\nSuccessfully exported data to %s and %s\n", serializer.LogbooksFile, serializer.LogbookEntriesFile)
}
