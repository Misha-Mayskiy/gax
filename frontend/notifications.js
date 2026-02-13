// notifications.js

// Функция для показа уведомлений
window.showNotification = function(message, type = 'info') {
    console.log(`Notification [${type}]: ${message}`);
    
    // Создаем элемент уведомления
    const notification = document.createElement('div');
    notification.textContent = message;
    notification.className = `notification notification-${type}`;
    
    // Стили для уведомления
    notification.style.cssText = `
        position: fixed;
        top: 20px;
        right: 20px;
        padding: 12px 20px;
        border-radius: 8px;
        z-index: 1001;
        font-weight: 500;
        box-shadow: 0 2px 10px rgba(0,0,0,0.1);
        max-width: 400px;
        animation: slideIn 0.3s ease;
        backdrop-filter: blur(10px);
        border: 1px solid;
    `;
    
    // Цвета в зависимости от типа
    switch (type) {
        case 'success':
            notification.style.backgroundColor = 'rgba(212, 237, 218, 0.9)';
            notification.style.color = '#155724';
            notification.style.borderColor = '#c3e6cb';
            break;
        case 'error':
            notification.style.backgroundColor = 'rgba(248, 215, 218, 0.9)';
            notification.style.color = '#721c24';
            notification.style.borderColor = '#f5c6cb';
            break;
        case 'warning':
            notification.style.backgroundColor = 'rgba(255, 243, 205, 0.9)';
            notification.style.color = '#856404';
            notification.style.borderColor = '#ffeaa7';
            break;
        default:
            notification.style.backgroundColor = 'rgba(209, 236, 241, 0.9)';
            notification.style.color = '#0c5460';
            notification.style.borderColor = '#bee5eb';
    }
    
    document.body.appendChild(notification);
    
    // Удаляем через 5 секунд
    setTimeout(() => {
        notification.style.animation = 'slideOut 0.3s ease';
        setTimeout(() => {
            if (notification.parentNode) {
                notification.parentNode.removeChild(notification);
            }
        }, 300);
    }, 5000);
};

// Добавьте стили для анимации в CSS
const style = document.createElement('style');
style.textContent = `
    @keyframes slideIn {
        from { 
            transform: translateX(100%); 
            opacity: 0; 
        }
        to { 
            transform: translateX(0); 
            opacity: 1; 
        }
    }
    
    @keyframes slideOut {
        from { 
            transform: translateX(0); 
            opacity: 1; 
        }
        to { 
            transform: translateX(100%); 
            opacity: 0; 
        }
    }
`;
document.head.appendChild(style);