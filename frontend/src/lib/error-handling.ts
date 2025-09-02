import * as Sentry from '@sentry/nextjs'
import { toast } from 'react-hot-toast'
import { AxiosError } from 'axios'
import React from 'react'

// Error types
export enum ErrorType {
  NETWORK = 'NETWORK',
  AUTHENTICATION = 'AUTHENTICATION',
  VALIDATION = 'VALIDATION',
  PERMISSION = 'PERMISSION',
  NOT_FOUND = 'NOT_FOUND',
  SERVER = 'SERVER',
  RATE_LIMIT = 'RATE_LIMIT',
  PAYMENT = 'PAYMENT',
  UNKNOWN = 'UNKNOWN',
}

export interface AppError {
  type: ErrorType
  message: string
  code?: string
  details?: any
  statusCode?: number
  timestamp: Date
  requestId?: string
}

// Error classification
export function classifyError(error: any): ErrorType {
  if (!error) return ErrorType.UNKNOWN

  // Network errors
  if (error.code === 'ECONNABORTED' || error.code === 'ENOTFOUND' || !navigator.onLine) {
    return ErrorType.NETWORK
  }

  // HTTP status based classification
  if (error.response?.status) {
    const status = error.response.status
    if (status === 401) return ErrorType.AUTHENTICATION
    if (status === 403) return ErrorType.PERMISSION
    if (status === 404) return ErrorType.NOT_FOUND
    if (status === 422 || status === 400) return ErrorType.VALIDATION
    if (status === 429) return ErrorType.RATE_LIMIT
    if (status >= 500) return ErrorType.SERVER
  }

  // Payment specific errors
  if (error.code?.startsWith('payment_') || error.type === 'StripeError') {
    return ErrorType.PAYMENT
  }

  return ErrorType.UNKNOWN
}

// User-friendly error messages
export function getUserMessage(errorType: ErrorType, defaultMessage?: string): string {
  const messages: Record<ErrorType, string> = {
    [ErrorType.NETWORK]: 'Connection error. Please check your internet and try again.',
    [ErrorType.AUTHENTICATION]: 'Please log in to continue.',
    [ErrorType.VALIDATION]: 'Please check your input and try again.',
    [ErrorType.PERMISSION]: 'You don\'t have permission to perform this action.',
    [ErrorType.NOT_FOUND]: 'The requested resource was not found.',
    [ErrorType.SERVER]: 'Something went wrong. Please try again later.',
    [ErrorType.RATE_LIMIT]: 'Too many requests. Please wait a moment and try again.',
    [ErrorType.PAYMENT]: 'Payment processing error. Please verify your payment details.',
    [ErrorType.UNKNOWN]: defaultMessage || 'An unexpected error occurred.',
  }

  return messages[errorType]
}

// Global error handler
export class ErrorHandler {
  private static instance: ErrorHandler
  private errorQueue: AppError[] = []
  private maxQueueSize = 100

  static getInstance(): ErrorHandler {
    if (!ErrorHandler.instance) {
      ErrorHandler.instance = new ErrorHandler()
    }
    return ErrorHandler.instance
  }

  handle(error: any, context?: { silent?: boolean; retry?: () => Promise<any> }): AppError {
    const errorType = classifyError(error)
    const appError: AppError = {
      type: errorType,
      message: error.response?.data?.message || error.message || getUserMessage(errorType),
      code: error.code || error.response?.data?.code,
      details: error.response?.data?.details,
      statusCode: error.response?.status,
      timestamp: new Date(),
      requestId: error.response?.headers?.['x-request-id'],
    }

    // Add to error queue
    this.addToQueue(appError)

    // Log to Sentry
    this.logToSentry(error, appError, context)

    // Show user notification (unless silent)
    if (!context?.silent) {
      this.notifyUser(appError, context?.retry)
    }

    return appError
  }

  private addToQueue(error: AppError) {
    this.errorQueue.push(error)
    if (this.errorQueue.length > this.maxQueueSize) {
      this.errorQueue.shift()
    }
  }

  private logToSentry(originalError: any, appError: AppError, context?: any) {
    Sentry.withScope((scope) => {
      scope.setTag('errorType', appError.type)
      scope.setLevel('error')
      scope.setContext('appError', {
        ...appError,
        timestamp: appError.timestamp.toISOString(),
      })
      
      if (context) {
        scope.setContext('errorContext', context)
      }

      if (appError.requestId) {
        scope.setTag('requestId', appError.requestId)
      }

      Sentry.captureException(originalError)
    })
  }

  private notifyUser(error: AppError, retry?: () => Promise<any>) {
    const message = getUserMessage(error.type, error.message)

    // Different notification styles based on error type
    switch (error.type) {
      case ErrorType.NETWORK:
      case ErrorType.SERVER:
        if (retry) {
          toast.error(
            (t) => (
              <div className="flex flex-col space-y-2">
                <span>{message}</span>
                <button
                  onClick={() => {
                    toast.dismiss(t.id)
                    retry()
                  }}
                  className="text-sm underline"
                >
                  Retry
                </button>
              </div>
            ),
            { duration: 5000 }
          )
        } else {
          toast.error(message)
        }
        break

      case ErrorType.AUTHENTICATION:
        toast.error(message, {
          duration: 4000,
          icon: '🔒',
        })
        // Redirect to login if needed
        if (typeof window !== 'undefined') {
          setTimeout(() => {
            window.location.href = '/login'
          }, 2000)
        }
        break

      case ErrorType.VALIDATION:
        toast.error(message, {
          duration: 3000,
          icon: '⚠️',
        })
        break

      case ErrorType.RATE_LIMIT:
        toast.error(message, {
          duration: 6000,
          icon: '⏱️',
        })
        break

      default:
        toast.error(message)
    }
  }

