// Theme management
(function() {
    const savedTheme = localStorage.getItem('theme') || 'dark';
    document.documentElement.setAttribute('data-theme', savedTheme);
})();

const API_BASE = 'http://localhost:8080/api/pagination';

function addSessionHeader(options = {}) {
    if (typeof window.SDS_SESSION !== 'undefined') {
        return window.SDS_SESSION.addSessionHeader(options);
    }
    return options;
}

let state = null;
let allData = [];
let currentPage = 1;

// Initialize
document.addEventListener('DOMContentLoaded', () => {
    console.log('🔧 Initializing Pagination vs Virtualization...');
    
    // Setup event listeners
    document.getElementById('load-data-btn').addEventListener('click', loadInitialData);
    document.getElementById('reset-btn').addEventListener('click', resetSimulation);
    
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
    
    // Show initial message
    showInitialMessage();
});

function showInitialMessage() {
    const paginationContainer = document.getElementById('pagination-container');
    const virtualContainer = document.getElementById('virtual-container');
    const virtualScrollArea = document.getElementById('virtual-scroll-area');
    
    if (paginationContainer) {
        paginationContainer.innerHTML = '<div class="loading-message">Click "Load Data" to start</div>';
    }
    
    if (virtualContainer) {
        const loadingMsg = virtualContainer.querySelector('.loading-message');
        if (loadingMsg) {
            loadingMsg.textContent = 'Click "Load Data" to start';
            loadingMsg.style.display = 'block';
        }
    }
    
    if (virtualScrollArea) {
        virtualScrollArea.style.display = 'none';
    }
}

// === LOAD INITIAL DATA ===
function loadInitialData() {
    const loadBtn = document.getElementById('load-data-btn');
    loadBtn.disabled = true;
    loadBtn.textContent = '⏳ Loading...';
    
    // Load first page for pagination
    loadPage(1);
    
    // Load all data for virtualization
    loadAllData();
    
    setTimeout(() => {
        loadBtn.disabled = false;
        loadBtn.textContent = '📥 Load Data';
    }, 1000);
}

// === PAGINATION ===
function loadPage(pageNumber) {
    const container = document.getElementById('pagination-container');
    container.innerHTML = '<div class="loading-message">⏳ Loading page ' + pageNumber + '...</div>';
    
    fetch(`${API_BASE}/get-page?page=${pageNumber}`, addSessionHeader())
        .then(response => response.json())
        .then(data => {
            currentPage = data.pageNumber;
            renderPage(data);
            renderPaginationControls(data);
            updatePaginationStats();
        })
        .catch(error => {
            console.error('Error loading page:', error);
            container.innerHTML = '<div class="loading-message">❌ Error loading page</div>';
        });
}

function renderPage(pageData) {
    const container = document.getElementById('pagination-container');
    container.innerHTML = '';
    
    pageData.items.forEach(item => {
        const itemEl = document.createElement('div');
        itemEl.className = 'item-row';
        itemEl.innerHTML = `
            <div class="item-title">${item.title}</div>
            <div class="item-description">${item.description}</div>
        `;
        container.appendChild(itemEl);
    });
    
    // Add scroll listener for infinite scroll
    container.onscroll = () => {
        const scrollTop = container.scrollTop;
        const scrollHeight = container.scrollHeight;
        const clientHeight = container.clientHeight;
        
        // Check if scrolled to bottom (with 50px threshold)
        if (scrollTop + clientHeight >= scrollHeight - 50) {
            // Load next page if available
            if (pageData.hasNextPage && !container.dataset.loading) {
                container.dataset.loading = 'true';
                
                // Add loading indicator
                const loadingEl = document.createElement('div');
                loadingEl.className = 'loading-message';
                loadingEl.textContent = '⏳ Loading more...';
                container.appendChild(loadingEl);
                
                // Load next page
                setTimeout(() => {
                    loadPage(currentPage + 1);
                }, 300);
            }
        }
    };
    
    delete container.dataset.loading;
}

