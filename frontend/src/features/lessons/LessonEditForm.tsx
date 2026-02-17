import { useTranslation } from 'react-i18next'
import { MarkdownStepEditor } from './MarkdownStepEditor'
import type { EditForm } from './hooks'
import styles from './LessonPage.module.css'

type LessonEditFormProps = {
  form: EditForm
  saveError: string
  onTitleChange: (value: string) => void
  onDescriptionChange: (value: string) => void
  onStepTitleChange: (stepId: string, value: string) => void
  onStepContentChange: (stepId: string, value: string) => void
  onRemoveStep: (stepId: string) => void
  onAddStep: () => void
  onSave: () => void
  onCancel: () => void
}

export function LessonEditForm({
  form,
  saveError,
  onTitleChange,
  onDescriptionChange,
  onStepTitleChange,
  onStepContentChange,
  onRemoveStep,
  onAddStep,
  onSave,
  onCancel,
}: LessonEditFormProps) {
  const { t } = useTranslation()

  return (
    <section
      className={styles.editForm}
      data-lesson-edit-form
      aria-labelledby="lesson-edit-form-title"
    >
      <h2 id="lesson-edit-form-title">{t('lesson.editFormTitle')}</h2>
      {saveError && <p className="error" role="alert">{saveError}</p>}
      <div className="form-group">
        <label htmlFor="edit-lesson-title">{t('catalog.courseTitle')}</label>
        <input
          id="edit-lesson-title"
          value={form.title}
          onChange={(e) => onTitleChange(e.target.value)}
          aria-required
        />
      </div>
      <div className="form-group">
        <label htmlFor="edit-lesson-desc">{t('catalog.description')}</label>
        <textarea
          id="edit-lesson-desc"
          rows={2}
          value={form.description}
          onChange={(e) => onDescriptionChange(e.target.value)}
        />
      </div>
      <h3>{t('lesson.steps')}</h3>
      {form.steps.map((step, idx) => (
        <div
          key={step.id}
          className={styles.editStep}
          data-lesson-edit-step
        >
          <div className="form-group">
            <label htmlFor={`edit-step-title-${step.id}`}>
              {t('lesson.stepHeading', { n: idx + 1 })}
            </label>
            <input
              id={`edit-step-title-${step.id}`}
              value={step.title}
              onChange={(e) => onStepTitleChange(step.id, e.target.value)}
              aria-required
            />
          </div>
          <div className="form-group">
            <MarkdownStepEditor
              value={step.content}
              onChange={(v) => onStepContentChange(step.id, v)}
              label={t('lesson.stepContentLabel')}
            />
          </div>
          <button
            type="button"
            className="button-danger-outline"
            onClick={() => onRemoveStep(step.id)}
            disabled={form.steps.length <= 1}
            aria-label={t('lesson.deleteStep')}
          >
            {t('lesson.deleteStep')}
          </button>
        </div>
      ))}
      <div className={styles.formActions}>
        <button type="button" className="button-secondary" onClick={onAddStep}>
          {t('module.addStep')}
        </button>
        <button type="button" className="button-primary" onClick={onSave}>
          {t('common.save')}
        </button>
        <button type="button" className="button-secondary" onClick={onCancel}>
          {t('common.cancel')}
        </button>
      </div>
    </section>
  )
}
