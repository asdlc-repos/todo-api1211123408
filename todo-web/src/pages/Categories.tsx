import { useCallback, useEffect, useState } from 'react';
import type { FormEvent } from 'react';
import { useNavigate } from 'react-router-dom';
import { ApiError, categories as categoriesApi } from '../api';
import type { Category } from '../types';
import { useAuth } from '../auth';

export function CategoriesPage() {
  const navigate = useNavigate();
  const { clearSession } = useAuth();

  const [categories, setCategories] = useState<Category[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);
  const [formError, setFormError] = useState<string | null>(null);
  const [name, setName] = useState('');
  const [submitting, setSubmitting] = useState(false);
  const [editingId, setEditingId] = useState<string | null>(null);
  const [editingName, setEditingName] = useState('');

  const handleUnauthorized = useCallback(() => {
    clearSession();
    navigate('/login', { replace: true });
  }, [clearSession, navigate]);

  const applyError = useCallback(
    (err: unknown, setter: (msg: string) => void) => {
      if (err instanceof ApiError) {
        if (err.status === 401) {
          handleUnauthorized();
          return;
        }
        setter(err.message);
      } else {
        setter('Unable to reach the server.');
      }
    },
    [handleUnauthorized],
  );

  const load = useCallback(async () => {
    setLoading(true);
    setError(null);
    try {
      const list = await categoriesApi.list();
      setCategories(Array.isArray(list) ? list : []);
    } catch (err) {
      applyError(err, setError);
    } finally {
      setLoading(false);
    }
  }, [applyError]);

  useEffect(() => {
    load();
  }, [load]);

  const submitCreate = async (e: FormEvent) => {
    e.preventDefault();
    setFormError(null);
    const trimmed = name.trim();
    if (!trimmed) {
      setFormError('Name is required.');
      return;
    }
    if (trimmed.length > 50) {
      setFormError('Name must be 50 characters or fewer.');
      return;
    }
    setSubmitting(true);
    try {
      await categoriesApi.create(trimmed);
      setName('');
      await load();
    } catch (err) {
      if (err instanceof ApiError && err.status === 409) {
        setFormError('A category with that name already exists.');
      } else {
        applyError(err, setFormError);
      }
    } finally {
      setSubmitting(false);
    }
  };

  const beginEdit = (c: Category) => {
    setEditingId(c.id);
    setEditingName(c.name);
  };

  const cancelEdit = () => {
    setEditingId(null);
    setEditingName('');
  };

  const saveEdit = async (c: Category) => {
    const trimmed = editingName.trim();
    if (!trimmed) {
      setError('Name is required.');
      return;
    }
    try {
      await categoriesApi.update(c.id, trimmed);
      cancelEdit();
      await load();
    } catch (err) {
      if (err instanceof ApiError && err.status === 409) {
        setError('A category with that name already exists.');
      } else {
        applyError(err, setError);
      }
    }
  };

  const deleteCategory = async (c: Category) => {
    if (!confirm(`Delete category "${c.name}"? Todos will be unassigned.`)) return;
    try {
      await categoriesApi.remove(c.id);
      await load();
    } catch (err) {
      applyError(err, setError);
    }
  };

  return (
    <div className="container">
      <div className="card">
        <h2>New category</h2>
        {formError && <div className="error">{formError}</div>}
        <form onSubmit={submitCreate}>
          <div className="row">
            <div className="field" style={{ flex: 1, marginBottom: 0 }}>
              <label htmlFor="name">Name</label>
              <input
                id="name"
                type="text"
                maxLength={50}
                value={name}
                onChange={(e) => setName(e.target.value)}
                required
              />
            </div>
            <button type="submit" disabled={submitting} style={{ alignSelf: 'flex-end' }}>
              {submitting ? 'Adding…' : 'Add category'}
            </button>
          </div>
        </form>
      </div>

      <div className="card">
        <h2>Categories</h2>
        {error && <div className="error">{error}</div>}
        {loading ? (
          <div className="empty-state">Loading…</div>
        ) : categories.length === 0 ? (
          <div className="empty-state">No categories yet. Create one above.</div>
        ) : (
          <ul className="categories-list">
            {categories.map((c) => (
              <li key={c.id} className="category-row">
                {editingId === c.id ? (
                  <div className="inline-edit">
                    <input
                      autoFocus
                      type="text"
                      maxLength={50}
                      value={editingName}
                      onChange={(e) => setEditingName(e.target.value)}
                    />
                    <button onClick={() => saveEdit(c)}>Save</button>
                    <button className="secondary" onClick={cancelEdit}>
                      Cancel
                    </button>
                  </div>
                ) : (
                  <>
                    <span style={{ flex: 1 }}>{c.name}</span>
                    <button className="secondary" onClick={() => beginEdit(c)}>
                      Rename
                    </button>
                    <button className="danger" onClick={() => deleteCategory(c)}>
                      Delete
                    </button>
                  </>
                )}
              </li>
            ))}
          </ul>
        )}
      </div>
    </div>
  );
}
