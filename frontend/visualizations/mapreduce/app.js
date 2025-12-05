// Theme management
(function() {
    const savedTheme = localStorage.getItem('theme') || 'dark';
    document.documentElement.setAttribute('data-theme', savedTheme);
})();

const API_BASE = 'http://localhost:8080/api/mapreduce';

function addSessionHeader(options = {}) {
    if (typeof window.SDS_SESSION !== 'undefined') {
        return window.SDS_SESSION.addSessionHeader(options);
    }
    return options;
}

let jobState = null;

function initVisualization() {
    console.log('🔧 Initializing MapReduce visualization...');
    
    // Setup event listeners
    document.getElementById('start-btn').addEventListener('click', startJob);
    document.getElementById('map-btn').addEventListener('click', executeMapPhase);
    document.getElementById('shuffle-btn').addEventListener('click', executeShufflePhase);
    document.getElementById('reduce-btn').addEventListener('click', executeReducePhase);
    document.getElementById('reset-btn').addEventListener('click', resetJob);
    
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
            if (jobState) renderJobState();
        });
        const currentTheme = document.documentElement.getAttribute('data-theme');
        const icon = themeToggle.querySelector('.theme-icon');
        if (icon) icon.textContent = currentTheme === 'dark' ? '🌙' : '☀️';
    }
    
    loadJobState();
}

function loadJobState() {
    fetch(`${API_BASE}/state`, addSessionHeader())
        .then(response => response.json())
        .then(data => {
            jobState = data;
            renderJobState();
            updateButtons();
        })
        .catch(error => {
            console.error('Error loading job state:', error);
        });
}

function renderJobState() {
    if (!jobState) return;
    
    // Update status
    document.getElementById('job-status').textContent = jobState.status;
    document.getElementById('current-stage').textContent = jobState.currentStage;
    document.getElementById('progress').textContent = `${jobState.progress}%`;
    document.getElementById('progress-fill').style.width = `${jobState.progress}%`;
    
    // Render input data
    renderInputData();
    
    // Render map tasks
    renderMapTasks();
    
    // Render shuffle data
    renderShuffleData();
    
    // Render reduce tasks
    renderReduceTasks();
    
    // Render final output
    renderFinalOutput();
}

function renderInputData() {
    const container = document.getElementById('input-data');
    container.innerHTML = '';
    
    if (!jobState.inputData || jobState.inputData.length === 0) {
        container.innerHTML = '<div style="color: var(--text-muted); padding: 10px;">No input data</div>';
        return;
    }
    
    jobState.inputData.forEach((data, index) => {
        const item = document.createElement('div');
        item.className = 'data-item';
        item.textContent = `Split ${index + 1}: "${data}"`;
        container.appendChild(item);
    });
}

function renderMapTasks() {
    const container = document.getElementById('map-workers');
    container.innerHTML = '';
    
    if (!jobState.mapTasks || jobState.mapTasks.length === 0) {
        container.innerHTML = '<div style="color: var(--text-muted); padding: 10px;">No map tasks yet</div>';
        return;
    }
    
    // Group tasks by worker
    const workerTasks = {};
    jobState.mapTasks.forEach(task => {
        if (!workerTasks[task.workerId]) {
            workerTasks[task.workerId] = [];
        }
        workerTasks[task.workerId].push(task);
    });
    
    // Render each worker
    Object.keys(workerTasks).sort().forEach(workerId => {
        const tasks = workerTasks[workerId];
        const workerCard = document.createElement('div');
        workerCard.className = 'worker-card';
        if (tasks.some(t => t.status === 'running')) {
            workerCard.classList.add('active');
        }
        
        const allCompleted = tasks.every(t => t.status === 'completed');
        const anyRunning = tasks.some(t => t.status === 'running');
        const status = allCompleted ? 'completed' : anyRunning ? 'running' : 'pending';
        
        let outputPairsHTML = '';
        if (allCompleted) {
            const allPairs = tasks.flatMap(t => t.outputPairs);
            outputPairsHTML = '<div class="worker-data"><strong>Output:</strong><br>';
            allPairs.forEach(pair => {
                outputPairsHTML += `<span class="kv-pair">${pair.key}: ${pair.value}</span>`;
            });
            outputPairsHTML += '</div>';
        }
        
        workerCard.innerHTML = `
            <div class="worker-header">
                <span class="worker-title">Mapper ${workerId}</span>
                <span class="worker-status ${status}">${status}</span>
            </div>
            <div class="worker-content">
                ${tasks.map(t => `<div class="worker-data"><strong>Input:</strong> "${t.inputData}"</div>`).join('')}
                ${outputPairsHTML}
            </div>
        `;
        
        container.appendChild(workerCard);
    });
}

