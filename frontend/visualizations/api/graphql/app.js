// GraphQL Simulator Frontend
const API_BASE = 'http://localhost:8080/api/graphql';

let currentState = null;
let currentDataType = 'users';

// Initialize on page load
document.addEventListener('DOMContentLoaded', () => {
    initTheme();
    initVisualization();
    setupEventListeners();
});

// Theme management
function initTheme() {
    const savedTheme = localStorage.getItem('theme') || 'dark';
    document.documentElement.setAttribute('data-theme', savedTheme);
    updateThemeIcon(savedTheme);
}

function updateThemeIcon(theme) {
    const icon = document.querySelector('.theme-icon');
    icon.textContent = theme === 'dark' ? '🌙' : '☀️';
}

document.getElementById('theme-toggle')?.addEventListener('click', () => {
    const current = document.documentElement.getAttribute('data-theme');
    const next = current === 'dark' ? 'light' : 'dark';
    document.documentElement.setAttribute('data-theme', next);
    localStorage.setItem('theme', next);
    updateThemeIcon(next);
});

// Initialize visualization
async function initVisualization() {
    await fetchState();
    renderCurrentData();
    renderHistory();
}

// Setup event listeners
function setupEventListeners() {
    // Execute query button
    document.getElementById('execute-btn').addEventListener('click', executeQuery);
    
    // Reset button
    document.getElementById('reset-btn').addEventListener('click', resetAPI);
    
    // Example buttons
    document.querySelectorAll('.example-btn').forEach(btn => {
        btn.addEventListener('click', (e) => {
            const example = e.target.dataset.example;
            loadExample(example);
        });
    });
    
    // Data tabs
    document.querySelectorAll('.data-tab-btn').forEach(btn => {
        btn.addEventListener('click', (e) => {
            document.querySelectorAll('.data-tab-btn').forEach(b => b.classList.remove('active'));
            e.target.classList.add('active');
            currentDataType = e.target.dataset.type;
            renderCurrentData();
        });
    });
    
    // Real-time validation
    document.getElementById('variables-editor').addEventListener('input', validateVariables);
}

// Fetch current state
async function fetchState() {
    try {
        const response = await fetch(`${API_BASE}/state`, addSessionHeader());
        
        if (!response.ok) {
            throw new Error(`HTTP ${response.status}`);
        }
        
        currentState = await response.json();
        return currentState;
    } catch (error) {
        console.error('Failed to fetch state:', error);
        showError('Failed to connect to API');
    }
}

// Execute GraphQL query
async function executeQuery() {
    const queryEditor = document.getElementById('query-editor');
    const variablesEditor = document.getElementById('variables-editor');
    const endpointInput = document.getElementById('endpoint-input');
    const executeBtn = document.getElementById('execute-btn');
    
    const query = queryEditor.value.trim();
    const endpoint = endpointInput.value.trim();
    
    if (!query) {
        showValidationError('query-validation', 'Query cannot be empty');
        return;
    }
    
    // Parse variables
    let variables = {};
    const variablesText = variablesEditor.value.trim();
    if (variablesText) {
        try {
            variables = JSON.parse(variablesText);
        } catch (error) {
            showValidationError('variables-validation', 'Invalid JSON: ' + error.message);
            return;
        }
    }
    
    // Clear validation messages
    clearValidationError('query-validation');
    clearValidationError('variables-validation');
    
    // Show loading state
    executeBtn.disabled = true;
    executeBtn.textContent = 'Executing...';
    
    try {
        const requestBody = {
            query,
            variables,
        };
        
        // Add endpoint if provided
        if (endpoint) {
            requestBody.endpoint = endpoint;
        }
        
        const response = await fetch(`${API_BASE}/query`, addSessionHeader({
            method: 'POST',
            headers: {
                'Content-Type': 'application/json'
            },
            body: JSON.stringify(requestBody)
        }));
        
        if (!response.ok) {
            throw new Error(`HTTP ${response.status}`);
        }
        
        const result = await response.json();
        
        // Update state if simulated API
        if (!endpoint) {
            currentState = result.state;
            renderCurrentData();
        }
        
        // Render response
        renderResponse(result.response, result.latency);
        
        // Update history if simulated API
        if (!endpoint && currentState) {
            renderHistory();
        }
        
    } catch (error) {
        console.error('Query execution failed:', error);
        showError('Query execution failed: ' + error.message);
    } finally {
        executeBtn.disabled = false;
        executeBtn.textContent = 'Execute Query';
    }
}

