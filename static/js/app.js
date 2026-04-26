const API_BASE = window.location.origin;
let token = localStorage.getItem('token');
let currentRoundID = null;
let selectedHeroID = null;


window.onload = () => {
    const path = window.location.pathname;

   
    if (!token && !path.includes('login')) {
        window.location.href = '/login';
        return;
    }

    if (path.includes('login')) {

        if (token) window.location.href = '/';
    } else if (path.includes('game')) {
       
        loadGame();
    } else {
       
        loadStats();
        loadTopPlayers();
    }
};


async function handleLogin() {
    const username = document.getElementById('username').value;
    const password = document.getElementById('password').value;
    const msgEl = document.getElementById('auth-message');
    
    if(!username || !password) {
        msgEl.textContent = 'Enter username and password';
        return;
    }

    try {
        const res = await fetch(`${API_BASE}/login`, {
            method: 'POST',
            headers: { 'Content-Type': 'application/json' },
            body: JSON.stringify({ username, password })
        });
        
        if (!res.ok) {
            msgEl.textContent = await res.text();
            return;
        }
        
        const data = await res.json();
        token = data.token;
        localStorage.setItem('token', token);
        window.location.href = '/'; 
    } catch(e) {
        msgEl.textContent = 'Error connecting to server';
    }
}

async function handleRegister() {
    const username = document.getElementById('username').value;
    const password = document.getElementById('password').value;
    const msgEl = document.getElementById('auth-message');
    
    if(!username || !password) {
        msgEl.textContent = 'Enter username and password';
        return;
    }

    try {
        const res = await fetch(`${API_BASE}/register`, {
            method: 'POST',
            headers: { 'Content-Type': 'application/json' },
            body: JSON.stringify({ username, password })
        });
        
        if (!res.ok) {
            msgEl.textContent = await res.text();
            return;
        }
        
        const data = await res.json();
        msgEl.style.color = 'var(--md-sys-color-primary)';
        msgEl.textContent = data.message || 'Epic! Now log in.';
        setTimeout(() => msgEl.style.color = 'var(--md-sys-color-error)', 3000);
    } catch(e) {
        msgEl.textContent = 'Error connecting to server';
    }
}

function handleLogout() {
    localStorage.removeItem('token');
    token = null;
    window.location.href = '/login'; 
}


async function loadStats() {
    try {
        const res = await fetch(`${API_BASE}/stats`, {
            headers: { 'Authorization': `Bearer ${token}` }
        });
        if(res.status === 401) { handleLogout(); return; }
        const stats = await res.json();
        document.getElementById('user-stats').innerHTML = `
            <div style="font-size: 18px; margin-bottom: 12px;">Player: <span style="color: var(--md-sys-color-on-background)">${stats.name}</span></div>
            <div style="display: flex; gap: 24px; justify-content: center;">
                <div>Wins: <br><span class="stats-value" style="color: #69f0ae">${stats.wins}</span></div>
                <div>Losses: <br><span class="stats-value" style="color: #ff5252">${stats.losses}</span></div>
            </div>
        `;
    } catch(e) {
        const el = document.getElementById('user-stats');
        if(el) el.innerHTML = '<span style="color: var(--md-sys-color-error)">Error loading stats</span>';
    }
}

async function loadTopPlayers() {
    try {
        const res = await fetch(`${API_BASE}/top-players`);
        if (!res.ok) throw new Error();
        const players = await res.json();
        const container = document.getElementById('top-players');
        
        if (!container) return;

        if(players.length === 0) {
            container.innerHTML = 'No players yet';
            return;
        }

        container.innerHTML = players.map((p, i) => `
            <div class="player-row">
                <span><strong>#${i+1}</strong> ${p.username}</span>
                <span style="color: var(--md-sys-color-primary)">${p.wins} wins</span>
            </div>
        `).join('');
    } catch(e) {
        const el = document.getElementById('top-players');
        if(el) el.textContent = 'Error loading top players';
    }
}

function startGame() {
    window.location.href = '/game'; 
}

function goHome() {
    window.location.href = '/'; 
}


async function loadGame() {
    try {
        const res = await fetch(`${API_BASE}/round`, {
            headers: { 'Authorization': `Bearer ${token}` }
        });
        if(res.status === 401) { handleLogout(); return; }
        const game = await res.json();
        
        currentRoundID = game.round_id;
        document.getElementById('game-message').textContent = game.message || '';
        
        displayItems(game.main_items, 'items');
        
        if (game.backpack_items && game.backpack_items !== "[]" && game.backpack_items.length > 0) {
            displayItems(game.backpack_items, 'backpack');
        } else {
            document.getElementById('backpack').innerHTML = '<span style="color: var(--md-sys-color-on-surface-variant); font-size: 14px;">Empty</span>';
        }
        
        displayHints(game);

        if (game.status && game.status !== 'playing') {
            showEndScreen(game.status, game.correct_hero);
        } else {
            hideEndScreen();
        }

        loadHeroes(game.hero_attribute);
        
        selectedHeroID = null;
        document.getElementById('submit').disabled = true;

    } catch(e) {
        console.error(e);
        const msgEl = document.getElementById('game-message');
        if(msgEl) msgEl.textContent = 'Error loading game';
    }
}

