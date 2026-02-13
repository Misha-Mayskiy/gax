// api-functions.js

const API_BASE_URL = 'http://localhost:8080';

// ========== УТИЛИТЫ ==========
function parseServerResponse(text) {
    if (!text || text.trim() === '') {
        return { success: false, message: 'Пустой ответ сервера' };
    }
    
    try {
        return JSON.parse(text);
    } catch (e) {
        console.log('Raw response:', text);
        
        let message = text;
        if (text.includes('rpc error:')) {
            message = text.replace('rpc error: code = Unknown desc = ', '');
            
            if (message.includes('name taken')) {
                message = 'Имя пользователя уже занято';
            } else if (message.includes('email taken')) {
                message = 'Email уже используется';
            } else if (message.includes('user not found')) {
                message = 'Пользователь не найден';
            }
        }
        
        return { success: false, message: message, raw: text };
    }
}

async function makeRequest(url, options = {}) {
    try {
        const response = await fetch(url, {
            ...options,
            headers: {
                'Content-Type': 'application/json',
                'Authorization': currentToken ? `Bearer ${currentToken}` : '',
                ...options.headers
            }
        });
        
        const text = await response.text();
        const data = parseServerResponse(text);
        
        return {
            status: response.status,
            ok: response.ok,
            data: data
        };
    } catch (error) {
        console.error('Request error:', error);
        return {
            status: 0,
            ok: false,
            error: error.message
        };
    }
}

// ========== ПОЛЬЗОВАТЕЛИ ==========
window.getUserInfo = async function() {
    const uuid = document.getElementById('getUserUUID')?.value;
    if (!uuid) {
        alert('Введите UUID пользователя');
        return;
    }
    
    const result = await makeRequest(`${API_BASE_URL}/user/get?uuid=${uuid}`);
    const resultDiv = document.getElementById('userInfoResult');
    
    if (resultDiv) {
        if (result.ok) {
            resultDiv.innerHTML = `<pre>${JSON.stringify(result.data, null, 2)}</pre>`;
        } else {
            resultDiv.innerHTML = `<div class="error">Ошибка: ${result.data?.message || result.error}</div>`;
        }
    }
};

window.updateUser = async function() {
    const uuid = document.getElementById('updateUUID')?.value;
    const email = document.getElementById('updateEmail')?.value;
    const username = document.getElementById('updateUsername')?.value;
    const about = document.getElementById('updateAbout')?.value;
    
    if (!uuid) {
        alert('Введите UUID');
        return;
    }
    
    const body = {};
    if (email) body.email = email;
    if (username) body.user_name = username;
    if (about) body.about_me = about;
    
    const result = await makeRequest(`${API_BASE_URL}/user/update`, {
        method: 'PATCH',
        body: JSON.stringify(body)
    });
    
    if (result.ok) {
        alert('Профиль обновлен');
    } else {
        alert(`Ошибка: ${result.data?.message || result.error}`);
    }
};

window.deleteUser = async function() {
    if (!confirm('Удалить аккаунт? Это действие нельзя отменить.')) {
        return;
    }
    
    const uuid = document.getElementById('updateUUID')?.value || currentUser?.uuid;
    if (!uuid) {
        alert('Введите UUID');
        return;
    }
    
    const result = await makeRequest(`${API_BASE_URL}/user/delete`, {
        method: 'DELETE',
        body: JSON.stringify({ uuid: uuid })
    });
    
    if (result.ok) {
        alert('Аккаунт удален');
        if (typeof logout === 'function') {
            logout();
        }
    } else {
        alert(`Ошибка: ${result.data?.message || result.error}`);
    }
};

// ========== ЧАТЫ ==========
window.createDirectChat = async function() {
    const userId = document.getElementById('directChatUser')?.value;
    const peerId = document.getElementById('directChatPeer')?.value;
    
    if (!userId || !peerId) {
        alert('Заполните оба ID');
        return;
    }
    
    const result = await makeRequest(`${API_BASE_URL}/chat/create-direct`, {
        method: 'POST',
        body: JSON.stringify({ user_id: userId, peer_id: peerId })
    });
    
    if (result.ok) {
        alert(`Чат создан: ${result.data?.chat?.id || 'ID не получен'}`);
    } else {
        alert(`Ошибка: ${result.data?.message || result.error}`);
    }
};

window.createGroupChat = async function() {
    const userId = document.getElementById('groupChatUser')?.value;
    const title = document.getElementById('groupChatTitle')?.value;
    const members = document.getElementById('groupChatMembers')?.value;
    
    if (!userId || !title || !members) {
        alert('Заполните все поля');
        return;
    }
    
    const memberIds = members.split(',').map(m => m.trim()).filter(m => m);
    
    const result = await makeRequest(`${API_BASE_URL}/chat/create-group`, {
        method: 'POST',
        body: JSON.stringify({
            user_id: userId,
            member_ids: memberIds,
            title: title
        })
    });
    
    if (result.ok) {
        alert(`Групповой чат создан: ${result.data?.chat?.id || 'ID не получен'}`);
    } else {
        alert(`Ошибка: ${result.data?.message || result.error}`);
    }
};

