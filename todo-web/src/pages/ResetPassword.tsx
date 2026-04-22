import { useEffect, useState } from 'react';
import type { FormEvent } from 'react';
import { Link, useNavigate, useSearchParams } from 'react-router-dom';
import { auth, ApiError } from '../api';

export function ResetPassword() {
  const [searchParams] = useSearchParams();
  const navigate = useNavigate();

  const [token, setToken] = useState('');
  const [password, setPassword] = useState('');
  const [confirm, setConfirm] = useState('');
  const [error, setError] = useState<string | null>(null);
  const [submitting, setSubmitting] = useState(false);
  const [done, setDone] = useState(false);

  useEffect(() => {
    const qsToken = searchParams.get('token');
    if (qsToken) setToken(qsToken);
  }, [searchParams]);

  const onSubmit = async (e: FormEvent) => {
    e.preventDefault();
    setError(null);

    if (password !== confirm) {
      setError('Passwords do not match.');
      return;
    }
    if (password.length < 8) {
      setError('Password must be at least 8 characters.');
      return;
    }

    setSubmitting(true);
    try {
      await auth.confirmReset(token.trim(), password);
      setDone(true);
      setTimeout(() => navigate('/login', { replace: true }), 1500);
    } catch (err) {
      if (err instanceof ApiError) {
        setError(err.message);
      } else {
        setError('Unable to reach the server.');
      }
    } finally {
      setSubmitting(false);
    }
  };

  return (
    <div className="auth-container">
      <h1>Set a new password</h1>
      {error && <div className="error">{error}</div>}
      {done ? (
        <div className="success">Password updated. Redirecting to sign in…</div>
      ) : (
        <form onSubmit={onSubmit}>
          <div className="field">
            <label htmlFor="token">Reset token</label>
            <input
              id="token"
              type="text"
              required
              value={token}
              onChange={(e) => setToken(e.target.value)}
            />
          </div>
          <div className="field">
            <label htmlFor="password">New password</label>
            <input
              id="password"
              type="password"
              autoComplete="new-password"
              required
              minLength={8}
              value={password}
              onChange={(e) => setPassword(e.target.value)}
            />
          </div>
          <div className="field">
            <label htmlFor="confirm">Confirm new password</label>
            <input
              id="confirm"
              type="password"
              autoComplete="new-password"
              required
              minLength={8}
              value={confirm}
              onChange={(e) => setConfirm(e.target.value)}
            />
          </div>
          <button type="submit" disabled={submitting}>
            {submitting ? 'Updating…' : 'Update password'}
          </button>
        </form>
      )}
      <div className="auth-footer">
        <Link to="/login">Back to sign in</Link>
      </div>
    </div>
  );
}
