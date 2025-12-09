// REST API Simulator

// Initialize theme immediately
(function() {
    const savedTheme = localStorage.getItem('theme') || 'dark';
    document.documentElement.setAttribute('data-theme', savedTheme);
})();

const API_BASE = 'http://localhost:8080/api/restapi';

// Helper function to add session header to fetch options
function addSessionHeader(options = {}) {
    if (typeof window.SDS_SESSION !== 'undefined') {
        return window.SDS_SESSION.addSessionHeader(options);
    }
    return options;
}

let currentState = null;
let currentResourceType = 'users';

// Initialize visualization
function initVisualization() {
    console.log('🔧 Initializing REST API visualization...');
    loadState();
    setupEventListeners();
    setupExamples();
}

function setupEventListeners() {
    // Send button
    document.getElementById('send-btn').addEventListener('click', sendRequest);
    
    // Enter key in URL input
    document.getElementById('url-input').addEventListener('keypress', (e) => {
        if (e.key === 'Enter') {
            sendRequest();
        }
    });
    
    // Reset button
    document.getElementById('reset-btn').addEventListener('click', reset);
    
    // Tab switching
    document.querySelectorAll('.tab-btn').forEach(btn => {
        btn.addEventListener('click', () => switchTab(btn.dataset.tab));
    });
    
    // Resource tab switching
    document.querySelectorAll('.resource-tab-btn').forEach(btn => {
        btn.addEventListener('click', () => switchResourceTab(btn.dataset.resource));
    });
    
    // Add row buttons
    document.querySelectorAll('.add-row-btn').forEach(btn => {
        btn.addEventListener('click', () => addKeyValueRow(btn.dataset.editor));
    });
    
    // Body format selector
    document.querySelectorAll('input[name="body-format"]').forEach(radio => {
        radio.addEventListener('change', (e) => {
            const bodyEditor = document.getElementById('body-editor');
            bodyEditor.disabled = e.target.value === 'none';
            if (e.target.value === 'none') {
                bodyEditor.value = '';
            }
        });
    });
    
    // Body editor validation
    document.getElementById('body-editor').addEventListener('input', validateJSON);
    
    // Method change - show/hide body tab
    document.getElementById('method-select').addEventListener('change', (e) => {
        const method = e.target.value;
        const bodyTab = document.querySelector('.tab-btn[data-tab="body"]');
        if (method === 'GET' || method === 'DELETE') {
            bodyTab.style.display = 'none';
            if (bodyTab.classList.contains('active')) {
                switchTab('headers');
            }
        } else {
            bodyTab.style.display = 'inline-block';
        }
    });
    
    // Theme toggle
    const themeToggle = document.getElementById('theme-toggle');
    if (themeToggle) {
        themeToggle.addEventListener('click', toggleTheme);
    }
}

function setupExamples() {
    const examples = {
        'get-users': {
            method: 'GET',
            url: '/users',
            headers: {'Content-Type': 'application/json'},
            body: null
        },
        'get-user': {
            method: 'GET',
            url: '/users/1',
            headers: {'Content-Type': 'application/json'},
            body: null
        },
        'post-user': {
            method: 'POST',
            url: '/users',
            headers: {'Content-Type': 'application/json'},
            body: {
                name: 'New User',
                email: 'newuser@example.com',
                role: 'user'
            }
        },
        'put-user': {
            method: 'PUT',
            url: '/users/1',
            headers: {'Content-Type': 'application/json'},
            body: {
                name: 'Alice Updated',
                email: 'alice.updated@example.com'
            }
        },
        'patch-user': {
            method: 'PATCH',
            url: '/users/1',
            headers: {'Content-Type': 'application/json'},
            body: {
                role: 'superadmin'
            }
        },
        'delete-user': {
            method: 'DELETE',
            url: '/users/1',
            headers: {'Content-Type': 'application/json'},
            body: null
        }
    };
    
    document.querySelectorAll('.example-btn').forEach(btn => {
        btn.addEventListener('click', () => {
            const example = examples[btn.dataset.example];
            loadExample(example);
        });
    });
}


function loadExample(example) {
    document.getElementById('method-select').value = example.method;
    document.getElementById('url-input').value = example.url;
    
    // Load headers
    const headersEditor = document.getElementById('headers-editor');
    headersEditor.innerHTML = '';
    Object.entries(example.headers).forEach(([key, value]) => {
        addKeyValueRow('headers', key, value);
    });
    
    // Load body
    if (example.body) {
        document.querySelector('input[name="body-format"][value="json"]').checked = true;
        document.getElementById('body-editor').disabled = false;
        document.getElementById('body-editor').value = JSON.stringify(example.body, null, 2);
    } else {
        document.querySelector('input[name="body-format"][value="none"]').checked = true;
        document.getElementById('body-editor').disabled = true;
        document.getElementById('body-editor').value = '';
    }
    
    // Trigger method change to show/hide body tab
    document.getElementById('method-select').dispatchEvent(new Event('change'));
}

