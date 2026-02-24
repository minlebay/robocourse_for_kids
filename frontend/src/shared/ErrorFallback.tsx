import { useTranslation } from 'react-i18next'

export function ErrorFallback({
  message,
  onRetry,
}: {
  message: string
  onRetry: () => void
}) {
  const { t } = useTranslation()
  return (
    <div className="error-boundary" role="alert">
      <h2>{t('errors.somethingWrong')}</h2>
      <p className="error-boundary-message">{message}</p>
      <button type="button" className="button-primary" onClick={onRetry}>
        {t('common.retry')}
      </button>
    </div>
  )
}
