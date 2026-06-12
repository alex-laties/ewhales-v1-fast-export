package main

import (
	"encoding/csv"
	"fmt"
	"os"
)

// Serializer is an interface for serializing the PivotData.
// We are documenting this to show how easy it is to add a JSONSerializer, XMLSerializer, etc.
// The main application logic will rely on this interface to decouple querying from output formatting.
type Serializer interface {
	Serialize(data PivotData) error
}

// CSVSerializer implements Serializer for CSV format.
type CSVSerializer struct {
	LogbooksFile       string
	LogbookEntriesFile string
	IdsToFields        map[string]string
}

// Serialize writes the PivotData out to CSV files.
// It uses IdsToFields to map internal struct names to user-friendly headers.
// It also maps LogbookEntry LogbookIDs to their corresponding Logbook string names.
func (s *CSVSerializer) Serialize(data PivotData) error {
	// Build Logbook mapping for entries
	logbookNameMap := make(map[uint]string)
	for _, lb := range data.Logbooks {
		logbookNameMap[lb.PostID] = lb.LogbookID
	}

	// Helper to get header
	getHeader := func(key string) string {
		if val, ok := s.IdsToFields[key]; ok {
			return val
		}
		return key
	}

	// Serialize Logbooks
	lbFile, err := os.Create(s.LogbooksFile)
	if err != nil {
		return fmt.Errorf("error creating logbooks file: %v", err)
	}
	defer lbFile.Close()

	lbWriter := csv.NewWriter(lbFile)
	if err := lbWriter.Write([]string{getHeader("post_id"), getHeader("logbook_id")}); err != nil {
		return err
	}
	for _, lb := range data.Logbooks {
		if err := lbWriter.Write([]string{fmt.Sprint(lb.PostID), lb.LogbookID}); err != nil {
			return err
		}
	}
	lbWriter.Flush()

	// Serialize Logbook Entries
	lbeFile, err := os.Create(s.LogbookEntriesFile)
	if err != nil {
		return fmt.Errorf("error creating logbook entries file: %v", err)
	}
	defer lbeFile.Close()

	lbeWriter := csv.NewWriter(lbeFile)
	keys := []string{
		"post_id", "logbook_id", "bottom", "cloud_cover", "depth", "depth_unit",
		"entry_date", "landmark", "latitude", "local_time", "longitude",
		"page", "sea_state", "ship_heading", "ship_sightings", "wind_direction", "wind_force",
	}
	var headers []string
	for _, k := range keys {
		headers = append(headers, getHeader(k))
	}

	if err := lbeWriter.Write(headers); err != nil {
		return err
	}
	for _, lbe := range data.LogbookEntries {
		logbookName := logbookNameMap[lbe.LogbookID]
		if logbookName == "" {
			logbookName = fmt.Sprint(lbe.LogbookID) // Fallback if missing
		}

		record := []string{
			fmt.Sprint(lbe.PostID), logbookName, lbe.Bottom, lbe.CloudCover,
			lbe.Depth, lbe.DepthUnit, lbe.EntryDate, lbe.Landmark, lbe.Latitude,
			lbe.LocalTime, lbe.Longitude, lbe.Page, lbe.SeaState, lbe.ShipHeading,
			lbe.ShipSightings, lbe.WindDirection, lbe.WindForce,
		}
		if err := lbeWriter.Write(record); err != nil {
			return err
		}
	}
	lbeWriter.Flush()

	return nil
}
