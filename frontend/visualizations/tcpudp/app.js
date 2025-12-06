// Theme management
(function() {
    const savedTheme = localStorage.getItem('theme') || 'dark';
    document.documentElement.setAttribute('data-theme', savedTheme);
})();

const API_BASE = 'http://localhost:8080/api/tcpudp';

function addSessionHeader(options = {}) {
    if (typeof window.SDS_SESSION !== 'undefined') {
        return window.SDS_SESSION.addSessionHeader(options);
    }
    return options;
}

let state = null;
let isProcessing = false;

function initVisualization() {
    console.log('🔧 Initializing TCP/UDP visualization...');
    
    // Setup event listeners
    document.getElementById('send-btn').addEventListener('click', sendBothPackets);
    document.getElementById('connect-btn').addEventListener('click', establishConnection);
    document.getElementById('disconnect-btn').addEventListener('click', closeConnection);
    document.getElementById('reset-btn').addEventListener('click', resetSimulation);
    
    // Allow Enter key to send
    document.getElementById('packet-data').addEventListener('keypress', (e) => {
        if (e.key === 'Enter') {
            sendBothPackets();
        }
    });
    
    // Setup loss rate slider
    const slider = document.getElementById('loss-rate-slider');
    slider.addEventListener('input', (e) => {
        const rate = e.target.value;
        document.getElementById('loss-rate-display').textContent = `${rate}%`;
    });
    
    slider.addEventListener('change', (e) => {
        const rate = e.target.value;
        setPacketLossRate(rate / 100);
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
        });
        const currentTheme = document.documentElement.getAttribute('data-theme');
        const icon = themeToggle.querySelector('.theme-icon');
        if (icon) icon.textContent = currentTheme === 'dark' ? '🌙' : '☀️';
    }
    
    loadState();
    updateConnectionButtons();
}

function loadState() {
    fetch(`${API_BASE}/state`, addSessionHeader())
        .then(response => response.json())
        .then(data => {
            state = data;
            renderState();
            updateConnectionButtons();
        })
        .catch(error => {
            console.error('Error loading state:', error);
            showLaneMessage('tcp', '❌ Backend not running');
            showLaneMessage('udp', '❌ Backend not running');
        });
}

function updateConnectionButtons() {
    const connectBtn = document.getElementById('connect-btn');
    const disconnectBtn = document.getElementById('disconnect-btn');
    
    if (state && state.tcpConnection) {
        if (state.tcpConnection.isConnected) {
            connectBtn.disabled = true;
            connectBtn.textContent = '✅ TCP Connected';
            disconnectBtn.disabled = false;
        } else {
            connectBtn.disabled = false;
            connectBtn.textContent = '🔌 Establish TCP Connection';
            disconnectBtn.disabled = true;
        }
    }
}

function renderState() {
    if (!state) return;
    
    renderTCPPackets();
    renderUDPPackets();
    renderTCPStats();
    renderUDPStats();
}

function sendBothPackets() {
    if (isProcessing) {
        alert('Please wait for current packets to finish processing');
        return;
    }
    
    // Check if TCP connection is established
    if (!state || !state.tcpConnection || !state.tcpConnection.isConnected) {
        // Auto-establish connection first
        showLaneMessage('tcp', '⚠️ TCP not connected - establishing connection first...');
        showLaneMessage('udp', 'ℹ️ UDP is connectionless - ready to send');
        
        establishConnection();
        
        // Wait for connection to complete, then send packets
        setTimeout(() => {
            sendBothPacketsInternal();
        }, 4000); // Wait for handshake animation (3.5s + buffer)
        return;
    }
    
    sendBothPacketsInternal();
}

