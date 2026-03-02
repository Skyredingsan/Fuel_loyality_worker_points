import React from 'react';
import { NavLink } from 'react-router-dom';
import { useAuth } from '../contexts/AuthContext';

export const Sidebar: React.FC = () => {
    const { hasRole } = useAuth();

    const menuItems = [
        { path: '/dashboard', label: 'Дашборд', icon: '📊', roles: ['tm', 'expert', 'coordinator'] },
        { path: '/results', label: 'Мои результаты', icon: '📈', roles: ['tm'] },
        { path: '/enter-results', label: 'Ввод результатов', icon: '✏️', roles: ['expert', 'coordinator'] },
        { path: '/kpi', label: 'KPI Показатели', icon: '🎯', roles: ['coordinator'] },
        { path: '/users', label: 'Пользователи', icon: '👥', roles: ['coordinator'] },
        { path: '/reports', label: 'Отчеты', icon: '📑', roles: ['coordinator'] },
    ];

    return (
        <aside className="w-64 bg-white shadow min-h-[calc(100vh-73px)]">
            <nav className="p-4">
                <ul className="space-y-2">
                    {menuItems
                        .filter(item => hasRole(item.roles))
                        .map(item => (
                            <li key={item.path}>
                                <NavLink
                                    to={item.path}
                                    className={({ isActive }) =>
                                        `flex items-center space-x-3 px-4 py-2 rounded transition ${
                                            isActive
                                                ? 'bg-blue-50 text-blue-600'
                                                : 'text-gray-700 hover:bg-gray-100'
                                        }`
                                    }
                                >
                                    <span>{item.icon}</span>
                                    <span>{item.label}</span>
                                </NavLink>
                            </li>
                        ))}
                </ul>
            </nav>
        </aside>
    );
};