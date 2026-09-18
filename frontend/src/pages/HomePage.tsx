import { Link } from 'react-router-dom';
import { useAuth } from '../contexts/AuthContext';

const HomePage = () => {
  const { isAuthenticated } = useAuth();

  return (
    <div className="page fade-in" style={{ textAlign: 'center', gap: '2rem' }}>
      {/* Hero */}
      <div style={{ maxWidth: '680px' }}>
        <div className="hero-badge">
          <span style={{ width: 6, height: 6, borderRadius: '50%', background: 'var(--accent)', display: 'inline-block', animation: 'pulse 1.5s infinite' }} />
          REAL-TIME AUDIENCE POLLING
        </div>

        <h1 className="hero-heading">
          Create polls.<br />Watch results live.
        </h1>

        <p style={{ fontSize: '1.15rem', color: 'var(--text-secondary)', maxWidth: '500px', margin: '0 auto 2.5rem', lineHeight: 1.7 }}>
          Build a poll in seconds, share a link, and watch votes roll in with <strong style={{ color: 'var(--text-primary)' }}>real-time live results</strong> — no refresh needed.
        </p>

        <div style={{ display: 'flex', gap: '1rem', justifyContent: 'center', flexWrap: 'wrap' }}>
          {isAuthenticated ? (
            <Link to="/create" className="btn btn-primary btn-lg" id="hero-create-btn">
              ✦ Create a Poll
            </Link>
          ) : (
            <>
              <Link to="/signup" className="btn btn-primary btn-lg" id="hero-signup-btn">
                Get started free →
              </Link>
              <Link to="/login" className="btn btn-secondary btn-lg" id="hero-login-btn">
                Log in
              </Link>
            </>
          )}
        </div>
      </div>

      {/* Feature cards */}
      <div style={{ display: 'grid', gridTemplateColumns: 'repeat(auto-fit, minmax(200px, 1fr))', gap: '1rem', maxWidth: '800px', width: '100%', marginTop: '2rem' }}>
        {[
          { icon: '⚡', title: 'Instant Results', desc: 'WebSocket-powered — updates the moment votes land.' },
          { icon: '🔗', title: 'Shareable Link', desc: 'One unique URL/code per poll. No app install needed.' },
          { icon: '🔒', title: 'No Login to Vote', desc: 'Audience votes instantly — only creators need accounts.' },
          { icon: '📊', title: 'Live Bar Charts', desc: 'Animated results with percentages and vote counts.' },
        ].map(f => (
          <div key={f.title} className="card" style={{ textAlign: 'left', padding: '1.5rem' }}>
            <div style={{ fontSize: '2rem', marginBottom: '0.75rem' }}>{f.icon}</div>
            <h3 style={{ fontWeight: 700, marginBottom: '0.35rem', color: 'var(--text-primary)' }}>{f.title}</h3>
            <p style={{ color: 'var(--text-secondary)', fontSize: '0.875rem', lineHeight: 1.6 }}>{f.desc}</p>
          </div>
        ))}
      </div>
    </div>
  );
};

export default HomePage;
