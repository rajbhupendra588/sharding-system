import { useState, useEffect } from 'react';
import { useNavigate, useLocation } from 'react-router-dom';
import { useAuthStore } from '@/store/auth-store';
import Button from '@/components/ui/Button';
import Input from '@/components/ui/Input';
import { appConfig } from '@/core/config';

interface OAuthProvider {
  name: string;
  displayName: string;
  icon: string;
  color: string;
}

const OAUTH_PROVIDERS: Record<string, OAuthProvider> = {
  google: {
    name: 'google',
    displayName: 'Google',
    icon: '🔍',
    color: 'bg-white hover:bg-gray-50 text-gray-700 border-gray-300',
  },
  github: {
    name: 'github',
    displayName: 'GitHub',
    icon: '🐙',
    color: 'bg-gray-800 hover:bg-gray-900 text-white border-gray-700',
  },
  facebook: {
    name: 'facebook',
    displayName: 'Facebook',
    icon: '📘',
    color: 'bg-blue-600 hover:bg-blue-700 text-white border-blue-600',
  },
};

export default function Login() {
  const [username, setUsername] = useState('');
  const [password, setPassword] = useState('');
  const [error, setError] = useState('');
  const [loading, setLoading] = useState(false);
  const [oauthProviders, setOAuthProviders] = useState<string[]>([]);
  const [loadingProviders, setLoadingProviders] = useState(true);
  const [serviceStatus, setServiceStatus] = useState<{ available: boolean; message?: string }>({ available: true });
  const { setToken } = useAuthStore();
  const navigate = useNavigate();
  const location = useLocation();

  // Handle OAuth callback
  useEffect(() => {
    // Check hash first (more secure)
    const hash = location.hash;
    if (hash) {
      const params = new URLSearchParams(hash.substring(1));
      const token = params.get('token');
      const username = params.get('username');
      
      if (token && username) {
        setToken(token);
        navigate('/dashboard');
        return;
      }
    }
    
    // Also check query params (fallback)
    const searchParams = new URLSearchParams(location.search);
    const token = searchParams.get('token');
    const username = searchParams.get('username');
    
    if (token && username) {
      setToken(token);
      navigate('/dashboard');
    }
  }, [location, setToken, navigate]);

  // Check manager service availability
  useEffect(() => {
    const checkServiceStatus = async () => {
      try {
        const managerUrl = appConfig.getConfig().managerUrl;
        let baseUrl = managerUrl || '';
        
        // If managerUrl points to the same origin (dev server), use relative URL instead
        if (baseUrl && (baseUrl.includes('localhost:3000') || baseUrl.includes('127.0.0.1:3000'))) {
          baseUrl = ''; // Use relative URL to leverage Vite proxy
        }
        
        const healthUrl = baseUrl ? `${baseUrl}/api/v1/health` : '/api/v1/health';
        const response = await fetch(healthUrl, { 
          method: 'GET',
          signal: AbortSignal.timeout(3000) // 3 second timeout
        });
        
        if (response.ok) {
          setServiceStatus({ available: true });
        } else {
          setServiceStatus({ 
            available: false, 
            message: 'Manager service is not responding correctly' 
          });
        }
      } catch (err) {
        setServiceStatus({ 
          available: false, 
          message: 'Manager service is not available. Please ensure the manager server is running on port 8081. You can start it with: ./scripts/start.sh' 
        });
      }
    };
    checkServiceStatus();
  }, []);

  // Fetch available OAuth providers
  useEffect(() => {
    const fetchProviders = async () => {
      // Only fetch if service is available
      if (!serviceStatus.available) {
        setLoadingProviders(false);
        return;
      }

      try {
        const managerUrl = appConfig.getConfig().managerUrl;
        let baseUrl = managerUrl || '';
        
        // If managerUrl points to the same origin (dev server), use relative URL instead
        if (baseUrl && (baseUrl.includes('localhost:3000') || baseUrl.includes('127.0.0.1:3000'))) {
          baseUrl = ''; // Use relative URL to leverage Vite proxy
        }
        
        const url = baseUrl ? `${baseUrl}/api/v1/auth/oauth/providers` : '/api/v1/auth/oauth/providers';
        console.log('Fetching OAuth providers from:', url);
        
        const response = await fetch(url);
        console.log('OAuth providers response status:', response.status);
        
        if (response.ok) {
          const data = await response.json();
          console.log('OAuth providers data:', data);
          setOAuthProviders(data.providers || []);
        } else {
          console.warn('Failed to fetch OAuth providers:', response.status, response.statusText);
        }
      } catch (err) {
        console.error('Failed to fetch OAuth providers:', err);
      } finally {
        setLoadingProviders(false);
      }
    };
    fetchProviders();
  }, [serviceStatus.available]);

  const handleOAuthLogin = (provider: string) => {
    const managerUrl = appConfig.getConfig().managerUrl;
    let baseUrl = managerUrl || '';
    
    // If managerUrl points to the same origin (dev server), use relative URL instead
    if (baseUrl && (baseUrl.includes('localhost:3000') || baseUrl.includes('127.0.0.1:3000'))) {
      baseUrl = ''; // Use relative URL to leverage Vite proxy
    }
    
    const redirectUri = encodeURIComponent(window.location.origin + '/login');
    const oauthUrl = baseUrl 
      ? `${baseUrl}/api/v1/auth/oauth/${provider}?redirect_uri=${redirectUri}`
      : `/api/v1/auth/oauth/${provider}?redirect_uri=${redirectUri}`;
    
    window.location.href = oauthUrl;
  };

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    setError('');
    setLoading(true);

    try {
      // Construct URL properly: use relative URL if managerUrl is empty (for Vite proxy)
      // In development, always use relative URLs to leverage Vite proxy
      const managerUrl = appConfig.getConfig().managerUrl;
      let baseUrl = managerUrl || '';
      
      // If managerUrl points to the same origin (dev server), use relative URL instead
      if (baseUrl && (baseUrl.includes('localhost:3000') || baseUrl.includes('127.0.0.1:3000'))) {
        baseUrl = ''; // Use relative URL to leverage Vite proxy
      }
      
      const url = baseUrl ? `${baseUrl}/api/v1/auth/login` : '/api/v1/auth/login';
      console.log('Attempting login to:', url);

      const response = await fetch(url, {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
        },
        body: JSON.stringify({ username, password }),
      });

      console.log('Login response status:', response.status);

      if (!response.ok) {
        // Try to parse JSON error, but handle non-JSON responses (like 404 HTML pages)
        let errorMessage = 'Login failed';
        const contentType = response.headers.get('content-type');
        
        if (contentType && contentType.includes('application/json')) {
          try {
            const errorData = await response.json();
            errorMessage = errorData.error?.message || errorData.message || `Login failed (${response.status})`;
          } catch (parseError) {
            errorMessage = `Login failed (${response.status})`;
          }
        } else {
          // Non-JSON response (likely HTML 404 page)
          if (response.status === 404) {
            errorMessage = 'Authentication endpoint not found. Please check if the manager server is running on port 8081.';
          } else {
            errorMessage = `Login failed (${response.status} ${response.statusText})`;
          }
        }
        
        console.error('Login error:', errorMessage);
        throw new Error(errorMessage);
      }

      const data = await response.json();
      console.log('Login successful, token received');
      setToken(data.token);
      navigate('/dashboard');
    } catch (err) {
      console.error('Login exception:', err);
      setError(err instanceof Error ? err.message : 'Login failed');
    } finally {
      setLoading(false);
    }
  };

  return (
    <div className="min-h-screen flex items-center justify-center bg-gray-50 dark:bg-gray-900 py-12 px-4 sm:px-6 lg:px-8">
      <div className="max-w-md w-full space-y-8">
        <div>
          <h2 className="mt-6 text-center text-3xl font-extrabold text-gray-900 dark:text-white">
            Sign in to Sharding System
          </h2>
          <p className="mt-2 text-center text-sm text-gray-600 dark:text-gray-400">
            Demo credentials: admin/admin123, operator/operator123, viewer/viewer123
          </p>
        </div>
        <form className="mt-8 space-y-6" onSubmit={handleSubmit}>
          {!serviceStatus.available && (
            <div className="rounded-md bg-yellow-50 dark:bg-yellow-900/20 border border-yellow-200 dark:border-yellow-800 p-4 mb-4">
              <div className="flex">
                <div className="flex-shrink-0">
                  <svg className="h-5 w-5 text-yellow-400" viewBox="0 0 20 20" fill="currentColor">
                    <path fillRule="evenodd" d="M8.257 3.099c.765-1.36 2.722-1.36 3.486 0l5.58 9.92c.75 1.334-.213 2.98-1.742 2.98H4.42c-1.53 0-2.493-1.646-1.743-2.98l5.58-9.92zM11 13a1 1 0 11-2 0 1 1 0 012 0zm-1-8a1 1 0 00-1 1v3a1 1 0 002 0V6a1 1 0 00-1-1z" clipRule="evenodd" />
                  </svg>
                </div>
                <div className="ml-3">
                  <h3 className="text-sm font-medium text-yellow-800 dark:text-yellow-200">
                    Manager Service Unavailable
                  </h3>
                  <div className="mt-2 text-sm text-yellow-700 dark:text-yellow-300">
                    <p>{serviceStatus.message || 'The manager server is not running. Please start it to continue.'}</p>
                    <p className="mt-2 font-mono text-xs bg-yellow-100 dark:bg-yellow-900/40 p-2 rounded">
                      Run: <code className="font-bold">./scripts/start.sh</code>
                    </p>
                  </div>
                </div>
              </div>
            </div>
          )}
          {error && (
            <div className="rounded-md bg-red-50 dark:bg-red-900/20 border border-red-200 dark:border-red-800 p-4">
              <div className="text-sm text-red-800 dark:text-red-200">{error}</div>
            </div>
          )}
          <div className="rounded-md shadow-sm -space-y-px">
            <div>
              <Input
                id="username"
                name="username"
                type="text"
                required
                value={username}
                onChange={(e) => setUsername(e.target.value)}
                placeholder="Username"
                className="appearance-none rounded-none relative block w-full px-3 py-2 border border-gray-300 placeholder-gray-500 text-gray-900 rounded-t-md focus:outline-none focus:ring-blue-500 focus:border-blue-500 focus:z-10 sm:text-sm dark:bg-gray-800 dark:border-gray-600 dark:text-white dark:placeholder-gray-400"
              />
            </div>
            <div>
              <Input
                id="password"
                name="password"
                type="password"
                required
                value={password}
                onChange={(e) => setPassword(e.target.value)}
                placeholder="Password"
                className="appearance-none rounded-none relative block w-full px-3 py-2 border border-gray-300 placeholder-gray-500 text-gray-900 rounded-b-md focus:outline-none focus:ring-blue-500 focus:border-blue-500 focus:z-10 sm:text-sm dark:bg-gray-800 dark:border-gray-600 dark:text-white dark:placeholder-gray-400"
              />
            </div>
          </div>

          <div>
            <Button
              type="submit"
              disabled={loading || !serviceStatus.available}
              className="group relative w-full flex justify-center py-2 px-4 border border-transparent text-sm font-medium rounded-md text-white bg-blue-600 hover:bg-blue-700 focus:outline-none focus:ring-2 focus:ring-offset-2 focus:ring-blue-500 disabled:opacity-50"
            >
              {loading ? 'Signing in...' : !serviceStatus.available ? 'Service Unavailable' : 'Sign in'}
            </Button>
          </div>
        </form>

        {!loadingProviders && (
          <div className="mt-6">
            {oauthProviders.length > 0 ? (
              <>
                <div className="relative">
                  <div className="absolute inset-0 flex items-center">
                    <div className="w-full border-t border-gray-300 dark:border-gray-600" />
                  </div>
                  <div className="relative flex justify-center text-sm">
                    <span className="px-2 bg-gray-50 dark:bg-gray-900 text-gray-500 dark:text-gray-400">
                      Or continue with
                    </span>
                  </div>
                </div>

                <div className="mt-6 grid grid-cols-3 gap-3">
                  {oauthProviders.map((provider) => {
                    const providerInfo = OAUTH_PROVIDERS[provider];
                    if (!providerInfo) return null;
                    
                    return (
                      <button
                        key={provider}
                        type="button"
                        onClick={() => handleOAuthLogin(provider)}
                        className={`w-full inline-flex justify-center py-2 px-4 border rounded-md shadow-sm text-sm font-medium ${providerInfo.color} focus:outline-none focus:ring-2 focus:ring-offset-2 focus:ring-blue-500 transition-colors`}
                      >
                        <span className="text-lg">{providerInfo.icon}</span>
                        <span className="ml-2 hidden sm:inline">{providerInfo.displayName}</span>
                      </button>
                    );
                  })}
                </div>
              </>
            ) : (
              <div className="text-center text-xs text-gray-500 dark:text-gray-400 mt-4">
                <p>No social login providers configured.</p>
                <p className="mt-1">Set OAuth environment variables to enable social login.</p>
              </div>
            )}
          </div>
        )}
      </div>
    </div>
  );
}

