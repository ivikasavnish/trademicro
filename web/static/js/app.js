// TradeMicro Dashboard Application

// Global state
const state = {
    user: null,
    token: null,
    trades: [],
    symbols: [],
    brokerTokens: [],
    tasks: []
};

// API Base URL
const API_BASE_URL = '/api';

// DOM Elements
document.addEventListener('DOMContentLoaded', () => {
    // Add Family Member button
    const addFamilyBtn = document.createElement('button');
    addFamilyBtn.id = 'new-family-btn';
    addFamilyBtn.className = 'primary-btn';
    addFamilyBtn.innerHTML = '<i class="fas fa-plus"></i> New Family Member';
    const dashboardSection = document.getElementById('dashboard');
    dashboardSection && dashboardSection.prepend(addFamilyBtn);

    addFamilyBtn.addEventListener('click', openFamilyMemberModal);
    setupFamilyMemberModal();
    // Check if user is logged in
    const token = localStorage.getItem('token');
    if (token) {
        state.token = token;
        fetchUserInfo();
        initializeDashboard();
    } else {
        showLoginModal();
    }

    // Navigation
    setupNavigation();

    // Form submissions
    setupFormHandlers();

    // Initialize modals
    setupModals();
});

// Authentication functions
async function login(username, password) {
    try {
        const response = await fetch(`${API_BASE_URL}/login`, {
            method: 'POST',
            headers: {
                'Content-Type': 'application/json'
            },
            body: JSON.stringify({ username, password })
        });

        if (!response.ok) {
            throw new Error('Login failed');
        }

        const data = await response.json();
        state.token = data.token;
        localStorage.setItem('token', data.token);
        
        hideLoginModal();
        fetchUserInfo();
        initializeDashboard();
    } catch (error) {
        showError('Login failed: ' + error.message);
    }
}

async function logout() {
    localStorage.removeItem('token');
    state.token = null;
    state.user = null;
    showLoginModal();
}

async function fetchUserInfo() {
    try {
        const response = await fetchWithAuth(`${API_BASE_URL}/user`);
        const data = await response.json();
        state.user = data;
        document.getElementById('username').textContent = data.username;
    } catch (error) {
        console.error('Error fetching user info:', error);
    }
}

// Dashboard initialization
async function initializeDashboard() {
    checkApiHealth();
    loadCounts();
    loadAllData();
}

async function checkApiHealth() {
    try {
        const response = await fetch(`${API_BASE_URL}/health`);
        const data = await response.json();
        
        document.getElementById('api-status').textContent = data.status;
        document.getElementById('api-version').textContent = data.version;
        document.getElementById('server-time').textContent = new Date(data.timestamp).toLocaleString();
    } catch (error) {
        document.getElementById('api-status').textContent = 'Error';
        console.error('Health check failed:', error);
    }
}

async function loadCounts() {
    try {
        // Fetch counts for dashboard
        const [trades, symbols, tokens, tasks] = await Promise.all([
            fetchWithAuth(`${API_BASE_URL}/trades?count=true`),
            fetchWithAuth(`${API_BASE_URL}/symbols?count=true`),
            fetchWithAuth(`${API_BASE_URL}/broker-tokens?count=true`),
            fetchWithAuth(`${API_BASE_URL}/tasks?count=true`)
        ]);

        const tradesData = await trades.json();
        const symbolsData = await symbols.json();
        const tokensData = await tokens.json();
        const tasksData = await tasks.json();

        document.getElementById('trades-count').textContent = tradesData.count;
        document.getElementById('symbols-count').textContent = symbolsData.count;
        document.getElementById('tokens-count').textContent = tokensData.count;
        document.getElementById('tasks-count').textContent = tasksData.count;
    } catch (error) {
        console.error('Error loading counts:', error);
    }
}

async function loadAllData() {
    await Promise.all([
        loadTrades(),
        loadSymbols(),
        loadBrokerTokens(),
        loadTasks()
    ]);
}

// Data loading functions
async function loadTrades() {
    try {
        const response = await fetchWithAuth(`${API_BASE_URL}/trades`);
        const data = await response.json();
        state.trades = data;
        renderTrades();
    } catch (error) {
        console.error('Error loading trades:', error);
    }
}

async function loadSymbols() {
    try {
        const response = await fetchWithAuth(`${API_BASE_URL}/symbols`);
        const data = await response.json();
        state.symbols = data;
        renderSymbols();
    } catch (error) {
        console.error('Error loading symbols:', error);
    }
}

