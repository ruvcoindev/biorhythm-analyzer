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
	http.HandleFunc("/zones", s.logMiddleware(s.zonesPageHandler))
	http.HandleFunc("/help", s.logMiddleware(s.helpPageHandler))
	http.HandleFunc("/api/pairs", s.logMiddleware(s.pairsHandler))
	http.HandleFunc("/api/matrix", s.logMiddleware(s.matrixHandler))
	http.HandleFunc("/api/biorhythms", s.logMiddleware(s.biorhythmsAllHandler))
	http.HandleFunc("/api/timeline", s.logMiddleware(s.timelineDataHandler))
	http.HandleFunc("/api/forecast", s.logMiddleware(s.forecastDataHandler))

	s.log.Info("Web server starting on port %s", s.port)
	fmt.Printf("\n🌐 Веб-интерфейс: http://localhost:%s\n", s.port)
	fmt.Printf("📅 Текущее время сервера: %s\n", time.Now().Format("02.01.2006 15:04:05"))
	fmt.Println("\n📊 Доступные страницы:")
	fmt.Printf("   - http://localhost:%s/          - Главная\n", s.port)
	fmt.Printf("   - http://localhost:%s/matrix    - Матрица\n", s.port)
	fmt.Printf("   - http://localhost:%s/biorhythms - Биоритмы\n", s.port)
	fmt.Printf("   - http://localhost:%s/timeline  - Динамика\n", s.port)
	fmt.Printf("   - http://localhost:%s/forecast  - Прогноз\n", s.port)
	fmt.Printf("   - http://localhost:%s/zones     - 31 зона\n", s.port)
	fmt.Printf("   - http://localhost:%s/help      - Справка\n", s.port)
	fmt.Println("\n📡 API эндпоинты:")
	fmt.Printf("   - /api/pairs\n")
	fmt.Printf("   - /api/matrix\n")
	fmt.Printf("   - /api/biorhythms\n")
	fmt.Printf("   - /api/timeline\n")
	fmt.Printf("   - /api/forecast\n")
	fmt.Println("\n[!] Ctrl+C для остановки")

	return http.ListenAndServe(":"+s.port, nil)
}

