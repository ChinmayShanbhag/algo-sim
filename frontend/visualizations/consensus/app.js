// Raft Consensus Algorithm Visualization

// Initialize theme immediately (before DOM loads)
(function() {
    const savedTheme = localStorage.getItem('theme') || 'dark';
    document.documentElement.setAttribute('data-theme', savedTheme);
})();

const API_BASE = 'http://localhost:8080/api/consensus/raft';

// Helper function to add session header to fetch options
function addSessionHeader(options = {}) {
    if (typeof window.SDS_SESSION !== 'undefined') {
        return window.SDS_SESSION.addSessionHeader(options);
    }
    return options;
}

let svg;
let clusterData = null;
let electionSteps = [];
let currentStepIndex = -1;
let initialNodesState = null;
let nodePositions = {}; // Store node positions for animations
let currentAnimation = null; // Track current animation
let electionMode = 'manual'; // 'manual' or 'automatic'
let automaticInterval = null; // Track automatic simulation interval
let isAutomaticRunning = false; // Flag to prevent multiple automatic simulations
let nodeTimeouts = {}; // Track timeout progress for each node (0-1)
let timeoutAnimationFrame = null; // Track animation frame for timeout rings
let ongoingElections = new Set(); // Track nodes currently in election
let shouldCancelElections = false; // Flag to cancel all ongoing elections
let animationModeWhenStarted = null; // Track which mode started the animation

// Color mapping for node states (theme-aware)
function getStateColors() {
    const isDark = document.documentElement.getAttribute('data-theme') === 'dark';
    return {
        follower: isDark ? '#475569' : '#94a3b8',   // gray - darker in dark mode for visibility
        candidate: '#fbbf24',  // yellow (works in both themes)
        leader: '#10b981'      // green (works in both themes)
    };
}

function initVisualization() {
    console.log('🔧 Initializing visualization...');
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
    console.log('✅ Loading message shown');
    
    // Load initial state
    console.log('🔧 Calling loadRaftState...');
    loadRaftState().catch((error) => {
        console.error('❌ Error in loadRaftState catch:', error);
        // Error already handled in loadRaftState
    });
    
    // Refresh every 2 seconds (only if no election in progress AND not in automatic mode)
    setInterval(() => {
        if (electionSteps.length === 0 && electionMode === 'manual') {
            loadRaftState().catch(() => {});
        }
    }, 2000);
}

function showLoading() {
    svg.selectAll('*').remove();
    const isDark = document.documentElement.getAttribute('data-theme') === 'dark';
    const textColor = isDark ? '#94a3b8' : '#64748b';
    svg.append('text')
        .attr('class', 'loading-text')
        .attr('x', 400)
        .attr('y', 300)
        .attr('text-anchor', 'middle')
        .attr('font-size', '18px')
        .attr('fill', textColor)
        .text('Loading Raft cluster state...');
}

function loadRaftState() {
    console.log('Loading Raft state from:', `${API_BASE}/state`);
    return fetch(`${API_BASE}/state`, addSessionHeader())
        .then(response => {
            console.log('Response status:', response.status, response.statusText);
            if (!response.ok) {
                throw new Error(`HTTP error! status: ${response.status}`);
            }
            return response.json();
        })
        .then(data => {
            console.log('Received data:', data);
            clusterData = data;
            if (!clusterData || !clusterData.nodes) {
                console.error('Invalid data structure:', clusterData);
                throw new Error('Invalid data structure received from backend');
            }
            console.log('Rendering cluster with', clusterData.nodes.length, 'nodes');
            renderCluster();
            hideError();
            return data;
        })
        .catch(error => {
            console.error('Error loading Raft state:', error);
            showError(`Cannot connect to backend: ${error.message}. Make sure the backend is running on port 8080.`);
            // Keep showing loading state but with error message
            showLoadingWithError('Backend connection failed. Check console for details.');
            throw error;
        });
}

function showLoadingWithError(message) {
    svg.selectAll('*').remove();
    const isDark = document.documentElement.getAttribute('data-theme') === 'dark';
    const errorColor = '#f87171';
    const textColor = isDark ? '#94a3b8' : '#64748b';
    svg.append('text')
        .attr('x', 400)
        .attr('y', 280)
        .attr('text-anchor', 'middle')
        .attr('font-size', '18px')
        .attr('fill', errorColor)
        .text('⚠️ Connection Error');
    svg.append('text')
        .attr('x', 400)
        .attr('y', 310)
        .attr('text-anchor', 'middle')
        .attr('font-size', '14px')
        .attr('fill', textColor)
        .text(message);
}

function showError(message) {
    let errorDiv = document.getElementById('error-message');
    if (!errorDiv) {
        errorDiv = document.createElement('div');
        errorDiv.id = 'error-message';
        errorDiv.className = 'error-message';
        document.querySelector('#container').insertBefore(errorDiv, document.querySelector('#visualization'));
    }
    errorDiv.textContent = message;
}

