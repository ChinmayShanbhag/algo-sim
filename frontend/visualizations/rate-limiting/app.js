// Theme management (initialize immediately)
(function() {
    const savedTheme = localStorage.getItem('theme') || 'dark';
    document.documentElement.setAttribute('data-theme', savedTheme);
})();

const API_BASE = 'http://localhost:8080/api/rate-limiting';

let allStates = null;

// Track previous values for drop animations
let previousTokens = null;
let previousQueue = null;

// Initialize visualization
function initVisualization() {
    console.log('🔧 Initializing Rate Limiting visualization...');
    loadStates();
}

function loadStates() {
    fetch(`${API_BASE}/state`)
        .then(response => response.json())
        .then(data => {
            allStates = data;
            renderAllAlgorithms();
        })
        .catch(error => {
            console.error('Error loading states:', error);
            showFeedback('Error connecting to backend. Make sure it\'s running on port 8080.', 'error');
        });
}

function renderAllAlgorithms() {
    if (!allStates) return;
    
    // Fixed Window
    renderFixedWindow(allStates.fixedWindow);
    
    // Sliding Log
    renderSlidingLog(allStates.slidingLog);
    
    // Sliding Window
    renderSlidingWindow(allStates.slidingWindow);
    
    // Token Bucket
    renderTokenBucket(allStates.tokenBucket);
    
    // Leaky Bucket
    renderLeakyBucket(allStates.leakyBucket);
}

function renderFixedWindow(state) {
    const countEl = document.getElementById('fw-count');
    const barEl = document.getElementById('fw-bar');
    const historyEl = document.getElementById('fw-history');
    
    if (countEl) countEl.textContent = state.currentCount;
    
    // Calculate window progress
    if (barEl && state.windowStart && state.windowEnd) {
        const now = new Date();
        const start = new Date(state.windowStart);
        const end = new Date(state.windowEnd);
        const progress = Math.min(100, ((now - start) / (end - start)) * 100);
        barEl.style.width = progress + '%';
    }
    
    renderHistory(historyEl, state.requestHistory);
}

function renderSlidingLog(state) {
    const countEl = document.getElementById('sl-count');
    const timelineEl = document.getElementById('sl-timeline-entries');
    const historyEl = document.getElementById('sl-history');
    
    // Calculate actual count of requests within the window
    const now = new Date();
    const windowSize = state.windowSize * 1000; // Convert to ms
    let actualCount = 0;
    
    // Show log entries on timeline
    if (timelineEl && state.requestLog) {
        timelineEl.innerHTML = '';
        
        state.requestLog.forEach(timestamp => {
            const requestTime = new Date(timestamp);
            const age = now - requestTime;
            
            // Only show if within window
            if (age <= windowSize) {
                actualCount++; // Count visible requests
                
                const dot = document.createElement('div');
                dot.className = 'log-entry-timeline';
                
                // Position on timeline (0% = -60s ago, 100% = now)
                const position = 100 - ((age / windowSize) * 100);
                dot.style.left = position + '%';
                dot.style.top = '50%';
                
                // Add age class for visual effect
                if (age > windowSize * 0.7) {
                    dot.classList.add('old');
                }
                
                // Tooltip
                const secondsAgo = Math.floor(age / 1000);
                dot.title = `${secondsAgo}s ago`;
                
                timelineEl.appendChild(dot);
            }
        });
    }
    
    // Update count with actual visible count
    if (countEl) countEl.textContent = actualCount;
    
    renderHistory(historyEl, state.requestHistory);
}

function renderSlidingWindow(state) {
    const countEl = document.getElementById('sw-count');
    const prevCountEl = document.getElementById('sw-prev-count');
    const currCountEl = document.getElementById('sw-curr-count');
    const historyEl = document.getElementById('sw-history');
    
    if (countEl) countEl.textContent = state.estimatedCount.toFixed(1);
    if (prevCountEl) prevCountEl.textContent = state.prevCount;
    if (currCountEl) currCountEl.textContent = state.currentCount;
    
    renderHistory(historyEl, state.requestHistory);
}

