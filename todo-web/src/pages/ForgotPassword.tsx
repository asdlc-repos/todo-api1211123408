import { useState } from 'react';
import type { FormEvent } from 'react';
import { Link } from 'react-router-dom';
import { auth, ApiError } from '../api';

export function ForgotPassword() {
  const [email, setEmail] = useState('');
  const [submitting, setSubmitting] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const [sent, setSent] = useState(false);

  const onSubmit = async (e: FormEvent) => {
    e.preventDefault();
    setError(null);
    setSubmitting(true);
    try {
      await auth.requestReset(email.trim());
      setSent(true);
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
      <h1>Reset password</h1>
      <p className="muted">
        Enter your email and we'll send a reset link if an account exists.
      </p>
      {error && <div className="error">{error}</div>}
      {sent ? (
        <div className="success">
          If an account exists for that email, a password reset link has been sent.
        </div>
      ) : (
        <form onSubmit={onSubmit}>
          <div className="field">
            <label htmlFor="email">Email</label>
            <input
              id="email"
              type="email"
              autoComplete="email"
              required
              value={email}
              onChange={(e) => setEmail(e.target.value)}
            />
          </div>
          <button type="submit" disabled={submitting}>
            {submitting ? 'Sending…' : 'Send reset link'}
          </button>
        </form>
      )}
      <div className="auth-footer">
        <Link to="/login">Back to sign in</Link>
        <Link to="/reset-password">Have a token?</Link>
      </div>
    </div>
  );
}
