import React from 'react';
import { useAuth } from '../contexts/AuthContext';

export const Header: React.FC = () => {
    const { user, logout } = useAuth();

    return (
        <header className="bg-white shadow">
            <div className="flex justify-between items-center px-6 py-4">
                <h1 className="text-2xl font-bold text-gray-800">
                    Топливный Альянс
                </h1>
                <div className="flex items-center space-x-4">
          <span className="text-sm text-gray-600">
            {user?.fio} ({user?.role === 'tm' ? 'ТМ' :
              user?.role === 'expert' ? 'Эксперт' : 'Координатор'})
          </span>
                    <button
                        onClick={logout}
                        className="px-3 py-1 text-sm text-red-600 hover:bg-red-50 rounded"
                    >
                        Выйти
                    </button>
                </div>
            </div>
        </header>
    );
};