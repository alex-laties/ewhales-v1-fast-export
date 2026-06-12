package main

import (
	"database/sql"
	"os"
	"strings"
	"testing"

	_ "github.com/mattn/go-sqlite3"
)

func TestQueryPivotData(t *testing.T) {
	// 1. Read the test SQL data
	sqlBytes, err := os.ReadFile("test/test_simple.sql")
	if err != nil {
		t.Fatalf("Failed to read test_simple.sql: %v", err)
	}
	queries := strings.Split(string(sqlBytes), ";")

	// 2. Open SQLite in-memory connection
	db, err := sql.Open("sqlite3", "file::memory:?cache=shared")
	if err != nil {
		t.Fatalf("Failed to open RamSQL: %v", err)
	}
	defer db.Close()

	// 3. Populate database
	for _, q := range queries {
		q = strings.TrimSpace(q)
		if q == "" {
			continue
		}
		if _, err := db.Exec(q); err != nil {
			t.Fatalf("Failed to execute query '%s': %v", q, err)
		}
	}

	// 4. Create Mock Config
	config := &Config{
		Table: "logswp_postmeta",
		PostTypeToMetaKeys: map[string][]string{
			"logbook":       {"logbook_id"},
			"logbook_entry": {"logbook_id", "cloud_cover", "sea_state"},
		},
	}

	// 5. Test QueryPivotData
	pivotData, err := QueryPivotData(db, config, nil)
	if err != nil {
		t.Fatalf("QueryPivotData returned error: %v", err)
	}

	// 6. Assertions
	if len(pivotData.Logbooks) != 1 {
		t.Errorf("Expected 1 Logbook, got %d", len(pivotData.Logbooks))
	} else {
		if pivotData.Logbooks[0].LogbookID != "Test Logbook 1888" {
			t.Errorf("Expected LogbookID to be 'Test Logbook 1888', got %s", pivotData.Logbooks[0].LogbookID)
		}
		if pivotData.Logbooks[0].PostID != 100 {
			t.Errorf("Expected Logbook PostID to be 100, got %d", pivotData.Logbooks[0].PostID)
		}
	}

	if len(pivotData.LogbookEntries) != 1 {
		t.Errorf("Expected 1 LogbookEntry, got %d", len(pivotData.LogbookEntries))
	} else {
		entry := pivotData.LogbookEntries[0]
		if entry.PostID != 200 {
			t.Errorf("Expected LogbookEntry PostID to be 200, got %d", entry.PostID)
		}
		if entry.LogbookID != 100 {
			t.Errorf("Expected LogbookEntry LogbookID to be 100, got %d", entry.LogbookID)
		}
		if entry.CloudCover != "Partly Cloudy" {
			t.Errorf("Expected CloudCover 'Partly Cloudy', got '%s'", entry.CloudCover)
		}
		if entry.SeaState != "Calm" {
			t.Errorf("Expected SeaState 'Calm', got '%s'", entry.SeaState)
		}
	}
}

func TestQueryPivotDataRealistic(t *testing.T) {
	// 1. Read the realistic test SQL data
	sqlBytes, err := os.ReadFile("test/test_multiple_logbook_logbook_entries.sql")
	if err != nil {
		t.Fatalf("Failed to read sql file: %v", err)
	}
	
	// Strip all CREATE TABLE and CREATE INDEX statements by finding the first INSERT
	sqlStr := string(sqlBytes)
	sqlStr = strings.ReplaceAll(sqlStr, "test.logswp_postmeta", "logswp_postmeta")
	sqlStr = strings.ReplaceAll(sqlStr, "\\'", "''")
	
	idx := strings.Index(sqlStr, "INSERT INTO")
	if idx == -1 {
		t.Fatalf("Could not find any INSERT statements in the realistic data file")
	}
	sqlStr = sqlStr[idx:]

	// 2. Open SQLite in-memory connection
	// Use a new shared memory connection to not conflict with the other test
	db, err := sql.Open("sqlite3", "file::memory:?cache=shared&mode=memory")
	if err != nil {
		t.Fatalf("Failed to open SQLite: %v", err)
	}
	defer db.Close()

	// 3. Create table manually and populate it
	if _, err := db.Exec("CREATE TABLE logswp_postmeta (meta_id INT, post_id INT, meta_key VARCHAR, meta_value VARCHAR);"); err != nil {
		t.Fatalf("Failed to create table: %v", err)
	}

	if _, err := db.Exec(sqlStr); err != nil {
		t.Fatalf("Failed to execute test data script: %v", err)
	}

	// 4. Create Mock Config
	config := &Config{
		Table: "logswp_postmeta",
		PostTypeToMetaKeys: map[string][]string{
			"logbook":       {"logbook_id"},
			"logbook_entry": {"logbook_id", "cloud_cover", "sea_state", "depth", "latitude", "longitude", "ship_heading", "wind_direction", "wind_force"},
		},
	}

	// 5. Test QueryPivotData
	pivotData, err := QueryPivotData(db, config, nil)
	if err != nil {
		t.Fatalf("QueryPivotData returned error: %v", err)
	}

	// 6. Assertions
	if len(pivotData.Logbooks) != 10 {
		t.Errorf("Expected 10 Logbooks, got %d", len(pivotData.Logbooks))
	}

	// Build a map of Logbook PostIDs to verify entries point to valid logbooks
	validLogbooks := make(map[uint]bool)
	for _, lb := range pivotData.Logbooks {
		validLogbooks[lb.PostID] = true
	}

	if len(pivotData.LogbookEntries) == 0 {
		t.Errorf("Expected some LogbookEntries, got 0")
	}

	for i, entry := range pivotData.LogbookEntries {
		if !validLogbooks[entry.LogbookID] {
			t.Errorf("Entry %d (PostID %d) points to unknown LogbookID %d", i, entry.PostID, entry.LogbookID)
		}
	}
}
