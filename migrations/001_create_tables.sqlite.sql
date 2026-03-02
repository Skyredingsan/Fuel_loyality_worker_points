-- SQLite не поддерживает ENUM, используем CHECK constraints или TEXT
-- Включаем поддержку внешних ключей
PRAGMA foreign_keys = ON;

-- Таблица пользователей
CREATE TABLE IF NOT EXISTS users (
                                     id INTEGER PRIMARY KEY AUTOINCREMENT,
                                     email TEXT UNIQUE NOT NULL,
                                     password_hash TEXT NOT NULL,
                                     role TEXT NOT NULL CHECK(role IN ('tm', 'expert', 'coordinator')),
    fio TEXT NOT NULL,
    cluster_name TEXT,
    azs_count INTEGER DEFAULT 0,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
    );

-- Таблица категорий KPI
CREATE TABLE IF NOT EXISTS kpi_categories (
                                              id INTEGER PRIMARY KEY AUTOINCREMENT,
                                              name TEXT NOT NULL,
                                              code TEXT UNIQUE NOT NULL,
                                              description TEXT,
                                              created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Таблица показателей KPI
CREATE TABLE IF NOT EXISTS kpi_indicators (
                                              id INTEGER PRIMARY KEY AUTOINCREMENT,
                                              category_id INTEGER REFERENCES kpi_categories(id) ON DELETE CASCADE,
    code TEXT UNIQUE NOT NULL,
    name TEXT NOT NULL,
    description TEXT,
    unit TEXT,
    indicator_type TEXT NOT NULL CHECK(indicator_type IN ('base', 'extra', 'penalty')),

    -- Для базовых показателей
    base_value REAL,
    base_weight INTEGER,

    -- Для дополнительных
    extra_weight INTEGER,

    -- Для штрафных
    penalty_weight INTEGER,

    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
    );

-- Таблица ежемесячных результатов
CREATE TABLE IF NOT EXISTS monthly_results (
                                               id INTEGER PRIMARY KEY AUTOINCREMENT,
                                               user_id INTEGER REFERENCES users(id) ON DELETE CASCADE,
    expert_id INTEGER REFERENCES users(id) ON DELETE SET NULL,
    period DATE NOT NULL,
    status TEXT DEFAULT 'draft' CHECK(status IN ('draft', 'confirmed')),
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(user_id, period)
    );

-- Таблица результатов по показателям
CREATE TABLE IF NOT EXISTS indicator_results (
                                                 id INTEGER PRIMARY KEY AUTOINCREMENT,
                                                 monthly_result_id INTEGER REFERENCES monthly_results(id) ON DELETE CASCADE,
    indicator_id INTEGER REFERENCES kpi_indicators(id) ON DELETE CASCADE,
    fact_value REAL,
    calculated_points INTEGER DEFAULT 0,
    supporting_document_url TEXT,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(monthly_result_id, indicator_id)
    );

-- Таблица уровней
CREATE TABLE IF NOT EXISTS levels (
                                      id INTEGER PRIMARY KEY AUTOINCREMENT,
                                      name TEXT NOT NULL,
                                      min_points_per_year INTEGER NOT NULL,
                                      privileges TEXT, -- JSON stored as TEXT
                                      created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Таблица истории уровней пользователей
CREATE TABLE IF NOT EXISTS user_level_history (
                                                  id INTEGER PRIMARY KEY AUTOINCREMENT,
                                                  user_id INTEGER REFERENCES users(id) ON DELETE CASCADE,
    level_id INTEGER REFERENCES levels(id) ON DELETE CASCADE,
    assigned_at DATE NOT NULL,
    points_year INTEGER,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
    );

-- Индексы для производительности
CREATE INDEX IF NOT EXISTS idx_monthly_results_user_period ON monthly_results(user_id, period);
CREATE INDEX IF NOT EXISTS idx_indicator_results_monthly_result ON indicator_results(monthly_result_id);
CREATE INDEX IF NOT EXISTS idx_users_role ON users(role);

-- Вставка начальных данных (уровни)
INSERT OR IGNORE INTO levels (name, min_points_per_year, privileges) VALUES
    ('Специалист Трассы', 0, '{"bonus": "Стандартный пакет мотивации"}'),
    ('Тактик Магистрали', 4321, '{"bonus": "Доплата 20% к окладу"}'),
    ('Стратег Гран-при', 4321, '{"bonus": "Доплата 50% к окладу", "prize": "Поездка на Кубу"}');

-- Вставка категорий KPI
INSERT OR IGNORE INTO kpi_categories (code, name, description) VALUES
    ('ПМ', 'Продажи и маржа', 'Показатели продаж топлива, СТ и маржинальности'),
    ('ОЭК', 'Операционная эффективность и качество', 'Тайный покупатель, аудиты, штрафы'),
    ('ЭКЛ', 'Эффективность команды и лидерство', 'Текучесть, обучение, дисциплина'),
    ('КБ', 'Культура безопасности', 'Травматизм, ПБОТОС, инициативы');