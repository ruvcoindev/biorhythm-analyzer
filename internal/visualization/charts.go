package visualization

import (
	"fmt"
	"math"
	"strings"
	"time"

	"biorhythm-analyzer/internal/metrics"
	"biorhythm-analyzer/internal/models"
)

func PrintCorrelationMatrix(people []models.Person, now time.Time) {
	if len(people) == 0 {
		return
	}

	fmt.Println("\n📊 МАТРИЦА КОРРЕЛЯЦИЙ")
	fmt.Println(strings.Repeat("─", 80))

	fmt.Printf("%-15s", "")
	for _, p := range people {
		name := p.Name
		if len(name) > 10 {
			name = name[:9] + "."
		}
		fmt.Printf("%-12s", name)
	}
	fmt.Println()

	for i, p1 := range people {
		name1 := p1.Name
		if len(name1) > 12 {
			name1 = name1[:11] + "."
		}
		fmt.Printf("%-15s", name1)
		for j, p2 := range people {
			var r float64
			if i == j {
				r = 1.0
			} else {
				r = metrics.CalculateCorrelation(p1, p2, now)
			}
			color := getColorForCorrelation(r)
			fmt.Printf("%s%8.4f %s", color, r, "\033[0m")
		}
		fmt.Println()
	}
	fmt.Println(strings.Repeat("─", 80))
}

func PrintBiorhythmStatusWithExplanation(person models.Person, now time.Time) {
	fmt.Printf("\n👤 %s:\n", person.Name)
	fmt.Println(strings.Repeat("─", 60))
	
	for _, br := range models.Biorhythms {
		val := metrics.GetBiorhythm(person.BirthDate, now, br.Period)
		icon := getBiorhythmIconDetailed(val)
		status := getBiorhythmStatusText(val)
		bar := getBarVisual(val, 25)
		
		fmt.Printf("   %s %-14s: %6.3f %s\n", icon, br.Name, val, bar)
		fmt.Printf("      └─ %s\n", status)
	}
	
	// Общая оценка
	avg := 0.0
	for _, br := range models.Biorhythms {
		avg += metrics.GetBiorhythm(person.BirthDate, now, br.Period)
	}
	avg /= float64(len(models.Biorhythms))
	
	fmt.Printf("\n   📊 ОБЩАЯ ОЦЕНКА: ")
	if avg > 0.3 {
		fmt.Printf("ПОДЪЕМ — благоприятный период для активности\n")
	} else if avg < -0.3 {
		fmt.Printf("СПАД — рекомендуется отдых и восстановление\n")
	} else {
		fmt.Printf("НЕЙТРАЛЬНО — стабильное состояние\n")
	}
	fmt.Println()
}

func getColorForCorrelation(r float64) string {
	switch {
	case r > 0.7:
		return "\033[32m" // зеленый
	case r > 0.3:
		return "\033[36m" // голубой
	case r > -0.3:
		return "\033[33m" // желтый
	case r > -0.7:
		return "\033[35m" // фиолетовый
	default:
		return "\033[31m" // красный
	}
}

func getBiorhythmIconDetailed(val float64) string {
	switch {
	case val > 0.7:
		return "🔴⬆️⬆️"
	case val > 0.3:
		return "🟡⬆️"
	case val > -0.3:
		return "⚪➖"
	case val > -0.7:
		return "🔵⬇️"
	default:
		return "🔻⬇️⬇️"
	}
}

func getBiorhythmStatusText(val float64) string {
	switch {
	case val > 0.7:
		return "🔥 Пик активности, максимальная продуктивность"
	case val > 0.3:
		return "📈 Умеренный подъем, хорошее состояние"
	case val > -0.3:
		return "⚖️ Нейтральная зона, стабильность"
	case val > -0.7:
		return "📉 Умеренный спад, требуется внимание"
	default:
		return "❄️ Глубокий спад, необходим отдых"
	}
}

func getBarVisual(val float64, width int) string {
	absVal := math.Abs(val)
	barLen := int(absVal * float64(width))
	if barLen > width {
		barLen = width
	}
	if val > 0 {
		return "[" + strings.Repeat("▓", barLen) + strings.Repeat("░", width-barLen) + "]"
	} else if val < 0 {
		return "[" + strings.Repeat("░", width-barLen) + strings.Repeat("▓", barLen) + "]"
	}
	return "[" + strings.Repeat("░", width) + "]"
}
