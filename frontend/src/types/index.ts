// Типы пользователей
export type UserRole = 'tm' | 'expert' | 'coordinator';

export interface User {
    id: number;
    email: string;
    role: UserRole;
    fio: string;
    cluster_name?: string;
    azs_count: number;
    created_at: string;
    updated_at: string;
}

// Аутентификация
export interface LoginRequest {
    email: string;
    password: string;
}

export interface LoginResponse {
    token: string;
    user: User;
}

// KPI
export interface KPICategory {
    id: number;
    name: string;
    code: string;
    description: string;
    created_at: string;
}

export type IndicatorType = 'base' | 'extra' | 'penalty';

export interface KPIIndicator {
    id: number;
    category_id: number;
    code: string;
    name: string;
    description: string;
    unit: string;
    indicator_type: IndicatorType;
    base_value?: number | null;
    base_weight?: number | null;
    extra_weight?: number | null;
    penalty_weight?: number | null;
    created_at: string;
    category_code?: string;
    category_name?: string;
}

// Результаты
export type ResultStatus = 'draft' | 'confirmed';

export interface MonthlyResult {
    id: number;
    user_id: number;
    expert_id?: number | null;
    period: string;
    status: ResultStatus;
    created_at: string;
    updated_at: string;
    user?: User;
    expert?: User;
}

export interface IndicatorResult {
    id: number;
    monthly_result_id: number;
    indicator_id: number;
    fact_value?: number | null;
    calculated_points: number;
    supporting_document_url?: string | null;
    created_at: string;
    indicator?: KPIIndicator;
}

export interface IndicatorResultInput {
    indicator_code: string;
    fact_value?: number | null;
    document_url?: string | null;
}

export interface EnterResultRequest {
    user_id: number;
    period: string; // YYYY-MM
    results: IndicatorResultInput[];
}

// Сводка по категориям
export interface CategorySummary {
    category_code: string;
    category_name: string;
    base_points: number;
    extra_points: number;
    penalty_points: number;
    total_points: number;
}

export interface FullResultSummary {
    user_id: number;
    user_fio: string;
    period: string;
    categories: CategorySummary[];
    total_points: number;
    level?: Level | null;
}

// Уровни
export interface Level {
    id: number;
    name: string;
    min_points_per_year: number;
    privileges: Record<string, any>;
    created_at: string;
}

export interface UserLevelHistory {
    id: number;
    user_id: number;
    level_id: number;
    assigned_at: string;
    points_year: number;
    created_at: string;
    level?: Level;
}

// Для API ответов с ошибками
export interface ApiError {
    message: string;
    status?: number;
}

export interface FullResultSummary {
    user_id: number;
    user_fio: string;
    period: string;
    categories: CategorySummary[];
    detailed_results?: IndicatorResult[]; // Добавлено
    total_points: number;
    level?: Level | null;
}