import { lazy, Suspense } from 'react';
import { BrowserRouter, Routes, Route, Navigate, useLocation } from 'react-router-dom';
import Layout from './components/Layout';
import LoadingSpinner from './components/ui/LoadingSpinner';
import { useAuthStore } from './store/auth-store';
import { useThemeStore } from './store/theme-store';
import { useEffect } from 'react';
import { AnimatePresence, motion } from 'framer-motion';

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
const Pricing = lazy(() => import('./pages/Pricing'));
const ClientApps = lazy(() => import('./pages/ClientApps'));
const Databases = lazy(() => import('./pages/Databases'));
const DatabaseDetail = lazy(() => import('./pages/DatabaseDetail'));
const Autoscale = lazy(() => import('./pages/Autoscale'));
const Branches = lazy(() => import('./pages/Branches'));
const MultiRegion = lazy(() => import('./pages/MultiRegion'));
const DisasterRecovery = lazy(() => import('./pages/DisasterRecovery'));
const PostgresStats = lazy(() => import('./pages/PostgresStats'));
const ClusterScanner = lazy(() => import('./pages/ClusterScanner'));

// Protected Route Component
function ProtectedRoute({ children }: { children: React.ReactNode }) {
  const { isAuthenticated } = useAuthStore();

  if (!isAuthenticated) {
    return <Navigate to="/login" replace />;
  }

  return <>{children}</>;
}

const PageTransition = ({ children }: { children: React.ReactNode }) => (
  <motion.div
    initial={{ opacity: 0, y: 20 }}
    animate={{ opacity: 1, y: 0 }}
    exit={{ opacity: 0, y: -20 }}
    transition={{ duration: 0.2 }}
  >
    {children}
  </motion.div>
);

function AnimatedRoutes() {
  const location = useLocation();

  return (
    <AnimatePresence mode="wait">
      <Routes location={location} key={location.pathname}>
        <Route path="/login" element={
          <Suspense fallback={<LoadingSpinner />}>
            <PageTransition>
              <Login />
            </PageTransition>
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
                <PageTransition>
                  <Dashboard />
                </PageTransition>
              </Suspense>
            }
          />
          <Route
            path="shards"
            element={
              <Suspense fallback={<LoadingSpinner size="lg" />}>
                <PageTransition>
                  <Shards />
                </PageTransition>
              </Suspense>
            }
          />
          <Route
            path="shards/:id"
            element={
              <Suspense fallback={<LoadingSpinner size="lg" />}>
                <PageTransition>
                  <ShardDetail />
                </PageTransition>
              </Suspense>
            }
          />
          <Route
            path="query"
            element={
              <Suspense fallback={<LoadingSpinner size="lg" />}>
                <PageTransition>
                  <QueryExecutor />
                </PageTransition>
              </Suspense>
            }
          />
          <Route
            path="resharding"
            element={
              <Suspense fallback={<LoadingSpinner size="lg" />}>
                <PageTransition>
                  <Resharding />
                </PageTransition>
              </Suspense>
            }
          />
          <Route
            path="resharding/jobs/:id"
            element={
              <Suspense fallback={<LoadingSpinner size="lg" />}>
                <PageTransition>
                  <ReshardJobDetail />
                </PageTransition>
              </Suspense>
            }
          />
          <Route
            path="health"
            element={
              <Suspense fallback={<LoadingSpinner size="lg" />}>
                <PageTransition>
                  <Health />
                </PageTransition>
              </Suspense>
            }
          />
          <Route
            path="metrics"
            element={
              <Suspense fallback={<LoadingSpinner size="lg" />}>
                <PageTransition>
                  <Metrics />
                </PageTransition>
              </Suspense>
            }
          />
          <Route
            path="settings"
            element={
              <Suspense fallback={<LoadingSpinner size="lg" />}>
                <PageTransition>
                  <Settings />
                </PageTransition>
              </Suspense>
            }
          />
          <Route
            path="pricing"
            element={
              <Suspense fallback={<LoadingSpinner size="lg" />}>
                <PageTransition>
                  <Pricing />
                </PageTransition>
              </Suspense>
            }
          />
          <Route
            path="client-apps"
            element={
              <Suspense fallback={<LoadingSpinner size="lg" />}>
                <PageTransition>
                  <ClientApps />
                </PageTransition>
              </Suspense>
            }
          />
          <Route
            path="databases"
            element={
              <Suspense fallback={<LoadingSpinner size="lg" />}>
                <PageTransition>
                  <Databases />
                </PageTransition>
              </Suspense>
            }
          />
          <Route
            path="databases/:id"
            element={
              <Suspense fallback={<LoadingSpinner size="lg" />}>
                <PageTransition>
                  <DatabaseDetail />
                </PageTransition>
              </Suspense>
            }
          />
          <Route
            path="databases/:dbName/branches"
            element={
              <Suspense fallback={<LoadingSpinner size="lg" />}>
                <PageTransition>
                  <Branches />
                </PageTransition>
              </Suspense>
            }
          />
          <Route
            path="autoscale"
            element={
              <Suspense fallback={<LoadingSpinner size="lg" />}>
                <PageTransition>
                  <Autoscale />
                </PageTransition>
              </Suspense>
            }
          />
          <Route
            path="multi-region"
            element={
              <Suspense fallback={<LoadingSpinner size="lg" />}>
                <PageTransition>
                  <MultiRegion />
                </PageTransition>
              </Suspense>
            }
          />
          <Route
            path="disaster-recovery"
            element={
              <Suspense fallback={<LoadingSpinner size="lg" />}>
                <PageTransition>
                  <DisasterRecovery />
                </PageTransition>
              </Suspense>
            }
          />
          <Route
            path="postgres-stats"
            element={
              <Suspense fallback={<LoadingSpinner size="lg" />}>
                <PageTransition>
                  <PostgresStats />
                </PageTransition>
              </Suspense>
            }
          />
          <Route
            path="cluster-scanner"
            element={
              <Suspense fallback={<LoadingSpinner size="lg" />}>
                <PageTransition>
                  <ClusterScanner />
                </PageTransition>
              </Suspense>
            }
          />
        </Route>
      </Routes>
    </AnimatePresence>
  );
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
      <AnimatedRoutes />
    </BrowserRouter>
  );
}

export default App;