function hideError() {
    const errorDiv = document.getElementById('error-message');
    if (errorDiv) {
        errorDiv.remove();
    }
}

function getNodePosition(nodeId, nodes) {
    // First check if we have a stored position (most reliable)
    if (nodePositions[nodeId]) {
        return nodePositions[nodeId];
    }
    
    // Otherwise calculate it based on node's position in array
    const nodeRadius = 40;
    const centerX = 400;
    const centerY = 300;
    const radius = 200;
    const angleStep = (2 * Math.PI) / nodes.length;
    
    // Find the index of the node with this ID
    const nodeIndex = nodes.findIndex(n => n.id === nodeId);
    if (nodeIndex === -1) return null;
    
    const angle = nodeIndex * angleStep - Math.PI / 2;
    const pos = {
        x: centerX + radius * Math.cos(angle),
        y: centerY + radius * Math.sin(angle)
    };
    
    // Store it for future use
    nodePositions[nodeId] = pos;
    return pos;
}

function renderCluster() {
    if (!clusterData || !clusterData.nodes) {
        showLoading();
        return;
    }
    
    const nodes = clusterData.nodes;
    const nodeRadius = 40;
    const centerX = 400;
    const centerY = 300;
    const radius = 200;
    
    // Clear previous render (but DON'T remove message animations)
    svg.selectAll('g.node').remove();
    svg.selectAll('text.loading-text').remove();
    
    // Update node buttons
    updateNodeButtons(nodes);
    
    // Calculate positions in a circle
    // IMPORTANT: Sort nodes by ID to ensure consistent ordering
    const sortedNodes = [...nodes].sort((a, b) => a.id - b.id);
    const angleStep = (2 * Math.PI) / sortedNodes.length;
    
    const nodeGroups = svg.selectAll('g.node')
        .data(sortedNodes)
        .enter()
        .append('g')
        .attr('class', 'node')
        .attr('transform', (d, i) => {
            const angle = i * angleStep - Math.PI / 2;
            const x = centerX + radius * Math.cos(angle);
            const y = centerY + radius * Math.sin(angle);
            // Store position for animations - use node ID as key
            nodePositions[d.id] = { x, y };
            return `translate(${x}, ${y})`;
        })
        .style('cursor', 'pointer')
        .on('click', (event, d) => {
            startElection(d.id);
        });
    
    // Ensure all node positions are stored correctly (already stored above, but double-check)
    // sortedNodes is already defined above, so we reuse it
    sortedNodes.forEach((node, index) => {
        const angle = index * angleStep - Math.PI / 2;
        const x = centerX + radius * Math.cos(angle);
        const y = centerY + radius * Math.sin(angle);
        nodePositions[node.id] = { x, y };
    });
    
    // Draw node circles
    const colors = getStateColors();
    const isDark = document.documentElement.getAttribute('data-theme') === 'dark';
    const strokeColor = isDark ? '#475569' : '#1e293b';
    
    nodeGroups.append('circle')
        .attr('r', nodeRadius)
        .attr('fill', d => {
            // Ensure candidate color is applied correctly
            if (d.state === 'candidate') return colors.candidate;
            if (d.state === 'leader') return colors.leader;
            return colors.follower;
        })
        .attr('stroke', strokeColor)
        .attr('stroke-width', 2.5);
    
    // Draw timeout ring for followers in automatic mode (AFTER main circle)
    if (electionMode === 'automatic') {
        nodeGroups.each(function(d) {
            if (d.state === 'follower') {
                const group = d3.select(this);
                const timeoutProgress = nodeTimeouts[d.id] || 0;
                
                // Only show ring if there's progress
                if (timeoutProgress > 0) {
                    // Create arc for timeout visualization
                    const arc = d3.arc()
                        .innerRadius(nodeRadius + 5)
                        .outerRadius(nodeRadius + 10)
                        .startAngle(0)
                        .endAngle(timeoutProgress * 2 * Math.PI);
                    
                    group.append('path')
                        .attr('class', 'timeout-ring')
                        .attr('d', arc)
                        .attr('fill', timeoutProgress > 0.7 ? '#ef4444' : '#fbbf24')
                        .attr('opacity', 0.8);
                }
            }
        });
    }
    
    // Add node ID
    const textColor = isDark ? '#f1f5f9' : '#0f172a';
    const mutedColor = isDark ? '#94a3b8' : '#64748b';
    
    nodeGroups.append('text')
        .attr('text-anchor', 'middle')
        .attr('dy', -5)
        .attr('font-size', '16px')
        .attr('font-weight', 'bold')
        .attr('fill', textColor)
        .text(d => `Node ${d.id}`);
    
    // Add state label
    nodeGroups.append('text')
        .attr('text-anchor', 'middle')
        .attr('dy', 15)
        .attr('font-size', '12px')
        .attr('fill', textColor)
        .text(d => d.state);
    
    // Add term
    nodeGroups.append('text')
        .attr('text-anchor', 'middle')
        .attr('dy', 30)
        .attr('font-size', '10px')
        .attr('fill', mutedColor)
        .text(d => `Term: ${d.currentTerm}`);
    
    // Don't auto-animate here - let goToStep handle it
}

