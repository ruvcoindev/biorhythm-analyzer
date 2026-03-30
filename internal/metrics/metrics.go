package metrics

import (
	"fmt"
	"math"
	"sort"
	"strings"
	"time"

	"biorhythm-analyzer/internal/models"
)

// ==================== БАЗОВЫЕ ФУНКЦИИ БИОРИТМОВ ====================

// GetBiorhythm вычисляет значение биоритма по периоду

// CalculateCorrelation вычисляет корреляцию Пирсона между двумя людьми
func CalculateCorrelation(p1, p2 models.Person, now time.Time) float64 {
	n := float64(len(models.Biorhythms))
	var sumX, sumY, sumXY, sumX2, sumY2 float64

	for _, br := range models.Biorhythms {
		valX := GetBiorhythm(p1.BirthDate, now, br.Period)
		valY := GetBiorhythm(p2.BirthDate, now, br.Period)
		sumX += valX
		sumY += valY
		sumXY += valX * valY
		sumX2 += valX * valX
		sumY2 += valY * valY
	}

	num := n*sumXY - sumX*sumY
	den := math.Sqrt((n*sumX2 - sumX*sumX) * (n*sumY2 - sumY*sumY))
	if den == 0 {
		return 0
	}
	r := num / den
	if r > 1 {
		return 1
	}
	if r < -1 {
		return -1
	}
	return r
}

// ==================== ПСИХОЛОГИЧЕСКАЯ КЛАССИФИКАЦИЯ (31 зона) ====================