function sendBothPacketsInternal() {
    const input = document.getElementById('packet-data');
    const data = input.value.trim() || 'Data';
    const sendBtn = document.getElementById('send-btn');
    
    // Disable button during processing
    sendBtn.disabled = true;
    sendBtn.textContent = '⏳ Sending...';
    
    isProcessing = true;
    
    // Send TCP packet
    const tcpPromise = fetch(`${API_BASE}/send-tcp?data=${encodeURIComponent(data)}`, 
        addSessionHeader({ method: 'POST' }))
        .then(response => response.json());
    
    // Send UDP packet
    const udpPromise = fetch(`${API_BASE}/send-udp?data=${encodeURIComponent(data)}`, 
        addSessionHeader({ method: 'POST' }))
        .then(response => response.json());
    
    // Wait for both to complete
    Promise.all([tcpPromise, udpPromise])
        .then(([tcpData, udpData]) => {
            state = udpData; // Use the latest state
            input.value = '';
            
            // Show visual feedback
            showLaneMessage('tcp', `📦 TCP packet #${state.tcpPackets.length} queued...`);
            showLaneMessage('udp', `📦 UDP packet #${state.udpPackets.length} queued...`);
            
            // Auto-process after a short delay
            setTimeout(() => {
                processPackets();
                sendBtn.disabled = false;
                sendBtn.textContent = '📤 Send to Both Protocols';
            }, 500);
        })
        .catch(error => {
            console.error('Error sending packets:', error);
            showLaneMessage('tcp', `❌ Error sending packet`);
            showLaneMessage('udp', `❌ Error sending packet`);
            isProcessing = false;
            sendBtn.disabled = false;
            sendBtn.textContent = '📤 Send to Both Protocols';
        });
}

function processPackets() {
    showLaneMessage('tcp', '⏳ Processing TCP packets...');
    showLaneMessage('udp', '⏳ Processing UDP packets...');
    
    fetch(`${API_BASE}/process`, addSessionHeader({ method: 'POST' }))
        .then(response => response.json())
        .then(data => {
            state = data;
            
            // Calculate animation duration based on packet status
            let tcpAnimationTime = 0;
            let udpAnimationTime = 0;
            
            // TCP animation time calculation
            if (state.tcpPackets && state.tcpPackets.length > 0) {
                const lastTCP = state.tcpPackets[state.tcpPackets.length - 1];
                if (lastTCP.status === 'lost') {
                    // Max retries: 4 attempts * 1.2s + 1s final = ~5.8s
                    tcpAnimationTime = (lastTCP.retryCount + 1) * 1200 + 1000;
                } else if (lastTCP.retryCount > 0) {
                    // Retries + success: retries * 1.2s + 2s success + 1.5s ACK
                    tcpAnimationTime = lastTCP.retryCount * 1200 + 2000 + 1500;
                } else {
                    // Success on first attempt: 2s + 1.5s ACK
                    tcpAnimationTime = 2000 + 1500;
                }
            }
            
            // UDP animation time calculation
            if (state.udpPackets && state.udpPackets.length > 0) {
                const lastUDP = state.udpPackets[state.udpPackets.length - 1];
                if (lastUDP.isLost) {
                    // Lost: 1s animation + 1s message
                    udpAnimationTime = 1000 + 1000;
                } else {
                    // Delivered: 2s animation + 1s message
                    udpAnimationTime = 2000 + 1000;
                }
            }
            
            // Use the longer animation time + small buffer
            const maxAnimationTime = Math.max(tcpAnimationTime, udpAnimationTime) + 500;
            
            // Animate packets based on their status
            animatePackets();
            
            // Update UI after animations complete
            setTimeout(() => {
                renderState();
                isProcessing = false;
            }, maxAnimationTime);
        })
        .catch(error => {
            console.error('Error processing packets:', error);
            isProcessing = false;
            showLaneMessage('tcp', '❌ Error processing');
            showLaneMessage('udp', '❌ Error processing');
        });
}