function switchTab(tabName) {
    // Update buttons
    document.querySelectorAll('.tab-btn').forEach(btn => {
        btn.classList.toggle('active', btn.dataset.tab === tabName);
    });
    
    // Update content
    document.querySelectorAll('.tab-content').forEach(content => {
        content.classList.toggle('active', content.id === `${tabName}-tab`);
    });
}

function switchResourceTab(resourceType) {
    currentResourceType = resourceType;
    
    // Update buttons
    document.querySelectorAll('.resource-tab-btn').forEach(btn => {
        btn.classList.toggle('active', btn.dataset.resource === resourceType);
    });
    
    renderResources();
}

function addKeyValueRow(editorType, key = '', value = '') {
    const editor = document.getElementById(`${editorType}-editor`);
    const row = document.createElement('div');
    row.className = 'kv-row';
    
    row.innerHTML = `
        <input type="text" class="kv-key" placeholder="${editorType === 'headers' ? 'Header Name' : 'Param Name'}" value="${key}" />
        <input type="text" class="kv-value" placeholder="${editorType === 'headers' ? 'Header Value' : 'Param Value'}" value="${value}" />
        <button class="kv-remove">×</button>
    `;
    
    row.querySelector('.kv-remove').addEventListener('click', () => row.remove());
    
    editor.appendChild(row);
}

function validateJSON() {
    const editor = document.getElementById('body-editor');
    const validation = document.getElementById('body-validation');
    
    if (!editor.value.trim()) {
        validation.textContent = '';
        validation.className = 'validation-message';
        return true;
    }
    
    try {
        JSON.parse(editor.value);
        validation.textContent = '✓ Valid JSON';
        validation.className = 'validation-message valid';
        return true;
    } catch (e) {
        validation.textContent = `✗ Invalid JSON: ${e.message}`;
        validation.className = 'validation-message invalid';
        return false;
    }
}

function loadState() {
    fetch(`${API_BASE}/state`, addSessionHeader())
        .then(response => response.json())
        .then(data => {
            currentState = data;
            renderResources();
            renderHistory();
        })
        .catch(error => {
            console.error('Error loading state:', error);
            showNotification('Error connecting to backend. Make sure it\'s running on port 8080.', 'error');
        });
}

function sendRequest() {
    const method = document.getElementById('method-select').value;
    const path = document.getElementById('url-input').value;
    
    if (!path.trim()) {
        showNotification('Please enter a URL', 'error');
        return;
    }
    
    // Collect headers
    const headers = {};
    document.querySelectorAll('#headers-editor .kv-row').forEach(row => {
        const key = row.querySelector('.kv-key').value.trim();
        const value = row.querySelector('.kv-value').value.trim();
        if (key) headers[key] = value;
    });
    
    // Collect query params
    const query = {};
    document.querySelectorAll('#query-editor .kv-row').forEach(row => {
        const key = row.querySelector('.kv-key').value.trim();
        const value = row.querySelector('.kv-value').value.trim();
        if (key) query[key] = value;
    });
    
    // Collect body
    let body = null;
    const bodyFormat = document.querySelector('input[name="body-format"]:checked').value;
    if (bodyFormat === 'json') {
        const bodyText = document.getElementById('body-editor').value.trim();
        if (bodyText) {
            if (!validateJSON()) {
                showNotification('Invalid JSON in request body', 'error');
                return;
            }
            body = JSON.parse(bodyText);
        }
    }
    
    const request = {
        method,
        path,
        headers,
        query,
        body
    };
    
    // Determine if it's a real or simulated request based on URL
    const isRealAPI = path.startsWith('http://') || path.startsWith('https://');
    showNotification(`Sending request...`, 'info');
    
    fetch(`${API_BASE}/request`, addSessionHeader({
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify(request)
    }))
        .then(response => response.json())
        .then(data => {
            currentState = data.state;
            renderResponse(data.response);
            
            // Only render resources for simulated API
            if (!isRealAPI) {
                renderResources();
            }
            
            renderHistory();
            
            const statusEmoji = data.response.statusCode < 400 ? '✅' : '❌';
            showNotification(`${statusEmoji} ${data.response.statusCode} ${data.response.statusText} (${data.response.latency}ms)`, 
                data.response.statusCode < 400 ? 'success' : 'error');
        })
        .catch(error => {
            console.error('Error sending request:', error);
            showNotification('Error sending request: ' + error.message, 'error');
        });
}

