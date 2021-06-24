package main

import (
	"bytes"
	"encoding/csv"
	"fmt"
	"io/ioutil"
	"path/filepath"
	"strconv"
	"strings"
)

type Hits struct {
	MaxScore int
	csv      map[string]int
}

func getHitsScorer(csvPath string, maxScore int) Hits {
	csvContent, err := parseCSV(csvPath)
	if err != nil {
		fmt.Printf("Could not get hits scorer:%s\n", err.Error())
	}

	return Hits{
		MaxScore: maxScore,
		csv:      csvContent,
	}
}

func (hits Hits) Score(fileEntries []fileEntry) []fileEntry {
	for index, fileEntry := range fileEntries {
		score, hits := hits.scoreFile(filepath.Dir(fileEntry.relative))
		fileEntry.score += score
		fileEntry.notes = append(fileEntry.notes, fmt.Sprintf("Hits: %d\n", hits))
		fileEntries[index] = fileEntry
	}
	return fileEntries
}

//scoreFile returns the score as a portion of max score and the full page hits
func (hits Hits) scoreFile(url string) (int, int) {
	url = strings.TrimPrefix(url, "en/")
	url = url + "/"

	switch score := hits.csv[url]; {
	case score > 2000:
		return maxScore, score
	case score > 400:
		return maxScore * 2 / 3, score
	case score > 10:
		return maxScore * 1 / 3, score
	default:
		return 0, score
	}
}

func parseCSV(path string) (map[string]int, error) {
	content, err := ioutil.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("unable to open CSV: %s", err.Error())
	}

	reader := csv.NewReader(bytes.NewReader(content))
	records, err := reader.ReadAll()
	if err != nil {
		return nil, fmt.Errorf("unable to parse CSV: %s", err.Error())
	}

	recordsMap := map[string]int{}
	for _, record := range records {
		recordCount, _ := strconv.Atoi(record[2])
		recordName := strings.TrimSuffix(record[0], "index.html")
		recordName = strings.TrimPrefix(recordName, "/latest/")
		recordsMap[recordName] = recordsMap[recordName] + recordCount
	}

	for key, val := range recordsMap {
		fmt.Printf("%s: %+v\n", key, val)
	}

	return recordsMap, nil
}
