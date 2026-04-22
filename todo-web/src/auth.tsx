import { createContext, useCallback, useContext, useMemo, useState } from 'react';
import type { ReactNode } from 'react';
import { auth } from './api';

const STORAGE_KEY = 'todo-web:auth';

interface AuthState {
  email: string | null;
}

interface AuthContextValue {
  email: string | null;
  isAuthenticated: boolean;
  login: (email: string, password: string) => Promise<void>;
  register: (email: string, password: string) => Promise<void>;
  logout: () => Promise<void>;
  clearSession: () => void;
}

const AuthContext = createContext<AuthContextValue | undefined>(undefined);

function readStored(): AuthState {
  try {
    const raw = localStorage.getItem(STORAGE_KEY);
    if (!raw) return { email: null };
    const parsed = JSON.parse(raw) as AuthState;
    return { email: parsed.email ?? null };
  } catch {
    return { email: null };
  }
}

function writeStored(state: AuthState) {
  try {
    if (state.email) {
      localStorage.setItem(STORAGE_KEY, JSON.stringify(state));
    } else {
      localStorage.removeItem(STORAGE_KEY);
    }
  } catch {
    // ignore storage errors
  }
}

export function AuthProvider({ children }: { children: ReactNode }) {
  const [state, setState] = useState<AuthState>(() => readStored());

  const setEmail = useCallback((email: string | null) => {
    const next = { email };
    writeStored(next);
    setState(next);
  }, []);

  const login = useCallback(
    async (email: string, password: string) => {
      await auth.login(email, password);
      setEmail(email);
    },
    [setEmail],
  );

  const register = useCallback(
    async (email: string, password: string) => {
      await auth.register(email, password);
      await auth.login(email, password);
      setEmail(email);
    },
    [setEmail],
  );

  const logout = useCallback(async () => {
    try {
      await auth.logout();
    } catch {
      // ignore — clear local state regardless
    }
    setEmail(null);
  }, [setEmail]);

  const clearSession = useCallback(() => {
    setEmail(null);
  }, [setEmail]);

  const value = useMemo<AuthContextValue>(
    () => ({
      email: state.email,
      isAuthenticated: !!state.email,
      login,
      register,
      logout,
      clearSession,
    }),
    [state.email, login, register, logout, clearSession],
  );

  return <AuthContext.Provider value={value}>{children}</AuthContext.Provider>;
}

export function useAuth(): AuthContextValue {
  const ctx = useContext(AuthContext);
  if (!ctx) throw new Error('useAuth must be used within AuthProvider');
  return ctx;
}
