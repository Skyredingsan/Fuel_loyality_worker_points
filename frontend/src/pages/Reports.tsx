import React, { useState, useEffect } from 'react';
import { resultsService } from '../services/results';
import { userService } from '../services/users';
import type { User } from '../types';

export const Reports: React.FC = () => {
    const [users, setUsers] = useState<User[]>([]);
    const [results, setResults] = useState<any[]>([]);
    const [loading, setLoading] = useState(true);
    const [period, setPeriod] = useState(() => {
        const now = new Date();
        return `${now.getFullYear()}-${String(now.getMonth() + 1).padStart(2, '0')}`;
    });
    const [expandedRow, setExpandedRow] = useState<number | null>(null);
    const [message, setMessage] = useState<{ type: 'success' | 'error', text: string } | null>(null);

    // Состояния для отклонения
    const [showRejectModal, setShowRejectModal] = useState(false);
    const [rejectReason, setRejectReason] = useState('');
    const [selectedResultId, setSelectedResultId] = useState<number | null>(null);

    useEffect(() => {
        loadData();
    }, [period]);

    const loadData = async () => {
        try {
            setLoading(true);

            // Получаем всех пользователей
            const allUsers = await userService.getAllUsers();
            console.log('All users loaded:', allUsers);
            setUsers(allUsers);

            // Получаем результаты за период
            const resultsData = await resultsService.getAllResults(period);
            console.log('Results loaded:', resultsData);
            setResults(resultsData);

        } catch (error) {
            console.error('Failed to load reports:', error);
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
            setTimeout(() => setMessage(null), 3000);
        } catch (error) {
            console.error('Failed to confirm results:', error);
            setMessage({ type: 'error', text: 'Ошибка при подтверждении' });
        }
    };

    const handleEdit = async (resultId: number) => {
        try {
            // Перенаправляем на страницу ввода результатов с параметром edit
            window.location.href = `/enter-results?edit=${resultId}`;
        } catch (error) {
            console.error('Failed to edit result:', error);
            setMessage({ type: 'error', text: 'Ошибка при редактировании' });
        }
    };

    const handleReject = (resultId: number) => {
        setSelectedResultId(resultId);
        setShowRejectModal(true);
    };

    const confirmReject = async () => {
        if (!selectedResultId || !rejectReason.trim()) {
            setMessage({ type: 'error', text: 'Укажите причину отклонения' });
            return;
        }

        try {
            await resultsService.rejectResults(selectedResultId, rejectReason);
            await loadData();
            setShowRejectModal(false);
            setRejectReason('');
            setSelectedResultId(null);
            setMessage({ type: 'success', text: 'Результаты отклонены' });
            setTimeout(() => setMessage(null), 3000);
        } catch (error) {
            console.error('Failed to reject results:', error);
            setMessage({ type: 'error', text: 'Ошибка при отклонении результатов' });
        }
    };

    // Функция для подсчета баллов по категории
    const getCategoryPoints = (indicators: any[], categoryCode: string) => {
        return indicators
            .filter((ind: any) => ind.indicator?.category_code === categoryCode)
            .reduce((sum: number, ind: any) => sum + (ind.calculated_points || 0), 0);
    };

    // Получаем цвет для баллов
    const getPointsColor = (points: number) => {
        if (points > 500) return 'text-green-600 font-bold';
        if (points > 200) return 'text-blue-600 font-bold';
        if (points > 0) return 'text-yellow-600 font-bold';
        if (points < 0) return 'text-red-600 font-bold';
        return 'text-gray-600';
    };

    // Определяем уровень по баллам
    const getLevel = (points: number) => {
        if (points >= 4320) return { name: 'Стратег Гран-при', color: 'text-purple-600' };
        if (points >= 2160) return { name: 'Тактик Магистрали', color: 'text-blue-600' };
        return { name: 'Специалист Трассы', color: 'text-gray-600' };
    };

    if (loading) {
        return (
            <div className="flex justify-center items-center h-64">
                <div className="animate-spin rounded-full h-12 w-12 border-b-2 border-blue-600"></div>
            </div>
        );
    }

    // Подсчет итогов
    const totals = {
        pm: 0,
        oek: 0,
        ekl: 0,
        kb: 0,
        total: 0,
        count: results.length,
        confirmed: results.filter(r => r.status === 'confirmed').length,
        draft: results.filter(r => r.status === 'draft').length
    };

    results.forEach(result => {
        const indicators = result.indicators || [];
        totals.pm += getCategoryPoints(indicators, 'ПМ');
        totals.oek += getCategoryPoints(indicators, 'ОЭК');
        totals.ekl += getCategoryPoints(indicators, 'ЭКЛ');
        totals.kb += getCategoryPoints(indicators, 'КБ');
        totals.total += indicators.reduce((sum: number, ind: any) => sum + (ind.calculated_points || 0), 0);
    });

    return (
        <div className="space-y-6 p-6">
            <div className="flex justify-between items-center">
                <h1 className="text-2xl font-bold">Отчеты</h1>
                <div className="flex items-center gap-2">
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
            <div className="grid grid-cols-1 md:grid-cols-4 gap-4">
                <div className="bg-white p-4 rounded-lg shadow">
                    <div className="text-sm text-gray-500">Всего результатов</div>
                    <div className="text-2xl font-bold">{totals.count}</div>
                </div>
                <div className="bg-white p-4 rounded-lg shadow">
                    <div className="text-sm text-gray-500">Подтверждено</div>
                    <div className="text-2xl font-bold text-green-600">{totals.confirmed}</div>
                </div>
                <div className="bg-white p-4 rounded-lg shadow">
                    <div className="text-sm text-gray-500">Черновиков</div>
                    <div className="text-2xl font-bold text-yellow-600">{totals.draft}</div>
                </div>
                <div className="bg-white p-4 rounded-lg shadow">
                    <div className="text-sm text-gray-500">Сумма баллов</div>
                    <div className="text-2xl font-bold">{totals.total}</div>
                </div>
            </div>

            {results.length === 0 ? (
                <div className="bg-white p-12 rounded-lg shadow text-center">
                    <div className="text-6xl mb-4">📭</div>
                    <h2 className="text-xl font-semibold text-gray-700 mb-2">Нет данных за выбранный период</h2>
                    <p className="text-gray-500">За период {period} нет результатов для отображения</p>
                </div>
            ) : (
                <div className="bg-white rounded-lg shadow overflow-hidden">
                    <table className="min-w-full divide-y divide-gray-200">
                        <thead className="bg-gray-50">
                        <tr>
                            <th className="px-4 py-3 text-left text-xs font-medium text-gray-500 uppercase">ТМ</th>
                            <th className="px-4 py-3 text-left text-xs font-medium text-gray-500 uppercase">Кластер</th>
                            <th className="px-4 py-3 text-center text-xs font-medium text-gray-500 uppercase">Продажи и маржа</th>
                            <th className="px-4 py-3 text-center text-xs font-medium text-gray-500 uppercase">Операционная эффективность</th>
                            <th className="px-4 py-3 text-center text-xs font-medium text-gray-500 uppercase">Эффективность команды</th>
                            <th className="px-4 py-3 text-center text-xs font-medium text-gray-500 uppercase">Культура безопасности</th>
                            <th className="px-4 py-3 text-center text-xs font-medium text-gray-500 uppercase">ИТОГО</th>
                            <th className="px-4 py-3 text-center text-xs font-medium text-gray-500 uppercase">Уровень</th>
                            <th className="px-4 py-3 text-center text-xs font-medium text-gray-500 uppercase">Статус</th>
                            <th className="px-4 py-3 text-center text-xs font-medium text-gray-500 uppercase">Действия</th>
                        </tr>
                        </thead>
                        <tbody className="bg-white divide-y divide-gray-200">
                        {results.map((result) => {
                            const indicators = result.indicators || [];
                            const pmPoints = getCategoryPoints(indicators, 'ПМ');
                            const oekPoints = getCategoryPoints(indicators, 'ОЭК');
                            const eklPoints = getCategoryPoints(indicators, 'ЭКЛ');
                            const kbPoints = getCategoryPoints(indicators, 'КБ');
                            const total = pmPoints + oekPoints + eklPoints + kbPoints;
                            const level = getLevel(total);

                            return (
                                <React.Fragment key={result.id}>
                                    <tr
                                        className="hover:bg-gray-50 cursor-pointer"
                                        onClick={() => setExpandedRow(expandedRow === result.id ? null : result.id)}
                                    >
                                        <td className="px-4 py-3 font-medium">
                                            {result.user?.fio || `ТМ #${result.user_id}`}
                                        </td>
                                        <td className="px-4 py-3 text-gray-600">
                                            {result.user?.cluster_name || '—'}
                                        </td>
                                        <td className={`px-4 py-3 text-center ${getPointsColor(pmPoints)}`}>{pmPoints}</td>
                                        <td className={`px-4 py-3 text-center ${getPointsColor(oekPoints)}`}>{oekPoints}</td>
                                        <td className={`px-4 py-3 text-center ${getPointsColor(eklPoints)}`}>{eklPoints}</td>
                                        <td className={`px-4 py-3 text-center ${getPointsColor(kbPoints)}`}>{kbPoints}</td>
                                        <td className={`px-4 py-3 text-center font-bold ${getPointsColor(total)}`}>{total}</td>
                                        <td className="px-4 py-3 text-center">
                        <span className={`text-sm font-medium ${level.color}`}>
                          {level.name}
                        </span>
                                        </td>
                                        <td className="px-4 py-3 text-center">
                        <span className={`px-2 py-1 text-xs rounded-full ${
                            result.status === 'confirmed'
                                ? 'bg-green-100 text-green-800'
                                : 'bg-yellow-100 text-yellow-800'
                        }`}>
                          {result.status === 'confirmed' ? 'Подтверждено' : 'Черновик'}
                        </span>
                                        </td>
                                        <td className="px-4 py-3 text-center">
                                            <div className="flex justify-center space-x-2">
                                                {result.status === 'draft' && (
                                                    <>
                                                        <button
                                                            onClick={(e) => {
                                                                e.stopPropagation();
                                                                handleEdit(result.id);
                                                            }}
                                                            className="bg-blue-500 text-white px-3 py-1 rounded-lg text-sm hover:bg-blue-600 transition flex items-center"
                                                            title="Редактировать"
                                                        >
                                                            ✏️
                                                        </button>
                                                        <button
                                                            onClick={(e) => {
                                                                e.stopPropagation();
                                                                handleReject(result.id);
                                                            }}
                                                            className="bg-red-500 text-white px-3 py-1 rounded-lg text-sm hover:bg-red-600 transition flex items-center"
                                                            title="Отклонить"
                                                        >
                                                            ✗
                                                        </button>
                                                        <button
                                                            onClick={(e) => {
                                                                e.stopPropagation();
                                                                handleConfirm(result.id);
                                                            }}
                                                            className="bg-green-500 text-white px-3 py-1 rounded-lg text-sm hover:bg-green-600 transition flex items-center"
                                                            title="Подтвердить"
                                                        >
                                                            ✓
                                                        </button>
                                                    </>
                                                )}
                                                {result.status === 'confirmed' && (
                                                    <span className="text-green-600 text-sm font-medium">✓ Подтверждено</span>
                                                )}
                                            </div>
                                        </td>
                                    </tr>

                                    {/* Детализация при раскрытии */}
                                    {expandedRow === result.id && (
                                        <tr>
                                            <td colSpan={10} className="bg-gray-50 p-4">
                                                <div className="space-y-2">
                                                    <h3 className="font-semibold mb-2">Детализация показателей:</h3>
                                                    <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-2">
                                                        {indicators.map((ind: any) => (
                                                            <div key={ind.id} className="bg-white p-2 rounded border text-sm">
                                                                <div className="font-medium">{ind.indicator?.name}</div>
                                                                <div className="text-xs text-gray-500">{ind.indicator?.code}</div>
                                                                <div className="flex justify-between mt-1">
                                                                    <span>Факт: {ind.fact_value} {ind.indicator?.unit}</span>
                                                                    <span className={getPointsColor(ind.calculated_points)}>
                                      {ind.calculated_points} баллов
                                    </span>
                                                                </div>
                                                                {ind.supporting_document_url && (
                                                                    <a
                                                                        href={`http://localhost:8080${ind.supporting_document_url}`}
                                                                        target="_blank"
                                                                        rel="noopener noreferrer"
                                                                        className="text-xs text-blue-600 hover:underline mt-1 inline-block"
                                                                        onClick={(e) => e.stopPropagation()}
                                                                    >
                                                                        📎 Документ
                                                                    </a>
                                                                )}
                                                            </div>
                                                        ))}
                                                    </div>
                                                </div>
                                            </td>
                                        </tr>
                                    )}
                                </React.Fragment>
                            );
                        })}
                        </tbody>
                        <tfoot className="bg-gray-100 font-medium">
                        <tr>
                            <td colSpan={2} className="px-4 py-3 text-right">Среднее:</td>
                            <td className="px-4 py-3 text-center">
                                {totals.count > 0 ? Math.round(totals.pm / totals.count) : 0}
                            </td>
                            <td className="px-4 py-3 text-center">
                                {totals.count > 0 ? Math.round(totals.oek / totals.count) : 0}
                            </td>
                            <td className="px-4 py-3 text-center">
                                {totals.count > 0 ? Math.round(totals.ekl / totals.count) : 0}
                            </td>
                            <td className="px-4 py-3 text-center">
                                {totals.count > 0 ? Math.round(totals.kb / totals.count) : 0}
                            </td>
                            <td className="px-4 py-3 text-center font-bold">
                                {totals.count > 0 ? Math.round(totals.total / totals.count) : 0}
                            </td>
                            <td colSpan={3}></td>
                        </tr>
                        </tfoot>
                    </table>
                </div>
            )}

            {/* Модальное окно отклонения */}
            {showRejectModal && (
                <div className="fixed inset-0 bg-black bg-opacity-50 flex items-center justify-center z-50">
                    <div className="bg-white rounded-lg shadow-xl max-w-md w-full mx-4">
                        <div className="p-6">
                            <h2 className="text-xl font-bold mb-4">Отклонение результатов</h2>
                            <p className="text-sm text-gray-600 mb-4">
                                Укажите причину отклонения результатов. Эта информация будет видна эксперту.
                            </p>

                            <textarea
                                value={rejectReason}
                                onChange={(e) => setRejectReason(e.target.value)}
                                placeholder="Причина отклонения..."
                                rows={4}
                                className="w-full px-3 py-2 border border-gray-300 rounded-lg focus:outline-none focus:ring-2 focus:ring-red-500 mb-4"
                                autoFocus
                            />

                            <div className="flex justify-end space-x-3">
                                <button
                                    onClick={() => {
                                        setShowRejectModal(false);
                                        setRejectReason('');
                                        setSelectedResultId(null);
                                    }}
                                    className="px-4 py-2 border border-gray-300 rounded-lg hover:bg-gray-50 transition"
                                >
                                    Отмена
                                </button>
                                <button
                                    onClick={confirmReject}
                                    disabled={!rejectReason.trim()}
                                    className="px-4 py-2 bg-red-600 text-white rounded-lg hover:bg-red-700 transition disabled:opacity-50"
                                >
                                    Отклонить
                                </button>
                            </div>
                        </div>
                    </div>
                </div>
            )}
        </div>
    );
};