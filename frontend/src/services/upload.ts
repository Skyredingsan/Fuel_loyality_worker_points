import api from './api';

export const uploadService = {
    async uploadFile(file: File, type: string, entityId: string): Promise<string> {
        const formData = new FormData();
        formData.append('file', file);
        formData.append('type', type);
        formData.append('entity_id', entityId);

        const response = await api.post<{ url: string }>('/upload', formData, {
            headers: {
                'Content-Type': 'multipart/form-data',
            },
        });

        return response.data.url;
    },

    async deleteFile(fileUrl: string): Promise<void> {
        // Извлекаем type и filename из URL
        // URL формата: /uploads/indicator_result/123456_filename.pdf
        const matches = fileUrl.match(/\/uploads\/([^\/]+)\/(.+)$/);
        if (!matches) {
            throw new Error('Invalid file URL');
        }

        const [, type, filename] = matches;
        await api.delete(`/upload/${type}/${filename}`);
    },

    getFileUrl(fileUrl: string): string {
        // Если URL уже полный, возвращаем как есть
        if (fileUrl.startsWith('http')) {
            return fileUrl;
        }
        // Иначе добавляем базовый URL
        return `http://localhost:8080${fileUrl}`;
    }
};