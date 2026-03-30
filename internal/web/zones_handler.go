func (s *Server) zonesPageHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.Write([]byte(`<!DOCTYPE html>
<html><head><meta charset="UTF-8"><title>31 зона | Психометрический анализатор</title>
<style>
*{margin:0;padding:0;box-sizing:border-box}
body{font-family:system-ui;background:linear-gradient(135deg,#667eea,#764ba2);padding:20px}
.navbar{background:rgba(0,0,0,0.3);padding:15px 30px;display:flex;gap:20px;flex-wrap:wrap;border-radius:15px;margin-bottom:20px}
.navbar a{color:#fff;text-decoration:none;padding:8px 16px;border-radius:8px}
.navbar a:hover,.navbar a.active{background:rgba(255,255,255,0.2)}
.container{max-width:1200px;margin:0 auto;background:#fff;border-radius:20px;padding:30px}
h1{color:#667eea;margin-bottom:20px;border-bottom:3px solid #667eea;padding-bottom:10px}
h2{color:#667eea;margin:25px 0 15px 0}
.zone-card{background:#f9f9f9;border-radius:15px;margin-bottom:15px;overflow:hidden;transition:transform 0.2s}
.zone-card:hover{transform:translateX(5px)}
.zone-header{display:flex;align-items:center;gap:15px;padding:15px 20px;cursor:pointer;background:linear-gradient(135deg,#f5f7fa,#e8ecf1)}
.zone-header .emoji{font-size:32px}
.zone-header .range{font-family:monospace;font-weight:bold;min-width:100px}
.zone-header .name{flex:1;font-weight:bold;font-size:18px}
.zone-header .toggle{font-size:20px;color:#667eea}
.zone-detail{display:none;padding:20px;background:#fff;border-top:1px solid #eee}
.zone-detail.show{display:block}
.zone-desc{font-size:15px;line-height:1.5;margin-bottom:15px;color:#333}
.zone-interpret{background:#e8f5e9;padding:12px;border-radius:10px;margin-top:10px}
.zone-interpret p{margin:5px 0}
.zone-interpret strong{color:#2ecc71}
.back-link{display:inline-block;margin-top:30px;padding:10px 20px;background:linear-gradient(135deg,#667eea,#764ba2);color:#fff;text-decoration:none;border-radius:8px}
.back-link:hover{transform:translateY(-2px)}
.footer{background:#333;color:#fff;text-align:center;padding:20px;margin-top:30px;border-radius:15px}
@media (max-width: 768px){
    .zone-header{flex-wrap:wrap}
    .zone-header .range{order:3}
}
</style>
</head><body>
<div class="navbar">
<a href="/">🏠 Главная</a>
<a href="/matrix">📊 Матрица</a>
<a href="/biorhythms">📈 Биоритмы</a>
<a href="/timeline">📉 Динамика</a>
<a href="/forecast">🔮 Анализ</a>
<a href="/zones" class="active">📖 31 зона</a>
<a href="/help">❓ Справка</a>
</div>
<div class="container">
<h1>📖 31 зона психологической классификации</h1>
<p style="margin-bottom:20px">7 уровней сознания, 31 психологическое состояние. Чем выше уровень, тем глубже и осознаннее отношения. Золотое сечение (Φ = 0.618) отмечает точки идеальной гармонии.</p>

<div id="zones-container"></div>
<a href="/" class="back-link">← Вернуться на главную</a>
</div>
<div class="footer">
<p>31 зона — уникальная авторская классификация психологических состояний на основе корреляции Пирсона и золотого сечения.</p>
</div>

<script>
const zones = [
    // Уровень 7: Самоактуализация
    {level:7, emoji:"🔥", range:"r > 0.98", name:"СИМБИОЗ", 
     desc:"Полное слияние личностей, потеря границ. Вы чувствуете друг друга на расстоянии. Опасность — растворение собственного 'Я'.",
     strength:"Максимальное взаимопонимание, телепатическая связь", 
     weakness:"Потеря индивидуальности, созависимость",
     advice:"Сохраняйте личное пространство. Практикуйте раздельные активности."},
    {level:7, emoji:"💫", range:"r > 0.96", name:"ТРАНСЦЕНДЕНТНОСТЬ",
     desc:"Выход за пределы личности. Отношения становятся путём к чему-то большему.",
     strength:"Духовный рост через партнёрство", weakness:"Отрыв от реальности, бытовые проблемы игнорируются",
     advice:"Не забывайте о материальном мире. Баланс духовного и бытового."},
    {level:7, emoji:"🕊️", range:"r > 0.95", name:"САМОАКТУАЛИЗАЦИЯ",
     desc:"Полная реализация потенциала в паре. Каждый помогает другому стать лучшей версией себя.",
     strength:"Взаимное развитие, поддержка роста", weakness:"Высокие ожидания, риск разочарования",
     advice:"Цените маленькие шаги, не требуйте мгновенных результатов."},
    
    // Уровень 6: Гармония
    {level:6, emoji:"💎", range:"r > 0.90", name:"КОСМИЧЕСКАЯ ЛЮБОВЬ",
     desc:"Безусловное принятие. Любовь без условий и требований.",
     strength:"Безусловная любовь, глубокое принятие", weakness:"Риск жертвенности, потеря себя в другом",
     advice:"Любите, но не теряйте себя. Границы важны даже в безусловной любви."},
    {level:6, emoji:"💖", range:"r > 0.85", name:"БОЖЕСТВЕННАЯ ГАРМОНИЯ",
     desc:"Идеальный резонанс. Всё происходит естественно, без усилий.",
     strength:"Лёгкость отношений, естественность", weakness:"Иллюзия 'вечного счастья', шок при первом конфликте",
     advice:"Не бойтесь конфликтов. Они делают отношения живыми."},
    {level:6, emoji:"💞", range:"r > 0.80", name:"ГЛУБОКАЯ ПРИВЯЗАННОСТЬ",
     desc:"Душевное родство. Чувство, что вы знаете друг друга вечность.",
     strength:"Надёжность, предсказуемость, безопасность", weakness:"Рутина, потеря новизны",
     advice:"Вносите разнообразие. Удивляйте друг друга."},
    
    // Уровень 5: Любовь и принятие
    {level:5, emoji:"💛", range:"r > 0.75", name:"ИСКРЕННЯЯ БЛИЗОСТЬ",
     desc:"Тёплые доверительные отношения. Вы можете быть собой без маски.",
     strength:"Доверие, искренность, уязвимость без страха", weakness:"Риск эмоционального выгорания при излишней открытости",
     advice:"Делитесь чувствами, но не перегружайте партнёра."},
    {level:5, emoji:"🌸", range:"r > 0.70", name:"ДРУЖЕСКАЯ СИМПАТИЯ",
     desc:"Естественное притяжение. Приятно быть рядом, легко общаться.",
     strength:"Лёгкость, радость от общения, отсутствие напряжения", weakness:"Поверхностность, страх глубоких чувств",
     advice:"Позвольте себе быть уязвимым. Глубина стоит риска."},
    {level:5, emoji:"🤝", range:"r > 0.65", name:"ВЗАИМОПОНИМАНИЕ",
     desc:"Согласованность ценностей. Вы смотрите в одну сторону.",
     strength:"Общие цели, единые ценности, предсказуемость", weakness:"Риск 'скучности', отсутствие вызовов",
     advice:"Ставьте совместные цели. Двигайтесь к ним вместе."},
    {level:5, emoji:"🌱", range:"r > 0.60", name:"СИМПАТИЯ",
     desc:"Начало близости. Зарождается интерес и тёплые чувства.",
     strength:"Надежда, ожидание, предвкушение", weakness:"Хрупкость, легко разрушить неосторожным словом",
     advice:"Будьте бережны. Дайте отношениям время созреть."},
    
    // Уровень 4: Стабильность
    {level:4, emoji:"👋", range:"r > 0.50", name:"ДОБРОЖЕЛАТЕЛЬНОСТЬ",
     desc:"Открытость к контакту. Вы готовы к общению, но не ищете его активно.",
     strength:"Безопасность, отсутствие угрозы", weakness:"Пассивность, отсутствие инициативы",
     advice:"Делайте первый шаг. Инициатива не наказуема."},
    {level:4, emoji:"😐", range:"r > 0.45", name:"НЕЙТРАЛЬНО-ПОЗИТИВНОЕ",
     desc:"Комфортное сосуществование. Нет конфликтов, но нет и страсти.",
     strength:"Спокойствие, стабильность, предсказуемость", weakness:"Отсутствие развития, застой",
     advice:"Ищите точки роста. Комфорт не должен быть болотом."},
    {level:4, emoji:"📍", range:"r > 0.40", name:"ЛЁГКАЯ СИМПАТИЯ",
     desc:"Без обязательств. Приятно, но не более.",
     strength:"Свобода, отсутствие давления, игра", weakness:"Поверхностность, боязнь обязательств",
     advice:"Если чувства глубже — скажите. Если нет — не обманывайте ожиданий."},
    {level:4, emoji:"🧘", range:"r > 0.30", name:"ЭМОЦИОНАЛЬНЫЙ НОЛЬ",
     desc:"Спокойное равнодушие. Эмоции не включены.",
     strength:"Объективность, рациональность", weakness:"Отсутствие тепла, механистичность",
     advice:"Позвольте себе чувствовать. Эмоции — это топливо отношений."},
    
    // Уровень 3: Нейтральность
    {level:3, emoji:"🧊", range:"r > 0.20", name:"ЛЁГКАЯ ОТСТРАНЁННОСТЬ",
     desc:"Дипломатичность. Вы вежливы, но держите дистанцию.",
     strength:"Безопасность, контроль над ситуацией", weakness:"Холодность, отсутствие близости",
     advice:"Приоткройтесь. Дистанция спасает от боли, но лишает радости."},
    {level:3, emoji:"🤨", range:"r > 0.10", name:"НАБЛЮДАТЕЛЬ",
     desc:"Сторонний анализ. Вы смотрите со стороны, не включаясь.",
     strength:"Объективность, отсутствие предвзятости", weakness:"Отсутствие вовлечённости, холодность",
     advice:"Решитесь на шаг. Наблюдение не меняет реальность."},
    {level:3, emoji:"❓", range:"r > 0.00", name:"НЕОПРЕДЕЛЁННОСТЬ",
     desc:"Формирование отношения. Вы ещё не поняли, что чувствуете.",
     strength:"Открытость новому, отсутствие предрассудков", weakness:"Нестабильность, непредсказуемость",
     advice:"Дайте себе время. Не торопитесь с выводами."},
    
    // Уровень 2: Напряжение
    {level:2, emoji:"😌", range:"r > -0.10", name:"ЛЁГКОЕ НАПРЯЖЕНИЕ",
     desc:"Притирка. Вы ищете общий язык.",
     strength:"Динамика, развитие, преодоление", weakness:"Дискомфорт, неловкость",
     advice:"Терпение. Притирка — нормальный этап."},
    {level:2, emoji:"😤", range:"r > -0.20", name:"РАЗДРАЖЕНИЕ",
     desc:"Мелкие конфликты. Мелочи начинают бесить.",
     strength:"Честность, отсутствие лицемерия", weakness:"Излишняя реакция на мелочи, взрывчатость",
     advice:"Дышите. Считайте до десяти. Не всё стоит реакции."},
    {level:2, emoji:"🥀", range:"r > -0.30", name:"ОТЧУЖДЕНИЕ",
     desc:"Эмоциональная дистанция. Вы отдаляетесь.",
     strength:"Защита от боли, самосохранение", weakness:"Потеря связи, одиночество вдвоём",
     advice:"Поговорите. Молчание только увеличивает расстояние."},
    
    // Уровень 1: Конфликт
    {level:1, emoji:"⚡", range:"r > -0.40", name:"НАПРЯЖЕНИЕ",
     desc:"Постоянные трения. Каждый разговор требует усилий.",
     strength:"Прямота, отсутствие недомолвок", weakness:"Усталость от общения, желание избегать",
     advice:"Возьмите паузу. Отдых от общения может помочь."},
    {level:1, emoji:"🔥", range:"r > -0.50", name:"КОНФРОНТАЦИЯ",
     desc:"Открытое противостояние. Вы на разных сторонах баррикад.",
     strength:"Честность, ясность позиций", weakness:"Истощение, разрушение связей",
     advice:"Если не можете договориться — согласитесь не соглашаться."},
    {level:1, emoji:"💔", range:"r > -0.60", name:"РАЗРЫВ",
     desc:"Потеря эмоциональной связи. Вы больше не чувствуете друг друга.",
     strength:"Конец страданиям, начало исцеления", weakness:"Боль потери, сожаление",
     advice:"Примите. Иногда разрыв — это правильное решение."},
    
    // Уровень 0: Антагонизм
    {level:0, emoji:"🗡️", range:"r > -0.70", name:"ВРАЖДЕБНОСТЬ",
     desc:"Системный конфликт. Каждый шаг вызывает сопротивление.",
     strength:"Ясность — вы точно знаете, что не подходите", weakness:"Постоянная борьба, отсутствие покоя",
     advice:"Минимизируйте контакты. Не всё можно исправить."},
    {level:0, emoji:"💀", range:"r > -0.80", name:"АНТАГОНИЗМ",
     desc:"Непримиримое противостояние. Вы как огонь и вода.",
     strength:"Предсказуемость реакции", weakness:"Полная несовместимость, разрушительное влияние",
     advice:"Признайте несовместимость. Это не поражение, а осознание."},
    {level:0, emoji:"🌑", range:"r > -0.90", name:"ПСИХОЛОГИЧЕСКОЕ ОТТОРЖЕНИЕ",
     desc:"Аверсия. Вас буквально тошнит от присутствия друг друга.",
     strength:"Чёткая граница 'не мой человек'", weakness:"Интенсивный дискомфорт, избегание",
     advice:"Полное прекращение контакта — единственный выход."},
    {level:0, emoji:"🕳️", range:"r > -0.95", name:"ЭКЗИСТЕНЦИАЛЬНАЯ НЕСОВМЕСТИМОСТЬ",
     desc:"Полное неприятие на уровне бытия. Вы из разных вселенных.",
     strength:"Абсолютная ясность", weakness:"Невозможность какого-либо взаимодействия",
     advice:"Не пытайтесь. Некоторые вещи нельзя изменить."},
    {level:0, emoji:"🌌", range:"r ≤ -0.95", name:"ТОТАЛЬНЫЙ РАЗРЫВ",
     desc:"Энергетический вакуум. Даже мысль о человеке вызывает пустоту.",
     strength:"Окончательная точка", weakness:"Опустошение, потеря веры в людей",
     advice:"Исцеляйтесь. Мир не состоит из одного человека."}
];

function renderZones() {
    let html = '';
    let currentLevel = null;
    
    for(let z of zones) {
        if(z.level !== currentLevel) {
            if(currentLevel !== null) html+='</div>';
            currentLevel = z.level;
            let levelName = '';
            if(currentLevel===7) levelName = '🔴 Уровень 7: Самоактуализация';
            else if(currentLevel===6) levelName = '🟠 Уровень 6: Гармония';
            else if(currentLevel===5) levelName = '🟡 Уровень 5: Любовь и принятие';
            else if(currentLevel===4) levelName = '🟢 Уровень 4: Стабильность';
            else if(currentLevel===3) levelName = '🔵 Уровень 3: Нейтральность';
            else if(currentLevel===2) levelName = '🟣 Уровень 2: Напряжение';
            else if(currentLevel===1) levelName = '⚫ Уровень 1: Конфликт';
            else if(currentLevel===0) levelName = '⚪ Уровень 0: Антагонизм';
            html+='<h2>'+levelName+'</h2><div style="margin-bottom:20px">';
        }
        
        html+=`
        <div class="zone-card">
            <div class="zone-header" onclick="this.nextElementSibling.classList.toggle('show')">
                <div class="emoji">${z.emoji}</div>
                <div class="range">${z.range}</div>
                <div class="name">${z.name}</div>
                <div class="toggle">▼</div>
            </div>
            <div class="zone-detail">
                <div class="zone-desc">${z.desc}</div>
                <div class="zone-interpret">
                    <p><strong>💪 Сильная сторона:</strong> ${z.strength}</p>
                    <p><strong>⚠️ Слабая сторона:</strong> ${z.weakness}</p>
                    <p><strong>💡 Рекомендация:</strong> ${z.advice}</p>
                </div>
            </div>
        </div>`;
    }
    html+='</div>';
    document.getElementById('zones-container').innerHTML = html;
}
renderZones();
</script>
</body></html>`))
}