// REWRITTEN: Clean voting animation function
function animateMessage(fromNodeId, toNodeId, messageType, onComplete) {
    // Validate inputs
    if (!clusterData || !clusterData.nodes) {
        if (onComplete) onComplete();
        return;
    }
    
    // Convert to numbers
    const fromId = Number(fromNodeId);
    const toId = Number(toNodeId);
    
    // Get positions from stored positions (set during renderCluster)
    const fromPos = nodePositions[fromId];
    const toPos = nodePositions[toId];
    
    if (!fromPos || !toPos) {
        console.warn(`Positions not found for nodes ${fromId} -> ${toId}`);
        if (onComplete) onComplete();
        return;
    }
    
    // Message color
    const isDark = document.documentElement.getAttribute('data-theme') === 'dark';
    const messageColor = messageType === 'vote_request' ? '#fbbf24' : 
                        messageType === 'vote_response' ? '#10b981' : 
                        (isDark ? '#60a5fa' : '#3b82f6');
    
    console.log(`📨 Animating message: ${fromId} -> ${toId}, type: ${messageType}, mode: ${electionMode}`);
    console.log(`  From position:`, fromPos);
    console.log(`  To position:`, toPos);
    
    // Store which mode started this animation
    const startedInMode = electionMode;
    
    // Check if we should cancel
    if (shouldCancelElections) {
        console.log(`  ❌ Animation canceled (shouldCancel: ${shouldCancelElections})`);
        if (onComplete) onComplete();
        return;
    }
    
    // Create message circle at source node
    const message = svg.append('circle')
        .attr('class', 'message')
        .attr('data-mode', startedInMode) // Tag with mode
        .attr('r', 10)
        .attr('fill', messageColor)
        .attr('stroke', isDark ? '#1e293b' : '#ffffff')
        .attr('stroke-width', 2)
        .attr('cx', fromPos.x)
        .attr('cy', fromPos.y)
        .attr('opacity', 0.9)
        .style('filter', 'drop-shadow(0 0 6px ' + messageColor + ')');
    
    console.log(`  Message created at (${fromPos.x}, ${fromPos.y})`);
    
    // Animate to destination - exactly 1 second
    message.transition()
        .duration(1000)
        .ease(d3.easeLinear)
        .attr('cx', toPos.x)
        .attr('cy', toPos.y)
        .on('interrupt', () => {
            console.log(`  ⚠️ Animation interrupted`);
            message.remove();
            if (onComplete) onComplete();
        })
        .on('end', () => {
            // Check if mode changed since animation started
            if (startedInMode !== electionMode) {
                console.log(`  ❌ Animation canceled - mode changed from ${startedInMode} to ${electionMode}`);
                message.remove();
                if (onComplete) onComplete();
                return;
            }
            
            // Check if we should cancel
            if (shouldCancelElections) {
                console.log(`  ❌ Animation end canceled`);
                message.remove();
                if (onComplete) onComplete();
                return;
            }
            
            console.log(`  ✅ Message reached destination (${toPos.x}, ${toPos.y})`);
            
            // Arrival effect
            const arrival = svg.append('circle')
                .attr('class', 'arrival-effect') // Add class for easy cleanup
                .attr('cx', toPos.x)
                .attr('cy', toPos.y)
                .attr('r', 15)
                .attr('fill', 'none')
                .attr('stroke', messageColor)
                .attr('stroke-width', 3)
                .attr('opacity', 0.8);
            
            arrival.transition()
                .duration(500)
                .attr('r', 35)
                .attr('opacity', 0)
                .on('end', () => arrival.remove())
                .on('interrupt', () => arrival.remove());
            
            message.remove();
            
            // Call completion callback
            if (onComplete) onComplete();
        });
}

function updateNodeButtons(nodes) {
    const buttonsContainer = document.getElementById('node-buttons');
    if (!buttonsContainer) return;
    
    buttonsContainer.innerHTML = '';
    nodes.forEach(node => {
        const btn = document.createElement('button');
        btn.className = 'node-button';
        btn.textContent = `Node ${node.id}`;
        btn.onclick = () => startElection(node.id);
        buttonsContainer.appendChild(btn);
    });
}

function startElection(nodeId) {
    // Save initial state before election
    loadRaftState().then(() => {
        if (!clusterData || !clusterData.nodes) return;
        
        // Deep copy initial state
        initialNodesState = JSON.parse(JSON.stringify(clusterData.nodes));
        
        showFeedback(`Node ${nodeId} is starting an election...`);
        
        fetch(`${API_BASE}/election?nodeId=${nodeId}`, addSessionHeader({
            method: 'POST'
        }))
            .then(response => {
                if (!response.ok) {
                    throw new Error(`HTTP error! status: ${response.status}`);
                }
                return response.json();
            })
            .then(data => {
                // Store final state and steps
                electionSteps = data.electionSteps || [];
                currentStepIndex = -1;
                
                // Reset to initial state and show step 0
                clusterData = { nodes: JSON.parse(JSON.stringify(initialNodesState)) };
                goToStep(0);
                
                // Enable navigation
                updateNavigationButtons();
            })
            .catch(error => {
                console.error('Error starting election:', error);
                showError('Failed to start election. Make sure backend is running.');
            });
    });
}

