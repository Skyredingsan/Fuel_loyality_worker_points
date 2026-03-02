import React, { useState, useEffect } from 'react';
import { useAuth } from '../contexts/AuthContext';
import { resultsService } from '../services/results';
import { kpiService } from '../services/kpi';
import { userService } from '../services/users';
import { uploadService } from '../services/upload';
import type { User, KPIIndicator, IndicatorResultInput } from '../types';

export const EnterResults: React.FC = () => {
    const { user } = useAuth();
    const [tms, setTms] = useState<User[]>([]);
    const [indicators, setIndicators] = useState<KPIIndicator[]>([]);
    const [selectedTM, setSelectedTM] = useState<number | ''>('');
    const [period, setPeriod] = useState(() => {
        const now = new Date();
        return `${now.getFullYear()}-${String(now.getMonth() + 1).padStart(2, '0')}`;
    });
    const [results, setResults] = useState<Record<string, { value: string; file?: File; fileUrl?: string }>>({});
    const [uploading, setUploading] = useState<Record<string, boolean>>({});
    const [loading, setLoading] = useState(false);
    const [saving, setSaving] = useState(false);
    const [message, setMessage] = useState<{ type: 'success' | 'error'; text: string } | null>(null);
    const [isEditing, setIsEditing] = useState(false);
    const [editingId, setEditingId] = useState<number | null>(null);

    useEffect(() => {
        // Проверяем URL на наличие параметра edit
        const params = new URLSearchParams(window.location.search);
        const editId = params.get('edit');

        if (editId) {
            setIsEditing(true);
            setEditingId(parseInt(editId));
            loadDataForEditing(parseInt(editId));
        } else {
            loadData();
        }
    }, []);

    const loadData = async () => {
        try {
            setLoading(true);
            // Загружаем список ТМ и индикаторы
            const [tmsData, indicatorsData] = await Promise.all([
                userService.getTMs(),
                kpiService.getAllIndicators()
            ]);
            setTms(tmsData);
            setIndicators(indicatorsData);
        } catch (error) {
            console.error('Failed to load data:', error);
            setMessage({ type: 'error', text: 'Ошибка загрузки данных' });
        } finally {
            setLoading(false);
        }
    };

    const loadDataForEditing = async (resultId: number) => {
        try {
            setLoading(true);

            // Загружаем список ТМ и индикаторы
            const [tmsData, indicatorsData, result] = await Promise.all([
                userService.getTMs(),
                kpiService.getAllIndicators(),
                resultsService.getResultById(resultId)
            ]);

            console.log('Editing result:', result);
            setTms(tmsData);
            setIndicators(indicatorsData);

            // Устанавливаем выбранного ТМ
            setSelectedTM(result.user_id);

            // Устанавливаем период
            setPeriod(result.period);

            // Заполняем результаты
            const loadedResults: Record<string, { value: string; fileUrl?: string }> = {};
            result.indicators.forEach((ind: any) => {
                loadedResults[ind.indicator.code] = {
                    value: ind.fact_value?.toString() || '',
                    fileUrl: ind.supporting_document_url
                };
            });
            setResults(loadedResults);

        } catch (error) {
            console.error('Failed to load result for editing:', error);
            setMessage({ type: 'error', text: 'Ошибка загрузки данных для редактирования' });
        } finally {
            setLoading(false);
        }
    };

    const handleInputChange = (indicatorCode: string, value: string) => {
        setResults(prev => ({
            ...prev,
            [indicatorCode]: { ...prev[indicatorCode], value }
        }));
    };

    const handleFileChange = async (indicatorCode: string, file: File | undefined) => {
        if (!file) {
            setResults(prev => ({
                ...prev,
                [indicatorCode]: { ...prev[indicatorCode], file: undefined, fileUrl: undefined }
            }));
            return;
        }

        try {
            setUploading(prev => ({ ...prev, [indicatorCode]: true }));

            const fileUrl = await uploadService.uploadFile(file, 'indicator_result', indicatorCode);

            setResults(prev => ({
                ...prev,
                [indicatorCode]: {
                    ...prev[indicatorCode],
                    file,
                    fileUrl
                }
            }));

            setMessage({ type: 'success', text: 'Файл загружен' });
        } catch (error) {
            console.error('Failed to upload file:', error);
            setMessage({ type: 'error', text: 'Ошибка загрузки файла' });
        } finally {
            setUploading(prev => ({ ...prev, [indicatorCode]: false }));
        }
    };

    const handleSubmit = async (e: React.FormEvent) => {
        e.preventDefault();

        if (!selectedTM) {
            setMessage({ type: 'error', text: 'Выберите ТМ' });
            return;
        }

        try {
            setSaving(true);
            setMessage(null);

            const resultsToSend: IndicatorResultInput[] = Object.entries(results)
                .filter(([_, data]) => data.value.trim() !== '')
                .map(([code, data]) => ({
                    indicator_code: code,
                    fact_value: parseFloat(data.value) || 0,
                    document_url: data.fileUrl
                }));

            if (resultsToSend.length === 0) {
                setMessage({ type: 'error', text: 'Введите хотя бы один результат' });
                return;
            }

            if (isEditing && editingId) {
                // Обновляем существующий результат
                await resultsService.updateResults(editingId, {
                    user_id: Number(selectedTM),
                    period: period,
                    results: resultsToSend
                });
                setMessage({ type: 'success', text: 'Результаты успешно обновлены' });
            } else {
                // Создаем новый результат
                await resultsService.enterResults({
                    user_id: Number(selectedTM),
                    period: period,
                    results: resultsToSend
                });
                setMessage({ type: 'success', text: 'Результаты успешно сохранены' });
            }

            // Очищаем форму если это не редактирование
            if (!isEditing) {
                setResults({});
            }

        } catch (error: any) {
            console.error('Failed to save results:', error);
            setMessage({
                type: 'error',
                text: error.response?.data || 'Ошибка сохранения результатов'
            });
        } finally {
            setSaving(false);
        }
    };

    const groupedIndicators = indicators.reduce((acc, ind) => {
        const category = ind.category_code || 'Другое';
        if (!acc[category]) {
            acc[category] = [];
        }
        acc[category].push(ind);
        return acc;
    }, {} as Record<string, KPIIndicator[]>);

    if (loading) {
        return (
            <div className="flex justify-center items-center h-64">
                <div className="animate-spin rounded-full h-12 w-12 border-b-2 border-blue-600"></div>
            </div>
        );
    }

    return (
        <div className="max-w-4xl mx-auto">
            <h1 className="text-2xl font-bold mb-6">
                {isEditing ? 'Редактирование результатов' : 'Ввод результатов'}
            </h1>

            {message && (
                <div className={`p-4 rounded-lg mb-4 ${
                    message.type === 'success' ? 'bg-green-50 text-green-800' : 'bg-red-50 text-red-800'
                }`}>
                    {message.text}
                </div>
            )}

            <form onSubmit={handleSubmit} className="space-y-6">
                {/* Выбор ТМ и периода */}
                <div className="bg-white p-6 rounded-lg shadow">
                    <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
                        <div>
                            <label className="block text-sm font-medium text-gray-700 mb-2">
                                Территориальный менеджер
                            </label>
                            <select
                                value={selectedTM}
                                onChange={(e) => setSelectedTM(e.target.value ? Number(e.target.value) : '')}
                                className="w-full px-3 py-2 border border-gray-300 rounded-lg focus:outline-none focus:ring-2 focus:ring-blue-500"
                                required
                                disabled={isEditing} // При редактировании нельзя изменить ТМ
                            >
                                <option value="">Выберите ТМ</option>
                                {tms.map(tm => (
                                    <option key={tm.id} value={tm.id}>
                                        {tm.fio} ({tm.cluster_name || 'нет кластера'})
                                    </option>
                                ))}
                            </select>
                            {isEditing && (
                                <p className="text-xs text-gray-500 mt-1">
                                    ТМ нельзя изменить при редактировании
                                </p>
                            )}
                        </div>
                        <div>
                            <label className="block text-sm font-medium text-gray-700 mb-2">
                                Период
                            </label>
                            <input
                                type="month"
                                value={period}
                                onChange={(e) => setPeriod(e.target.value)}
                                className="w-full px-3 py-2 border border-gray-300 rounded-lg focus:outline-none focus:ring-2 focus:ring-blue-500"
                                required
                            />
                        </div>
                    </div>
                </div>

                {/* Показатели по категориям */}
                {Object.entries(groupedIndicators).map(([category, cats]) => (
                    <div key={category} className="bg-white p-6 rounded-lg shadow">
                        <h2 className="text-lg font-semibold mb-4">
                            {category === 'ПМ' && '📊 Продажи и маржа'}
                            {category === 'ОЭК' && '⚙️ Операционная эффективность'}
                            {category === 'ЭКЛ' && '👥 Эффективность команды'}
                            {category === 'КБ' && '🔒 Культура безопасности'}
                            {!['ПМ', 'ОЭК', 'ЭКЛ', 'КБ'].includes(category) && category}
                        </h2>
                        <div className="space-y-6">
                            {cats.map(indicator => (
                                <div key={indicator.id} className="border-b pb-4 last:border-b-0">
                                    <div className="grid grid-cols-1 md:grid-cols-3 gap-4 items-start mb-2">
                                        <div>
                                            <label className="block text-sm font-medium text-gray-700">
                                                {indicator.name}
                                            </label>
                                            <span className="text-xs text-gray-500">
                        {indicator.indicator_type === 'base' && '🎯 Базовый'}
                                                {indicator.indicator_type === 'extra' && '➕ Дополнительный'}
                                                {indicator.indicator_type === 'penalty' && '⚠️ Штрафной'}
                                                {' • '}{indicator.unit}
                      </span>
                                        </div>
                                        <div>
                                            <input
                                                type="number"
                                                step="0.01"
                                                value={results[indicator.code]?.value || ''}
                                                onChange={(e) => handleInputChange(indicator.code, e.target.value)}
                                                placeholder={`Введите значение в ${indicator.unit}`}
                                                className="w-full px-3 py-2 border border-gray-300 rounded-lg focus:outline-none focus:ring-2 focus:ring-blue-500"
                                            />
                                        </div>
                                        <div>
                                            <div className="flex items-center space-x-2">
                                                <input
                                                    type="file"
                                                    id={`file-${indicator.code}`}
                                                    onChange={(e) => handleFileChange(
                                                        indicator.code,
                                                        e.target.files?.[0]
                                                    )}
                                                    className="hidden"
                                                    accept=".pdf,.jpg,.jpeg,.png,.doc,.docx,.xls,.xlsx"
                                                />
                                                <label
                                                    htmlFor={`file-${indicator.code}`}
                                                    className={`px-3 py-2 border rounded-lg cursor-pointer hover:bg-gray-50 transition ${
                                                        results[indicator.code]?.file ? 'bg-green-50 border-green-300' : ''
                                                    }`}
                                                >
                                                    {uploading[indicator.code] ? (
                                                        <span className="flex items-center">
                              <svg className="animate-spin h-4 w-4 mr-2" viewBox="0 0 24 24">
                                <circle className="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" strokeWidth="4" fill="none" />
                                <path className="opacity-75" fill="currentColor" d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4zm2 5.291A7.962 7.962 0 014 12H0c0 3.042 1.135 5.824 3 7.938l3-2.647z" />
                              </svg>
                              Загрузка...
                            </span>
                                                    ) : (
                                                        <span className="flex items-center">
                              <span className="mr-2">📎</span>
                                                            {results[indicator.code]?.file ? 'Файл выбран' : 'Выбрать файл'}
                            </span>
                                                    )}
                                                </label>
                                                {results[indicator.code]?.file && (
                                                    <button
                                                        type="button"
                                                        onClick={() => handleFileChange(indicator.code, undefined)}
                                                        className="text-red-600 hover:text-red-800"
                                                        title="Удалить файл"
                                                    >
                                                        ✕
                                                    </button>
                                                )}
                                            </div>
                                            {results[indicator.code]?.file && (
                                                <div className="mt-1 text-xs text-gray-500">
                                                    {results[indicator.code].file?.name}
                                                </div>
                                            )}
                                            {results[indicator.code]?.fileUrl && !results[indicator.code]?.file && (
                                                <div className="mt-1 text-xs text-blue-600">
                                                    <a
                                                        href={`http://localhost:8080${results[indicator.code].fileUrl}`}
                                                        target="_blank"
                                                        rel="noopener noreferrer"
                                                        className="hover:underline"
                                                    >
                                                        📎 Текущий файл
                                                    </a>
                                                </div>
                                            )}
                                        </div>
                                    </div>
                                    {indicator.base_value !== undefined && indicator.indicator_type === 'base' && (
                                        <div className="text-xs text-gray-500 ml-1">
                                            Цель: {indicator.base_value}{indicator.unit} (вес: {indicator.base_weight})
                                        </div>
                                    )}
                                </div>
                            ))}
                        </div>
                    </div>
                ))}

                {/* Кнопка сохранения */}
                <div className="flex justify-end">
                    <button
                        type="submit"
                        disabled={saving || !selectedTM}
                        className="bg-blue-600 text-white px-6 py-2 rounded-lg hover:bg-blue-700 transition disabled:opacity-50 flex items-center"
                    >
                        {saving ? (
                            <>
                                <svg className="animate-spin h-5 w-5 mr-2" viewBox="0 0 24 24">
                                    <circle className="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" strokeWidth="4" fill="none" />
                                    <path className="opacity-75" fill="currentColor" d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4zm2 5.291A7.962 7.962 0 014 12H0c0 3.042 1.135 5.824 3 7.938l3-2.647z" />
                                </svg>
                                {isEditing ? 'Обновление...' : 'Сохранение...'}
                            </>
                        ) : (
                            isEditing ? 'Обновить результаты' : 'Сохранить результаты'
                        )}
                    </button>
                </div>
            </form>
        </div>
    );
};