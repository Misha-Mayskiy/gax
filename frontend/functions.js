// functions.js

// ========== НАВИГАЦИЯ И ОБЩИЕ ФУНКЦИИ ==========
window.showSection = function(sectionId) {
    // Скрыть все секции
    document.querySelectorAll('.content-section').forEach(section => {
        section.classList.remove('active');
    });
    
    // Показать нужную секцию
    const targetSection = document.getElementById(`${sectionId}Section`);
    if (targetSection) {
        targetSection.classList.add('active');
    }
    
    // Обновить активный пункт меню
    document.querySelectorAll('.menu-item').forEach(item => {
        item.classList.remove('active');
    });
    
    // Найти и активировать соответствующий пункт меню
    const menuItem = Array.from(document.querySelectorAll('.menu-item')).find(item => 
        item.onclick && item.onclick.toString().includes(sectionId)
    );
    
    if (menuItem) {
        menuItem.classList.add('active');
    }
};

window.toggleOnlineStatus = function() {
    if (!currentUser) {
        showAuthModal();
        return;
    }
    
    const statusText = document.getElementById('userStatus').textContent;
    if (statusText.includes('онлайн')) {
        setOfflineStatus();
    } else {
        setOnlineStatus();
    }
};

// ========== ФУНКЦИИ ПОЛЬЗОВАТЕЛЕЙ ==========
window.getUserInfo = async function() {
    const uuid = document.getElementById('getUserUUID')?.value;
    
    if (!uuid) {
        alert('Введите UUID пользователя');
        return;
    }
    
    // Показываем индикатор загрузки
    const button = event?.target || document.getElementById('findUserBtn');
    if (button) {
        const originalText = button.textContent;
        button.innerHTML = '<span class="loading"></span> Поиск...';
        button.disabled = true;
    }
    
    try {
        const controller = new AbortController();
        const timeoutId = setTimeout(() => controller.abort(), 10000);
        
        const response = await fetch(`${API_BASE_URL}/user/get?uuid=${uuid}`, {
            signal: controller.signal,
            headers: currentToken ? { 'Authorization': `Bearer ${currentToken}` } : {}
        });
        
        clearTimeout(timeoutId);
        
        if (!response.ok) {
            if (response.status === 404) {
                throw new Error('Пользователь не найден');
            } else if (response.status === 429) {
                throw new Error('Слишком много запросов. Подождите немного.');
            } else {
                throw new Error(`HTTP ${response.status}`);
            }
        }
        
        const data = await response.json();
        
        const resultDiv = document.getElementById('userInfoResult');
        if (resultDiv) {
            resultDiv.innerHTML = `<pre>${JSON.stringify(data, null, 2)}</pre>`;
        }
    } catch (error) {
        console.error('Error getting user info:', error);
        
        let errorMessage = error.message;
        if (error.name === 'AbortError') {
            errorMessage = 'Запрос занял слишком много времени';
        }
        
        alert(errorMessage);
    } finally {
        // Восстанавливаем кнопку
        if (button) {
            button.textContent = originalText;
            button.disabled = false;
        }
    }
};

// ========== АУТЕНТИФИКАЦИЯ ==========
window.showAuthModal = function() {
    const authModal = document.getElementById('authModal');
    if (authModal) {
        authModal.style.display = 'block';
    }
    switchTab('login');
};

window.switchTab = function(tabName) {
    const loginForm = document.getElementById('loginForm');
    const registerForm = document.getElementById('registerForm');
    const passwordChangeForm = document.getElementById('passwordChangeForm');
    
    if (!loginForm || !registerForm) return;
    
    // Скрыть все табы
    [loginForm, registerForm, passwordChangeForm].forEach(form => {
        if (form) form.classList.remove('active');
    });
    
    // Убрать активный класс со всех кнопок табов
    document.querySelectorAll('.tab').forEach(tab => {
        tab.classList.remove('active');
    });
    
    // Показать нужный таб
    if (tabName === 'login') {
        loginForm.classList.add('active');
        const loginTab = document.querySelector('.tab:nth-child(1)');
        if (loginTab) loginTab.classList.add('active');
    } else if (tabName === 'register') {
        registerForm.classList.add('active');
        const registerTab = document.querySelector('.tab:nth-child(2)');
        if (registerTab) registerTab.classList.add('active');
    }
};