async function loadBrokerTokens() {
    try {
        const response = await fetchWithAuth(`${API_BASE_URL}/broker-tokens`);
        const data = await response.json();
        state.brokerTokens = data;
        renderBrokerTokens();
    } catch (error) {
        console.error('Error loading broker tokens:', error);
    }
}

async function loadTasks() {
    try {
        const response = await fetchWithAuth(`${API_BASE_URL}/tasks`);
        const data = await response.json();
        state.tasks = data;
        renderTasks();
    } catch (error) {
        console.error('Error loading tasks:', error);
    }
}

// Rendering functions
function renderTrades() {
    const tbody = document.querySelector('#trades-table tbody');
    tbody.innerHTML = '';

    if (state.trades.length === 0) {
        tbody.innerHTML = `<tr><td colspan="8" class="empty-message">No trades found</td></tr>`;
        return;
    }

    state.trades.forEach(trade => {
        const tr = document.createElement('tr');
        tr.innerHTML = `
            <td>${trade.id}</td>
            <td>${trade.symbol}</td>
            <td>${trade.type}</td>
            <td>${trade.quantity}</td>
            <td>${trade.price}</td>
            <td><span class="status-badge ${trade.status.toLowerCase()}">${trade.status}</span></td>
            <td>${new Date(trade.created_at).toLocaleString()}</td>
            <td>
                <button class="action-btn edit" data-id="${trade.id}"><i class="fas fa-edit"></i></button>
                <button class="action-btn delete" data-id="${trade.id}"><i class="fas fa-trash"></i></button>
            </td>
        `;
        tbody.appendChild(tr);
    });
}

function renderSymbols() {
    const tbody = document.querySelector('#symbols-table tbody');
    tbody.innerHTML = '';

    if (state.symbols.length === 0) {
        tbody.innerHTML = `<tr><td colspan="5" class="empty-message">No symbols found</td></tr>`;
        return;
    }

    state.symbols.forEach(symbol => {
        const tr = document.createElement('tr');
        tr.innerHTML = `
            <td>${symbol.id}</td>
            <td>${symbol.name}</td>
            <td>${symbol.exchange}</td>
            <td>${symbol.type}</td>
            <td>
                <button class="action-btn edit" data-id="${symbol.id}"><i class="fas fa-edit"></i></button>
                <button class="action-btn delete" data-id="${symbol.id}"><i class="fas fa-trash"></i></button>
            </td>
        `;
        tbody.appendChild(tr);
    });
}

function renderBrokerTokens() {
    const tbody = document.querySelector('#tokens-table tbody');
    tbody.innerHTML = '';

    if (state.brokerTokens.length === 0) {
        tbody.innerHTML = `<tr><td colspan="6" class="empty-message">No broker tokens found</td></tr>`;
        return;
    }

    state.brokerTokens.forEach(token => {
        // Calculate expiry date (1 month from creation)
        const createdAt = new Date(token.created_at);
        const expiresAt = new Date(createdAt);
        expiresAt.setMonth(expiresAt.getMonth() + 1);
        
        // Check if token is expired or about to expire
        const now = new Date();
        const daysUntilExpiry = Math.floor((expiresAt - now) / (1000 * 60 * 60 * 24));
        let expiryClass = '';
        
        if (daysUntilExpiry < 0) {
            expiryClass = 'status-badge failed';
        } else if (daysUntilExpiry < 7) {
            expiryClass = 'status-badge pending';
        }

        const tr = document.createElement('tr');
        tr.innerHTML = `
            <td>${token.id}</td>
            <td>${token.broker}</td>
            <td>${maskToken(token.token)}</td>
            <td>${new Date(token.created_at).toLocaleString()}</td>
            <td><span class="${expiryClass}">${expiresAt.toLocaleDateString()} (${daysUntilExpiry} days)</span></td>
            <td>
                <button class="action-btn edit" data-id="${token.id}"><i class="fas fa-edit"></i></button>
                <button class="action-btn delete" data-id="${token.id}"><i class="fas fa-trash"></i></button>
            </td>
        `;
        tbody.appendChild(tr);
        
        // Add event listeners for token actions
        tr.querySelector('.edit').addEventListener('click', () => openTokenModal(token));
        tr.querySelector('.delete').addEventListener('click', () => deleteToken(token.id));
    });
}