// GetDetailedStatus - 31 зона психологической аналитики
// 7 основных уровней (по аналогии с уровнями сознания)
func GetDetailedStatus(r float64) string {
	p := (r + 1) / 2 * 100
	absR := math.Abs(r)
	diff := math.Abs(absR - models.Phi)
	golden := ""
	if diff < 0.05 {
		golden = " [✨ ЗОЛОТОЕ СЕЧЕНИЕ]"
	}
	
	// ==================== УРОВЕНЬ 7: САМОАКТУАЛИЗАЦИЯ (r > 0.95) ====================
	if r > 0.98 {
		return fmt.Sprintf("%5.1f%% | 🔥 СИМБИОЗ (Слияние душ, потеря эго)%s", p, golden)
	}
	if r > 0.96 {
		return fmt.Sprintf("%5.1f%% | 💫 ТРАНСЦЕНДЕНТНОСТЬ (Выход за пределы личности)%s", p, golden)
	}
	if r > 0.95 {
		return fmt.Sprintf("%5.1f%% | 🕊️ САМОАКТУАЛИЗАЦИЯ (Полная реализация потенциала)%s", p, golden)
	}
	
	// ==================== УРОВЕНЬ 6: ГАРМОНИЯ (0.80 - 0.95) ====================
	if r > 0.90 {
		return fmt.Sprintf("%5.1f%% | 💎 КОСМИЧЕСКАЯ ЛЮБОВЬ (Безусловное принятие)%s", p, golden)
	}
	if r > 0.85 {
		return fmt.Sprintf("%5.1f%% | 💖 БОЖЕСТВЕННАЯ ГАРМОНИЯ (Идеальный резонанс)%s", p, golden)
	}
	if r > 0.80 {
		return fmt.Sprintf("%5.1f%% | 💞 ГЛУБОКАЯ ПРИВЯЗАННОСТЬ (Душевное родство)%s", p, golden)
	}
	
	// ==================== УРОВЕНЬ 5: ЛЮБОВЬ И ПРИНЯТИЕ (0.60 - 0.80) ====================
	if r > 0.75 {
		return fmt.Sprintf("%5.1f%% | 💛 ИСКРЕННЯЯ БЛИЗОСТЬ (Тёплые доверительные)%s", p, golden)
	}
	if r > 0.70 {
		return fmt.Sprintf("%5.1f%% | 🌸 ДРУЖЕСКАЯ СИМПАТИЯ (Естественное притяжение)%s", p, golden)
	}
	if r > 0.65 {
		return fmt.Sprintf("%5.1f%% | 🤝 ВЗАИМОПОНИМАНИЕ (Согласованность ценностей)%s", p, golden)
	}
	if r > 0.60 {
		return fmt.Sprintf("%5.1f%% | 🌱 СИМПАТИЯ (Начало близости)%s", p, golden)
	}
	
	// ==================== УРОВЕНЬ 4: СТАБИЛЬНОСТЬ (0.30 - 0.60) ====================
	if r > 0.50 {
		return fmt.Sprintf("%5.1f%% | 👋 ДОБРОЖЕЛАТЕЛЬНОСТЬ (Открытость к контакту)%s", p, golden)
	}
	if r > 0.45 {
		return fmt.Sprintf("%5.1f%% | 😐 НЕЙТРАЛЬНО-ПОЗИТИВНОЕ (Комфортное сосуществование)%s", p, golden)
	}
	if r > 0.40 {
		return fmt.Sprintf("%5.1f%% | 📍 ЛЁГКАЯ СИМПАТИЯ (Без обязательств)%s", p, golden)
	}
	if r > 0.30 {
		return fmt.Sprintf("%5.1f%% | 🧘 ЭМОЦИОНАЛЬНЫЙ НОЛЬ (Спокойное равнодушие)%s", p, golden)
	}
	
	// ==================== УРОВЕНЬ 3: НЕЙТРАЛЬНОСТЬ (0.00 - 0.30) ====================
	if r > 0.20 {
		return fmt.Sprintf("%5.1f%% | 🧊 ЛЁГКАЯ ОТСТРАНЁННОСТЬ (Дипломатичность)%s", p, golden)
	}
	if r > 0.10 {
		return fmt.Sprintf("%5.1f%% | 🤨 НАБЛЮДАТЕЛЬ (Сторонний анализ)%s", p, golden)
	}
	if r > 0.00 {
		return fmt.Sprintf("%5.1f%% | ❓ НЕОПРЕДЕЛЁННОСТЬ (Формирование отношения)%s", p, golden)
	}
	
	// ==================== УРОВЕНЬ 2: НАПРЯЖЕНИЕ (-0.30 - 0.00) ====================
	if r > -0.10 {
		return fmt.Sprintf("%5.1f%% | 😌 ЛЁГКОЕ НАПРЯЖЕНИЕ (Притирка)%s", p, golden)
	}
	if r > -0.20 {
		return fmt.Sprintf("%5.1f%% | 😤 РАЗДРАЖЕНИЕ (Мелкие конфликты)%s", p, golden)
	}
	if r > -0.30 {
		return fmt.Sprintf("%5.1f%% | 🥀 ОТЧУЖДЕНИЕ (Эмоциональная дистанция)%s", p, golden)
	}
	
	// ==================== УРОВЕНЬ 1: КОНФЛИКТ (-0.60 - -0.30) ====================
	if r > -0.40 {
		return fmt.Sprintf("%5.1f%% | ⚡ НАПРЯЖЕНИЕ (Постоянные трения)%s", p, golden)
	}
	if r > -0.50 {
		return fmt.Sprintf("%5.1f%% | 🔥 КОНФРОНТАЦИЯ (Открытое противостояние)%s", p, golden)
	}
	if r > -0.60 {
		return fmt.Sprintf("%5.1f%% | 💔 РАЗРЫВ (Потеря эмоциональной связи)%s", p, golden)
	}
	
	// ==================== УРОВЕНЬ 0: АНТАГОНИЗМ (r < -0.60) ====================
	if r > -0.70 {
		return fmt.Sprintf("%5.1f%% | 🗡️ ВРАЖДЕБНОСТЬ (Системный конфликт)%s", p, golden)
	}
	if r > -0.80 {
		return fmt.Sprintf("%5.1f%% | 💀 АНТАГОНИЗМ (Непримиримое противостояние)%s", p, golden)
	}
	if r > -0.90 {
		return fmt.Sprintf("%5.1f%% | 🌑 ПСИХОЛОГИЧЕСКОЕ ОТТОРЖЕНИЕ (Аверсия)%s", p, golden)
	}
	if r > -0.95 {
		return fmt.Sprintf("%5.1f%% | 🕳️ ЭКЗИСТЕНЦИАЛЬНАЯ НЕСОВМЕСТИМОСТЬ (Полное неприятие)%s", p, golden)
	}
	return fmt.Sprintf("%5.1f%% | 🌌 ТОТАЛЬНЫЙ РАЗРЫВ (Энергетический вакуум)%s", p, golden)
}