  getRecentErrors(): AppError[] {
    return [...this.errorQueue]
  }

  clearErrors() {
    this.errorQueue = []
  }
}

// React Error Boundary
export class ErrorBoundary extends React.Component<
  { children: React.ReactNode; fallback?: React.ComponentType<{ error: Error }> },
  { hasError: boolean; error?: Error }
> {
  constructor(props: any) {
    super(props)
    this.state = { hasError: false }
  }

  static getDerivedStateFromError(error: Error) {
    return { hasError: true, error }
  }

  componentDidCatch(error: Error, errorInfo: React.ErrorInfo) {
    console.error('Error Boundary caught:', error, errorInfo)
    
    Sentry.withScope((scope) => {
      scope.setTag('errorBoundary', true)
      scope.setContext('errorInfo', errorInfo)
      Sentry.captureException(error)
    })
  }

  render() {
    if (this.state.hasError) {
      if (this.props.fallback) {
        const FallbackComponent = this.props.fallback
        return <FallbackComponent error={this.state.error!} />
      }

      return (
        <div className="min-h-screen flex items-center justify-center p-4">
          <div className="max-w-md w-full text-center">
            <h2 className="text-2xl font-bold mb-4">Something went wrong</h2>
            <p className="text-muted-foreground mb-6">
              We're sorry for the inconvenience. Please try refreshing the page.
            </p>
            <button
              onClick={() => window.location.reload()}
              className="px-6 py-2 bg-primary text-primary-foreground rounded-lg hover:bg-primary/90 transition-colors"
            >
              Refresh Page
            </button>
          </div>
        </div>
      )
    }

    return this.props.children
  }
}

// API error interceptor
export function setupErrorInterceptor(apiClient: any) {
  apiClient.interceptors.response.use(
    (response: any) => response,
    (error: AxiosError) => {
      const errorHandler = ErrorHandler.getInstance()
      
      // Don't handle authentication errors here if they're being handled elsewhere
      if (error.response?.status === 401 && error.config?.url !== '/auth/refresh') {
        // Let auth interceptor handle this
        return Promise.reject(error)
      }

      // Handle all other errors
      errorHandler.handle(error, { silent: error.config?.headers?.['X-Silent-Error'] })
      
      return Promise.reject(error)
    }
  )
}

// Custom error hooks
export function useErrorHandler() {
  const errorHandler = ErrorHandler.getInstance()

  const handleError = React.useCallback(
    (error: any, options?: { silent?: boolean; retry?: () => Promise<any> }) => {
      return errorHandler.handle(error, options)
    },
    []
  )

  const clearErrors = React.useCallback(() => {
    errorHandler.clearErrors()
  }, [])

  const getRecentErrors = React.useCallback(() => {
    return errorHandler.getRecentErrors()
  }, [])

  return { handleError, clearErrors, getRecentErrors }
}

// Retry mechanism
export async function retryWithBackoff<T>(
  fn: () => Promise<T>,
  options: {
    maxRetries?: number
    initialDelay?: number
    maxDelay?: number
    factor?: number
    onRetry?: (attempt: number, error: any) => void
  } = {}
): Promise<T> {
  const {
    maxRetries = 3,
    initialDelay = 1000,
    maxDelay = 10000,
    factor = 2,
    onRetry,
  } = options

  let lastError: any
  let delay = initialDelay

  for (let attempt = 1; attempt <= maxRetries; attempt++) {
    try {
      return await fn()
    } catch (error) {
      lastError = error
      
      if (attempt === maxRetries) {
        throw error
      }

      // Don't retry on certain errors
      const errorType = classifyError(error)
      if ([ErrorType.AUTHENTICATION, ErrorType.PERMISSION, ErrorType.VALIDATION].includes(errorType)) {
        throw error
      }

      if (onRetry) {
        onRetry(attempt, error)
      }

      await new Promise(resolve => setTimeout(resolve, delay))
      delay = Math.min(delay * factor, maxDelay)
    }
  }

  throw lastError
}

// Form validation error handler
export function handleFormErrors(error: any, setError: (field: string, error: any) => void) {
  if (error.response?.data?.errors) {
    const errors = error.response.data.errors
    
    if (Array.isArray(errors)) {
      errors.forEach((err: any) => {
        if (err.field) {
          setError(err.field, {
            type: 'server',
            message: err.message,
          })
        }
      })
    } else if (typeof errors === 'object') {
      Object.entries(errors).forEach(([field, message]) => {
        setError(field, {
          type: 'server',
          message: message as string,
        })
      })
    }
  }
}

// Performance monitoring
export function trackPerformance(name: string, fn: () => void | Promise<void>) {
  const startTime = performance.now()
  
  const complete = () => {
    const duration = performance.now() - startTime
    
    // Log slow operations
    if (duration > 1000) {
      console.warn(`Slow operation: ${name} took ${duration.toFixed(2)}ms`)
      
      Sentry.withScope((scope) => {
        scope.setTag('performance', 'slow')
        scope.setContext('performance', {
          operation: name,
          duration,
        })
        Sentry.captureMessage(`Slow operation: ${name}`, 'warning')
      })
    }
    
    // Track in analytics
    if (typeof window !== 'undefined' && window.gtag) {
      window.gtag('event', 'timing_complete', {
        name,
        value: Math.round(duration),
      })
    }
  }

  const result = fn()
  
  if (result instanceof Promise) {
    return result.finally(complete)
  } else {
    complete()
    return result
  }
}
