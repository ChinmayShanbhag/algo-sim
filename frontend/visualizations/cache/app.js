// Theme management (initialize immediately)
(function() {
    const savedTheme = localStorage.getItem('theme') || 'dark';
    document.documentElement.setAttribute('data-theme', savedTheme);
})();

const API_BASE = 'http://localhost:8080/api/cache';

// Helper function to add session header to fetch options
function addSessionHeader(options = {}) {
    if (typeof window.SDS_SESSION !== 'undefined') {
        return window.SDS_SESSION.addSessionHeader(options);
    }
    return options;
}

let allStates = null;

// Initialize visualization
function initVisualization() {
    console.log('🔧 Initializing Cache Eviction visualization...');
    loadStates();
    
    // Set up event listeners
    document.getElementById('execute-btn').addEventListener('click', executeOperation);
    document.getElementById('reset-btn').addEventListener('click', resetAll);
    document.getElementById('operation-select').addEventListener('change', updateValueFieldVisibility);
    
    // Allow Enter key to execute
    document.getElementById('key-input').addEventListener('keypress', (e) => {
        if (e.key === 'Enter') executeOperation();
    });
    document.getElementById('value-input').addEventListener('keypress', (e) => {
        if (e.key === 'Enter') executeOperation();
    });
    
    updateValueFieldVisibility();
}

function updateValueFieldVisibility() {
    const operation = document.getElementById('operation-select').value;
    const valueGroup = document.getElementById('value-group');
    valueGroup.style.display = operation === 'PUT' ? 'flex' : 'none';
}

function loadStates() {
    fetch(`${API_BASE}/state`, addSessionHeader())
        .then(response => response.json())
        .then(data => {
            allStates = data;
            renderAllCaches();
        })
        .catch(error => {
            console.error('Error loading states:', error);
            showFeedback('Error connecting to backend. Make sure it\'s running on port 8080.', 'error');
        });
}

function renderAllCaches() {
    if (!allStates) return;
    
    renderCache('lru', allStates.lru);
    renderCache('lfu', allStates.lfu);
    renderCache('fifo', allStates.fifo);
}

function renderCache(type, state) {
    // Update capacity and size
    document.getElementById(`${type}-capacity`).textContent = state.capacity;
    document.getElementById(`${type}-size`).textContent = state.size;
    
    // Render cache items as a vertical stack
    const vizContainer = document.getElementById(`${type}-viz`);
    vizContainer.innerHTML = '';
    
    if (state.items.length === 0) {
        vizContainer.innerHTML = '<div style="color: var(--text-muted); text-align: center; padding: 40px; font-size: 0.9rem;">Cache is empty<br><span style="font-size: 0.8rem; opacity: 0.7;">Add items using PUT operation</span></div>';
    } else {
        state.items.forEach((item, index) => {
            const itemDiv = document.createElement('div');
            itemDiv.className = 'cache-item';
            
            // Add position-based color variation
            const hue = 210 + (index * 15); // Vary blue hue
            itemDiv.style.background = `hsl(${hue}, 70%, 50%)`;
            
            let metaInfo = '';
            if (type === 'lfu' && item.frequency) {
                metaInfo = `Frequency: ${item.frequency}`;
            } else if (type === 'lru') {
                metaInfo = index === 0 ? 'Most Recent' : index === state.items.length - 1 ? 'Least Recent' : `Position ${index + 1}`;
            } else if (type === 'fifo') {
                metaInfo = index === 0 ? 'Newest' : index === state.items.length - 1 ? 'Oldest (Next Out)' : `Position ${index + 1}`;
            }
            
            itemDiv.innerHTML = `
                <div style="display: flex; align-items: center; gap: 10px;">
                    <span class="key">${item.key}</span>
                    <span class="value">${item.value}</span>
                </div>
                ${metaInfo ? `<span class="meta">${metaInfo}</span>` : ''}
            `;
            
            vizContainer.appendChild(itemDiv);
        });
    }
    
    // Render history
    renderHistory(type, state.history);
}

function renderHistory(type, history) {
    const historyContainer = document.getElementById(`${type}-history`);
    historyContainer.innerHTML = '';
    
    if (!history || history.length === 0) {
        historyContainer.innerHTML = '<div style="color: var(--text-muted); text-align: center; padding: 10px;">No operations yet</div>';
        return;
    }
    
    // Show last 10 operations
    const recentHistory = history.slice(-10).reverse();
    
    recentHistory.forEach(event => {
        const itemDiv = document.createElement('div');
        let className = 'history-item';
        let text = '';
        
        if (event.operation === 'GET') {
            className += event.hit ? ' hit' : ' miss';
            text = `GET ${event.key}: ${event.hit ? '✓ HIT' : '✗ MISS'} `;
        } else if (event.operation === 'PUT') {
            if (event.evictedKey) {
                className += ' eviction';
                text = `PUT ${event.key}: Evicted ${event.evictedKey}`;
            } else {
                className += ' hit';
                text = `PUT ${event.key}: Added`;
            }
        }
        
        itemDiv.className = className;
        itemDiv.textContent = text;
        historyContainer.appendChild(itemDiv);
    });
}

