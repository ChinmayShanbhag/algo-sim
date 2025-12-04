// Theme management (initialize immediately)
(function() {
    const savedTheme = localStorage.getItem('theme') || 'dark';
    document.documentElement.setAttribute('data-theme', savedTheme);
})();

const API_BASE = 'http://localhost:8080/api/atomic-commit/2pc';

let svg;
let systemData = null;  // Contains coordinator and participants
let protocolSteps = [];
let currentStepIndex = -1;
let nodePositions = {};  // Store positions for animations
let initialSystemState = null;  // Store initial state before transaction

// Color mapping for node states (theme-aware)
function getStateColors() {
    const isDark = document.documentElement.getAttribute('data-theme') === 'dark';
    return {
        coordinator: '#8b5cf6',  // purple (coordinator always purple)
        idle: isDark ? '#64748b' : '#94a3b8',  // gray
        prepared: '#fbbf24',  // yellow
        committed: '#10b981',  // green
        aborted: '#ef4444',  // red
        failed: '#6b7280',  // dark gray
        committing: '#8b5cf6',  // purple (coordinator committing)
        aborting: '#8b5cf6'  // purple (coordinator aborting)
    };
}

// Initialize visualization
function initVisualization() {
    console.log('🔧 Initializing 2PC visualization...');
    const vizElement = document.getElementById('visualization');
    if (!vizElement) {
        console.error('❌ Visualization element not found!');
        return;
    }
    
    svg = d3.select('#visualization')
        .append('svg')
        .attr('width', 800)
        .attr('height', 600);
    
    console.log('✅ SVG created');
    
    // Show loading message
    showLoading();
    
    // Load initial state
    loadSystemState().catch(error => {
        console.error('Failed to load initial state:', error);
    });
}

function showLoading() {
    svg.selectAll('*').remove();
    svg.append('text')
        .attr('class', 'loading-text')
        .attr('x', 400)
        .attr('y', 300)
        .attr('text-anchor', 'middle')
        .attr('font-size', '18px')
        .attr('fill', getComputedStyle(document.documentElement).getPropertyValue('--text-secondary'))
        .text('Loading 2PC system state...');
}

function loadSystemState() {
    console.log('Loading system state from:', `${API_BASE}/state`);
    return fetch(`${API_BASE}/state`)
        .then(response => {
            if (!response.ok) {
                throw new Error(`HTTP error! status: ${response.status}`);
            }
            return response.json();
        })
        .then(data => {
            console.log('Received data:', data);
            systemData = data;
            renderSystem();
            hideError();
            return data;
        })
        .catch(error => {
            console.error('Error loading system state:', error);
            showError(`Cannot connect to backend: ${error.message}. Make sure the backend is running on port 8080.`);
            throw error;
        });
}

function renderSystem() {
    if (!systemData) {
        showLoading();
        return;
    }
    
    const participants = systemData.participants;
    const coordinator = systemData.coordinator;
    
    // Clear previous render
    svg.selectAll('g.node').remove();
    svg.selectAll('text.loading-text').remove();
    
    // Update participant buttons
    updateParticipantButtons(participants);
    
    const nodeRadius = 40;
    const centerX = 400;
    const centerY = 300;
    const radius = 180;
    
    // Position coordinator at center
    const coordGroup = svg.append('g')
        .attr('class', 'node coordinator-node')
        .attr('transform', `translate(${centerX}, ${centerY})`);
    
    // Store coordinator position
    nodePositions[-1] = { x: centerX, y: centerY };
    
    // Draw coordinator circle
    const colors = getStateColors();
    const isDark = document.documentElement.getAttribute('data-theme') === 'dark';
    const strokeColor = isDark ? '#475569' : '#1e293b';
    
    coordGroup.append('circle')
        .attr('r', nodeRadius + 10)  // Larger than participants
        .attr('fill', colors.coordinator)
        .attr('stroke', strokeColor)
        .attr('stroke-width', 3);
    
    // Coordinator label
    const textColor = isDark ? '#f1f5f9' : '#0f172a';
    coordGroup.append('text')
        .attr('text-anchor', 'middle')
        .attr('dy', -5)
        .attr('font-size', '14px')
        .attr('font-weight', 'bold')
        .attr('fill', textColor)
        .text('Coordinator');
    
    coordGroup.append('text')
        .attr('text-anchor', 'middle')
        .attr('dy', 15)
        .attr('font-size', '11px')
        .attr('fill', textColor)
        .text(coordinator.state);
    
    // Position participants in a circle around coordinator
    const angleStep = (2 * Math.PI) / participants.length;
    
    participants.forEach((participant, i) => {
        const angle = i * angleStep - Math.PI / 2;
        const x = centerX + radius * Math.cos(angle);
        const y = centerY + radius * Math.sin(angle);
        
        // Store position
        nodePositions[participant.id] = { x, y };
        
        const participantGroup = svg.append('g')
            .attr('class', 'node participant-node')
            .attr('transform', `translate(${x}, ${y})`);
        
        // Participant circle
        const participantColor = colors[participant.state] || colors.idle;
        
        participantGroup.append('circle')
            .attr('r', nodeRadius)
            .attr('fill', participantColor)
            .attr('stroke', strokeColor)
            .attr('stroke-width', 2.5);
        
        // Participant label
        participantGroup.append('text')
            .attr('text-anchor', 'middle')
            .attr('dy', -8)
            .attr('font-size', '13px')
            .attr('font-weight', 'bold')
            .attr('fill', textColor)
            .text(`P${participant.id}`);
        
        participantGroup.append('text')
            .attr('text-anchor', 'middle')
            .attr('dy', 10)
            .attr('font-size', '10px')
            .attr('fill', textColor)
            .text(participant.state);
        
        // Show vote if present
        if (participant.vote) {
            participantGroup.append('text')
                .attr('text-anchor', 'middle')
                .attr('dy', 25)
                .attr('font-size', '9px')
                .attr('fill', participant.vote === 'YES' ? colors.committed : colors.aborted)
                .attr('font-weight', 'bold')
                .text(participant.vote);
        }
    });
}

