// Theme management
(function() {
    const savedTheme = localStorage.getItem('theme') || 'dark';
    document.documentElement.setAttribute('data-theme', savedTheme);
})();

const API_BASE = 'http://localhost:8080/api/cdc';

function addSessionHeader(options = {}) {
    if (typeof window.SDS_SESSION !== 'undefined') {
        return window.SDS_SESSION.addSessionHeader(options);
    }
    return options;
}

let systemState = null;

function initVisualization() {
    console.log('🔧 Initializing CDC visualization...');
    
    // Setup event listeners
    document.getElementById('insert-btn').addEventListener('click', insertRecord);
    document.getElementById('update-btn').addEventListener('click', updateRecord);
    document.getElementById('delete-btn').addEventListener('click', deleteRecord);
    document.getElementById('stream-btn').addEventListener('click', streamToKafka);
    document.getElementById('consume-btn').addEventListener('click', consumeFromKafka);
    document.getElementById('reset-btn').addEventListener('click', resetSystem);
    
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
            if (systemState) renderSystemState();
        });
        const currentTheme = document.documentElement.getAttribute('data-theme');
        const icon = themeToggle.querySelector('.theme-icon');
        if (icon) icon.textContent = currentTheme === 'dark' ? '🌙' : '☀️';
    }
    
    loadSystemState();
    // Auto-refresh every 3 seconds
    setInterval(loadSystemState, 3000);
}

function loadSystemState() {
    fetch(`${API_BASE}/state`, addSessionHeader())
        .then(response => response.json())
        .then(data => {
            systemState = data;
            renderSystemState();
        })
        .catch(error => {
            console.error('Error loading system state:', error);
            showFeedback('Error connecting to backend', 'error');
        });
}

function renderSystemState() {
    if (!systemState) return;
    
    renderDatabaseRecords();
    renderReplicationLog();
    renderKafkaMessages();
    renderSearchIndex();
    renderCache();
    renderStatistics();
}

function renderDatabaseRecords() {
    const container = document.getElementById('database-records');
    container.innerHTML = '';
    
    if (!systemState.databaseRecords || systemState.databaseRecords.length === 0) {
        container.innerHTML = '<div style="color: var(--text-muted); padding: 10px;">No records</div>';
        return;
    }
    
    // Sort by ID to prevent jumping around
    const sortedRecords = [...systemState.databaseRecords].sort((a, b) => a.id - b.id);
    
    sortedRecords.forEach(record => {
        const item = document.createElement('div');
        item.className = 'record-item';
        item.innerHTML = `
            <div class="record-header">
                <span class="record-id">ID: ${record.id}</span>
                <span class="record-status ${record.status}">${record.status}</span>
            </div>
            <div class="record-details">
                <strong>${record.name}</strong><br>
                ${record.email}<br>
                <small>Updated: ${new Date(record.updatedAt).toLocaleString()}</small>
            </div>
        `;
        container.appendChild(item);
    });
}

function renderReplicationLog() {
    const container = document.getElementById('replication-log');
    container.innerHTML = '';
    
    if (!systemState.replicationLog || systemState.replicationLog.length === 0) {
        container.innerHTML = '<div style="color: var(--text-muted); padding: 10px;">No changes captured</div>';
        return;
    }
    
    // Show most recent first
    const sortedLog = [...systemState.replicationLog].reverse();
    
    sortedLog.forEach(event => {
        const item = document.createElement('div');
        item.className = 'log-item';
        item.innerHTML = `
            <div class="log-header">
                <span class="operation-badge ${event.operation.toLowerCase()}">${event.operation}</span>
                <small style="color: var(--text-muted);">LSN: ${event.lsn}</small>
            </div>
            <div style="color: var(--text-secondary); font-size: 0.8rem; margin-top: 4px;">
                ${event.table} | ID: ${event.record.id} | ${event.record.name}
            </div>
        `;
        container.appendChild(item);
    });
}

