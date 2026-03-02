import api from './api';
import type { KPICategory, KPIIndicator } from '../types';

export const kpiService = {
    async getCategories(): Promise<KPICategory[]> {
        const response = await api.get<KPICategory[]>('/kpi/categories');
        return response.data;
    },

    async getAllIndicators(): Promise<KPIIndicator[]> {
        const response = await api.get<KPIIndicator[]>('/kpi/indicators');
        return response.data;
    },

    async getIndicatorsByCategory(categoryCode: string): Promise<KPIIndicator[]> {
        const response = await api.get<KPIIndicator[]>(`/kpi/categories/${categoryCode}/indicators`);
        return response.data;
    },

    async createIndicator(data: Partial<KPIIndicator>): Promise<KPIIndicator> {
        const response = await api.post<KPIIndicator>('/kpi/indicators', data);
        return response.data;
    },

    async updateIndicator(id: number, data: Partial<KPIIndicator>): Promise<KPIIndicator> {
        const response = await api.put<KPIIndicator>(`/kpi/indicators/${id}`, data);
        return response.data;
    },

    async deleteIndicator(id: number): Promise<void> {
        await api.delete(`/kpi/indicators/${id}`);
    }
};