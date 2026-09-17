import { useEffect, useState } from 'react';
import { useParams, Link } from 'react-router-dom';
import { getPoll } from '../api/polls';
import type { PollWithCounts, VoteCount } from '../api/polls';
import { useLivePoll } from '../hooks/useLivePoll';

const ResultsPage = () => {
  const { shareCode } = useParams<{ shareCode: string }>();
  const [poll, setPoll] = useState<PollWithCounts | null>(null);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState('');

  // Fetch poll on mount
  useEffect(() => {
    if (!shareCode) return;
    getPoll(shareCode)
      .then(({ data }) => setPoll(data))
      .catch(() => setError('Poll not found.'))
      .finally(() => setLoading(false));
  }, [shareCode]);

  // Connect to WebSocket for live updates
  const { counts: liveCounts, connected } = useLivePoll(shareCode!, poll?.id ?? null);

  // Merge: prefer live counts from WebSocket, fall back to REST snapshot
  const effectiveCounts: VoteCount[] = liveCounts.length > 0
    ? liveCounts
    : (poll?.counts ?? []);

  const totalVotes = effectiveCounts.reduce((s, c) => s + c.count, 0);

  const getCount = (optionId: string) =>
    effectiveCounts.find(c => c.optionId === optionId)?.count ?? 0;

  const getPercent = (optionId: string) => {
    if (totalVotes === 0) return 0;
    return Math.round((getCount(optionId) / totalVotes) * 100);
  };

  // Determine winner
  const maxCount = Math.max(...effectiveCounts.map(c => c.count), 0);

  const BAR_COLORS = [
    'linear-gradient(90deg, #8b5cf6, #06b6d4)',
    'linear-gradient(90deg, #06b6d4, #10b981)',
    'linear-gradient(90deg, #f59e0b, #ef4444)',
    'linear-gradient(90deg, #ec4899, #8b5cf6)',
    'linear-gradient(90deg, #10b981, #06b6d4)',
    'linear-gradient(90deg, #f97316, #eab308)',
    'linear-gradient(90deg, #6366f1, #8b5cf6)',
    'linear-gradient(90deg, #14b8a6, #3b82f6)',
    'linear-gradient(90deg, #a855f7, #ec4899)',
    'linear-gradient(90deg, #22d3ee, #818cf8)',
  ];

  if (loading) {
    return (
      <div className="page">
        <div className="spinner-wrapper">
          <div className="spinner" />
          <span>Loading results…</span>
        </div>
      </div>
    );
  }

  if (error || !poll) {
    return (
      <div className="page">
        <div className="empty-state">
          <div style={{ fontSize: '3rem', marginBottom: '1rem' }}>🔍</div>
          <h3>Poll not found</h3>
          <p>{error}</p>
        </div>
      </div>
    );
  }

  return (
    <div className="page fade-in">
      <div className="results-wrapper">
        {/* Live indicator */}
        <div className={`live-badge ${connected ? '' : 'offline'}`}>
          <span className="dot" />
          {connected ? 'Live results — updating in real time' : 'Reconnecting…'}
        </div>

        <div className="card">
          <div style={{ marginBottom: '0.5rem' }}>
            <span style={{ fontSize: '0.8rem', color: 'var(--text-muted)', fontWeight: 500, textTransform: 'uppercase', letterSpacing: '0.05em' }}>
              Poll Results
            </span>
          </div>
          <h1 className="poll-question" style={{ fontSize: '1.3rem', marginBottom: '0' }}>
            {poll.question}
          </h1>

          <div className="result-bar-wrapper">
            {poll.options.map((opt, i) => {
              const count = getCount(opt.id);
              const pct = getPercent(opt.id);
              const isWinner = count === maxCount && maxCount > 0;
              return (
                <div key={opt.id} className="result-bar-item" id={`result-${opt.id}`}>
                  <div className="result-bar-meta">
                    <span className="result-bar-label">
                      {isWinner && <span style={{ marginRight: '0.4rem' }}>👑</span>}
                      {opt.text}
                    </span>
                    <span className="result-bar-count">
                      {count} vote{count !== 1 ? 's' : ''}
                    </span>
                  </div>
                  <div className="result-bar-track">
                    <div
                      className="result-bar-fill"
                      style={{
                        width: `${pct}%`,
                        background: BAR_COLORS[i % BAR_COLORS.length],
                      }}
                    />
                  </div>
                  <span className="result-bar-percent">{pct}%</span>
                </div>
              );
            })}
          </div>

          <div className="total-votes">
            <strong>{totalVotes}</strong> total vote{totalVotes !== 1 ? 's' : ''} cast
          </div>
        </div>

        {/* Actions */}
        <div style={{ display: 'flex', gap: '0.75rem', marginTop: '1.5rem', justifyContent: 'center', flexWrap: 'wrap' }}>
          <Link to={`/poll/${shareCode}`} className="btn btn-secondary" id="back-to-vote-btn">
            ← Vote
          </Link>
          <button
            className="btn btn-ghost"
            onClick={() => {
              navigator.clipboard.writeText(window.location.href);
            }}
            id="share-results-btn"
          >
            Share results
          </button>
        </div>
      </div>
    </div>
  );
};

export default ResultsPage;
