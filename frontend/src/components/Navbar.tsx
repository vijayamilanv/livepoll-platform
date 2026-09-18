import { Link, useNavigate } from 'react-router-dom';
import { useAuth } from '../contexts/AuthContext';
import { useTheme } from '../contexts/ThemeContext';

const Navbar = () => {
  const { isAuthenticated, user, logout } = useAuth();
  const { theme, toggleTheme } = useTheme();
  const navigate = useNavigate();

  const handleLogout = () => {
    logout();
    navigate('/login');
  };

  return (
    <nav className="navbar">
      <Link to="/" className="navbar-brand">
        <span className="dot" />
        LivePoll
      </Link>
      <div className="navbar-actions">
        <button
          onClick={toggleTheme}
          className="theme-toggle-btn"
          title={`Switch to ${theme === 'dark' ? 'light' : 'dark'} mode`}
          aria-label="Toggle Theme"
          id="theme-toggle-btn"
        >
          {theme === 'dark' ? '☀️' : '🌙'}
        </button>

        {isAuthenticated ? (
          <>
            <span style={{ color: 'var(--text-muted)', fontSize: '0.85rem' }}>
              {user?.email}
            </span>
            <Link to="/my-polls" className="btn btn-secondary btn-sm">
              My Polls
            </Link>
            <Link to="/create" className="btn btn-primary btn-sm">
              + Create Poll
            </Link>
            <button onClick={handleLogout} className="btn btn-ghost btn-sm">
              Logout
            </button>
          </>
        ) : (
          <>
            <Link to="/login" className="btn btn-secondary btn-sm">Log in</Link>
            <Link to="/signup" className="btn btn-primary btn-sm">Sign up</Link>
          </>
        )}
      </div>
    </nav>
  );
};

export default Navbar;
