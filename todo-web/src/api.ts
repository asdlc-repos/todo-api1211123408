import type { Category, Todo, TodoFilters, TodoInput } from './types';

const API_BASE =
  (import.meta.env.VITE_API_BASE as string | undefined)?.replace(/\/$/, '') || '/api/v1';

export class ApiError extends Error {
  status: number;
  constructor(message: string, status: number) {
    super(message);
    this.status = status;
  }
}

async function request<T>(path: string, init: RequestInit = {}): Promise<T> {
  const headers = new Headers(init.headers);
  if (init.body && !headers.has('Content-Type')) {
    headers.set('Content-Type', 'application/json');
  }
  headers.set('Accept', 'application/json');

  const res = await fetch(`${API_BASE}${path}`, {
    ...init,
    headers,
    credentials: 'include',
  });

  if (res.status === 204) {
    return undefined as unknown as T;
  }

  const text = await res.text();
  let body: unknown = undefined;
  if (text) {
    try {
      body = JSON.parse(text);
    } catch {
      body = text;
    }
  }

  if (!res.ok) {
    const message =
      (body && typeof body === 'object' && 'error' in body && typeof (body as { error: unknown }).error === 'string'
        ? (body as { error: string }).error
        : res.statusText) || `Request failed (${res.status})`;
    throw new ApiError(message, res.status);
  }

  return body as T;
}

function buildQuery(params: Record<string, string | number | boolean | undefined | null>): string {
  const usp = new URLSearchParams();
  for (const [key, value] of Object.entries(params)) {
    if (value === undefined || value === null || value === '') continue;
    usp.append(key, String(value));
  }
  const qs = usp.toString();
  return qs ? `?${qs}` : '';
}

export const auth = {
  register: (email: string, password: string) =>
    request<void>('/auth/register', {
      method: 'POST',
      body: JSON.stringify({ email, password }),
    }),
  login: (email: string, password: string) =>
    request<void>('/auth/login', {
      method: 'POST',
      body: JSON.stringify({ email, password }),
    }),
  logout: () => request<void>('/auth/logout', { method: 'POST' }),
  requestReset: (email: string) =>
    request<void>('/auth/password-reset/request', {
      method: 'POST',
      body: JSON.stringify({ email }),
    }),
  confirmReset: (token: string, password: string) =>
    request<void>('/auth/password-reset/confirm', {
      method: 'POST',
      body: JSON.stringify({ token, password }),
    }),
};

export const todos = {
  list: (filters: TodoFilters = {}) =>
    request<Todo[]>(`/todos${buildQuery({ ...filters })}`),
  get: (id: string) => request<Todo>(`/todos/${encodeURIComponent(id)}`),
  create: (input: TodoInput) =>
    request<Todo>('/todos', { method: 'POST', body: JSON.stringify(input) }),
  update: (id: string, input: TodoInput) =>
    request<Todo>(`/todos/${encodeURIComponent(id)}`, {
      method: 'PUT',
      body: JSON.stringify(input),
    }),
  remove: (id: string) =>
    request<void>(`/todos/${encodeURIComponent(id)}`, { method: 'DELETE' }),
};

export const categories = {
  list: () => request<Category[]>('/categories'),
  create: (name: string) =>
    request<Category>('/categories', {
      method: 'POST',
      body: JSON.stringify({ name }),
    }),
  update: (id: string, name: string) =>
    request<Category>(`/categories/${encodeURIComponent(id)}`, {
      method: 'PUT',
      body: JSON.stringify({ name }),
    }),
  remove: (id: string) =>
    request<void>(`/categories/${encodeURIComponent(id)}`, { method: 'DELETE' }),
};
