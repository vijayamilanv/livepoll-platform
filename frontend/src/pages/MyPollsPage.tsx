import { useEffect, useState } from 'react';
import { Link } from 'react-router-dom';
import toast from 'react-hot-toast';
import { getMyPolls } from '../api/polls';
import type { PollWithCounts } from '../api/polls';

const MyPollsPage = () => {
  const [polls, setPolls] = useState<PollWithCounts[]>([]);
  const [loading, setLoading] = useState(true);

  useEffect(() => {
    getMyPolls()
      .then(({ data }) => setPolls(data.polls))
      .catch(() => toast.error('Failed to load your polls'))
      .finally(() => setLoading(false));
  }, []);

  const totalVotesForPoll = (poll: PollWithCounts) =>
    poll.counts?.reduce((s, c) => s + c.count, 0) ?? 0;

  const copyLink = (shareCode: string) => {
    const url = `${window.location.origin}/poll/${shareCode}`;
    navigator.clipboard.writeText(url);
    toast.success('Link copied!');
  };

  if (loading) {
    return (
      <div className="page">
        <div className="spinner-wrapper">
          <div className="spinner" />
          <span>Loading your polls…</span>
        </div>
      </div>
    );
  }

  return (
    <div className="page fade-in" style={{ alignItems: 'flex-start', paddingTop: 0 }}>
      <div className="dashboard-wrapper">
        <div className="page-header">
          <h1>My Polls</h1>
          <p>Manage and monitor your polls. Click any poll to view live results.</p>
        </div>

        <div style={{ display: 'flex', justifyContent: 'flex-end', marginBottom: '1.5rem' }}>
          <Link to="/create" className="btn btn-primary" id="create-new-poll-btn">
            + Create New Poll
          </Link>
        </div>

        {polls.length === 0 ? (
          <div className="empty-state card">
            <div style={{ fontSize: '3rem', marginBottom: '1rem' }}>📊</div>
            <h3>No polls yet</h3>
            <p style={{ marginBottom: '1.5rem' }}>Create your first poll and share it with the world.</p>
            <Link to="/create" className="btn btn-primary">Create a Poll</Link>
          </div>
        ) : (
          <div className="polls-grid">
            {polls.map(poll => (
              <div key={poll.id} className="poll-card fade-in-up">
                <p className="poll-card-question">{poll.question}</p>

                <div className="poll-card-meta">
                  <span className="meta-chip">
                    🔑 {poll.shareCode}
                  </span>
                  <span className="meta-chip">
                    🗳️ {totalVotesForPoll(poll)} votes
                  </span>
                  <span className="meta-chip">
                    📝 {poll.options.length} options
                  </span>
                  <span className="meta-chip">
                    {new Date(poll.createdAt).toLocaleDateString()}
                  </span>
                </div>

                <div className="poll-card-actions">
                  <Link
                    to={`/poll/${poll.shareCode}`}
                    className="btn btn-secondary btn-sm"
                    id={`vote-link-${poll.shareCode}`}
                  >
                    Vote Page
                  </Link>
                  <Link
                    to={`/poll/${poll.shareCode}/results`}
                    className="btn btn-primary btn-sm"
                    id={`results-link-${poll.shareCode}`}
                  >
                    Live Results
                  </Link>
                  <button
                    className="btn btn-ghost btn-sm"
                    onClick={() => copyLink(poll.shareCode)}
                    id={`copy-link-${poll.shareCode}`}
                  >
                    Copy Link
                  </button>
                </div>
              </div>
            ))}
          </div>
        )}
      </div>
    </div>
  );
};

export default MyPollsPage;