function renderTasks() {
    const tbody = document.querySelector('#tasks-table tbody');
    tbody.innerHTML = '';

    if (state.tasks.length === 0) {
        tbody.innerHTML = `<tr><td colspan="6" class="empty-message">No tasks found</td></tr>`;
        return;
    }

    // Filter tasks based on selected status
    const statusFilter = document.getElementById('task-status-filter').value;
    const filteredTasks = statusFilter === 'all' ? 
        state.tasks : 
        state.tasks.filter(task => task.status.toLowerCase() === statusFilter);

    if (filteredTasks.length === 0) {
        tbody.innerHTML = `<tr><td colspan="6" class="empty-message">No tasks with status "${statusFilter}" found</td></tr>`;
        return;
    }

    filteredTasks.forEach(task => {
        const tr = document.createElement('tr');
        tr.innerHTML = `
            <td>${task.id}</td>
            <td>${task.type}</td>
            <td><span class="status-badge ${task.status.toLowerCase()}">${task.status}</span></td>
            <td>${new Date(task.created_at).toLocaleString()}</td>
            <td>${task.updated_at ? new Date(task.updated_at).toLocaleString() : '-'}</td>
            <td>
                <button class="action-btn view" data-id="${task.id}"><i class="fas fa-eye"></i></button>
                ${task.status === 'PENDING' ? `<button class="action-btn delete" data-id="${task.id}"><i class="fas fa-trash"></i></button>` : ''}
            </td>
        `;
        tbody.appendChild(tr);
    });
}

// Broker Token functions
async function createToken(broker, token) {
    try {
        const response = await fetchWithAuth(`${API_BASE_URL}/broker-tokens`, {
            method: 'POST',
            headers: {
                'Content-Type': 'application/json'
            },
            body: JSON.stringify({ broker, token })
        });

        if (!response.ok) {
            throw new Error('Failed to create token');
        }

        await loadBrokerTokens();
        closeTokenModal();
    } catch (error) {
        showError('Error creating token: ' + error.message);
    }
}

async function updateToken(id, token) {
    try {
        const response = await fetchWithAuth(`${API_BASE_URL}/broker-tokens/${id}`, {
            method: 'PUT',
            headers: {
                'Content-Type': 'application/json'
            },
            body: JSON.stringify({ token })
        });

        if (!response.ok) {
            throw new Error('Failed to update token');
        }

        await loadBrokerTokens();
        closeTokenModal();
    } catch (error) {
        showError('Error updating token: ' + error.message);
    }
}

async function deleteToken(id) {
    if (!confirm('Are you sure you want to delete this token?')) {
        return;
    }

    try {
        const response = await fetchWithAuth(`${API_BASE_URL}/broker-tokens/${id}`, {
            method: 'DELETE'
        });

        if (!response.ok) {
            throw new Error('Failed to delete token');
        }

        await loadBrokerTokens();
    } catch (error) {
        showError('Error deleting token: ' + error.message);
    }
}

// --- Family Member Modal ---
function openFamilyMemberModal(member = null) {
    const modal = document.getElementById('family-member-modal');
    const form = document.getElementById('family-member-form');
    document.getElementById('family-modal-title').textContent = member ? 'Edit Family Member' : 'Add Family Member';
    document.getElementById('family-id').value = member ? member.id : '';
    document.getElementById('family-name-input').value = member ? member.name : '';
    document.getElementById('family-email-input').value = member ? member.email : '';
    document.getElementById('family-phone-input').value = member ? member.phone : '';
    document.getElementById('family-pin-input').value = '';
    document.getElementById('family-client-id-input').value = member ? member.client_id || '' : '';
    document.getElementById('family-client-token-input').value = member ? member.client_token || '' : '';
    modal.classList.add('active');
}

function closeFamilyMemberModal() {
    document.getElementById('family-member-modal').classList.remove('active');
}

