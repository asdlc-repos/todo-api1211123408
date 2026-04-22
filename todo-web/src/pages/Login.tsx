import { useState } from 'react';
import type { FormEvent } from 'react';
import { Link, useLocation, useNavigate } from 'react-router-dom';
import { useAuth } from '../auth';
import { ApiError } from '../api';

interface LocationState {
  from?: string;
}

export function Login() {
  const { login } = useAuth();
  const navigate = useNavigate();
  const location = useLocation();
  const from = (location.state as LocationState | null)?.from || '/todos';

  const [email, setEmail] = useState('');
  const [password, setPassword] = useState('');
  const [error, setError] = useState<string | null>(null);
  const [submitting, setSubmitting] = useState(false);

  const onSubmit = async (e: FormEvent) => {
    e.preventDefault();
    setError(null);
    setSubmitting(true);
    try {
      await login(email.trim(), password);
      navigate(from, { replace: true });
    } catch (err) {
      if (err instanceof ApiError) {
        if (err.status === 423) {
          setError('Account locked. Try again in 15 minutes.');
        } else if (err.status === 401) {
          setError('Invalid email or password.');
        } else {
          setError(err.message);
        }
      } else {
        setError('Unable to reach the server.');
      }
    } finally {
      setSubmitting(false);
    }
  };

  return (
    <div className="auth-container">
      <h1>Sign in</h1>
      <p className="muted">Welcome back. Enter your credentials to continue.</p>
      {error && <div className="error">{error}</div>}
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
        <div className="field">
          <label htmlFor="password">Password</label>
          <input
            id="password"
            type="password"
            autoComplete="current-password"
            required
            value={password}
            onChange={(e) => setPassword(e.target.value)}
          />
        </div>
        <button type="submit" disabled={submitting}>
          {submitting ? 'Signing in…' : 'Sign in'}
        </button>
      </form>
      <div className="auth-footer">
        <Link to="/forgot-password">Forgot password?</Link>
        <span>
          New here? <Link to="/register">Create account</Link>
        </span>
      </div>
    </div>
  );
}