function updateParticipantButtons(participants) {
    const container = document.getElementById('participant-buttons');
    if (!container) return;
    
    container.innerHTML = '';
    participants.forEach(participant => {
        const btn = document.createElement('button');
        btn.className = `participant-vote-button ${participant.canCommit ? 'vote-yes' : 'vote-no'}`;
        btn.innerHTML = `
            <span class="vote-indicator"></span>
            <span>P${participant.id}: ${participant.canCommit ? 'YES' : 'NO'}</span>
        `;
        btn.onclick = () => toggleParticipantVote(participant.id, !participant.canCommit);
        container.appendChild(btn);
    });
}

function toggleParticipantVote(participantId, canCommit) {
    fetch(`${API_BASE}/set-participant-vote?participantId=${participantId}&canCommit=${canCommit}`, {
        method: 'POST'
    })
        .then(response => response.json())
        .then(data => {
            systemData = data;
            renderSystem();
            showFeedback(`Participant ${participantId} will vote ${canCommit ? 'YES' : 'NO'}`);
        })
        .catch(error => {
            console.error('Error setting participant vote:', error);
            showError('Failed to set participant vote');
        });
}

function startTransaction() {
    const transactionData = document.getElementById('transaction-data').value || 'Sample Transaction';
    
    showFeedback('Starting transaction...');
    
    // Save initial state before transaction (deep copy)
    initialSystemState = JSON.parse(JSON.stringify(systemData));
    
    fetch(`${API_BASE}/start-transaction?data=${encodeURIComponent(transactionData)}`, {
        method: 'POST'
    })
        .then(response => response.json())
        .then(data => {
            // Store the final state but don't render it yet
            const finalSystemState = data;
            protocolSteps = data.protocolSteps || [];
            currentStepIndex = -1;
            
            // Reset to initial state for stepping
            systemData = JSON.parse(JSON.stringify(initialSystemState));
            renderSystem();
            
            if (protocolSteps.length > 0) {
                showFeedback(`Transaction started with ${protocolSteps.length} steps. Use navigation to step through.`);
                updateNavigationButtons();
                showProtocolSteps(protocolSteps, -1);
                
                // Automatically go to first step
                setTimeout(() => goToStep(0), 500);
            } else {
                showFeedback('Transaction completed with no steps.');
            }
        })
        .catch(error => {
            console.error('Error starting transaction:', error);
            showError('Failed to start transaction');
        });
}

function goToStep(stepIndex) {
    if (stepIndex < 0 || stepIndex >= protocolSteps.length) return;
    
    currentStepIndex = stepIndex;
    
    // Reset to initial state
    systemData = JSON.parse(JSON.stringify(initialSystemState));
    
    // Apply all steps up to and including current step
    for (let i = 0; i <= stepIndex; i++) {
        applyStep(protocolSteps[i]);
    }
    
    // Re-render with updated state
    renderSystem();
    
    // Show current step explanation
    const step = protocolSteps[stepIndex];
    showCurrentStep(step);
    
    // Update step list highlighting
    showProtocolSteps(protocolSteps, stepIndex);
    
    // Update navigation buttons
    updateNavigationButtons();
    
    // Animate message if applicable (after render)
    setTimeout(() => {
        if (step.fromNode !== undefined && step.toNode !== undefined) {
            animateMessage(step.fromNode, step.toNode, step.messageType, step.voteResponse);
        }
    }, 100);
}