// Render GraphQL response
function renderResponse(response, latency) {
    const container = document.getElementById('response-container');
    container.innerHTML = '';
    
    // Create response header
    const header = document.createElement('div');
    header.className = 'response-header';
    
    const statusBadge = document.createElement('span');
    statusBadge.className = response.errors && response.errors.length > 0 ? 'status-badge error' : 'status-badge success';
    statusBadge.textContent = response.errors && response.errors.length > 0 ? 'Error' : 'Success';
    
    const latencyBadge = document.createElement('span');
    latencyBadge.className = 'latency-badge';
    latencyBadge.textContent = `${latency}ms`;
    
    header.appendChild(statusBadge);
    header.appendChild(latencyBadge);
    container.appendChild(header);
    
    // Create response body
    const body = document.createElement('pre');
    body.className = 'response-body';
    body.textContent = JSON.stringify(response, null, 2);
    container.appendChild(body);
}

// Render current data
function renderCurrentData() {
    if (!currentState) return;
    
    const container = document.getElementById('data-container');
    container.innerHTML = '';
    
    if (currentDataType === 'users') {
        renderUsers(container);
    } else if (currentDataType === 'posts') {
        renderPosts(container);
    }
}

// Render users table
function renderUsers(container) {
    if (!currentState.users || Object.keys(currentState.users).length === 0) {
        container.innerHTML = '<div class="empty-state">No users available</div>';
        return;
    }
    
    const table = document.createElement('table');
    table.className = 'data-table';
    
    // Header
    const thead = document.createElement('thead');
    thead.innerHTML = `
        <tr>
            <th>ID</th>
            <th>Name</th>
            <th>Email</th>
            <th>Role</th>
        </tr>
    `;
    table.appendChild(thead);
    
    // Body
    const tbody = document.createElement('tbody');
    Object.values(currentState.users).forEach(user => {
        const row = document.createElement('tr');
        row.innerHTML = `
            <td>${user.id}</td>
            <td>${user.name}</td>
            <td>${user.email || '-'}</td>
            <td><span class="role-badge">${user.role || 'user'}</span></td>
        `;
        tbody.appendChild(row);
    });
    table.appendChild(tbody);
    
    container.appendChild(table);
}

// Render posts table
function renderPosts(container) {
    if (!currentState.posts || Object.keys(currentState.posts).length === 0) {
        container.innerHTML = '<div class="empty-state">No posts available</div>';
        return;
    }
    
    const table = document.createElement('table');
    table.className = 'data-table';
    
    // Header
    const thead = document.createElement('thead');
    thead.innerHTML = `
        <tr>
            <th>ID</th>
            <th>Title</th>
            <th>Body</th>
            <th>Author ID</th>
        </tr>
    `;
    table.appendChild(thead);
    
    // Body
    const tbody = document.createElement('tbody');
    Object.values(currentState.posts).forEach(post => {
        const row = document.createElement('tr');
        row.innerHTML = `
            <td>${post.id}</td>
            <td>${post.title}</td>
            <td>${post.body.substring(0, 50)}${post.body.length > 50 ? '...' : ''}</td>
            <td>${post.authorId}</td>
        `;
        tbody.appendChild(row);
    });
    table.appendChild(tbody);
    
    container.appendChild(table);
}

// Render query history
function renderHistory() {
    if (!currentState || !currentState.queryHistory) return;
    
    const container = document.getElementById('history-container');
    
    if (currentState.queryHistory.length === 0) {
        container.innerHTML = '<div class="empty-state">No queries executed yet</div>';
        return;
    }
    
    container.innerHTML = '';
    
    // Show last 10 queries
    const recentQueries = currentState.queryHistory.slice(-10).reverse();
    
    recentQueries.forEach((item, index) => {
        const historyItem = document.createElement('div');
        historyItem.className = 'history-item';
        
        const header = document.createElement('div');
        header.className = 'history-header';
        
        const timestamp = new Date(item.timestamp).toLocaleTimeString();
        const queryType = item.request.query.trim().startsWith('mutation') ? 'Mutation' : 'Query';
        
        const statusBadge = document.createElement('span');
        statusBadge.className = item.isError ? 'status-badge error' : 'status-badge success';
        statusBadge.textContent = item.isError ? 'Error' : 'Success';
        
        const typeBadge = document.createElement('span');
        typeBadge.className = 'type-badge';
        typeBadge.textContent = queryType;
        
        const latencyBadge = document.createElement('span');
        latencyBadge.className = 'latency-badge';
        latencyBadge.textContent = `${item.latency}ms`;
        
        const timeSpan = document.createElement('span');
        timeSpan.className = 'history-time';
        timeSpan.textContent = timestamp;
        
        header.appendChild(statusBadge);
        header.appendChild(typeBadge);
        header.appendChild(latencyBadge);
        header.appendChild(timeSpan);
        
        const queryPreview = document.createElement('div');
        queryPreview.className = 'query-preview';
        queryPreview.textContent = item.request.query.split('\n')[0].substring(0, 80) + '...';
        
        historyItem.appendChild(header);
        historyItem.appendChild(queryPreview);
        
        container.appendChild(historyItem);
    });
}