window.sendMessage = async function() {
    const chatId = document.getElementById('messageChatId')?.value;
    const authorId = document.getElementById('messageAuthor')?.value;
    const text = document.getElementById('messageText')?.value;
    
    if (!chatId || !authorId || !text) {
        alert('Заполните все поля');
        return;
    }
    
    const result = await makeRequest(`${API_BASE_URL}/chat/message/send`, {
        method: 'POST',
        body: JSON.stringify({
            chat_id: chatId,
            author_id: authorId,
            text: text
        })
    });
    
    if (result.ok) {
        alert('Сообщение отправлено');
        document.getElementById('messageText').value = '';
    } else {
        alert(`Ошибка: ${result.data?.message || result.error}`);
    }
};

window.listMessages = async function() {
    const chatId = document.getElementById('listMessagesChat')?.value;
    const limit = document.getElementById('messagesLimit')?.value || 10;
    
    if (!chatId) {
        alert('Введите ID чата');
        return;
    }
    
    const result = await makeRequest(`${API_BASE_URL}/chat/messages/list?chat_id=${chatId}&limit=${limit}`);
    const resultDiv = document.getElementById('messagesResult');
    
    if (resultDiv) {
        if (result.ok && result.data?.messages) {
            let html = `<h4>Сообщения (${result.data.messages.length}):</h4>`;
            result.data.messages.forEach(msg => {
                html += `
                    <div class="message">
                        <strong>${msg.author_id}:</strong> ${msg.text}
                        <br><small>${new Date(msg.created_at).toLocaleString()}</small>
                    </div>
                    <hr>
                `;
            });
            resultDiv.innerHTML = html;
        } else {
            resultDiv.innerHTML = `<div class="error">Ошибка: ${result.data?.message || result.error || 'Сообщений нет'}</div>`;
        }
    }
};

// ========== КОМНАТЫ ==========
window.createRoom = async function() {
    const hostId = document.getElementById('roomHostId')?.value;
    const trackUrl = document.getElementById('roomTrackUrl')?.value;
    
    if (!hostId || !trackUrl) {
        alert('Заполните все поля');
        return;
    }
    
    const result = await makeRequest(`${API_BASE_URL}/room/create`, {
        method: 'POST',
        body: JSON.stringify({
            host_id: hostId,
            track_url: trackUrl
        })
    });
    
    const resultDiv = document.getElementById('roomResult');
    if (resultDiv) {
        if (result.ok) {
            resultDiv.innerHTML = `<pre>${JSON.stringify(result.data, null, 2)}</pre>`;
        } else {
            resultDiv.innerHTML = `<div class="error">Ошибка: ${result.data?.message || result.error}</div>`;
        }
    }
};

window.joinRoom = async function() {
    const roomId = document.getElementById('joinRoomId')?.value;
    const userId = document.getElementById('joinUserId')?.value;
    
    if (!roomId || !userId) {
        alert('Заполните все поля');
        return;
    }
    
    const result = await makeRequest(`${API_BASE_URL}/room/join`, {
        method: 'POST',
        body: JSON.stringify({
            room_id: roomId,
            user_id: userId
        })
    });
    
    if (result.ok) {
        alert(`Присоединились к комнате ${roomId}`);
    } else {
        alert(`Ошибка: ${result.data?.message || result.error}`);
    }
};

window.setPlayback = async function() {
    const roomId = document.getElementById('playbackRoomId')?.value;
    const action = document.getElementById('playbackAction')?.value;
    const timestamp = document.getElementById('playbackTimestamp')?.value;
    
    if (!roomId) {
        alert('Введите ID комнаты');
        return;
    }
    
    const result = await makeRequest(`${API_BASE_URL}/room/playback`, {
        method: 'POST',
        body: JSON.stringify({
            room_id: roomId,
            action: action,
            timestamp: parseInt(timestamp) || 0
        })
    });
    
    if (result.ok) {
        alert(`Действие "${action}" применено`);
    } else {
        alert(`Ошибка: ${result.data?.message || result.error}`);
    }
};

window.getRoomState = async function() {
    const roomId = document.getElementById('stateRoomId')?.value;
    
    if (!roomId) {
        alert('Введите ID комнаты');
        return;
    }
    
    const result = await makeRequest(`${API_BASE_URL}/room/state?room_id=${roomId}`);
    const resultDiv = document.getElementById('roomStateResult');
    
    if (resultDiv) {
        if (result.ok) {
            resultDiv.innerHTML = `<pre>${JSON.stringify(result.data, null, 2)}</pre>`;
        } else {
            resultDiv.innerHTML = `<div class="error">Ошибка: ${result.data?.message || result.error}</div>`;
        }
    }
};

