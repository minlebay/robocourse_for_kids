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
}

export const modules = {
  list: (params?: { platform?: string; tag?: string }) => {
    const q = new URLSearchParams(params as Record<string, string>).toString()
    return api<import('./types').Module[]>(`/modules${q ? `?${q}` : ''}`)
  },
  get: (id: string) => api<import('./types').Module>(`/modules/${id}`),
}

export const lessons = {
  get: (id: string) => api<import('./types').Lesson>(`/lessons/${id}`),
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
