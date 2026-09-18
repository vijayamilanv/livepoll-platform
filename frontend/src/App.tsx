import { BrowserRouter, Routes, Route, Navigate } from 'react-router-dom';
import { Toaster } from 'react-hot-toast';
import { AuthProvider } from './contexts/AuthContext';
import { ThemeProvider } from './contexts/ThemeContext';
import AuthGuard from './components/AuthGuard';
import Navbar from './components/Navbar';
import HomePage from './pages/HomePage';
import LoginPage from './pages/LoginPage';
import SignupPage from './pages/SignupPage';
import CreatePollPage from './pages/CreatePollPage';
import VotePage from './pages/VotePage';
import ResultsPage from './pages/ResultsPage';
import MyPollsPage from './pages/MyPollsPage';

function App() {
  return (
    <ThemeProvider>
      <AuthProvider>
        <BrowserRouter>
          <Navbar />
          <Routes>
            <Route path="/" element={<HomePage />} />
            <Route path="/login" element={<LoginPage />} />
            <Route path="/signup" element={<SignupPage />} />
            <Route
              path="/create"
              element={
                <AuthGuard>
                  <CreatePollPage />
                </AuthGuard>
              }
            />
            <Route
              path="/my-polls"
              element={
                <AuthGuard>
                  <MyPollsPage />
                </AuthGuard>
              }
            />
            <Route path="/poll/:shareCode" element={<VotePage />} />
            <Route path="/poll/:shareCode/results" element={<ResultsPage />} />
            {/* Catch-all */}
            <Route path="*" element={<Navigate to="/" replace />} />
          </Routes>
          <Toaster
            position="top-right"
            toastOptions={{
              style: {
                background: 'var(--bg-card)',
                color: 'var(--text-primary)',
                border: '1px solid var(--border)',
                borderRadius: '12px',
                backdropFilter: 'blur(16px)',
                fontFamily: 'inherit',
                fontSize: '0.9rem',
                fontWeight: 500,
              },
              success: { iconTheme: { primary: 'var(--success)', secondary: '#ffffff' } },
              error: { iconTheme: { primary: 'var(--error)', secondary: '#ffffff' } },
            }}
          />
        </BrowserRouter>
      </AuthProvider>
    </ThemeProvider>
  );
}

export default App;