function setupFamilyMemberModal() {
    document.getElementById('cancel-family-btn').addEventListener('click', closeFamilyMemberModal);
    document.querySelector('#family-member-modal .close-btn').addEventListener('click', closeFamilyMemberModal);
    document.getElementById('family-member-form').addEventListener('submit', async function(e) {
        e.preventDefault();
        const id = document.getElementById('family-id').value;
        const name = document.getElementById('family-name-input').value;
        const email = document.getElementById('family-email-input').value;
        const phone = document.getElementById('family-phone-input').value;
        const pin = document.getElementById('family-pin-input').value;
        const client_id = document.getElementById('family-client-id-input').value;
        const client_token = document.getElementById('family-client-token-input').value;
        const memberData = { name, email, phone, pin, client_id, client_token };
        try {
            const response = await fetchWithAuth(`${API_BASE_URL}/family-members`, {
                method: 'POST',
                headers: { 'Content-Type': 'application/json' },
                body: JSON.stringify(memberData)
            });
            if (!response.ok) throw new Error('Failed to add family member');
            closeFamilyMemberModal();
            // Optionally reload family members list here
        } catch (error) {
            showError('Error adding family member: ' + error.message);
        }
    });
}
// UI Setup functions
function setupNavigation() {
    const navLinks = document.querySelectorAll('nav a');
    navLinks.forEach(link => {
        link.addEventListener('click', (e) => {
            e.preventDefault();
            const section = link.getAttribute('data-section');
            
            // Update active link
            navLinks.forEach(l => l.classList.remove('active'));
            link.classList.add('active');
            
            // Show selected section
            document.querySelectorAll('main section').forEach(s => {
                s.classList.remove('active');
            });
            document.getElementById(section).classList.add('active');
        });
    });

    // Logout button
    document.getElementById('logout-btn').addEventListener('click', logout);

    // Task filter
    document.getElementById('task-status-filter').addEventListener('change', renderTasks);
}

function setupFormHandlers() {
    // Login form
    document.getElementById('login-form').addEventListener('submit', (e) => {
        e.preventDefault();
        const username = document.getElementById('username-input').value;
        const password = document.getElementById('password-input').value;
        login(username, password);
    });

    // Token form
    document.getElementById('token-form').addEventListener('submit', (e) => {
        e.preventDefault();
        const id = document.getElementById('token-id').value;
        const broker = document.getElementById('broker-input').value;
        const token = document.getElementById('token-input').value;
        
        if (id) {
            updateToken(id, token);
        } else {
            createToken(broker, token);
        }
    });

    // New token button
    document.getElementById('new-token-btn').addEventListener('click', () => {
        openTokenModal();
    });

    // Cancel token button
    document.getElementById('cancel-token-btn').addEventListener('click', closeTokenModal);
}

function setupModals() {
    // Close modal when clicking outside content
    document.querySelectorAll('.modal').forEach(modal => {
        modal.addEventListener('click', (e) => {
            if (e.target === modal) {
                modal.classList.remove('active');
            }
        });
    });

    // Close buttons
    document.querySelectorAll('.close-btn').forEach(btn => {
        btn.addEventListener('click', () => {
            btn.closest('.modal').classList.remove('active');
        });
    });
}

// Modal functions
function showLoginModal() {
    document.getElementById('login-modal').classList.add('active');
}

function hideLoginModal() {
    document.getElementById('login-modal').classList.remove('active');
}

function openTokenModal(token = null) {
    const modal = document.getElementById('token-modal');
    const form = document.getElementById('token-form');
    const title = document.getElementById('token-modal-title');
    const idInput = document.getElementById('token-id');
    const brokerInput = document.getElementById('broker-input');
    const tokenInput = document.getElementById('token-input');
    
    // Reset form
    form.reset();
    
    if (token) {
        // Edit mode
        title.textContent = 'Update Broker Token';
        idInput.value = token.id;
        brokerInput.value = token.broker;
        brokerInput.disabled = true;
        tokenInput.value = '';
        tokenInput.placeholder = 'Enter new token value';
    } else {
        // Create mode
        title.textContent = 'Add Broker Token';
        idInput.value = '';
        brokerInput.disabled = false;
        tokenInput.placeholder = 'Enter token value';
    }
    
    modal.classList.add('active');
}

function closeTokenModal() {
    document.getElementById('token-modal').classList.remove('active');
}

// Utility functions
function maskToken(token) {
    if (!token) return '';
    if (token.length <= 8) return '********';
    return token.substring(0, 4) + '...' + token.substring(token.length - 4);
}

function showError(message) {
    alert(message);
}

async function fetchWithAuth(url, options = {}) {
    if (!state.token) {
        throw new Error('Not authenticated');
    }
    
    const headers = options.headers || {};
    headers['Authorization'] = `Bearer ${state.token}`;
    
    const response = await fetch(url, {
        ...options,
        headers
    });
    
    if (response.status === 401) {
        // Token expired or invalid
        logout();
        throw new Error('Session expired. Please login again.');
    }
    
    return response;
}
