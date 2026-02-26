const API_BASE = import.meta.env.VITE_API_URL || 'http://localhost:8080/api';

async function request(path, options = {}) {
  const res = await fetch(`${API_BASE}${path}`, {
    credentials: 'include',
    headers: {
      'Content-Type': 'application/json',
      ...(options.headers || {})
    },
    ...options
  });

  const contentType = res.headers.get('Content-Type') || '';
  const isJSON = contentType.includes('application/json');
  const data = isJSON ? await res.json().catch(() => null) : null;

  if (!res.ok) {
    throw new Error(data?.error || 'İstek başarısız');
  }
  return data;
}

export const api = {
  register: (payload) =>
    request('/auth/register', {
      method: 'POST',
      body: JSON.stringify(payload)
    }),
  login: (payload) =>
    request('/auth/login', {
      method: 'POST',
      body: JSON.stringify(payload)
    }),
  logout: () =>
    request('/auth/logout', {
      method: 'POST'
    }),
  resetPassword: (payload) =>
    request('/auth/reset-password', {
      method: 'POST',
      body: JSON.stringify(payload)
    }),
  me: () => request('/me'),
  updateMe: (payload) =>
    request('/me', {
      method: 'PUT',
      body: JSON.stringify(payload)
    }),
  listTasks: () => request('/tasks'),
  createTask: (payload) =>
    request('/tasks', {
      method: 'POST',
      body: JSON.stringify(payload)
    }),
  updateTask: (id, payload) =>
    request(`/tasks/${id}`, {
      method: 'PUT',
      body: JSON.stringify(payload)
    }),
  updateTaskStatus: (id, status) =>
    request(`/tasks/${id}/status`, {
      method: 'PATCH',
      body: JSON.stringify({ status })
    }),
  deleteTask: (id) =>
    request(`/tasks/${id}`, {
      method: 'DELETE'
    })
};