function animatePackets() {
    // Animate TCP packets
    if (state.tcpPackets && state.tcpPackets.length > 0) {
        const tcpLane = document.getElementById('tcp-lane');
        const lastPacket = state.tcpPackets[state.tcpPackets.length - 1];
        
        // Only animate the most recent packet if it's just been processed
        if (lastPacket.status === 'acknowledged' || lastPacket.status === 'lost') {
            clearLaneMessage('tcp');
            
            if (lastPacket.status === 'lost') {
                // TCP gave up after max retries (very rare!)
                const packetDiv = createPacketElement('tcp', lastPacket.data, false);
                tcpLane.appendChild(packetDiv);
                
                showLaneMessage('tcp', `🔁 TCP attempting delivery (max 3 retries)...`);
                
                // Show multiple retry attempts
                let retryDelay = 0;
                for (let i = 0; i <= lastPacket.retryCount; i++) {
                    setTimeout(() => {
                        if (i > 0) {
                            packetDiv.style.animation = 'none';
                            setTimeout(() => {
                                packetDiv.style.left = '0';
                                packetDiv.style.opacity = '1';
                                packetDiv.style.animation = 'packetLost 1s ease-out forwards';
                            }, 10);
                        } else {
                            packetDiv.style.animation = 'packetLost 1s ease-out forwards';
                        }
                        showLaneMessage('tcp', `❌ Attempt ${i + 1} failed${i < 3 ? ' - retrying...' : ' - MAX RETRIES REACHED!'}`);
                    }, retryDelay);
                    retryDelay += 1200;
                }
                
                setTimeout(() => {
                    packetDiv.remove();
                    showLaneMessage('tcp', '💔 TCP connection timeout - packet lost after 3 retries');
                    setTimeout(() => clearLaneMessage('tcp'), 3000);
                }, retryDelay + 1000);
                
            } else if (lastPacket.isLost && lastPacket.retryCount > 0) {
                // Show initial loss and successful retry
                const packetDiv = createPacketElement('tcp', lastPacket.data, false);
                tcpLane.appendChild(packetDiv);
                
                // Show loss animations for each retry
                let retryDelay = 0;
                for (let i = 0; i < lastPacket.retryCount; i++) {
                    setTimeout(() => {
                        if (i > 0) {
                            packetDiv.style.animation = 'none';
                            setTimeout(() => {
                                packetDiv.style.left = '0';
                                packetDiv.style.opacity = '1';
                                packetDiv.style.animation = 'packetLost 1s ease-out forwards';
                            }, 10);
                        } else {
                            packetDiv.style.animation = 'packetLost 1s ease-out forwards';
                        }
                        showLaneMessage('tcp', `❌ Attempt ${i + 1} lost - retrying...`);
                    }, retryDelay);
                    retryDelay += 1200;
                }
                
                // Final successful attempt
                setTimeout(() => {
                    packetDiv.style.animation = 'none';
                    setTimeout(() => {
                        packetDiv.style.left = '0';
                        packetDiv.style.opacity = '1';
                        packetDiv.style.borderColor = '#22c55e';
                        packetDiv.style.borderWidth = '3px';
                        packetDiv.style.animation = 'movePacket 2s ease-in-out forwards';
                        showLaneMessage('tcp', `🔁 Retry #${lastPacket.retryCount} successful!`);
                    }, 10);
                    
                    setTimeout(() => {
                        packetDiv.remove();
                        showLaneMessage('tcp', '✅ Packet delivered!');
                        
                        // Show ACK coming back
                        setTimeout(() => {
                            const ackDiv = createAckElement();
                            tcpLane.appendChild(ackDiv);
                            showLaneMessage('tcp', '✅ ACK received - Complete!');
                            
                            setTimeout(() => {
                                ackDiv.remove();
                                clearLaneMessage('tcp');
                            }, 1000);
                        }, 500);
                    }, 2000);
                }, retryDelay);
                
            } else {
                // Successful on first attempt
                const packetDiv = createPacketElement('tcp', lastPacket.data, false);
                tcpLane.appendChild(packetDiv);
                
                setTimeout(() => {
                    packetDiv.remove();
                    showLaneMessage('tcp', '✅ Packet delivered!');
                    
                    // Show ACK coming back
                    setTimeout(() => {
                        const ackDiv = createAckElement();
                        tcpLane.appendChild(ackDiv);
                        showLaneMessage('tcp', '✅ ACK received - Complete!');
                        
                        setTimeout(() => {
                            ackDiv.remove();
                            clearLaneMessage('tcp');
                        }, 1000);
                    }, 500);
                }, 2000);
            }
        }
    }
    
    // Animate UDP packets (unchanged)
    if (state.udpPackets && state.udpPackets.length > 0) {
        const udpLane = document.getElementById('udp-lane');
        const lastPacket = state.udpPackets[state.udpPackets.length - 1];
        
        // Only animate the most recent packet if it's just been processed
        if (lastPacket.status === 'delivered' || lastPacket.status === 'lost') {
            clearLaneMessage('udp');
            
            const packetDiv = createPacketElement('udp', lastPacket.data, lastPacket.isLost);
            udpLane.appendChild(packetDiv);
            
            setTimeout(() => {
                if (lastPacket.isLost) {
                    // Packet lost animation
                    packetDiv.style.animation = 'packetLost 1s ease-out forwards';
                    showLaneMessage('udp', '❌ Packet lost!');
                    
                    setTimeout(() => {
                        packetDiv.remove();
                        showLaneMessage('udp', '❌ Lost - No retry (UDP)');
                        setTimeout(() => clearLaneMessage('udp'), 2000);
                    }, 1000);
                } else {
                    // Packet delivered
                    setTimeout(() => {
                        packetDiv.remove();
                        showLaneMessage('udp', '✅ Delivered - No ACK (UDP)');
                        setTimeout(() => clearLaneMessage('udp'), 2000);
                    }, 2000);
                }
            }, 100);
        }
    }
}

