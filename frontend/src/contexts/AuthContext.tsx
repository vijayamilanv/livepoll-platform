import { createContext, useContext, useState, useCallback } from 'react';
import type { ReactNode } from 'react';
import { signup as apiSignup, login as apiLogin } from '../api/auth';

interface User {
  id: string;
  email: string;
}

interface AuthContextType {
  user: User | null;
  token: string | null;
  isAuthenticated: boolean;
  login: (email: string, password: string) => Promise<void>;
  signup: (email: string, password: string) => Promise<void>;
  logout: () => void;
}

const AuthContext = createContext<AuthContextType | null>(null);

export const AuthProvider = ({ children }: { children: ReactNode }) => {
  const [user, setUser] = useState<User | null>(() => {
    const stored = localStorage.getItem('livepoll_user');
    return stored ? JSON.parse(stored) : null;
  });
  const [token, setToken] = useState<string | null>(() =>
    localStorage.getItem('livepoll_token')
  );

  const persist = (tok: string, u: User) => {
    localStorage.setItem('livepoll_token', tok);
    localStorage.setItem('livepoll_user', JSON.stringify(u));
    setToken(tok);
    setUser(u);
  };

  const login = useCallback(async (email: string, password: string) => {
    const { data } = await apiLogin(email, password);
    persist(data.token, data.user);
  }, []);

  const signup = useCallback(async (email: string, password: string) => {
    const { data } = await apiSignup(email, password);
    persist(data.token, data.user);
  }, []);

  const logout = useCallback(() => {
    localStorage.removeItem('livepoll_token');
    localStorage.removeItem('livepoll_user');
    setToken(null);
    setUser(null);
  }, []);

  return (
    <AuthContext.Provider
      value={{ user, token, isAuthenticated: !!token, login, signup, logout }}
    >
      {children}
    </AuthContext.Provider>
  );
};

export const useAuth = () => {
  const ctx = useContext(AuthContext);
  if (!ctx) throw new Error('useAuth must be used within AuthProvider');
  return ctx;
};