function renderKafkaMessages() {
    const container = document.getElementById('kafka-messages');
    container.innerHTML = '';
    
    if (!systemState.kafkaMessages || systemState.kafkaMessages.length === 0) {
        container.innerHTML = '<div style="color: var(--text-muted); padding: 10px;">No messages</div>';
        return;
    }
    
    // Show most recent first
    const sortedMessages = [...systemState.kafkaMessages].reverse();
    
    sortedMessages.forEach(message => {
        const item = document.createElement('div');
        item.className = 'message-item';
        item.innerHTML = `
            <div class="message-header">
                <span class="message-status ${message.status}">${message.status}</span>
                <small style="color: var(--text-muted);">Offset: ${message.offset}</small>
            </div>
            <div style="color: var(--text-secondary); font-size: 0.8rem; margin-top: 4px;">
                <span class="operation-badge ${message.event.operation.toLowerCase()}">${message.event.operation}</span>
                Partition: ${message.partition} | ID: ${message.event.record.id}
            </div>
        `;
        container.appendChild(item);
    });
}

function renderSearchIndex() {
    const container = document.getElementById('search-index');
    container.innerHTML = '';
    
    if (!systemState.searchIndex || systemState.searchIndex.length === 0) {
        container.innerHTML = '<div style="color: var(--text-muted); padding: 10px;">Empty</div>';
        return;
    }
    
    // Sort by ID to prevent jumping around
    const sortedIndex = [...systemState.searchIndex].sort((a, b) => a.id - b.id);
    
    sortedIndex.forEach(entry => {
        const item = document.createElement('div');
        item.className = 'index-item';
        item.innerHTML = `
            <div style="display: flex; justify-content: space-between; align-items: center;">
                <span style="font-weight: 700; color: var(--accent-blue);">ID: ${entry.id}</span>
                <span class="record-status ${entry.status}">${entry.status}</span>
            </div>
            <div style="color: var(--text-secondary); font-size: 0.8rem; margin-top: 4px;">
                ${entry.name}<br>
                ${entry.email}
            </div>
        `;
        container.appendChild(item);
    });
}

function renderCache() {
    const container = document.getElementById('cache');
    container.innerHTML = '';
    
    if (!systemState.cache || systemState.cache.length === 0) {
        container.innerHTML = '<div style="color: var(--text-muted); padding: 10px;">Empty</div>';
        return;
    }
    
    // Sort by key to prevent jumping around
    const sortedCache = [...systemState.cache].sort((a, b) => a.key.localeCompare(b.key));
    
    sortedCache.forEach(entry => {
        const item = document.createElement('div');
        item.className = 'cache-item';
        item.innerHTML = `
            <div style="display: flex; justify-content: space-between; align-items: center;">
                <span style="font-weight: 700; color: var(--accent-green);">${entry.key}</span>
                <span style="font-size: 0.7rem; color: var(--text-muted);">TTL: ${entry.ttl}s</span>
            </div>
            <div style="color: var(--text-secondary); font-size: 0.8rem; margin-top: 4px;">
                ${entry.value.name}<br>
                ${entry.value.email}
            </div>
        `;
        container.appendChild(item);
    });
}

function renderStatistics() {
    if (!systemState.stats) return;
    
    document.getElementById('stat-total').textContent = systemState.stats.totalChanges;
    document.getElementById('stat-processed').textContent = systemState.stats.changesProcessed;
    document.getElementById('stat-messages').textContent = systemState.stats.messagesSent;
    document.getElementById('stat-search').textContent = systemState.stats.searchIndexed;
    document.getElementById('stat-cache').textContent = systemState.stats.cacheUpdated;
    document.getElementById('stat-latency').textContent = `${systemState.stats.averageLatencyMs.toFixed(1)}ms`;
}