function applyStep(step) {
    if (!step || !systemData) return;
    
    const participants = systemData.participants;
    
    console.log(`Applying step ${step.stepNumber}: ${step.action}`, step);
    
    // Apply state changes based on step action
    switch (step.action) {
        case 'transaction_initiated':
            // Transaction started, no state changes yet
            break;
            
        case 'prepare_request_sent':
            // Prepare request sent to a participant, no state change yet
            break;
            
        case 'vote_received':
            // Participant voted, update their state
            if (step.fromNode !== undefined && step.fromNode >= 0) {
                const participant = participants[step.fromNode];
                if (participant) {
                    participant.vote = step.voteResponse;
                    if (step.voteResponse === 'YES') {
                        participant.state = 'prepared';  // Participant is prepared
                    } else {
                        participant.state = 'aborted';  // Participant voted NO
                    }
                }
            }
            break;
            
        case 'decision_commit':
            // Coordinator decided to commit, no participant state change yet
            systemData.coordinator.state = 'committing';
            break;
            
        case 'decision_abort':
            // Coordinator decided to abort, no participant state change yet
            systemData.coordinator.state = 'aborting';
            break;
            
        case 'commit_sent':
            // Commit message sent to participant, no state change yet
            break;
            
        case 'commit_ack':
            // Participant acknowledged commit, update their state
            if (step.fromNode !== undefined && step.fromNode >= 0) {
                const participant = participants[step.fromNode];
                if (participant) {
                    participant.state = 'committed';
                }
            }
            break;
            
        case 'abort_sent':
            // Abort message sent to participant, no state change yet
            break;
            
        case 'abort_ack':
            // Participant acknowledged abort, update their state
            if (step.fromNode !== undefined && step.fromNode >= 0) {
                const participant = participants[step.fromNode];
                if (participant) {
                    participant.state = 'aborted';
                }
            }
            break;
            
        case 'transaction_committed':
            // Transaction complete, coordinator back to idle
            systemData.coordinator.state = 'idle';
            break;
            
        case 'transaction_aborted':
            // Transaction complete, coordinator back to idle
            systemData.coordinator.state = 'idle';
            break;
    }
}

function animateMessage(fromNodeId, toNodeId, messageType, voteResponse) {
    const fromPos = nodePositions[fromNodeId];
    const toPos = nodePositions[toNodeId];
    
    if (!fromPos || !toPos) {
        console.warn(`Positions not found for nodes ${fromNodeId} -> ${toNodeId}`);
        return;
    }
    
    // Message color based on type
    const isDark = document.documentElement.getAttribute('data-theme') === 'dark';
    let messageColor;
    
    if (messageType === 'prepare') {
        messageColor = '#fbbf24';  // yellow
    } else if (messageType === 'vote') {
        messageColor = voteResponse === 'YES' ? '#10b981' : '#ef4444';  // green or red
    } else if (messageType === 'commit') {
        messageColor = '#10b981';  // green
    } else if (messageType === 'abort') {
        messageColor = '#ef4444';  // red
    } else if (messageType === 'ack') {
        messageColor = '#60a5fa';  // blue
    } else {
        messageColor = isDark ? '#60a5fa' : '#3b82f6';
    }
    
    console.log(`📨 Animating message: ${fromNodeId} -> ${toNodeId}, type: ${messageType}`);
    
    // Create message circle
    const message = svg.append('circle')
        .attr('class', 'message')
        .attr('r', 8)
        .attr('fill', messageColor)
        .attr('stroke', isDark ? '#1e293b' : '#ffffff')
        .attr('stroke-width', 2)
        .attr('cx', fromPos.x)
        .attr('cy', fromPos.y)
        .attr('opacity', 0.9)
        .style('filter', 'drop-shadow(0 0 6px ' + messageColor + ')');
    
    // Animate to destination
    message.transition()
        .duration(800)
        .ease(d3.easeLinear)
        .attr('cx', toPos.x)
        .attr('cy', toPos.y)
        .on('end', () => {
            // Arrival effect
            const arrival = svg.append('circle')
                .attr('class', 'arrival-effect')
                .attr('cx', toPos.x)
                .attr('cy', toPos.y)
                .attr('r', 15)
                .attr('fill', 'none')
                .attr('stroke', messageColor)
                .attr('stroke-width', 3)
                .attr('opacity', 0.8);
            
            arrival.transition()
                .duration(400)
                .attr('r', 30)
                .attr('opacity', 0)
                .on('end', () => arrival.remove());
            
            message.remove();
        });
}