// Load example query
function loadExample(example) {
    const queryEditor = document.getElementById('query-editor');
    const variablesEditor = document.getElementById('variables-editor');
    const endpointInput = document.getElementById('endpoint-input');
    
    // Clear endpoint for examples (use simulated API)
    endpointInput.value = '';
    
    const examples = {
        'get-users': {
            query: `query {
  users {
    id
    name
    email
    role
  }
}`,
            variables: ''
        },
        'get-user': {
            query: `query GetUser($userId: ID!) {
  user(id: $userId) {
    id
    name
    email
    role
  }
}`,
            variables: `{
  "userId": 1
}`
        },
        'get-posts': {
            query: `query {
  posts {
    id
    title
    body
    author {
      id
      name
      email
    }
  }
}`,
            variables: ''
        },
        'create-user': {
            query: `mutation CreateUser($input: CreateUserInput!) {
  createUser(input: $input) {
    id
    name
    email
    role
  }
}`,
            variables: `{
  "input": {
    "name": "New User",
    "email": "newuser@example.com",
    "role": "user"
  }
}`
        },
        'update-user': {
            query: `mutation UpdateUser($userId: ID!, $input: UpdateUserInput!) {
  updateUser(id: $userId, input: $input) {
    id
    name
    email
    role
  }
}`,
            variables: `{
  "userId": 1,
  "input": {
    "name": "Updated Name"
  }
}`
        },
        'delete-user': {
            query: `mutation DeleteUser($userId: ID!) {
  deleteUser(id: $userId) {
    success
    id
  }
}`,
            variables: `{
  "userId": 3
}`
        }
    };
    
    const example_data = examples[example];
    if (example_data) {
        queryEditor.value = example_data.query;
        variablesEditor.value = example_data.variables;
        clearValidationError('query-validation');
        clearValidationError('variables-validation');
    }
}

// Reset API data
async function resetAPI() {
    if (!confirm('Reset all API data to initial state?')) {
        return;
    }
    
    try {
        const response = await fetch(`${API_BASE}/reset`, addSessionHeader({
            method: 'POST'
        }));
        
        if (!response.ok) {
            throw new Error(`HTTP ${response.status}`);
        }
        
        currentState = await response.json();
        renderCurrentData();
        renderHistory();
        
        // Clear response
        const container = document.getElementById('response-container');
        container.innerHTML = '<div class="empty-response"><p>API data has been reset</p></div>';
        
    } catch (error) {
        console.error('Reset failed:', error);
        showError('Reset failed: ' + error.message);
    }
}

// Validate variables JSON
function validateVariables() {
    const editor = document.getElementById('variables-editor');
    const text = editor.value.trim();
    
    if (!text) {
        clearValidationError('variables-validation');
        return;
    }
    
    try {
        JSON.parse(text);
        clearValidationError('variables-validation');
    } catch (error) {
        showValidationError('variables-validation', 'Invalid JSON: ' + error.message);
    }
}

// Show validation error
function showValidationError(elementId, message) {
    const element = document.getElementById(elementId);
    element.textContent = '⚠️ ' + message;
    element.style.display = 'block';
}

// Clear validation error
function clearValidationError(elementId) {
    const element = document.getElementById(elementId);
    element.textContent = '';
    element.style.display = 'none';
}

// Show error message
function showError(message) {
    const container = document.getElementById('response-container');
    container.innerHTML = `
        <div class="error-message">
            <strong>Error:</strong> ${message}
        </div>
    `;
}

