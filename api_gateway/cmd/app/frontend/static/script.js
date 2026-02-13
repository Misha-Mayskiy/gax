// ==============================================
// КОНФИГУРАЦИЯ И ГЛОБАЛЬНЫЕ ПЕРЕМЕННЫЕ
// ==============================================

const API_BASE_URL = 'http://localhost:8080';
let currentUser = null;
let currentChat = null;
let chats = [];
let messages = new Map();

// ==============================================
// УТИЛИТЫ ДЛЯ ОТЛАДКИ
// ==============================================

function debugLog(message, data = null) {
    console.log(`[DEBUG] ${message}`, data || '');
}

// ==============================================
// ФУНКЦИИ ДЛЯ РАБОТЫ С LOCALSTORAGE
// ==============================================

function saveMessagesToStorage() {
    try {
        // Преобразуем Map в обычный объект
        const messagesObj = {};
        messages.forEach((value, key) => {
            messagesObj[key] = value;
        });
        
        localStorage.setItem('gax_messages', JSON.stringify(messagesObj));
        debugLog('Сообщения сохранены в localStorage');
    } catch (error) {
        console.error('Error saving messages:', error);
        debugLog('Ошибка сохранения сообщений', error);
    }
}

function loadMessagesFromStorage() {
    try {
        const savedMessages = localStorage.getItem('gax_messages');
        if (savedMessages) {
            const messagesObj = JSON.parse(savedMessages);
            
            // Преобразуем обратно в Map
            messages = new Map();
            Object.keys(messagesObj).forEach(key => {
                messages.set(key, messagesObj[key]);
            });
            
            debugLog(`Загружено ${Object.keys(messagesObj).length} чатов с сообщениями из localStorage`);
            return true;
        }
    } catch (error) {
        console.error('Error loading messages:', error);
        debugLog('Ошибка загрузки сообщений', error);
    }
    return false;
}

function saveChatsToStorage() {
    try {
        localStorage.setItem('gax_chats', JSON.stringify(chats));
        debugLog('Чаты сохранены в localStorage');
    } catch (error) {
        console.error('Error saving chats:', error);
        debugLog('Ошибка сохранения чатов', error);
    }
}

function loadChatsFromStorage() {
    try {
        const savedChats = localStorage.getItem('gax_chats');
        if (savedChats) {
            chats = JSON.parse(savedChats);
            debugLog(`Загружено ${chats.length} чатов из localStorage`);
            return true;
        }
    } catch (error) {
        console.error('Error loading chats:', error);
        debugLog('Ошибка загрузки чатов', error);
    }
    return false;
}

// ==============================================
// КЛАСС ДЛЯ РАБОТЫ С API (ТОЛЬКО МОК)
// ==============================================

class GAXAPI {
    constructor() {
        this.token = localStorage.getItem('gax_token');
        debugLog('API инициализирован');
    }

    setToken(token) {
        this.token = token;
        localStorage.setItem('gax_token', token);
        debugLog('Токен установлен');
    }

    async request(endpoint, options = {}) {
        // Всегда используем мок данные
        debugLog(`Запрос: ${options.method || 'GET'} ${endpoint}`);
        
        // Имитируем задержку сети
        await new Promise(resolve => setTimeout(resolve, 100));
        
        // Возвращаем мок данные в зависимости от эндпоинта
        return this.getMockResponse(endpoint, options);
    }

