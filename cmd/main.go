package main

import (
	"flag"
	"fmt"
	"math"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"biorhythm-analyzer/internal/logger"
	"biorhythm-analyzer/internal/metrics"
	"biorhythm-analyzer/internal/models"
	"biorhythm-analyzer/internal/storage"
	"biorhythm-analyzer/internal/visualization"
	"biorhythm-analyzer/internal/web"
)

func main() {
	var (
		addName      string
		addBirth     string
		list         bool
		advanced     bool
		webServer    bool
		port         string
		matrix       bool
		timeline     bool
		forecast     bool
		forecastDays int
		help         bool

		logLevel   string
		logFile    string
		logConsole bool
	)

	flag.StringVar(&addName, "name", "", "Имя для добавления")
	flag.StringVar(&addBirth, "birth", "", "Дата рождения (ДД.ММ.ГГГГ)")
	flag.BoolVar(&list, "list", false, "Показать всех людей")
	flag.BoolVar(&advanced, "advanced", false, "Расширенный анализ (6 метрик)")
	flag.BoolVar(&webServer, "web", false, "Запустить веб-интерфейс")
	flag.StringVar(&port, "port", "8095", "Порт для веб-сервера")
	flag.BoolVar(&matrix, "matrix", false, "Показать матрицу корреляций")
	flag.BoolVar(&timeline, "timeline", false, "Динамика корреляций за 90 дней")
	flag.BoolVar(&forecast, "forecast", false, "Прогноз на будущее")
	flag.IntVar(&forecastDays, "days", 30, "Количество дней прогноза")
	flag.BoolVar(&help, "help", false, "Показать справку")

	flag.StringVar(&logLevel, "log-level", "info", "Уровень логирования (debug, info, warn, error, fatal)")
	flag.StringVar(&logFile, "log-file", "", "Файл для записи логов")
	flag.BoolVar(&logConsole, "log-console", true, "Выводить логи в консоль")

	flag.Parse()

	// Инициализация логгера
	logConfig := logger.Config{
		Level:      logLevel,
		LogFile:    logFile,
		Console:    logConsole,
		Timestamp:  true,
		CallerInfo: true,
	}

	log, err := logger.NewLogger(logConfig)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to initialize logger: %v\n", err)
		os.Exit(1)
	}
	defer log.Close()

	log.Info("Starting biorhythm-analyzer")
	log.Debug("Debug mode enabled")

	if help {
		showHelp()
		return
	}

	storage.EnsureDataDir()

	switch {
	case webServer:
		log.Info("Starting web server on port %s", port)
		server := web.NewServer(port, log)

		sigChan := make(chan os.Signal, 1)
		signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

		go func() {
			if err := server.Start(); err != nil {
				log.Error("Server error: %v", err)
				os.Exit(1)
			}
		}()

		<-sigChan
		log.Info("Shutting down...")

	case list:
		runList(log)

	case addName != "" && addBirth != "":
		runAddPerson(addName, addBirth, log)

	case advanced:
		runAdvancedAnalysis(log)

	case matrix:
		runMatrix(log)

	case timeline:
		runTimeline(log)

	case forecast:
		runForecast(forecastDays, log)

	default:
		runBasic(log)
	}
}

func runList(log *logger.Logger) {
	people, err := storage.LoadPeople()
	if err != nil {
		log.Error("Failed to load people: %v", err)
		fmt.Printf("❌ Ошибка загрузки: %v\n", err)
		return
	}

	if len(people) == 0 {
		log.Warn("No people found in database")
		fmt.Println("База пуста. Добавьте людей через --name и --birth")
		return
	}

	log.Info("Listing %d people", len(people))
	fmt.Println("\n=== РЕЕСТР СУБЪЕКТОВ ===\n")
	for i, p := range people {
		fmt.Printf("%2d. %-15s (родился: %s)\n", i+1, p.Name, p.BirthDate.Format("02.01.2006"))
	}
	fmt.Printf("\nВсего: %d человек\n", len(people))
}

func runAddPerson(name, birth string, log *logger.Logger) {
	birthDate, err := time.Parse("02.01.2006", birth)
	if err != nil {
		log.Error("Invalid date format: %s, error: %v", birth, err)
		fmt.Printf("❌ Ошибка формата даты! Используйте ДД.ММ.ГГГГ\n")
		return
	}

	people, err := storage.LoadPeople()
	if err != nil {
		log.Error("Failed to load people: %v", err)
		fmt.Printf("❌ Ошибка загрузки: %v\n", err)
		return
	}

	people = append(people, models.Person{Name: name, BirthDate: birthDate})
	if err := storage.SavePeople(people); err != nil {
		log.Error("Failed to save people: %v", err)
		fmt.Printf("❌ Ошибка сохранения: %v\n", err)
		return
	}

	log.Info("Added person: %s (born: %s)", name, birthDate.Format("02.01.2006"))
	fmt.Printf("✓ %s добавлен в реестр\n", name)
}