func (s *Server) logMiddleware(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		s.log.Debug("Incoming: %s %s", r.Method, r.URL.Path)
		lrw := &loggingResponseWriter{ResponseWriter: w, statusCode: http.StatusOK}
		next(lrw, r)
		s.log.Info("Completed: %s %s -> %d (%v)", r.Method, r.URL.Path, lrw.statusCode, time.Since(start))
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

// ==================== API HANDLERS ====================

func (s *Server) pairsHandler(w http.ResponseWriter, r *http.Request) {
	now := time.Now()
	people, err := storage.LoadPeople()
	if err != nil {
		s.log.Error("Failed to load people: %v", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

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
		for j := i + 1; j < len(people); j++ {
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

	response := map[string]interface{}{
		"people_count": len(people),
		"pairs_count":  len(pairs),
		"pairs":        pairs,
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

func (s *Server) matrixHandler(w http.ResponseWriter, r *http.Request) {
	now := time.Now()
	people, err := storage.LoadPeople()
	if err != nil {
		s.log.Error("Failed to load people: %v", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if len(people) == 0 {
		json.NewEncoder(w).Encode(map[string]interface{}{
			"names":  []string{},
			"matrix": [][]float64{},
		})
		return
	}

	names := make([]string, len(people))
	for i, p := range people {
		names[i] = p.Name
	}

	matrix := make([][]float64, len(people))
	for i := 0; i < len(people); i++ {
		matrix[i] = make([]float64, len(people))
		matrix[i][i] = 1.0
		for j := i + 1; j < len(people); j++ {
			r := metrics.CalculateCorrelation(people[i], people[j], now)
			matrix[i][j] = r
			matrix[j][i] = r
		}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"names":  names,
		"matrix": matrix,
	})
}

func (s *Server) biorhythmsAllHandler(w http.ResponseWriter, r *http.Request) {
	now := time.Now()
	people, err := storage.LoadPeople()
	if err != nil {
		s.log.Error("Failed to load people: %v", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	type BiorhythmData struct {
		Name   string             `json:"name"`
		Values map[string]float64 `json:"values"`
	}

	result := []BiorhythmData{}
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

	personA := r.URL.Query().Get("a")
	personB := r.URL.Query().Get("b")

	var p1, p2 models.Person
	for _, p := range people {
		if p.Name == personA {
			p1 = p
		}
		if p.Name == personB {
			p2 = p
		}
	}

	if p1.Name == "" || p2.Name == "" {
		http.Error(w, "Люди не найдены", http.StatusBadRequest)
		return
	}

	days := 90
	startDate := now.AddDate(0, 0, -days)
	timeline := metrics.CorrelationTimeline(p1, p2, startDate, days)

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

	personA := r.URL.Query().Get("a")
	personB := r.URL.Query().Get("b")
	daysStr := r.URL.Query().Get("days")
	days := 30
	if daysStr != "" {
		fmt.Sscanf(daysStr, "%d", &days)
	}
	if days <= 0 || days > 365 {
		days = 30
	}

	var p1, p2 models.Person
	for _, p := range people {
		if p.Name == personA {
			p1 = p
		}
		if p.Name == personB {
			p2 = p
		}
	}

	if p1.Name == "" || p2.Name == "" {
		http.Error(w, "Люди не найдены", http.StatusBadRequest)
		return
	}

	startDate := now.AddDate(0, 0, 1)
	forecast := metrics.ForecastCorrelation(p1, p2, startDate, days)

	type ForecastPoint struct {
		Date         string             `json:"date"`
		R            float64            `json:"r"`
		Harmony      float64            `json:"harmony"`
		Status       string             `json:"status"`
		SphereScores map[string]float64 `json:"sphere_scores"`
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

	response := map[string]interface{}{
		"person_a": p1.Name,
		"person_b": p2.Name,
		"forecast": result,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// ==================== PAGE HANDLERS ====================

func (s *Server) indexHandler(w http.ResponseWriter, r *http.Request) {
	now := time.Now()
	html := `<!DOCTYPE html>
<html><head><meta charset="UTF-8"><title>Психометрический анализатор</title>
<style>
*{margin:0;padding:0;box-sizing:border-box}
body{font-family:system-ui;background:linear-gradient(135deg,#667eea,#764ba2);min-height:100vh}
.navbar{background:rgba(0,0,0,0.3);padding:15px 30px;display:flex;gap:20px;flex-wrap:wrap;align-items:center}
.navbar a{color:#fff;text-decoration:none;padding:8px 16px;border-radius:8px}
.navbar a:hover,.navbar a.active{background:rgba(255,255,255,0.2)}
.datetime{background:rgba(0,0,0,0.2);color:#fff;padding:8px 16px;border-radius:8px;margin-left:auto}
.container{max-width:1400px;margin:20px auto;background:#fff;border-radius:20px;overflow:hidden}
.header{background:linear-gradient(135deg,#667eea,#764ba2);color:#fff;padding:30px;text-align:center}
.content{padding:30px}
.stats{display:grid;grid-template-columns:repeat(auto-fit,minmax(200px,1fr));gap:20px;margin-bottom:30px}
.stat-card{background:linear-gradient(135deg,#f5f7fa,#c3cfe2);padding:20px;border-radius:15px;text-align:center}
.stat-card h3{color:#667eea}
.stat-card .value{font-size:2em;font-weight:bold;color:#764ba2}
table{width:100%;border-collapse:collapse;margin-top:20px}
th,td{padding:12px;text-align:left;border:1px solid #ddd}
th{background:linear-gradient(135deg,#667eea,#764ba2);color:#fff}
tr:hover{background:#f5f5f5}
.positive{color:#4ECDC4;font-weight:bold}
.negative{color:#FF6B6B;font-weight:bold}
.harmony-high{color:#4ECDC4}
.harmony-mid{color:#FFEAA7}
.harmony-low{color:#FF6B6B}
.loading{text-align:center;padding:40px}
.footer{background:#333;color:#fff;text-align:center;padding:20px}
button{background:linear-gradient(135deg,#667eea,#764ba2);color:#fff;border:none;padding:10px 20px;border-radius:5px;cursor:pointer;margin:10px}
</style></head><body>
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
<div class="header"><h1>🧠 Психометрический анализатор</h1><p>31 зона | Гармоничность | Прогнозы</p></div>
<div class="content">
<div class="stats"><div class="stat-card"><h3>👥 Субъектов</h3><div class="value" id="peopleCount">-</div></div>
<div class="stat-card"><h3>🔄 Пар</h3><div class="value" id="pairsCount">-</div></div>
<div class="stat-card"><h3>🎯 Эталон Φ</h3><div class="value">0.618</div></div>
<div class="stat-card"><h3>📊 Зон анализа</h3><div class="value">31</div></div></div>
<h2>📊 Анализ пар</h2>
<div id="pairsTable"><div class="loading">Загрузка...</div></div>
<div style="text-align:center"><button onclick="location.reload()">🔄 Обновить</button></div>
</div>
<div class="footer"><p>Гармоничность (0-10) — близость к Φ = 0.618 | 31 зона</p></div>
</div>
<script>
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
return '📊';}
fetch('/api/pairs').then(r=>r.json()).then(d=>{
document.getElementById('peopleCount').innerText=d.people_count;
document.getElementById('pairsCount').innerText=d.pairs_count;
let h='<table border="1"><thead><tr><th>#</th><th>А</th><th>Б</th><th>r</th><th>Гармония</th><th>|r-Φ|</th><th>Статус</th></tr></thead><tbody>';
let n=1;
for(let p of d.pairs){
let cls=p.r>=0?'positive':'negative';
let hc='harmony-mid';
if(p.harmony>=7)hc='harmony-high';
if(p.harmony<=3)hc='harmony-low';
h+='<tr><td>'+n+++'</td><td>'+escape(p.person_a)+'</td><td>'+escape(p.person_b)+'</td><td class="'+cls+'">'+p.r.toFixed(4)+'</td><td class="'+hc+'">'+p.harmony.toFixed(2)+'</td><td>'+p.phi_diff.toFixed(4)+'</td><td>'+getEmoji(p.status)+' '+escape(p.status)+'</td></tr>';
}
h+='</tbody></table>';
document.getElementById('pairsTable').innerHTML=h;
});
function escape(t){const d=document.createElement('div');d.textContent=t;return d.innerHTML;}
</script>
</body></html>`
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.Write([]byte(html))
}

func (s *Server) matrixPageHandler(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte(`<!DOCTYPE html>
<html><head><meta charset="UTF-8"><title>Матрица</title>
<style>body{font-family:system-ui;background:linear-gradient(135deg,#667eea,#764ba2);padding:20px}
.navbar{background:rgba(0,0,0,0.3);padding:15px;display:flex;gap:20px;border-radius:15px;margin-bottom:20px}
.navbar a{color:#fff;text-decoration:none;padding:8px 16px;border-radius:8px}
.navbar a:hover{background:rgba(255,255,255,0.2)}
.container{max-width:1200px;margin:0 auto;background:#fff;border-radius:20px;padding:20px}
table{border-collapse:collapse;width:100%}
th,td{border:1px solid #ddd;padding:8px;text-align:center}
th{background:#667eea;color:#fff}
</style></head><body>
<div class="navbar"><a href="/">🏠 Главная</a><a href="/matrix" class="active">📊 Матрица</a><a href="/biorhythms">📈 Биоритмы</a><a href="/timeline">📉 Динамика</a><a href="/forecast">🔮 Прогноз</a><a href="/zones">📖 31 зона</a><a href="/help">❓ Справка</a></div>
<div class="container"><h1>📊 Матрица корреляций</h1><div id="matrix">Загрузка...</div></div>
<script>fetch('/api/matrix').then(r=>r.json()).then(d=>{
let h='<table><thead><tr><th>Субъект</th>';
for(let n of d.names)h+='<th>'+n+'</th>';
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
});</script></body></html>`))
}

func (s *Server) biorhythmsPageHandler(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte(`<!DOCTYPE html>
<html><head><meta charset="UTF-8"><title>Биоритмы</title>
<style>body{font-family:system-ui;background:linear-gradient(135deg,#667eea,#764ba2);padding:20px}
.navbar{background:rgba(0,0,0,0.3);padding:15px;display:flex;gap:20px;border-radius:15px;margin-bottom:20px}
.navbar a{color:#fff;text-decoration:none;padding:8px 16px;border-radius:8px}
.navbar a:hover{background:rgba(255,255,255,0.2)}
.container{max-width:1200px;margin:0 auto;background:#fff;border-radius:20px;padding:20px}
.biorhythm-card{background:#f9f9f9;border-radius:15px;padding:20px;margin-bottom:20px}
.bar{height:30px;background:#e0e0e0;border-radius:15px;overflow:hidden}
.bar-fill{height:100%}
.positive-bar{background:linear-gradient(90deg,#4ECDC4,#2ecc71)}
.negative-bar{background:linear-gradient(90deg,#e74c3c,#FF6B6B)}
.biorhythm-item{display:flex;align-items:center;margin:10px 0;gap:15px;flex-wrap:wrap}
.biorhythm-name{width:120px;font-weight:bold}
.biorhythm-value{width:80px}
.bar-container{flex:1;min-width:200px}
</style></head><body>
<div class="navbar"><a href="/">🏠 Главная</a><a href="/matrix">📊 Матрица</a><a href="/biorhythms" class="active">📈 Биоритмы</a><a href="/timeline">📉 Динамика</a><a href="/forecast">🔮 Прогноз</a><a href="/zones">📖 31 зона</a><a href="/help">❓ Справка</a></div>
<div class="container"><h1>📈 Биоритмы</h1><div id="biorhythms">Загрузка...</div></div>
<script>
fetch('/api/biorhythms').then(r=>r.json()).then(d=>{
let h='';
for(let p of d){
h+='<div class="biorhythm-card"><h2>👤 '+escape(p.name)+'</h2>';
for(let [name,val] of Object.entries(p.values)){
let pc=((val+1)/2*100).toFixed(0);
let bc=val>=0?'positive-bar':'negative-bar';
let bw=Math.abs(val)*100;
h+='<div class="biorhythm-item"><div class="biorhythm-name">'+name+'</div>';
h+='<div class="biorhythm-value">'+val.toFixed(3)+'</div>';
h+='<div class="bar-container"><div class="bar"><div class="bar-fill '+bc+'" style="width:'+bw+'%"></div></div></div>';
h+='<div>'+pc+'%</div></div>';
}
h+='</div>';
}
document.getElementById('biorhythms').innerHTML=h;
});
function escape(t){const d=document.createElement('div');d.textContent=t;return d.innerHTML;}
</script></body></html>`))
}

func (s *Server) timelinePageHandler(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte(`<!DOCTYPE html>
<html><head><meta charset="UTF-8"><title>Динамика</title>
<style>body{font-family:system-ui;background:linear-gradient(135deg,#667eea,#764ba2);padding:20px}
.navbar{background:rgba(0,0,0,0.3);padding:15px;display:flex;gap:20px;border-radius:15px;margin-bottom:20px}
.navbar a{color:#fff;text-decoration:none;padding:8px 16px;border-radius:8px}
.navbar a:hover{background:rgba(255,255,255,0.2)}
.container{max-width:1400px;margin:0 auto;background:#fff;border-radius:20px;padding:20px}
.selector{background:#f5f5f5;padding:20px;border-radius:15px;margin-bottom:20px;display:flex;gap:20px;flex-wrap:wrap}
.selector select,button{padding:10px;border-radius:8px}
button{background:linear-gradient(135deg,#667eea,#764ba2);color:#fff;border:none;cursor:pointer}
.stats-card{background:#f5f7fa;padding:20px;border-radius:15px;margin-bottom:20px}
.graph{background:#f9f9f9;padding:20px;border-radius:15px;font-size:24px;margin-bottom:20px;overflow-x:auto}
.loading{text-align:center;padding:40px}
</style></head><body>
<div class="navbar"><a href="/">🏠 Главная</a><a href="/matrix">📊 Матрица</a><a href="/biorhythms">📈 Биоритмы</a><a href="/timeline" class="active">📉 Динамика</a><a href="/forecast">🔮 Прогноз</a><a href="/zones">📖 31 зона</a><a href="/help">❓ Справка</a></div>
<div class="container"><h1>📉 Динамика</h1>
<div class="selector"><select id="personA"></select> ↔ <select id="personB"></select><button onclick="load()">Показать</button></div>
<div id="result"><div class="loading">Выберите пару</div></div></div>
<script>
let people=[];
fetch('/api/pairs').then(r=>r.json()).then(d=>{
people=[...new Set(d.pairs.flatMap(p=>[p.person_a,p.person_b]))];
let a=document.getElementById('personA'),b=document.getElementById('personB');
people.forEach(p=>{a.innerHTML+='<option>'+p+'</option>';b.innerHTML+='<option>'+p+'</option>'});
if(people.length>=2){a.value=people[0];b.value=people[1];}
});
function getCandle(r){
if(r>0.6)return '🟩';if(r>0.2)return '🟢';if(r>-0.2)return '🟡';if(r>-0.6)return '🟠';return '🔴';
}
function getHarmony(h){
if(h>8)return '💎';if(h>6)return '✨';if(h>4)return '🌱';if(h>2)return '🍂';return '💔';
}
async function load(){
let a=document.getElementById('personA').value,b=document.getElementById('personB').value;
if(!a||!b)return;
document.getElementById('result').innerHTML='<div class="loading">Загрузка...</div>';
let r=await fetch('/api/timeline?a='+encodeURIComponent(a)+'&b='+encodeURIComponent(b));
let d=await r.json();
let html='<div class="stats-card"><h3>Статистика</h3><p>Среднее: '+d.statistics.mean.toFixed(4)+' | Медиана: '+d.statistics.median.toFixed(4)+' | Дисперсия: '+d.statistics.variance.toFixed(4)+'</p>';
html+='<p>📈 Максимум: '+d.statistics.peak_date+' r='+d.statistics.peak_r.toFixed(4)+'<br>'+d.statistics.peak_status+'</p>';
html+='<p>📉 Минимум: '+d.statistics.low_date+' r='+d.statistics.low_r.toFixed(4)+'<br>'+d.statistics.low_status+'</p></div>';
html+='<div class="graph"><h3>Свечной график</h3><div>';
for(let i=0;i<d.timeline.length;i++)html+=getCandle(d.timeline[i].r);
html+='</div></div><div class="graph"><h3>Гармоничность</h3><div>';
for(let i=0;i<d.timeline.length;i++)html+=getHarmony(d.timeline[i].harmony);
html+='</div></div>';
document.getElementById('result').innerHTML=html;
}
</script></body></html>`))
}

func (s *Server) forecastPageHandler(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte(`<!DOCTYPE html>
<html><head><meta charset="UTF-8"><title>Прогноз</title>
<style>body{font-family:system-ui;background:linear-gradient(135deg,#667eea,#764ba2);padding:20px}
.navbar{background:rgba(0,0,0,0.3);padding:15px;display:flex;gap:20px;border-radius:15px;margin-bottom:20px}
.navbar a{color:#fff;text-decoration:none;padding:8px 16px;border-radius:8px}
.navbar a:hover{background:rgba(255,255,255,0.2)}
.container{max-width:1400px;margin:0 auto;background:#fff;border-radius:20px;padding:20px}
.selector{background:#f5f5f5;padding:20px;border-radius:15px;margin-bottom:20px;display:flex;gap:20px;flex-wrap:wrap}
.selector select,button{padding:10px;border-radius:8px}
button{background:linear-gradient(135deg,#667eea,#764ba2);color:#fff;border:none;cursor:pointer}
.period-btns{display:flex;gap:10px}
.period-btn{padding:10px 15px;background:#fff;border:1px solid #667eea;border-radius:8px;cursor:pointer}
.period-btn.active{background:linear-gradient(135deg,#667eea,#764ba2);color:#fff}
.forecast-grid{display:grid;grid-template-columns:repeat(auto-fill,minmax(350px,1fr));gap:20px;margin-top:20px}
.forecast-card{background:#fff;border-radius:15px;overflow:hidden;box-shadow:0 4px 15px rgba(0,0,0,0.1)}
.card-header{background:linear-gradient(135deg,#667eea,#764ba2);color:#fff;padding:15px;text-align:center}
.card-status{padding:10px;text-align:center;background:#f9f9f9}
.card-scores{padding:15px}
.score-item{display:flex;align-items:center;gap:10px;margin:8px 0}
.score-name{width:130px}
.score-bar{flex:1;height:25px;background:#e0e0e0;border-radius:12px;overflow:hidden}
.score-fill{height:100%;background:linear-gradient(90deg,#4ECDC4,#2ecc71);display:flex;align-items:center;justify-content:flex-end;padding-right:5px;color:#fff}
.best-worst{background:#f9f9f9;border-radius:15px;padding:20px;margin-bottom:30px;display:grid;grid-template-columns:1fr 1fr;gap:20px}
.best-card,.worst-card{padding:15px;border-radius:12px;text-align:center}
.best-card{background:#e8f5e9}
.worst-card{background:#ffebee}
.loading{text-align:center;padding:40px}
</style></head><body>
<div class="navbar"><a href="/">🏠 Главная</a><a href="/matrix">📊 Матрица</a><a href="/biorhythms">📈 Биоритмы</a><a href="/timeline">📉 Динамика</a><a href="/forecast" class="active">🔮 Прогноз</a><a href="/zones">📖 31 зона</a><a href="/help">❓ Справка</a></div>
<div class="container"><h1>🔮 Прогноз</h1>
<div class="selector"><select id="personA"></select> ↔ <select id="personB"></select>
<div class="period-btns"><button class="period-btn" data-days="7">Неделя</button><button class="period-btn active" data-days="30">Месяц</button><button class="period-btn" data-days="90">Квартал</button><button class="period-btn" data-days="365">Год</button></div>
<button onclick="loadForecast()">Прогноз</button></div>
<div id="result"><div class="loading">Выберите пару</div></div></div>
<script>
let people=[],curDays=30;
fetch('/api/pairs').then(r=>r.json()).then(d=>{
people=[...new Set(d.pairs.flatMap(p=>[p.person_a,p.person_b]))];
let a=document.getElementById('personA'),b=document.getElementById('personB');
people.forEach(p=>{a.innerHTML+='<option>'+p+'</option>';b.innerHTML+='<option>'+p+'</option>'});
if(people.length>=2){a.value=people[0];b.value=people[1];}
});
document.querySelectorAll('.period-btn').forEach(b=>{
b.onclick=function(){
document.querySelectorAll('.period-btn').forEach(x=>x.classList.remove('active'));
this.classList.add('active');
curDays=parseInt(this.dataset.days);
if(document.getElementById('personA').value)loadForecast();
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
if(!a||!b)return;
document.getElementById('result').innerHTML='<div class="loading">Загрузка...</div>';
let r=await fetch('/api/forecast?a='+encodeURIComponent(a)+'&b='+encodeURIComponent(b)+'&days='+curDays);
let d=await r.json();
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
}
</script></body></html>`))
}

func (s *Server) zonesPageHandler(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte(`<!DOCTYPE html>
<html><head><meta charset="UTF-8"><title>31 зона</title>
<style>body{font-family:system-ui;background:linear-gradient(135deg,#667eea,#764ba2);padding:20px}
.navbar{background:rgba(0,0,0,0.3);padding:15px;display:flex;gap:20px;border-radius:15px;margin-bottom:20px}
.navbar a{color:#fff;text-decoration:none;padding:8px 16px;border-radius:8px}
.navbar a:hover{background:rgba(255,255,255,0.2)}
.container{max-width:1200px;margin:0 auto;background:#fff;border-radius:20px;padding:30px}
h1{color:#667eea;border-bottom:3px solid #667eea;padding-bottom:10px;margin-bottom:20px}
h2{color:#667eea;margin:25px 0 15px}
.zone-card{background:#f9f9f9;border-radius:15px;margin-bottom:15px}
.zone-header{display:flex;align-items:center;gap:15px;padding:15px 20px;cursor:pointer;background:#f5f7fa}
.zone-header .emoji{font-size:32px}
.zone-header .range{font-family:monospace;font-weight:bold;min-width:100px}
.zone-header .name{flex:1;font-weight:bold;font-size:18px}
.zone-header .toggle{font-size:20px;color:#667eea}
.zone-detail{display:none;padding:20px;background:#fff;border-top:1px solid #eee}
.zone-detail.show{display:block}
.zone-desc{margin-bottom:15px}
.zone-interpret{background:#e8f5e9;padding:12px;border-radius:10px}
.back-link{display:inline-block;margin-top:30px;padding:10px 20px;background:linear-gradient(135deg,#667eea,#764ba2);color:#fff;text-decoration:none;border-radius:8px}
.footer{background:#333;color:#fff;text-align:center;padding:20px;margin-top:30px;border-radius:15px}
</style></head><body>
<div class="navbar"><a href="/">🏠 Главная</a><a href="/matrix">📊 Матрица</a><a href="/biorhythms">📈 Биоритмы</a><a href="/timeline">📉 Динамика</a><a href="/forecast">🔮 Прогноз</a><a href="/zones" class="active">📖 31 зона</a><a href="/help">❓ Справка</a></div>
<div class="container"><h1>📖 31 зона психологической классификации</h1><p>7 уровней сознания, 31 состояние. Золотое сечение Φ = 0.618 — идеальная гармония.</p><div id="zones"></div><a href="/" class="back-link">← На главную</a></div>
<div class="footer"><p>31 зона — авторская классификация на основе корреляции Пирсона и золотого сечения</p></div>
<script>
const zones=[
{level:7,emoji:"🔥",range:"r>0.98",name:"СИМБИОЗ",desc:"Полное слияние личностей",strength:"Максимальное взаимопонимание",weakness:"Потеря индивидуальности",advice:"Сохраняйте личное пространство"},
{level:7,emoji:"💫",range:"r>0.96",name:"ТРАНСЦЕНДЕНТНОСТЬ",desc:"Выход за пределы личности",strength:"Духовный рост",weakness:"Отрыв от реальности",advice:"Баланс духовного и бытового"},
{level:7,emoji:"🕊️",range:"r>0.95",name:"САМОАКТУАЛИЗАЦИЯ",desc:"Полная реализация потенциала",strength:"Взаимное развитие",weakness:"Высокие ожидания",advice:"Цените маленькие шаги"},
{level:6,emoji:"💎",range:"r>0.90",name:"КОСМИЧЕСКАЯ ЛЮБОВЬ",desc:"Безусловное принятие",strength:"Безусловная любовь",weakness:"Риск жертвенности",advice:"Любите, но не теряйте себя"},
{level:6,emoji:"💖",range:"r>0.85",name:"БОЖЕСТВЕННАЯ ГАРМОНИЯ",desc:"Идеальный резонанс",strength:"Лёгкость",weakness:"Иллюзия вечного счастья",advice:"Не бойтесь конфликтов"},
{level:6,emoji:"💞",range:"r>0.80",name:"ГЛУБОКАЯ ПРИВЯЗАННОСТЬ",desc:"Душевное родство",strength:"Надёжность",weakness:"Рутина",advice:"Вносите разнообразие"},
{level:5,emoji:"💛",range:"r>0.75",name:"ИСКРЕННЯЯ БЛИЗОСТЬ",desc:"Тёплые доверительные отношения",strength:"Доверие",weakness:"Эмоциональное выгорание",advice:"Делитесь чувствами"},
{level:5,emoji:"🌸",range:"r>0.70",name:"ДРУЖЕСКАЯ СИМПАТИЯ",desc:"Естественное притяжение",strength:"Лёгкость",weakness:"Поверхностность",advice:"Позвольте себе быть уязвимым"},
{level:5,emoji:"🤝",range:"r>0.65",name:"ВЗАИМОПОНИМАНИЕ",desc:"Согласованность ценностей",strength:"Общие цели",weakness:"Риск скучности",advice:"Ставьте совместные цели"},
{level:5,emoji:"🌱",range:"r>0.60",name:"СИМПАТИЯ",desc:"Начало близости",strength:"Надежда",weakness:"Хрупкость",advice:"Будьте бережны"},
{level:4,emoji:"👋",range:"r>0.50",name:"ДОБРОЖЕЛАТЕЛЬНОСТЬ",desc:"Открытость к контакту",strength:"Безопасность",weakness:"Пассивность",advice:"Делайте первый шаг"},
{level:4,emoji:"😐",range:"r>0.45",name:"НЕЙТРАЛЬНО-ПОЗИТИВНОЕ",desc:"Комфортное сосуществование",strength:"Спокойствие",weakness:"Застой",advice:"Ищите точки роста"},
{level:4,emoji:"📍",range:"r>0.40",name:"ЛЁГКАЯ СИМПАТИЯ",desc:"Без обязательств",strength:"Свобода",weakness:"Боязнь обязательств",advice:"Будьте честны"},
{level:4,emoji:"🧘",range:"r>0.30",name:"ЭМОЦИОНАЛЬНЫЙ НОЛЬ",desc:"Спокойное равнодушие",strength:"Объективность",weakness:"Отсутствие тепла",advice:"Позвольте себе чувствовать"},
{level:3,emoji:"🧊",range:"r>0.20",name:"ЛЁГКАЯ ОТСТРАНЁННОСТЬ",desc:"Дипломатичность",strength:"Безопасность",weakness:"Холодность",advice:"Приоткройтесь"},
{level:3,emoji:"🤨",range:"r>0.10",name:"НАБЛЮДАТЕЛЬ",desc:"Сторонний анализ",strength:"Объективность",weakness:"Отсутствие вовлечённости",advice:"Решитесь на шаг"},
{level:3,emoji:"❓",range:"r>0.00",name:"НЕОПРЕДЕЛЁННОСТЬ",desc:"Формирование отношения",strength:"Открытость",weakness:"Нестабильность",advice:"Дайте себе время"},
{level:2,emoji:"😌",range:"r>-0.10",name:"ЛЁГКОЕ НАПРЯЖЕНИЕ",desc:"Притирка",strength:"Динамика",weakness:"Дискомфорт",advice:"Терпение"},
{level:2,emoji:"😤",range:"r>-0.20",name:"РАЗДРАЖЕНИЕ",desc:"Мелкие конфликты",strength:"Честность",weakness:"Взрывчатость",advice:"Дышите"},
{level:2,emoji:"🥀",range:"r>-0.30",name:"ОТЧУЖДЕНИЕ",desc:"Эмоциональная дистанция",strength:"Защита",weakness:"Потеря связи",advice:"Поговорите"},
{level:1,emoji:"⚡",range:"r>-0.40",name:"НАПРЯЖЕНИЕ",desc:"Постоянные трения",strength:"Прямота",weakness:"Усталость",advice:"Возьмите паузу"},
{level:1,emoji:"🔥",range:"r>-0.50",name:"КОНФРОНТАЦИЯ",desc:"Открытое противостояние",strength:"Ясность",weakness:"Истощение",advice:"Согласитесь не соглашаться"},
{level:1,emoji:"💔",range:"r>-0.60",name:"РАЗРЫВ",desc:"Потеря эмоциональной связи",strength:"Конец страданиям",weakness:"Боль потери",advice:"Примите"},
{level:0,emoji:"🗡️",range:"r>-0.70",name:"ВРАЖДЕБНОСТЬ",desc:"Системный конфликт",strength:"Ясность",weakness:"Постоянная борьба",advice:"Минимизируйте контакты"},
{level:0,emoji:"💀",range:"r>-0.80",name:"АНТАГОНИЗМ",desc:"Непримиримое противостояние",strength:"Предсказуемость",weakness:"Разрушительное влияние",advice:"Признайте несовместимость"},
{level:0,emoji:"🌑",range:"r>-0.90",name:"ПСИХОЛОГИЧЕСКОЕ ОТТОРЖЕНИЕ",desc:"Аверсия",strength:"Чёткая граница",weakness:"Интенсивный дискомфорт",advice:"Прекратите контакт"},
{level:0,emoji:"🕳️",range:"r>-0.95",name:"ЭКЗИСТЕНЦИАЛЬНАЯ НЕСОВМЕСТИМОСТЬ",desc:"Полное неприятие",strength:"Абсолютная ясность",weakness:"Невозможность взаимодействия",advice:"Не пытайтесь"},
{level:0,emoji:"🌌",range:"r≤-0.95",name:"ТОТАЛЬНЫЙ РАЗРЫВ",desc:"Энергетический вакуум",strength:"Окончательная точка",weakness:"Опустошение",advice:"Исцеляйтесь"}
];
function render(){
let h="",l=null;
for(let z of zones){
if(z.level!==l){
if(l!==null)h+="</div>";
l=z.level;
let n=l===7?"🔴 Уровень 7: Самоактуализация":l===6?"🟠 Уровень 6: Гармония":l===5?"🟡 Уровень 5: Любовь и принятие":l===4?"🟢 Уровень 4: Стабильность":l===3?"🔵 Уровень 3: Нейтральность":l===2?"🟣 Уровень 2: Напряжение":l===1?"⚫ Уровень 1: Конфликт":"⚪ Уровень 0: Антагонизм";
h+="<h2>"+n+"</h2><div>";
}
h+='<div class="zone-card"><div class="zone-header" onclick="this.nextElementSibling.classList.toggle(\'show\')"><div class="emoji">'+z.emoji+'</div><div class="range">'+z.range+'</div><div class="name">'+z.name+'</div><div class="toggle">▼</div></div><div class="zone-detail"><div class="zone-desc">'+z.desc+'</div><div class="zone-interpret"><p><strong>💪 Сильная сторона:</strong> '+z.strength+'</p><p><strong>⚠️ Слабая сторона:</strong> '+z.weakness+'</p><p><strong>💡 Рекомендация:</strong> '+z.advice+'</p></div></div></div>';
}
h+="</div>";
document.getElementById("zones").innerHTML=h;
}
render();
</script></body></html>`))
}

func (s *Server) helpPageHandler(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte(`<!DOCTYPE html>
<html><head><meta charset="UTF-8"><title>Справка</title>
<style>body{font-family:system-ui;background:linear-gradient(135deg,#667eea,#764ba2);padding:20px}
.navbar{background:rgba(0,0,0,0.3);padding:15px;display:flex;gap:20px;border-radius:15px;margin-bottom:20px}
.navbar a{color:#fff;text-decoration:none;padding:8px 16px;border-radius:8px}
.navbar a:hover{background:rgba(255,255,255,0.2)}
.container{max-width:1000px;margin:0 auto;background:#fff;border-radius:20px;padding:30px}
h1,h2{color:#667eea}
.zone-list{display:grid;grid-template-columns:repeat(auto-fit,minmax(300px,1fr));gap:15px;margin:20px 0}
.zone-category{background:#f5f5f5;padding:15px;border-radius:10px}
</style></head><body>
<div class="navbar"><a href="/">🏠 Главная</a><a href="/matrix">📊 Матрица</a><a href="/biorhythms">📈 Биоритмы</a><a href="/timeline">📉 Динамика</a><a href="/forecast">🔮 Прогноз</a><a href="/zones">📖 31 зона</a><a href="/help" class="active">❓ Справка</a></div>
<div class="container"><h1>❓ Справка</h1>
<h2>31 зона психологической классификации</h2>
<p>7 уровней сознания:</p>
<div class="zone-list">
<div class="zone-category"><h3>🔴 Уровень 7: Самоактуализация (r>0.95)</h3><ul><li>r>0.98 — Симбиоз</li><li>r>0.96 — Трансцендентность</li><li>r>0.95 — Самоактуализация</li></ul></div>
<div class="zone-category"><h3>🟠 Уровень 6: Гармония (0.80-0.95)</h3><ul><li>r>0.90 — Космическая любовь</li><li>r>0.85 — Божественная гармония</li><li>r>0.80 — Глубокая привязанность</li></ul></div>
<div class="zone-category"><h3>🟡 Уровень 5: Любовь (0.60-0.80)</h3><ul><li>r>0.75 — Искренняя близость</li><li>r>0.70 — Дружеская симпатия</li><li>r>0.65 — Взаимопонимание</li><li>r>0.60 — Симпатия</li></ul></div>
<div class="zone-category"><h3>🟢 Уровень 4: Стабильность (0.30-0.60)</h3><ul><li>r>0.50 — Доброжелательность</li><li>r>0.45 — Нейтрально-позитивное</li><li>r>0.40 — Лёгкая симпатия</li><li>r>0.30 — Эмоциональный ноль</li></ul></div>
<div class="zone-category"><h3>🔵 Уровень 3: Нейтральность (0.00-0.30)</h3><ul><li>r>0.20 — Лёгкая отстранённость</li><li>r>0.10 — Наблюдатель</li><li>r>0.00 — Неопределённость</li></ul></div>
<div class="zone-category"><h3>🟣 Уровень 2: Напряжение (-0.30-0.00)</h3><ul><li>r>-0.10 — Лёгкое напряжение</li><li>r>-0.20 — Раздражение</li><li>r>-0.30 — Отчуждение</li></ul></div>
<div class="zone-category"><h3>⚫ Уровень 1: Конфликт (-0.60 - -0.30)</h3><ul><li>r>-0.40 — Напряжение</li><li>r>-0.50 — Конфронтация</li><li>r>-0.60 — Разрыв</li></ul></div>
<div class="zone-category"><h3>⚪ Уровень 0: Антагонизм (r<-0.60)</h3><ul><li>r>-0.70 — Враждебность</li><li>r>-0.80 — Антагонизм</li><li>r>-0.90 — Психологическое отторжение</li><li>r>-0.95 — Экзистенциальная несовместимость</li><li>r≤-0.95 — Тотальный разрыв</li></ul></div>
</div>
<h2>🎵 Гармоничность</h2><p>Оценка от 0 до 10, где 10 — идеальная близость к Φ=0.618</p>
<h2>🔮 Прогноз</h2><p>7 сфер: любовь, карьера, дружба, творчество, здоровье, обучение, финансы</p>
<h2>📈 Биоритмы</h2><p>Физический(23), эмоциональный(28), интеллектуальный(33), духовный(38), интуитивный(42)</p>
</div></body></html>`))
}