    getMockResponse(endpoint, options) {
        const path = endpoint.split('?')[0];
        
        switch(path) {
            case '/user/create':
                const userData = options.body ? JSON.parse(options.body) : {};
                return {
                    success: true,
                    user: {
                        id: 'user-' + Date.now(),
                        username: userData.username || 'Пользователь',
                        email: userData.email || 'user@example.com',
                        created_at: new Date().toISOString()
                    },
                    token: 'mock-token-' + Date.now()
                };
                
            case '/chat/create-direct':
                const directData = options.body ? JSON.parse(options.body) : {};
                return {
                    success: true,
                    chat: {
                        id: 'direct-' + Date.now(),
                        name: 'Новый чат',
                        type: 'direct',
                        lastMessage: '',
                        timestamp: new Date().toISOString(),
                        peerId: directData.user2_id
                    }
                };
                
            case '/chat/create-group':
                const groupData = options.body ? JSON.parse(options.body) : {};
                return {
                    success: true,
                    chat: {
                        id: 'group-' + Date.now(),
                        name: groupData.title || 'Новая группа',
                        type: 'group',
                        lastMessage: '',
                        timestamp: new Date().toISOString(),
                        members: groupData.member_ids || []
                    }
                };
                
            case '/chat/list':
                return {
                    success: true,
                    chats: getDemoChats()
                };
                
            case '/chat/messages/list':
                const chatId = new URLSearchParams(endpoint.split('?')[1]).get('chat_id');
                const chatMessages = getStoredMessages(chatId);
                return {
                    success: true,
                    messages: chatMessages
                };
                
            case '/chat/message/send':
                const sendData = options.body ? JSON.parse(options.body) : {};
                return {
                    success: true,
                    message: {
                        id: 'msg-' + Date.now(),
                        chatId: sendData.chat_id,
                        senderId: 'current-user',
                        content: sendData.content,
                        type: 'text',
                        timestamp: new Date().toISOString(),
                        status: 'sent'
                    }
                };
                
            case '/user/set_online':
                return {
                    success: true,
                    message: 'Online status updated'
                };
                
            default:
                return {
                    success: true,
                    message: 'Mock response'
                };
        }
    }

    // API методы
    async createUser(userData) {
        return this.request('/user/create', {
            method: 'PUT',
            body: JSON.stringify(userData)
        });
    }

    async createDirectChat(user1Id, user2Id) {
        return this.request('/chat/create-direct', {
            method: 'POST',
            body: JSON.stringify({ user1_id: user1Id, user2_id: user2Id })
        });
    }

    async createGroupChat(title, memberIds) {
        return this.request('/chat/create-group', {
            method: 'POST',
            body: JSON.stringify({ title, member_ids: memberIds })
        });
    }

    async listChats(userId) {
        return this.request(`/chat/list?user_id=${userId}`);
    }

    async listMessages(chatId) {
        return this.request(`/chat/messages/list?chat_id=${chatId}`);
    }

    async sendMessage(chatId, content) {
        return this.request('/chat/message/send', {
            method: 'POST',
            body: JSON.stringify({ chat_id: chatId, content })
        });
    }

    async setOnline(userId, isOnline = true) {
        return this.request('/user/set_online', {
            method: 'POST',
            body: JSON.stringify({ user_id: userId, is_online: isOnline })
        });
    }
}

// ==============================================
// ВСПОМОГАТЕЛЬНЫЕ ФУНКЦИИ ДЛЯ ДЕМО ДАННЫХ
// ==============================================

function getDemoChats() {
    // Проверяем, есть ли сохраненные чаты
    if (chats.length > 0) {
        return chats;
    }
    
    // Возвращаем демо чаты если нет сохраненных
    return [
        {
            id: 'chat-1',
            name: 'Общий чат',
            type: 'group',
            lastMessage: 'Добро пожаловать в GAX Messenger!',
            timestamp: new Date().toISOString(),
            members: ['user-1', 'user-2', 'user-3'],
            unread: 0
        },
        {
            id: 'chat-2',
            name: 'Мария',
            type: 'direct',
            lastMessage: 'Привет! Как дела?',
            timestamp: new Date(Date.now() - 3600000).toISOString(),
            peerId: 'user-2',
            unread: 2
        },
        {
            id: 'chat-3',
            name: 'Команда проекта',
            type: 'group',
            lastMessage: 'Завтра созвон в 10:00',
            timestamp: new Date(Date.now() - 7200000).toISOString(),
            members: ['user-1', 'user-2', 'user-3'],
            unread: 0
        }
    ];
}