function renderPaginationControls(pageData) {
    const pageNumbersContainer = document.getElementById('page-numbers');
    pageNumbersContainer.innerHTML = '';
    
    // Show 5 page numbers around current page
    const start = Math.max(1, currentPage - 2);
    const end = Math.min(pageData.totalPages, start + 4);
    
    for (let i = start; i <= end; i++) {
        const pageBtn = document.createElement('div');
        pageBtn.className = 'page-number' + (i === currentPage ? ' active' : '');
        pageBtn.textContent = i;
        pageBtn.addEventListener('click', () => loadPage(i));
        pageNumbersContainer.appendChild(pageBtn);
    }
    
    // Update prev/next buttons
    const prevBtn = document.querySelector('[data-page="prev"]');
    const nextBtn = document.querySelector('[data-page="next"]');
    
    prevBtn.disabled = !pageData.hasPrevPage;
    nextBtn.disabled = !pageData.hasNextPage;
    
    prevBtn.onclick = () => loadPage(currentPage - 1);
    nextBtn.onclick = () => loadPage(currentPage + 1);
}

function updatePaginationStats() {
    fetch(`${API_BASE}/state`, addSessionHeader())
        .then(response => response.json())
        .then(data => {
            const stats = data.paginationStats;
            document.getElementById('pagination-api-calls').textContent = stats.totalApiCalls;
            document.getElementById('pagination-items-fetched').textContent = stats.totalItemsFetched;
            document.getElementById('pagination-latency').textContent = stats.averageLatencyMs.toFixed(0) + 'ms';
            document.getElementById('pagination-current-page').textContent = stats.lastPageViewed || '-';
            
            // Render network log
            renderNetworkLog(stats.recentRequests);
        });
}

function renderNetworkLog(requests) {
    const logContainer = document.getElementById('pagination-network');
    logContainer.innerHTML = '';
    
    if (requests.length === 0) {
        logContainer.innerHTML = '<div style="color: var(--text-muted); font-size: 0.85rem;">No API calls yet</div>';
        return;
    }
    
    requests.slice().reverse().forEach(req => {
        const callEl = document.createElement('div');
        callEl.className = 'network-call';
        callEl.textContent = `📡 GET /page?page=${req.pageNumber} (${req.pageSize} items)`;
        logContainer.appendChild(callEl);
    });
}

// === VIRTUALIZATION ===
function loadAllData() {
    const container = document.getElementById('virtual-container');
    const scrollArea = document.getElementById('virtual-scroll-area');
    const loadingMsg = container.querySelector('.loading-message');
    
    // Hide scroll area and show loading message
    if (scrollArea) scrollArea.style.display = 'none';
    if (loadingMsg) {
        loadingMsg.textContent = '⏳ Loading 10,000 items...';
        loadingMsg.style.display = 'block';
    }
    
    fetch(`${API_BASE}/load-all`, addSessionHeader())
        .then(response => response.json())
        .then(data => {
            allData = data.items;
            
            // Hide loading message, show scroll area
            if (loadingMsg) loadingMsg.style.display = 'none';
            if (scrollArea) scrollArea.style.display = 'block';
            
            // Initialize virtual scroll
            initVirtualScroll();
            updateVirtualizationStats();
        })
        .catch(error => {
            console.error('Error loading data:', error);
            if (loadingMsg) {
                loadingMsg.textContent = '❌ Error loading data';
                loadingMsg.style.display = 'block';
            }
        });
}

let virtualScrollHandler = null; // Store handler reference