function goToStep(stepIndex) {
    if (!electionSteps || electionSteps.length === 0) return;
    
    if (stepIndex < 0) stepIndex = 0;
    if (stepIndex > electionSteps.length) stepIndex = electionSteps.length;
    
    currentStepIndex = stepIndex;
    
    // Reset to initial state
    clusterData = { nodes: JSON.parse(JSON.stringify(initialNodesState)) };
    
    // Apply steps up to current step
    const candidateId = electionSteps[0]?.votedNodes[0];
    
    for (let i = 0; i < stepIndex; i++) {
        applyStep(electionSteps[i], candidateId);
    }
    
    // Show current step info
    if (stepIndex > 0 && stepIndex <= electionSteps.length) {
        const currentStep = electionSteps[stepIndex - 1];
        showCurrentStep(currentStep);
        showElectionSteps(electionSteps, stepIndex - 1);
    } else {
        showCurrentStep(null);
        showElectionSteps([], -1);
    }
    
    renderCluster();
    
    // Trigger animation for current step after nodes are rendered
    setTimeout(() => {
        if (stepIndex > 0 && stepIndex <= electionSteps.length) {
            const step = electionSteps[stepIndex - 1];
            // Only animate if it's a message step
            if (step.fromNode !== null && step.fromNode !== undefined && 
                step.toNode !== null && step.toNode !== undefined &&
                step.messageType) {
                // Ensure we have cluster data and nodes are rendered
                if (clusterData && clusterData.nodes) {
                    // Delay animation slightly to ensure nodes are positioned
                    setTimeout(() => {
                        // Ensure node IDs are numbers
                        const fromId = typeof step.fromNode === 'number' ? step.fromNode : parseInt(step.fromNode);
                        const toId = typeof step.toNode === 'number' ? step.toNode : parseInt(step.toNode);
                        animateMessage(fromId, toId, step.messageType);
                    }, 200);
                }
            }
        }
    }, 500);
    
    updateNavigationButtons();
}

function applyStep(step, candidateId) {
    if (!clusterData || !clusterData.nodes) return;
    
    const nodes = clusterData.nodes;
    const candidate = nodes.find(n => n.id === candidateId);
    
    if (!candidate) return;
    
    switch(step.action) {
        case 'increment_term_and_vote_self':
            candidate.state = 'candidate';
            candidate.currentTerm++;
            candidate.votedFor = candidateId;
            break;
            
        case 'vote_request_sent':
            // Message is being sent, no state change yet
            break;
            
        case 'vote_received':
            const votedNodeId = step.votedNodes[step.votedNodes.length - 1];
            const votedNode = nodes.find(n => n.id === votedNodeId);
            if (votedNode) {
                votedNode.votedFor = candidateId;
                votedNode.currentTerm = candidate.currentTerm;
                votedNode.state = 'follower';
            }
            break;
            
        case 'vote_rejected':
            // Node already voted, no change needed
            break;
            
        case 'election_success':
            candidate.state = 'leader';
            nodes.forEach(node => {
                if (node.id !== candidateId) {
                    node.state = 'follower';
                    node.currentTerm = candidate.currentTerm;
                }
            });
            break;
            
        case 'election_failed':
            candidate.state = 'follower';
            candidate.votedFor = null;
            break;
    }
}

function updateNavigationButtons() {
    const prevBtn = document.getElementById('prev-step-btn');
    const nextBtn = document.getElementById('next-step-btn');
    const stepInfo = document.getElementById('step-info');
    
    if (!prevBtn || !nextBtn) return;
    
    const hasSteps = electionSteps && electionSteps.length > 0;
    const canGoPrev = hasSteps && currentStepIndex > 0;
    const canGoNext = hasSteps && currentStepIndex < electionSteps.length;
    
    prevBtn.disabled = !canGoPrev;
    nextBtn.disabled = !canGoNext;
    
    if (stepInfo && hasSteps) {
        stepInfo.textContent = `Step ${currentStepIndex}/${electionSteps.length}`;
    } else if (stepInfo) {
        stepInfo.textContent = 'No election';
    }
}

