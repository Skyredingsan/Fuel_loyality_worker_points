import React, { useEffect, useState } from 'react';
import { useAuth } from '../contexts/AuthContext';
import { resultsService } from '../services/results';
import type { FullResultSummary } from '../types';

export const Dashboard: React.FC = () => {
    const { user } = useAuth();
    const [summary, setSummary] = useState<FullResultSummary | null>(null);
    const [loading, setLoading] = useState(true);
    const [error, setError] = useState('');
    const [period, setPeriod] = useState(() => {
        const now = new Date();
        return `${now.getFullYear()}-${String(now.getMonth() + 1).padStart(2, '0')}`;
    });

    useEffect(() => {
        loadResults();
    }, [period]);

    const loadResults = async () => {
        try {
            setLoading(true);
            setError('');

            if (user?.role === 'tm') {
                const data = await resultsService.getMyResults(period);
                setSummary(data);
            } else {
                // Для экспертов и координаторов показываем заглушку
                setSummary(null);
            }
        } catch (err: any) {
            console.error('Failed to load results:', err);
            setError(err.message || 'Ошибка загрузки данных');
        } finally {
            setLoading(false);
        }
    };

    if (loading) {
        return (
            <div className="flex justify-center items-center h-64">
                <div className="animate-spin rounded-full h-12 w-12 border-b-2 border-blue-600"></div>
            </div>
        );
    }

    return (
        <div className="space-y-6">
            <div className="flex justify-between items-center">
                <h1 className="text-2xl font-bold">Дашборд</h1>
                <input
                    type="month"
                    value={period}
                    onChange={(e) => setPeriod(e.target.value)}
                    className="px-3 py-2 border rounded-lg focus:outline-none focus:ring-2 focus:ring-blue-500"
                />
            </div>

            {error && (
                <div className="bg-red-50 text-red-600 p-4 rounded-lg">
                    {error}
                </div>
            )}

            {user?.role === 'tm' && summary ? (
                <>
                    {/* Итоговые баллы */}
                    <div className="bg-white p-6 rounded-lg shadow">
                        <h2 className="text-lg font-semibold mb-4">Общий итог</h2>
                        <div className="flex items-center justify-between">
                            <div>
                                <div className="text-4xl font-bold text-blue-600">
                                    {summary.total_points}
                                </div>
                                <div className="text-sm text-gray-500">баллов за месяц</div>
                            </div>
                            {summary.level && (
                                <div className="text-right">
                                    <div className="text-xl font-semibold text-green-600">
                                        {summary.level.name}
                                    </div>
                                    <div className="text-sm text-gray-500">
                                        {summary.level.privileges?.bonus || 'Нет привилегий'}
                                    </div>
                                </div>
                            )}
                        </div>
                    </div>

                    {/* Категории */}
                    {summary.categories.length > 0 ? (
                        <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
                            {summary.categories.map((cat) => (
                                <div key={cat.category_code} className="bg-white p-4 rounded-lg shadow">
                                    <h3 className="font-semibold mb-2">{cat.category_name}</h3>
                                    <div className="space-y-1 text-sm">
                                        <div className="flex justify-between">
                                            <span className="text-gray-600">Базовые:</span>
                                            <span className="font-medium">{cat.base_points}</span>
                                        </div>
                                        <div className="flex justify-between">
                                            <span className="text-gray-600">Дополнительные:</span>
                                            <span className="font-medium text-green-600">{cat.extra_points}</span>
                                        </div>
                                        <div className="flex justify-between">
                                            <span className="text-gray-600">Штрафы:</span>
                                            <span className="font-medium text-red-600">{cat.penalty_points}</span>
                                        </div>
                                        <div className="border-t pt-1 mt-1 flex justify-between font-bold">
                                            <span>Итого:</span>
                                            <span>{cat.total_points}</span>
                                        </div>
                                    </div>
                                </div>
                            ))}
                        </div>
                    ) : (
                        <div className="bg-white p-8 rounded-lg shadow text-center">
                            <p className="text-gray-500">Нет данных за выбранный период</p>
                            <p className="text-sm text-gray-400 mt-2">
                                Данные появятся после ввода результатов экспертом
                            </p>
                        </div>
                    )}
                </>
            ) : (
                <div className="bg-white p-8 rounded-lg shadow text-center">
                    <p className="text-gray-500">
                        {user?.role === 'expert'
                            ? 'Перейдите в раздел "Ввод результатов" для внесения данных'
                            : user?.role === 'coordinator'
                                ? 'Перейдите в раздел "Отчеты" для просмотра сводной информации'
                                : 'Нет данных за выбранный период'}
                    </p>
                </div>
            )}
        </div>
    );
};