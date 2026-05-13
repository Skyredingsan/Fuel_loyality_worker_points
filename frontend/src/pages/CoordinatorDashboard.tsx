import React, { useState, useEffect } from 'react';
import { useAuth } from '../contexts/AuthContext';
import { resultsService } from '../services/results';
import { userService } from '../services/users';
import { Link } from 'react-router-dom';

export const CoordinatorDashboard: React.FC = () => {
    const { user } = useAuth();
    console.log(user);
    const [stats, setStats] = useState({
        totalUsers: 0,
        totalTMs: 0,
        totalExperts: 0,
        tmsWithResults: 0,
        totalPoints: 0,
        averagePoints: 0,
        maxPoints: 0,
        pendingResults: 0
    });
    const [recentResults, setRecentResults] = useState<any[]>([]);
    const [loading, setLoading] = useState(true);
    const [message, setMessage] = useState<{ type: 'success' | 'error', text: string } | null>(null);
    const [period, setPeriod] = useState(() => {
        const now = new Date();
        return `${now.getFullYear()}-${String(now.getMonth() + 1).padStart(2, '0')}`;
    });

    useEffect(() => {
        loadData();
    }, [period]);

    const loadData = async () => {
        try {
            setLoading(true);

            // Получаем всех пользователей
            const users = await userService.getAllUsers();
            const tms = users.filter(u => u.role === 'tm');
            const experts = users.filter(u => u.role === 'expert');

            // Получаем результаты за период
            const results = await resultsService.getAllResults(period);
            console.log('Results loaded:', results); // Для отладки

            // Считаем статистику
            const tmsWithResults = new Set(results.map((r: any) => r.user_id)).size;
            const totalPoints = results.reduce((sum: number, r: any) => {
                // Суммируем баллы из детальных результатов
                const points = r.indicators?.reduce((acc: number, ind: any) =>
                    acc + (ind.calculated_points || 0), 0) || 0;
                return sum + points;
            }, 0);

            const maxPoints = results.reduce((max: number, r: any) => {
                const points = r.indicators?.reduce((acc: number, ind: any) =>
                    acc + (ind.calculated_points || 0), 0) || 0;
                return Math.max(max, points);
            }, 0);

            const pendingCount = results.filter((r: any) => r.status === 'draft').length;

            setStats({
                totalUsers: users.length,
                totalTMs: tms.length,
                totalExperts: experts.length,
                tmsWithResults,
                totalPoints,
                averagePoints: results.length ? Math.round(totalPoints / results.length) : 0,
                maxPoints,
                pendingResults: pendingCount
            });

            // Берем последние 5 результатов (сортировка по дате)
            const sorted = [...results].sort((a, b) =>
                new Date(b.created_at).getTime() - new Date(a.created_at).getTime()
            );
            setRecentResults(sorted.slice(0, 5));

        } catch (error) {
            console.error('Failed to load dashboard data:', error);
            setMessage({ type: 'error', text: 'Ошибка загрузки данных' });
        } finally {
            setLoading(false);
        }
    };

    const handleConfirm = async (resultId: number) => {
        if (!window.confirm('Подтвердить результаты? Это действие нельзя отменить.')) {
            return;
        }

        try {
            await resultsService.confirmResults(resultId);
            await loadData();
            setMessage({ type: 'success', text: 'Результаты подтверждены' });

            // Скрываем сообщение через 3 секунды
            setTimeout(() => setMessage(null), 3000);
        } catch (error) {
            console.error('Failed to confirm results:', error);
            setMessage({ type: 'error', text: 'Ошибка при подтверждении' });
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
                <h1 className="text-2xl font-bold">Панель координатора</h1>
                <div className="flex items-center space-x-2">
                    <span className="text-sm text-gray-500">Период:</span>
                    <input
                        type="month"
                        value={period}
                        onChange={(e) => setPeriod(e.target.value)}
                        className="px-3 py-2 border rounded-lg focus:outline-none focus:ring-2 focus:ring-blue-500"
                    />
                </div>
            </div>

            {message && (
                <div className={`p-4 rounded-lg ${
                    message.type === 'success' ? 'bg-green-50 text-green-800' : 'bg-red-50 text-red-800'
                }`}>
                    {message.text}
                </div>
            )}

            {/* Статистика */}
            <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-4 gap-4">
                <div className="bg-white p-6 rounded-lg shadow">
                    <div className="flex items-center justify-between">
                        <div>
                            <p className="text-sm text-gray-500">Всего пользователей</p>
                            <p className="text-3xl font-bold">{stats.totalUsers}</p>
                        </div>
                        <div className="text-3xl text-blue-500">👥</div>
                    </div>
                    <div className="mt-2 text-sm text-gray-600">
                        ТМ: {stats.totalTMs} | Экспертов: {stats.totalExperts}
                    </div>
                </div>

                <div className="bg-white p-6 rounded-lg shadow">
                    <div className="flex items-center justify-between">
                        <div>
                            <p className="text-sm text-gray-500">ТМ с результатами</p>
                            <p className="text-3xl font-bold">{stats.tmsWithResults}</p>
                        </div>
                        <div className="text-3xl text-green-500">📊</div>
                    </div>
                    <div className="mt-2 text-sm text-gray-600">
                        из {stats.totalTMs} ТМ
                    </div>
                </div>

                <div className="bg-white p-6 rounded-lg shadow">
                    <div className="flex items-center justify-between">
                        <div>
                            <p className="text-sm text-gray-500">Сумма баллов</p>
                            <p className="text-3xl font-bold">{stats.totalPoints}</p>
                        </div>
                        <div className="text-3xl text-purple-500">⭐</div>
                    </div>
                    <div className="mt-2 text-sm text-gray-600">
                        Средний: {stats.averagePoints} | Макс: {stats.maxPoints}
                    </div>
                </div>

                <div className="bg-white p-6 rounded-lg shadow">
                    <div className="flex items-center justify-between">
                        <div>
                            <p className="text-sm text-gray-500">Ожидают подтверждения</p>
                            <p className="text-3xl font-bold">{stats.pendingResults}</p>
                        </div>
                        <div className="text-3xl text-yellow-500">⏳</div>
                    </div>
                    <div className="mt-2 text-sm text-gray-600">
                        <Link to="/reports" className="text-blue-600 hover:underline">
                            Перейти к отчетам →
                        </Link>
                    </div>
                </div>
            </div>

            {/* Быстрые действия */}
            <div className="grid grid-cols-1 md:grid-cols-3 gap-4">
                <Link
                    to="/users"
                    className="bg-white p-6 rounded-lg shadow hover:shadow-lg transition flex items-center space-x-4"
                >
                    <div className="text-3xl">👥</div>
                    <div>
                        <h3 className="font-semibold">Управление пользователями</h3>
                        <p className="text-sm text-gray-500">Добавление, редактирование, удаление</p>
                    </div>
                </Link>

                <Link
                    to="/kpi"
                    className="bg-white p-6 rounded-lg shadow hover:shadow-lg transition flex items-center space-x-4"
                >
                    <div className="text-3xl">🎯</div>
                    <div>
                        <h3 className="font-semibold">Настройка KPI</h3>
                        <p className="text-sm text-gray-500">Редактирование весов и целей</p>
                    </div>
                </Link>

                <Link
                    to="/reports"
                    className="bg-white p-6 rounded-lg shadow hover:shadow-lg transition flex items-center space-x-4"
                >
                    <div className="text-3xl">📊</div>
                    <div>
                        <h3 className="font-semibold">Отчеты</h3>
                        <p className="text-sm text-gray-500">Просмотр, подтверждение и экспорт</p>
                    </div>
                </Link>
            </div>

            {/* Последние результаты */}
            <div className="bg-white rounded-lg shadow">
                <div className="px-6 py-4 border-b flex justify-between items-center">
                    <h2 className="font-semibold">Последние результаты</h2>
                    <Link to="/reports" className="text-sm text-blue-600 hover:underline">
                        Все отчеты →
                    </Link>
                </div>
                <div className="p-6">
                    {recentResults.length > 0 ? (
                        <table className="w-full">
                            <thead>
                            <tr className="text-xs text-gray-500 uppercase">
                                <th className="text-left pb-2">ТМ</th>
                                <th className="text-left pb-2">Эксперт</th>
                                <th className="text-center pb-2">Баллы</th>
                                <th className="text-center pb-2">Статус</th>
                                <th className="text-center pb-2">Действия</th>
                                <th className="text-right pb-2">Дата</th>
                            </tr>
                            </thead>
                            <tbody className="divide-y">
                            {recentResults.map((result: any) => {
                                // Считаем общие баллы
                                const totalPoints = result.indicators?.reduce(
                                    (sum: number, ind: any) => sum + (ind.calculated_points || 0), 0
                                ) || 0;

                                return (
                                    <tr key={result.id}>
                                        <td className="py-2">{result.user?.fio || '—'}</td>
                                        <td className="py-2">{result.expert?.fio || '—'}</td>
                                        <td className="py-2 text-center font-bold">{totalPoints}</td>
                                        <td className="py-2 text-center">
                        <span className={`px-2 py-1 text-xs rounded-full ${
                            result.status === 'confirmed'
                                ? 'bg-green-100 text-green-800'
                                : 'bg-yellow-100 text-yellow-800'
                        }`}>
                          {result.status === 'confirmed' ? 'Подтверждено' : 'Черновик'}
                        </span>
                                        </td>
                                        <td className="py-2 text-center">
                                            {result.status !== 'confirmed' && (
                                                <button
                                                    onClick={() => handleConfirm(result.id)}
                                                    className="bg-green-500 text-white px-3 py-1 rounded-lg text-xs hover:bg-green-600 transition"
                                                >
                                                    Подтвердить
                                                </button>
                                            )}
                                        </td>
                                        <td className="py-2 text-right text-sm text-gray-500">
                                            {new Date(result.created_at).toLocaleDateString()}
                                        </td>
                                    </tr>
                                );
                            })}
                            </tbody>
                        </table>
                    ) : (
                        <p className="text-center text-gray-500 py-4">
                            Нет результатов за выбранный период
                        </p>
                    )}
                </div>
            </div>
        </div>
    );
};