// ========== ЗАГРУЗКА ИНФОРМАЦИИ О ПОЛЬЗОВАТЕЛЕ ==========
window.loadUserInfo = async function() {
    if (!currentUser) return;
    
    try {
        const response = await fetch(`${API_BASE_URL}/user/get?uuid=${currentUser.uuid}`, {
            headers: currentToken ? { 'Authorization': `Bearer ${currentToken}` } : {}
        });
        
        if (response.ok) {
            const data = await response.json();
            
            const userInfoDiv = document.getElementById('currentUserInfo');
            if (userInfoDiv) {
                userInfoDiv.innerHTML = `
                    <div class="user-detail">
                        <strong>UUID:</strong> ${data.uuid || 'Не указан'}
                    </div>
                    <div class="user-detail">
                        <strong>Email:</strong> ${data.email || 'Не указан'}
                    </div>
                    <div class="user-detail">
                        <strong>Имя:</strong> ${data.user_name || 'Не указано'}
                    </div>
                    <div class="user-detail">
                        <strong>О себе:</strong> ${data.about_me || 'Не указано'}
                    </div>
                    <div class="user-detail">
                        <strong>Статус:</strong> ${data.status || 'Не указан'}
                    </div>
                `;
            }
        }
    } catch (error) {
        console.error('Error loading user info:', error);
    }
};

window.loadOnlineUsers = async function() {
    if (!currentUser) return;
    
    try {
        const response = await fetch(`${API_BASE_URL}/user/get_online_users`, {
            headers: currentToken ? { 'Authorization': `Bearer ${currentToken}` } : {}
        });
        
        if (response.ok) {
            const data = await response.json();
            
            const listElement = document.getElementById('onlineUsersList');
            if (listElement) {
                listElement.innerHTML = '';
                
                if (data.uuids && data.uuids.length > 0) {
                    data.uuids.forEach(uuid => {
                        const li = document.createElement('li');
                        li.innerHTML = `<i class="fas fa-user-circle"></i> ${uuid}`;
                        listElement.appendChild(li);
                    });
                } else {
                    listElement.innerHTML = '<li>Нет пользователей онлайн</li>';
                }
            }
        }
    } catch (error) {
        console.error('Error loading online users:', error);
    }
};

// ========== ОБНОВЛЕНИЕ UI ==========
window.updateUIAfterLogin = function() {
    if (!currentUser) return;
    
    const userEmailEl = document.getElementById('userEmail');
    const footerUserEl = document.getElementById('footerUser');
    const userNameEl = document.getElementById('userName');
    const userAvatarEl = document.getElementById('userAvatar');
    
    if (userEmailEl) userEmailEl.textContent = currentUser.email || 'Пользователь';
    if (footerUserEl) footerUserEl.textContent = currentUser.email || 'Пользователь';
    if (userNameEl) userNameEl.textContent = currentUser.username || currentUser.email?.split('@')[0] || 'Пользователь';
    if (userAvatarEl) userAvatarEl.textContent = (currentUser.email?.[0] || 'U').toUpperCase();
    
    const logoutBtn = document.getElementById('logoutBtn');
    const loginBtn = document.getElementById('loginBtn');
    
    if (logoutBtn) logoutBtn.style.display = 'inline-block';
    if (loginBtn) loginBtn.style.display = 'none';
    
    // Загружаем информацию о пользователе
    loadUserInfo();
    loadOnlineUsers();
};

window.updateUIAfterLogout = function() {
    const userEmailEl = document.getElementById('userEmail');
    const footerUserEl = document.getElementById('footerUser');
    const userNameEl = document.getElementById('userName');
    const userAvatarEl = document.getElementById('userAvatar');
    const userStatusEl = document.getElementById('userStatus');
    
    if (userEmailEl) userEmailEl.textContent = 'Гость';
    if (footerUserEl) footerUserEl.textContent = 'Гость';
    if (userNameEl) userNameEl.textContent = 'Гость';
    if (userAvatarEl) userAvatarEl.textContent = 'U';
    if (userStatusEl) userStatusEl.textContent = 'Статус: не в сети';
    
    const logoutBtn = document.getElementById('logoutBtn');
    const loginBtn = document.getElementById('loginBtn');
    
    if (logoutBtn) logoutBtn.style.display = 'none';
    if (loginBtn) loginBtn.style.display = 'inline-block';
    
    const currentUserInfoEl = document.getElementById('currentUserInfo');
    const onlineUsersListEl = document.getElementById('onlineUsersList');
    
    if (currentUserInfoEl) currentUserInfoEl.innerHTML = '';
    if (onlineUsersListEl) onlineUsersListEl.innerHTML = '';
};

