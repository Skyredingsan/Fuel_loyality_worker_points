import api from './api';
import type {
    EnterResultRequest,
    FullResultSummary,
    MonthlyResult,
    IndicatorResult
} from '../types';

export const resultsService = {
    async enterResults(data: EnterResultRequest): Promise<MonthlyResult> {
        const response = await api.post<MonthlyResult>('/results/enter', data);
        return response.data;
    },

    async confirmResults(id: number): Promise<void> {
        await api.post(`/results/${id}/confirm`);
    },

    // НОВЫЙ МЕТОД: Отклонить результаты
    async rejectResults(id: number, reason: string): Promise<void> {
        await api.post(`/results/${id}/reject`, { reason });
    },

    // НОВЫЙ МЕТОД: Получить результат по ID для редактирования
    async getResultById(id: number): Promise<any> {
        const response = await api.get(`/results/${id}`);
        return response.data;
    },

    // НОВЫЙ МЕТОД: Обновить результаты (для черновиков)
    async updateResults(id: number, data: EnterResultRequest): Promise<MonthlyResult> {
        const response = await api.put<MonthlyResult>(`/results/${id}`, data);
        return response.data;
    },

    async getMyResults(period?: string): Promise<FullResultSummary> {
        const url = period ? `/results/my?period=${period}` : '/results/my';
        const response = await api.get<FullResultSummary>(url);
        return response.data;
    },

    async getUserResults(userId: number, period: string): Promise<FullResultSummary> {
        const response = await api.get<FullResultSummary>(`/results/user/${userId}?period=${period}`);
        return response.data;
    },

    async getDetailedResults(monthlyResultId: number): Promise<IndicatorResult[]> {
        const response = await api.get<IndicatorResult[]>(`/results/${monthlyResultId}/indicators`);
        return response.data;
    },

    async getYearlySummary(userId: number, year: number): Promise<any> {
        const response = await api.get(`/results/user/${userId}/yearly?year=${year}`);
        return response.data;
    },

    async getAllResults(period: string): Promise<any[]> {
        const response = await api.get(`/results?period=${period}`);
        return response.data;
    },

    async getIndicatorResults(monthlyResultId: number): Promise<IndicatorResult[]> {
        const response = await api.get(`/results/${monthlyResultId}/indicators`);
        return response.data;
    }
};