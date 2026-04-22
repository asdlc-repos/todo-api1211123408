import { Navigate, Route, Routes } from 'react-router-dom';
import { ProtectedRoute } from './components/ProtectedRoute';
import { Navbar } from './components/Navbar';
import { Login } from './pages/Login';
import { Register } from './pages/Register';
import { ForgotPassword } from './pages/ForgotPassword';
import { ResetPassword } from './pages/ResetPassword';
import { TodosPage } from './pages/Todos';
import { CategoriesPage } from './pages/Categories';
import { useAuth } from './auth';

export default function App() {
  const { isAuthenticated } = useAuth();

  return (
    <div className="app-shell">
      {isAuthenticated && <Navbar />}
      <Routes>
        <Route
          path="/login"
          element={isAuthenticated ? <Navigate to="/todos" replace /> : <Login />}
        />
        <Route
          path="/register"
          element={isAuthenticated ? <Navigate to="/todos" replace /> : <Register />}
        />
        <Route
          path="/forgot-password"
          element={isAuthenticated ? <Navigate to="/todos" replace /> : <ForgotPassword />}
        />
        <Route
          path="/reset-password"
          element={isAuthenticated ? <Navigate to="/todos" replace /> : <ResetPassword />}
        />
        <Route
          path="/todos"
          element={
            <ProtectedRoute>
              <TodosPage />
            </ProtectedRoute>
          }
        />
        <Route
          path="/categories"
          element={
            <ProtectedRoute>
              <CategoriesPage />
            </ProtectedRoute>
          }
        />
        <Route path="/" element={<Navigate to={isAuthenticated ? '/todos' : '/login'} replace />} />
        <Route path="*" element={<Navigate to={isAuthenticated ? '/todos' : '/login'} replace />} />
      </Routes>
    </div>
  );
}