func runBasic(log *logger.Logger) {
	now := time.Now()
	people, err := storage.LoadPeople()
	if err != nil {
		log.Error("Failed to load people: %v", err)
		fmt.Printf("❌ Ошибка загрузки: %v\n", err)
		return
	}

	if len(people) < 2 {
		log.Warn("Insufficient people for analysis: %d (need at least 2)", len(people))
		fmt.Println("❌ Нужно минимум 2 человека. Добавьте через --name и --birth")
		return
	}

	log.Info("Running basic analysis for %d people", len(people))

	fmt.Printf("\n%s\n", strings.Repeat("═", 125))
	fmt.Printf("🧠 ПСИХОМЕТРИЧЕСКИЙ АНАЛИЗ")
	fmt.Printf("\n📅 ДАТА РАСЧЁТА: %s", now.Format("02.01.2006"))
	fmt.Printf("  🕐 ВРЕМЯ: %s", now.Format("15:04:05"))
	fmt.Printf("\n%s\n", strings.Repeat("═", 125))

	fmt.Printf("║ %-3s ║ %-20s ║ %-20s ║ %-10s ║ %-10s ║ %-10s ║ %-45s ║\n",
		"№", "Субъект А", "Субъект Б", "r (Пирсон)", "Гармония", "Φ-близость", "Статус (31 зона)")
	fmt.Printf("%s\n", strings.Repeat("═", 125))

	pairNum := 1
	for i := 0; i < len(people); i++ {
		for j := i + 1; j < len(people); j++ {
			r := metrics.CalculateCorrelation(people[i], people[j], now)
			harmony := metrics.CalculateHarmony(r)
			phiDiff := math.Abs(math.Abs(r) - models.Phi)
			status := metrics.GetDetailedStatus(r)

			fmt.Printf("║ %-3d ║ %-20s ║ %-20s ║ %10.4f ║ %9.2f ║ %9.4f ║ %-45s ║\n",
				pairNum, people[i].Name, people[j].Name, r, harmony, phiDiff, status)
			pairNum++
		}
	}
	fmt.Printf("%s\n\n", strings.Repeat("═", 125))

	fmt.Println("📊 ТЕКУЩИЕ ЗНАЧЕНИЯ БИОРИТМОВ")
	fmt.Println(strings.Repeat("─", 80))
	for _, p := range people {
		visualization.PrintBiorhythmStatusWithExplanation(p, now)
	}
}

func runAdvancedAnalysis(log *logger.Logger) {
	now := time.Now()
	people, err := storage.LoadPeople()
	if err != nil {
		log.Error("Failed to load people: %v", err)
		fmt.Printf("❌ Ошибка загрузки: %v\n", err)
		return
	}

	if len(people) < 2 {
		log.Warn("Insufficient people for advanced analysis: %d (need at least 2)", len(people))
		fmt.Println("❌ Нужно минимум 2 человека")
		return
	}

	log.Info("Running advanced analysis for %d people", len(people))

	fmt.Printf("\n%s\n", strings.Repeat("═", 100))
	fmt.Printf("🔬 РАСШИРЕННЫЙ АНАЛИЗ (6 метрик)")
	fmt.Printf("\n📅 ДАТА РАСЧЁТА: %s %s", now.Format("02.01.2006"), now.Format("15:04:05"))
	fmt.Printf("\n%s\n", strings.Repeat("═", 100))

	for i := 0; i < len(people); i++ {
		for j := i + 1; j < len(people); j++ {
			mA := metrics.PrepareBiorhythmMatrix(people[i], now, 60)
			mB := metrics.PrepareBiorhythmMatrix(people[j], now, 60)
			analysis := metrics.AnalyzePair(mA, mB)

			fmt.Printf("\n📌 ПАРА: %s ↔ %s\n", people[i].Name, people[j].Name)
			fmt.Printf("   ┌─────────────────────────────────────────────────────────────────────────────┐\n")
			fmt.Printf("   │ 📊 СТАТИСТИКА ПО 6 МЕТРИКАМ                                                │\n")
			fmt.Printf("   ├─────────────────────────────────────────────────────────────────────────────┤\n")
			fmt.Printf("   │ Пирсон (r):           %8.4f  │ Гармоничность:      %8.2f\n", analysis.Pearson, metrics.CalculateHarmony(analysis.Pearson))
			fmt.Printf("   │ Косинусное сходство:  %8.4f  │ Близость к Φ:       %8.4f\n", analysis.Cosine, math.Abs(math.Abs(analysis.Pearson)-models.Phi))
			fmt.Printf("   │ Спирмен (ρ):          %8.4f  │\n", analysis.Spearman)
			fmt.Printf("   │ Кендалл (τ):          %8.4f  │\n", analysis.Kendall)
			fmt.Printf("   │ Взаимная информация:  %8.4f бит\n", analysis.MutualInfo)
			fmt.Printf("   │ DTW сходство:         %8.4f\n", analysis.DTW)
			fmt.Printf("   ├─────────────────────────────────────────────────────────────────────────────┤\n")
			fmt.Printf("   │ ВЗВЕШЕННЫЙ СКОР:      %8.4f\n", analysis.WeightedScore)
			fmt.Printf("   │ %s\n", analysis.Recommendation)
			fmt.Printf("   └─────────────────────────────────────────────────────────────────────────────┘\n")
		}
	}
}

