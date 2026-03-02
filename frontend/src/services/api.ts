import axios from 'axios';

// Убедись что URL правильный
const API_URL = import.meta.env.VITE_API_URL || 'http://localhost:8080/api';

console.log('API URL:', API_URL); // Добавь для отладки

const api = axios.create({
    baseURL: API_URL,
    headers: {
        'Content-Type': 'application/json',
    },
});

// Добавляем токен к каждому запросу
api.interceptors.request.use((config) => {
    const token = localStorage.getItem('token');
    if (token) {
        config.headers.Authorization = `Bearer ${token}`;
    }
    console.log('Request:', config.method?.toUpperCase(), config.url); // Добавь логи
    return config;
});

// Обработка ошибок
api.interceptors.response.use(
    (response) => {
        console.log('Response:', response.status, response.config.url); // Добавь логи
        return response;
    },
    (error) => {
        console.error('API Error:', error.response?.status, error.response?.data); // Подробный лог ошибок
        return Promise.reject(error);
    }
);

export default api;