function renderTokenBucket(state) {
    const tokensEl = document.getElementById('tb-tokens');
    const fillEl = document.getElementById('tb-fill');
    const historyEl = document.getElementById('tb-history');
    
    // Calculate current tokens accounting for time passed since last refill
    const now = new Date();
    const lastRefill = new Date(state.lastRefill);
    const elapsedSeconds = (now - lastRefill) / 1000;
    const currentTokens = Math.min(state.capacity, state.currentTokens + (elapsedSeconds * state.refillRate));
    
    // Animate drops falling IN when tokens increase
    if (previousTokens !== null && currentTokens > previousTokens) {
        const tokensGained = Math.floor(currentTokens - previousTokens);
        for (let i = 0; i < Math.min(tokensGained, 3); i++) {
            setTimeout(() => createDropAnimation('token-bucket', 'falling-in'), i * 150);
        }
    }
    previousTokens = currentTokens;
    
    if (tokensEl) tokensEl.textContent = currentTokens.toFixed(1);
    
    // Show bucket fill level
    if (fillEl) {
        const fillPercent = (currentTokens / state.capacity) * 100;
        fillEl.style.height = fillPercent + '%';
    }
    
    renderHistory(historyEl, state.requestHistory);
}

function renderLeakyBucket(state) {
    const queueEl = document.getElementById('lb-queue');
    const fillEl = document.getElementById('lb-fill');
    const historyEl = document.getElementById('lb-history');
    
    // Calculate current queue accounting for processing since last update
    const now = new Date();
    const lastProcess = new Date(state.lastProcess);
    const elapsedSeconds = (now - lastProcess) / 1000;
    const processed = Math.floor(elapsedSeconds * state.processRate);
    const currentQueue = Math.max(0, state.currentQueue - processed);
    
    // Animate drops falling OUT when queue decreases
    if (previousQueue !== null && currentQueue < previousQueue) {
        const dropsLeaked = Math.floor(previousQueue - currentQueue);
        for (let i = 0; i < Math.min(dropsLeaked, 3); i++) {
            setTimeout(() => createDropAnimation('leaky-bucket', 'falling-out'), i * 150);
        }
    }
    previousQueue = currentQueue;
    
    if (queueEl) queueEl.textContent = currentQueue;
    
    // Show bucket fill level
    if (fillEl) {
        const fillPercent = (currentQueue / state.capacity) * 100;
        fillEl.style.height = fillPercent + '%';
    }
    
    renderHistory(historyEl, state.requestHistory);
}

function renderHistory(container, history) {
    if (!container || !history) return;
    
    const last32 = history.slice(-32);  // Show last 32 requests
    
    container.innerHTML = `
        <div class="history-label">Last ${last32.length} requests:</div>
        <div class="history-items">
            ${last32.map(req => `
                <div class="history-item ${req.allowed ? 'allowed' : 'rejected'}" 
                     title="${req.allowed ? 'Allowed' : 'Rejected'} at ${new Date(req.timestamp).toLocaleTimeString()}">
                </div>
            `).join('')}
        </div>
    `;
}

function sendSingleRequest() {
    fetch(`${API_BASE}/send-request`, { method: 'POST' })
        .then(response => response.json())
        .then(data => {
            allStates = data.states;
            renderAllAlgorithms();
            
            // Show results
            const results = data.results;
            const allowed = Object.values(results).filter(r => r).length;
            const rejected = 5 - allowed;
            
            showFeedback(`Request sent! ✅ ${allowed} allowed, ❌ ${rejected} rejected`, 
                         allowed > 0 ? 'success' : 'error');
        })
        .catch(error => {
            console.error('Error sending request:', error);
            showFeedback('Error sending request', 'error');
        });
}

function sendBurstRequests() {
    const burstBtn = document.getElementById('send-burst-btn');
    burstBtn.disabled = true;
    burstBtn.textContent = 'Sending 10 requests...';
    
    fetch(`${API_BASE}/send-burst?count=10`, { method: 'POST' })
        .then(response => response.json())
        .then(data => {
            allStates = data.states;
            renderAllAlgorithms();
            
            // Calculate totals
            const totals = {
                fixedWindow: 0,
                slidingLog: 0,
                slidingWindow: 0,
                tokenBucket: 0,
                leakyBucket: 0
            };
            
            data.results.forEach(result => {
                if (result.fixedWindow) totals.fixedWindow++;
                if (result.slidingLog) totals.slidingLog++;
                if (result.slidingWindow) totals.slidingWindow++;
                if (result.tokenBucket) totals.tokenBucket++;
                if (result.leakyBucket) totals.leakyBucket++;
            });
            
            showFeedback(
                `10 requests sent! Results: ` +
                `FW:${totals.fixedWindow} | SL:${totals.slidingLog} | ` +
                `SW:${totals.slidingWindow} | TB:${totals.tokenBucket} | LB:${totals.leakyBucket}`,
                'success'
            );
        })
        .catch(error => {
            console.error('Error sending burst:', error);
            showFeedback('Error sending burst requests', 'error');
        })
        .finally(() => {
            burstBtn.disabled = false;
            burstBtn.textContent = 'Send 10 Requests (Burst)';
        });
}

