package main

import (
	"database/sql"
	"encoding/csv"
	"flag"
	"fmt"
	"log"
	"os"

	_ "github.com/go-sql-driver/mysql"
)

// generateSQLQueries is a stub function that generates the initial SQL query,
// and can be extended to return multiple queries based on logic.
func generateSQLQueries(table string) []string {
	// For now, it just returns a basic SELECT * query.
	// We expect to create more queries based on the initial query later.
	return []string{
		fmt.Sprintf("SELECT * FROM %s", table),
	}
}

func main() {
	configPath := flag.String("config", "config.json", "Path to the configuration file")
	flag.Parse()

	config, err := LoadConfig(*configPath)
	if err != nil {
		log.Fatalf("Error loading config: %v", err)
	}

	if config.CSVOutputName == "" {
		config.CSVOutputName = "output.csv"
	}
	if config.Port == 0 {
		config.Port = 3306
	}

	dsn := fmt.Sprintf("%s:%s@tcp(%s:%d)/%s",
		config.Username, config.Password, config.Host, config.Port, config.Database)

	db, err := sql.Open("mysql", dsn)
	if err != nil {
		log.Fatalf("Error opening database: %v", err)
	}
	defer db.Close()

	if err := db.Ping(); err != nil {
		log.Fatalf("Error connecting to database: %v", err)
	}
	log.Println("Successfully connected to MySQL database.")

	queries := generateSQLQueries(config.Table)

	// Since we are exporting to a single CSV file, we'll execute the first query
	// and export its results. 
	// This might need adjustments once the query generation logic gets more complex.
	if len(queries) == 0 {
		log.Println("No queries generated.")
		return
	}

	query := queries[0]
	log.Printf("Executing query: %s\n", query)

	rows, err := db.Query(query)
	if err != nil {
		log.Fatalf("Error executing query: %v", err)
	}
	defer rows.Close()

	file, err := os.Create(config.CSVOutputName)
	if err != nil {
		log.Fatalf("Error creating CSV file: %v", err)
	}
	defer file.Close()

	writer := csv.NewWriter(file)
	defer writer.Flush()

	columns, err := rows.Columns()
	if err != nil {
		log.Fatalf("Error getting columns: %v", err)
	}

	if err := writer.Write(columns); err != nil {
		log.Fatalf("Error writing headers to CSV: %v", err)
	}

	values := make([]interface{}, len(columns))
	valuePtrs := make([]interface{}, len(columns))
	for i := range values {
		valuePtrs[i] = &values[i]
	}

	var rowCount int
	for rows.Next() {
		if err := rows.Scan(valuePtrs...); err != nil {
			log.Fatalf("Error scanning row: %v", err)
		}

		var stringRecord []string
		for _, val := range values {
			if val == nil {
				stringRecord = append(stringRecord, "")
			} else {
				b, ok := val.([]byte)
				if ok {
					stringRecord = append(stringRecord, string(b))
				} else {
					stringRecord = append(stringRecord, fmt.Sprintf("%v", val))
				}
			}
		}

		if err := writer.Write(stringRecord); err != nil {
			log.Fatalf("Error writing record to CSV: %v", err)
		}
		rowCount++
	}

	if err := rows.Err(); err != nil {
		log.Fatalf("Error iterating over rows: %v", err)
	}

	log.Printf("Successfully exported %d rows to %s", rowCount, config.CSVOutputName)
}