function getStoredMessages(chatId) {
    // Возвращаем сохраненные сообщения или демо
    if (messages.has(chatId)) {
        return messages.get(chatId);
    }
    
    // Демо сообщения для нового чата
    return [
        {
            id: 'msg-1-' + chatId,
            chatId: chatId,
            senderId: 'system',
            content: 'Добро пожаловать в чат!',
            type: 'system',
            timestamp: new Date(Date.now() - 3600000).toISOString(),
            status: 'read'
        },
        {
            id: 'msg-2-' + chatId,
            chatId: chatId,
            senderId: 'demo-peer',
            content: 'Привет! Рад тебя видеть!',
            type: 'text',
            timestamp: new Date(Date.now() - 3500000).toISOString(),
            status: 'read'
        },
        {
            id: 'msg-3-' + chatId,
            chatId: chatId,
            senderId: 'current-user',
            content: 'Привет! Я тоже рад! Как дела?',
            type: 'text',
            timestamp: new Date(Date.now() - 3400000).toISOString(),
            status: 'read'
        }
    ];
}

// ==============================================
// ИНИЦИАЛИЗАЦИЯ
// ==============================================

const api = new GAXAPI();

// ==============================================
// ОСНОВНЫЕ ФУНКЦИИ
// ==============================================

async function initApp() {
    debugLog('Начало инициализации приложения');
    
    const username = document.getElementById('username').value;
    const email = document.getElementById('email').value;
    
    if (!username || !email) {
        showAuthStatus('Заполните все поля', 'error');
        return;
    }

    try {
        showAuthStatus('Создание пользователя...', 'info');
        
        // Создаем пользователя
        const userResult = await api.createUser({
            username: username,
            email: email
        });
        
        if (userResult.success) {
            currentUser = userResult.user;
            api.setToken(userResult.token);
            
            // Сохраняем в localStorage
            localStorage.setItem('gax_user', JSON.stringify(currentUser));
            localStorage.setItem('gax_token', userResult.token);
            
            debugLog('Пользователь создан', currentUser);
            showAuthStatus('Пользователь создан!', 'success');
            
            // Устанавливаем онлайн статус
            await api.setOnline(currentUser.id, true);
            
            // Показываем основной интерфейс
            setTimeout(() => {
                showMainScreen();
                loadInitialData();
            }, 500);
            
        } else {
            showAuthStatus('Ошибка создания пользователя', 'error');
            debugLog('Ошибка создания пользователя');
        }
        
    } catch (error) {
        console.error('Init error:', error);
        debugLog('Ошибка инициализации', error);
        
        // Создаем демо пользователя в случае ошибки
        currentUser = {
            id: 'demo-user-' + Date.now(),
            username: username,
            email: email,
            created_at: new Date().toISOString()
        };
        
        localStorage.setItem('gax_user', JSON.stringify(currentUser));
        localStorage.setItem('gax_token', 'demo-token-' + Date.now());
        
        showAuthStatus('Используется демо-режим', 'warning');
        debugLog('Переключение в демо-режим');
        
        setTimeout(() => {
            showMainScreen();
            loadInitialData();
        }, 500);
    }
}

function showAuthScreen() {
    debugLog('Показ экрана авторизации');
    const authScreen = document.getElementById('authScreen');
    const mainScreen = document.getElementById('mainScreen');
    
    if (authScreen) authScreen.style.display = 'flex';
    if (mainScreen) {
        mainScreen.style.display = 'none';
        mainScreen.classList.remove('show');
    }
}

function showMainScreen() {
    debugLog('Показ основного интерфейса');
    const authScreen = document.getElementById('authScreen');
    const mainScreen = document.getElementById('mainScreen');
    
    if (authScreen) authScreen.style.display = 'none';
    if (mainScreen) {
        mainScreen.style.display = 'flex';
        
        // Анимация появления
        setTimeout(() => {
            mainScreen.classList.add('show');
        }, 50);
    }
    
    if (currentUser) {
        document.getElementById('currentUserName').textContent = currentUser.username;
        document.getElementById('currentUserStatus').textContent = '🟢 Онлайн';
    }
}