function initVirtualScroll() {
    const scrollArea = document.getElementById('virtual-scroll-area');
    const viewport = document.getElementById('virtual-viewport');
    
    if (!scrollArea || !viewport) {
        console.error('Virtual scroll elements not found');
        return;
    }
    
    // Item height includes padding (12px * 2) + margin-bottom (12px to match pagination) + content (~30px)
    const itemHeight = 74; // Increased to match pagination spacing
    const totalHeight = allData.length * itemHeight;
    
    // Set the total scrollable height
    scrollArea.style.height = '400px';
    viewport.style.height = totalHeight + 'px';
    
    // Reset scroll position to top
    scrollArea.scrollTop = 0;
    
    // Render initial visible items (starting from item 1)
    renderVisibleItems(0);
    
    // Remove existing scroll listener if any
    if (virtualScrollHandler) {
        scrollArea.removeEventListener('scroll', virtualScrollHandler);
    }
    
    // Create new scroll handler with debouncing
    let scrollTimeout = null;
    virtualScrollHandler = () => {
        const scrollTop = scrollArea.scrollTop;
        renderVisibleItems(scrollTop);
        
        // Update backend stats (debounced)
        if (scrollTimeout) {
            clearTimeout(scrollTimeout);
        }
        scrollTimeout = setTimeout(() => {
            fetch(`${API_BASE}/update-viewport?scroll=${Math.floor(scrollTop)}&height=400`, 
                addSessionHeader({ method: 'POST' }))
                .then(() => updateVirtualizationStats());
        }, 100); // Debounce by 100ms
    };
    
    // Add scroll listener
    scrollArea.addEventListener('scroll', virtualScrollHandler);
}

function renderVisibleItems(scrollTop) {
    const viewport = document.getElementById('virtual-viewport');
    const itemHeight = 74; // Match the value in initVirtualScroll
    const viewportHeight = 400;
    
    // Calculate visible range
    const startIndex = Math.floor(scrollTop / itemHeight);
    const endIndex = Math.min(allData.length, startIndex + Math.ceil(viewportHeight / itemHeight) + 1);
    
    // Clear and render only visible items
    viewport.innerHTML = '';
    
    for (let i = startIndex; i < endIndex; i++) {
        const item = allData[i];
        const itemEl = document.createElement('div');
        itemEl.className = 'item-row';
        itemEl.style.position = 'absolute';
        itemEl.style.top = (i * itemHeight) + 'px';
        itemEl.style.left = '0';
        itemEl.style.right = '0';
        itemEl.style.height = (itemHeight - 12) + 'px'; // Subtract margin-bottom (12px to match pagination)
        itemEl.style.marginBottom = '12px'; // Match pagination spacing
        itemEl.innerHTML = `
            <div class="item-title">${item.title}</div>
            <div class="item-description">${item.description}</div>
        `;
        viewport.appendChild(itemEl);
    }
}

function updateVirtualizationStats() {
    fetch(`${API_BASE}/state`, addSessionHeader())
        .then(response => response.json())
        .then(data => {
            const stats = data.virtualizationStats;
            document.getElementById('virtual-api-calls').textContent = stats.totalApiCalls;
            document.getElementById('virtual-in-memory').textContent = stats.itemsInMemory.toLocaleString();
            document.getElementById('virtual-rendered').textContent = stats.itemsRendered;
            document.getElementById('virtual-memory').textContent = stats.memoryUsageKB.toLocaleString() + ' KB';
            document.getElementById('virtual-render-time').textContent = stats.renderTimeMs.toFixed(1) + 'ms';
            document.getElementById('virtual-visible-range').textContent = stats.visibleRange || '-';
        });
}

// === RESET ===
function resetSimulation() {
    if (!confirm('Reset both simulations?')) return;
    
    fetch(`${API_BASE}/reset`, addSessionHeader({ method: 'POST' }))
        .then(response => response.json())
        .then(() => {
            // Reset pagination
            currentPage = 1;
            
            // Reset virtualization
            allData = [];
            
            // Show initial messages
            showInitialMessage();
            
            // Update stats to show zeros
            document.getElementById('pagination-api-calls').textContent = '0';
            document.getElementById('pagination-items-fetched').textContent = '0';
            document.getElementById('pagination-current-page').textContent = '-';
            document.getElementById('pagination-network').innerHTML = '<div style="color: var(--text-muted); font-size: 0.85rem;">No API calls yet</div>';
            
            document.getElementById('virtual-api-calls').textContent = '0';
            document.getElementById('virtual-in-memory').textContent = '0';
            document.getElementById('virtual-rendered').textContent = '0';
            document.getElementById('virtual-memory').textContent = '0 KB';
            document.getElementById('virtual-render-time').textContent = '0ms';
            document.getElementById('virtual-visible-range').textContent = '-';
        })
        .catch(error => console.error('Error resetting:', error));
}

