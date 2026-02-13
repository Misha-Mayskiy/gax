// auth.js

// Конфигурация
const API_BASE_URL = 'http://localhost:8080';

// Текущее состояние
let currentUser = null;
let currentToken = null;

// Улучшенная функция для парсинга ответов сервера
function parseServerResponse(text) {
    if (!text || text.trim() === '') {
        return { 
            success: false, 
            message: 'Пустой ответ сервера',
            errorCode: 'EMPTY_RESPONSE' 
        };
    }
    
    // Сначала пытаемся парсить как JSON
    try {
        const data = JSON.parse(text);
        return {
            success: true,
            ...data,
            errorCode: null
        };
    } catch (e) {
        // Если не JSON, анализируем текст
        console.log('Raw server response (not JSON):', text);
        
        let message = text;
        let errorCode = 'UNKNOWN';
        
        // Извлекаем понятное сообщение об ошибке
        if (text.includes('rpc error:')) {
            // Упрощаем сообщение
            message = text.replace('rpc error: code = Unknown desc = ', '');
            errorCode = 'RPC_ERROR';
            
            // Преобразуем в понятные сообщения
            if (message.includes('name taken')) {
                message = 'Имя пользователя уже занято. Попробуйте другое.';
            } else if (message.includes('email taken')) {
                message = 'Email уже используется. Используйте другой email или восстановите пароль.';
            } else if (message.includes('already exists')) {
                message = 'Пользователь уже существует';
            } else if (message.includes('invalid') || message.includes('Invalid')) {
                message = 'Некорректные данные. Проверьте ввод.';
            } else if (message.includes('not found')) {
                message = 'Не найдено';
            } else if (message.includes('password')) {
                message = 'Ошибка пароля';
            }
        }
        
        return {
            success: false,
            message: message,
            errorCode: errorCode,
            raw: text
        };
    }
}

// Закрытие модального окна
function closeAuthModal() {
    const authModal = document.getElementById('authModal');
    if (authModal) {
        authModal.style.display = 'none';
    }
}

// Очистка форм
function clearAuthForms() {
    const loginEmail = document.getElementById('loginEmail');
    const loginPassword = document.getElementById('loginPassword');
    const regUsername = document.getElementById('regUsername');
    const regEmail = document.getElementById('regEmail');
    const regPassword = document.getElementById('regPassword');
    
    if (loginEmail) loginEmail.value = '';
    if (loginPassword) loginPassword.value = '';
    if (regUsername) regUsername.value = '';
    if (regEmail) regEmail.value = '';
    if (regPassword) regPassword.value = '';
}

// Простая функция alert если нет showNotification
function showAlert(message, type = 'info') {
    if (typeof showNotification === 'function') {
        showNotification(message, type);
    } else {
        console.log(`[${type.toUpperCase()}] ${message}`);
        alert(message);
    }
}

