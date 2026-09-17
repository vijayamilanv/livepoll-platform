import { useEffect, useState } from 'react';
import { useParams, useNavigate } from 'react-router-dom';
import toast from 'react-hot-toast';
import { getPoll, castVote } from '../api/polls';
import type { PollWithCounts } from '../api/polls';
import { useLivePoll } from '../hooks/useLivePoll';

const VotePage = () => {
  const { shareCode } = useParams<{ shareCode: string }>();
  const navigate = useNavigate();
  const [poll, setPoll] = useState<PollWithCounts | null>(null);
  const [loading, setLoading] = useState(true);
  const [selected, setSelected] = useState<string | null>(null);
  const [voting, setVoting] = useState(false);
  const [voted, setVoted] = useState(false);
  const [error, setError] = useState('');
  const { connected } = useLivePoll(shareCode!, poll?.id ?? null);

  useEffect(() => {
    if (!shareCode) return;
    setLoading(true);
    getPoll(shareCode)
      .then(({ data }) => setPoll(data))
      .catch(() => setError('Poll not found or has been removed.'))
      .finally(() => setLoading(false));
  }, [shareCode]);

  const handleVote = async () => {
    if (!selected || !shareCode) return;
    setVoting(true);
    try {
      await castVote(shareCode, selected);
      toast.success('Vote cast! 🗳️');
      setVoted(true);
      // Navigate to live results view
      navigate(`/poll/${shareCode}/results`);
    } catch (err: any) {
      const msg = err?.response?.data?.error || 'Failed to cast vote';
      if (msg.includes('already voted')) {
        toast('You already voted on this poll.', { icon: 'ℹ️' });
        navigate(`/poll/${shareCode}/results`);
      } else {
        toast.error(msg);
      }
    } finally {
      setVoting(false);
    }
  };

  if (loading) {
    return (
      <div className="page">
        <div className="spinner-wrapper">
          <div className="spinner" />
          <span>Loading poll…</span>
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
          <p>{error || 'This poll may have been removed.'}</p>
        </div>
      </div>
    );
  }

  return (
    <div className="page fade-in">
      <div className="vote-wrapper">
        {/* Live status badge */}
        <div className={`live-badge ${connected ? '' : 'offline'}`}>
          <span className="dot" />
          {connected ? 'Live' : 'Reconnecting…'}
        </div>

        <div className="card">
          <h1 className="poll-question">{poll.question}</h1>

          <div className="options-grid">
            {poll.options.map(opt => (
              <button
                key={opt.id}
                id={`vote-option-${opt.id}`}
                className={`vote-option ${selected === opt.id ? 'selected' : ''}`}
                onClick={() => setSelected(opt.id)}
                disabled={voted || voting}
              >
                <span className="option-dot" />
                {opt.text}
              </button>
            ))}
          </div>

          <div style={{ marginTop: '1.5rem', display: 'flex', gap: '0.75rem', flexWrap: 'wrap' }}>
            <button
              className="btn btn-primary btn-lg"
              onClick={handleVote}
              disabled={!selected || voting || voted}
              id="submit-vote-btn"
              style={{ flex: 1 }}
            >
              {voting ? 'Submitting…' : 'Submit Vote'}
            </button>
            <a
              href={`/poll/${shareCode}/results`}
              className="btn btn-secondary btn-lg"
              id="view-results-btn"
            >
              View Results
            </a>
          </div>

          <p style={{ marginTop: '1rem', fontSize: '0.8rem', color: 'var(--text-muted)', textAlign: 'center' }}>
            Select an option and click Submit. Voting is open to everyone.
          </p>
        </div>
      </div>
    </div>
  );
};

export default VotePage;