// ==================== ГАРМОНИЧНОСТЬ (вместо ликвидности) ====================

// CalculateHarmony вычисляет гармоничность отношений по близости к золотому сечению
// Значение от 0 до 10, где 10 — идеальная гармония (|r| = Φ)
func CalculateHarmony(r float64) float64 {
	diff := math.Abs(math.Abs(r) - models.Phi)
	// Максимальная гармония = 10 при diff = 0
	// Минимальная = 0 при diff = 1
	harmony := 10.0 * (1.0 - math.Min(1.0, diff))
	return math.Round(harmony*100) / 100
}

// GetHarmonyDescription возвращает описание уровня гармонии
func GetHarmonyDescription(harmony float64) string {
	switch {
	case harmony >= 9.0:
		return "🔥 ИДЕАЛЬНАЯ ГАРМОНИЯ — золотое сечение, высший резонанс"
	case harmony >= 7.0:
		return "💫 ВЫСОКАЯ ГАРМОНИЯ — естественная синхронизация"
	case harmony >= 5.0:
		return "✨ УМЕРЕННАЯ ГАРМОНИЯ — баланс, комфортное взаимодействие"
	case harmony >= 3.0:
		return "🌱 СЛАБАЯ ГАРМОНИЯ — потенциал для развития"
	default:
		return "⚖️ ДИСГАРМОНИЯ — требуется работа над отношениями"
	}
}

// ==================== КОСИНУСНОЕ СХОДСТВО ====================

func CosineSimilarity(vecA, vecB []float64) float64 {
	if len(vecA) != len(vecB) || len(vecA) == 0 {
		return 0
	}
	var dot, normA, normB float64
	for i := 0; i < len(vecA); i++ {
		dot += vecA[i] * vecB[i]
		normA += vecA[i] * vecA[i]
		normB += vecB[i] * vecB[i]
	}
	if normA == 0 || normB == 0 {
		return 0
	}
	return dot / (math.Sqrt(normA) * math.Sqrt(normB))
}

// ==================== РАНГОВЫЕ КОРРЕЛЯЦИИ ====================

func SpearmanCorrelation(vecA, vecB []float64) float64 {
	if len(vecA) != len(vecB) {
		return 0
	}
	ranksA := getRanks(vecA)
	ranksB := getRanks(vecB)
	return pearsonCorrelationSimple(ranksA, ranksB)
}

func KendallTau(vecA, vecB []float64) float64 {
	if len(vecA) != len(vecB) {
		return 0
	}
	n := len(vecA)
	concordant, discordant := 0, 0
	for i := 0; i < n-1; i++ {
		for j := i + 1; j < n; j++ {
			signA := vecA[i] - vecA[j]
			signB := vecB[i] - vecB[j]
			if signA*signB > 0 {
				concordant++
			} else if signA*signB < 0 {
				discordant++
			}
		}
	}
	total := float64(concordant + discordant)
	if total == 0 {
		return 0
	}
	return float64(concordant-discordant) / total
}

func getRanks(values []float64) []float64 {
	type pair struct {
		val   float64
		index int
	}
	pairs := make([]pair, len(values))
	for i, v := range values {
		pairs[i] = pair{v, i}
	}
	sort.Slice(pairs, func(i, j int) bool {
		return pairs[i].val < pairs[j].val
	})
	ranks := make([]float64, len(values))
	for i, p := range pairs {
		ranks[p.index] = float64(i + 1)
	}
	return ranks
}

