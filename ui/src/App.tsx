import { lazy, Suspense } from 'react';
import { BrowserRouter, Routes, Route, Navigate } from 'react-router-dom';
import Layout from './components/Layout';
import LoadingSpinner from './components/ui/LoadingSpinner';
import { useAuthStore } from './store/auth-store';
import { useThemeStore } from './store/theme-store';
import { useEffect } from 'react';

// Lazy load pages for code splitting
const Login = lazy(() => import('./pages/Login'));
const Dashboard = lazy(() => import('./pages/Dashboard'));
const Shards = lazy(() => import('./pages/Shards'));
const ShardDetail = lazy(() => import('./pages/ShardDetail'));
const QueryExecutor = lazy(() => import('./pages/QueryExecutor'));
const Resharding = lazy(() => import('./pages/Resharding'));
const ReshardJobDetail = lazy(() => import('./pages/ReshardJobDetail'));
const Health = lazy(() => import('./pages/Health'));
const Metrics = lazy(() => import('./pages/Metrics'));
const Settings = lazy(() => import('./pages/Settings'));

// Protected Route Component
function ProtectedRoute({ children }: { children: React.ReactNode }) {
  const { isAuthenticated } = useAuthStore();

  if (!isAuthenticated) {
    return <Navigate to="/login" replace />;
  }

  return <>{children}</>;
}

function App() {
  const { theme } = useThemeStore();

  useEffect(() => {
    if (theme === 'dark') {
      document.documentElement.classList.add('dark');
    } else {
      document.documentElement.classList.remove('dark');
    }
  }, [theme]);


  return (
    <BrowserRouter>
      <Routes>
        <Route path="/login" element={
          <Suspense fallback={<LoadingSpinner />}>
            <Login />
          </Suspense>
        } />
        <Route path="/" element={
          <ProtectedRoute>
            <Layout />
          </ProtectedRoute>
        }>
          <Route index element={<Navigate to="/dashboard" replace />} />
          <Route
            path="dashboard"
            element={
              <Suspense fallback={<LoadingSpinner size="lg" />}>
                <Dashboard />
              </Suspense>
            }
          />
          <Route
            path="shards"
            element={
              <Suspense fallback={<LoadingSpinner size="lg" />}>
                <Shards />
              </Suspense>
            }
          />
          <Route
            path="shards/:id"
            element={
              <Suspense fallback={<LoadingSpinner size="lg" />}>
                <ShardDetail />
              </Suspense>
            }
          />
          <Route
            path="query"
            element={
              <Suspense fallback={<LoadingSpinner size="lg" />}>
                <QueryExecutor />
              </Suspense>
            }
          />
          <Route
            path="resharding"
            element={
              <Suspense fallback={<LoadingSpinner size="lg" />}>
                <Resharding />
              </Suspense>
            }
          />
          <Route
            path="resharding/jobs/:id"
            element={
              <Suspense fallback={<LoadingSpinner size="lg" />}>
                <ReshardJobDetail />
              </Suspense>
            }
          />
          <Route
            path="health"
            element={
              <Suspense fallback={<LoadingSpinner size="lg" />}>
                <Health />
              </Suspense>
            }
          />
          <Route
            path="metrics"
            element={
              <Suspense fallback={<LoadingSpinner size="lg" />}>
                <Metrics />
              </Suspense>
            }
          />
          <Route
            path="settings"
            element={
              <Suspense fallback={<LoadingSpinner size="lg" />}>
                <Settings />
              </Suspense>
            }
          />
        </Route>
      </Routes>
    </BrowserRouter>
  );
}

export default App;

