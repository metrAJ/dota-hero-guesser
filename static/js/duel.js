const API_BASE = window.location.origin;
const token = localStorage.getItem('token');

let duelSocket = null;
let selectedHeroID = null;
let allHeroesList = []; 

let globalP1Wrong = []; 
let globalP2Wrong = [];
let myPlayerName = null;   // <--- Changed to Name
let opponentName = null;


window.onload = async () => {
    if (!token) {
        window.location.href = '/login';
        return;
    }
    
    // 1. FORCE CSS OVERRIDE: Guarantee the Queue shows and the Game is hidden on load
    const queue = document.getElementById('queue-view');
    const game = document.getElementById('game-view');
    if (queue) queue.style.display = 'flex';
    if (game) game.style.display = 'none'; 

    await fetchHeroes();
    startMatchmaking();
};

// --- 1. HERO UI LOGIC (Bulletproof & Fuzzy Matching) ---
async function fetchHeroes() {
    try {
        const res = await fetch(`${API_BASE}/heroes`);
        const heroes = await res.json();
        
        allHeroesList = heroes; 
        console.log(`📡 Database sent ${heroes.length} heroes.`); // DEBUG LOG

        // Get exactly the containers we need
        const containers = {
            'Strength': document.getElementById('strength-heroes'),
            'Agility': document.getElementById('agility-heroes'),
            'Intelligence': document.getElementById('intelligence-heroes'),
            'Universal': document.getElementById('universal-heroes')
        };

        // SAFETY CHECK: Are the HTML elements actually on the page?
        if (!containers['Strength'] || !containers['Agility']) {
            console.error("❌ HTML ERROR: JavaScript cannot find <div id='strength-heroes'> in your duel.html! Did it get deleted?");
            return;
        }

        Object.values(containers).forEach(c => { if(c) c.innerHTML = ''; });

        let heroCount = 0;
        heroes.forEach(h => {
            const img = document.createElement('img');
            img.src = h.image_url || h.ImageURL || h.image || `/static/images/heroes/${h.id || h.ID}.png`; 
            img.className = 'hero-img';
            img.id = `hero-img-${h.id}`;
            img.onclick = () => selectHero(h.id);
            
            // Try Fuzzy Match First
            const attr = String(h.type || "").toLowerCase();
            let targetId = '';
            if (attr.includes('str')) targetId = 'strength-heroes';
            else if (attr.includes('agi')) targetId = 'agility-heroes';
            else if (attr.includes('int')) targetId = 'intelligence-heroes';
            else if (attr.includes('uni') || attr.includes('all')) targetId = 'universal-heroes';
            
            if (targetId && document.getElementById(targetId)) {
                document.getElementById(targetId).appendChild(img);
                heroCount++;
            } else if (containers[h.type]) {
                // Strict Match Fallback
                containers[h.type].appendChild(img);
                heroCount++;
            } else {
                console.warn(`Could not figure out where to put hero: ${h.name || h.id} with type: ${h.type}`);
            }
        });
        console.log(`✅ Successfully loaded ${heroCount} heroes into the DOM.`);
    } catch (e) {
        console.error("❌ Failed to fetch heroes:", e);
    }
}

function getHeroImageUrl(id) {
    const hero = allHeroesList.find(h => h.id === parseInt(id));
    return hero ? (hero.image_url || hero.ImageURL || '') : '';
}

function selectHero(id) {
    document.querySelectorAll('.hero-img').forEach(img => img.classList.remove('selected'));
    const img = document.getElementById(`hero-img-${id}`);
    
    if (img && !img.classList.contains('eliminated')) {
        img.classList.add('selected');
        selectedHeroID = id;
        const submitBtn = document.getElementById('submit');
        if (submitBtn) submitBtn.disabled = false;
    }
}

function cancelMatchmaking() {
    if (duelSocket) {
        duelSocket.close(); 
        duelSocket = null;
    }
    window.location.href = '/'; 
}