function displayItems(items, containerID) {
    const container = document.getElementById(containerID);
    if (!container) return;
    container.innerHTML = '';
    if(!items || items.length === 0) return;
    
    items.forEach(item => {
        const img = document.createElement('img');
        img.src = item.image_url;
        img.className = 'item-img';
        img.title = item.name;
        container.appendChild(img);
    });
}

function displayHints(game) {
    const hints = document.getElementById('hints');
    if (!hints) return;
    hints.innerHTML = '';
    
    if (game.is_won !== undefined) {
        const wonText = game.is_won ? 'Win' : 'Loss';
        const color = game.is_won ? '#69f0ae' : '#ff5252';
        hints.innerHTML += `<div class="hint-chip" style="border-left: 3px solid ${color}">Dota Match Outcome: ${wonText}</div>`;
    }
    if (game.hero_attribute) {
        const attrMap = { 'str': 'Strength', 'agi': 'Agility', 'int': 'Intelligence', 'all': 'Universal' };
        const attrName = attrMap[game.hero_attribute] || game.hero_attribute;
        hints.innerHTML += `<div class="hint-chip">Attribute: ${attrName}</div>`;
    }
}

async function loadHeroes(attributeHint) {
    try {
        const res = await fetch(`${API_BASE}/heroes`);
        if (!res.ok) throw new Error();
        const heroes = await res.json();
        
        const groups = { str: [], agi: [], int: [], all: [] };
        heroes.forEach(h => { if(groups[h.type]) groups[h.type].push(h); });
        
        Object.keys(groups).forEach(key => {
            groups[key].sort((a, b) => a.name.localeCompare(b.name));
        });

        renderHeroGroup('group-str', 'strength-heroes', groups.str, attributeHint === 'str' || !attributeHint);
        renderHeroGroup('group-agi', 'agility-heroes', groups.agi, attributeHint === 'agi' || !attributeHint);
        renderHeroGroup('group-int', 'intelligence-heroes', groups.int, attributeHint === 'int' || !attributeHint);
        renderHeroGroup('group-uni', 'universal-heroes', groups.all, attributeHint === 'all' || !attributeHint);
        
    } catch(e) {
        console.error('Heroes loading error', e);
    }
}

function renderHeroGroup(cardId, listId, heroes, show) {
    const card = document.getElementById(cardId);
    const list = document.getElementById(listId);
    if (!card || !list) return;
    list.innerHTML = '';
    
    if (!show || heroes.length === 0) {
        card.style.display = 'none';
        return;
    }
    
    card.style.display = 'block';
    heroes.forEach(hero => {
        const img = document.createElement('img');
        img.src = hero.image_url;
        img.className = 'hero-img';
        img.title = hero.name;
        img.onclick = () => selectHero(hero.id, img);
        list.appendChild(img);
    });
}

function selectHero(id, imgElement) {
    document.querySelectorAll('.hero-img').forEach(el => el.classList.remove('selected'));
    imgElement.classList.add('selected');
    selectedHeroID = id;
    document.getElementById('submit').disabled = false;
}

async function submitGuess() {
    if(!selectedHeroID) return;
    
    try {
        const res = await fetch(`${API_BASE}/round`, {
            method: 'POST',
            headers: {
                'Authorization': `Bearer ${token}`,
                'Content-Type': 'application/json'
            },
            body: JSON.stringify({ round_id: currentRoundID, guess_id: selectedHeroID })
        });
        
        const game = await res.json();
        if (!res.ok) {
            alert('Error submitting: ' + (game.message || 'Unknown error'));
            return;
        }

        if (game.status && game.status !== 'playing') {
            showEndScreen(game.status, game.correct_hero);
        } else {
            loadGame();
        }
    } catch(e) {
        console.error('Error submitting result', e);
    }
}

function showEndScreen(status, correctHero) {
    const overlay = document.getElementById('end-screen');
    const message = document.getElementById('end-message');
    const icon = document.getElementById('end-icon');
    
    if (!overlay) return;

    if (status === 'won') {
        icon.innerHTML = '<span class="material-icon" style="color: #69f0ae; font-size: 64px;">verified</span>';
        message.innerHTML = `You won!<br><span style="color: var(--md-sys-color-primary); font-size: 20px; font-weight: normal;">Correct answer: ${correctHero || ''}</span>`;
    } else {
        icon.innerHTML = '<span class="material-icon" style="color: #ff5252; font-size: 64px;">cancel</span>';
        message.innerHTML = `You were wrong!<br><span style="color: var(--md-sys-color-primary); font-size: 20px; font-weight: normal;">The hero was: ${correctHero || ''}</span>`;
    }
    
    overlay.classList.add('visible');
}

function hideEndScreen() {
    const overlay = document.getElementById('end-screen');
    if (overlay) overlay.classList.remove('visible');
}

function startNewRound() {
    hideEndScreen();
    loadGame(); 
}