func runMatrix(log *logger.Logger) {
	now := time.Now()
	people, err := storage.LoadPeople()
	if err != nil {
		log.Error("Failed to load people: %v", err)
		fmt.Printf("❌ Ошибка загрузки: %v\n", err)
		return
	}

	if len(people) == 0 {
		log.Warn("No people for matrix")
		fmt.Println("❌ Нет данных")
		return
	}

	log.Info("Generating correlation matrix for %d people", len(people))
	fmt.Printf("\n📅 ДАТА РАСЧЁТА: %s %s\n", now.Format("02.01.2006"), now.Format("15:04:05"))
	visualization.PrintCorrelationMatrix(people, now)
}

func runTimeline(log *logger.Logger) {
	now := time.Now()
	people, err := storage.LoadPeople()
	if err != nil {
		log.Error("Failed to load people: %v", err)
		fmt.Printf("❌ Ошибка загрузки: %v\n", err)
		return
	}

	if len(people) < 2 {
		log.Warn("Insufficient people for timeline: %d (need at least 2)", len(people))
		fmt.Println("❌ Нужно минимум 2 человека")
		return
	}

	log.Info("Generating timeline for %d people", len(people))

	fmt.Printf("\n%s\n", strings.Repeat("═", 120))
	fmt.Printf("📈 ДИНАМИКА КОРРЕЛЯЦИЙ (90 дней)")
	fmt.Printf("\n📅 ДАТА РАСЧЁТА: %s %s", now.Format("02.01.2006"), now.Format("15:04:05"))
	fmt.Printf("\n%s\n", strings.Repeat("═", 120))

	for i := 0; i < len(people); i++ {
		for j := i + 1; j < len(people); j++ {
			p1, p2 := people[i], people[j]
			startDate := now.AddDate(0, 0, -90)

			mean, median, variance := metrics.CorrelationStability(p1, p2, startDate, 90)
			peakDate, peakR, peakStatus := metrics.FindPeakCorrelation(p1, p2, startDate, 90)
			lowDate, lowR, lowStatus := metrics.FindLowestCorrelation(p1, p2, startDate, 90)

			fmt.Printf("\n📌 %s ↔ %s\n", p1.Name, p2.Name)
			fmt.Printf("   ┌─────────────────────────────────────────────────────────────────────────────┐\n")
			fmt.Printf("   │ 📊 СТАТИСТИКА ЗА 90 ДНЕЙ                                                    │\n")
			fmt.Printf("   ├─────────────────────────────────────────────────────────────────────────────┤\n")
			fmt.Printf("   │ Среднее значение r:     %8.4f                                              │\n", mean)
			fmt.Printf("   │ Медианное значение r:   %8.4f                                              │\n", median)
			fmt.Printf("   │ Дисперсия:               %8.4f                                              │\n", variance)
			fmt.Printf("   │ Гармоничность (сред.):  %8.2f                                              │\n", metrics.CalculateHarmony(mean))
			fmt.Printf("   │                                                                             │\n")
			fmt.Printf("   │ 📈 МАКСИМУМ: %s  | r = %8.4f\n", peakDate.Format("02.01.2006"), peakR)
			fmt.Printf("   │              %s\n", peakStatus)
			fmt.Printf("   │                                                                             │\n")
			fmt.Printf("   │ 📉 МИНИМУМ:  %s  | r = %8.4f\n", lowDate.Format("02.01.2006"), lowR)
			fmt.Printf("   │              %s\n", lowStatus)
			fmt.Printf("   └─────────────────────────────────────────────────────────────────────────────┘\n")

			timeline := metrics.CorrelationTimeline(p1, p2, startDate, 90)
			fmt.Printf("\n   📈 СВЕЧНОЙ ГРАФИК (цвет = знак и сила корреляции):\n   ")
			for day := 0; day < 90 && day < len(timeline); day++ {
				r := timeline[day].R
				if r > 0.6 {
					fmt.Print("🟩")
				} else if r > 0.2 {
					fmt.Print("🟢")
				} else if r > -0.2 {
					fmt.Print("🟡")
				} else if r > -0.6 {
					fmt.Print("🟠")
				} else {
					fmt.Print("🔴")
				}
				if (day+1)%30 == 0 {
					fmt.Printf(" ")
				}
			}
			fmt.Println(" → 90 дней")

			fmt.Printf("\n   📊 СПАРКЛАЙН (высота = сила связи |r|):\n   ")
			for day := 0; day < 90 && day < len(timeline); day++ {
				absR := math.Abs(timeline[day].R)
				if absR > 0.8 {
					fmt.Print("▇")
				} else if absR > 0.6 {
					fmt.Print("▆")
				} else if absR > 0.4 {
					fmt.Print("▅")
				} else if absR > 0.2 {
					fmt.Print("▃")
				} else {
					fmt.Print("▁")
				}
				if (day+1)%30 == 0 {
					fmt.Printf(" ")
				}
			}
			fmt.Println(" → 90 дней")

			fmt.Printf("\n   🎵 ГАРМОНИЧНОСТЬ (близость к Φ = 0.618):\n   ")
			for day := 0; day < 90 && day < len(timeline); day++ {
				harmony := metrics.CalculateHarmony(timeline[day].R)
				if harmony > 8 {
					fmt.Print("💎")
				} else if harmony > 6 {
					fmt.Print("✨")
				} else if harmony > 4 {
					fmt.Print("🌱")
				} else if harmony > 2 {
					fmt.Print("🍂")
				} else {
					fmt.Print("💔")
				}
				if (day+1)%15 == 0 {
					fmt.Printf(" ")
				}
			}
			fmt.Println(" → 90 дней")
			fmt.Println()
		}
	}
}

