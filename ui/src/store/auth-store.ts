/**
 * Authentication Store
 * Manages user authentication state
 */

import { create } from 'zustand';
import { persist } from 'zustand/middleware';

interface AuthState {
  token: string | null;
  isAuthenticated: boolean;
  setToken: (token: string | null) => void;
  logout: () => void;
}

export const useAuthStore = create<AuthState>()(
  persist(
    (set) => ({
      token: null,
      isAuthenticated: false,
      setToken: (token) => {
        if (token) {
          localStorage.setItem('auth_token', token);
          set({ token, isAuthenticated: true });
        } else {
          localStorage.removeItem('auth_token');
          set({ token: null, isAuthenticated: false });
        }
      },
      logout: () => {
        localStorage.removeItem('auth_token');
        localStorage.removeItem('auth-storage');
        set({ token: null, isAuthenticated: false });
        // Navigate will be handled by App.tsx ProtectedRoute
      },
    }),
    {
      name: 'auth-storage',
    }
  )
);
