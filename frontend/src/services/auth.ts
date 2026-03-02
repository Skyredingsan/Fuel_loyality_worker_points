import api from './api';
import type {
    LoginRequest,
    LoginResponse,
    User
} from '../types';

export const authService = {
    async login(data: LoginRequest): Promise<LoginResponse> {
        try {
            const response = await api.post<LoginResponse>('/login', data);
            if (response.data.token) {
                localStorage.setItem('token', response.data.token);
                localStorage.setItem('user', JSON.stringify(response.data.user));
            }
            return response.data;
        } catch (error: any) {
            throw error.response?.data || { message: 'Login failed' };
        }
    },

    logout(): void {
        localStorage.removeItem('token');
        localStorage.removeItem('user');
        window.location.href = '/login';
    },

    getCurrentUser(): User | null {
        const userStr = localStorage.getItem('user');
        if (userStr) {
            try {
                return JSON.parse(userStr) as User;
            } catch {
                return null;
            }
        }
        return null;
    },

    async refreshToken(): Promise<string> {
        const response = await api.post<{ token: string }>('/users/refresh');
        localStorage.setItem('token', response.data.token);
        return response.data.token;
    },

    async getMe(): Promise<User> {
        const response = await api.get<User>('/users/me');
        localStorage.setItem('user', JSON.stringify(response.data));
        return response.data;
    }
};