func runForecast(days int, log *logger.Logger) {
	now := time.Now()
	people, err := storage.LoadPeople()
	if err != nil {
		log.Error("Failed to load people: %v", err)
		fmt.Printf("❌ Ошибка загрузки: %v\n", err)
		return
	}

	if len(people) < 2 {
		log.Warn("Insufficient people for forecast: %d (need at least 2)", len(people))
		fmt.Println("❌ Нужно минимум 2 человека")
		return
	}

	log.Info("Generating %d-day forecast for %d people", days, len(people))

	fmt.Printf("\n%s\n", strings.Repeat("═", 120))
	fmt.Printf("🔮 ПРОГНОЗ НА БУДУЩИЕ %d ДНЕЙ", days)
	fmt.Printf("\n📅 НАЧАЛО ПРОГНОЗА: %s", now.Format("02.01.2006"))
	fmt.Printf("\n%s\n", strings.Repeat("═", 120))

	for i := 0; i < len(people); i++ {
		for j := i + 1; j < len(people); j++ {
			p1, p2 := people[i], people[j]
			startDate := now.AddDate(0, 0, 1)

			fmt.Printf("\n📌 ПРОГНОЗ: %s ↔ %s\n", p1.Name, p2.Name)
			fmt.Printf("   ┌─────────────────────────────────────────────────────────────────────────────┐\n")

			forecast := metrics.ForecastCorrelation(p1, p2, startDate, days)

			bestDay := forecast[0]
			worstDay := forecast[0]
			for _, f := range forecast {
				if f.R > bestDay.R {
					bestDay = f
				}
				if f.R < worstDay.R {
					worstDay = f
				}
			}

			fmt.Printf("   │ 🌟 ЛУЧШИЙ ДЕНЬ:     %s  | r = %8.4f  | Гармония: %5.1f\n",
				bestDay.Date.Format("02.01.2006"), bestDay.R, bestDay.Harmony)
			fmt.Printf("   │                   %s\n", bestDay.Status)
			fmt.Printf("   │                                                                             │\n")
			fmt.Printf("   │ ⚠️ ХУДШИЙ ДЕНЬ:     %s  | r = %8.4f  | Гармония: %5.1f\n",
				worstDay.Date.Format("02.01.2006"), worstDay.R, worstDay.Harmony)
			fmt.Printf("   │                   %s\n", worstDay.Status)
			fmt.Printf("   └─────────────────────────────────────────────────────────────────────────────┘\n")

			fmt.Printf("\n   📊 ПРОГНОЗ ПО СФЕРАМ ЖИЗНИ (следующие 7 дней):\n")
			fmt.Printf("   ┌─────────────────────────────────────────────────────────────────────────────┐\n")

			for day := 0; day < 7 && day < len(forecast); day++ {
				f := forecast[day]
				fmt.Printf("   │ 📅 %s (r = %+.4f, гармония: %.1f):\n",
					f.Date.Format("02.01.2006"), f.R, f.Harmony)
				for sphere, score := range f.SphereScores {
					bar := getScoreBar(score, 25)
					fmt.Printf("   │    %-22s: %5.1f%% %s\n", sphere, score, bar)
				}
				if day < 6 && day+1 < len(forecast) {
					fmt.Printf("   │\n")
				}
			}
			fmt.Printf("   └─────────────────────────────────────────────────────────────────────────────┘\n")

			fmt.Printf("\n   💡 РЕКОМЕНДАЦИИ ПО СФЕРАМ (лучшие дни в ближайшие %d дней):\n", days)
			spheres := []string{"💖 Любовь и отношения", "💼 Карьера и бизнес", "🎨 Творчество", "📚 Обучение и развитие", "💰 Финансы"}
			for _, sphere := range spheres {
				bestDays := metrics.FindBestDaysForSphere(p1, p2, startDate, days, sphere)
				if len(bestDays) > 0 {
					fmt.Printf("   • %s: ", sphere)
					for idx, d := range bestDays[:min(3, len(bestDays))] {
						if idx > 0 {
							fmt.Printf(", ")
						}
						fmt.Printf("%s (%.0f%%)", d.Date.Format("02.01"), d.Score)
					}
					fmt.Println()
				}
			}
			fmt.Println()
		}
	}
}

