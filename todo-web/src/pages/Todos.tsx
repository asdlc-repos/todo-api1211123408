import { useCallback, useEffect, useMemo, useState } from 'react';
import type { FormEvent } from 'react';
import { useNavigate } from 'react-router-dom';
import { ApiError, categories as categoriesApi, todos as todosApi } from '../api';
import type { Category, Todo, TodoFilters, TodoInput } from '../types';
import { useAuth } from '../auth';

interface FormState {
  id: string | null;
  title: string;
  description: string;
  dueDate: string;
  categoryId: string;
  version: number;
}

const EMPTY_FORM: FormState = {
  id: null,
  title: '',
  description: '',
  dueDate: '',
  categoryId: '',
  version: 0,
};

function toDatetimeLocal(iso: string | null | undefined): string {
  if (!iso) return '';
  const d = new Date(iso);
  if (Number.isNaN(d.getTime())) return '';
  const pad = (n: number) => n.toString().padStart(2, '0');
  const yyyy = d.getFullYear();
  const mm = pad(d.getMonth() + 1);
  const dd = pad(d.getDate());
  const hh = pad(d.getHours());
  const mi = pad(d.getMinutes());
  return `${yyyy}-${mm}-${dd}T${hh}:${mi}`;
}

function fromDatetimeLocal(value: string): string | null {
  if (!value) return null;
  const d = new Date(value);
  if (Number.isNaN(d.getTime())) return null;
  return d.toISOString();
}

function formatDueDate(iso: string | null | undefined): string {
  if (!iso) return '';
  const d = new Date(iso);
  if (Number.isNaN(d.getTime())) return '';
  return d.toLocaleString();
}