function showCurrentStep(step) {
    const stepExplanation = document.getElementById('current-step-explanation');
    if (!stepExplanation) return;
    
    if (!step) {
        stepExplanation.innerHTML = '<p class="feedback-text">Click a node to start an election and see step-by-step process.</p>';
        return;
    }
    
    let html = `<div class="current-step">`;
    html += `<div class="step-header-large">`;
    html += `<span class="step-number-large">Step ${step.stepNumber}</span>`;
    html += `<span class="step-votes-large">Votes: ${step.votes}/5</span>`;
    html += `</div>`;
    html += `<div class="step-description-large">${step.description}</div>`;
    
    if (step.votedNodes && step.votedNodes.length > 0) {
        html += `<div class="step-voters-large">Voters: ${step.votedNodes.map(n => `Node ${n}`).join(', ')}</div>`;
    }
    
    html += `</div>`;
    stepExplanation.innerHTML = html;
}

function showElectionSteps(steps, currentIndex = -1) {
    const stepsContainer = document.getElementById('election-steps');
    if (!stepsContainer) return;
    
    if (steps.length === 0) {
        stepsContainer.innerHTML = '<p class="feedback-text">Click a node to see the step-by-step election process!</p>';
        return;
    }
    
    let html = '<div class="steps-list">';
    steps.forEach((step, index) => {
        const isCurrent = index === currentIndex;
        const stepClass = step.action === 'election_success' ? 'step-success' : 
                         step.action === 'election_failed' ? 'step-failed' :
                         step.action === 'vote_received' ? 'step-vote-yes' :
                         step.action === 'vote_rejected' ? 'step-vote-no' : 'step-info';
        
        const currentClass = isCurrent ? 'current-step-highlight' : '';
        
        html += `<div class="election-step ${stepClass} ${currentClass}">`;
        html += `<div class="step-header">`;
        html += `<span class="step-number">Step ${step.stepNumber}</span>`;
        html += `<span class="step-votes">Votes: ${step.votes}/5</span>`;
        html += `</div>`;
        html += `<div class="step-description">${step.description}</div>`;
        
        if (step.votedNodes && step.votedNodes.length > 0) {
            html += `<div class="step-voters">Voters: ${step.votedNodes.map(n => `Node ${n}`).join(', ')}</div>`;
        }
        
        html += `</div>`;
    });
    html += '</div>';
    
    stepsContainer.innerHTML = html;
}

function resetCluster() {
    showFeedback('Resetting cluster... All nodes back to Followers, Term 0.');
    
    // Stop any animations
    if (currentAnimation) {
        currentAnimation.interrupt();
        currentAnimation = null;
    }
    svg.selectAll('circle.message').remove();
    
    // Reset election state
    electionSteps = [];
    currentStepIndex = -1;
    initialNodesState = null;
    nodePositions = {};
    updateNavigationButtons();
    showCurrentStep(null);
    showElectionSteps([], -1);
    
    fetch(`${API_BASE}/reset`, addSessionHeader({
        method: 'POST'
    }))
        .then(response => {
            if (!response.ok) {
                throw new Error(`HTTP error! status: ${response.status}`);
            }
            return response.json();
        })
        .then(data => {
            clusterData = data;
            
            // If in manual mode, set a random leader
            if (electionMode === 'manual') {
                const leaderId = Math.floor(Math.random() * data.nodes.length);
                return fetch(`${API_BASE}/set-leader?nodeId=${leaderId}`, addSessionHeader({ method: 'POST' }));
            }
            return Promise.resolve({ json: () => Promise.resolve(data) });
        })
        .then(response => response.json())
        .then(data => {
            clusterData = data;
            renderCluster();
            
            // If in automatic mode, restart simulation
            if (electionMode === 'automatic') {
                startAutomaticSimulation();
            } else {
                updateModeUI();
                const leader = data.nodes.find(n => n.state === 'leader');
                showFeedback(`✅ Cluster reset! Node ${leader?.id || '?'} is the initial leader. Click a node to start an election.`);
            }
        })
        .catch(error => {
            console.error('Error resetting cluster:', error);
            showError('Failed to reset cluster.');
        });
}

function showFeedback(message) {
    const feedbackDiv = document.getElementById('action-feedback');
    if (feedbackDiv) {
        feedbackDiv.innerHTML = `<p class="feedback-text">${message}</p>`;
    }
}

// Theme management
function initTheme() {
    // Default to dark mode
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
    if (clusterData) {
        renderCluster();
    }
}

function updateThemeIcon(theme) {
    const themeIcon = document.querySelector('.theme-icon');
    if (themeIcon) {
        themeIcon.textContent = theme === 'dark' ? '☀️' : '🌙';
    }
}

