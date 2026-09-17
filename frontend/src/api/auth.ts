import api from './axios';

export interface AuthResponse {
  token: string;
  user: { id: string; email: string };
}

export const signup = (email: string, password: string) =>
  api.post<AuthResponse>('/auth/signup', { email, password });

export const login = (email: string, password: string) =>
  api.post<AuthResponse>('/auth/login', { email, password });
