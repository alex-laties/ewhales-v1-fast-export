package main

import (
	"database/sql"
	"fmt"
	"log"
	"strconv"
	"strings"
)

// PivotData acts as the central in-memory structure holding all parsed entities.
// It is the resulting data structure after querying the EAV database and parsing the properties.
type PivotData struct {
	Logbooks       []Logbook
	LogbookEntries []LogbookEntry
}

// QueryPivotData performs the database interactions to extract and pivot the EAV data.
// It fetches all relevant post_ids first, then batches them to avoid high memory spikes
// and database timeouts. It returns the central PivotData structure.
func QueryPivotData(db *sql.DB, config *Config, onProgress func(processed int, total int)) (*PivotData, error) {
	log.Println("Fetching distinct post_ids...")
	postIDs, err := getDistinctPostIDs(db, config.Table)
	if err != nil {
		return nil, fmt.Errorf("error fetching post IDs: %v", err)
	}

	log.Printf("Found %d distinct post_ids to process.", len(postIDs))
	return processBatches(db, config, postIDs, onProgress)
}

func getDistinctPostIDs(db *sql.DB, tableName string) ([]uint, error) {
	rows, err := db.Query(fmt.Sprintf("SELECT DISTINCT post_id FROM %s WHERE meta_key = 'logbook_id'", tableName))
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var postIDs []uint
	for rows.Next() {
		var id uint
		if err := rows.Scan(&id); err != nil {
			return nil, fmt.Errorf("error scanning post ID: %v", err)
		}
		postIDs = append(postIDs, id)
	}

	return postIDs, nil
}

func processBatches(db *sql.DB, config *Config, postIDs []uint, onProgress func(processed int, total int)) (*PivotData, error) {
	batchSize := 100
	var pivotData PivotData

	metaKeysMap := make(map[string]bool)
	for _, keys := range config.PostTypeToMetaKeys {
		for _, key := range keys {
			metaKeysMap[key] = true
		}
	}

	var metaKeys []string
	var metaKeysArgs []interface{}
	for k := range metaKeysMap {
		metaKeys = append(metaKeys, "?")
		metaKeysArgs = append(metaKeysArgs, k)
	}
	metaKeysInClause := strings.Join(metaKeys, ",")

	for i := 0; i < len(postIDs); i += batchSize {
		end := i + batchSize
		if end > len(postIDs) {
			end = len(postIDs)
		}
		batchIDs := postIDs[i:end]

		var idPlaceholders []string
		var args []interface{}
		for _, id := range batchIDs {
			idPlaceholders = append(idPlaceholders, "?")
			args = append(args, id)
		}

		args = append(args, metaKeysArgs...)

		query := fmt.Sprintf(`
			SELECT post_id, meta_key, meta_value 
			FROM %s 
			WHERE post_id IN (%s) AND meta_key IN (%s) 
			ORDER BY post_id
		`, config.Table, strings.Join(idPlaceholders, ","), metaKeysInClause)

		rows, err := db.Query(query, args...)
		if err != nil {
			return nil, fmt.Errorf("error executing batch query: %v", err)
		}

		var currentPostID uint
		entityProps := make(map[string]string)

		processEntity := func(id uint, props map[string]string) {
			if id == 0 || len(props) == 0 {
				return
			}
			logbookIDVal := props["logbook_id"]
			if numID, err := strconv.ParseUint(logbookIDVal, 10, 64); err == nil {
				entry := LogbookEntry{
					PostID:        id,
					LogbookID:     uint(numID),
					Bottom:        props["bottom"],
					CloudCover:    props["cloud_cover"],
					Depth:         props["depth"],
					DepthUnit:     props["depth_unit"],
					EntryDate:     props["entry_date"],
					Landmark:      props["landmark"],
					Latitude:      props["latitude"],
					LocalTime:     props["local_time"],
					Longitude:     props["longitude"],
					Page:          props["page"],
					SeaState:      props["sea_state"],
					ShipHeading:   props["ship_heading"],
					ShipSightings: props["ship_sightings"],
					WindDirection: props["wind_direction"],
					WindForce:     props["wind_force"],
				}
				pivotData.LogbookEntries = append(pivotData.LogbookEntries, entry)
			} else {
				lb := Logbook{
					PostID:    id,
					LogbookID: logbookIDVal,
				}
				pivotData.Logbooks = append(pivotData.Logbooks, lb)
			}
		}

		for rows.Next() {
			var postID uint
			var metaKey string
			var metaValue sql.NullString

			if err := rows.Scan(&postID, &metaKey, &metaValue); err != nil {
				rows.Close()
				return nil, fmt.Errorf("error scanning row: %v", err)
			}

			if currentPostID != postID {
				processEntity(currentPostID, entityProps)
				currentPostID = postID
				entityProps = make(map[string]string)
			}

			if metaValue.Valid {
				entityProps[metaKey] = metaValue.String
			}
		}
		processEntity(currentPostID, entityProps)
		rows.Close()

		if onProgress != nil {
			onProgress(end, len(postIDs))
		} else {
			log.Printf("Processed batch %d to %d (out of %d)", i+1, end, len(postIDs))
		}
	}

	log.Printf("Found %d Logbooks and %d Logbook Entries", len(pivotData.Logbooks), len(pivotData.LogbookEntries))
	return &pivotData, nil
}