function updateModeUI() {
    const nodeButtons = document.getElementById('node-buttons');
    const stepNav = document.querySelector('.step-navigation-compact');
    
    if (!nodeButtons || !clusterData || !clusterData.nodes) return;
    
    const nodes = clusterData.nodes;
    
    if (electionMode === 'automatic') {
        // Hide node buttons and step controls in automatic mode
        nodeButtons.style.display = 'none';
        if (stepNav) stepNav.style.display = 'none';
        
        // Start automatic simulation
        startAutomaticSimulation();
    } else {
        // Show node buttons and step controls in manual mode
        nodeButtons.style.display = 'flex';
        if (stepNav) stepNav.style.display = 'flex';
        
        // Stop automatic simulation
        stopAutomaticSimulation();
        
        // FORCE remove all animations immediately
        console.log('🧹 Forcing removal of all animations on mode switch');
        svg.selectAll('circle.message').interrupt().remove();
        svg.selectAll('circle.arrival-effect').interrupt().remove();
        
        // Reset cluster to get a clean state for manual mode
        fetch(`${API_BASE}/reset`, { method: 'POST' })
            .then(response => response.json())
            .then(data => {
                clusterData = data;
                
                // Set a random node as initial leader
                const leaderId = Math.floor(Math.random() * data.nodes.length);
                return fetch(`${API_BASE}/set-leader?nodeId=${leaderId}`, addSessionHeader({ method: 'POST' }));
            })
            .then(response => response.json())
            .then(data => {
                clusterData = data;
                renderCluster();
                
                // Update button text
                const buttons = nodeButtons.querySelectorAll('.node-button');
                buttons.forEach((btn, idx) => {
                    if (idx < nodes.length) {
                        const nodeId = nodes[idx].id;
                        btn.textContent = `Node ${nodeId}`;
                        btn.onclick = () => startElection(nodeId);
                    }
                });
                
                const leader = data.nodes.find(n => n.state === 'leader');
                showFeedback(`Manual mode: Node ${leader?.id || '?'} is leader. Click a node to start an election.`);
            })
            .catch(error => {
                console.error('Error resetting cluster:', error);
                showError('Failed to reset cluster for manual mode.');
            });
    }
}

function startAutomaticSimulation() {
    // Prevent multiple instances
    if (isAutomaticRunning) {
        console.log('⚠️ Automatic simulation already running, skipping...');
        return;
    }
    
    // Stop any existing simulation
    stopAutomaticSimulation();
    
    // Mark as running
    isAutomaticRunning = true;
    console.log('▶️ Starting automatic simulation...');
    
    // Reset cluster and set initial leader
    fetch(`${API_BASE}/reset`, { method: 'POST' })
        .then(response => response.json())
        .then(data => {
            clusterData = data;
            
            // Set a random node as initial leader
            const leaderId = Math.floor(Math.random() * data.nodes.length);
            return fetch(`${API_BASE}/set-leader?nodeId=${leaderId}`, { method: 'POST' });
        })
        .then(response => response.json())
        .then(data => {
            clusterData = data;
            renderCluster();
            showFeedback(`Automatic mode: Node ${data.nodes.find(n => n.state === 'leader')?.id || '?'} is leader. Simulation starting...`);
            
            // Start timeout animations - this handles the automatic election cycle
            startTimeoutAnimations();
        })
        .catch(error => {
            console.error('Error starting automatic simulation:', error);
            showError('Failed to start automatic simulation.');
        });
}

function stopAutomaticSimulation() {
    console.log('⏹️ Stopping automatic simulation...');
    
    // Stop the interval
    if (automaticInterval) {
        clearInterval(automaticInterval);
        automaticInterval = null;
    }
    
    // Stop the animation frame
    if (timeoutAnimationFrame) {
        cancelAnimationFrame(timeoutAnimationFrame);
        timeoutAnimationFrame = null;
    }
    
    // Cancel all ongoing elections
    shouldCancelElections = true;
    ongoingElections.clear();
    
    // INTERRUPT all D3 transitions (this stops animations mid-flight)
    svg.selectAll('circle.message').interrupt();
    svg.selectAll('circle.arrival-effect').interrupt();
    
    // Remove all message and arrival animations
    svg.selectAll('circle.message').remove();
    svg.selectAll('circle.arrival-effect').remove();
    
    console.log('🧹 Cleared all automatic mode animations');
    
    // Clear all timeouts
    nodeTimeouts = {};
    
    // Mark as stopped
    isAutomaticRunning = false;
    
    // Reset cancel flag after cleanup
    setTimeout(() => {
        shouldCancelElections = false;
    }, 100);
}

