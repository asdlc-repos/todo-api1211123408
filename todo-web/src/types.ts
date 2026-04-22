export interface Todo {
  id: string;
  title: string;
  description: string;
  dueDate: string | null;
  categoryId: string | null;
  completed: boolean;
  completedAt: string | null;
  version: number;
  createdAt: string;
}

export interface TodoInput {
  title: string;
  description?: string;
  dueDate?: string | null;
  categoryId?: string | null;
  completed?: boolean;
  version?: number;
}

export interface Category {
  id: string;
  name: string;
}

export interface TodoFilters {
  completed?: boolean;
  categoryId?: string;
  dueFrom?: string;
  dueTo?: string;
}

export interface ApiError {
  error: string;
  status: number;
}
