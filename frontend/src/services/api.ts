import axios from 'axios';

// URL бэкенда в продакшене и разработке
const API_URL = 'https://fuel-loyality-worker-points.onrender.com/api';

const api = axios.create({
    baseURL: API_URL,
    headers: {
        'Content-Type': 'application/json',
    },
    timeout: 30000, // 30 секунд таймаут (на случай медленного ответа)
});

// Интерсептор: добавляем токен к каждому запросу
api.interceptors.request.use(
    (config) => {
        const token = localStorage.getItem('token');
        if (token) {
            config.headers.Authorization = `Bearer ${token}`;
        }

        // Для отладки в продакшене (можно убрать после настройки)
        if (import.meta.env.DEV) {
            console.log(`[API] ${config.method?.toUpperCase()} ${config.url}`);
        }

        return config;
    },
    (error) => {
        console.error('[API] Request error:', error);
        return Promise.reject(error);
    }
);

// Интерсептор: обработка ошибок ответа
api.interceptors.response.use(
    (response) => {
        // Успешный ответ
        return response;
    },
    (error) => {
        // Ошибка от сервера
        if (error.response) {
            // Сервер ответил с ошибкой
            const { status, data } = error.response;

            console.error(`[API] Error ${status}:`, data);

            // Если 401 Unauthorized -> токен истек или невалидный
            if (status === 401) {
                console.warn('[API] Token expired or invalid, logging out...');
                localStorage.removeItem('token');
                localStorage.removeItem('user');

                // Не редиректим на /login, если уже на странице логина
                if (window.location.pathname !== '/login') {
                    window.location.href = '/login';
                }
            }

            // Если 403 Forbidden -> недостаточно прав
            if (status === 403) {
                console.warn('[API] Access forbidden:', data?.message || 'No permission');
            }

            return Promise.reject(error);
        }

        // Ошибка сети (сервер не отвечает)
        if (error.request) {
            console.error('[API] Network error - no response from server:', error.request);
            return Promise.reject(new Error('Сервер не отвечает. Проверьте подключение.'));
        }

        // Другие ошибки
        console.error('[API] Unknown error:', error.message);
        return Promise.reject(error);
    }
);

export default api;