// Animate timeout rings for followers
function startTimeoutAnimations() {
    if (!clusterData || !clusterData.nodes) return;
    
    console.log('🔄 Starting timeout animations...');
    const TIMEOUT_DURATION_MIN = 3000; // 3 seconds minimum
    const TIMEOUT_DURATION_MAX = 6000; // 6 seconds maximum
    
    function animate() {
        if (electionMode !== 'automatic' || !isAutomaticRunning) {
            console.log('⏹️ Stopping timeout animations');
            return;
        }
        
        if (!clusterData || !clusterData.nodes) {
            timeoutAnimationFrame = requestAnimationFrame(animate);
            return;
        }
        
        const nodes = clusterData.nodes;
        const leader = nodes.find(n => n.state === 'leader');
        
        // Update timeout progress for each follower
        nodes.forEach(node => {
            if (node.state === 'follower') {
                // Initialize timeout for this node if not exists
                if (!nodeTimeouts[node.id + '_duration']) {
                    nodeTimeouts[node.id + '_duration'] = TIMEOUT_DURATION_MIN + 
                        Math.random() * (TIMEOUT_DURATION_MAX - TIMEOUT_DURATION_MIN);
                    nodeTimeouts[node.id + '_start'] = Date.now();
                    console.log(`⏱️ Node ${node.id} timeout set to ${(nodeTimeouts[node.id + '_duration'] / 1000).toFixed(1)}s`);
                }
                
                const duration = nodeTimeouts[node.id + '_duration'];
                const nodeElapsed = Date.now() - nodeTimeouts[node.id + '_start'];
                nodeTimeouts[node.id] = Math.min(nodeElapsed / duration, 1);
                
                // If timeout reached, trigger election (don't stop other timers)
                if (nodeTimeouts[node.id] >= 1 && !ongoingElections.has(node.id)) {
                    console.log(`⚡ Node ${node.id} timeout reached! Starting election...`);
                    
                    // Reset this node's timeout
                    delete nodeTimeouts[node.id];
                    delete nodeTimeouts[node.id + '_duration'];
                    delete nodeTimeouts[node.id + '_start'];
                    
                    // Trigger election for this node
                    const message = leader 
                        ? `Node ${node.id} timeout: No heartbeat from leader. Starting election...`
                        : `Node ${node.id} timeout: Starting election...`;
                    
                    showFeedback(message);
                    
                    // Start election (non-blocking - other timers continue)
                    startElectionAutomatic(node.id, () => {
                        console.log(`✅ Node ${node.id} election complete`);
                        // Check if there's a leader now
                        const newLeader = clusterData.nodes.find(n => n.state === 'leader');
                        if (newLeader) {
                            console.log(`👑 Node ${newLeader.id} is now leader, canceling all other elections`);
                            // Cancel all ongoing elections
                            shouldCancelElections = true;
                            // Clear all message circles from canceled elections
                            svg.selectAll('circle.message').remove();
                            // Reset all follower timeouts when a leader is elected
                            Object.keys(nodeTimeouts).forEach(key => {
                                delete nodeTimeouts[key];
                                delete nodeTimeouts[key + '_duration'];
                                delete nodeTimeouts[key + '_start'];
                            });
                            
                            // Reset cancel flag after a short delay (allow canceled elections to finish)
                            setTimeout(() => {
                                console.log('🔄 Resetting shouldCancelElections flag for next cycle');
                                shouldCancelElections = false;
                            }, 2000);
                        }
                    });
                }
            } else if (node.state === 'candidate') {
                // Candidates don't have timeout rings, but keep their timeout data
                // They might lose the election and become followers again
            } else {
                // Leader - clear timeout
                if (nodeTimeouts[node.id] !== undefined) {
                    console.log(`🧹 Clearing timeout for Node ${node.id} (${node.state})`);
                    delete nodeTimeouts[node.id];
                    delete nodeTimeouts[node.id + '_duration'];
                    delete nodeTimeouts[node.id + '_start'];
                }
            }
        });
        
        // Re-render to update timeout rings (but not too frequently)
        // Only render every ~100ms to avoid interfering with election animations
        const now = Date.now();
        if (!animate.lastRender || now - animate.lastRender > 100) {
            renderCluster();
            animate.lastRender = now;
        }
        
        // Continue animation loop (never stop)
        timeoutAnimationFrame = requestAnimationFrame(animate);
    }
    
    animate();
}

// This function is no longer used - timeout animations handle elections
function runAutomaticStep() {
    // Deprecated - startTimeoutAnimations() handles automatic elections now
    console.log('runAutomaticStep called but not needed');
}

// REWRITTEN: Clean automatic election with proper animation waiting
function startElectionAutomatic(nodeId, callback) {
    // Mark this election as ongoing
    ongoingElections.add(nodeId);
    console.log(`🗳️ Starting election for Node ${nodeId}. Ongoing elections:`, Array.from(ongoingElections));
    
    // Temporarily mark this node as candidate in the UI
    const candidateNode = clusterData.nodes.find(n => n.id === nodeId);
    if (candidateNode) {
        candidateNode.state = 'candidate';
        renderCluster();
    }
    
    fetch(`${API_BASE}/election?nodeId=${nodeId}`, {
        method: 'POST'
    })
        .then(response => response.json())
        .then(data => {
            // Check if we should cancel this election
            if (shouldCancelElections) {
                console.log(`❌ Canceling election for Node ${nodeId} - leader already elected`);
                ongoingElections.delete(nodeId);
                if (callback) callback();
                return;
            }
            
            // DON'T update cluster data yet - wait for animations to complete
            const finalState = data.nodes;
            const electionSteps = data.electionSteps || [];
            
            // Animate the election steps
            if (electionSteps.length > 0) {
                animateElectionSteps(electionSteps, 0, () => {
                    // Check again if we should cancel
                    if (shouldCancelElections) {
                        console.log(`❌ Canceling election for Node ${nodeId} after animation - leader already elected`);
                        ongoingElections.delete(nodeId);
                        if (callback) callback();
                        return;
                    }
                    
                    // NOW update cluster data with final state
                    clusterData = { nodes: finalState };
                    renderCluster();
                    
                    const leader = finalState.find(n => n.state === 'leader');
                    if (leader) {
                        showFeedback(`✅ Node ${leader.id} elected as leader (Term ${leader.currentTerm})`);
                        // Don't set shouldCancelElections here - it's already set in the callback
                    } else {
                        showFeedback(`❌ Election failed for Node ${nodeId}`);
                    }
                    
                    // Remove from ongoing elections
                    ongoingElections.delete(nodeId);
                    
                    // Call callback when done
                    if (callback) callback();
                });
            } else {
                // No steps, update immediately
                clusterData = { nodes: finalState };
                renderCluster();
                ongoingElections.delete(nodeId);
                if (callback) callback();
            }
        })
        .catch(error => {
            console.error('Error in automatic election:', error);
            ongoingElections.delete(nodeId);
            if (callback) callback();
        });
}