function showAuthStatus(message, type) {
    debugLog(`Статус авторизации: ${message}`);
    const element = document.getElementById('authStatus');
    if (element) {
        element.textContent = message;
        element.style.color = type === 'error' ? '#dc3545' : 
                            type === 'success' ? '#28a745' : 
                            type === 'warning' ? '#ffc107' : '#007bff';
    }
}

async function loadInitialData() {
    debugLog('Загрузка начальных данных');
    
    if (!currentUser) {
        debugLog('Нет пользователя, загрузка невозможна');
        return;
    }
    
    // Загружаем данные из localStorage
    const loadedChats = loadChatsFromStorage();
    const loadedMessages = loadMessagesFromStorage();
    
    if (!loadedChats) {
        // Загружаем чаты из API, если нет в localStorage
        await loadChats();
    } else {
        renderChatList();
    }
    
    if (loadedMessages && currentChat) {
        // Если есть загруженные сообщения и выбран чат, показываем их
        const chatMessages = messages.get(currentChat.id);
        if (chatMessages) {
            renderMessages(chatMessages);
        }
    }
    
    // Автоматически выбираем первый чат
    if (chats.length > 0 && !currentChat) {
        setTimeout(() => {
            debugLog('Автовыбор первого чата');
            selectChat(chats[0]);
        }, 200);
    }
}

async function loadChats() {
    debugLog('Загрузка списка чатов');
    
    if (!currentUser) {
        debugLog('Нет пользователя для загрузки чатов');
        return;
    }
    
    try {
        const result = await api.listChats(currentUser.id);
        
        if (result.success && result.chats) {
            chats = result.chats;
            // Сохраняем чаты
            saveChatsToStorage();
            debugLog(`Загружено ${chats.length} чатов`, chats);
        } else {
            // Используем демо чаты
            chats = getDemoChats();
            debugLog('Используются демо чаты', chats);
        }
        
        renderChatList();
        
    } catch (error) {
        console.error('Error loading chats:', error);
        debugLog('Ошибка загрузки чатов', error);
        chats = getDemoChats();
        renderChatList();
    }
}

async function createDirectChat() {
    debugLog('Создание директ-чата');
    
    if (!currentUser) {
        alert('Сначала войдите в систему');
        return;
    }
    
    const peerName = prompt('Имя собеседника:', 'Тестовый собеседник');
    if (!peerName) return;
    
    try {
        // Создаем ID для собеседника
        const peerId = 'peer-' + Date.now();
        
        const result = await api.createDirectChat(currentUser.id, peerId);
        
        if (result.success && result.chat) {
            // Обновляем имя чата
            result.chat.name = peerName;
            result.chat.peerId = peerId;
            
            // Добавляем в начало списка
            chats.unshift(result.chat);
            
            // Сохраняем чаты
            saveChatsToStorage();
            
            renderChatList();
            
            debugLog('Директ-чат создан', result.chat);
            
            // Выбираем новый чат
            selectChat(result.chat);
            
            // Добавляем приветственное сообщение
            addMessageToChat(result.chat.id, {
                id: 'welcome-' + Date.now(),
                senderId: 'system',
                content: `Чат с "${peerName}" создан! Начните общение.`,
                type: 'system',
                timestamp: new Date().toISOString(),
                status: 'read'
            });
        }
    } catch (error) {
        console.error('Error creating chat:', error);
        debugLog('Ошибка создания чата', error);
        alert('Ошибка создания чата');
    }
}