// --- 2. WEBSOCKET LOGIC ---
async function startMatchmaking() {
    try {
        const res = await fetch(`${API_BASE}/ws-ticket`, {
            method: 'POST',
            headers: { 'Authorization': `Bearer ${token}` }
        });
        
        if (!res.ok) throw new Error("Ticket request failed");
        
        const data = await res.json();

        const wsProtocol = window.location.protocol === 'https:' ? 'wss://' : 'ws://';
        const WS_BASE = wsProtocol + window.location.host;
        duelSocket = new WebSocket(`${WS_BASE}/ws?ticket=${data.ticket}`);

        duelSocket.onmessage = (event) => {
            try {
                const rawData = JSON.parse(event.data);
                console.log("📩 Raw WebSocket Message:", rawData);
                handleServerMessage(rawData);
            } catch (err) {
                console.error("❌ Error parsing message:", err);
            }
        };
        
        duelSocket.onclose = () => {
            const endScreen = document.getElementById('end-screen');
            if (!endScreen || !endScreen.classList.contains('active')) {
                console.warn("Socket closed.");
            }
        };
    } catch (e) {
        console.error(e);
        alert("Помилка підключення до сервера.");
        window.location.href = '/';
    }
}

function submitDuelGuess() {
    if (!selectedHeroID || !duelSocket || duelSocket.readyState !== WebSocket.OPEN) return;
    
    duelSocket.send(JSON.stringify({
        type: "guess",
        hero_id: selectedHeroID
    }));
}

// --- 3. THE SAFE EVENT HANDLER ---
function handleServerMessage(msg) {
    const msgType = String(msg.type || msg.Type || "").toLowerCase().trim();
    const payload = msg.payload || msg.Payload || {};

    const queueEl = document.getElementById('queue-view');
    const gameEl = document.getElementById('game-view');
    const msgEl = document.getElementById('game-message');

    switch (msgType) {
        case "waiting":
            console.log("Waiting for opponent...");
            break;

        case "match_found":
            globalP1Wrong = [];
            globalP2Wrong = [];
            myPlayerName = payload.my_name;
            opponentName = payload.opponent_name;

            const myNameEl = document.getElementById('my-name-display');
            const oppNameEl = document.getElementById('opponent-name-display');

            if (myNameEl) myNameEl.innerText = myPlayerName || "You";
            if (oppNameEl) oppNameEl.innerText = opponentName || "Opponent";


            console.log("✅ Match Found! Forcing screen swap...");
            if (queueEl) {
                queueEl.classList.remove('active');
                queueEl.style.display = 'none'; 
            }
            if (gameEl) {
                gameEl.classList.add('active');
                gameEl.style.display = 'block'; 
            }
            
            let startMsg = "Opponent found! Get ready...";
            if (msgEl) msgEl.innerHTML = startMsg;
            break;

        case "phase_start":
            let roundHtml = `<div style="font-size: 18px; font-weight: bold; color: var(--md-sys-color-primary);">Round ${payload.round || 1} - You have ${payload.time_limit || 30} seconds!</div>`;
            
            // 1. DYNAMIC ITEM RENDERING: Extract and draw Main Items
            if (payload.main_items && Array.isArray(payload.main_items) && payload.main_items.length > 0) {
                roundHtml += renderItemsHtml(payload.main_items);
            }
            
            // 2. DYNAMIC ITEM RENDERING: Extract and draw Backpack Items
            if (payload.backpack_items && Array.isArray(payload.backpack_items) && payload.backpack_items.length > 0) {
                roundHtml += `<div style="margin-top: 12px; font-size: 14px; color: var(--md-sys-color-on-surface-variant);">Рюкзак:</div>`;
                roundHtml += renderItemsHtml(payload.backpack_items);
            }
            
            // 3. HINT RENDERING: Extract and draw Hero Attribute
            if (payload.hero_attribute) {
               const attrNames = { 
               "str": "Strength", 
               "agi": "Agility", 
               "int": "Intelligence", 
               "all": "Universal" 
            };
            const attrName = attrNames[payload.hero_attribute] || payload.hero_attribute;  
            roundHtml += `<div style="margin-top: 12px; font-size: 16px; font-weight: bold; color: #ffeb3b;">Hint: This hero is ${attrName}</div>`;
        
             // Hide all heroes except the hinted category
            filterHeroesByAttribute(payload.hero_attribute);
            } else {
            // If there is no hint (e.g., Rounds 1-4 or a new game), reset and show all heroes
                 filterHeroesByAttribute("reset");
            }

            if (msgEl) msgEl.innerHTML = roundHtml;
            
            const btn = document.getElementById('submit');
            if (btn) btn.disabled = true; 
            
            selectedHeroID = null;
            document.querySelectorAll('.hero-img').forEach(img => img.classList.remove('selected'));
            
            applyEliminatedHeroes(payload);
            break;

        case "guess_locked":
            if (msgEl) msgEl.innerHTML = "Answer received. Waiting for opponent...";
            const sBtn = document.getElementById('submit');
            if (sBtn) sBtn.disabled = true;
            document.querySelectorAll('.hero-img').forEach(img => img.style.pointerEvents = 'none');
            break;

        case "phase_results":
            applyEliminatedHeroes(payload);
            break;

       case "game_over": {
            console.log("Game Over triggered! Winner:", payload.winner_name);
            
            // 1. FORCE THE GAME BOARD TO HIDE so it cannot block the overlay
            const gameBoard = document.getElementById('game-view');
            if (gameBoard) gameBoard.style.display = 'none';

            let statusText = "";
            
            if (payload.reason === "opponent_disconnected") {
                statusText = "You won! Opponent disconnected.";
            } else if (payload.winner_name === myPlayerName) {
                statusText = "You won! You guessed the hero.";
            } else if (payload.winner_name) {
                statusText = `You lost! ${payload.winner_name} guessed first.`;
            } else {
                statusText = "Match completed!";
            }
            
            showEndScreen(statusText, payload.hero_id);
            if (duelSocket) duelSocket.close();
            break;
        }

        case "error":
            if (msgEl) msgEl.innerHTML = `<span style="color: #ff5252;">${payload}</span>`;
            break;
    }
}

