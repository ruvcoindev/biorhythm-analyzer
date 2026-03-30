package main

import (
	"fmt"
	"log"
	"biorhythm-analyzer/internal/storage"
)

func main() {
	people, err := storage.LoadPeople()
	if err != nil {
		log.Fatal(err)
	}
	
	fmt.Printf("Loaded %d people\n", len(people))
	for i, p := range people {
		fmt.Printf("%d: %s, born: %s\n", i, p.Name, p.BirthDate.Format("02.01.2006"))
	}
}