// ========== WEBSOCKET ==========
window.connectWebSocket = function() {
    const roomId = document.getElementById('wsRoomId')?.value;
    const userId = document.getElementById('wsUserId')?.value;
    
    if (!roomId || !userId) {
        alert('Заполните ID комнаты и пользователя');
        return;
    }
    
    try {
        const wsUrl = `ws://localhost:8080/room/ws/?room_id=${roomId}&user_id=${userId}&token=${currentToken || 'test'}`;
        webSocket = new WebSocket(wsUrl);
        
        webSocket.onopen = function() {
            showNotification('WebSocket подключен', 'success');
            document.getElementById('wsConnectBtn').style.display = 'none';
            document.getElementById('wsDisconnectBtn').style.display = 'inline-block';
            addWsMessage('[SYSTEM] Подключение установлено');
        };
        
        webSocket.onmessage = function(event) {
            try {
                const data = JSON.parse(event.data);
                addWsMessage(`[${data.type}] ${JSON.stringify(data.payload || data)}`);
            } catch (e) {
                addWsMessage(`[RAW] ${event.data}`);
            }
        };
        
        webSocket.onerror = function(error) {
            console.error('WebSocket error:', error);
            addWsMessage('[SYSTEM] Ошибка соединения');
        };
        
        webSocket.onclose = function() {
            showNotification('WebSocket отключен', 'warning');
            document.getElementById('wsConnectBtn').style.display = 'inline-block';
            document.getElementById('wsDisconnectBtn').style.display = 'none';
            addWsMessage('[SYSTEM] Соединение закрыто');
        };
        
    } catch (error) {
        alert(`Ошибка WebSocket: ${error.message}`);
    }
};

window.disconnectWebSocket = function() {
    if (webSocket) {
        webSocket.close();
        webSocket = null;
    }
};

function addWsMessage(message) {
    const messagesDiv = document.getElementById('wsMessages');
    if (messagesDiv) {
        const messageElement = document.createElement('div');
        messageElement.textContent = `${new Date().toLocaleTimeString()}: ${message}`;
        messagesDiv.appendChild(messageElement);
        messagesDiv.scrollTop = messagesDiv.scrollHeight;
    }
}

// ========== ПОИСК ==========
window.performSearch = async function() {
    const query = document.getElementById('searchQuery')?.value;
    const type = document.getElementById('searchType')?.value;
    const limit = document.getElementById('searchLimit')?.value || 10;
    
    if (!query) {
        alert('Введите поисковый запрос');
        return;
    }
    
    const result = await makeRequest(`${API_BASE_URL}/search?q=${encodeURIComponent(query)}&type=${type}&limit=${limit}`);
    const resultsDiv = document.getElementById('searchResults');
    
    if (resultsDiv) {
        if (result.ok && result.data?.results) {
            const results = result.data.results;
            let html = `<h4>Найдено ${result.data.total || results.length} результатов:</h4>`;
            
            results.forEach(item => {
                html += `
                    <div class="search-result">
                        <strong>${item.title || item.name || item.id}:</strong>
                        <p>${item.description || ''}</p>
                        ${item.type ? `<small>Тип: ${item.type}</small>` : ''}
                    </div>
                    <hr>
                `;
            });
            
            resultsDiv.innerHTML = html;
        } else {
            resultsDiv.innerHTML = `<div class="error">Ошибка: ${result.data?.message || result.error || 'Ничего не найдено'}</div>`;
        }
    }
};

// ========== МЕДИА ==========
window.uploadFile = async function() {
    const fileInput = document.getElementById('fileInput');
    const userId = document.getElementById('uploadUserId')?.value;
    const chatId = document.getElementById('uploadChatId')?.value;
    
    if (!fileInput?.files[0] || !userId) {
        alert('Выберите файл и укажите ID пользователя');
        return;
    }
    
    const formData = new FormData();
    formData.append('file', fileInput.files[0]);
    formData.append('user_id', userId);
    if (chatId) formData.append('chat_id', chatId);
    
    try {
        const response = await fetch(`${API_BASE_URL}/media/upload`, {
            method: 'POST',
            body: formData,
            headers: currentToken ? { 'Authorization': `Bearer ${currentToken}` } : {}
        });
        
        const text = await response.text();
        const data = parseServerResponse(text);
        const resultDiv = document.getElementById('uploadResult');
        
        if (resultDiv) {
            if (response.ok) {
                resultDiv.innerHTML = `<pre>${JSON.stringify(data, null, 2)}</pre>`;
            } else {
                resultDiv.innerHTML = `<div class="error">Ошибка: ${data.message || text}</div>`;
            }
        }
    } catch (error) {
        alert(`Ошибка загрузки: ${error.message}`);
    }
};