// HELPER: Bulletproof Item Renderer
function renderItemsHtml(itemsArray) {
    let html = `<div style="display: flex; justify-content: center; gap: 8px; margin-top: 12px;">`;
    itemsArray.forEach(item => {
        let imgUrl = "";
        if (typeof item === 'object' && item !== null) {
            imgUrl = item.image_url || item.ImageURL || item.image || `/static/images/items/${item.id || item.ID}.png`;
        } else {
            imgUrl = `/static/images/items/${item}.png`;
        }
        html += `<img src="${imgUrl}" style="width: 48px; border-radius: 4px; box-shadow: 0 2px 4px rgba(0,0,0,0.5);">`;
    });
    html += `</div>`;
    return html;
}

function applyEliminatedHeroes(payload) {
    // 1. Ask the payload for the array tied to YOUR exact name and THEIR exact name!
    if (payload.wrong_guesses && Object.keys(payload.wrong_guesses).length > 0) {
        if (myPlayerName) globalP1Wrong = payload.wrong_guesses[myPlayerName] || [];
        if (opponentName) globalP2Wrong = payload.wrong_guesses[opponentName] || [];
    } 

    // 2. Combine for graying out the center board
    const allEliminated = [...globalP1Wrong, ...globalP2Wrong];
    
    document.querySelectorAll('.hero-img').forEach(img => {
        img.style.pointerEvents = 'auto'; 
        img.classList.remove('eliminated');
        
        const heroId = parseInt(img.id.replace('hero-img-', ''));
        if (allEliminated.includes(heroId)) {
            img.classList.add('eliminated');
            img.style.pointerEvents = 'none'; 
        }
    });

    // 3. Draw Left (You) and Right (Them)
    const p1Container = document.getElementById('p1-wrong-guesses');
    const p2Container = document.getElementById('p2-wrong-guesses');
    
    if (p1Container) p1Container.innerHTML = globalP1Wrong.map(id => `<img src="${getHeroImageUrl(id)}" style="width:100%; border-radius:4px; border: 2px solid #ff5252;">`).join('');
    if (p2Container) p2Container.innerHTML = globalP2Wrong.map(id => `<img src="${getHeroImageUrl(id)}" style="width:100%; border-radius:4px; border: 2px solid #ff5252;">`).join('');
}

