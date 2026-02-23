import { Component, type ErrorInfo, type ReactNode } from 'react'
import { useTranslation } from 'react-i18next'

type Props = {
  children: ReactNode
  fallback?: ReactNode
  onError?: (error: Error, errorInfo: ErrorInfo) => void
}

type State = {
  hasError: boolean
  error: Error | null
  /** Инкрементируется при retry, что форсирует полный ремаунт дочерних компонентов. */
  resetKey: number
}

function ErrorFallback({
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

export class ErrorBoundary extends Component<Props, State> {
  state: State = { hasError: false, error: null, resetKey: 0 }

  static getDerivedStateFromError(error: Error): Partial<State> {
    return { hasError: true, error }
  }

  componentDidCatch(error: Error, errorInfo: ErrorInfo) {
    if (this.props.onError) {
      this.props.onError(error, errorInfo)
    } else {
      console.error('ErrorBoundary caught an error:', error, errorInfo)
    }
  }

  render() {
    if (this.state.hasError && this.state.error) {
      if (this.props.fallback) return this.props.fallback
      return (
        <ErrorFallback
          message={this.state.error.message}
          onRetry={() =>
            this.setState((prev) => ({
              hasError: false,
              error: null,
              resetKey: prev.resetKey + 1,
            }))
          }
        />
      )
    }
    // key на обёртке гарантирует полный ремаунт детей при каждом retry
    return (
      <ChildrenWrapper key={this.state.resetKey}>
        {this.props.children}
      </ChildrenWrapper>
    )
  }
}

function ChildrenWrapper({ children }: { children: ReactNode }) {
  return <>{children}</>
}
