// Session Management for Multi-User Support
// This module handles session ID generation and storage

/**
 * Generates a unique session ID for the current user
 * Format: session_<timestamp>_<random>
 * Example: session_1234567890_abc123def456
 */
function generateSessionID() {
    const timestamp = Date.now();
    const random = Math.random().toString(36).substring(2, 15);
    return `session_${timestamp}_${random}`;
}

/**
 * Gets or creates a session ID for the current user
 * Session ID is stored in localStorage to persist across page reloads
 * @returns {string} The session ID
 */
function getSessionID() {
    let sessionID = localStorage.getItem('sds_session_id');
    
    if (!sessionID) {
        sessionID = generateSessionID();
        localStorage.setItem('sds_session_id', sessionID);
        console.log('🆕 New session created:', sessionID);
    } else {
        console.log('♻️ Existing session loaded:', sessionID);
    }
    
    return sessionID;
}

/**
 * Clears the current session (useful for testing multi-user scenarios)
 */
function clearSession() {
    localStorage.removeItem('sds_session_id');
    console.log('🧹 Session cleared - reload page to get new session');
}

/**
 * Adds session ID to fetch headers
 * @param {Object} options - Fetch options object
 * @returns {Object} Modified options with session header
 */
function addSessionHeader(options = {}) {
    const sessionID = getSessionID();
    
    if (!options.headers) {
        options.headers = {};
    }
    
    options.headers['X-Session-ID'] = sessionID;
    
    return options;
}

// Initialize session on page load
const SESSION_ID = getSessionID();

// Export for use in other scripts
if (typeof window !== 'undefined') {
    window.SDS_SESSION = {
        getSessionID,
        clearSession,
        addSessionHeader,
        SESSION_ID
    };
}