// ========== ОСТАЛЬНЫЕ ФУНКЦИИ ==========
window.logout = function() {
    currentUser = null;
    currentToken = null;
    
    // Удаляем из localStorage
    localStorage.removeItem('gax_user');
    localStorage.removeItem('gax_token');
    
    // Удаляем куки
    document.cookie = 'jwtToken=; path=/; expires=Thu, 01 Jan 1970 00:00:00 GMT';
    document.cookie = 'jwtToken=; path=/; max-age=0;';
    
    updateUIAfterLogout();
    showNotification('Вы вышли из системы', 'info');
};
// Добавьте в functions.js
window.syncUserWithUserService = async function(userUuid, email, username) {
    try {
        // Проверяем, существует ли пользователь в user-service
        const checkResponse = await fetch(`${API_BASE_URL}/user/get?uuid=${userUuid}`, {
            signal: AbortSignal.timeout(5000)
        });
        
        if (checkResponse.ok) {
            // Пользователь уже существует
            return true;
        } else if (checkResponse.status === 404) {
            // Создаем пользователя
            const createResponse = await fetch(`${API_BASE_URL}/user/create`, {
                method: 'PUT',
                headers: { 'Content-Type': 'application/json' },
                body: JSON.stringify({
                    uuid: userUuid,
                    email: email,
                    user_name: username || email.split('@')[0],
                    avatar: '',
                    about_me: '',
                    status: 'offline',
                    friends: []
                }),
                signal: AbortSignal.timeout(5000)
            });
            
            return createResponse.ok;
        }
    } catch (error) {
        console.error('Error syncing user:', error);
        return false;
    }
    
    return false;
};

// Обновите updateUIAfterLogin:
window.updateUIAfterLogin = function() {
    if (!currentUser) return;
    
    // Синхронизируем пользователя с user-service
    if (currentUser.uuid && currentUser.email) {
        syncUserWithUserService(currentUser.uuid, currentUser.email, currentUser.username)
            .then(success => {
                if (!success) {
                    console.warn('Failed to sync user with user-service');
                }
            });
    }
    
    // Остальной код обновления UI...
    const userEmailEl = document.getElementById('userEmail');
    const footerUserEl = document.getElementById('footerUser');
    const userNameEl = document.getElementById('userName');
    const userAvatarEl = document.getElementById('userAvatar');
    
    if (userEmailEl) userEmailEl.textContent = currentUser.email || 'Пользователь';
    if (footerUserEl) footerUserEl.textContent = currentUser.email || 'Пользователь';
    if (userNameEl) userNameEl.textContent = currentUser.username || currentUser.email?.split('@')[0] || 'Пользователь';
    if (userAvatarEl) userAvatarEl.textContent = (currentUser.email?.[0] || 'U').toUpperCase();
    
    const logoutBtn = document.getElementById('logoutBtn');
    const loginBtn = document.getElementById('loginBtn');
    
    if (logoutBtn) logoutBtn.style.display = 'inline-block';
    if (loginBtn) loginBtn.style.display = 'none';
    
    // Загружаем информацию о пользователе
    if (typeof loadUserInfo === 'function') {
        loadUserInfo();
    }
    if (typeof loadOnlineUsers === 'function') {
        loadOnlineUsers();
    }
};
// Простые заглушки для остальных функций
window.setOnlineStatus = function() {
    alert('Функция в разработке');
};

window.setOfflineStatus = function() {
    alert('Функция в разработке');
};

window.checkOnlineStatus = function() {
    alert('Функция в разработке');
};

window.updateUser = function() {
    alert('Функция в разработке');
};

window.deleteUser = function() {
    alert('Функция в разработке');
};

window.sendMessage = function() {
    alert('Функция в разработке');
};

window.performSearch = function() {
    alert('Функция в разработке');
};

window.createDirectChat = function() {
    alert('Функция в разработке');
};

window.createGroupChat = function() {
    alert('Функция в разработке');
};

window.listMessages = function() {
    alert('Функция в разработке');
};

window.createRoom = function() {
    alert('Функция в разработке');
};

window.joinRoom = function() {
    alert('Функция в разработке');
};

window.setPlayback = function() {
    alert('Функция в разработке');
};

window.getRoomState = function() {
    alert('Функция в разработке');
};

window.connectWebSocket = function() {
    alert('Функция в разработке');
};

window.downloadFile = function() {
    alert('Функция в разработке');
};

window.getFileMeta = function() {
    alert('Функция в разработке');
};

window.listUserFiles = function() {
    alert('Функция в разработке');
};