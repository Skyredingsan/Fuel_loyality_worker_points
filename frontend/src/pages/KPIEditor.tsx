import React, { useState, useEffect } from 'react';
import { kpiService } from '../services/kpi';
import type { KPIIndicator, KPICategory } from '../types';

export const KPIEditor: React.FC = () => {
    const [indicators, setIndicators] = useState<KPIIndicator[]>([]);
    const [categories, setCategories] = useState<KPICategory[]>([]);
    const [loading, setLoading] = useState(true);
    const [editingId, setEditingId] = useState<number | null>(null);
    const [showModal, setShowModal] = useState(false);
    const [selectedCategory, setSelectedCategory] = useState<string>('all');
    const [searchTerm, setSearchTerm] = useState('');
    const [message, setMessage] = useState<{ type: 'success' | 'error'; text: string } | null>(null);

    const [formData, setFormData] = useState({
        code: '',
        name: '',
        category_code: 'ПМ',
        description: '',
        unit: '%',
        indicator_type: 'base' as 'base' | 'extra' | 'penalty',
        base_value: '',
        base_weight: '',
        extra_weight: '',
        penalty_weight: ''
    });

    useEffect(() => {
        loadData();
    }, []);

    const loadData = async () => {
        try {
            setLoading(true);
            const [indicatorsData, categoriesData] = await Promise.all([
                kpiService.getAllIndicators(),
                kpiService.getCategories()
            ]);
            setIndicators(Array.isArray(indicatorsData) ? indicatorsData : []);
            setCategories(Array.isArray(categoriesData) ? categoriesData : []);
        } catch (error) {
            console.error('Failed to load KPI data:', error);
            setMessage({ type: 'error', text: 'Ошибка загрузки данных' });
        } finally {
            setLoading(false);
        }
    };

    const handleInputChange = (e: React.ChangeEvent<HTMLInputElement | HTMLSelectElement | HTMLTextAreaElement>) => {
        const { name, value } = e.target;
        setFormData(prev => ({ ...prev, [name]: value }));
    };

    const resetForm = () => {
        setFormData({
            code: '',
            name: '',
            category_code: 'ПМ',
            description: '',
            unit: '%',
            indicator_type: 'base',
            base_value: '',
            base_weight: '',
            extra_weight: '',
            penalty_weight: ''
        });
        setEditingId(null);
    };

    const handleEdit = (indicator: KPIIndicator) => {
        setFormData({
            code: indicator.code,
            name: indicator.name,
            category_code: indicator.category_code || 'ПМ',
            description: indicator.description || '',
            unit: indicator.unit,
            indicator_type: indicator.indicator_type,
            base_value: indicator.base_value?.toString() || '',
            base_weight: indicator.base_weight?.toString() || '',
            extra_weight: indicator.extra_weight?.toString() || '',
            penalty_weight: indicator.penalty_weight?.toString() || ''
        });
        setEditingId(indicator.id);
        setShowModal(true);
    };

    const handleSubmit = async (e: React.FormEvent) => {
        e.preventDefault();

        try {
            const data: any = {
                code: formData.code,
                name: formData.name,
                category_code: formData.category_code,
                description: formData.description,
                unit: formData.unit,
                indicator_type: formData.indicator_type,
            };

            if (formData.indicator_type === 'base') {
                data.base_value = parseFloat(formData.base_value) || 0;
                data.base_weight = parseInt(formData.base_weight) || 0;
            } else if (formData.indicator_type === 'extra') {
                data.extra_weight = parseInt(formData.extra_weight) || 0;
            } else if (formData.indicator_type === 'penalty') {
                data.penalty_weight = parseInt(formData.penalty_weight) || 0;
            }

            if (editingId) {
                await kpiService.updateIndicator(editingId, data);
                setMessage({ type: 'success', text: 'Показатель обновлён' });
            } else {
                await kpiService.createIndicator(data);
                setMessage({ type: 'success', text: 'Показатель создан' });
            }

            await loadData();
            setShowModal(false);
            resetForm();
            setTimeout(() => setMessage(null), 3000);
        } catch (error) {
            console.error('Failed to save indicator:', error);
            setMessage({ type: 'error', text: 'Ошибка при сохранении показателя' });
        }
    };

    const handleDelete = async (id: number, name: string) => {
        if (window.confirm(`Вы уверены, что хотите удалить показатель "${name}"?`)) {
            try {
                await kpiService.deleteIndicator(id);
                await loadData();
                setMessage({ type: 'success', text: 'Показатель удалён' });
                setTimeout(() => setMessage(null), 3000);
            } catch (error) {
                console.error('Failed to delete indicator:', error);
                setMessage({ type: 'error', text: 'Ошибка при удалении показателя' });
            }
        }
    };

    const filteredIndicators = (Array.isArray(indicators) ? indicators : []).filter(ind => {
        const matchesCategory = selectedCategory === 'all' || ind.category_code === selectedCategory;
        const matchesSearch =
            ind.name.toLowerCase().includes(searchTerm.toLowerCase()) ||
            ind.code.toLowerCase().includes(searchTerm.toLowerCase());
        return matchesCategory && matchesSearch;
    });

    const indicatorsByCategory = filteredIndicators.reduce((acc, ind) => {
        const cat = ind.category_code || 'Другое';
        if (!acc[cat]) acc[cat] = [];
        acc[cat].push(ind);
        return acc;
    }, {} as Record<string, KPIIndicator[]>);

    const getTypeBadgeColor = (type: string) => {
        switch (type) {
            case 'base': return 'bg-blue-100 text-blue-800';
            case 'extra': return 'bg-green-100 text-green-800';
            case 'penalty': return 'bg-red-100 text-red-800';
            default: return 'bg-gray-100 text-gray-800';
        }
    };

    const getTypeName = (type: string) => {
        switch (type) {
            case 'base': return 'Базовый';
            case 'extra': return 'Дополнительный';
            case 'penalty': return 'Штрафной';
            default: return type;
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
            <div className="flex flex-col sm:flex-row justify-between items-start sm:items-center gap-4">
                <h1 className="text-2xl font-bold">Управление KPI</h1>
                <button
                    onClick={() => {
                        resetForm();
                        setShowModal(true);
                    }}
                    className="bg-blue-600 text-white px-4 py-2 rounded-lg hover:bg-blue-700 transition flex items-center"
                >
                    <span className="mr-2">➕</span>
                    Новый показатель
                </button>
            </div>

            {message && (
                <div className={`p-4 rounded-lg ${
                    message.type === 'success' ? 'bg-green-50 text-green-800' : 'bg-red-50 text-red-800'
                }`}>
                    {message.text}
                </div>
            )}

            <div className="bg-white p-4 rounded-lg shadow flex flex-col sm:flex-row gap-4">
                <div className="flex-1">
                    <input
                        type="text"
                        placeholder="Поиск по названию или коду..."
                        value={searchTerm}
                        onChange={(e) => setSearchTerm(e.target.value)}
                        className="w-full px-3 py-2 border border-gray-300 rounded-lg focus:outline-none focus:ring-2 focus:ring-blue-500"
                    />
                </div>
                <div>
                    <select
                        value={selectedCategory}
                        onChange={(e) => setSelectedCategory(e.target.value)}
                        className="px-3 py-2 border border-gray-300 rounded-lg focus:outline-none focus:ring-2 focus:ring-blue-500"
                    >
                        <option value="all">Все категории</option>
                        {(Array.isArray(categories) ? categories : []).map(cat => (
                            <option key={cat.id} value={cat.code}>{cat.name}</option>
                        ))}
                    </select>
                </div>
            </div>

            <div className="space-y-6">
                {Object.entries(indicatorsByCategory).map(([categoryCode, cats]) => (
                    <div key={categoryCode} className="bg-white rounded-lg shadow overflow-hidden">
                        <div className="bg-gray-50 px-4 py-3 border-b">
                            <h2 className="font-semibold text-lg">
                                {(Array.isArray(categories) ? categories : []).find(c => c.code === categoryCode)?.name || categoryCode}
                            </h2>
                        </div>
                        <div className="overflow-x-auto">
                            <table className="min-w-full divide-y divide-gray-200">
                                <thead className="bg-gray-50">
                                <tr>
                                    <th className="px-4 py-3 text-left text-xs font-medium text-gray-500 uppercase">Код</th>
                                    <th className="px-4 py-3 text-left text-xs font-medium text-gray-500 uppercase">Название</th>
                                    <th className="px-4 py-3 text-center text-xs font-medium text-gray-500 uppercase">Тип</th>
                                    <th className="px-4 py-3 text-center text-xs font-medium text-gray-500 uppercase">Ед.изм</th>
                                    <th className="px-4 py-3 text-center text-xs font-medium text-gray-500 uppercase">Цель/Вес</th>
                                    <th className="px-4 py-3 text-right text-xs font-medium text-gray-500 uppercase">Действия</th>
                                </tr>
                                </thead>
                                <tbody className="bg-white divide-y divide-gray-200">
                                {cats.map((indicator) => (
                                    <tr key={indicator.id} className="hover:bg-gray-50">
                                        <td className="px-4 py-3 font-mono text-sm">{indicator.code}</td>
                                        <td className="px-4 py-3">
                                            <div className="font-medium">{indicator.name}</div>
                                            {indicator.description && (
                                                <div className="text-xs text-gray-500">{indicator.description}</div>
                                            )}
                                        </td>
                                        <td className="px-4 py-3 text-center">
                        <span className={`px-2 py-1 text-xs font-medium rounded-full ${getTypeBadgeColor(indicator.indicator_type)}`}>
                          {getTypeName(indicator.indicator_type)}
                        </span>
                                        </td>
                                        <td className="px-4 py-3 text-center text-sm">{indicator.unit}</td>
                                        <td className="px-4 py-3 text-center text-sm">
                                            {indicator.indicator_type === 'base' && (
                                                <span>≥{indicator.base_value}{indicator.unit} (вес: {indicator.base_weight})</span>
                                            )}
                                            {indicator.indicator_type === 'extra' && (
                                                <span className="text-green-600">+{indicator.extra_weight} за ед.</span>
                                            )}
                                            {indicator.indicator_type === 'penalty' && (
                                                <span className="text-red-600">{indicator.penalty_weight} за ед.</span>
                                            )}
                                        </td>
                                        <td className="px-4 py-3 text-right space-x-2">
                                            <button
                                                onClick={() => handleEdit(indicator)}
                                                className="text-blue-600 hover:text-blue-900"
                                                title="Редактировать"
                                            >
                                                ✏️
                                            </button>
                                            <button
                                                onClick={() => handleDelete(indicator.id, indicator.name)}
                                                className="text-red-600 hover:text-red-900"
                                                title="Удалить"
                                            >
                                                🗑️
                                            </button>
                                        </td>
                                    </tr>
                                ))}
                                </tbody>
                            </table>
                        </div>
                    </div>
                ))}
            </div>

            {/* Модальное окно */}
            {showModal && (
                <div className="fixed inset-0 bg-black bg-opacity-50 flex items-center justify-center z-50">
                    <div className="bg-white rounded-lg shadow-xl max-w-2xl w-full mx-4 max-h-[90vh] overflow-y-auto">
                        <div className="p-6">
                            <h2 className="text-xl font-bold mb-4">
                                {editingId ? 'Редактировать показатель' : 'Новый показатель'}
                            </h2>

                            <form onSubmit={handleSubmit} className="space-y-4">
                                <div className="grid grid-cols-2 gap-4">
                                    <div>
                                        <label className="block text-sm font-medium text-gray-700 mb-1">Код *</label>
                                        <input
                                            type="text"
                                            name="code"
                                            value={formData.code}
                                            onChange={handleInputChange}
                                            required
                                            className="w-full px-3 py-2 border border-gray-300 rounded-lg focus:outline-none focus:ring-2 focus:ring-blue-500"
                                        />
                                    </div>
                                    <div>
                                        <label className="block text-sm font-medium text-gray-700 mb-1">Категория *</label>
                                        <select
                                            name="category_code"
                                            value={formData.category_code}
                                            onChange={handleInputChange}
                                            required
                                            className="w-full px-3 py-2 border border-gray-300 rounded-lg focus:outline-none focus:ring-2 focus:ring-blue-500"
                                        >
                                            {(Array.isArray(categories) ? categories : []).map(cat => (
                                                <option key={cat.id} value={cat.code}>{cat.name}</option>
                                            ))}
                                        </select>
                                    </div>
                                </div>

                                <div>
                                    <label className="block text-sm font-medium text-gray-700 mb-1">Название *</label>
                                    <input
                                        type="text"
                                        name="name"
                                        value={formData.name}
                                        onChange={handleInputChange}
                                        required
                                        className="w-full px-3 py-2 border border-gray-300 rounded-lg focus:outline-none focus:ring-2 focus:ring-blue-500"
                                    />
                                </div>

                                <div>
                                    <label className="block text-sm font-medium text-gray-700 mb-1">Описание</label>
                                    <textarea
                                        name="description"
                                        value={formData.description}
                                        onChange={handleInputChange}
                                        rows={2}
                                        className="w-full px-3 py-2 border border-gray-300 rounded-lg focus:outline-none focus:ring-2 focus:ring-blue-500"
                                    />
                                </div>

                                <div className="grid grid-cols-2 gap-4">
                                    <div>
                                        <label className="block text-sm font-medium text-gray-700 mb-1">Тип *</label>
                                        <select
                                            name="indicator_type"
                                            value={formData.indicator_type}
                                            onChange={handleInputChange}
                                            required
                                            className="w-full px-3 py-2 border border-gray-300 rounded-lg focus:outline-none focus:ring-2 focus:ring-blue-500"
                                        >
                                            <option value="base">Базовый</option>
                                            <option value="extra">Дополнительный</option>
                                            <option value="penalty">Штрафной</option>
                                        </select>
                                    </div>
                                    <div>
                                        <label className="block text-sm font-medium text-gray-700 mb-1">Единица измерения *</label>
                                        <select
                                            name="unit"
                                            value={formData.unit}
                                            onChange={handleInputChange}
                                            required
                                            className="w-full px-3 py-2 border border-gray-300 rounded-lg focus:outline-none focus:ring-2 focus:ring-blue-500"
                                        >
                                            <option value="%">Проценты (%)</option>
                                            <option value="шт">Штуки (шт)</option>
                                            <option value="чел">Человеки (чел)</option>
                                        </select>
                                    </div>
                                </div>

                                {formData.indicator_type === 'base' && (
                                    <div className="grid grid-cols-2 gap-4 bg-blue-50 p-4 rounded-lg">
                                        <div>
                                            <label className="block text-sm font-medium text-gray-700 mb-1">Целевое значение *</label>
                                            <input
                                                type="number"
                                                name="base_value"
                                                value={formData.base_value}
                                                onChange={handleInputChange}
                                                required
                                                step="0.1"
                                                className="w-full px-3 py-2 border border-gray-300 rounded-lg focus:outline-none focus:ring-2 focus:ring-blue-500"
                                            />
                                        </div>
                                        <div>
                                            <label className="block text-sm font-medium text-gray-700 mb-1">Вес баллов *</label>
                                            <input
                                                type="number"
                                                name="base_weight"
                                                value={formData.base_weight}
                                                onChange={handleInputChange}
                                                required
                                                className="w-full px-3 py-2 border border-gray-300 rounded-lg focus:outline-none focus:ring-2 focus:ring-blue-500"
                                            />
                                        </div>
                                    </div>
                                )}

                                {formData.indicator_type === 'extra' && (
                                    <div className="bg-green-50 p-4 rounded-lg">
                                        <label className="block text-sm font-medium text-gray-700 mb-1">Вес за единицу *</label>
                                        <input
                                            type="number"
                                            name="extra_weight"
                                            value={formData.extra_weight}
                                            onChange={handleInputChange}
                                            required
                                            className="w-full px-3 py-2 border border-gray-300 rounded-lg focus:outline-none focus:ring-2 focus:ring-blue-500"
                                        />
                                    </div>
                                )}

                                {formData.indicator_type === 'penalty' && (
                                    <div className="bg-red-50 p-4 rounded-lg">
                                        <label className="block text-sm font-medium text-gray-700 mb-1">Штраф за единицу *</label>
                                        <input
                                            type="number"
                                            name="penalty_weight"
                                            value={formData.penalty_weight}
                                            onChange={handleInputChange}
                                            required
                                            className="w-full px-3 py-2 border border-gray-300 rounded-lg focus:outline-none focus:ring-2 focus:ring-blue-500"
                                        />
                                    </div>
                                )}

                                <div className="flex justify-end space-x-3 pt-4">
                                    <button
                                        type="button"
                                        onClick={() => {
                                            setShowModal(false);
                                            resetForm();
                                        }}
                                        className="px-4 py-2 border border-gray-300 rounded-lg hover:bg-gray-50 transition"
                                    >
                                        Отмена
                                    </button>
                                    <button
                                        type="submit"
                                        className="px-4 py-2 bg-blue-600 text-white rounded-lg hover:bg-blue-700 transition"
                                    >
                                        {editingId ? 'Сохранить' : 'Создать'}
                                    </button>
                                </div>
                            </form>
                        </div>
                    </div>
                </div>
            )}
        </div>
    );
};