package main

import (
	"fmt"
	"time"
	"biorhythm-analyzer/internal/metrics"
	"biorhythm-analyzer/internal/models"
)

func main() {
	now := time.Now()
	
	birth1, _ := time.Parse("2006-01-02", "1974-10-15")
	birth2, _ := time.Parse("2006-01-02", "1970-10-03")
	
	p1 := models.Person{Name: "Виталий", BirthDate: birth1}
	p2 := models.Person{Name: "Дина", BirthDate: birth2}
	
	fmt.Printf("Testing correlation between %s and %s\n", p1.Name, p2.Name)
	fmt.Printf("Biorhythms count: %d\n", len(models.Biorhythms))
	
	r := metrics.CalculateCorrelation(p1, p2, now)
	fmt.Printf("Correlation: %f\n", r)
}