function renderResponse(response) {
    const container = document.getElementById('response-container');
    container.innerHTML = '';
    
    // Status line
    const statusLine = document.createElement('div');
    statusLine.className = `status-line status-${Math.floor(response.statusCode / 100)}xx`;
    statusLine.innerHTML = `
        <span class="status-code">${response.statusCode}</span>
        <span class="status-text">${response.statusText}</span>
        <span class="latency">${response.latency}ms</span>
    `;
    container.appendChild(statusLine);
    
    // Headers
    const headersSection = document.createElement('div');
    headersSection.className = 'response-section';
    headersSection.innerHTML = '<h3>Headers</h3>';
    const headersList = document.createElement('div');
    headersList.className = 'headers-list';
    Object.entries(response.headers).forEach(([key, value]) => {
        const headerItem = document.createElement('div');
        headerItem.className = 'header-item';
        headerItem.innerHTML = `<span class="header-key">${key}:</span> <span class="header-value">${value}</span>`;
        headersList.appendChild(headerItem);
    });
    headersSection.appendChild(headersList);
    container.appendChild(headersSection);
    
    // Body
    if (response.body !== null && response.body !== undefined) {
        const bodySection = document.createElement('div');
        bodySection.className = 'response-section';
        bodySection.innerHTML = '<h3>Body</h3>';
        const bodyContent = document.createElement('pre');
        bodyContent.className = 'response-body';
        bodyContent.textContent = JSON.stringify(response.body, null, 2);
        bodySection.appendChild(bodyContent);
        container.appendChild(bodySection);
    }
}

function renderResources() {
    if (!currentState) return;
    
    const container = document.getElementById('resources-container');
    container.innerHTML = '';
    
    const resources = currentState.resources[currentResourceType];
    if (!resources || Object.keys(resources).length === 0) {
        container.innerHTML = '<div class="empty-state">No resources yet</div>';
        return;
    }
    
    const resourceList = document.createElement('div');
    resourceList.className = 'resource-list';
    
    Object.values(resources).forEach(resource => {
        const item = document.createElement('div');
        item.className = 'resource-item';
        
        const header = document.createElement('div');
        header.className = 'resource-header';
        header.innerHTML = `
            <span class="resource-id">ID: ${resource.id}</span>
            <span class="resource-type">${resource.type}</span>
        `;
        
        const data = document.createElement('pre');
        data.className = 'resource-data';
        data.textContent = JSON.stringify(resource.data, null, 2);
        
        item.appendChild(header);
        item.appendChild(data);
        resourceList.appendChild(item);
    });
    
    container.appendChild(resourceList);
}

function renderHistory() {
    if (!currentState) return;
    
    const container = document.getElementById('history-container');
    container.innerHTML = '';
    
    if (currentState.requestHistory.length === 0) {
        container.innerHTML = '<div class="empty-state">No requests yet. Make a request to see history.</div>';
        return;
    }
    
    const historyList = document.createElement('div');
    historyList.className = 'history-list';
    
    // Show last 10 requests
    const recentHistory = currentState.requestHistory.slice(-10).reverse();
    
    recentHistory.forEach(item => {
        const historyItem = document.createElement('div');
        historyItem.className = 'history-item';
        
        const methodClass = item.request.method.toLowerCase();
        const statusClass = `status-${Math.floor(item.response.statusCode / 100)}xx`;
        
        historyItem.innerHTML = `
            <div class="history-header">
                <span class="method-badge ${methodClass}">${item.request.method}</span>
                <span class="history-path">${item.request.path}</span>
                <span class="history-status ${statusClass}">${item.response.statusCode}</span>
                <span class="history-latency">${item.response.latency}ms</span>
            </div>
        `;
        
        // Click to load request
        historyItem.addEventListener('click', () => {
            loadExample({
                method: item.request.method,
                url: item.request.path,
                headers: item.request.headers,
                body: item.request.body
            });
        });
        
        historyList.appendChild(historyItem);
    });
    
    container.appendChild(historyList);
}

function reset() {
    if (!confirm('Reset REST API to initial state? This will delete all created resources.')) return;
    
    fetch(`${API_BASE}/reset`, addSessionHeader({ method: 'POST' }))
        .then(response => response.json())
        .then(data => {
            currentState = data;
            renderResources();
            renderHistory();
            document.getElementById('response-container').innerHTML = '<div class="empty-response"><p>👆 Make a request to see the response</p></div>';
            showNotification('✅ REST API reset to initial state', 'success');
        })
        .catch(error => {
            console.error('Error resetting:', error);
            showNotification('Error resetting API', 'error');
        });
}

function showNotification(message, type = 'info') {
    // Create notification element
    const notification = document.createElement('div');
    notification.className = `notification ${type}`;
    notification.textContent = message;
    
    // Add to body
    document.body.appendChild(notification);
    
    // Show with animation
    setTimeout(() => notification.classList.add('show'), 10);
    
    // Remove after 3 seconds
    setTimeout(() => {
        notification.classList.remove('show');
        setTimeout(() => notification.remove(), 300);
    }, 3000);
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