function insertRecord() {
    const name = document.getElementById('name-input').value.trim();
    const email = document.getElementById('email-input').value.trim();
    const status = document.getElementById('status-select').value;
    
    if (!name || !email) {
        showFeedback('Name and email are required', 'error');
        return;
    }
    
    fetch(`${API_BASE}/insert?name=${encodeURIComponent(name)}&email=${encodeURIComponent(email)}&status=${status}`, 
          addSessionHeader({ method: 'POST' }))
        .then(response => response.json())
        .then(data => {
            systemState = data;
            renderSystemState();
            showFeedback(`✅ Inserted: ${name}`, 'success');
            // Clear inputs
            document.getElementById('name-input').value = '';
            document.getElementById('email-input').value = '';
        })
        .catch(error => {
            console.error('Error inserting record:', error);
            showFeedback('Error inserting record', 'error');
        });
}

function updateRecord() {
    const id = document.getElementById('update-id-input').value;
    const name = document.getElementById('update-name-input').value.trim();
    const email = document.getElementById('update-email-input').value.trim();
    const status = document.getElementById('update-status-select').value;
    
    if (!id || !name || !email) {
        showFeedback('ID, name, and email are required', 'error');
        return;
    }
    
    fetch(`${API_BASE}/update?id=${id}&name=${encodeURIComponent(name)}&email=${encodeURIComponent(email)}&status=${status}`, 
          addSessionHeader({ method: 'POST' }))
        .then(response => {
            if (!response.ok) throw new Error('Record not found');
            return response.json();
        })
        .then(data => {
            systemState = data;
            renderSystemState();
            showFeedback(`✅ Updated ID ${id}`, 'success');
            // Clear inputs
            document.getElementById('update-id-input').value = '';
            document.getElementById('update-name-input').value = '';
            document.getElementById('update-email-input').value = '';
        })
        .catch(error => {
            console.error('Error updating record:', error);
            showFeedback('Error updating record (check ID)', 'error');
        });
}

function deleteRecord() {
    const id = document.getElementById('delete-id-input').value;
    
    if (!id) {
        showFeedback('ID is required', 'error');
        return;
    }
    
    fetch(`${API_BASE}/delete?id=${id}`, addSessionHeader({ method: 'POST' }))
        .then(response => {
            if (!response.ok) throw new Error('Record not found');
            return response.json();
        })
        .then(data => {
            systemState = data;
            renderSystemState();
            showFeedback(`✅ Deleted ID ${id}`, 'success');
            document.getElementById('delete-id-input').value = '';
        })
        .catch(error => {
            console.error('Error deleting record:', error);
            showFeedback('Error deleting record (check ID)', 'error');
        });
}

function streamToKafka() {
    fetch(`${API_BASE}/stream-to-kafka`, addSessionHeader({ method: 'POST' }))
        .then(response => response.json())
        .then(data => {
            systemState = data;
            renderSystemState();
            showFeedback('📤 Changes streamed to Kafka', 'success');
        })
        .catch(error => {
            console.error('Error streaming to Kafka:', error);
            showFeedback('Error streaming to Kafka', 'error');
        });
}

function consumeFromKafka() {
    fetch(`${API_BASE}/consume-from-kafka`, addSessionHeader({ method: 'POST' }))
        .then(response => response.json())
        .then(data => {
            systemState = data;
            renderSystemState();
            showFeedback('📥 Kafka messages consumed and applied', 'success');
        })
        .catch(error => {
            console.error('Error consuming from Kafka:', error);
            showFeedback('Error consuming from Kafka', 'error');
        });
}

function resetSystem() {
    if (!confirm('Reset the entire CDC system?')) return;
    
    fetch(`${API_BASE}/reset`, addSessionHeader({ method: 'POST' }))
        .then(response => response.json())
        .then(data => {
            systemState = data;
            renderSystemState();
            showFeedback('🔄 System reset', 'success');
        })
        .catch(error => {
            console.error('Error resetting system:', error);
            showFeedback('Error resetting system', 'error');
        });
}

function showFeedback(message, type) {
    const feedback = document.getElementById('feedback');
    feedback.textContent = message;
    feedback.className = `feedback ${type}`;
    
    setTimeout(() => {
        feedback.textContent = '';
        feedback.className = 'feedback';
    }, 3000);
}

document.addEventListener('DOMContentLoaded', initVisualization);

