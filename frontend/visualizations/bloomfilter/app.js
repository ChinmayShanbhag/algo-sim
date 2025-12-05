// Theme management
(function() {
    const savedTheme = localStorage.getItem('theme') || 'dark';
    document.documentElement.setAttribute('data-theme', savedTheme);
})();

const API_BASE = 'http://localhost:8080/api/bloomfilter';

function addSessionHeader(options = {}) {
    if (typeof window.SDS_SESSION !== 'undefined') {
        return window.SDS_SESSION.addSessionHeader(options);
    }
    return options;
}

let filterState = null;
let lastHighlightedBits = [];

function initVisualization() {
    console.log('🔧 Initializing Bloom Filter visualization...');
    
    // Setup event listeners
    document.getElementById('add-btn').addEventListener('click', addItem);
    document.getElementById('check-btn').addEventListener('click', checkItem);
    document.getElementById('reset-btn').addEventListener('click', resetFilter);
    
    // Allow Enter key
    document.getElementById('item-input').addEventListener('keypress', (e) => {
        if (e.key === 'Enter') addItem();
    });
    
    // Setup theme toggle
    const themeToggle = document.getElementById('theme-toggle');
    if (themeToggle) {
        themeToggle.addEventListener('click', () => {
            const currentTheme = document.documentElement.getAttribute('data-theme');
            const newTheme = currentTheme === 'dark' ? 'light' : 'dark';
            document.documentElement.setAttribute('data-theme', newTheme);
            localStorage.setItem('theme', newTheme);
            const icon = themeToggle.querySelector('.theme-icon');
            if (icon) icon.textContent = newTheme === 'dark' ? '🌙' : '☀️';
            if (filterState) renderFilterState();
        });
        const currentTheme = document.documentElement.getAttribute('data-theme');
        const icon = themeToggle.querySelector('.theme-icon');
        if (icon) icon.textContent = currentTheme === 'dark' ? '🌙' : '☀️';
    }
    
    loadFilterState();
}

function loadFilterState() {
    fetch(`${API_BASE}/state`, addSessionHeader())
        .then(response => response.json())
        .then(data => {
            filterState = data;
            renderFilterState();
        })
        .catch(error => {
            console.error('Error loading filter state:', error);
            showFeedback('Error connecting to backend', 'error');
        });
}

function renderFilterState() {
    if (!filterState) return;
    
    renderBitArray();
    renderItemsAdded();
    renderRecentOps();
    renderStatistics();
    clearHashFunctions();
}

function renderBitArray() {
    const container = document.getElementById('bit-array-container');
    container.innerHTML = '';
    
    filterState.bitArray.forEach((bit, index) => {
        const cell = document.createElement('div');
        cell.className = 'bit-cell';
        if (bit) cell.classList.add('set');
        
        cell.innerHTML = `
            <div class="bit-index">${index}</div>
            <div class="bit-value">${bit ? '1' : '0'}</div>
        `;
        
        cell.title = `Bit ${index}: ${bit ? 'Set (1)' : 'Unset (0)'}`;
        container.appendChild(cell);
    });
    
    // Update fill percentage
    const bitsSet = filterState.bitArray.filter(b => b).length;
    const fillPercentage = (bitsSet / filterState.size * 100).toFixed(1);
    document.getElementById('fill-percentage').textContent = `${fillPercentage}%`;
    document.getElementById('bits-set').textContent = `${bitsSet}/${filterState.size}`;
}

function highlightBits(bitPositions) {
    // Remove previous highlights
    document.querySelectorAll('.bit-cell.highlighted').forEach(cell => {
        cell.classList.remove('highlighted');
    });
    
    // Add new highlights
    const cells = document.querySelectorAll('.bit-cell');
    bitPositions.forEach(pos => {
        if (cells[pos]) {
            cells[pos].classList.add('highlighted');
        }
    });
    
    lastHighlightedBits = bitPositions;
}

function clearHashFunctions() {
    document.getElementById('hash-1').textContent = '-';
    document.getElementById('hash-2').textContent = '-';
    document.getElementById('hash-3').textContent = '-';
}

function showHashFunctions(hashBits) {
    document.getElementById('hash-1').textContent = hashBits[0] !== undefined ? hashBits[0] : '-';
    document.getElementById('hash-2').textContent = hashBits[1] !== undefined ? hashBits[1] : '-';
    document.getElementById('hash-3').textContent = hashBits[2] !== undefined ? hashBits[2] : '-';
}

function renderItemsAdded() {
    const container = document.getElementById('items-added');
    container.innerHTML = '';
    
    if (!filterState.itemsAdded || filterState.itemsAdded.length === 0) {
        container.innerHTML = '<div style="color: var(--text-muted); padding: 10px;">No items added yet</div>';
        return;
    }
    
    filterState.itemsAdded.forEach(item => {
        const tag = document.createElement('div');
        tag.className = 'item-tag';
        tag.textContent = item;
        container.appendChild(tag);
    });
}