function executeOperation() {
    const operation = document.getElementById('operation-select').value;
    const key = document.getElementById('key-input').value.trim();
    const value = document.getElementById('value-input').value.trim();
    
    if (!key) {
        showFeedback('Please enter a key', 'error');
        return;
    }
    
    if (operation === 'PUT' && !value) {
        showFeedback('Please enter a value for PUT operation', 'error');
        return;
    }
    
    let url = `${API_BASE}/access?operation=${operation}&key=${encodeURIComponent(key)}`;
    if (operation === 'PUT') {
        url += `&value=${encodeURIComponent(value)}`;
    }
    
    fetch(url, addSessionHeader({ method: 'POST' }))
        .then(response => response.json())
        .then(data => {
            allStates = data.states;
            renderAllCaches();
            
            // Show feedback
            const results = data.results;
            if (operation === 'GET') {
                const hits = Object.values(results).filter(r => r.hit).length;
                const misses = 3 - hits;
                
                // Get the value from any cache that had a hit
                let hitValue = '';
                for (const [algo, result] of Object.entries(results)) {
                    if (result.hit && result.value) {
                        hitValue = result.value;
                        break;
                    }
                }
                
                if (hits > 0) {
                    showFeedback(`GET ${key}: ${hits} hits, ${misses} misses | Value: "${hitValue}"`, 'success');
                } else {
                    showFeedback(`GET ${key}: ${hits} hits, ${misses} misses | Key not found`, 'error');
                }
            } else {
                const evictions = Object.entries(results)
                    .filter(([_, r]) => r.evicted)
                    .map(([algo, r]) => `${algo.toUpperCase()}: ${r.evicted}`)
                    .join(', ');
                
                if (evictions) {
                    showFeedback(`PUT ${key}: Evicted ${evictions}`, 'warning');
                } else {
                    showFeedback(`PUT ${key}: Added to all caches`, 'success');
                }
            }
            
            // Clear inputs
            if (operation === 'PUT') {
                document.getElementById('key-input').value = '';
                document.getElementById('value-input').value = '';
            }
        })
        .catch(error => {
            console.error('Error executing operation:', error);
            showFeedback('Error executing operation', 'error');
        });
}

function resetAll() {
    if (!confirm('Reset all caches? This will clear all data.')) {
        return;
    }
    
    fetch(`${API_BASE}/reset`, addSessionHeader({ method: 'POST' }))
        .then(response => response.json())
        .then(data => {
            allStates = data;
            renderAllCaches();
            showFeedback('All caches reset successfully', 'success');
            
            // Clear inputs
            document.getElementById('key-input').value = '';
            document.getElementById('value-input').value = '';
        })
        .catch(error => {
            console.error('Error resetting:', error);
            showFeedback('Error resetting caches', 'error');
        });
}

function showFeedback(message, type = 'info') {
    const feedbackEl = document.getElementById('operation-feedback');
    if (!feedbackEl) return;
    
    const colors = {
        success: 'var(--accent-green)',
        error: 'var(--accent-red)',
        warning: 'var(--accent-yellow)',
        info: 'var(--text-secondary)'
    };
    
    feedbackEl.style.color = colors[type] || colors.info;
    feedbackEl.style.background = type === 'error' ? 'rgba(239, 68, 68, 0.1)' : 
                                   type === 'success' ? 'rgba(34, 197, 94, 0.1)' :
                                   type === 'warning' ? 'rgba(251, 191, 36, 0.1)' :
                                   'transparent';
    feedbackEl.textContent = message;
    
    // Auto-clear after 5 seconds
    setTimeout(() => {
        feedbackEl.textContent = '';
        feedbackEl.style.background = 'transparent';
    }, 5000);
}

// Initialize on page load
document.addEventListener('DOMContentLoaded', () => {
    initVisualization();
    
    // Setup theme toggle
    const themeToggle = document.getElementById('theme-toggle');
    if (themeToggle) {
        themeToggle.addEventListener('click', () => {
            const currentTheme = document.documentElement.getAttribute('data-theme');
            const newTheme = currentTheme === 'dark' ? 'light' : 'dark';
            document.documentElement.setAttribute('data-theme', newTheme);
            localStorage.setItem('theme', newTheme);
            
            // Update icon
            const icon = themeToggle.querySelector('.theme-icon');
            if (icon) {
                icon.textContent = newTheme === 'dark' ? '🌙' : '☀️';
            }
            
            // Re-render to update colors
            if (allStates) {
                renderAllCaches();
            }
        });
        
        // Set initial icon
        const currentTheme = document.documentElement.getAttribute('data-theme');
        const icon = themeToggle.querySelector('.theme-icon');
        if (icon) {
            icon.textContent = currentTheme === 'dark' ? '🌙' : '☀️';
        }
    }
});

// Auto-refresh state every 2 seconds
setInterval(loadStates, 2000);

