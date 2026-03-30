package models

import (
	"time"
)

// Phi - золотое сечение
const Phi = 0.618033988749895

// BiorhythmType представляет тип биоритма
type BiorhythmType struct {
	Name   string  `json:"name"`
	Period float64 `json:"period"`
	Color  string  `json:"color"`
}

// Person представляет человека в системе
type Person struct {
	Name      string    `json:"name"`
	BirthDate time.Time `json:"birth_date"`
}

// HistoryEntry запись в истории анализа
type HistoryEntry struct {
	Timestamp string     `json:"timestamp"`
	Pairs     []PairData `json:"pairs"`
}

// PairData данные по одной паре
type PairData struct {
	PersonA   string  `json:"person_a"`
	PersonB   string  `json:"person_b"`
	R         float64 `json:"r"`
	Liquidity float64 `json:"liquidity"`
	Status    string  `json:"status"`
}

// CorrelationMatrix хранит матрицу корреляций
type CorrelationMatrix struct {
	Names []string    `json:"names"`
	Data  [][]float64 `json:"data"`
}

// ComprehensiveAnalysis результат комплексного анализа
type ComprehensiveAnalysis struct {
	Pearson        float64 `json:"pearson"`
	Cosine         float64 `json:"cosine"`
	Spearman       float64 `json:"spearman"`
	Kendall        float64 `json:"kendall"`
	MutualInfo     float64 `json:"mutual_info"`
	DTW            float64 `json:"dtw"`
	SpectralMatch  float64 `json:"spectral_match"`
	WeightedScore  float64 `json:"weighted_score"`
	Recommendation string  `json:"recommendation"`
}

// Расширенная система биоритмов (5 ритмов)
var Biorhythms = []BiorhythmType{
	{Name: "Физический", Period: 23, Color: "#FF6B6B"},
	{Name: "Эмоциональный", Period: 28, Color: "#4ECDC4"},
	{Name: "Интеллектуальный", Period: 33, Color: "#45B7D1"},
	{Name: "Духовный", Period: 38, Color: "#96CEB4"},
	{Name: "Интуитивный", Period: 42, Color: "#FFEAA7"},
}
