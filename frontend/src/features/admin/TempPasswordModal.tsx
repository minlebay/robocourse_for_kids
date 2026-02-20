import { useState } from 'react'
import { useTranslation } from 'react-i18next'

interface TempPasswordModalProps {
  password: string
  onClose: () => void
}

export function TempPasswordModal({ password, onClose }: TempPasswordModalProps) {
  const { t } = useTranslation()
  const [copied, setCopied] = useState(false)

  async function handleCopy() {
    try {
      await navigator.clipboard.writeText(password)
      setCopied(true)
      setTimeout(() => setCopied(false), 2000)
    } catch {
      // clipboard not available — do nothing
    }
  }

  return (
    <div className="temp-modal-overlay" role="dialog" aria-modal="true" aria-labelledby="temp-modal-title">
      <div className="temp-modal">
        <h2 id="temp-modal-title" className="temp-modal-title">{t('admin.tempPasswordTitle')}</h2>
        <p className="temp-modal-warning">{t('admin.tempPasswordWarning')}</p>
        <div className="temp-modal-password">{password}</div>
        <div className="temp-modal-actions">
          <button type="button" className="button-primary" onClick={handleCopy}>
            {copied ? t('admin.copied') : t('admin.copyPassword')}
          </button>
          <button type="button" className="button-secondary" onClick={onClose}>
            {t('common.cancel')}
          </button>
        </div>
      </div>
    </div>
  )
}
