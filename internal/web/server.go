package web

import (
	"encoding/json"
	"fmt"
	"math"
	"net/http"
	"sort"
	"time"

	"biorhythm-analyzer/internal/logger"
	"biorhythm-analyzer/internal/metrics"
	"biorhythm-analyzer/internal/models"
	"biorhythm-analyzer/internal/storage"
)

type Server struct {
	port string
	log  *logger.Logger
}

func NewServer(port string, log *logger.Logger) *Server {
	return &Server{
		port: port,
		log:  log,
	}
}

func (s *Server) Start() error {
	http.HandleFunc("/", s.logMiddleware(s.indexHandler))
	http.HandleFunc("/matrix", s.logMiddleware(s.matrixPageHandler))
	http.HandleFunc("/biorhythms", s.logMiddleware(s.biorhythmsPageHandler))
	http.HandleFunc("/timeline", s.logMiddleware(s.timelinePageHandler))
	http.HandleFunc("/forecast", s.logMiddleware(s.forecastPageHandler))
	http.HandleFunc("/help", s.logMiddleware(s.helpPageHandler))
	http.HandleFunc("/zones", s.zonesPageHandler)
	http.HandleFunc("/api/pairs", s.logMiddleware(s.pairsHandler))
	http.HandleFunc("/api/matrix", s.logMiddleware(s.matrixHandler))
	http.HandleFunc("/api/biorhythms", s.logMiddleware(s.biorhythmsAllHandler))
	http.HandleFunc("/api/timeline", s.logMiddleware(s.timelineDataHandler))
	http.HandleFunc("/api/forecast", s.logMiddleware(s.forecastDataHandler))

	s.log.Info("Web server starting on port %s", s.port)
	fmt.Printf("\n🌐 Веб-интерфейс: http://localhost:%s\n", s.port)
	fmt.Printf("📅 Текущее время сервера: %s\n", time.Now().Format("02.01.2006 15:04:05"))
	fmt.Println("\n📊 Доступные страницы:")
	fmt.Printf("   - http://localhost:%s/          - Главная (анализ пар)\n", s.port)
	fmt.Printf("   - http://localhost:%s/matrix    - Матрица корреляций\n", s.port)
	fmt.Printf("   - http://localhost:%s/biorhythms - Биоритмы всех субъектов\n", s.port)
	fmt.Printf("   - http://localhost:%s/timeline  - Динамика корреляций (90 дней)\n", s.port)
	fmt.Printf("   - http://localhost:%s/forecast  - Прогноз по сферам жизни\n", s.port)
	fmt.Printf("   - http://localhost:%s/help      - Справка\n", s.port)
	fmt.Println("\n📡 API эндпоинты:")
	fmt.Printf("   - /api/pairs      - анализ пар\n")
	fmt.Printf("   - /api/matrix     - матрица корреляций\n")
	fmt.Printf("   - /api/biorhythms - биоритмы всех\n")
	fmt.Printf("   - /api/timeline?a=Имя&b=Имя - динамика пары\n")
	fmt.Printf("   - /api/forecast?a=Имя&b=Имя - прогноз пары\n")
	fmt.Println("\n[!] Ctrl+C для остановки")

	return http.ListenAndServe(":"+s.port, nil)
}

func (s *Server) logMiddleware(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		s.log.Debug("Incoming: %s %s from %s", r.Method, r.URL.Path, r.RemoteAddr)

		lrw := &loggingResponseWriter{ResponseWriter: w, statusCode: http.StatusOK}
		next(lrw, r)

		duration := time.Since(start)
		s.log.Info("Completed: %s %s -> %d (%v)", r.Method, r.URL.Path, lrw.statusCode, duration)
	}
}

type loggingResponseWriter struct {
	http.ResponseWriter
	statusCode int
}

func (lrw *loggingResponseWriter) WriteHeader(code int) {
	lrw.statusCode = code
	lrw.ResponseWriter.WriteHeader(code)
}

