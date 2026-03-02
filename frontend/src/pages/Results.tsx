import React, { useState, useEffect } from 'react';
import { useAuth } from '../contexts/AuthContext';
import { resultsService } from '../services/results';
import { kpiService } from '../services/kpi';
import { uploadService } from '../services/upload';
import type { FullResultSummary, CategorySummary, KPIIndicator, IndicatorResult } from '../types';

export const Results: React.FC = () => {
    const { user } = useAuth();
    const [summary, setSummary] = useState<FullResultSummary | null>(null);
    const [indicators, setIndicators] = useState<KPIIndicator[]>([]);
    const [loading, setLoading] = useState(true);
    const [period, setPeriod] = useState(() => {
        const now = new Date();
        return `${now.getFullYear()}-${String(now.getMonth() + 1).padStart(2, '0')}`;
    });
    const [expandedCategory, setExpandedCategory] = useState<string | null>(null);

    useEffect(() => {
        loadData();
    }, [period]);

    const loadData = async () => {
        try {
            setLoading(true);

            // Получаем сводку за период с детализацией
            const summaryData = await resultsService.getMyResults(period);
            console.log('Real data with details:', summaryData);
            setSummary(summaryData);

            // Получаем все индикаторы для справки
            const indicatorsData = await kpiService.getAllIndicators();
            setIndicators(indicatorsData);

        } catch (error) {
            console.error('Failed to load results:', error);
        } finally {
            setLoading(false);
        }
    };

    // Группируем индикаторы по категориям
    const getIndicatorsByCategory = (categoryCode: string) => {
        return indicators.filter(ind => ind.category_code === categoryCode);
    };

    // Получаем детальные результаты для категории
    const getDetailedResultsForCategory = (categoryCode: string): IndicatorResult[] => {
        if (!summary?.detailed_results) return [];

        const categoryIndicators = getIndicatorsByCategory(categoryCode);
        const categoryIndicatorIds = new Set(categoryIndicators.map(ind => ind.id));

        return summary.detailed_results.filter(result =>
            categoryIndicatorIds.has(result.indicator_id)
        );
    };

    // Получаем цвет для категории
    const getCategoryColor = (categoryCode: string) => {
        switch (categoryCode) {
            case 'ПМ': return 'border-blue-500 bg-blue-50';
            case 'ОЭК': return 'border-green-500 bg-green-50';
            case 'ЭКЛ': return 'border-purple-500 bg-purple-50';
            case 'КБ': return 'border-orange-500 bg-orange-50';
            default: return 'border-gray-500 bg-gray-50';
        }
    };

    // Получаем иконку для категории
    const getCategoryIcon = (categoryCode: string) => {
        switch (categoryCode) {
            case 'ПМ': return '📊';
            case 'ОЭК': return '⚙️';
            case 'ЭКЛ': return '👥';
            case 'КБ': return '🔒';
            default: return '📋';
        }
    };

    // Получаем цвет для баллов
    const getPointsColor = (points: number) => {
        if (points > 0) return 'text-green-600 font-bold';
        if (points < 0) return 'text-red-600 font-bold';
        return 'text-gray-600';
    };

    // Форматирование значения
    const formatValue = (value: number | undefined | null, unit: string) => {
        if (value === undefined || value === null) return '—';

        const rounded = Math.round(value * 100) / 100;

        switch (unit) {
            case '%': return `${rounded}%`;
            case 'шт': return `${rounded} шт`;
            case 'чел': return `${rounded} чел`;
            default: return rounded.toString();
        }
    };

    // Получаем название типа показателя
    const getIndicatorTypeName = (type: string) => {
        switch (type) {
            case 'base': return 'Базовый';
            case 'extra': return 'Дополнительный';
            case 'penalty': return 'Штрафной';
            default: return type;
        }
    };

    // Получаем цвет для типа
    const getTypeColor = (type: string) => {
        switch (type) {
            case 'base': return 'bg-blue-100 text-blue-700';
            case 'extra': return 'bg-green-100 text-green-700';
            case 'penalty': return 'bg-red-100 text-red-700';
            default: return 'bg-gray-100 text-gray-700';
        }
    };

    // Находим категорию по коду
    const getCategoryData = (categoryCode: string): CategorySummary | undefined => {
        return summary?.categories.find(c => c.category_code === categoryCode);
    };

    if (loading) {
        return (
            <div className="flex justify-center items-center h-64">
                <div className="animate-spin rounded-full h-12 w-12 border-b-2 border-blue-600"></div>
            </div>
        );
    }

    if (!summary || summary.categories.length === 0) {
        return (
            <div className="bg-white p-12 rounded-lg shadow text-center">
                <div className="text-6xl mb-4">📭</div>
                <h2 className="text-xl font-semibold text-gray-700 mb-2">Нет данных за выбранный период</h2>
                <p className="text-gray-500">
                    Результаты появятся после того, как эксперт введет данные
                </p>
            </div>
        );
    }

    return (
        <div className="space-y-6">
            {/* Заголовок и выбор периода */}
            <div className="flex flex-col sm:flex-row justify-between items-start sm:items-center gap-4">
                <h1 className="text-2xl font-bold">Мои результаты</h1>
                <input
                    type="month"
                    value={period}
                    onChange={(e) => setPeriod(e.target.value)}
                    className="px-3 py-2 border rounded-lg focus:outline-none focus:ring-2 focus:ring-blue-500"
                />
            </div>

            {/* Сводная карточка с итогами */}
            <div className="bg-gradient-to-r from-blue-600 to-blue-700 text-white p-6 rounded-lg shadow-lg">
                <div className="flex flex-col md:flex-row justify-between items-start md:items-center gap-4">
                    <div>
                        <p className="text-blue-100 text-sm">Всего баллов за месяц</p>
                        <p className="text-5xl font-bold">{summary.total_points}</p>
                    </div>
                    {summary.level && (
                        <div className="bg-white/20 backdrop-blur-sm px-6 py-3 rounded-lg">
                            <p className="text-sm opacity-90">Текущий уровень</p>
                            <p className="text-2xl font-semibold">{summary.level.name}</p>
                            {summary.level.privileges && (
                                <p className="text-sm mt-1 opacity-90">
                                    🏆 {summary.level.privileges.bonus || 'Нет привилегий'}
                                </p>
                            )}
                        </div>
                    )}
                </div>
            </div>

            {/* Детальные результаты по категориям */}
            <div className="space-y-4">
                {['ПМ', 'ОЭК', 'ЭКЛ', 'КБ'].map((categoryCode) => {
                    const category = getCategoryData(categoryCode);
                    if (!category) return null;

                    const detailedResults = getDetailedResultsForCategory(categoryCode);
                    const hasDetails = detailedResults.length > 0;

                    return (
                        <div key={categoryCode} className="bg-white rounded-lg shadow overflow-hidden">
                            {/* Заголовок категории */}
                            <button
                                onClick={() => setExpandedCategory(
                                    expandedCategory === categoryCode ? null : categoryCode
                                )}
                                className={`w-full p-4 flex items-center justify-between border-l-4 ${getCategoryColor(categoryCode)} hover:bg-gray-50 transition`}
                            >
                                <div className="flex items-center space-x-3">
                                    <span className="text-2xl">{getCategoryIcon(categoryCode)}</span>
                                    <div className="text-left">
                                        <h2 className="text-lg font-semibold">
                                            {categoryCode === 'ПМ' && 'Продажи и маржа'}
                                            {categoryCode === 'ОЭК' && 'Операционная эффективность'}
                                            {categoryCode === 'ЭКЛ' && 'Эффективность команды'}
                                            {categoryCode === 'КБ' && 'Культура безопасности'}
                                        </h2>
                                        <p className="text-sm text-gray-600">
                                            Базовые: {category.base_points} •
                                            Доп: {category.extra_points} •
                                            Штрафы: {category.penalty_points}
                                        </p>
                                    </div>
                                </div>
                                <div className="flex items-center space-x-4">
                  <span className={`text-xl font-bold ${getPointsColor(category.total_points)}`}>
                    {category.total_points} баллов
                  </span>
                                    <span className="text-gray-400">
                    {expandedCategory === categoryCode ? '▼' : '▶'}
                  </span>
                                </div>
                            </button>

                            {/* Детализация показателей */}
                            {expandedCategory === categoryCode && (
                                <div className="p-4 bg-gray-50 border-t">
                                    {hasDetails ? (
                                        <div className="overflow-x-auto">
                                            <table className="w-full text-sm">
                                                <thead className="text-xs text-gray-500 uppercase">
                                                <tr>
                                                    <th className="px-3 py-2 text-left">Показатель</th>
                                                    <th className="px-3 py-2 text-center">Тип</th>
                                                    <th className="px-3 py-2 text-center">Факт</th>
                                                    <th className="px-3 py-2 text-center">Баллы</th>
                                                    <th className="px-3 py-2 text-center">Документ</th>
                                                </tr>
                                                </thead>
                                                <tbody className="divide-y divide-gray-200">
                                                {detailedResults.map((result) => {
                                                    const indicator = indicators.find(i => i.id === result.indicator_id);
                                                    if (!indicator) return null;

                                                    return (
                                                        <tr key={result.id} className="hover:bg-white transition">
                                                            <td className="px-3 py-2">
                                                                <div className="font-medium">{indicator.name}</div>
                                                                <div className="text-xs text-gray-400">{indicator.code}</div>
                                                            </td>
                                                            <td className="px-3 py-2 text-center">
                                  <span className={`px-2 py-1 rounded-full text-xs ${getTypeColor(indicator.indicator_type)}`}>
                                    {getIndicatorTypeName(indicator.indicator_type)}
                                  </span>
                                                            </td>
                                                            <td className="px-3 py-2 text-center font-mono">
                                                                {formatValue(result.fact_value, indicator.unit)}
                                                            </td>
                                                            <td className={`px-3 py-2 text-right font-mono ${getPointsColor(result.calculated_points)}`}>
                                                                {result.calculated_points}
                                                            </td>
                                                            <td className="px-3 py-2 text-center">
                                                                {result.supporting_document_url ? (
                                                                    <a
                                                                        href={uploadService.getFileUrl(result.supporting_document_url)}
                                                                        target="_blank"
                                                                        rel="noopener noreferrer"
                                                                        className="text-blue-600 hover:text-blue-800 inline-flex items-center"
                                                                        title="Открыть подтверждающий документ"
                                                                        onClick={(e) => e.stopPropagation()}
                                                                    >
                                                                        <span className="text-xl">📎</span>
                                                                    </a>
                                                                ) : (
                                                                    <span className="text-gray-300">—</span>
                                                                )}
                                                            </td>
                                                        </tr>
                                                    );
                                                })}
                                                </tbody>
                                                <tfoot className="bg-gray-100 font-medium">
                                                <tr>
                                                    <td colSpan={4} className="px-3 py-2 text-right">Итого по категории:</td>
                                                    <td className={`px-3 py-2 text-right font-bold ${getPointsColor(category.total_points)}`}>
                                                        {category.total_points}
                                                    </td>
                                                </tr>
                                                </tfoot>
                                            </table>
                                        </div>
                                    ) : (
                                        <p className="text-center text-gray-500 py-4">
                                            Детальные результаты по показателям отсутствуют
                                        </p>
                                    )}
                                </div>
                            )}
                        </div>
                    );
                })}
            </div>

            {/* Информация о расчете */}
            <div className="bg-blue-50 p-4 rounded-lg shadow border border-blue-200">
                <h3 className="font-semibold mb-2 flex items-center text-blue-800">
                    <span className="mr-2">ℹ️</span>
                    Как рассчитываются баллы
                </h3>
                <ul className="text-sm text-blue-700 space-y-1">
                    <li>• Базовые: вес × 1 при достижении цели, иначе 0</li>
                    <li>• Дополнительные: факт × вес (за каждый % или штуку)</li>
                    <li>• Штрафные: факт × вес (отрицательные баллы)</li>
                    <li>• Итоговый уровень зависит от суммы баллов за год</li>
                </ul>
            </div>
        </div>
    );
};