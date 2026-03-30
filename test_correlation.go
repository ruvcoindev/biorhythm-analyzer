package main

import (
	"fmt"
	"time"
	"biorhythm-analyzer/internal/metrics"
	"biorhythm-analyzer/internal/models"
)

func main() {
	now := time.Now()
	
	people := []models.Person{
		{Name: "Дина", BirthDate: time.Date(1970, 10, 3, 0, 0, 0, 0, time.UTC)},
		{Name: "Виталий", BirthDate: time.Date(1974, 10, 15, 0, 0, 0, 0, time.UTC)},
	}
	
	fmt.Printf("Testing correlation between %s and %s\n", people[0].Name, people[1].Name)
	
	result := metrics.CalculateCorrelation(people[0], people[1], now)
	
	fmt.Printf("Result: %f\n", result)
}
