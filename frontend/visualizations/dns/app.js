// DNS Resolution Visualization

// Initialize theme immediately
(function() {
    const savedTheme = localStorage.getItem('theme') || 'dark';
    document.documentElement.setAttribute('data-theme', savedTheme);
})();

const API_BASE = 'http://localhost:8080/api/dns';

// Helper function to add session header to fetch options
function addSessionHeader(options = {}) {
    if (typeof window.SDS_SESSION !== 'undefined') {
        return window.SDS_SESSION.addSessionHeader(options);
    }
    return options;
}

let currentState = null;
let ttlUpdateInterval = null;

// Initialize visualization
function initVisualization() {
    console.log('🔧 Initializing DNS visualization...');
    loadState();
    setupEventListeners();
    startTTLUpdates();
}

function setupEventListeners() {
    // Resolve button
    document.getElementById('resolve-btn').addEventListener('click', resolveDomain);
    
    // Enter key in domain input
    document.getElementById('domain-input').addEventListener('keypress', (e) => {
        if (e.key === 'Enter') {
            resolveDomain();
        }
    });
    
    // Clear cache button
    document.getElementById('clear-cache-btn').addEventListener('click', clearCache);
    
    // Reset button
    document.getElementById('reset-btn').addEventListener('click', reset);
    
    // Quick domain buttons
    document.querySelectorAll('.quick-domain-btn').forEach(btn => {
        btn.addEventListener('click', () => {
            const domain = btn.getAttribute('data-domain');
            document.getElementById('domain-input').value = domain;
            resolveDomain();
        });
    });
    
    // Theme toggle
    const themeToggle = document.getElementById('theme-toggle');
    if (themeToggle) {
        themeToggle.addEventListener('click', toggleTheme);
    }
}

function loadState() {
    fetch(`${API_BASE}/state`, addSessionHeader())
        .then(response => response.json())
        .then(data => {
            currentState = data;
            renderAll();
        })
        .catch(error => {
            console.error('Error loading state:', error);
            showFeedback('Error connecting to backend. Make sure it\'s running on port 8080.', 'error');
        });
}

function resolveDomain() {
    const domainInput = document.getElementById('domain-input');
    const domain = domainInput.value.trim();
    
    if (!domain) {
        showFeedback('Please enter a domain name', 'error');
        return;
    }
    
    showFeedback('Resolving domain...', 'info');
    
    fetch(`${API_BASE}/resolve`, addSessionHeader({
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ domain })
    }))
        .then(response => response.json())
        .then(data => {
            currentState = data.state;
            renderAll();
            
            // Show result feedback
            const result = data.result;
            if (result.success) {
                const method = result.method === 'cache' ? 'Cache Hit' : 'Recursive Query';
                const latencyInfo = result.cacheExpired ? ' (cache expired)' : '';
                showFeedback(
                    `✅ ${domain} → ${result.ipAddress} via ${method} (${result.totalLatency}ms)${latencyInfo}`,
                    result.method === 'cache' ? 'success' : 'warning'
                );
                
                // Animate the latest query
                animateLatestQuery(result);
            } else {
                showFeedback(`❌ Failed to resolve ${domain}`, 'error');
            }
        })
        .catch(error => {
            console.error('Error resolving domain:', error);
            showFeedback('Error resolving domain', 'error');
        });
}

function clearCache() {
    if (!confirm('Clear all DNS cache entries?')) return;
    
    fetch(`${API_BASE}/clear-cache`, addSessionHeader({ method: 'POST' }))
        .then(response => response.json())
        .then(data => {
            currentState = data;
            renderAll();
            showFeedback('✅ DNS cache cleared', 'success');
        })
        .catch(error => {
            console.error('Error clearing cache:', error);
            showFeedback('Error clearing cache', 'error');
        });
}

function reset() {
    if (!confirm('Reset DNS simulator to initial state?')) return;
    
    fetch(`${API_BASE}/reset`, addSessionHeader({ method: 'POST' }))
        .then(response => response.json())
        .then(data => {
            currentState = data;
            renderAll();
            showFeedback('✅ DNS simulator reset', 'success');
            document.getElementById('latest-query-section').style.display = 'none';
            // Restart TTL updates after reset
            startTTLUpdates();
        })
        .catch(error => {
            console.error('Error resetting:', error);
            showFeedback('Error resetting', 'error');
        });
}

function renderAll() {
    if (!currentState) return;
    
    renderStatistics();
    renderCache();
    renderQueryHistory();
}