function createPacketElement(protocol, data, isLost) {
    const div = document.createElement('div');
    div.className = `packet-transit ${protocol} ${isLost ? 'lost' : ''}`;
    div.textContent = data.substring(0, 3);
    div.style.top = '20px';
    div.style.left = '0';
    div.style.animation = 'movePacket 2s ease-in-out forwards';
    return div;
}

function createAckElement() {
    const div = document.createElement('div');
    div.className = 'packet-transit ack';
    div.textContent = 'ACK';
    div.style.top = '20px';
    div.style.right = '0';
    div.style.left = 'auto';
    div.style.animation = 'moveAck 1s ease-in-out forwards';
    return div;
}

function showLaneMessage(protocol, message) {
    const laneId = protocol === 'tcp' ? 'tcp-lane' : 'udp-lane';
    const lane = document.getElementById(laneId);
    
    // Remove existing message
    const existingMsg = lane.querySelector('.lane-message');
    if (existingMsg) existingMsg.remove();
    
    // Add new message
    const msgDiv = document.createElement('div');
    msgDiv.className = 'lane-message';
    msgDiv.textContent = message;
    lane.appendChild(msgDiv);
}

function clearLaneMessage(protocol) {
    const laneId = protocol === 'tcp' ? 'tcp-lane' : 'udp-lane';
    const lane = document.getElementById(laneId);
    const msg = lane.querySelector('.lane-message');
    if (msg) msg.remove();
}

function renderTCPPackets() {
    const container = document.getElementById('tcp-packets');
    container.innerHTML = '';
    
    if (!state.tcpPackets || state.tcpPackets.length === 0) {
        container.innerHTML = '<div style="color: var(--text-muted); padding: 10px;">No packets sent yet. Click "Send TCP Packet" to start.</div>';
        return;
    }
    
    // Filter out pending packets (not yet processed)
    const processedPackets = state.tcpPackets.filter(p => p.status !== 'pending');
    
    if (processedPackets.length === 0) {
        container.innerHTML = '<div style="color: var(--text-muted); padding: 10px;">Packets queued. Processing...</div>';
        return;
    }
    
    // Show most recent first
    const packets = [...processedPackets].reverse();
    
    packets.forEach(packet => {
        const item = document.createElement('div');
        item.className = `packet-item ${packet.status}`;
        item.innerHTML = `
            <span><strong>#${packet.seqNumber}</strong>: ${packet.data}</span>
            <span>${getStatusIcon(packet.status, packet.retryCount)}</span>
        `;
        container.appendChild(item);
    });
}