func pearsonCorrelationSimple(vecA, vecB []float64) float64 {
	n := float64(len(vecA))
	var sumA, sumB, sumAB, sumA2, sumB2 float64
	for i := 0; i < len(vecA); i++ {
		sumA += vecA[i]
		sumB += vecB[i]
		sumAB += vecA[i] * vecB[i]
		sumA2 += vecA[i] * vecA[i]
		sumB2 += vecB[i] * vecB[i]
	}
	num := n*sumAB - sumA*sumB
	den := math.Sqrt((n*sumA2 - sumA*sumA) * (n*sumB2 - sumB*sumB))
	if den == 0 {
		return 0
	}
	return num / den
}

// ==================== ВЗАИМНАЯ ИНФОРМАЦИЯ ====================

func MutualInformation(vecA, vecB []float64, bins int) float64 {
	if len(vecA) != len(vecB) || len(vecA) == 0 {
		return 0
	}
	discreteA := discretize(vecA, bins)
	discreteB := discretize(vecB, bins)
	joint := make([][]int, bins)
	for i := range joint {
		joint[i] = make([]int, bins)
	}
	margA := make([]int, bins)
	margB := make([]int, bins)
	n := len(discreteA)
	for i := 0; i < n; i++ {
		a := discreteA[i]
		b := discreteB[i]
		joint[a][b]++
		margA[a]++
		margB[b]++
	}
	mi := 0.0
	for i := 0; i < bins; i++ {
		for j := 0; j < bins; j++ {
			if joint[i][j] > 0 {
				pXY := float64(joint[i][j]) / float64(n)
				pX := float64(margA[i]) / float64(n)
				pY := float64(margB[j]) / float64(n)
				mi += pXY * math.Log(pXY/(pX*pY))
			}
		}
	}
	return mi / math.Log(2)
}

func discretize(values []float64, bins int) []int {
	if len(values) == 0 {
		return []int{}
	}
	sorted := make([]float64, len(values))
	copy(sorted, values)
	sort.Float64s(sorted)
	boundaries := make([]float64, bins-1)
	for i := 0; i < bins-1; i++ {
		quantile := float64(i+1) / float64(bins)
		idx := int(quantile * float64(len(sorted)-1))
		boundaries[i] = sorted[idx]
	}
	result := make([]int, len(values))
	for i, v := range values {
		bin := 0
		for _, boundary := range boundaries {
			if v > boundary {
				bin++
			}
		}
		result[i] = bin
	}
	return result
}

// ==================== DTW (DYNAMIC TIME WARPING) ====================

func DTWDistance(seriesA, seriesB []float64) float64 {
	n, m := len(seriesA), len(seriesB)
	dtw := make([][]float64, n+1)
	for i := range dtw {
		dtw[i] = make([]float64, m+1)
		for j := range dtw[i] {
			dtw[i][j] = math.Inf(1)
		}
	}
	dtw[0][0] = 0
	for i := 1; i <= n; i++ {
		for j := 1; j <= m; j++ {
			cost := math.Abs(seriesA[i-1] - seriesB[j-1])
			dtw[i][j] = cost + math.Min(
				math.Min(dtw[i-1][j], dtw[i][j-1]),
				dtw[i-1][j-1],
			)
		}
	}
	return dtw[n][m]
}

func DTWSimilarity(seriesA, seriesB []float64) float64 {
	distance := DTWDistance(seriesA, seriesB)
	maxDist := 0.0
	for i := 0; i < len(seriesA); i++ {
		for j := 0; j < len(seriesB); j++ {
			if d := math.Abs(seriesA[i] - seriesB[j]); d > maxDist {
				maxDist = d
			}
		}
	}
	if maxDist == 0 {
		return 1.0
	}
	maxLen := len(seriesA)
	if len(seriesB) > maxLen {
		maxLen = len(seriesB)
	}
	normalized := distance / (maxDist * float64(maxLen))
	if normalized > 1 {
		normalized = 1
	}
	return 1.0 - normalized
}

