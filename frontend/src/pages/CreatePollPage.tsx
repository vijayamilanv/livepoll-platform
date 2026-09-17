import React, { useState } from 'react';
import toast from 'react-hot-toast';
import { createPoll } from '../api/polls';

const MAX_OPTIONS = 10;
const MIN_OPTIONS = 2;
const MAX_QUESTION_LEN = 500;
const MAX_OPTION_LEN = 200;

const CreatePollPage = () => {
  const [question, setQuestion] = useState('');
  const [options, setOptions] = useState(['', '']);
  const [loading, setLoading] = useState(false);
  const [created, setCreated] = useState<{ shareCode: string } | null>(null);
  const [errors, setErrors] = useState<Record<string, string>>({});
  const [copied, setCopied] = useState(false);

  const shareUrl = created
    ? `${window.location.origin}/poll/${created.shareCode}`
    : '';

  const validate = () => {
    const e: Record<string, string> = {};
    if (!question.trim()) e.question = 'Question is required';
    else if (question.length > MAX_QUESTION_LEN) e.question = `Max ${MAX_QUESTION_LEN} characters`;
    options.forEach((o, i) => {
      if (!o.trim()) e[`option_${i}`] = 'Option text required';
      else if (o.length > MAX_OPTION_LEN) e[`option_${i}`] = `Max ${MAX_OPTION_LEN} chars`;
    });
    return e;
  };

  const handleOptionChange = (i: number, val: string) => {
    setOptions(prev => prev.map((o, idx) => idx === i ? val : o));
  };

  const addOption = () => {
    if (options.length < MAX_OPTIONS) setOptions(prev => [...prev, '']);
  };

  const removeOption = (i: number) => {
    if (options.length > MIN_OPTIONS) setOptions(prev => prev.filter((_, idx) => idx !== i));
  };

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    const errs = validate();
    if (Object.keys(errs).length > 0) { setErrors(errs); return; }
    setErrors({});
    setLoading(true);
    try {
      const { data } = await createPoll(question.trim(), options.map(o => o.trim()));
      setCreated({ shareCode: data.shareCode });
      toast.success('Poll created! 🎉');
    } catch (err: any) {
      toast.error(err?.response?.data?.error || 'Failed to create poll');
    } finally {
      setLoading(false);
    }
  };

  const handleCopy = async () => {
    await navigator.clipboard.writeText(shareUrl);
    setCopied(true);
    toast.success('Link copied!');
    setTimeout(() => setCopied(false), 2000);
  };

  const handleReset = () => {
    setCreated(null);
    setQuestion('');
    setOptions(['', '']);
  };

  if (created) {
    return (
      <div className="page fade-in">
        <div className="create-poll-wrapper">
          <div className="share-box">
            <div style={{ fontSize: '3rem', marginBottom: '0.5rem' }}>🎉</div>
            <h2 style={{ color: 'var(--text-primary)', marginBottom: '0.5rem', fontSize: '1.4rem', fontWeight: 700 }}>
              Poll Created!
            </h2>
            <p style={{ color: 'var(--text-secondary)', marginBottom: '1.5rem' }}>
              Share this code or link with your audience
            </p>
            <div className="share-code">{created.shareCode}</div>
            <div className="share-url" style={{ marginTop: '1.5rem' }}>
              <code>{shareUrl}</code>
              <button
                className="btn btn-primary btn-sm"
                onClick={handleCopy}
                id="copy-link-btn"
              >
                {copied ? '✓ Copied' : 'Copy'}
              </button>
            </div>
            <div style={{ display: 'flex', gap: '0.75rem', justifyContent: 'center', marginTop: '1.5rem', flexWrap: 'wrap' }}>
              <a
                href={`/poll/${created.shareCode}`}
                className="btn btn-primary"
                id="view-poll-btn"
              >
                Open Poll →
              </a>
              <button onClick={handleReset} className="btn btn-secondary" id="create-another-btn">
                + Create Another
              </button>
            </div>
          </div>
        </div>
      </div>
    );
  }

  return (
    <div className="page fade-in">
      <div className="create-poll-wrapper">
        <div className="page-header">
          <h1>Create a Poll</h1>
          <p>Craft your question and let the crowd decide — results update live.</p>
        </div>

        <div className="card">
          <form id="create-poll-form" onSubmit={handleSubmit}>
            {/* Question */}
            <div className="form-group" style={{ marginBottom: '2rem' }}>
              <label className="form-label" htmlFor="poll-question">
                Your question <span style={{ color: 'var(--text-muted)', fontSize: '0.8rem' }}>
                  ({question.length}/{MAX_QUESTION_LEN})
                </span>
              </label>
              <textarea
                id="poll-question"
                className="form-input"
                placeholder="e.g. Which framework do you prefer for your next project?"
                value={question}
                onChange={e => setQuestion(e.target.value)}
                rows={3}
                style={{ resize: 'vertical' }}
                maxLength={MAX_QUESTION_LEN}
              />
              {errors.question && <span className="form-error">{errors.question}</span>}
            </div>

            {/* Options */}
            <div className="form-group" style={{ marginBottom: '1.5rem' }}>
              <label className="form-label">
                Answer options <span style={{ color: 'var(--text-muted)', fontSize: '0.8rem' }}>
                  ({options.length}/{MAX_OPTIONS})
                </span>
              </label>
              <div className="options-list">
                {options.map((opt, i) => (
                  <div key={i} className="option-row">
                    <span className="option-number">{i + 1}</span>
                    <input
                      id={`option-${i}`}
                      type="text"
                      className="form-input"
                      placeholder={`Option ${i + 1}`}
                      value={opt}
                      onChange={e => handleOptionChange(i, e.target.value)}
                      maxLength={MAX_OPTION_LEN}
                    />
                    {options.length > MIN_OPTIONS && (
                      <button
                        type="button"
                        className="btn btn-ghost btn-sm"
                        onClick={() => removeOption(i)}
                        title="Remove option"
                        id={`remove-option-${i}`}
                        style={{ color: 'var(--error)', flexShrink: 0 }}
                      >
                        ✕
                      </button>
                    )}
                  </div>
                ))}
                {errors.option_0 && <span className="form-error">All options must have text</span>}
              </div>
            </div>

            <div style={{ display: 'flex', gap: '0.75rem', flexWrap: 'wrap', marginBottom: '1.5rem' }}>
              {options.length < MAX_OPTIONS && (
                <button
                  type="button"
                  className="btn btn-secondary btn-sm"
                  onClick={addOption}
                  id="add-option-btn"
                >
                  + Add option
                </button>
              )}
            </div>

            <button
              type="submit"
              className="btn btn-primary btn-full btn-lg"
              disabled={loading}
              id="create-poll-submit"
            >
              {loading ? 'Creating poll…' : '✦ Create Poll & Get Link'}
            </button>
          </form>
        </div>
      </div>
    </div>
  );
};

export default CreatePollPage;