function renderStatistics() {
    document.getElementById('stat-total').textContent = currentState.totalQueries;
    document.getElementById('stat-cache-hits').textContent = currentState.cacheHits;
    document.getElementById('stat-cache-misses').textContent = currentState.cacheMisses;
    document.getElementById('stat-expired').textContent = currentState.expiredCacheHits;
    
    document.getElementById('stat-cache-avg').textContent = 
        `~${currentState.avgCacheLatency.toFixed(1)}ms avg`;
    document.getElementById('stat-recursive-avg').textContent = 
        `~${currentState.avgRecursiveLatency.toFixed(0)}ms avg`;
}

function renderCache() {
    const container = document.getElementById('cache-visualization');
    container.innerHTML = '';
    
    const cacheEntries = Object.values(currentState.cache);
    
    if (cacheEntries.length === 0) {
        container.innerHTML = '<div class="empty-state">Cache is empty. Resolve a domain to populate it.</div>';
        return;
    }
    
    // Create cache table
    const table = document.createElement('table');
    table.className = 'cache-table';
    table.id = 'cache-table';
    
    const thead = document.createElement('thead');
    thead.innerHTML = `
        <tr>
            <th>Domain</th>
            <th>IP Address</th>
            <th>TTL Remaining</th>
            <th>Status</th>
        </tr>
    `;
    table.appendChild(thead);
    
    const tbody = document.createElement('tbody');
    tbody.id = 'cache-tbody';
    
    cacheEntries.forEach((entry, index) => {
        const now = new Date();
        const expiresAt = new Date(entry.expiresAt);
        const ttlRemaining = Math.max(0, Math.floor((expiresAt - now) / 1000));
        
        const row = document.createElement('tr');
        row.className = entry.isExpired ? 'expired' : (ttlRemaining < 30 ? 'expiring' : '');
        row.setAttribute('data-domain', entry.domain);
        row.setAttribute('data-expires-at', entry.expiresAt);
        
        let statusText = '';
        let statusClass = '';
        if (entry.isExpired) {
            statusText = '❌ Expired';
            statusClass = 'status-expired';
        } else if (ttlRemaining < 30) {
            statusText = '⚠️ Expiring Soon';
            statusClass = 'status-expiring';
        } else {
            statusText = '✅ Valid';
            statusClass = 'status-valid';
        }
        
        row.innerHTML = `
            <td class="domain-cell">${entry.domain}</td>
            <td class="ip-cell">${entry.ipAddress}</td>
            <td class="ttl-cell" data-ttl="${ttlRemaining}">${ttlRemaining}s</td>
            <td class="status-cell ${statusClass}">${statusText}</td>
        `;
        
        tbody.appendChild(row);
    });
    
    table.appendChild(tbody);
    container.appendChild(table);
}

// Update TTL countdown every second
function updateTTLCountdown() {
    const tbody = document.getElementById('cache-tbody');
    if (!tbody) return;
    
    const rows = tbody.querySelectorAll('tr');
    const now = new Date();
    
    rows.forEach(row => {
        const expiresAt = new Date(row.getAttribute('data-expires-at'));
        const ttlRemaining = Math.max(0, Math.floor((expiresAt - now) / 1000));
        
        const ttlCell = row.querySelector('.ttl-cell');
        const statusCell = row.querySelector('.status-cell');
        
        if (ttlCell) {
            ttlCell.textContent = `${ttlRemaining}s`;
            ttlCell.setAttribute('data-ttl', ttlRemaining);
        }
        
        // Update row class and status based on TTL
        if (ttlRemaining === 0) {
            row.className = 'expired';
            if (statusCell) {
                statusCell.className = 'status-cell status-expired';
                statusCell.textContent = '❌ Expired';
            }
        } else if (ttlRemaining < 30) {
            row.className = 'expiring';
            if (statusCell && !statusCell.classList.contains('status-expired')) {
                statusCell.className = 'status-cell status-expiring';
                statusCell.textContent = '⚠️ Expiring Soon';
            }
        } else {
            row.className = '';
            if (statusCell && !statusCell.classList.contains('status-expired') && !statusCell.classList.contains('status-expiring')) {
                statusCell.className = 'status-cell status-valid';
                statusCell.textContent = '✅ Valid';
            }
        }
    });
}

// Start TTL countdown updates
function startTTLUpdates() {
    // Clear any existing interval
    if (ttlUpdateInterval) {
        clearInterval(ttlUpdateInterval);
    }
    
    // Update every second
    ttlUpdateInterval = setInterval(updateTTLCountdown, 1000);
}

// Stop TTL updates (cleanup)
function stopTTLUpdates() {
    if (ttlUpdateInterval) {
        clearInterval(ttlUpdateInterval);
        ttlUpdateInterval = null;
    }
}