// ==================== ПОДГОТОВКА МАТРИЦ ====================

func PrepareBiorhythmMatrix(person models.Person, now time.Time, days int) [][]float64 {
	matrix := make([][]float64, len(models.Biorhythms))
	for i, br := range models.Biorhythms {
		matrix[i] = make([]float64, days)
		for d := 0; d < days; d++ {
			date := now.AddDate(0, 0, d-days/2)
			matrix[i][d] = GetBiorhythm(person.BirthDate, date, br.Period)
		}
	}
	return matrix
}

func flattenMatrix(matrix [][]float64) []float64 {
	if len(matrix) == 0 {
		return []float64{}
	}
	flat := make([]float64, 0, len(matrix)*len(matrix[0]))
	for _, row := range matrix {
		flat = append(flat, row...)
	}
	return flat
}

// ==================== КОМПЛЕКСНЫЙ АНАЛИЗ ====================

func AnalyzePair(matrixA, matrixB [][]float64) models.ComprehensiveAnalysis {
	result := models.ComprehensiveAnalysis{}
	vectorA := flattenMatrix(matrixA)
	vectorB := flattenMatrix(matrixB)

	result.Pearson = pearsonCorrelationSimple(vectorA, vectorB)
	result.Cosine = CosineSimilarity(vectorA, vectorB)
	result.Spearman = SpearmanCorrelation(vectorA, vectorB)
	result.Kendall = KendallTau(vectorA, vectorB)
	result.MutualInfo = MutualInformation(vectorA, vectorB, 10)
	result.DTW = DTWSimilarity(vectorA, vectorB)
	result.SpectralMatch = (result.Cosine + result.Pearson) / 2

	weights := map[string]float64{
		"pearson":     0.25,
		"cosine":      0.25,
		"spearman":    0.15,
		"mutual_info": 0.20,
		"dtw":         0.15,
	}

	result.WeightedScore =
		weights["pearson"]*(result.Pearson+1)/2 +
			weights["cosine"]*(result.Cosine+1)/2 +
			weights["spearman"]*(result.Spearman+1)/2 +
			weights["mutual_info"]*math.Min(1.0, result.MutualInfo/2.0) +
			weights["dtw"]*result.DTW

	if result.WeightedScore > 0.75 {
		result.Recommendation = "⭐ ВЫСОКАЯ СОВМЕСТИМОСТЬ — глубокое взаимопонимание"
	} else if result.WeightedScore > 0.5 {
		result.Recommendation = "🌱 ПОТЕНЦИАЛ РАЗВИТИЯ — есть основа для гармонизации"
	} else if result.WeightedScore > 0.25 {
		result.Recommendation = "⚖️ НЕЙТРАЛЬНО — требуется осознанная работа над отношениями"
	} else {
		result.Recommendation = "❄️ НИЗКАЯ СОВМЕСТИМОСТЬ — значительные различия в паттернах"
	}

	return result
}

func FormatAnalysisOutput(analysis models.ComprehensiveAnalysis) string {
	var s strings.Builder
	s.WriteString("\n" + strings.Repeat("═", 80) + "\n")
	s.WriteString("📊 КОМПЛЕКСНЫЙ АНАЛИЗ ВЗАИМОДЕЙСТВИЯ\n")
	s.WriteString(strings.Repeat("═", 80) + "\n")
	s.WriteString(fmt.Sprintf("\nПирсон (r):           %.4f\n", analysis.Pearson))
	s.WriteString(fmt.Sprintf("Косинусное сходство:  %.4f\n", analysis.Cosine))
	s.WriteString(fmt.Sprintf("Спирмен (ρ):          %.4f\n", analysis.Spearman))
	s.WriteString(fmt.Sprintf("Кендалл (τ):          %.4f\n", analysis.Kendall))
	s.WriteString(fmt.Sprintf("Взаимная информация:  %.4f бит\n", analysis.MutualInfo))
	s.WriteString(fmt.Sprintf("DTW сходство:         %.4f\n", analysis.DTW))
	s.WriteString(strings.Repeat("─", 80) + "\n")
	s.WriteString(fmt.Sprintf("ВЗВЕШЕННЫЙ СКОР:      %.4f\n", analysis.WeightedScore))
	s.WriteString("\n💡 РЕКОМЕНДАЦИЯ: " + analysis.Recommendation + "\n")
	s.WriteString(strings.Repeat("═", 80) + "\n")
	return s.String()
}

