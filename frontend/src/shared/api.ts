import type { LessonStatus, ReactionType, AdminCreateUserRequest, AdminStats, ActivityItem } from './types'

const API_BASE = '/api/v1'
const RETURN_URL_KEY = 'learn_kids_return_url'
const REQUEST_TIMEOUT_MS = 30_000

interface ErrorResponse {
  error?: string
  message?: string
}

function getToken(): string | null {
  return localStorage.getItem('token')
}

/**
 * При 401: сбрасываем токен, сохраняем текущий путь для returnUrl (если безопасный)
 * и редиректим на /login. throw гарантирует, что вызывающий код не продолжит выполнение.
 * После присвоения location.href страница выгружается, поэтому повторный 401 в другом
 * запросе приведёт к той же последовательности (идемпотентно).
 */
function handle401(): never {
  localStorage.removeItem('token')
  const path = window.location.pathname + window.location.search
  if (path !== '/login' && path !== '/register') {
    try {
      sessionStorage.setItem(RETURN_URL_KEY, path)
    } catch {
      /* ignore */
    }
  }
  window.location.href = '/login'
  throw new Error('Сессия истекла. Выполните вход заново.')
}

/** Возвращает сохранённый returnUrl и удаляет его из sessionStorage. Безопасные пути: /, /modules/*, /lessons/*, /progress, /dashboard. */
export function consumeReturnUrl(): string | null {
  try {
    const url = sessionStorage.getItem(RETURN_URL_KEY)
    if (!url) return null
    sessionStorage.removeItem(RETURN_URL_KEY)
    const path = url.split('?')[0]
    if (path === '/' || path === '/progress' || path === '/dashboard') return url
    if (path.startsWith('/modules/') || path.startsWith('/lessons/')) return url
    return null
  } catch {
    return null
  }
}

async function request(path: string, options: RequestInit = {}): Promise<Response> {
  const token = getToken()
  const headers: HeadersInit = {
    'Content-Type': 'application/json',
    ...(options.headers as Record<string, string>),
  }
  if (token) {
    (headers as Record<string, string>)['Authorization'] = `Bearer ${token}`
  }

  const controller = new AbortController()
  const timeoutId = setTimeout(() => controller.abort(), REQUEST_TIMEOUT_MS)
  try {
    return await fetch(API_BASE + path, { ...options, headers, signal: controller.signal })
  } catch (err) {
    if (err instanceof Error && err.name === 'AbortError') {
      throw new Error('Запрос превысил допустимое время ожидания. Проверьте соединение.')
    }
    throw err
  } finally {
    clearTimeout(timeoutId)
  }
}

const NEW_TOKEN_HEADER = 'X-New-Token'

async function handleResponse<T>(res: Response, expectBody: boolean): Promise<T> {
  if (res.status === 401) handle401()
  // Sliding session: сервер может вернуть новый токен в заголовке
  if (res.ok) {
    const newToken = res.headers.get(NEW_TOKEN_HEADER)
    if (newToken) localStorage.setItem('token', newToken)
  }
  if (res.status === 429) {
    throw new Error('Слишком много запросов. Подожди немного и попробуй снова.')
  }
  if (res.status === 204) return undefined as T
  const data = await res.json().catch(() => null) as ErrorResponse | null
  if (!res.ok) {
    const msg = data?.error ?? data?.message ?? 'Ошибка сервера. Попробуйте позже.'
    throw new Error(msg)
  }
  return expectBody ? (data as T) : (undefined as T)
}

export async function api<T>(
  path: string,
  options: RequestInit = {}
): Promise<T> {
  const res = await request(path, options)
  return handleResponse<T>(res, true)
}

/** Запросы без тела ответа (204). */
export async function apiVoid(path: string, options: RequestInit = {}): Promise<void> {
  const res = await request(path, options)
  await handleResponse<void>(res, false)
}

export const auth = {
  register: (body: { login: string; password: string; name: string; role?: string; invite_code?: string }) =>
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
  changePassword: async (currentPassword: string, newPassword: string): Promise<void> => {
    const res = await request('/auth/change-password', {
      method: 'POST',
      body: JSON.stringify({ current_password: currentPassword, new_password: newPassword }),
    })
    if (!res.ok) {
      if (res.status === 401) handle401()
      const data = await res.json().catch(() => null) as ErrorResponse | null
      const msg = data?.error ?? data?.message ?? 'Ошибка сервера. Попробуйте позже.'
      throw new Error(msg)
    }
  },
}