function animateElectionSteps(steps, startIndex, callback) {
    if (startIndex >= steps.length) {
        if (callback) callback();
        return;
    }
    
    const step = steps[startIndex];
    if (step.fromNode !== null && step.fromNode !== undefined && 
        step.toNode !== null && step.toNode !== undefined &&
        step.messageType) {
        // Ensure node IDs are numbers
        const fromId = typeof step.fromNode === 'number' ? step.fromNode : parseInt(step.fromNode);
        const toId = typeof step.toNode === 'number' ? step.toNode : parseInt(step.toNode);
        
        // Wait a bit before animating to ensure cluster is rendered
        setTimeout(() => {
            // animateMessage returns a promise-like that resolves when animation completes
            animateMessage(fromId, toId, step.messageType, () => {
                // Animation completed - wait a bit then move to next step
                setTimeout(() => {
                    animateElectionSteps(steps, startIndex + 1, callback);
                }, 200);
            });
        }, 100);
    } else {
        // No animation for this step, move to next immediately
        setTimeout(() => {
            animateElectionSteps(steps, startIndex + 1, callback);
        }, 200);
    }
}

function simulateTimeout() {
    if (electionMode !== 'automatic') return;
    
    // In automatic mode, user clicks to simulate timeout
    // This function can be extended for fully automatic simulation later
    showFeedback('Automatic mode: Click a node to simulate timeout scenario.');
}

document.addEventListener('DOMContentLoaded', () => {
    console.log('✅ DOM Content Loaded');
    console.log('✅ D3 available:', typeof d3 !== 'undefined');
    console.log('✅ Visualization element exists:', document.getElementById('visualization') !== null);
    
    initTheme();
    console.log('✅ Theme initialized');
    
    initVisualization();
    console.log('✅ Visualization initialized');
    
    // Setup theme toggle
    const themeToggle = document.getElementById('theme-toggle');
    if (themeToggle) {
        themeToggle.addEventListener('click', toggleTheme);
    }
    
    // Setup mode toggle
    const modeInputs = document.querySelectorAll('input[name="election-mode"]');
    modeInputs.forEach(input => {
        input.addEventListener('change', (e) => {
            const oldMode = electionMode;
            const newMode = e.target.value;
            
            console.log(`🔄 Mode changed from ${oldMode} to ${newMode}`);
            
            // If switching from automatic to manual, just reload the page for a clean state
            if (oldMode === 'automatic' && newMode === 'manual') {
                console.log('🔄 Reloading page for clean state...');
                window.location.reload();
                return;
            }
            
            // Otherwise, update mode normally
            electionMode = newMode;
            
            // Update UI immediately
            if (clusterData && clusterData.nodes) {
                updateModeUI();
            } else {
                // Load cluster data first
                loadRaftState().then(() => {
                    updateModeUI();
                }).catch(() => {});
            }
        });
    });
    
    // Setup reset button
    const resetBtn = document.getElementById('reset-btn');
    if (resetBtn) {
        resetBtn.addEventListener('click', () => {
            electionSteps = [];
            currentStepIndex = -1;
            initialNodesState = null;
            resetCluster();
        });
    }
    
    // Setup navigation buttons
    const prevBtn = document.getElementById('prev-step-btn');
    const nextBtn = document.getElementById('next-step-btn');
    
    if (prevBtn) {
        prevBtn.addEventListener('click', () => {
            if (currentStepIndex > 0) {
                goToStep(currentStepIndex - 1);
            }
        });
    }
    
    if (nextBtn) {
        nextBtn.addEventListener('click', () => {
            if (currentStepIndex < electionSteps.length) {
                goToStep(currentStepIndex + 1);
            }
        });
    }
    
    // Initialize mode UI after cluster data loads
    loadRaftState().then(() => {
        updateModeUI();
    }).catch(() => {});
});