// ==================== АНАЛИЗ ВО ВРЕМЕНИ ====================

// TimelinePoint структура точки временного ряда
type TimelinePoint struct {
	Date   time.Time
	R      float64
	Status string
}

// CorrelationTimeline возвращает массив корреляций за период дней
func CorrelationTimeline(p1, p2 models.Person, startDate time.Time, days int) []TimelinePoint {
	result := make([]TimelinePoint, days)
	for i := 0; i < days; i++ {
		date := startDate.AddDate(0, 0, i)
		r := CalculateCorrelation(p1, p2, date)
		result[i] = TimelinePoint{date, r, GetDetailedStatus(r)}
	}
	return result
}

// FindPeakCorrelation находит максимальную корреляцию за период
func FindPeakCorrelation(p1, p2 models.Person, startDate time.Time, days int) (time.Time, float64, string) {
	maxR := -2.0
	var maxDate time.Time
	var maxStatus string
	for i := 0; i < days; i++ {
		date := startDate.AddDate(0, 0, i)
		r := CalculateCorrelation(p1, p2, date)
		if r > maxR {
			maxR = r
			maxDate = date
			maxStatus = GetDetailedStatus(r)
		}
	}
	return maxDate, maxR, maxStatus
}

// FindLowestCorrelation находит минимальную корреляцию за период
func FindLowestCorrelation(p1, p2 models.Person, startDate time.Time, days int) (time.Time, float64, string) {
	minR := 2.0
	var minDate time.Time
	var minStatus string
	for i := 0; i < days; i++ {
		date := startDate.AddDate(0, 0, i)
		r := CalculateCorrelation(p1, p2, date)
		if r < minR {
			minR = r
			minDate = date
			minStatus = GetDetailedStatus(r)
		}
	}
	return minDate, minR, minStatus
}

// CorrelationStability вычисляет стабильность корреляции (среднее, медиана, дисперсия)
func CorrelationStability(p1, p2 models.Person, startDate time.Time, days int) (mean, median, variance float64) {
	values := make([]float64, days)
	sum := 0.0
	for i := 0; i < days; i++ {
		date := startDate.AddDate(0, 0, i)
		r := CalculateCorrelation(p1, p2, date)
		values[i] = r
		sum += r
	}
	mean = sum / float64(days)
	
	// Медиана
	sorted := make([]float64, days)
	copy(sorted, values)
	sort.Float64s(sorted)
	if days%2 == 0 {
		median = (sorted[days/2-1] + sorted[days/2]) / 2
	} else {
		median = sorted[days/2]
	}
	
	// Дисперсия
	sumSq := 0.0
	for _, v := range values {
		diff := v - mean
		sumSq += diff * diff
	}
	variance = sumSq / float64(days)
	
	return mean, median, variance
}

// ==================== ПРОГНОЗ НА БУДУЩЕЕ ====================

// SphereScores оценки по сферам жизни
type SphereScores map[string]float64

// ForecastPoint точка прогноза
type ForecastPoint struct {
	Date         time.Time
	R            float64
	Status       string
	Harmony      float64
	SphereScores SphereScores
}

// SpheresOfLife сферы жизни для прогноза
var SpheresOfLife = []string{
	"💖 Любовь и отношения",
	"💼 Карьера и бизнес",
	"👥 Дружба и социализация",
	"🎨 Творчество",
	"💪 Здоровье и энергия",
	"📚 Обучение и развитие",
	"💰 Финансы",
}