async function createGroupChat() {
    debugLog('Создание группового чата');
    
    if (!currentUser) {
        alert('Сначала войдите в систему');
        return;
    }
    
    const groupName = prompt('Название группы:', 'Новая группа');
    if (!groupName) return;
    
    try {
        // Создаем демо участников
        const members = [
            currentUser.id,
            'member-1-' + Date.now(),
            'member-2-' + Date.now(),
            'member-3-' + Date.now()
        ];
        
        const result = await api.createGroupChat(groupName, members);
        
        if (result.success && result.chat) {
            // Добавляем в начало списка
            chats.unshift(result.chat);
            
            // Сохраняем чаты
            saveChatsToStorage();
            
            renderChatList();
            
            debugLog('Групповой чат создан', result.chat);
            
            // Выбираем новый чат
            selectChat(result.chat);
            
            // Добавляем приветственное сообщение
            addMessageToChat(result.chat.id, {
                id: 'welcome-group-' + Date.now(),
                senderId: 'system',
                content: `Группа "${groupName}" создана! Добро пожаловать!`,
                type: 'system',
                timestamp: new Date().toISOString(),
                status: 'read'
            });
        }
    } catch (error) {
        console.error('Error creating group chat:', error);
        debugLog('Ошибка создания группы', error);
        alert('Ошибка создания группы');
    }
}

async function sendMessage() {
    debugLog('Отправка сообщения');
    
    const input = document.getElementById('messageInput');
    const text = input.value.trim();
    
    if (!text || !currentChat) {
        if (!currentChat) alert('Выберите чат для отправки сообщения');
        input.focus();
        return;
    }
    
    try {
        // Создаем временное сообщение
        const tempMessage = {
            id: 'temp-' + Date.now(),
            chatId: currentChat.id,
            senderId: currentUser.id,
            content: text,
            timestamp: new Date().toISOString(),
            status: 'sending'
        };
        
        // Добавляем в UI
        addMessageToChat(currentChat.id, tempMessage);
        input.value = '';
        
        // Отправляем через API
        const result = await api.sendMessage(currentChat.id, text);
        
        if (result.success && result.message) {
            // Обновляем статус
            updateMessageStatus(tempMessage.id, {
                id: result.message.id,
                status: 'sent'
            });
            
            // Обновляем последнее сообщение в чате
            updateChatLastMessage(currentChat.id, text);
            
            debugLog('Сообщение отправлено', result.message);
            
        } else {
            updateMessageStatus(tempMessage.id, { status: 'error' });
            debugLog('Ошибка отправки сообщения');
        }
    } catch (error) {
        console.error('Error sending message:', error);
        updateMessageStatus(tempMessage.id, { status: 'error' });
        debugLog('Ошибка отправки сообщения', error);
    }
}

async function loadMessages(chatId) {
    debugLog(`Загрузка сообщений для чата: ${chatId}`);
    
    if (!chatId) return;
    
    try {
        const result = await api.listMessages(chatId);
        
        if (result.success && result.messages) {
            messages.set(chatId, result.messages);
            
            // Сохраняем сообщения
            saveMessagesToStorage();
            
            renderMessages(result.messages);
            debugLog(`Загружено ${result.messages.length} сообщений`);
        } else {
            // Используем сохраненные сообщения
            const storedMessages = messages.get(chatId);
            if (storedMessages) {
                renderMessages(storedMessages);
            } else {
                // Используем демо сообщения
                const demoMessages = getStoredMessages(chatId);
                messages.set(chatId, demoMessages);
                saveMessagesToStorage();
                renderMessages(demoMessages);
            }
        }
    } catch (error) {
        console.error('Error loading messages:', error);
        debugLog('Ошибка загрузки сообщений', error);
        
        // Используем сохраненные сообщения
        const storedMessages = messages.get(chatId);
        if (storedMessages) {
            renderMessages(storedMessages);
        } else {
            // Используем демо сообщения
            const demoMessages = getStoredMessages(chatId);
            messages.set(chatId, demoMessages);
            saveMessagesToStorage();
            renderMessages(demoMessages);
        }
    }
}

// ==============================================
// ФУНКЦИИ РЕНДЕРИНГА
// ==============================================