function renderRecentOps() {
    const container = document.getElementById('recent-ops');
    container.innerHTML = '';
    
    if (!filterState.recentOps || filterState.recentOps.length === 0) {
        container.innerHTML = '<div style="color: var(--text-muted); padding: 10px;">No operations yet</div>';
        return;
    }
    
    // Show most recent first
    const sortedOps = [...filterState.recentOps].reverse();
    
    sortedOps.forEach(op => {
        const item = document.createElement('div');
        item.className = 'operation-item';
        
        let resultHTML = '';
        if (op.type === 'CHECK') {
            const isFalsePositive = op.result === 'probably_yes' && !op.actualIn;
            const resultClass = isFalsePositive ? 'false-positive' : 
                               op.result === 'probably_yes' ? 'probably-yes' : 'definitely-not';
            const resultText = isFalsePositive ? 'FALSE POSITIVE!' : 
                              op.result === 'probably_yes' ? 'Probably Yes' : 'Definitely Not';
            resultHTML = `<span class="operation-result ${resultClass}">${resultText}</span>`;
        }
        
        item.innerHTML = `
            <div class="operation-header">
                <span class="operation-type ${op.type.toLowerCase()}">${op.type}</span>
                ${resultHTML}
            </div>
            <div class="operation-details">
                <strong>${op.item}</strong> → Bits: [${op.hashBits.join(', ')}]
                ${op.type === 'CHECK' ? ` | Actually in set: ${op.actualIn ? 'Yes' : 'No'}` : ''}
                <br><small>${op.timestamp}</small>
            </div>
        `;
        container.appendChild(item);
    });
}

function renderStatistics() {
    document.getElementById('stat-tp').textContent = filterState.truePositives;
    document.getElementById('stat-fp').textContent = filterState.falsePositives;
    document.getElementById('stat-tn').textContent = filterState.trueNegatives;
    
    const totalChecks = filterState.truePositives + filterState.falsePositives + filterState.trueNegatives;
    const fpr = totalChecks > 0 ? (filterState.falsePositives / (filterState.falsePositives + filterState.trueNegatives) * 100).toFixed(1) : 0;
    document.getElementById('stat-fpr').textContent = `${fpr}%`;
}

function addItem() {
    const input = document.getElementById('item-input');
    const item = input.value.trim().toLowerCase();
    
    if (!item) {
        showFeedback('Please enter an item', 'error');
        return;
    }
    
    fetch(`${API_BASE}/add?item=${encodeURIComponent(item)}`, addSessionHeader({ method: 'POST' }))
        .then(response => response.json())
        .then(data => {
            filterState = data;
            renderFilterState();
            
            // Show hash functions and highlight bits
            const lastOp = data.recentOps[data.recentOps.length - 1];
            if (lastOp && lastOp.hashBits) {
                showHashFunctions(lastOp.hashBits);
                highlightBits(lastOp.hashBits);
            }
            
            showFeedback(`✅ Added "${item}" to filter. Bits set: [${lastOp.hashBits.join(', ')}]`, 'success');
            input.value = '';
        })
        .catch(error => {
            console.error('Error adding item:', error);
            showFeedback('Error adding item', 'error');
        });
}

function checkItem() {
    const input = document.getElementById('item-input');
    const item = input.value.trim().toLowerCase();
    
    if (!item) {
        showFeedback('Please enter an item', 'error');
        return;
    }
    
    fetch(`${API_BASE}/check?item=${encodeURIComponent(item)}`, addSessionHeader())
        .then(response => response.json())
        .then(result => {
            // Show hash functions and highlight bits FIRST
            showHashFunctions(result.hashBits);
            highlightBits(result.hashBits);
            
            let message = '';
            let type = 'info';
            
            if (result.isFalsePositive) {
                message = `🚨 FALSE POSITIVE! "${item}" is NOT in the set, but the filter says "probably yes". Bits checked: [${result.hashBits.join(', ')}]`;
                type = 'warning';
            } else if (result.result === 'probably_yes') {
                message = `✅ "${item}" is PROBABLY in the set (true positive). Bits checked: [${result.hashBits.join(', ')}]`;
                type = 'success';
            } else {
                message = `❌ "${item}" is DEFINITELY NOT in the set. Bits checked: [${result.hashBits.join(', ')}]`;
                type = 'info';
            }
            
            // Show feedback BEFORE reloading state (so it doesn't get cleared)
            showFeedback(message, type, 15000); // 15 seconds for CHECK results - longer for false positives
            
            // Reload state to get updated stats (after a longer delay to avoid clearing feedback)
            setTimeout(() => {
                fetch(`${API_BASE}/state`, addSessionHeader())
                    .then(response => response.json())
                    .then(data => {
                        filterState = data;
                        // Only update components that don't affect the feedback message
                        renderRecentOps();
                        renderStatistics();
                    });
            }, 500);
            
            input.value = '';
        })
        .catch(error => {
            console.error('Error checking item:', error);
            showFeedback('Error checking item', 'error', 15000);
        });
}

function resetFilter() {
    if (!confirm('Reset the Bloom Filter?')) return;
    
    fetch(`${API_BASE}/reset`, addSessionHeader({ method: 'POST' }))
        .then(response => response.json())
        .then(data => {
            filterState = data;
            renderFilterState();
            showFeedback('🔄 Filter reset', 'success');
        })
        .catch(error => {
            console.error('Error resetting filter:', error);
            showFeedback('Error resetting filter', 'error');
        });
}

function showFeedback(message, type, duration = 5000) {
    const feedback = document.getElementById('feedback');
    feedback.textContent = message;
    feedback.className = `feedback ${type}`;
    
    setTimeout(() => {
        feedback.textContent = '';
        feedback.className = 'feedback';
    }, duration);
}

document.addEventListener('DOMContentLoaded', initVisualization);

