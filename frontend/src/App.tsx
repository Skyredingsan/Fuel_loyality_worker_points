import { BrowserRouter, Routes, Route, Navigate } from 'react-router-dom';
import { AuthProvider } from './contexts/AuthContext';
import { ProtectedRoute } from './components/ProtectedRoute';
import { Layout } from './components/Layout';
import { Login } from './pages/Login';
import { Dashboard } from './pages/Dashboard'; // для ТМ и экспертов
import { CoordinatorDashboard } from './pages/CoordinatorDashboard'; // для координатора
import { EnterResults } from './pages/EnterResults';
import { Results } from './pages/Results';
import { KPIEditor } from './pages/KPIEditor';
import { Users } from './pages/Users';
import { Reports } from './pages/Reports';

function App() {
    return (
        <AuthProvider>
            <BrowserRouter>
                <Routes>
                    <Route path="/login" element={<Login />} />

                    <Route path="/" element={
                        <ProtectedRoute>
                            <Layout />
                        </ProtectedRoute>
                    }>
                        <Route index element={<Navigate to="/dashboard" replace />} />

                        {/* ВАЖНО: разные дашборды для разных ролей */}
                        <Route path="dashboard" element={
                            <ProtectedRoute allowedRoles={['coordinator']}>
                                <CoordinatorDashboard />
                            </ProtectedRoute>
                        } />

                        {/* Для ТМ и экспертов используем обычный Dashboard */}
                        <Route path="dashboard" element={
                            <ProtectedRoute allowedRoles={['tm', 'expert']}>
                                <Dashboard />
                            </ProtectedRoute>
                        } />

                        {/* Остальные маршруты */}
                        <Route path="results" element={
                            <ProtectedRoute allowedRoles={['tm']}>
                                <Results />
                            </ProtectedRoute>
                        } />

                        <Route path="enter-results" element={
                            <ProtectedRoute allowedRoles={['expert', 'coordinator']}>
                                <EnterResults />
                            </ProtectedRoute>
                        } />

                        <Route path="kpi" element={
                            <ProtectedRoute allowedRoles={['coordinator']}>
                                <KPIEditor />
                            </ProtectedRoute>
                        } />

                        <Route path="users" element={
                            <ProtectedRoute allowedRoles={['coordinator']}>
                                <Users />
                            </ProtectedRoute>
                        } />

                        <Route path="reports" element={
                            <ProtectedRoute allowedRoles={['coordinator']}>
                                <Reports />
                            </ProtectedRoute>
                        } />
                    </Route>

                    <Route path="*" element={<Navigate to="/dashboard" replace />} />
                </Routes>
            </BrowserRouter>
        </AuthProvider>
    );
}

export default App;