export type ThemeId = 'default' | 'light' | 'cyberpunk'

export interface User {
  id: string
  login: string
  name: string
  role: 'student' | 'teacher'
  theme?: ThemeId
  created_at: string
}

export interface Module {
  id: string
  title: string
  description: string
  sort_order: number
  lessons?: Lesson[]
  created_at: string
}

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
  kind: 'link' | 'file' | 'simulator' | 'mermaid'
  url_or_path: string
  title?: string
}

export interface ChecklistItem {
  id: string
  lesson_id: string
  title: string
  sort_order: number
}

export interface LessonProgressItem {
  lesson_id: string
  lesson_title?: string
  module_id?: string
  status: 'not_started' | 'in_progress' | 'completed'
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
}
