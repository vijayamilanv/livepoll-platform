import axios from 'axios';

const api = axios.create({
  baseURL: import.meta.env.VITE_API_URL || 'http://localhost:8080',
  headers: { 'Content-Type': 'application/json' },
});

// Attach JWT from memory store on every request
api.interceptors.request.use((config) => {
  const token = localStorage.getItem('livepoll_token'); // see README for token storage rationale
  if (token) config.headers.Authorization = `Bearer ${token}`;
  return config;
});

export default api;