window.downloadFile = function() {
    const fileId = document.getElementById('downloadFileId')?.value;
    if (!fileId) {
        alert('Введите ID файла');
        return;
    }
    
    window.open(`${API_BASE_URL}/media/download?id=${fileId}`, '_blank');
};

window.getFileMeta = async function() {
    const fileId = document.getElementById('metaFileId')?.value;
    if (!fileId) {
        alert('Введите ID файла');
        return;
    }
    
    const result = await makeRequest(`${API_BASE_URL}/media/meta?id=${fileId}`);
    const resultDiv = document.getElementById('metaResult');
    
    if (resultDiv) {
        if (result.ok) {
            resultDiv.innerHTML = `<pre>${JSON.stringify(result.data, null, 2)}</pre>`;
        } else {
            resultDiv.innerHTML = `<div class="error">Ошибка: ${result.data?.message || result.error}</div>`;
        }
    }
};

window.listUserFiles = async function() {
    const userId = document.getElementById('listUserId')?.value;
    if (!userId) {
        alert('Введите ID пользователя');
        return;
    }
    
    const result = await makeRequest(`${API_BASE_URL}/media/list?user_id=${userId}`);
    const resultDiv = document.getElementById('filesList');
    
    if (resultDiv) {
        if (result.ok && Array.isArray(result.data)) {
            let html = `<h4>Файлы пользователя (${result.data.length}):</h4>`;
            result.data.forEach(file => {
                html += `
                    <div class="file-item">
                        <strong>${file.filename || file.id}:</strong> ${file.size || 0} байт
                        <br><small>${new Date(file.created_at).toLocaleString()}</small>
                    </div>
                    <hr>
                `;
            });
            resultDiv.innerHTML = html;
        } else {
            resultDiv.innerHTML = `<div class="error">Ошибка: ${result.data?.message || result.error || 'Файлы не найдены'}</div>`;
        }
    }
};

// ========== СТАТУС ОНЛАЙН ==========
window.setOnlineStatus = async function() {
    if (!currentUser?.uuid) {
        alert('Войдите в систему');
        return;
    }
    
    const result = await makeRequest(`${API_BASE_URL}/user/set_online`, {
        method: 'POST',
        body: JSON.stringify({
            uuid: currentUser.uuid,
            ttl_seconds: 3600
        })
    });
    
    if (result.ok) {
        alert('Статус установлен: онлайн');
        if (document.getElementById('userStatus')) {
            document.getElementById('userStatus').textContent = 'Статус: онлайн';
        }
    } else {
        alert(`Ошибка: ${result.data?.message || result.error}`);
    }
};

window.setOfflineStatus = async function() {
    if (!currentUser?.uuid) {
        alert('Войдите в систему');
        return;
    }
    
    const result = await makeRequest(`${API_BASE_URL}/user/set_offline`, {
        method: 'POST',
        body: JSON.stringify({ uuid: currentUser.uuid })
    });
    
    if (result.ok) {
        alert('Статус установлен: офлайн');
        if (document.getElementById('userStatus')) {
            document.getElementById('userStatus').textContent = 'Статус: офлайн';
        }
    } else {
        alert(`Ошибка: ${result.data?.message || result.error}`);
    }
};

window.checkOnlineStatus = async function() {
    const uuid = prompt('Введите UUID для проверки:', currentUser?.uuid || 'user-123');
    if (!uuid) return;
    
    const result = await makeRequest(`${API_BASE_URL}/user/is_online?uuid=${uuid}`);
    const resultDiv = document.getElementById('statusResult');
    
    if (resultDiv) {
        if (result.ok && result.data) {
            resultDiv.innerHTML = `
                <div class="${result.data.online ? 'success' : 'warning'}">
                    Пользователь ${uuid}: ${result.data.online ? 'онлайн' : 'офлайн'}
                </div>
            `;
        } else {
            resultDiv.innerHTML = `<div class="error">Ошибка: ${result.data?.message || result.error}</div>`;
        }
    }
};

// ========== ПРОВЕРКА API ==========
window.checkAPIStatus = async function() {
    try {
        const response = await fetch(`${API_BASE_URL}/health`);
        const statusElement = document.getElementById('apiStatus');
        
        if (statusElement) {
            if (response.ok) {
                statusElement.textContent = 'Подключен';
                statusElement.className = 'status-online';
            } else {
                statusElement.textContent = 'Ошибка';
                statusElement.className = 'status-offline';
            }
        }
    } catch (error) {
        const statusElement = document.getElementById('apiStatus');
        if (statusElement) {
            statusElement.textContent = 'Не подключен';
            statusElement.className = 'status-offline';
        }
    }
};