// Регистрация с улучшенной обработкой ошибок
async function register() {
    const usernameInput = document.getElementById('regUsername');
    const emailInput = document.getElementById('regEmail');
    const passwordInput = document.getElementById('regPassword');
    
    if (!usernameInput || !emailInput || !passwordInput) {
        alert('Ошибка: форма не найдена');
        return;
    }
    
    const username = usernameInput.value.trim();
    const email = emailInput.value.toLowerCase().trim();
    const password = passwordInput.value;
    
    // Валидация данных
    if (!username || !email || !password) {
        alert('Заполните все поля');
        return;
    }
    
    if (username.length < 3) {
        alert('Имя пользователя должно быть не менее 3 символов');
        return;
    }
    
    if (password.length < 6) {
        alert('Пароль должен быть не менее 6 символов');
        return;
    }
    
    if (!email.includes('@') || !email.includes('.')) {
        alert('Введите корректный email');
        return;
    }
    
    // Показываем индикатор загрузки
    const registerBtn = document.querySelector('#registerForm button');
    if (!registerBtn) return;
    
    const originalText = registerBtn.textContent;
    registerBtn.innerHTML = '<span class="loading"></span> Регистрация...';
    registerBtn.disabled = true;
    
    let registrationSuccessful = false;
    let userUuid = null;
    
    try {
        console.log('Starting registration for:', { username, email });
        
        // Шаг 1: Регистрация в auth-service
        const authResponse = await fetch(`${API_BASE_URL}/auth/register`, {
            method: 'POST',
            headers: { 'Content-Type': 'application/json' },
            body: JSON.stringify({ 
                username: username,
                email: email, 
                password: password 
            }),
            signal: AbortSignal.timeout(15000)
        });
        
        const authText = await authResponse.text();
        console.log('Auth service response:', authText);
        
        const authData = parseServerResponse(authText);
        
        if (authResponse.ok && authData.uuid) {
            registrationSuccessful = true;
            userUuid = authData.uuid;
            console.log('Auth registration successful, UUID:', userUuid);
            
            // Шаг 2: Создание пользователя в user-service
            try {
                console.log('Creating user in user-service with UUID:', userUuid);
                
                const userResponse = await fetch(`${API_BASE_URL}/user/create`, {
                    method: 'PUT',
                    headers: { 'Content-Type': 'application/json' },
                    body: JSON.stringify({
                        uuid: userUuid,
                        email: email,
                        user_name: username,
                        avatar: '',
                        about_me: 'Новый пользователь',
                        status: 'offline',
                        friends: []
                    }),
                    signal: AbortSignal.timeout(10000)
                });
                
                const userText = await userResponse.text();
                console.log('User service response:', userText);
                
                if (userResponse.ok) {
                    console.log('User created successfully in user-service');
                    showAlert('Регистрация успешна! Теперь войдите в систему.', 'success');
                } else {
                    console.warn('Failed to create user in user-service:', userText);
                    showAlert('Аккаунт создан, но произошла ошибка при создании профиля. Пожалуйста, войдите и обновите профиль.', 'warning');
                }
            } catch (userError) {
                console.error('Error creating user in user-service:', userError);
                showAlert('Аккаунт создан, но произошла ошибка при создании профиля. Вы можете войти и обновить профиль.', 'warning');
            }
            
            // Переключаем на форму входа
            if (typeof switchTab === 'function') {
                switchTab('login');
            }
            
            // Автоматически заполняем email в форме входа
            const loginEmail = document.getElementById('loginEmail');
            if (loginEmail) loginEmail.value = email;
            
        } else {
            // Обработка ошибок регистрации
            let errorMessage = authData.message || 'Ошибка регистрации';
            
            // Специфичные сообщения для пользователя
            if (errorMessage.includes('name taken')) {
                errorMessage = `Имя пользователя "${username}" уже занято. Выберите другое имя.`;
            } else if (errorMessage.includes('email taken')) {
                errorMessage = `Email "${email}" уже используется. Используйте другой email или восстановите пароль.`;
            } else if (errorMessage.includes('already exists')) {
                errorMessage = 'Пользователь с такими данными уже существует.';
            }
            
            alert(errorMessage);
        }
    } catch (error) {
        console.error('Register error:', error);
        
        let errorMessage = 'Ошибка регистрации';
        
        if (error.name === 'AbortError') {
            errorMessage = 'Регистрация заняла слишком много времени. Проверьте подключение к серверу и попробуйте снова.';
        } else if (error.message.includes('Failed to fetch')) {
            errorMessage = 'Не удалось подключиться к серверу. Проверьте подключение к интернету.';
        } else {
            errorMessage = error.message || 'Неизвестная ошибка';
        }
        
        alert(errorMessage);
    } finally {
        // Восстанавливаем кнопку
        registerBtn.textContent = originalText;
        registerBtn.disabled = false;
    }
}