function renderQueryHistory() {
    const container = document.getElementById('query-history');
    container.innerHTML = '';
    
    if (currentState.queryHistory.length === 0) {
        container.innerHTML = '<div class="empty-state">No queries yet. Resolve a domain to see history.</div>';
        return;
    }
    
    // Show last 10 queries
    const recentQueries = currentState.queryHistory.slice(-10).reverse();
    
    const historyList = document.createElement('div');
    historyList.className = 'history-list';
    
    recentQueries.forEach(query => {
        const item = document.createElement('div');
        item.className = `history-item ${query.method}`;
        
        const methodBadge = query.method === 'cache' ? 
            '<span class="badge badge-cache">Cache</span>' : 
            '<span class="badge badge-recursive">Recursive</span>';
        
        const expiredNote = query.cacheExpired ? 
            '<span class="badge badge-expired">Expired</span>' : '';
        
        const latencyClass = query.totalLatency < 10 ? 'latency-fast' : 
                            query.totalLatency < 100 ? 'latency-medium' : 'latency-slow';
        
        item.innerHTML = `
            <div class="history-header">
                <span class="history-domain">${query.domain}</span>
                ${methodBadge}
                ${expiredNote}
            </div>
            <div class="history-details">
                <span class="history-ip">${query.ipAddress}</span>
                <span class="history-latency ${latencyClass}">${query.totalLatency}ms</span>
            </div>
        `;
        
        historyList.appendChild(item);
    });
    
    container.appendChild(historyList);
}

function animateLatestQuery(result) {
    const section = document.getElementById('latest-query-section');
    const container = document.getElementById('latest-query-result');
    
    section.style.display = 'block';
    container.innerHTML = '';
    
    if (result.method === 'cache') {
        // Cache hit visualization
        container.innerHTML = `
            <div class="query-result cache-result">
                <div class="result-header">
                    <h3>⚡ Cache Lookup</h3>
                    <span class="result-time">${result.totalLatency}ms</span>
                </div>
                <div class="result-body">
                    <div class="result-flow">
                        <div class="flow-step instant">
                            <div class="step-icon">💾</div>
                            <div class="step-label">Local Cache</div>
                            <div class="step-time">~1ms</div>
                        </div>
                    </div>
                    <div class="result-summary">
                        <strong>✅ Instant Lookup</strong><br>
                        Zero network latency • No I/O overhead • Immediate response
                        ${result.cacheExpired ? '<br><strong>⚠️ Note:</strong> Cache was expired, performed recursive query' : ''}
                    </div>
                </div>
            </div>
        `;
    } else {
        // Recursive query visualization
        renderRecursiveQueryFlow(result, container);
    }
    
    // Scroll to result
    section.scrollIntoView({ behavior: 'smooth', block: 'nearest' });
}

function renderRecursiveQueryFlow(result, container) {
    const totalLatency = result.totalLatency;
    
    const flowHTML = `
        <div class="query-result recursive-result">
            <div class="result-header">
                <h3>🌐 Recursive DNS Query</h3>
                <span class="result-time">${totalLatency}ms</span>
            </div>
            <div class="result-body">
                <div class="result-flow-vertical">
                    ${result.steps.map((step, idx) => `
                        <div class="flow-step-vertical" style="animation-delay: ${idx * 0.2}s">
                            <div class="step-number">${step.stepNumber}</div>
                            <div class="step-content">
                                <div class="step-route">
                                    <span class="step-from">${step.fromServer}</span>
                                    <span class="step-arrow">→</span>
                                    <span class="step-to">${step.toServer}</span>
                                </div>
                                <div class="step-description">${step.description}</div>
                                <div class="step-latency">+${step.latency}ms</div>
                            </div>
                        </div>
                    `).join('')}
                </div>
                <div class="result-summary warning">
                    <strong>⚠️ High Latency Trade-off</strong><br>
                    Total: ${totalLatency}ms • ${result.steps.length} network round-trips • Multiple I/O operations<br>
                    <strong>Benefit:</strong> Fresh, authoritative data with TTL: ${result.ttl}s
                </div>
            </div>
        </div>
    `;
    
    container.innerHTML = flowHTML;
}

function showFeedback(message, type = 'info') {
    const feedback = document.getElementById('operation-feedback');
    feedback.textContent = message;
    feedback.className = `feedback ${type}`;
    feedback.style.display = 'block';
    
    setTimeout(() => {
        feedback.style.display = 'none';
    }, 5000);
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

// Initialize on page load
document.addEventListener('DOMContentLoaded', () => {
    const savedTheme = localStorage.getItem('theme') || 'dark';
    updateThemeIcon(savedTheme);
    initVisualization();
});

