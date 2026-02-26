import { describe, it, expect, vi, beforeEach } from 'vitest'
import { render, screen, fireEvent, waitFor } from '@testing-library/react'
import { MemoryRouter, Route, Routes } from 'react-router-dom'
import { ModulePage } from './ModulePage'
import type { Lesson } from '../../shared/types'

vi.mock('../../shared/api', () => ({
  modules: {
    get: vi.fn(),
    delete: vi.fn(),
    createLesson: vi.fn(),
  },
  lessons: {
    delete: vi.fn(),
  },
}))

vi.mock('../auth/AuthContext', () => ({
  useAuth: vi.fn(),
}))

// ConfirmModal требует портал — упрощаем
vi.mock('../../components', () => ({
  ConfirmModal: ({ open, onConfirm, onCancel, children }: {
    open: boolean
    onConfirm: () => void
    onCancel: () => void
    children: React.ReactNode
  }) =>
    open ? (
      <div role="dialog">
        {children}
        <button onClick={onConfirm}>Confirm</button>
        <button onClick={onCancel}>Cancel</button>
      </div>
    ) : null,
}))

vi.mock('./PageWithToc', () => ({
  PageWithToc: ({ children }: { children: React.ReactNode }) => <div>{children}</div>,
}))

vi.mock('./MarkdownStepEditor', () => ({
  MarkdownStepEditor: ({ onChange }: { onChange: (v: string) => void }) => (
    <textarea data-testid="md-editor" onChange={(e) => onChange(e.target.value)} />
  ),
}))

const api = await import('../../shared/api')
const useAuth = (await import('../auth/AuthContext').then((m) => m.useAuth)) as ReturnType<typeof vi.fn>

const TEACHER = { id: 'teacher1', login: 'teacher', name: 'Teacher', role: 'teacher' }
const STUDENT = { id: 'student1', login: 'student', name: 'Student', role: 'student' }

const MODULE_DATA = {
  id: 'mod1',
  title: 'Введение в Arduino',
  description: 'Базовый курс',
  sort_order: 0,
  created_at: '2025-01-01T00:00:00Z',
  lessons: [
    { id: 'l1', title: 'Урок 1: Мигаем светодиодом', sort_order: 0, lesson_type: 'theory', module_id: 'mod1', description: '', created_at: '2025-01-01T00:00:00Z' } as Lesson,
    { id: 'l2', title: 'Урок 2: Кнопка', sort_order: 1, lesson_type: 'practice', module_id: 'mod1', description: '', created_at: '2025-01-01T00:00:00Z' } as Lesson,
  ],
}

function renderModulePage() {
  return render(
    <MemoryRouter initialEntries={['/modules/mod1']}>
      <Routes>
        <Route path="/modules/:id" element={<ModulePage />} />
      </Routes>
    </MemoryRouter>
  )
}

describe('ModulePage', () => {
  beforeEach(() => {
    vi.mocked(api.modules.get).mockResolvedValue(MODULE_DATA)
  })

  it('renders module title and lessons', async () => {
    useAuth.mockReturnValue({ user: STUDENT })
    renderModulePage()

    expect(await screen.findByRole('heading', { name: /введение в arduino/i })).toBeInTheDocument()
    expect(await screen.findByText('Урок 1: Мигаем светодиодом')).toBeInTheDocument()
    expect(await screen.findByText('Урок 2: Кнопка')).toBeInTheDocument()
  })

  it('shows add lesson button for teacher', async () => {
    useAuth.mockReturnValue({ user: TEACHER })
    renderModulePage()

    expect(await screen.findByRole('button', { name: /добавить урок/i })).toBeInTheDocument()
  })

  it('hides teacher actions for student', async () => {
    useAuth.mockReturnValue({ user: STUDENT })
    renderModulePage()

    await screen.findByText('Урок 1: Мигаем светодиодом')
    expect(screen.queryByRole('button', { name: /добавить урок/i })).not.toBeInTheDocument()
    expect(screen.queryByRole('button', { name: /удалить курс/i })).not.toBeInTheDocument()
  })

  it('shows create form when teacher clicks add lesson', async () => {
    useAuth.mockReturnValue({ user: TEACHER })
    renderModulePage()

    const addBtn = await screen.findByRole('button', { name: /добавить урок/i })
    fireEvent.click(addBtn)

    expect(await screen.findByRole('heading', { name: /новый урок/i })).toBeInTheDocument()
    expect(screen.getByLabelText(/название/i)).toBeInTheDocument()
  })

  it('creates a lesson and shows it in the list', async () => {
    useAuth.mockReturnValue({ user: TEACHER })
    const newLesson = { id: 'l3', title: 'Урок 3: Сервопривод', sort_order: 2, lesson_type: 'practice', module_id: 'mod1', description: '', created_at: '2025-01-01T00:00:00Z' } as Lesson
    vi.mocked(api.modules.createLesson).mockResolvedValue(newLesson)
    renderModulePage()

    fireEvent.click(await screen.findByRole('button', { name: /добавить урок/i }))
    const titleInput = await screen.findByLabelText(/название/i)
    fireEvent.change(titleInput, { target: { value: 'Урок 3: Сервопривод' } })
    fireEvent.click(screen.getByRole('button', { name: /создать/i }))

    expect(await screen.findByText('Урок 3: Сервопривод')).toBeInTheDocument()
    expect(api.modules.createLesson).toHaveBeenCalledWith('mod1', expect.objectContaining({ title: 'Урок 3: Сервопривод' }))
  })

  it('shows error when module load fails', async () => {
    useAuth.mockReturnValue({ user: STUDENT })
    vi.mocked(api.modules.get).mockRejectedValue(new Error('Сервер недоступен'))
    renderModulePage()

    expect(await screen.findByText(/сервер недоступен/i)).toBeInTheDocument()
  })

  it('asks confirmation before deleting lesson', async () => {
    useAuth.mockReturnValue({ user: TEACHER })
    renderModulePage()

    const deleteBtns = await screen.findAllByRole('button', { name: /удалить/i })
    // первый не относящийся к уроку — берём правильный
    const lessonDeleteBtn = deleteBtns.find((b) => !b.closest('.module-teacher-buttons'))
    expect(lessonDeleteBtn).toBeTruthy()
    fireEvent.click(lessonDeleteBtn!)

    expect(await screen.findByRole('dialog')).toBeInTheDocument()
  })

  it('deletes lesson after confirmation', async () => {
    useAuth.mockReturnValue({ user: TEACHER })
    vi.mocked(api.lessons.delete).mockResolvedValue(undefined)
    renderModulePage()

    const deleteBtns = await screen.findAllByRole('button', { name: /удалить/i })
    const lessonDeleteBtn = deleteBtns.find((b) => !b.closest('.module-teacher-buttons'))
    fireEvent.click(lessonDeleteBtn!)

    const confirmBtn = await screen.findByRole('button', { name: /confirm/i })
    fireEvent.click(confirmBtn)

    await waitFor(() => {
      expect(api.lessons.delete).toHaveBeenCalledWith('l1')
    })
  })
})