export function TodosPage() {
  const navigate = useNavigate();
  const { clearSession } = useAuth();

  const [todos, setTodos] = useState<Todo[]>([]);
  const [categories, setCategories] = useState<Category[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);

  const [filters, setFilters] = useState<{
    completed: '' | 'true' | 'false';
    categoryId: string;
    dueFrom: string;
    dueTo: string;
  }>({
    completed: '',
    categoryId: '',
    dueFrom: '',
    dueTo: '',
  });

  const [form, setForm] = useState<FormState>(EMPTY_FORM);
  const [formError, setFormError] = useState<string | null>(null);
  const [submitting, setSubmitting] = useState(false);

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

  const activeFilters = useMemo<TodoFilters>(() => {
    const f: TodoFilters = {};
    if (filters.completed === 'true') f.completed = true;
    else if (filters.completed === 'false') f.completed = false;
    if (filters.categoryId) f.categoryId = filters.categoryId;
    if (filters.dueFrom) {
      const iso = fromDatetimeLocal(filters.dueFrom);
      if (iso) f.dueFrom = iso;
    }
    if (filters.dueTo) {
      const iso = fromDatetimeLocal(filters.dueTo);
      if (iso) f.dueTo = iso;
    }
    return f;
  }, [filters]);

  const loadTodos = useCallback(
    async (f: TodoFilters) => {
      setLoading(true);
      setError(null);
      try {
        const list = await todosApi.list(f);
        setTodos(Array.isArray(list) ? list : []);
      } catch (err) {
        applyError(err, setError);
      } finally {
        setLoading(false);
      }
    },
    [applyError],
  );

  const loadCategories = useCallback(async () => {
    try {
      const list = await categoriesApi.list();
      setCategories(Array.isArray(list) ? list : []);
    } catch (err) {
      applyError(err, setError);
    }
  }, [applyError]);

  useEffect(() => {
    loadCategories();
  }, [loadCategories]);

  useEffect(() => {
    loadTodos(activeFilters);
  }, [activeFilters, loadTodos]);

  const categoryLookup = useMemo(() => {
    const map = new Map<string, string>();
    for (const c of categories) map.set(c.id, c.name);
    return map;
  }, [categories]);

  const beginEdit = (todo: Todo) => {
    setForm({
      id: todo.id,
      title: todo.title || '',
      description: todo.description || '',
      dueDate: toDatetimeLocal(todo.dueDate),
      categoryId: todo.categoryId || '',
      version: todo.version || 0,
    });
    setFormError(null);
  };

  const cancelEdit = () => {
    setForm(EMPTY_FORM);
    setFormError(null);
  };

  const submitForm = async (e: FormEvent) => {
    e.preventDefault();
    setFormError(null);

    if (!form.title.trim()) {
      setFormError('Title is required.');
      return;
    }

    const payload: TodoInput = {
      title: form.title.trim(),
      description: form.description,
      dueDate: fromDatetimeLocal(form.dueDate),
      categoryId: form.categoryId || null,
    };

    setSubmitting(true);
    try {
      if (form.id) {
        payload.version = form.version;
        await todosApi.update(form.id, payload);
      } else {
        await todosApi.create(payload);
      }
      setForm(EMPTY_FORM);
      await loadTodos(activeFilters);
    } catch (err) {
      if (err instanceof ApiError && err.status === 409) {
        setFormError('This todo was modified elsewhere. Reload and try again.');
      } else {
        applyError(err, setFormError);
      }
    } finally {
      setSubmitting(false);
    }
  };

  const toggleComplete = async (todo: Todo) => {
    try {
      await todosApi.update(todo.id, {
        title: todo.title,
        description: todo.description,
        dueDate: todo.dueDate,
        categoryId: todo.categoryId,
        completed: !todo.completed,
        version: todo.version,
      });
      await loadTodos(activeFilters);
    } catch (err) {
      if (err instanceof ApiError && err.status === 409) {
        setError('Todo was modified elsewhere. Refreshing…');
        await loadTodos(activeFilters);
      } else {
        applyError(err, setError);
      }
    }
  };

  const deleteTodo = async (todo: Todo) => {
    if (!confirm(`Delete "${todo.title}"?`)) return;
    try {
      await todosApi.remove(todo.id);
      if (form.id === todo.id) setForm(EMPTY_FORM);
      await loadTodos(activeFilters);
    } catch (err) {
      applyError(err, setError);
    }
  };

  const clearFilters = () => {
    setFilters({ completed: '', categoryId: '', dueFrom: '', dueTo: '' });
  };

  return (
    <div className="container">
      <div className="card">
        <h2>{form.id ? 'Edit todo' : 'Create todo'}</h2>
        {formError && <div className="error">{formError}</div>}
        <form onSubmit={submitForm}>
          <div className="field">
            <label htmlFor="title">Title</label>
            <input
              id="title"
              type="text"
              required
              maxLength={200}
              value={form.title}
              onChange={(e) => setForm({ ...form, title: e.target.value })}
            />
          </div>
          <div className="field">
            <label htmlFor="description">Description</label>
            <textarea
              id="description"
              maxLength={2000}
              value={form.description}
              onChange={(e) => setForm({ ...form, description: e.target.value })}
            />
          </div>
          <div className="form-row">
            <div className="field">
              <label htmlFor="dueDate">Due date</label>
              <input
                id="dueDate"
                type="datetime-local"
                value={form.dueDate}
                onChange={(e) => setForm({ ...form, dueDate: e.target.value })}
              />
            </div>
            <div className="field">
              <label htmlFor="category">Category</label>
              <select
                id="category"
                value={form.categoryId}
                onChange={(e) => setForm({ ...form, categoryId: e.target.value })}
              >
                <option value="">— None —</option>
                {categories.map((c) => (
                  <option key={c.id} value={c.id}>
                    {c.name}
                  </option>
                ))}
              </select>
            </div>
          </div>
          <div className="row">
            <button type="submit" disabled={submitting}>
              {submitting ? 'Saving…' : form.id ? 'Update todo' : 'Add todo'}
            </button>
            {form.id && (
              <button type="button" className="secondary" onClick={cancelEdit}>
                Cancel
              </button>
            )}
          </div>
        </form>
      </div>

      <div className="card">
        <h2>Filters</h2>
        <div className="filter-bar">
          <div className="field">
            <label htmlFor="filter-completed">Status</label>
            <select
              id="filter-completed"
              value={filters.completed}
              onChange={(e) =>
                setFilters({
                  ...filters,
                  completed: e.target.value as '' | 'true' | 'false',
                })
              }
            >
              <option value="">All</option>
              <option value="false">Active</option>
              <option value="true">Completed</option>
            </select>
          </div>
          <div className="field">
            <label htmlFor="filter-category">Category</label>
            <select
              id="filter-category"
              value={filters.categoryId}
              onChange={(e) => setFilters({ ...filters, categoryId: e.target.value })}
            >
              <option value="">All</option>
              {categories.map((c) => (
                <option key={c.id} value={c.id}>
                  {c.name}
                </option>
              ))}
            </select>
          </div>
          <div className="field">
            <label htmlFor="filter-from">Due from</label>
            <input
              id="filter-from"
              type="datetime-local"
              value={filters.dueFrom}
              onChange={(e) => setFilters({ ...filters, dueFrom: e.target.value })}
            />
          </div>
          <div className="field">
            <label htmlFor="filter-to">Due to</label>
            <input
              id="filter-to"
              type="datetime-local"
              value={filters.dueTo}
              onChange={(e) => setFilters({ ...filters, dueTo: e.target.value })}
            />
          </div>
          <button type="button" className="secondary" onClick={clearFilters}>
            Clear
          </button>
        </div>
      </div>

      {error && <div className="error">{error}</div>}

      {loading ? (
        <div className="empty-state">Loading todos…</div>
      ) : todos.length === 0 ? (
        <div className="card empty-state">
          No todos match your filters. Create one above to get started.
        </div>
      ) : (
        <ul className="todo-list">
          {todos.map((todo) => (
            <li
              key={todo.id}
              className={`todo-item ${todo.completed ? 'completed' : ''}`}
            >
              <input
                type="checkbox"
                className="todo-checkbox"
                checked={todo.completed}
                onChange={() => toggleComplete(todo)}
                aria-label={todo.completed ? 'Mark incomplete' : 'Mark complete'}
              />
              <div className="todo-body">
                <p className="todo-title">{todo.title}</p>
                <div className="todo-meta">
                  {todo.categoryId && categoryLookup.has(todo.categoryId) && (
                    <span className="badge">{categoryLookup.get(todo.categoryId)}</span>
                  )}
                  {todo.dueDate && <span>Due: {formatDueDate(todo.dueDate)}</span>}
                  {todo.completed && todo.completedAt && (
                    <span>Completed: {formatDueDate(todo.completedAt)}</span>
                  )}
                </div>
                {todo.description && <p className="todo-description">{todo.description}</p>}
              </div>
              <div className="todo-actions">
                <button className="secondary" onClick={() => beginEdit(todo)}>
                  Edit
                </button>
                <button className="danger" onClick={() => deleteTodo(todo)}>
                  Delete
                </button>
              </div>
            </li>
          ))}
        </ul>
      )}
    </div>
  );
}
