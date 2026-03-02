import api from './api';
import type { User } from '../types';

export const userService = {
    async getTMs(): Promise<User[]> {
        const response = await api.get<User[]>('/tms');
        return response.data;
    },

    async getAllUsers(role?: string): Promise<User[]> {
        const url = role ? `/users?role=${role}` : '/users';
        const response = await api.get<User[]>(url);
        return response.data;
    },

    async getUserById(id: number): Promise<User> {
        const response = await api.get<User>(`/users/${id}`);
        return response.data;
    },

    async createUser(data: Partial<User>): Promise<User> {
        const response = await api.post<User>('/users/register', data);
        return response.data;
    },

    async updateUser(id: number, data: Partial<User>): Promise<User> {
        const response = await api.put<User>(`/users/${id}`, data);
        return response.data;
    },

    async deleteUser(id: number): Promise<void> {
        await api.delete(`/users/${id}`);
    }
};