function renderChatList() {
    debugLog('Рендеринг списка чатов');
    
    const container = document.getElementById('chatList');
    if (!container) {
        debugLog('Контейнер чатов не найден!');
        return;
    }
    
    container.innerHTML = '';
    
    if (chats.length === 0) {
        container.innerHTML = `
            <div style="padding: 20px; text-align: center; color: #6c757d;">
                Нет чатов. Создайте первый чат!
            </div>
        `;
        debugLog('Нет чатов для отображения');
        return;
    }
    
    chats.forEach(chat => {
        const element = document.createElement('div');
        element.className = `chat-item ${currentChat?.id === chat.id ? 'active' : ''}`;
        element.onclick = () => {
            debugLog(`Выбран чат: ${chat.name}`);
            selectChat(chat);
        };
        
        const time = chat.timestamp ? 
            new Date(chat.timestamp).toLocaleTimeString([], { hour: '2-digit', minute: '2-digit' }) : '';
        
        const icon = chat.type === 'group' ? '👥' : '👤';
        
        element.innerHTML = `
            <div style="display: flex; align-items: center; gap: 12px;">
                <div style="font-size: 24px;">${icon}</div>
                <div style="flex: 1; min-width: 0;">
                    <div style="font-weight: 600; margin-bottom: 4px;">${chat.name}</div>
                    <div style="font-size: 14px; opacity: 0.8; white-space: nowrap; overflow: hidden; text-overflow: ellipsis;">
                        ${chat.lastMessage || 'Нет сообщений'}
                    </div>
                </div>
                <div style="font-size: 12px; color: ${currentChat?.id === chat.id ? 'rgba(255,255,255,0.8)' : '#6c757d'}">
                    ${time}
                </div>
            </div>
            ${chat.unread > 0 ? `
                <div style="position: absolute; top: 12px; right: 12px; background: #dc3545; color: white; border-radius: 50%; width: 20px; height: 20px; display: flex; align-items: center; justify-content: center; font-size: 12px; font-weight: bold;">
                    ${chat.unread}
                </div>
            ` : ''}
        `;
        
        container.appendChild(element);
    });
    
    debugLog(`Отрендерено ${chats.length} чатов`);
}

function renderMessages(messagesArray) {
    debugLog(`Рендеринг сообщений: ${messagesArray?.length || 0} шт.`);
    
    const container = document.getElementById('messagesContainer');
    if (!container) {
        debugLog('Контейнер сообщений не найден!');
        return;
    }
    
    container.innerHTML = '';
    
    if (!messagesArray || messagesArray.length === 0) {
        container.innerHTML = `
            <div style="text-align: center; padding: 40px; color: #6c757d;">
                Нет сообщений. Начните общение!
            </div>
        `;
        debugLog('Нет сообщений для отображения');
        return;
    }
    
    messagesArray.forEach(msg => {
        const isSent = msg.senderId === currentUser?.id;
        const isSystem = msg.type === 'system' || msg.senderId === 'system';
        const time = msg.timestamp ? 
            new Date(msg.timestamp).toLocaleTimeString([], { hour: '2-digit', minute: '2-digit' }) : '';
        
        const element = document.createElement('div');
        element.className = `message ${isSystem ? 'system' : isSent ? 'sent' : 'received'}`;
        
        element.innerHTML = `
            <div class="message-content">
                <div>${msg.content || ''}</div>
                <div class="message-time">${time}</div>
                ${msg.status === 'error' ? 
                    '<div style="color: #dc3545; font-size: 12px; margin-top: 4px;">Ошибка отправки</div>' : 
                    msg.status === 'sending' ?
                    '<div style="color: #ffc107; font-size: 12px; margin-top: 4px;">Отправка...</div>' : ''
                }
            </div>
        `;
        
        container.appendChild(element);
    });
    
    // Прокручиваем вниз
    setTimeout(() => {
        container.scrollTop = container.scrollHeight;
    }, 100);
    
    debugLog('Сообщения отрендерены');
}

function selectChat(chat) {
    debugLog(`Выбор чата: ${chat.name} (${chat.id})`);
    
    currentChat = chat;
    renderChatList();
    document.getElementById('chatTitle').textContent = chat.name;
    loadMessages(chat.id);
    
    // Фокусируемся на поле ввода
    setTimeout(() => {
        const input = document.getElementById('messageInput');
        if (input) input.focus();
    }, 100);
}