func getScoreBar(score float64, width int) string {
	barLen := int(score / 100 * float64(width))
	if barLen > width {
		barLen = width
	}
	if barLen < 0 {
		barLen = 0
	}
	return "[" + strings.Repeat("█", barLen) + strings.Repeat("░", width-barLen) + "]"
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func showHelp() {
	fmt.Println(`
╔═══════════════════════════════════════════════════════════════════════════════╗
║                    🧠 ПСИХОМЕТРИЧЕСКИЙ АНАЛИЗАТОР                             ║
║                         Версия 2.0 — 31 зона, прогнозы, гармоничность         ║
╚═══════════════════════════════════════════════════════════════════════════════╝

📖 ИСПОЛЬЗОВАНИЕ:
  biorhythm-analyzer [OPTIONS]

🔧 ОПЦИИ:
  --name="Имя"           Добавить человека
  --birth="ДД.ММ.ГГГГ"  Дата рождения
  --list                 Показать всех людей
  --matrix               Матрица корреляций
  --timeline             Динамика за 90 дней (свечной график, спарклайн, гармония)
  --forecast             Прогноз на будущее по 7 сферам жизни
  --days=N               Количество дней прогноза (по умолчанию 30)
  --advanced             Расширенный анализ (6 метрик)
  --web                  Веб-интерфейс
  --port=8095            Порт для веб-сервера (по умолчанию 8095)
  --help                 Справка

📊 НОВЫЕ МЕТРИКИ:
  • Гармоничность (0-10) — близость к золотому сечению (Φ = 0.618)
  • 31 зона классификации — 7 уровней психологических состояний
  • Прогноз по 7 сферам: любовь, карьера, дружба, творчество, здоровье, обучение, финансы

📈 ДИНАМИКА --timeline:
  • Среднее, медиана, дисперсия
  • Свечной график (🟩🟢🟡🟠🔴)
  • Спарклайн (▇▆▅▃▁) — высота = сила связи
  • Гармоничность (💎✨🌱🍂💔)

🔮 ПРОГНОЗ --forecast:
  • Лучшие и худшие дни
  • Оценки по 7 сферам жизни на каждый день
  • Рекомендации по лучшим дням для каждой сферы

📝 ЛОГИРОВАНИЕ:
  --log-level=info       Уровень логирования (debug, info, warn, error, fatal)
  --log-file=./logs/app.log  Файл для записи логов
  --log-console=true      Выводить логи в консоль
`)
}