// ForecastCorrelation прогноз корреляции на будущие даты
func ForecastCorrelation(p1, p2 models.Person, startDate time.Time, days int) []ForecastPoint {
	result := make([]ForecastPoint, days)
	for i := 0; i < days; i++ {
		date := startDate.AddDate(0, 0, i)
		r := CalculateCorrelation(p1, p2, date)
		harmony := CalculateHarmony(r)
		sphereScores := calculateSphereScores(r)
		
		result[i] = ForecastPoint{
			Date:         date,
			R:            r,
			Status:       GetDetailedStatus(r),
			Harmony:      harmony,
			SphereScores: sphereScores,
		}
	}
	return result
}

// calculateSphereScores рассчитывает оценки по сферам жизни на основе корреляции
func calculateSphereScores(r float64) SphereScores {
	scores := make(SphereScores)
	absR := math.Abs(r)
	baseScore := (r + 1) / 2 * 100
	
	for _, sphere := range SpheresOfLife {
		var score float64
		switch sphere {
		case "💖 Любовь и отношения":
			score = baseScore * (0.7 + 0.3*absR)
		case "💼 Карьера и бизнес":
			careerMod := 1.0 - math.Abs(absR-0.5)*0.5
			score = baseScore * careerMod
		case "👥 Дружба и социализация":
			score = baseScore * (0.8 + 0.2*math.Max(0, r))
		case "🎨 Творчество":
			creativeMod := math.Max(0, r) * 0.5
			score = baseScore * (0.5 + creativeMod)
		case "💪 Здоровье и энергия":
			healthMod := 1.0 - absR*0.3
			score = baseScore * healthMod
		case "📚 Обучение и развитие":
			learnMod := 1.0 - math.Abs(absR-0.3)*0.4
			score = baseScore * learnMod
		case "💰 Финансы":
			financeMod := 0.6 + math.Max(0, r)*0.4
			score = baseScore * financeMod
		default:
			score = baseScore
		}
		if score > 100 {
			score = 100
		}
		if score < 0 {
			score = 0
		}
		scores[sphere] = math.Round(score*10) / 10
	}
	return scores
}

// FindBestDaysForSphere находит лучшие дни для конкретной сферы жизни
func FindBestDaysForSphere(p1, p2 models.Person, startDate time.Time, days int, sphere string) []struct {
	Date  time.Time
	Score float64
	R     float64
} {
	forecast := ForecastCorrelation(p1, p2, startDate, days)
	
	type DayScore struct {
		Date  time.Time
		Score float64
		R     float64
	}
	
	scores := make([]DayScore, 0)
	for _, f := range forecast {
		if score, ok := f.SphereScores[sphere]; ok {
			scores = append(scores, DayScore{f.Date, score, f.R})
		}
	}
	
	// Сортируем по убыванию оценки
	for i := 0; i < len(scores)-1; i++ {
		for j := i + 1; j < len(scores); j++ {
			if scores[i].Score < scores[j].Score {
				scores[i], scores[j] = scores[j], scores[i]
			}
		}
	}
	
	// Возвращаем топ-5
	result := make([]struct {
		Date  time.Time
		Score float64
		R     float64
	}, 0)
	for i := 0; i < len(scores) && i < 5; i++ {
		result = append(result, struct {
			Date  time.Time
			Score float64
			R     float64
		}{scores[i].Date, scores[i].Score, scores[i].R})
	}
	return result
}
func GetBiorhythm(birthDate time.Time, now time.Time, period float64) float64 {
	// Защита от паники
	defer func() {
		if r := recover(); r != nil {
			fmt.Printf("PANIC in GetBiorhythm: birth=%s, now=%s, period=%f, error=%v\n", 
				birthDate.Format("2006-01-02"), now.Format("2006-01-02"), period, r)
		}
	}()
	
	days := now.Sub(birthDate).Hours() / 24
	// Защита от слишком больших чисел
	if math.IsNaN(days) || math.IsInf(days, 0) {
		return 0
	}
	val := math.Sin(2 * math.Pi * days / period)
	if math.IsNaN(val) || math.IsInf(val, 0) {
		return 0
	}
	return val
}