function showEndScreen(status, correctHeroId) {
    try {
        const overlay = document.getElementById('end-screen');
        const message = document.getElementById('end-message');
        const icon = document.getElementById('end-icon');
        
        if (!overlay) {
            console.error("❌ end-screen div missing from HTML!");
            return;
        }
        
        // 🚨 OVERRIDE ANY !IMPORTANT CSS RULES PREVENTING DISPLAY 🚨
        overlay.style.setProperty('position', 'fixed', 'important');
        overlay.style.setProperty('top', '0', 'important');
        overlay.style.setProperty('left', '0', 'important');
        overlay.style.setProperty('width', '100vw', 'important');
        overlay.style.setProperty('height', '100vh', 'important');
        overlay.style.setProperty('background-color', 'rgba(0, 0, 0, 0.95)', 'important');
        overlay.style.setProperty('z-index', '999999', 'important');
        overlay.style.setProperty('display', 'flex', 'important');
        overlay.style.setProperty('flex-direction', 'column', 'important');
        overlay.style.setProperty('justify-content', 'center', 'important');
        overlay.style.setProperty('align-items', 'center', 'important');
        overlay.style.setProperty('opacity', '1', 'important');
        overlay.style.setProperty('visibility', 'visible', 'important');
        overlay.classList.add('active');

        // Force child container to be visible
        const endContent = overlay.querySelector('.end-content');
        if (endContent) {
            endContent.style.setProperty('display', 'block', 'important');
            endContent.style.setProperty('opacity', '1', 'important');
            endContent.style.setProperty('visibility', 'visible', 'important');
            overlay.style.pointerEvents = 'auto';
        }

        const heroImgUrl = getHeroImageUrl(correctHeroId);

        if (status.includes("Перемога") || status.includes("Win") || status === 'won') {
            if (icon) icon.innerHTML = `<img src="${heroImgUrl}" style="width: 120px; border-radius: 8px; border: 4px solid #69f0ae; box-shadow: 0 0 40px rgba(105, 240, 174, 0.8);">`;
            if (message) message.innerHTML = `<span style="color: #69f0ae; font-size: 32px; font-weight: bold; text-shadow: 0 2px 4px rgba(0,0,0,0.8);">${status}</span><br><span style="font-size: 16px; margin-top: 12px; display: block; color: white;">Правильна відповідь</span>`;
        } else {
            if (icon) icon.innerHTML = `<img src="${heroImgUrl}" style="width: 120px; border-radius: 8px; border: 4px solid #ff5252; filter: grayscale(100%); box-shadow: 0 0 40px rgba(255, 82, 82, 0.8);">`;
            if (message) message.innerHTML = `<span style="color: #ff5252; font-size: 32px; font-weight: bold; text-shadow: 0 2px 4px rgba(0,0,0,0.8);">${status}</span><br><span style="font-size: 16px; margin-top: 12px; display: block; color: white;">Герой був:</span>`;
        }
    } catch (err) {
        console.error("❌ Crash inside showEndScreen:", err);
    }
}


function filterHeroesByAttribute(targetAttr) {
    // Map the server's attribute strings to your exact DOM element IDs
    const categories = {
        'str': 'strength-heroes',
        'agi': 'agility-heroes',
        'int': 'intelligence-heroes',
        'all': 'universal-heroes' 
    };

    const showAll = (targetAttr === "reset");

    for (let attrKey in categories) {
        const container = document.getElementById(categories[attrKey]);
        if (container) {
            const shouldShow = showAll || (attrKey === targetAttr);
            
            // Using "" restores your default CSS (like display: grid), "none" hides it
            container.style.display = shouldShow ? "" : "none";
            
            // Grab the Header (h2/h3) right above the container and hide it too
            const header = container.previousElementSibling;
            if (header && header.tagName.match(/^H\d$/)) {
                header.style.display = shouldShow ? "" : "none";
            }
        }
    }
}