function showCurrentStep(step) {
    const container = document.getElementById('current-step-explanation');
    if (!container) return;
    
    container.innerHTML = `
        <div class="step-detail">
            <h4>Step ${step.stepNumber}: ${step.action.replace(/_/g, ' ').toUpperCase()}</h4>
            <p>${step.description}</p>
            ${step.messageType ? `<p><strong>Message Type:</strong> ${step.messageType.toUpperCase()}</p>` : ''}
            ${step.voteResponse ? `<p><strong>Vote:</strong> ${step.voteResponse}</p>` : ''}
            <p><strong>Votes:</strong> ${step.yesVotes} YES, ${step.noVotes} NO</p>
        </div>
    `;
}

function showProtocolSteps(steps, activeIndex) {
    const container = document.getElementById('protocol-steps');
    if (!container) return;
    
    if (steps.length === 0) {
        container.innerHTML = '<p class="feedback-text">Start a transaction to see the step-by-step protocol!</p>';
        return;
    }
    
    container.innerHTML = '<ol class="steps-list">' +
        steps.map((step, idx) => `
            <li class="step-item ${idx === activeIndex ? 'active' : ''} ${idx < activeIndex ? 'completed' : ''}">
                <strong>Step ${step.stepNumber}:</strong> ${step.description}
            </li>
        `).join('') +
        '</ol>';
}

function updateNavigationButtons() {
    const prevBtn = document.getElementById('prev-step-btn');
    const nextBtn = document.getElementById('next-step-btn');
    const stepInfo = document.getElementById('step-info');
    
    if (!prevBtn || !nextBtn || !stepInfo) return;
    
    if (protocolSteps.length === 0) {
        prevBtn.disabled = true;
        nextBtn.disabled = true;
        stepInfo.textContent = 'No transaction';
        return;
    }
    
    prevBtn.disabled = currentStepIndex <= 0;
    nextBtn.disabled = currentStepIndex >= protocolSteps.length - 1;
    stepInfo.textContent = `Step ${currentStepIndex + 1} / ${protocolSteps.length}`;
}

function resetSystem() {
    // Clear animations
    svg.selectAll('circle.message').remove();
    svg.selectAll('circle.arrival-effect').remove();
    
    fetch(`${API_BASE}/reset`, { method: 'POST' })
        .then(response => response.json())
        .then(data => {
            systemData = data;
            protocolSteps = [];
            currentStepIndex = -1;
            initialSystemState = null;  // Clear initial state
            
            renderSystem();
            updateNavigationButtons();
            showProtocolSteps([], -1);
            
            document.getElementById('current-step-explanation').innerHTML = 
                '<p class="feedback-text">Click "Start Transaction" to begin a 2PC protocol.</p>';
            
            showFeedback('System reset successfully');
        })
        .catch(error => {
            console.error('Error resetting system:', error);
            showError('Failed to reset system');
        });
}

function showFeedback(message) {
    const container = document.getElementById('action-feedback');
    if (container) {
        container.innerHTML = `<p class="feedback-text">${message}</p>`;
    }
}

function showError(message) {
    const container = document.getElementById('action-feedback');
    if (container) {
        container.innerHTML = `<p class="error-text">${message}</p>`;
    }
}

function hideError() {
    // Error is shown in action-feedback, so this is handled by showFeedback
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
    // Re-render visualization with new colors
    if (systemData) {
        renderSystem();
    }
}

function updateThemeIcon(theme) {
    const themeIcon = document.querySelector('.theme-icon');
    if (themeIcon) {
        themeIcon.textContent = theme === 'dark' ? '☀️' : '🌙';
    }
}

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
    const startBtn = document.getElementById('start-transaction-btn');
    if (startBtn) {
        startBtn.addEventListener('click', startTransaction);
    }
    
    const resetBtn = document.getElementById('reset-btn');
    if (resetBtn) {
        resetBtn.addEventListener('click', resetSystem);
    }
    
    const prevBtn = document.getElementById('prev-step-btn');
    if (prevBtn) {
        prevBtn.addEventListener('click', () => {
            if (currentStepIndex > 0) {
                goToStep(currentStepIndex - 1);
            }
        });
    }
    
    const nextBtn = document.getElementById('next-step-btn');
    if (nextBtn) {
        nextBtn.addEventListener('click', () => {
            if (currentStepIndex < protocolSteps.length - 1) {
                goToStep(currentStepIndex + 1);
            }
        });
    }
});