// Вход
async function login() {
    const emailInput = document.getElementById('loginEmail');
    const passwordInput = document.getElementById('loginPassword');
    
    if (!emailInput || !passwordInput) {
        alert('Ошибка: форма не найдена');
        return;
    }
    
    const email = emailInput.value.trim();
    const password = passwordInput.value;
    
    if (!email || !password) {
        alert('Заполните все поля');
        return;
    }
    
    // Показываем индикатор загрузки
    const loginBtn = document.querySelector('#loginForm button');
    if (!loginBtn) return;
    
    const originalText = loginBtn.textContent;
    loginBtn.innerHTML = '<span class="loading"></span> Вход...';
    loginBtn.disabled = true;
    
    try {
        // Шаг 1: Вход в auth-service
        const authResponse = await fetch(`${API_BASE_URL}/auth/login`, {
            method: 'POST',
            headers: { 'Content-Type': 'application/json' },
            body: JSON.stringify({ 
                email: email, 
                password: password 
            }),
            signal: AbortSignal.timeout(10000)
        });
        
        const authText = await authResponse.text();
        const authData = parseServerResponse(authText);
        
        if (authResponse.ok && authData.uuid) {
            let userUuid = authData.uuid;
            
            // Шаг 2: Проверяем существование в user-service
            try {
                const userCheckResponse = await fetch(`${API_BASE_URL}/user/get?uuid=${userUuid}`, {
                    signal: AbortSignal.timeout(5000)
                });
                
                if (!userCheckResponse.ok) {
                    // Если пользователь не найден в user-service, создаем его
                    console.log('User not found in user-service, creating...');
                    
                    const username = email.split('@')[0];
                    
                    const createUserResponse = await fetch(`${API_BASE_URL}/user/create`, {
                        method: 'PUT',
                        headers: { 'Content-Type': 'application/json' },
                        body: JSON.stringify({
                            uuid: userUuid,
                            email: email,
                            user_name: username,
                            avatar: '',
                            about_me: '',
                            status: 'offline',
                            friends: []
                        }),
                        signal: AbortSignal.timeout(5000)
                    });
                    
                    if (!createUserResponse.ok) {
                        console.warn('Failed to create user in user-service');
                    }
                }
                
                // Сохраняем данные пользователя
                currentUser = { 
                    uuid: userUuid,
                    email: email,
                    username: email.split('@')[0]
                };
                
                currentToken = userUuid;
                
                // Сохраняем в localStorage
                localStorage.setItem('gax_user', JSON.stringify(currentUser));
                localStorage.setItem('gax_token', currentToken);
                
                // Сохраняем токен в куках
                document.cookie = `jwtToken=${currentToken}; path=/; max-age=86400; SameSite=Lax`;
                
                // Обновляем интерфейс
                if (typeof updateUIAfterLogin === 'function') {
                    updateUIAfterLogin();
                }
                
                closeAuthModal();
                showAlert('Успешный вход!', 'success');
                
            } catch (syncError) {
                console.error('Error syncing with user-service:', syncError);
                
                // Все равно сохраняем пользователя
                currentUser = { 
                    uuid: userUuid,
                    email: email,
                    username: email.split('@')[0]
                };
                currentToken = userUuid;
                
                localStorage.setItem('gax_user', JSON.stringify(currentUser));
                localStorage.setItem('gax_token', currentToken);
                document.cookie = `jwtToken=${currentToken}; path=/; max-age=86400; SameSite=Lax`;
                
                if (typeof updateUIAfterLogin === 'function') {
                    updateUIAfterLogin();
                }
                
                closeAuthModal();
                showAlert('Вход успешен!', 'success');
            }
            
        } else {
            let errorMessage = authData.message || 'Ошибка входа';
            
            if (errorMessage.includes('user with this email not found') || errorMessage.includes('not found')) {
                errorMessage = 'Пользователь не найден. Зарегистрируйтесь сначала.';
            } else if (errorMessage.includes('wrong password') || errorMessage.includes('invalid password')) {
                errorMessage = 'Неверный пароль';
            } else if (errorMessage.includes('invalid credentials')) {
                errorMessage = 'Неверный email или пароль';
            }
            
            alert(errorMessage);
        }
    } catch (error) {
        console.error('Login error:', error);
        
        let errorMessage = 'Ошибка входа';
        
        if (error.name === 'AbortError') {
            errorMessage = 'Вход занял слишком много времени. Проверьте подключение к серверу.';
        } else if (error.message.includes('Failed to fetch')) {
            errorMessage = 'Не удалось подключиться к серверу. Проверьте подключение к интернету.';
        }
        
        alert(errorMessage);
    } finally {
        // Восстанавливаем кнопку
        loginBtn.textContent = originalText;
        loginBtn.disabled = false;
    }
}

// Проверка статуса аутентификации
function checkAuthStatus() {
    // Проверяем localStorage
    const savedUser = localStorage.getItem('gax_user');
    const savedToken = localStorage.getItem('gax_token');
    
    if (savedUser && savedToken) {
        try {
            currentUser = JSON.parse(savedUser);
            currentToken = savedToken;
            
            if (typeof updateUIAfterLogin === 'function') {
                updateUIAfterLogin();
            }
            return;
        } catch (e) {
            console.error('Failed to parse saved user:', e);
            localStorage.removeItem('gax_user');
            localStorage.removeItem('gax_token');
        }
    }
    
    // Проверяем куки
    const cookies = document.cookie.split(';');
    const jwtCookie = cookies.find(c => c.trim().startsWith('jwtToken='));
    
    if (jwtCookie) {
        const token = jwtCookie.split('=')[1];
        if (token && token !== 'undefined') {
            currentToken = token;
            // Создаем временного пользователя
            currentUser = {
                uuid: 'temp-user-' + Date.now(),
                email: 'user@example.com',
                username: 'User'
            };
            
            if (typeof updateUIAfterLogin === 'function') {
                updateUIAfterLogin();
            }
        }
    }
}

// Инициализация
window.addEventListener('DOMContentLoaded', () => {
    // Закрытие модального окна при клике вне его
    const authModal = document.getElementById('authModal');
    if (authModal) {
        window.addEventListener('click', (e) => {
            if (e.target === authModal) {
                closeAuthModal();
            }
        });
    }
    
    // Проверяем статус аутентификации
    checkAuthStatus();
});