function renderUDPPackets() {
    const container = document.getElementById('udp-packets');
    container.innerHTML = '';
    
    if (!state.udpPackets || state.udpPackets.length === 0) {
        container.innerHTML = '<div style="color: var(--text-muted); padding: 10px;">No packets sent yet. Click "Send UDP Packet" to start.</div>';
        return;
    }
    
    // Filter out pending packets (not yet processed)
    const processedPackets = state.udpPackets.filter(p => p.status !== 'pending');
    
    if (processedPackets.length === 0) {
        container.innerHTML = '<div style="color: var(--text-muted); padding: 10px;">Packets queued. Processing...</div>';
        return;
    }
    
    // Show most recent first
    const packets = [...processedPackets].reverse();
    
    packets.forEach(packet => {
        const item = document.createElement('div');
        item.className = `packet-item ${packet.status}`;
        item.innerHTML = `
            <span><strong>#${packet.seqNumber}</strong>: ${packet.data}</span>
            <span>${getStatusIcon(packet.status, 0)}</span>
        `;
        container.appendChild(item);
    });
}

function getStatusIcon(status, retryCount) {
    switch(status) {
        case 'delivered': return '✅ Delivered';
        case 'acknowledged': 
            if (retryCount > 0) {
                return `✅ ACK (${retryCount} ${retryCount === 1 ? 'retry' : 'retries'})`;
            }
            return '✅ ACK';
        case 'lost': return '❌ Lost';
        case 'in_transit': return '📤 Sending...';
        case 'pending': return '⏳ Queued';
        default: return '❓ Unknown';
    }
}

function renderTCPStats() {
    if (!state.stats || !state.stats.tcp) return;
    
    const tcp = state.stats.tcp;
    document.getElementById('tcp-sent').textContent = tcp.packetsSent;
    document.getElementById('tcp-delivered').textContent = tcp.packetsDelivered;
    document.getElementById('tcp-retried').textContent = tcp.packetsRetried;
    document.getElementById('tcp-lost').textContent = tcp.packetsLost || 0;
    document.getElementById('tcp-acks').textContent = tcp.acknowledgments;
    document.getElementById('tcp-rate').textContent = `${tcp.deliveryRate.toFixed(1)}%`;
    
    // Highlight delivery rate
    const rateEl = document.getElementById('tcp-rate');
    if (tcp.deliveryRate === 100) {
        rateEl.style.color = 'var(--accent-green)';
    } else if (tcp.deliveryRate >= 95) {
        rateEl.style.color = '#f59e0b'; // Orange for slightly less than perfect
    } else {
        rateEl.style.color = 'var(--accent-red)';
    }
    
    // Highlight lost packets
    const lostEl = document.getElementById('tcp-lost');
    if (tcp.packetsLost > 0) {
        lostEl.style.color = 'var(--accent-red)';
        lostEl.style.fontWeight = 'bold';
    } else {
        lostEl.style.color = 'var(--accent-green)';
    }
}

function renderUDPStats() {
    if (!state.stats || !state.stats.udp) return;
    
    const udp = state.stats.udp;
    document.getElementById('udp-sent').textContent = udp.packetsSent;
    document.getElementById('udp-delivered').textContent = udp.packetsDelivered;
    document.getElementById('udp-lost').textContent = udp.packetsLost;
    document.getElementById('udp-rate').textContent = `${udp.deliveryRate.toFixed(1)}%`;
    
    // Color code delivery rate
    const rateEl = document.getElementById('udp-rate');
    if (udp.deliveryRate >= 80) {
        rateEl.style.color = 'var(--accent-green)';
    } else if (udp.deliveryRate >= 50) {
        rateEl.style.color = '#f59e0b';
    } else {
        rateEl.style.color = 'var(--accent-red)';
    }
}

function setPacketLossRate(rate) {
    fetch(`${API_BASE}/set-loss-rate?rate=${rate}`, addSessionHeader({ method: 'POST' }))
        .then(response => response.json())
        .then(data => {
            state = data;
            showLaneMessage('tcp', `📡 Packet loss rate: ${(rate * 100).toFixed(0)}%`);
            showLaneMessage('udp', `📡 Packet loss rate: ${(rate * 100).toFixed(0)}%`);
            setTimeout(() => {
                clearLaneMessage('tcp');
                clearLaneMessage('udp');
            }, 2000);
        })
        .catch(error => console.error('Error setting loss rate:', error));
}

