import { Component, type ErrorInfo, type ReactNode } from 'react'

type Props = {
  children: ReactNode
  fallback?: ReactNode
  onError?: (error: Error, errorInfo: ErrorInfo) => void
}

type State = {
  hasError: boolean
  error: Error | null
}

export class ErrorBoundary extends Component<Props, State> {
  state: State = { hasError: false, error: null }

  static getDerivedStateFromError(error: Error): State {
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
        <div className="error-boundary" role="alert">
          <h2>Что-то пошло не так</h2>
          <p className="error-boundary-message">{this.state.error.message}</p>
          <button
            type="button"
            className="button-primary"
            onClick={() => this.setState({ hasError: false, error: null })}
          >
            Попробовать снова
          </button>
        </div>
      )
    }
    return this.props.children
  }
}
