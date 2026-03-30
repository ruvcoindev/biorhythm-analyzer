package storage

import (
	"encoding/json"
	"os"
	"time"

	"biorhythm-analyzer/internal/models"
)

func LoadPeople() ([]models.Person, error) {
	data, err := os.ReadFile("data/people.json")
	if err != nil {
		if os.IsNotExist(err) {
			return []models.Person{}, nil
		}
		return nil, err
	}
	var people []models.Person
	err = json.Unmarshal(data, &people)
	return people, err
}

func SavePeople(people []models.Person) error {
	data, err := json.MarshalIndent(people, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile("data/people.json", data, 0644)
}

func SaveHistory(people []models.Person, now time.Time, pairs []models.PairData) error {
	entry := models.HistoryEntry{
		Timestamp: now.Format("2006-01-02 15:04:05"),
		Pairs:     pairs,
	}
	var history []models.HistoryEntry
	data, _ := os.ReadFile("data/history.log.json")
	if len(data) > 0 {
		json.Unmarshal(data, &history)
	}
	history = append(history, entry)
	newData, _ := json.MarshalIndent(history, "", "  ")
	return os.WriteFile("data/history.log.json", newData, 0644)
}

func EnsureDataDir() error {
	return os.MkdirAll("data", 0755)
}
