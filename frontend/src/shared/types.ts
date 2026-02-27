export type ThemeId =
  | 'default'
  | 'light'
  | 'cyberpunk'
  | 'contrast-light'
  | 'contrast-dark'
  | 'cream'
  | 'snow'
  | 'midnight'
  | 'forest'

export type UserRole = 'student' | 'teacher' | 'course_owner' | 'administrator'

export interface User {
  id: string
  login: string
  name: string
  role: 'student' | 'teacher' | 'administrator'
  roles?: UserRole[]
  theme?: ThemeId
  created_at: string
  email?: string
  must_change_password?: boolean
  is_blocked?: boolean
}

export interface AdminStats {
  users: number
  modules: number
  lessons: number
}

export interface ActivityItem {
  id: string
  login: string
  name: string
  role: 'student' | 'teacher' | 'administrator'
  created_at: string
}

export interface AdminCreateUserRequest {
  login: string
  password: string
  name: string
  role: 'student' | 'teacher'
  email?: string
}

export interface Module {
  id: string
  title: string
  description: string
  sort_order: number
  lessons?: Lesson[]
  created_at: string
  owner_id?: string
  is_owner?: boolean
}

/** Реакция пользователя: лайк или дизлайк */
export type ReactionType = 'like' | 'dislike'

export interface Lesson {
  id: string
  module_id: string
  title: string
  description: string
  lesson_type: 'theory' | 'practice' | 'project'
  sort_order: number
  steps?: LessonStep[]
  materials?: LessonMaterial[]
  tags?: string[]
  checklist?: ChecklistItem[]
  created_at: string
  likes_count?: number
  dislikes_count?: number
  user_reaction?: ReactionType
}

export interface LessonStep {
  id: string
  lesson_id: string
  title: string
  content: string
  sort_order: number
}

export interface LessonMaterial {
  id: string
  lesson_id: string
  kind: 'link' | 'file' | 'simulator'
  url_or_path: string
  title?: string
}

export interface ChecklistItem {
  id: string
  lesson_id: string
  title: string
  sort_order: number
}

/** Статус прохождения урока (контракт progress API). Подписи для UI — через i18n (progress.status_*). */
export type LessonStatus = 'not_started' | 'in_progress' | 'completed'

export interface LessonProgressItem {
  lesson_id: string
  lesson_title?: string
  module_id?: string
  status: LessonStatus
  updated_at: string
}

export interface UserProgress {
  lessons: LessonProgressItem[]
  checklist: { checklist_item_id: string; completed_at?: string }[]
}

export interface LessonComment {
  id: string
  lesson_id: string
  user_id: string
  user_name: string
  text: string
  created_at: string
  likes_count?: number
  dislikes_count?: number
  user_reaction?: ReactionType
}