function resetAll() {
    fetch(`${API_BASE}/reset`, { method: 'POST' })
        .then(response => response.json())
        .then(data => {
            allStates = data;
            renderAllAlgorithms();
            showFeedback('All algorithms reset!', 'success');
        })
        .catch(error => {
            console.error('Error resetting:', error);
            showFeedback('Error resetting algorithms', 'error');
        });
}

function showFeedback(message, type = 'info') {
    const feedbackEl = document.getElementById('request-feedback');
    if (!feedbackEl) return;
    
    const color = type === 'success' ? 'var(--accent-green)' : 
                  type === 'error' ? 'var(--accent-red)' : 
                  'var(--text-secondary)';
    
    feedbackEl.innerHTML = `<span style="color: ${color}; font-weight: 600;">${message}</span>`;
}

// Create drop animation for buckets
function createDropAnimation(bucketType, animationClass) {
    const containerId = bucketType === 'token-bucket' ? 'token-bucket-card' : 'leaky-bucket-card';
    const container = document.getElementById(containerId);
    if (!container) return;
    
    const drop = document.createElement('div');
    drop.className = `water-drop ${animationClass}`;
    
    // Position at bucket center
    const bucketViz = container.querySelector(`.${bucketType === 'token-bucket' ? 'token-bucket-viz' : 'leaky-bucket-viz'}`);
    if (bucketViz) {
        const rect = bucketViz.getBoundingClientRect();
        const containerRect = container.getBoundingClientRect();
        
        drop.style.left = (rect.left - containerRect.left + rect.width / 2 - 4) + 'px';
        
        if (animationClass === 'falling-in') {
            drop.style.top = (rect.top - containerRect.top - 30) + 'px';
        } else {
            drop.style.top = (rect.bottom - containerRect.top) + 'px';
        }
    }
    
    container.appendChild(drop);
    
    // Remove after animation
    setTimeout(() => {
        if (drop.parentNode) {
            drop.parentNode.removeChild(drop);
        }
    }, 1000);
}

// Theme management
function initTheme() {
    const savedTheme = localStorage.getItem('theme') || 'dark';
    document.documentElement.setAttribute('data-theme', savedTheme);
    updateThemeIcon(savedTheme);
}

function toggleTheme() {
    const currentTheme = document.documentElement.getAttribute('data-theme');
    const newTheme = currentTheme === 'dark' ? 'light' : 'dark';
    document.documentElement.setAttribute('data-theme', newTheme);
    localStorage.setItem('theme', newTheme);
    updateThemeIcon(newTheme);
}

function updateThemeIcon(theme) {
    const themeIcon = document.querySelector('.theme-icon');
    if (themeIcon) {
        themeIcon.textContent = theme === 'dark' ? '☀️' : '🌙';
    }
}

// Auto-refresh to show live state changes (token refill, leaky bucket drain)
setInterval(() => {
    if (allStates) {
        loadStates();
    }
}, 1000);  // Refresh every second

// Initialize when DOM is ready
document.addEventListener('DOMContentLoaded', () => {
    console.log('✅ DOM Content Loaded');
    
    initTheme();
    initVisualization();
    
    // Setup theme toggle
    const themeToggle = document.getElementById('theme-toggle');
    if (themeToggle) {
        themeToggle.addEventListener('click', toggleTheme);
    }
    
    // Setup control buttons
    const sendBtn = document.getElementById('send-request-btn');
    if (sendBtn) {
        sendBtn.addEventListener('click', sendSingleRequest);
    }
    
    const burstBtn = document.getElementById('send-burst-btn');
    if (burstBtn) {
        burstBtn.addEventListener('click', sendBurstRequests);
    }
    
    const resetBtn = document.getElementById('reset-btn');
    if (resetBtn) {
        resetBtn.addEventListener('click', resetAll);
    }
});

