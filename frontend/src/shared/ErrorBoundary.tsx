import { Component, Fragment, type ErrorInfo, type ReactNode } from 'react'
import { ErrorFallback } from './ErrorFallback'

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
    // key на Fragment гарантирует полный ремаунт детей при каждом retry
    return <Fragment key={this.state.resetKey}>{this.props.children}</Fragment>
  }
}