export const modules = {
  list: (params?: { platform?: string; tag?: string }) => {
    const filtered: Record<string, string> = {}
    if (params) {
      for (const [k, v] of Object.entries(params)) {
        if (v !== undefined && v !== null) filtered[k] = v
      }
    }
    const q = new URLSearchParams(filtered).toString()
    return api<import('./types').Module[]>(`/modules${q ? `?${q}` : ''}`)
  },
  get: (id: string) => api<import('./types').Module>(`/modules/${id}`),
  create: (body: { title: string; description?: string; sort_order?: number }) =>
    api<import('./types').Module>('/modules', {
      method: 'POST',
      body: JSON.stringify(body),
    }),
  delete: (id: string) =>
    apiVoid(`/modules/${id}`, { method: 'DELETE' }),
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
  delete: (id: string) => apiVoid(`/lessons/${id}`, { method: 'DELETE' }),
  getComments: (id: string) =>
    api<import('./types').LessonComment[]>(`/lessons/${id}/comments`),
  addComment: (id: string, text: string) =>
    api<import('./types').LessonComment>(`/lessons/${id}/comments`, {
      method: 'POST',
      body: JSON.stringify({ text }),
    }),
  deleteComment: (lessonId: string, commentId: string) =>
    apiVoid(`/lessons/${lessonId}/comments/${commentId}`, { method: 'DELETE' }),
  setReaction: (lessonId: string, reaction: ReactionType) =>
    apiVoid(`/lessons/${lessonId}/reaction`, {
      method: 'PUT',
      body: JSON.stringify({ reaction }),
    }),
  deleteReaction: (lessonId: string) =>
    apiVoid(`/lessons/${lessonId}/reaction`, { method: 'DELETE' }),
  setCommentReaction: (lessonId: string, commentId: string, reaction: ReactionType) =>
    apiVoid(`/lessons/${lessonId}/comments/${commentId}/reaction`, {
      method: 'PUT',
      body: JSON.stringify({ reaction }),
    }),
  deleteCommentReaction: (lessonId: string, commentId: string) =>
    apiVoid(`/lessons/${lessonId}/comments/${commentId}/reaction`, { method: 'DELETE' }),
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
  setLesson: (lessonId: string, status: LessonStatus) =>
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
    apiVoid(`/users/${userId}`, { method: 'DELETE' }),
}

export interface ChatMessage {
  role: 'user' | 'model'
  text: string
  /** Стабильный ключ для React (генерируется на клиенте, бэкенд может не возвращать). */
  id?: string
}

export const chat = {
  send: (lessonId: string, message: string) =>
    api<{ text: string }>('/chat', {
      method: 'POST',
      body: JSON.stringify({ lesson_id: lessonId, message }),
    }),
  getHistory: (lessonId: string) =>
    api<{ messages: ChatMessage[] }>(`/chat/${lessonId}/history`),
  clearHistory: (lessonId: string) =>
    apiVoid(`/chat/${lessonId}/history`, { method: 'DELETE' }),
}

export const admin = {
  listUsers: () => api<import('./types').User[]>('/admin/users'),
  createUser: (data: AdminCreateUserRequest) =>
    api<{ user: import('./types').User; temp_password: string }>('/admin/users', {
      method: 'POST',
      body: JSON.stringify(data),
    }),
  deleteUser: (id: string) =>
    apiVoid(`/admin/users/${id}`, { method: 'DELETE' }),
  blockUser: (id: string, block: boolean) =>
    apiVoid(`/admin/users/${id}/block`, {
      method: 'POST',
      body: JSON.stringify({ block }),
    }),
  resetPassword: (id: string) =>
    api<{ temp_password: string }>(`/admin/users/${id}/reset-password`, {
      method: 'POST',
    }),
  getStats: () => api<AdminStats>('/admin/stats'),
  getActivity: () => api<ActivityItem[]>('/admin/activity'),
}
