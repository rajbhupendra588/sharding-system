import { lazy, Suspense } from 'react';
import { BrowserRouter, Routes, Route, Navigate } from 'react-router-dom';
import Layout from './components/Layout';
import LoadingSpinner from './components/ui/LoadingSpinner';

// Lazy load pages for code splitting
const Dashboard = lazy(() => import('./pages/Dashboard'));
const Shards = lazy(() => import('./pages/Shards'));
const ShardDetail = lazy(() => import('./pages/ShardDetail'));
const QueryExecutor = lazy(() => import('./pages/QueryExecutor'));
const Resharding = lazy(() => import('./pages/Resharding'));
const ReshardJobDetail = lazy(() => import('./pages/ReshardJobDetail'));
const Health = lazy(() => import('./pages/Health'));
const Metrics = lazy(() => import('./pages/Metrics'));
const Settings = lazy(() => import('./pages/Settings'));

function App() {
  // For now, allow access without authentication
  // In production, add proper auth check here
  // const { isAuthenticated } = useAuthStore();
  // if (!isAuthenticated) {
  //   return <Navigate to="/login" replace />;
  // }

  return (
    <BrowserRouter>
      <Routes>
        <Route path="/" element={<Layout />}>
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