func (s *Server) pairsHandler(w http.ResponseWriter, r *http.Request) {
	now := time.Now()
	people, err := storage.LoadPeople()
	if err != nil {
		s.log.Error("Failed to load people: %v", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	response := map[string]interface{}{
		"people_count": len(people),
		"pairs_count":  0,
		"pairs":        []interface{}{},
	}

	if len(people) == 0 {
		s.log.Warn("No people data available")
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(response)
		return
	}

	s.log.Debug("Processing %d people", len(people))

	type PairResponse struct {
		PersonA string  `json:"person_a"`
		PersonB string  `json:"person_b"`
		R       float64 `json:"r"`
		Harmony float64 `json:"harmony"`
		PhiDiff float64 `json:"phi_diff"`
		Status  string  `json:"status"`
	}

	pairs := []PairResponse{}
	for i := 0; i < len(people); i++ {
		if s.log != nil { s.log.Debug("Processing row %d for %s", i, people[i].Name) }
		for j := i + 1; j < len(people); j++ {
			if s.log != nil { s.log.Debug("  Computing correlation %s vs %s", people[i].Name, people[j].Name) }
			if people[i].Name == "" || people[j].Name == "" {
				if s.log != nil { s.log.Error("Empty name in matrix at i=%d, j=%d", i, j) }
				continue
			}
			r := metrics.CalculateCorrelation(people[i], people[j], now)
			pairs = append(pairs, PairResponse{
				PersonA: people[i].Name,
				PersonB: people[j].Name,
				R:       r,
				Harmony: metrics.CalculateHarmony(r),
				PhiDiff: math.Abs(math.Abs(r) - models.Phi),
				Status:  metrics.GetDetailedStatus(r),
			})
		}
	}

	sort.Slice(pairs, func(i, j int) bool {
		return math.Abs(pairs[i].R) > math.Abs(pairs[j].R)
	})

	response["pairs_count"] = len(pairs)
	response["pairs"] = pairs

	s.log.Debug("Generated %d pairs", len(pairs))

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}


func (s *Server) biorhythmsAllHandler(w http.ResponseWriter, r *http.Request) {
	now := time.Now()
	people, err := storage.LoadPeople()
	if err != nil {
		s.log.Error("Failed to load people: %v", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if len(people) == 0 {
		s.log.Warn("No people data available for biorhythms")
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode([]interface{}{})
		return
	}

	s.log.Debug("Calculating biorhythms for %d people", len(people))

	type BiorhythmData struct {
		Name   string             `json:"name"`
		Values map[string]float64 `json:"values"`
	}

	result := make([]BiorhythmData, 0)
	for _, p := range people {
		values := make(map[string]float64)
		for _, br := range models.Biorhythms {
			values[br.Name] = metrics.GetBiorhythm(p.BirthDate, now, br.Period)
		}
		result = append(result, BiorhythmData{Name: p.Name, Values: values})
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(result)
}

func (s *Server) timelineDataHandler(w http.ResponseWriter, r *http.Request) {
	now := time.Now()
	people, err := storage.LoadPeople()
	if err != nil {
		s.log.Error("Failed to load people: %v", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if len(people) == 0 {
		s.log.Warn("No people data available for timeline")
		http.Error(w, "Нет данных о субъектах", http.StatusBadRequest)
		return
	}

	personA := r.URL.Query().Get("a")
	personB := r.URL.Query().Get("b")

	if personA == "" || personB == "" {
		s.log.Warn("Missing person names: a=%s, b=%s", personA, personB)
		http.Error(w, "Не указаны имена субъектов", http.StatusBadRequest)
		return
	}

	s.log.Debug("Timeline request for %s - %s", personA, personB)

	var p1, p2 models.Person
	foundA, foundB := false, false
	for _, p := range people {
		if p.Name == personA {
			p1 = p
			foundA = true
		}
		if p.Name == personB {
			p2 = p
			foundB = true
		}
	}

	if !foundA || !foundB {
		s.log.Warn("Person not found: a=%s (found=%v), b=%s (found=%v)", personA, foundA, personB, foundB)
		http.Error(w, "Один или оба субъекта не найдены", http.StatusBadRequest)
		return
	}

	days := 90
	startDate := now.AddDate(0, 0, -days)
	timeline := metrics.CorrelationTimeline(p1, p2, startDate, days)

	s.log.Debug("Generated timeline with %d points", len(timeline))

	type TimelinePoint struct {
		Date    string  `json:"date"`
		R       float64 `json:"r"`
		Harmony float64 `json:"harmony"`
		Status  string  `json:"status"`
	}

	result := make([]TimelinePoint, len(timeline))
	for i, point := range timeline {
		result[i] = TimelinePoint{
			Date:    point.Date.Format("02.01.2006"),
			R:       point.R,
			Harmony: metrics.CalculateHarmony(point.R),
			Status:  point.Status,
		}
	}

	mean, median, variance := metrics.CorrelationStability(p1, p2, startDate, days)
	peakDate, peakR, peakStatus := metrics.FindPeakCorrelation(p1, p2, startDate, days)
	lowDate, lowR, lowStatus := metrics.FindLowestCorrelation(p1, p2, startDate, days)

	response := map[string]interface{}{
		"person_a": p1.Name,
		"person_b": p2.Name,
		"timeline": result,
		"statistics": map[string]interface{}{
			"mean":        mean,
			"median":      median,
			"variance":    variance,
			"peak_date":   peakDate.Format("02.01.2006"),
			"peak_r":      peakR,
			"peak_status": peakStatus,
			"low_date":    lowDate.Format("02.01.2006"),
			"low_r":       lowR,
			"low_status":  lowStatus,
		},
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

func (s *Server) forecastDataHandler(w http.ResponseWriter, r *http.Request) {
	now := time.Now()
	people, err := storage.LoadPeople()
	if err != nil {
		s.log.Error("Failed to load people: %v", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if len(people) == 0 {
		s.log.Warn("No people data available for forecast")
		http.Error(w, "Нет данных о субъектах", http.StatusBadRequest)
		return
	}

	personA := r.URL.Query().Get("a")
	personB := r.URL.Query().Get("b")
	daysStr := r.URL.Query().Get("days")
	var days int
	if daysStr != "" {
		fmt.Sscanf(daysStr, "%d", &days)
	}
	if days <= 0 || days > 365 {
		days = 30
	}

	if personA == "" || personB == "" {
		s.log.Warn("Missing person names: a=%s, b=%s", personA, personB)
		http.Error(w, "Не указаны имена субъектов", http.StatusBadRequest)
		return
	}

	s.log.Debug("Forecast request for %s - %s, days=%d", personA, personB, days)

	var p1, p2 models.Person
	foundA, foundB := false, false
	for _, p := range people {
		if p.Name == personA {
			p1 = p
			foundA = true
		}
		if p.Name == personB {
			p2 = p
			foundB = true
		}
	}

	if !foundA || !foundB {
		s.log.Warn("Person not found: a=%s (found=%v), b=%s (found=%v)", personA, foundA, personB, foundB)
		http.Error(w, "Один или оба субъекта не найдены", http.StatusBadRequest)
		return
	}

	startDate := now.AddDate(0, 0, 1)
	forecast := metrics.ForecastCorrelation(p1, p2, startDate, days)

	s.log.Debug("Generated forecast with %d points", len(forecast))

	type ForecastPoint struct {
		Date         string                       `json:"date"`
		R            float64                      `json:"r"`
		Harmony      float64                      `json:"harmony"`
		Status       string                       `json:"status"`
		SphereScores map[string]float64           `json:"sphere_scores"`
	}

	result := make([]ForecastPoint, len(forecast))
	for i, f := range forecast {
		result[i] = ForecastPoint{
			Date:         f.Date.Format("02.01.2006"),
			R:            f.R,
			Harmony:      f.Harmony,
			Status:       f.Status,
			SphereScores: f.SphereScores,
		}
	}

	spheres := []string{"💖 Любовь и отношения", "💼 Карьера и бизнес", "👥 Дружба и социализация", "🎨 Творчество", "💪 Здоровье и энергия", "📚 Обучение и развитие", "💰 Финансы"}
	bestDays := make(map[string][]map[string]interface{})

	for _, sphere := range spheres {
		best := metrics.FindBestDaysForSphere(p1, p2, startDate, days, sphere)
		daysList := make([]map[string]interface{}, 0)
		for _, d := range best {
			daysList = append(daysList, map[string]interface{}{
				"date":  d.Date.Format("02.01.2006"),
				"score": d.Score,
				"r":     d.R,
			})
		}
		bestDays[sphere] = daysList
	}

	response := map[string]interface{}{
		"person_a":  p1.Name,
		"person_b":  p2.Name,
		"forecast":  result,
		"best_days": bestDays,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

func (s *Server) indexHandler(w http.ResponseWriter, r *http.Request) {
	now := time.Now()

	html := `<!DOCTYPE html>
<html>
<head>
    <meta charset="UTF-8">
    <title>Главная | Психометрический анализатор</title>
    <style>
        * { margin: 0; padding: 0; box-sizing: border-box; }
        body {
            font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, sans-serif;
            background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
            min-height: 100vh;
        }
        .navbar {
            background: rgba(0,0,0,0.3);
            padding: 15px 30px;
            display: flex;
            gap: 20px;
            flex-wrap: wrap;
            align-items: center;
        }
        .navbar a {
            color: white;
            text-decoration: none;
            padding: 8px 16px;
            border-radius: 8px;
            transition: background 0.3s;
        }
        .navbar a:hover, .navbar a.active {
            background: rgba(255,255,255,0.2);
        }
        .datetime {
            background: rgba(0,0,0,0.2);
            color: white;
            padding: 8px 16px;
            border-radius: 8px;
            font-family: monospace;
            margin-left: auto;
        }
        .container {
            max-width: 1400px;
            margin: 20px auto;
            background: white;
            border-radius: 20px;
            overflow: hidden;
            box-shadow: 0 20px 60px rgba(0,0,0,0.3);
        }
        .header {
            background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
            color: white;
            padding: 30px;
            text-align: center;
        }
        .header h1 { font-size: 2em; margin-bottom: 10px; }
        .header .datetime-badge {
            background: rgba(255,255,255,0.2);
            display: inline-block;
            padding: 5px 15px;
            border-radius: 20px;
            font-size: 0.9em;
            margin-top: 10px;
        }
        .content { padding: 30px; }
        .stats {
            display: grid;
            grid-template-columns: repeat(auto-fit, minmax(200px, 1fr));
            gap: 20px;
            margin-bottom: 30px;
        }
        .stat-card {
            background: linear-gradient(135deg, #f5f7fa 0%, #c3cfe2 100%);
            padding: 20px;
            border-radius: 15px;
            text-align: center;
        }
        .stat-card h3 { color: #667eea; margin-bottom: 10px; }
        .stat-card .value { font-size: 2em; font-weight: bold; color: #764ba2; }
        table {
            width: 100%;
            border-collapse: collapse;
            margin-top: 20px;
        }
        th, td {
            padding: 12px;
            text-align: left;
            border: 1px solid #ddd;
        }
        th {
            background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
            color: white;
        }
        tr:hover { background-color: #f5f5f5; }
        .positive { color: #4ECDC4; font-weight: bold; }
        .negative { color: #FF6B6B; font-weight: bold; }
        .harmony-high { color: #4ECDC4; }
        .harmony-mid { color: #FFEAA7; }
        .harmony-low { color: #FF6B6B; }
        .loading { text-align: center; padding: 40px; color: #666; }
        .error-message {
            background: #ffe0e0;
            border-left: 4px solid #c00;
            padding: 20px;
            border-radius: 10px;
            color: #c00;
            text-align: center;
            margin: 20px 0;
        }
        .footer {
            background: #333;
            color: white;
            text-align: center;
            padding: 20px;
            font-size: 0.9em;
        }
        button {
            background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
            color: white;
            border: none;
            padding: 10px 20px;
            border-radius: 5px;
            cursor: pointer;
            margin: 10px;
        }
        button:hover { transform: translateY(-2px); }
    </style>
</head>
<body>
    <div class="navbar">
        <a href="/" class="active">🏠 Главная</a>
        <a href="/matrix">📊 Матрица</a>
        <a href="/biorhythms">📈 Биоритмы</a>
        <a href="/timeline">📉 Динамика</a>
        <a href="/forecast">🔮 Прогноз</a>
<a href="/zones">📖 31 зона</a>
        <a href="/help">❓ Справка</a>
        <div class="datetime">📅 ` + now.Format("02.01.2006 15:04:05") + `</div>
    </div>
    <div class="container">
        <div class="header">
            <h1>🧠 Психометрический анализатор</h1>
            <p>Анализ биоритмов и корреляций | 31 зона | Гармоничность | Прогнозы</p>
            <div class="datetime-badge">📅 Расчёт на: ` + now.Format("02.01.2006 15:04:05") + `</div>
        </div>
        <div class="content">
            <div class="stats" id="stats">
                <div class="stat-card"><h3>👥 Субъектов</h3><div class="value" id="peopleCount">-</div></div>
                <div class="stat-card"><h3>🔄 Пар</h3><div class="value" id="pairsCount">-</div></div>
                <div class="stat-card"><h3>🎯 Эталон Φ</h3><div class="value">0.618</div></div>
                <div class="stat-card"><h3>📊 Зон анализа</h3><div class="value">31</div></div>
            </div>
            
            <h2>📊 Анализ пар (31 зона психологической классификации)</h2>
            <div id="pairsTable"><div class="loading">Загрузка данных...</div></div>
            
            <div style="text-align: center; margin-top: 30px;">
                <button onclick="location.reload()">🔄 Обновить</button>
            </div>
        </div>
        <div class="footer">
            <p>🎵 Гармоничность (0-10) — близость к золотому сечению Φ = 0.618 | 7 уровней сознания, 31 зона</p>
            <p>📅 Дата расчёта: ` + now.Format("02.01.2006 15:04:05") + `</p>
        </div>
    </div>
    
    <script>
        function getStatusEmoji(status) {
            if (status.includes('СИМБИОЗ')) return '🔥';
            if (status.includes('ТРАНСЦЕНДЕНТНОСТЬ')) return '💫';
            if (status.includes('САМОАКТУАЛИЗАЦИЯ')) return '🕊️';
            if (status.includes('КОСМИЧЕСКАЯ ЛЮБОВЬ')) return '💎';
            if (status.includes('БОЖЕСТВЕННАЯ ГАРМОНИЯ')) return '💖';
            if (status.includes('ГЛУБОКАЯ ПРИВЯЗАННОСТЬ')) return '💞';
            if (status.includes('ИСКРЕННЯЯ БЛИЗОСТЬ')) return '💛';
            if (status.includes('ДРУЖЕСКАЯ СИМПАТИЯ')) return '🌸';
            if (status.includes('ВЗАИМОПОНИМАНИЕ')) return '🤝';
            if (status.includes('СИМПАТИЯ') && !status.includes('ДРУЖЕСКАЯ')) return '🌱';
            if (status.includes('ДОБРОЖЕЛАТЕЛЬНОСТЬ')) return '👋';
            if (status.includes('НЕЙТРАЛЬНО-ПОЗИТИВНОЕ')) return '😐';
            if (status.includes('ЛЁГКАЯ СИМПАТИЯ')) return '📍';
            if (status.includes('ЭМОЦИОНАЛЬНЫЙ НОЛЬ')) return '🧘';
            if (status.includes('ЛЁГКАЯ ОТСТРАНЁННОСТЬ')) return '🧊';
            if (status.includes('НАБЛЮДАТЕЛЬ')) return '🤨';
            if (status.includes('НЕОПРЕДЕЛЁННОСТЬ')) return '❓';
            if (status.includes('ЛЁГКОЕ НАПРЯЖЕНИЕ')) return '😌';
            if (status.includes('РАЗДРАЖЕНИЕ')) return '😤';
            if (status.includes('ОТЧУЖДЕНИЕ')) return '🥀';
            if (status.includes('НАПРЯЖЕНИЕ') && !status.includes('ЛЁГКОЕ')) return '⚡';
            if (status.includes('КОНФРОНТАЦИЯ')) return '🔥';
            if (status.includes('РАЗРЫВ')) return '💔';
            if (status.includes('ВРАЖДЕБНОСТЬ')) return '🗡️';
            if (status.includes('АНТАГОНИЗМ')) return '💀';
            if (status.includes('ПСИХОЛОГИЧЕСКОЕ ОТТОРЖЕНИЕ')) return '🌑';
            if (status.includes('ЭКЗИСТЕНЦИАЛЬНАЯ НЕСОВМЕСТИМОСТЬ')) return '🕳️';
            if (status.includes('ТОТАЛЬНЫЙ РАЗРЫВ')) return '🌌';
            return '📊';
        }
        
        async function loadData() {
            try {
                const response = await fetch('/api/pairs');
                if (!response.ok) {
                    throw new Error('HTTP ' + response.status);
                }
                const data = await response.json();
                
                document.getElementById('peopleCount').innerText = data.people_count || 0;
                document.getElementById('pairsCount').innerText = data.pairs_count || 0;
                
                if (!data.pairs || data.pairs.length === 0) {
                    document.getElementById('pairsTable').innerHTML = 
                        '<div class="error-message">⚠️ Нет данных для отображения. Добавьте субъектов в систему.</div>';
                    return;
                }
                
                let html = '<table border="1" style="border-collapse: collapse; width: 100%;"><thead>起源' +
                    '<th>#</th><th>Субъект А</th><th>Субъект Б</th><th>r (Пирсон)</th><th>Гармоничность</th><th>Близость к Φ</th><th>Статус (31 зона)</th>' +
                    '</thead><tbody>';
                
                let num = 1;
                for (const p of data.pairs) {
                    const rClass = p.r >= 0 ? 'positive' : 'negative';
                    let harmonyClass = 'harmony-mid';
                    if (p.harmony >= 7) harmonyClass = 'harmony-high';
                    if (p.harmony <= 3) harmonyClass = 'harmony-low';
                    
                    html += '<tr>' +
                        '<td>' + num++ + '</td>' +
                        '<td>' + escapeHtml(p.person_a) + '</td>' +
                        '<td>' + escapeHtml(p.person_b) + '</td>' +
                        '<td class="' + rClass + '">' + p.r.toFixed(4) + '</td>' +
                        '<td class="' + harmonyClass + '">' + p.harmony.toFixed(2) + '</td>' +
                        '<td>' + p.phi_diff.toFixed(4) + '</td>' +
                        '<td>' + getStatusEmoji(p.status) + ' ' + escapeHtml(p.status) + '</td>' +
                        '</tr>';
                }
                html += '</tbody></table>';
                document.getElementById('pairsTable').innerHTML = html;
            } catch (error) {
                console.error('Load error:', error);
                document.getElementById('pairsTable').innerHTML = 
                    '<div class="error-message">❌ Ошибка загрузки данных: ' + error.message + '</div>';
            }
        }
        
        function escapeHtml(text) {
            const div = document.createElement('div');
            div.textContent = text;
            return div.innerHTML;
        }
        
        loadData();
    </script>
</body>
</html>`
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.Write([]byte(html))
}

func (s *Server) matrixPageHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.Write([]byte(`<!DOCTYPE html>
<html><head><meta charset="UTF-8"><title>Матрица корреляций</title>
<style>
body{font-family:system-ui;background:linear-gradient(135deg,#667eea,#764ba2);padding:20px}
.navbar{background:rgba(0,0,0,0.3);padding:15px;display:flex;gap:20px;border-radius:15px;margin-bottom:20px}
.navbar a{color:#fff;text-decoration:none;padding:8px 16px;border-radius:8px}
.navbar a:hover{background:rgba(255,255,255,0.2)}
.container{max-width:1200px;margin:0 auto;background:#fff;border-radius:20px;padding:20px}
table{border-collapse:collapse;width:100%}
th,td{border:1px solid #ddd;padding:8px;text-align:center}
th{background:#667eea;color:#fff}
.error-message{background:#ffe0e0;padding:20px;border-radius:10px;color:#c00;text-align:center;margin:20px 0}
.loading{text-align:center;padding:40px;color:#666}
</style></head><body>
<a href="/zones">📖 31 зона</a>
<div class="navbar"><a href="/">🏠 Главная</a><a href="/matrix" class="active">📊 Матрица</a><a href="/biorhythms">📈 Биоритмы</a><a href="/timeline">📉 Динамика</a><a href="/forecast">🔮 Прогноз</a><a href="/help">❓ Справка</a></div>
<div class="container"><h1>📊 Матрица корреляций</h1><div id="matrix"><div class="loading">Загрузка...</div></div></div>
<script>
fetch('/api/matrix')
    .then(r => {
        if (!r.ok) throw new Error('HTTP ' + r.status);
        return r.json();
    })
    .then(d => {
        if(!d.names || d.names.length === 0) {
            document.getElementById('matrix').innerHTML = 
                '<div class="error-message">⚠️ Нет данных о субъектах. Добавьте людей в систему.</div>';
            return;
        }
        
        let h='<table><thead><tr><th>Субъект</th>';
        for(let n of d.names) h+='<th>'+n+'</th>';
        h+='</tr></thead><tbody>';
        
        for(let i=0;i<d.names.length;i++){
            h+='<tr><th>'+d.names[i]+'</th>';
            for(let j=0;j<d.names.length;j++){
                let v=d.matrix[i][j];
                let c=v>0.7?'#4ECDC4':v>0.3?'#96CEB4':v>-0.3?'#FFEAA7':v>-0.7?'#FFB6B6':'#FF6B6B';
                h+='<td style="background:'+c+'">'+v.toFixed(3)+'</td>';
            }
            h+='</tr>';
        }
        h+='</tbody></table>';
        document.getElementById('matrix').innerHTML=h;
    })
    .catch(err => {
        console.error('Matrix error:', err);
        document.getElementById('matrix').innerHTML = 
            '<div class="error-message">❌ Ошибка загрузки данных: ' + err.message + '</div>';
    });
</script></body></html>`))
}

func (s *Server) biorhythmsPageHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.Write([]byte(`<!DOCTYPE html>
<html><head><meta charset="UTF-8"><title>Биоритмы</title>
<style>
* { margin: 0; padding: 0; box-sizing: border-box; }
body{font-family:system-ui;background:linear-gradient(135deg,#667eea,#764ba2);padding:20px}
.navbar{background:rgba(0,0,0,0.3);padding:15px;display:flex;gap:20px;border-radius:15px;margin-bottom:20px}
.navbar a{color:#fff;text-decoration:none;padding:8px 16px;border-radius:8px}
.navbar a:hover{background:rgba(255,255,255,0.2)}
.container{max-width:1200px;margin:0 auto;background:#fff;border-radius:20px;padding:20px}
.biorhythm-card{background:#f9f9f9;border-radius:15px;padding:20px;margin-bottom:20px}
.biorhythm-card h2{color:#667eea;margin-bottom:15px;border-bottom:2px solid #667eea;padding-bottom:10px}
.bar-container{display:flex;align-items:center;gap:15px;margin:12px 0;flex-wrap:wrap}
.bar-label{width:140px;font-weight:bold;font-size:14px}
.bar-value{width:70px;font-family:monospace;font-weight:bold;text-align:right}
.bar-wrapper{flex:1;min-width:200px}
.bar{height:32px;background:#e0e0e0;border-radius:16px;overflow:hidden}
.bar-fill{height:100%;transition:width 0.3s ease;display:flex;align-items:center;justify-content:flex-end;padding-right:10px;color:#fff;font-size:12px;font-weight:bold}
.positive-bar{background:linear-gradient(90deg,#4ECDC4,#2ecc71)}
.negative-bar{background:linear-gradient(90deg,#e74c3c,#FF6B6B)}
.percentage{width:50px;font-size:12px;color:#666}
.error-message{background:#ffe0e0;padding:20px;border-radius:10px;color:#c00;text-align:center;margin:20px 0}
.loading{text-align:center;padding:40px;color:#666}
.grid{display:grid;grid-template-columns:repeat(auto-fill,minmax(500px,1fr));gap:20px}
@media (max-width: 768px){.grid{grid-template-columns:1fr}}
</style></head><body>
<a href="/zones">📖 31 зона</a>
<div class="navbar"><a href="/">🏠 Главная</a><a href="/matrix">📊 Матрица</a><a href="/biorhythms" class="active">📈 Биоритмы</a><a href="/timeline">📉 Динамика</a><a href="/forecast">🔮 Прогноз</a><a href="/help">❓ Справка</a></div>
<div class="container"><h1>📈 Биоритмы</h1><div id="biorhythms"><div class="loading">Загрузка...</div></div></div>
<script>
fetch('/api/biorhythms')
    .then(r => {
        if (!r.ok) throw new Error('HTTP ' + r.status);
        return r.json();
    })
    .then(d => {
        if(!d || d.length === 0) {
            document.getElementById('biorhythms').innerHTML = 
                '<div class="error-message">⚠️ Нет данных о субъектах. Добавьте людей в систему.</div>';
            return;
        }
        
        let h='<div class="grid">';
        for(let p of d){
            h+='<div class="biorhythm-card"><h2>👤 '+escape(p.name)+'</h2>';
            let items = Object.entries(p.values);
            for(let [name,val] of items){
                let pc=((val+1)/2*100).toFixed(0);
                let bc=val>=0?'positive-bar':'negative-bar';
                let bw=Math.abs(val)*100;
                let sign = val>=0 ? '+' : '';
                h+='<div class="bar-container">';
                h+='<div class="bar-label">'+name+'</div>';
                h+='<div class="bar-value">'+sign+val.toFixed(3)+'</div>';
                h+='<div class="bar-wrapper"><div class="bar"><div class="bar-fill '+bc+'" style="width:'+bw+'%">'+pc+'%</div></div></div>';
                h+='</div>';
            }
            h+='</div>';
        }
        h+='</div>';
        document.getElementById('biorhythms').innerHTML=h;
    })
    .catch(err => {
        console.error('Biorhythms error:', err);
        document.getElementById('biorhythms').innerHTML = 
            '<div class="error-message">❌ Ошибка загрузки данных: ' + err.message + '</div>';
    });
function escape(t){const d=document.createElement('div');d.textContent=t;return d.innerHTML;}
</script></body></html>`))
}


function getCandleChar(r){
    if(r>0.6) return '🟩';
    if(r>0.2) return '🟢';
    if(r>-0.2) return '🟡';
    if(r>-0.6) return '🟠';
    return '🔴';
}

function getHarmonyChar(h){
    if(h>8) return '💎';
    if(h>6) return '✨';
    if(h>4) return '🌱';
    if(h>2) return '🍂';
    return '💔';
}

function formatDateAxis(dates, periodDays){
    let axisHtml='<div class="date-axis">';
    let step=periodDays<=7?periodDays:Math.max(7,Math.floor(periodDays/10));
    for(let i=0;i<dates.length;i+=step){
        let d=new Date(dates[i]);
        let label=d.getDate()+'.'+(d.getMonth()+1);
        let leftPos=(i/dates.length)*100;
        axisHtml+='<span style="position:relative;left:'+leftPos+'%;margin-left:-20px;">'+label+'</span>';
    }
    axisHtml+='</div>';
    return axisHtml;
}

function generateWeekSummary(timeline){
    let weeks=[];
    let currentWeek=[];
    let currentWeekStart=null;
    for(let i=0;i<timeline.length;i++){
        let date=new Date(timeline[i].date);
        let weekNum=Math.floor(date.getTime()/(7*24*60*60*1000));
        if(currentWeekStart===null || weekNum!==currentWeekStart){
            if(currentWeek.length>0) weeks.push(currentWeek);
            currentWeek=[timeline[i]];
            currentWeekStart=weekNum;
        }else{
            currentWeek.push(timeline[i]);
        }
    }
    if(currentWeek.length>0) weeks.push(currentWeek);

    let weekSummaries='<div class="week-summaries"><h3>📅 Недельная аналитика</h3>';
    for(let w=0;w<weeks.length;w++){
        let week=weeks[w];
        if(week.length===0) continue;
        let avgR=week.reduce((s,p)=>s+p.r,0)/week.length;
        let avgHarmony=week.reduce((s,p)=>s+p.harmony,0)/week.length;
        let dominantCandle=getCandleChar(avgR);
        let dominantHarmony=getHarmonyChar(avgHarmony);
        let weekStart=new Date(week[0].date);
        let weekEnd=new Date(week[week.length-1].date);
        let weekLabel=weekStart.getDate()+'.'+(weekStart.getMonth()+1)+' — '+weekEnd.getDate()+'.'+(weekEnd.getMonth()+1);
        let climate='';
        if(avgR>0.4) climate='🌿 Период созидания и роста';
        else if(avgR>0.1) climate='🌱 Зарождение понимания';
        else if(avgR>-0.1) climate='⚖️ Период стабильности';
        else if(avgR>-0.4) climate='🍂 Охлаждение. Потеря ритма';
        else climate='🔥 КРИЗИС. Режим тишины';

        weekSummaries+='<div class="week-summary">';
        weekSummaries+='<h4>📅 '+weekLabel+'</h4>';
        weekSummaries+='<div>Свечной фон: '+dominantCandle+' | Гармония: '+dominantHarmony+'</div>';
        weekSummaries+='<div>Средний r: '+avgR.toFixed(3)+' | Средняя гармония: '+avgHarmony.toFixed(1)+'</div>';
        weekSummaries+='<div><strong>'+climate+'</strong></div>';
        weekSummaries+='</div>';
    }
    weekSummaries+='</div>';

    let totalAvgR=timeline.reduce((s,p)=>s+p.r,0)/timeline.length;
    let totalAvgHarmony=timeline.reduce((s,p)=>s+p.harmony,0)/timeline.length;
    let totalClimate='';
    if(totalAvgR>0.4) totalClimate='🌿 Период созидания и роста';
    else if(totalAvgR>0.1) totalClimate='🌱 Зарождение понимания';
    else if(totalAvgR>-0.1) totalClimate='⚖️ Период стабильности';
    else if(totalAvgR>-0.4) totalClimate='🍂 Охлаждение. Потеря ритма';
    else totalClimate='🔥 КРИЗИС. Режим тишины';

    weekSummaries+='<div class="week-summary" style="background:#e8f5e9"><h4>📊 ОБЩИЙ КЛИМАТ ЗА ПЕРИОД</h4>';
    weekSummaries+='<div>Средний r: '+totalAvgR.toFixed(3)+' | Средняя гармония: '+totalAvgHarmony.toFixed(1)+'</div>';
    weekSummaries+='<div><strong>'+totalClimate+'</strong></div>';
    weekSummaries+='</div>';

    return weekSummaries;
}

async function loadTimeline(){
    let a=document.getElementById('personA').value,b=document.getElementById('personB').value;
    if(!a||!b || a==='Нет данных' || b==='Нет данных'){
        document.getElementById('result').innerHTML='<div class="error-message">⚠️ Недостаточно данных для анализа</div>';
        return;
    }
    document.getElementById('result').innerHTML='<div class="loading">Загрузка...</div>';
    try {
        let r=await fetch('/api/timeline?a='+encodeURIComponent(a)+'&b='+encodeURIComponent(b));
        if(!r.ok){
            throw new Error('Ошибка загрузки данных');
        }
        let d=await r.json();
        
        if(!d.timeline || d.timeline.length===0){
            document.getElementById('result').innerHTML='<div class="error-message">⚠️ Нет данных для отображения</div>';
            return;
        }

        let html='<div class="stats-card"><h3>📊 Статистика за 90 дней</h3>';
        html+='<p>Среднее r: '+d.statistics.mean.toFixed(4)+' | Медиана: '+d.statistics.median.toFixed(4)+' | Дисперсия: '+d.statistics.variance.toFixed(4)+'</p>';
        html+='<p>📈 Максимум: '+d.statistics.peak_date+' r='+d.statistics.peak_r.toFixed(4)+'<br>'+d.statistics.peak_status+'</p>';
        html+='<p>📉 Минимум: '+d.statistics.low_date+' r='+d.statistics.low_r.toFixed(4)+'<br>'+d.statistics.low_status+'</p></div>';

        let candleSymbols='';
        let dates=[];
        for(let i=0;i<d.timeline.length;i++){
            candleSymbols+=getCandleChar(d.timeline[i].r);
            dates.push(d.timeline[i].date);
        }
        html+='<div class="graph-container"><h3>📈 Свечной график (цвет = знак и сила)</h3>';
        html+='<div class="candle-graph" id="candleGraph">'+candleSymbols+'</div>';
        html+=formatDateAxis(dates,90);
        html+='<div class="legend"><div class="legend-item">🟩 r>0.6</div><div class="legend-item">🟢 0.2-0.6</div><div class="legend-item">🟡 -0.2-0.2</div><div class="legend-item">🟠 -0.6 - -0.2</div><div class="legend-item">🔴 r<-0.6</div></div></div>';

        let harmonySymbols='';
        for(let i=0;i<d.timeline.length;i++){
            harmonySymbols+=getHarmonyChar(d.timeline[i].harmony);
        }
        html+='<div class="graph-container"><h3>🎵 Гармоничность (близость к Φ = 0.618)</h3>';
        html+='<div class="harmony-graph" id="harmonyGraph">'+harmonySymbols+'</div>';
        html+=formatDateAxis(dates,90);
        html+='<div class="legend"><div class="legend-item">💎 >8</div><div class="legend-item">✨ 6-8</div><div class="legend-item">🌱 4-6</div><div class="legend-item">🍂 2-4</div><div class="legend-item">💔 <2</div></div></div>';

        html+=generateWeekSummary(d.timeline);
        document.getElementById('result').innerHTML=html;
    } catch(err) {
        console.error('Timeline error:', err);
        document.getElementById('result').innerHTML='<div class="error-message">❌ Ошибка загрузки данных: ' + err.message + '</div>';
    }
}
</script></body></html>`))
}

func (s *Server) forecastPageHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.Write([]byte(`<!DOCTYPE html>
<html><head><meta charset="UTF-8"><title>Прогноз</title>
<style>
body{font-family:system-ui;background:linear-gradient(135deg,#667eea,#764ba2);padding:20px}
.navbar{background:rgba(0,0,0,0.3);padding:15px;display:flex;gap:20px;border-radius:15px;margin-bottom:20px}
.navbar a{color:#fff;text-decoration:none;padding:8px 16px;border-radius:8px}
.navbar a:hover{background:rgba(255,255,255,0.2)}
.container{max-width:1400px;margin:0 auto;background:#fff;border-radius:20px;padding:20px}
.selector{background:#f5f5f5;padding:20px;border-radius:15px;margin-bottom:20px;display:flex;gap:20px;flex-wrap:wrap}
.selector select,.selector button{padding:10px;border-radius:8px}
.selector button{background:linear-gradient(135deg,#667eea,#764ba2);color:#fff;border:none;cursor:pointer}
.period-btns{display:flex;gap:10px}
.period-btn{padding:10px 15px;background:#fff;border:1px solid #667eea;border-radius:8px;cursor:pointer}
.period-btn.active{background:linear-gradient(135deg,#667eea,#764ba2);color:#fff}
.forecast-grid{display:grid;grid-template-columns:repeat(auto-fill,minmax(350px,1fr));gap:20px;margin-top:20px}
.forecast-card{background:#fff;border-radius:15px;overflow:hidden;box-shadow:0 4px 15px rgba(0,0,0,0.1)}
.card-header{background:linear-gradient(135deg,#667eea,#764ba2);color:#fff;padding:15px;text-align:center}
.card-status{padding:10px;text-align:center;background:#f9f9f9}
.card-scores{padding:15px}
.score-item{display:flex;align-items:center;gap:10px;margin:8px 0}
.score-name{width:130px;font-size:0.85em}
.score-bar{flex:1;height:25px;background:#e0e0e0;border-radius:12px;overflow:hidden}
.score-fill{height:100%;background:linear-gradient(90deg,#4ECDC4,#2ecc71);display:flex;align-items:center;justify-content:flex-end;padding-right:5px;color:#fff;font-size:0.7em}
.best-worst{background:#f9f9f9;border-radius:15px;padding:20px;margin-bottom:30px;display:grid;grid-template-columns:1fr 1fr;gap:20px}
.best-card,.worst-card{padding:15px;border-radius:12px;text-align:center}
.best-card{background:#e8f5e9}
.worst-card{background:#ffebee}
.big-date{font-size:1.3em;font-weight:bold;margin:10px 0}
.loading{text-align:center;padding:40px}
.positive{color:#4ECDC4}
.negative{color:#FF6B6B}
.error-message{background:#ffe0e0;padding:20px;border-radius:10px;color:#c00;text-align:center;margin:20px 0}
</style></head><body>
<a href="/zones">📖 31 зона</a>
<div class="navbar"><a href="/">🏠 Главная</a><a href="/matrix">📊 Матрица</a><a href="/biorhythms">📈 Биоритмы</a><a href="/timeline">📉 Динамика</a><a href="/forecast" class="active">🔮 Прогноз</a><a href="/help">❓ Справка</a></div>
<div class="container"><h1>🔮 Прогноз</h1>
<div class="selector"><select id="personA"></select> ↔ <select id="personB"></select>
<div class="period-btns"><button class="period-btn" data-days="7">Неделя</button><button class="period-btn active" data-days="30">Месяц</button><button class="period-btn" data-days="90">Квартал</button><button class="period-btn" data-days="365">Год</button></div>
<button onclick="loadForecast()">🔮 Прогноз</button></div>
<div id="result"><div class="loading">Выберите пару</div></div></div>
<script>
let people=[],curDays=30;
fetch('/api/pairs').then(r=>r.json()).then(d=>{
    if(d.pairs && d.pairs.length > 0){
        people=[...new Set(d.pairs.flatMap(p=>[p.person_a,p.person_b]))];
    }
    let a=document.getElementById('personA'),b=document.getElementById('personB');
    if(people.length > 0){
        people.forEach(p=>{a.innerHTML+='<option>'+p+'</option>';b.innerHTML+='<option>'+p+'</option>'});
        a.value=people[0];
        if(people[1]) b.value=people[1];
    } else {
        a.innerHTML='<option disabled>Нет данных</option>';
        b.innerHTML='<option disabled>Нет данных</option>';
    }
});
document.querySelectorAll('.period-btn').forEach(b=>{
b.onclick=function(){
document.querySelectorAll('.period-btn').forEach(x=>x.classList.remove('active'));
this.classList.add('active');
curDays=parseInt(this.dataset.days);
if(document.getElementById('personA').value && document.getElementById('personA').value !== 'Нет данных')loadForecast();
};
});
function getEmoji(s){
if(s.includes('СИМБИОЗ'))return '🔥';if(s.includes('ТРАНСЦЕНДЕНТНОСТЬ'))return '💫';
if(s.includes('САМОАКТУАЛИЗАЦИЯ'))return '🕊️';if(s.includes('КОСМИЧЕСКАЯ ЛЮБОВЬ'))return '💎';
if(s.includes('БОЖЕСТВЕННАЯ ГАРМОНИЯ'))return '💖';if(s.includes('ГЛУБОКАЯ ПРИВЯЗАННОСТЬ'))return '💞';
if(s.includes('ИСКРЕННЯЯ БЛИЗОСТЬ'))return '💛';if(s.includes('ДРУЖЕСКАЯ СИМПАТИЯ'))return '🌸';
if(s.includes('ВЗАИМОПОНИМАНИЕ'))return '🤝';if(s.includes('СИМПАТИЯ'))return '🌱';
if(s.includes('ДОБРОЖЕЛАТЕЛЬНОСТЬ'))return '👋';if(s.includes('НЕЙТРАЛЬНО-ПОЗИТИВНОЕ'))return '😐';
if(s.includes('ЛЁГКАЯ СИМПАТИЯ'))return '📍';if(s.includes('ЭМОЦИОНАЛЬНЫЙ НОЛЬ'))return '🧘';
if(s.includes('ЛЁГКАЯ ОТСТРАНЁННОСТЬ'))return '🧊';if(s.includes('НАБЛЮДАТЕЛЬ'))return '🤨';
if(s.includes('НЕОПРЕДЕЛЁННОСТЬ'))return '❓';if(s.includes('ЛЁГКОЕ НАПРЯЖЕНИЕ'))return '😌';
if(s.includes('РАЗДРАЖЕНИЕ'))return '😤';if(s.includes('ОТЧУЖДЕНИЕ'))return '🥀';
if(s.includes('КОНФРОНТАЦИЯ'))return '🔥';if(s.includes('РАЗРЫВ'))return '💔';
if(s.includes('ВРАЖДЕБНОСТЬ'))return '🗡️';if(s.includes('АНТАГОНИЗМ'))return '💀';
if(s.includes('ПСИХОЛОГИЧЕСКОЕ ОТТОРЖЕНИЕ'))return '🌑';if(s.includes('ЭКЗИСТЕНЦИАЛЬНАЯ НЕСОВМЕСТИМОСТЬ'))return '🕳️';
return '📊';
}
async function loadForecast(){
let a=document.getElementById('personA').value,b=document.getElementById('personB').value;
if(!a||!b || a==='Нет данных' || b==='Нет данных'){
document.getElementById('result').innerHTML='<div class="error-message">⚠️ Недостаточно данных для прогноза</div>';
return;
}
document.getElementById('result').innerHTML='<div class="loading">Загрузка...</div>';
try {
let r=await fetch('/api/forecast?a='+encodeURIComponent(a)+'&b='+encodeURIComponent(b)+'&days='+curDays);
if(!r.ok) throw new Error('Ошибка загрузки');
let d=await r.json();
if(!d.forecast || d.forecast.length===0){
document.getElementById('result').innerHTML='<div class="error-message">⚠️ Нет данных для прогноза</div>';
return;
}
let fc=[...d.forecast].sort((x,y)=>new Date(x.date)-new Date(y.date));
let best=fc[0],worst=fc[0];
for(let f of fc){if(f.r>best.r)best=f;if(f.r<worst.r)worst=f;}
let html='<div class="best-worst"><div class="best-card"><h4>🌟 Лучший день</h4><div class="big-date">'+best.date+'</div><div class="positive">r='+best.r.toFixed(4)+'</div><div>Гармония: '+best.harmony.toFixed(1)+'</div><div>'+getEmoji(best.status)+' '+best.status+'</div></div>';
html+='<div class="worst-card"><h4>⚠️ Худший день</h4><div class="big-date">'+worst.date+'</div><div class="negative">r='+worst.r.toFixed(4)+'</div><div>Гармония: '+worst.harmony.toFixed(1)+'</div><div>'+getEmoji(worst.status)+' '+worst.status+'</div></div></div>';
html+='<div class="forecast-grid">';
for(let f of fc){
html+='<div class="forecast-card"><div class="card-header"><div>📅 '+f.date+'</div><div class="'+(f.r>=0?'positive':'negative')+'">r='+f.r.toFixed(4)+'</div><div>🎵 '+f.harmony.toFixed(1)+'</div></div>';
html+='<div class="card-status">'+getEmoji(f.status)+' '+f.status+'</div><div class="card-scores">';
for(let [sp,sc] of Object.entries(f.sphere_scores)){
html+='<div class="score-item"><div class="score-name">'+sp+'</div><div class="score-bar"><div class="score-fill" style="width:'+sc+'%">'+Math.round(sc)+'%</div></div></div>';
}
html+='</div></div>';
}
html+='</div>';
document.getElementById('result').innerHTML=html;
} catch(err) {
console.error('Forecast error:', err);
document.getElementById('result').innerHTML='<div class="error-message">❌ Ошибка загрузки данных: ' + err.message + '</div>';
}
}
</script></body></html>`))
}

func (s *Server) helpPageHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.Write([]byte(`<!DOCTYPE html>
<html><head><meta charset="UTF-8"><title>Справка</title>
<style>
body{font-family:system-ui;background:linear-gradient(135deg,#667eea,#764ba2);padding:20px}
.navbar{background:rgba(0,0,0,0.3);padding:15px;display:flex;gap:20px;border-radius:15px;margin-bottom:20px}
.navbar a{color:#fff;text-decoration:none;padding:8px 16px;border-radius:8px}
.navbar a:hover{background:rgba(255,255,255,0.2)}
.container{max-width:1000px;margin:0 auto;background:#fff;border-radius:20px;padding:30px}
h1,h2{color:#667eea}
.zone-list{display:grid;grid-template-columns:repeat(auto-fit,minmax(300px,1fr));gap:15px;margin:20px 0}
.zone-category{background:#f5f5f5;padding:15px;border-radius:10px}
</style></head><body>
<div class="navbar"><a href="/">🏠 Главная</a><a href="/matrix">📊 Матрица</a><a href="/biorhythms">📈 Биоритмы</a><a href="/timeline">📉 Динамика</a><a href="/forecast">🔮 Прогноз</a><a href="/help" class="active">❓ Справка</a></div>
<div class="container"><h1>❓ Справка</h1>
<h2>31 зона психологической классификации</h2>
<p>7 уровней сознания, каждый разделён на подуровни:</p>
<div class="zone-list"><div class="zone-category"><h3>🔴 Уровень 7: Самоактуализация (r > 0.95)</h3><ul><li>r>0.98 — Симбиоз</li><li>r>0.96 — Трансцендентность</li><li>r>0.95 — Самоактуализация</li></ul></div>
<div class="zone-category"><h3>🟠 Уровень 6: Гармония (0.80-0.95)</h3><ul><li>r>0.90 — Космическая любовь</li><li>r>0.85 — Божественная гармония</li><li>r>0.80 — Глубокая привязанность</li></ul></div>
<div class="zone-category"><h3>🟡 Уровень 5: Любовь (0.60-0.80)</h3><ul><li>r>0.75 — Искренняя близость</li><li>r>0.70 — Дружеская симпатия</li><li>r>0.65 — Взаимопонимание</li><li>r>0.60 — Симпатия</li></ul></div>
<div class="zone-category"><h3>🟢 Уровень 4: Стабильность (0.30-0.60)</h3><ul><li>r>0.50 — Доброжелательность</li><li>r>0.45 — Нейтрально-позитивное</li><li>r>0.40 — Лёгкая симпатия</li><li>r>0.30 — Эмоциональный ноль</li></ul></div>
<div class="zone-category"><h3>🔵 Уровень 3: Нейтральность (0.00-0.30)</h3><ul><li>r>0.20 — Лёгкая отстранённость</li><li>r>0.10 — Наблюдатель</li><li>r>0.00 — Неопределённость</li></ul></div>
<div class="zone-category"><h3>🟣 Уровень 2: Напряжение (-0.30-0.00)</h3><ul><li>r>-0.10 — Лёгкое напряжение</li><li>r>-0.20 — Раздражение</li><li>r>-0.30 — Отчуждение</li></ul></div>
<div class="zone-category"><h3>⚫ Уровень 1: Конфликт (-0.60 - -0.30)</h3><ul><li>r>-0.40 — Напряжение</li><li>r>-0.50 — Конфронтация</li><li>r>-0.60 — Разрыв</li></ul></div>
<div class="zone-category"><h3>⚪ Уровень 0: Антагонизм (r < -0.60)</h3><ul><li>r>-0.70 — Враждебность</li><li>r>-0.80 — Антагонизм</li><li>r>-0.90 — Психологическое отторжение</li><li>r>-0.95 — Экзистенциальная несовместимость</li><li>r≤-0.95 — Тотальный разрыв</li></ul></div></div>
<h2>🎵 Гармоничность</h2><p>Оценка от 0 до 10, где 10 — идеальная близость к золотому сечению (Φ = 0.618).</p>
<h2>🔮 Прогноз по сферам жизни</h2><p>7 сфер: любовь, карьера, дружба, творчество, здоровье, обучение, финансы.</p>
<h2>📈 Биоритмы</h2><p>5 биоритмов: физический(23), эмоциональный(28), интеллектуальный(33), духовный(38), интуитивный(42).</p>
</div></body></html>`))
}
func (s *Server) matrixHandler(w http.ResponseWriter, r *http.Request) {
	now := time.Now()
	people, err := storage.LoadPeople()
	if err != nil {
		if s.log != nil {
			s.log.Error("Failed to load people: %v", err)
		}
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if s.log != nil {
		s.log.Debug("Processing %d people for matrix", len(people))
	}

	if len(people) == 0 {
		response := map[string]interface{}{
			"names":  []string{},
			"matrix": [][]float64{},
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(response)
		return
	}

	names := make([]string, len(people))
	for i, p := range people {
		names[i] = p.Name
	}

	// Создаём матрицу
	matrix := make([][]float64, len(people))
	for i := 0; i < len(people); i++ {
		matrix[i] = make([]float64, len(people))
		matrix[i][i] = 1.0
	}

	// ТОЧНО ТАКОЙ ЖЕ ЦИКЛ, как в pairsHandler
	for i := 0; i < len(people); i++ {
		for j := i + 1; j < len(people); j++ {
			r := metrics.CalculateCorrelation(people[i], people[j], now)
			matrix[i][j] = r
			matrix[j][i] = r
			if s.log != nil {
				s.log.Debug("Matrix[%d][%d] = %f (%s-%s)", i, j, r, people[i].Name, people[j].Name)
			}
		}
	}

	response := map[string]interface{}{
		"names":  names,
		"matrix": matrix,
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}
func (s *Server) timelinePageHandler(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte(`<!DOCTYPE html>
<html><head><meta charset="UTF-8"><title>Динамика | Психометрический анализатор</title>
<style>
body{font-family:system-ui;background:linear-gradient(135deg,#667eea,#764ba2);padding:20px}
.navbar{background:rgba(0,0,0,0.3);padding:15px;display:flex;gap:20px;border-radius:15px;margin-bottom:20px}
.navbar a{color:#fff;text-decoration:none;padding:8px 16px;border-radius:8px}
.navbar a:hover{background:rgba(255,255,255,0.2)}
.container{max-width:1400px;margin:0 auto;background:#fff;border-radius:20px;padding:20px}
.selector{background:#f5f5f5;padding:20px;border-radius:15px;margin-bottom:20px;display:flex;gap:20px;flex-wrap:wrap;align-items:center}
.selector select,button{padding:10px;border-radius:8px}
button{background:linear-gradient(135deg,#667eea,#764ba2);color:#fff;border:none;cursor:pointer}
.stats-card{background:#f5f7fa;padding:20px;border-radius:15px;margin-bottom:20px}
.graph-container{background:#f9f9f9;padding:20px;border-radius:15px;margin-bottom:20px;overflow-x:auto}
.candle-graph{font-size:24px;letter-spacing:2px;font-family:monospace;white-space:nowrap}
.harmony-graph{font-size:24px;letter-spacing:2px;font-family:monospace;white-space:nowrap}
.date-axis{font-family:monospace;font-size:12px;margin-top:8px;white-space:nowrap}
.date-axis span{display:inline-block;text-align:center;margin-right:20px}
.legend{display:flex;gap:20px;flex-wrap:wrap;margin-top:15px;padding:10px;background:#f0f0f0;border-radius:10px}
.legend-item{display:flex;align-items:center;gap:8px}
.loading{text-align:center;padding:40px}
.positive{color:#4ECDC4}
.negative{color:#FF6B6B}
.week-summary{background:#f0f0f0;padding:10px;border-radius:10px;margin:10px 0}
.week-summary h4{margin:0 0 5px 0}
.error-message{background:#ffe0e0;padding:20px;border-radius:10px;color:#c00;text-align:center;margin:20px 0}
</style></head><body>
<div class="navbar">
<a href="/">🏠 Главная</a>
<a href="/matrix">📊 Матрица</a>
<a href="/biorhythms">📈 Биоритмы</a>
<a href="/timeline" class="active">📉 Динамика</a>
<a href="/forecast">🔮 Анализ</a>
<a href="/zones">📖 31 зона</a>
<a href="/help">❓ Справка</a>
</div>
<div class="container">
<h1>📉 Динамика корреляций</h1>
<div class="selector">
<select id="personA"></select> ↔ <select id="personB"></select>
<button onclick="loadTimeline()">📊 Показать</button>
</div>
<div id="result"><div class="loading">Выберите пару</div></div>
</div>
<script>
let people=[];
fetch('/api/pairs').then(r=>r.json()).then(d=>{
    if(d.pairs && d.pairs.length > 0){
        people=[...new Set(d.pairs.flatMap(p=>[p.person_a,p.person_b]))];
    }
    let a=document.getElementById('personA'),b=document.getElementById('personB');
    if(people.length > 0){
        people.forEach(p=>{a.innerHTML+='<option>'+p+'</option>';b.innerHTML+='<option>'+p+'</option>'});
        a.value=people[0];
        if(people[1]) b.value=people[1];
    }
});

function getCandleChar(r){
    if(r>0.6) return '🟩';
    if(r>0.2) return '🟢';
    if(r>-0.2) return '🟡';
    if(r>-0.6) return '🟠';
    return '🔴';
}

function getHarmonyChar(h){
    if(h>8) return '💎';
    if(h>6) return '✨';
    if(h>4) return '🌱';
    if(h>2) return '🍂';
    return '💔';
}

function formatDateAxis(dates){
    if(!dates || dates.length===0) return '';
    let axisHtml='<div class="date-axis">';
    let step=Math.max(1, Math.floor(dates.length/10));
    for(let i=0;i<dates.length;i+=step){
        let d=dates[i];
        if(d && d!=='Invalid Date'){
            let parts=d.split('.');
            let label=parts[0]+'.'+parts[1];
            axisHtml+='<span>'+label+'</span>';
        }
    }
    axisHtml+='</div>';
    return axisHtml;
}

function formatDate(d){
    if(!d) return '?';
    let parts=d.split('.');
    if(parts.length===3){
        return parts[0]+'.'+parts[1];
    }
    return d;
}

function generateWeekSummary(timeline){
    if(!timeline || timeline.length===0) return '';
    
    let weeks=[];
    let currentWeek=[];
    for(let i=0;i<timeline.length;i++){
        currentWeek.push(timeline[i]);
        if(currentWeek.length===7 || i===timeline.length-1){
            weeks.push([...currentWeek]);
            currentWeek=[];
        }
    }
    
    let weekSummaries='<div class="week-summaries"><h3>📅 Недельная аналитика</h3>';
    for(let w=0;w<weeks.length;w++){
        let week=weeks[w];
        if(week.length===0) continue;
        let avgR=week.reduce((s,p)=>s+p.r,0)/week.length;
        let avgHarmony=week.reduce((s,p)=>s+p.harmony,0)/week.length;
        let startDate=week[0].date;
        let endDate=week[week.length-1].date;
        
        let climate='';
        if(avgR>0.4) climate='🌿 Период созидания и роста';
        else if(avgR>0.1) climate='🌱 Зарождение понимания';
        else if(avgR>-0.1) climate='⚖️ Период стабильности';
        else if(avgR>-0.4) climate='🍂 Охлаждение. Потеря ритма';
        else climate='🔥 КРИЗИС. Режим тишины';
        
        weekSummaries+='<div class="week-summary">';
        weekSummaries+='<h4>📅 '+formatDate(startDate)+' — '+formatDate(endDate)+'</h4>';
        weekSummaries+='<div>Средний r: '+avgR.toFixed(3)+' | Средняя гармония: '+avgHarmony.toFixed(1)+'</div>';
        weekSummaries+='<div><strong>'+climate+'</strong></div>';
        weekSummaries+='</div>';
    }
    
    let totalAvgR=timeline.reduce((s,p)=>s+p.r,0)/timeline.length;
    let totalAvgHarmony=timeline.reduce((s,p)=>s+p.harmony,0)/timeline.length;
    let totalClimate='';
    if(totalAvgR>0.4) totalClimate='🌿 Период созидания и роста';
    else if(totalAvgR>0.1) totalClimate='🌱 Зарождение понимания';
    else if(totalAvgR>-0.1) totalClimate='⚖️ Период стабильности';
    else if(totalAvgR>-0.4) totalClimate='🍂 Охлаждение. Потеря ритма';
    else totalClimate='🔥 КРИЗИС. Режим тишины';
    
    weekSummaries+='<div class="week-summary" style="background:#e8f5e9;margin-top:20px"><h4>📊 ОБЩИЙ КЛИМАТ ЗА ПЕРИОД</h4>';
    weekSummaries+='<div>Средний r: '+totalAvgR.toFixed(3)+' | Средняя гармония: '+totalAvgHarmony.toFixed(1)+'</div>';
    weekSummaries+='<div><strong>'+totalClimate+'</strong></div>';
    weekSummaries+='</div>';
    
    return weekSummaries;
}

async function loadTimeline(){
    let a=document.getElementById('personA').value,b=document.getElementById('personB').value;
    if(!a||!b){
        document.getElementById('result').innerHTML='<div class="error-message">Выберите двух человек</div>';
        return;
    }
    document.getElementById('result').innerHTML='<div class="loading">Загрузка...</div>';
    try {
        let r=await fetch('/api/timeline?a='+encodeURIComponent(a)+'&b='+encodeURIComponent(b));
        if(!r.ok) throw new Error('Ошибка загрузки');
        let d=await r.json();
        
        if(!d.timeline || d.timeline.length===0){
            document.getElementById('result').innerHTML='<div class="error-message">Нет данных за выбранный период</div>';
            return;
        }
        
        let html='<div class="stats-card"><h3>📊 Статистика за 90 дней</h3>';
        html+='<p>Среднее r: '+d.statistics.mean.toFixed(4)+' | Медиана: '+d.statistics.median.toFixed(4)+' | Дисперсия: '+d.statistics.variance.toFixed(4)+'</p>';
        html+='<p>📈 Максимум: '+d.statistics.peak_date+' r='+d.statistics.peak_r.toFixed(4)+'<br>'+d.statistics.peak_status+'</p>';
        html+='<p>📉 Минимум: '+d.statistics.low_date+' r='+d.statistics.low_r.toFixed(4)+'<br>'+d.statistics.low_status+'</p></div>';
        
        let candleSymbols='';
        let dates=[];
        for(let i=0;i<d.timeline.length;i++){
            candleSymbols+=getCandleChar(d.timeline[i].r);
            dates.push(d.timeline[i].date);
        }
        html+='<div class="graph-container"><h3>📈 Свечной график (цвет = знак и сила)</h3>';
        html+='<div class="candle-graph">'+candleSymbols+'</div>';
        html+=formatDateAxis(dates);
        html+='<div class="legend"><div class="legend-item">🟩 r>0.6</div><div class="legend-item">🟢 0.2-0.6</div><div class="legend-item">🟡 -0.2-0.2</div><div class="legend-item">🟠 -0.6 - -0.2</div><div class="legend-item">🔴 r<-0.6</div></div></div>';
        
        let harmonySymbols='';
        for(let i=0;i<d.timeline.length;i++){
            harmonySymbols+=getHarmonyChar(d.timeline[i].harmony);
        }
        html+='<div class="graph-container"><h3>🎵 Гармоничность (близость к Φ = 0.618)</h3>';
        html+='<div class="harmony-graph">'+harmonySymbols+'</div>';
        html+=formatDateAxis(dates);
        html+='<div class="legend"><div class="legend-item">💎 >8</div><div class="legend-item">✨ 6-8</div><div class="legend-item">🌱 4-6</div><div class="legend-item">🍂 2-4</div><div class="legend-item">💔 <2</div></div></div>';
        
        html+=generateWeekSummary(d.timeline);
        document.getElementById('result').innerHTML=html;
    } catch(err) {
        console.error('Timeline error:', err);
        document.getElementById('result').innerHTML='<div class="error-message">❌ Ошибка: '+err.message+'</div>';
    }
}
</script></body></html>`))
}