function addMessageToChat(chatId, message) {
    debugLog(`Добавление сообщения в чат ${chatId}`);
    
    if (!messages.has(chatId)) {
        messages.set(chatId, []);
    }
    
    const chatMessages = messages.get(chatId);
    chatMessages.push(message);
    
    // Сохраняем сообщения
    saveMessagesToStorage();
    
    if (currentChat && currentChat.id === chatId) {
        renderMessages(chatMessages);
    }
}

function updateMessageStatus(tempId, updates) {
    debugLog(`Обновление статуса сообщения ${tempId}`, updates);
    
    if (!currentChat) return;
    
    const chatMessages = messages.get(currentChat.id);
    if (chatMessages) {
        const index = chatMessages.findIndex(m => m.id === tempId);
        if (index !== -1) {
            chatMessages[index] = { ...chatMessages[index], ...updates };
            
            // Сохраняем обновленные сообщения
            saveMessagesToStorage();
            
            renderMessages(chatMessages);
        }
    }
}

function updateChatLastMessage(chatId, lastMessage) {
    debugLog(`Обновление последнего сообщения чата ${chatId}: ${lastMessage}`);
    
    const chatIndex = chats.findIndex(c => c.id === chatId);
    if (chatIndex !== -1) {
        chats[chatIndex].lastMessage = lastMessage;
        chats[chatIndex].timestamp = new Date().toISOString();
        
        // Сохраняем чаты
        saveChatsToStorage();
        
        renderChatList();
    }
}

function clearData() {
    if (confirm('Очистить все данные и выйти?')) {
        debugLog('Очистка данных');
        
        localStorage.clear();
        currentUser = null;
        currentChat = null;
        chats = [];
        messages.clear();
        
        showAuthScreen();
        
        // Обновляем поля авторизации
        const randomNum = Math.floor(Math.random() * 1000);
        document.getElementById('username').value = `Пользователь ${randomNum}`;
        document.getElementById('email').value = `user${randomNum}@example.com`;
        
        debugLog('Данные очищены, показан экран авторизации');
    }
}

function checkExistingSession() {
    debugLog('Проверка существующей сессии');
    
    const userData = localStorage.getItem('gax_user');
    
    if (userData) {
        try {
            currentUser = JSON.parse(userData);
            debugLog('Сессия найдена', currentUser);
            
            // Проверяем, есть ли токен
            const token = localStorage.getItem('gax_token');
            if (token) {
                api.setToken(token);
            }
            
            // Загружаем чаты и сообщения из localStorage
            loadChatsFromStorage();
            loadMessagesFromStorage();
            
            showMainScreen();
            loadInitialData();
            return true;
            
        } catch (e) {
            console.error('Error restoring session:', e);
            debugLog('Ошибка восстановления сессии', e);
            localStorage.clear();
            return false;
        }
    }
    
    debugLog('Сессия не найдена');
    return false;
}

// ==============================================
// ИНИЦИАЛИЗАЦИЯ ПРИ ЗАГРУЗКЕ
// ==============================================

document.addEventListener('DOMContentLoaded', function() {
    debugLog('Документ загружен');
    
    // Проверяем существующую сессию
    const hasSession = checkExistingSession();
    
    if (!hasSession) {
        showAuthScreen();
        
        // Автозаполняем демо данные
        const randomNum = Math.floor(Math.random() * 1000);
        document.getElementById('username').value = `Пользователь ${randomNum}`;
        document.getElementById('email').value = `user${randomNum}@example.com`;
        
        debugLog('Автозаполнение демо данных');
    }
});

// ==============================================
// ЭКСПОРТ ФУНКЦИЙ
// ==============================================

window.initApp = initApp;
window.createDirectChat = createDirectChat;
window.createGroupChat = createGroupChat;
window.sendMessage = sendMessage;
window.clearData = clearData;
window.debugLog = debugLog;