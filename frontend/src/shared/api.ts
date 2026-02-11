const API_BASE = '/api/v1'

function getToken(): string | null {
  return localStorage.getItem('token')
}

export async function api<T>(
  path: string,
  options: RequestInit = {}
): Promise<T> {
  const token = getToken()
  const headers: HeadersInit = {
    'Content-Type': 'application/json',
    ...(options.headers as Record<string, string>),
  }
  if (token) {
    (headers as Record<string, string>)['Authorization'] = `Bearer ${token}`
  }
  const res = await fetch(API_BASE + path, { ...options, headers })
  const data = await res.json().catch(() => ({}))
  if (!res.ok) {
    throw new Error((data.error as string) || data.message || res.statusText)
  }
  return data as T
}

export const auth = {
  register: (body: { login: string; password: string; name: string; role?: string }) =>
    api<{ user: import('./types').User; token: string }>('/auth/register', {
      method: 'POST',
      body: JSON.stringify(body),
    }),
  login: (body: { login: string; password: string }) =>
    api<{ user: import('./types').User; token: string }>('/auth/login', {
      method: 'POST',
      body: JSON.stringify(body),
    }),
  me: () => api<import('./types').User>('/auth/me'),
  updateTheme: (theme: import('./types').ThemeId) =>
    api<import('./types').User>('/auth/me', {
      method: 'PATCH',
      body: JSON.stringify({ theme }),
    }),
}

export const modules = {
  list: (params?: { platform?: string; tag?: string }) => {
    const q = new URLSearchParams(params as Record<string, string>).toString()
    return api<import('./types').Module[]>(`/modules${q ? `?${q}` : ''}`)
  },
  get: (id: string) => api<import('./types').Module>(`/modules/${id}`),
  create: (body: { title: string; description?: string; sort_order?: number }) =>
    api<import('./types').Module>('/modules', {
      method: 'POST',
      body: JSON.stringify(body),
    }),
  delete: (id: string) =>
    api<void>(`/modules/${id}`, { method: 'DELETE' }),
  createLesson: (
    moduleId: string,
    body: {
      title: string
      description?: string
      lesson_type?: string
      sort_order?: number
      steps?: { title: string; content?: string }[]
    }
  ) =>
    api<import('./types').Lesson>(`/modules/${moduleId}/lessons`, {
      method: 'POST',
      body: JSON.stringify(body),
    }),
}

export const lessons = {
  get: (id: string) => api<import('./types').Lesson>(`/lessons/${id}`),
  delete: (id: string) => api<void>(`/lessons/${id}`, { method: 'DELETE' }),
  getComments: (id: string) =>
    api<import('./types').LessonComment[]>(`/lessons/${id}/comments`),
  addComment: (id: string, text: string) =>
    api<import('./types').LessonComment>(`/lessons/${id}/comments`, {
      method: 'POST',
      body: JSON.stringify({ text }),
    }),
  deleteComment: (lessonId: string, commentId: string) =>
    api<void>(`/lessons/${lessonId}/comments/${commentId}`, { method: 'DELETE' }),
  update: (
    id: string,
    body: {
      title?: string
      description?: string
      steps?: { title: string; content: string; sort_order: number }[]
    }
  ) =>
    api<import('./types').Lesson>(`/lessons/${id}`, {
      method: 'PUT',
      body: JSON.stringify(body),
    }),
}

export const progress = {
  get: () => api<import('./types').UserProgress>('/progress'),
  setLesson: (lessonId: string, status: string) =>
    api<{ ok: boolean }>(`/lessons/${lessonId}/progress`, {
      method: 'PUT',
      body: JSON.stringify({ status }),
    }),
  setChecklist: (lessonId: string, itemId: string, completed: boolean) =>
    api<{ ok: boolean }>(`/lessons/${lessonId}/checklist/${itemId}`, {
      method: 'PUT',
      body: JSON.stringify({ completed }),
    }),
}

export const users = {
  list: () => api<import('./types').User[]>('/users'),
  progress: (userId: string) =>
    api<import('./types').UserProgress>(`/users/${userId}/progress`),
  delete: (userId: string) =>
    api<void>(`/users/${userId}`, { method: 'DELETE' }),
}

export interface ChatMessage {
  role: 'user' | 'model'
  text: string
}

export const chat = {
  send: (lessonContext: string, messages: ChatMessage[]) =>
    api<{ text: string }>('/chat', {
      method: 'POST',
      body: JSON.stringify({ lesson_context: lessonContext, messages }),
    }),
}