function renderShuffleData() {
    const container = document.getElementById('shuffle-data');
    container.innerHTML = '';
    
    if (!jobState.shuffleData || jobState.shuffleData.length === 0) {
        container.innerHTML = '<div style="color: var(--text-muted); padding: 10px;">No shuffle data yet</div>';
        return;
    }
    
    jobState.shuffleData.forEach(partition => {
        const item = document.createElement('div');
        item.className = 'data-item';
        item.style.background = `hsl(${(partition.reducerId * 60) % 360}, 70%, 50%)`;
        item.textContent = `${partition.key} → [${partition.values.join(', ')}] (Reducer ${partition.reducerId})`;
        container.appendChild(item);
    });
}

function renderReduceTasks() {
    const container = document.getElementById('reduce-workers');
    container.innerHTML = '';
    
    if (!jobState.reduceTasks || jobState.reduceTasks.length === 0) {
        container.innerHTML = '<div style="color: var(--text-muted); padding: 10px;">No reduce tasks yet</div>';
        return;
    }
    
    // Group tasks by worker
    const workerTasks = {};
    jobState.reduceTasks.forEach(task => {
        if (!workerTasks[task.workerId]) {
            workerTasks[task.workerId] = [];
        }
        workerTasks[task.workerId].push(task);
    });
    
    // Render each worker
    Object.keys(workerTasks).sort().forEach(workerId => {
        const tasks = workerTasks[workerId];
        const workerCard = document.createElement('div');
        workerCard.className = 'worker-card';
        
        const allCompleted = tasks.every(t => t.status === 'completed');
        const status = allCompleted ? 'completed' : 'running';
        
        let tasksHTML = '';
        tasks.forEach(task => {
            tasksHTML += `
                <div class="worker-data">
                    <strong>${task.key}:</strong> [${task.values.join(', ')}] → <strong style="color: var(--accent-green);">${task.result}</strong>
                </div>
            `;
        });
        
        workerCard.innerHTML = `
            <div class="worker-header">
                <span class="worker-title">Reducer ${workerId}</span>
                <span class="worker-status ${status}">${status}</span>
            </div>
            <div class="worker-content">
                ${tasksHTML}
            </div>
        `;
        
        container.appendChild(workerCard);
    });
}

function renderFinalOutput() {
    const container = document.getElementById('final-output');
    container.innerHTML = '';
    
    if (!jobState.finalOutput || jobState.finalOutput.length === 0) {
        container.innerHTML = '<div style="color: var(--text-muted); padding: 10px;">No output yet</div>';
        return;
    }
    
    jobState.finalOutput.forEach(kv => {
        const item = document.createElement('div');
        item.className = 'data-item';
        item.style.background = 'var(--accent-green)';
        item.textContent = `${kv.key}: ${kv.value}`;
        container.appendChild(item);
    });
}

function updateButtons() {
    if (!jobState) return;
    
    const startBtn = document.getElementById('start-btn');
    const mapBtn = document.getElementById('map-btn');
    const shuffleBtn = document.getElementById('shuffle-btn');
    const reduceBtn = document.getElementById('reduce-btn');
    
    startBtn.disabled = jobState.status !== 'idle';
    mapBtn.disabled = jobState.status !== 'mapping';
    shuffleBtn.disabled = jobState.status !== 'shuffling';
    reduceBtn.disabled = jobState.status !== 'reducing';
}

function startJob() {
    fetch(`${API_BASE}/start`, addSessionHeader({ method: 'POST' }))
        .then(response => response.json())
        .then(data => {
            jobState = data;
            renderJobState();
            updateButtons();
        })
        .catch(error => console.error('Error starting job:', error));
}

function executeMapPhase() {
    fetch(`${API_BASE}/execute-map`, addSessionHeader({ method: 'POST' }))
        .then(response => response.json())
        .then(data => {
            jobState = data;
            renderJobState();
            updateButtons();
        })
        .catch(error => console.error('Error executing map phase:', error));
}

function executeShufflePhase() {
    fetch(`${API_BASE}/execute-shuffle`, addSessionHeader({ method: 'POST' }))
        .then(response => response.json())
        .then(data => {
            jobState = data;
            renderJobState();
            updateButtons();
        })
        .catch(error => console.error('Error executing shuffle phase:', error));
}

function executeReducePhase() {
    fetch(`${API_BASE}/execute-reduce`, addSessionHeader({ method: 'POST' }))
        .then(response => response.json())
        .then(data => {
            jobState = data;
            renderJobState();
            updateButtons();
        })
        .catch(error => console.error('Error executing reduce phase:', error));
}

function resetJob() {
    fetch(`${API_BASE}/reset`, addSessionHeader({ method: 'POST' }))
        .then(response => response.json())
        .then(data => {
            jobState = data;
            renderJobState();
            updateButtons();
        })
        .catch(error => console.error('Error resetting job:', error));
}

document.addEventListener('DOMContentLoaded', initVisualization);

