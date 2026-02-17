import { Link } from 'react-router-dom'
import { useTranslation } from 'react-i18next'
import type { Lesson, Module } from '../../shared/types'

type CourseNavProps = {
  module: Module | null
  currentLesson: Lesson
}

export function CourseNav({ module, currentLesson }: CourseNavProps) {
  const { t } = useTranslation()
  const orderedLessons = (module?.lessons ?? []).slice().sort((a, b) => a.sort_order - b.sort_order)
  const currentIndex = orderedLessons.findIndex((l) => l.id === currentLesson.id)
  const prevLesson = currentIndex > 0 ? orderedLessons[currentIndex - 1] : null
  const nextLesson = currentIndex >= 0 && currentIndex < orderedLessons.length - 1 ? orderedLessons[currentIndex + 1] : null

  return (
    <nav className="course-nav" aria-label={t('courseNav.ariaLabel')}>
      <Link to={module ? `/modules/${module.id}` : '/'} className="course-nav-home">
        {t('courseNav.home')}
      </Link>
      {prevLesson ? (
        <Link to={`/lessons/${prevLesson.id}`} className="course-nav-prev">
          {t('courseNav.back')}
        </Link>
      ) : (
        <span className="course-nav-prev disabled" aria-hidden>{t('courseNav.back')}</span>
      )}
      {nextLesson ? (
        <Link to={`/lessons/${nextLesson.id}`} className="course-nav-next">
          {t('courseNav.forward')}
        </Link>
      ) : (
        <span className="course-nav-next disabled" aria-hidden>{t('courseNav.forward')}</span>
      )}
    </nav>
  )
}