function resetSimulation() {
    if (!confirm('Reset the TCP/UDP simulation? All packets and statistics will be cleared.')) return;
    
    fetch(`${API_BASE}/reset`, addSessionHeader({ method: 'POST' }))
        .then(response => response.json())
        .then(data => {
            state = data;
            renderState();
            updateConnectionButtons();
            showLaneMessage('tcp', '🔄 Simulation reset');
            showLaneMessage('udp', '🔄 Simulation reset');
            setTimeout(() => {
                clearLaneMessage('tcp');
                clearLaneMessage('udp');
            }, 2000);
        })
        .catch(error => console.error('Error resetting simulation:', error));
}

function establishConnection() {
    const connectBtn = document.getElementById('connect-btn');
    connectBtn.disabled = true;
    connectBtn.textContent = '⏳ Connecting...';
    
    showLaneMessage('tcp', '🔌 Establishing TCP connection...');
    showLaneMessage('udp', 'ℹ️ UDP is connectionless');
    
    fetch(`${API_BASE}/establish-connection`, addSessionHeader({ method: 'POST' }))
        .then(response => response.json())
        .then(data => {
            state = data.state;
            const steps = data.steps;
            
            // Animate 3-way handshake
            animateHandshake(steps);
            
            setTimeout(() => {
                updateConnectionButtons();
                showLaneMessage('tcp', '✅ TCP connection established!');
                setTimeout(() => clearLaneMessage('tcp'), 2000);
            }, 3500);
            
            setTimeout(() => clearLaneMessage('udp'), 2000);
        })
        .catch(error => {
            console.error('Error establishing connection:', error);
            showLaneMessage('tcp', '❌ Connection failed');
            connectBtn.disabled = false;
            connectBtn.textContent = '🔌 Establish TCP Connection';
        });
}

function closeConnection() {
    const disconnectBtn = document.getElementById('disconnect-btn');
    disconnectBtn.disabled = true;
    disconnectBtn.textContent = '⏳ Closing...';
    
    showLaneMessage('tcp', '👋 Closing TCP connection...');
    
    fetch(`${API_BASE}/close-connection`, addSessionHeader({ method: 'POST' }))
        .then(response => response.json())
        .then(data => {
            state = data.state;
            updateConnectionButtons();
            showLaneMessage('tcp', '🔌 TCP connection closed');
            setTimeout(() => clearLaneMessage('tcp'), 2000);
        })
        .catch(error => {
            console.error('Error closing connection:', error);
            showLaneMessage('tcp', '❌ Error closing connection');
        });
}

function animateHandshake(steps) {
    const tcpLane = document.getElementById('tcp-lane');
    
    // Step 1: SYN (Client → Server)
    setTimeout(() => {
        const synPacket = createHandshakePacket('SYN');
        tcpLane.appendChild(synPacket);
        showLaneMessage('tcp', '→ SYN sent');
        setTimeout(() => synPacket.remove(), 1000);
    }, 100);
    
    // Step 2: SYN-ACK (Server → Client)
    setTimeout(() => {
        const synAckPacket = createHandshakePacket('SYN-ACK');
        synAckPacket.style.right = '0';
        synAckPacket.style.left = 'auto';
        synAckPacket.style.animation = 'moveAck 1s ease-in-out forwards';
        tcpLane.appendChild(synAckPacket);
        showLaneMessage('tcp', '← SYN-ACK received');
        setTimeout(() => synAckPacket.remove(), 1000);
    }, 1200);
    
    // Step 3: ACK (Client → Server)
    setTimeout(() => {
        const ackPacket = createHandshakePacket('ACK');
        tcpLane.appendChild(ackPacket);
        showLaneMessage('tcp', '→ ACK sent');
        setTimeout(() => ackPacket.remove(), 1000);
    }, 2400);
}

function createHandshakePacket(label) {
    const div = document.createElement('div');
    div.className = 'packet-transit tcp';
    div.textContent = label;
    div.style.top = '20px';
    div.style.left = '0';
    div.style.fontSize = '0.7rem';
    div.style.animation = 'movePacket 1s ease-in-out forwards';
    return div;
}

document.addEventListener('DOMContentLoaded', initVisualization);
