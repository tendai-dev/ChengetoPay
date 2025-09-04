import { useEffect, useCallback, useRef } from 'react'
import { useInView } from 'react-intersection-observer'
import { useQueryClient } from '@tanstack/react-query'
import { prefetchHelpers, memoryCache, CACHE_TIME } from '@/lib/cache'

// Virtual scrolling hook for large lists
export function useVirtualScroll<T>(
  items: T[],
  itemHeight: number,
  containerHeight: number
) {
  const scrollTop = useRef(0)
  const startIndex = Math.floor(scrollTop.current / itemHeight)
  const endIndex = Math.min(
    startIndex + Math.ceil(containerHeight / itemHeight) + 1,
    items.length
  )

  const visibleItems = items.slice(startIndex, endIndex)
  const totalHeight = items.length * itemHeight
  const offsetY = startIndex * itemHeight

  const handleScroll = useCallback((e: React.UIEvent<HTMLDivElement>) => {
    scrollTop.current = e.currentTarget.scrollTop
  }, [])

  return {
    visibleItems,
    totalHeight,
    offsetY,
    handleScroll,
  }
}

// Lazy loading hook with prefetching
export function useLazyLoad(
  fetchFn: () => Promise<any>,
  options?: {
    threshold?: number
    rootMargin?: string
    prefetch?: boolean
  }
) {
  const { ref, inView } = useInView({
    threshold: options?.threshold || 0.1,
    rootMargin: options?.rootMargin || '100px',
    triggerOnce: true,
  })

  const queryClient = useQueryClient()
  const hasPrefetched = useRef(false)

  useEffect(() => {
    if (inView && options?.prefetch && !hasPrefetched.current) {
      hasPrefetched.current = true
      queryClient.prefetchQuery({
        queryKey: ['prefetch', fetchFn.toString()],
        queryFn: fetchFn,
      })
    }
  }, [inView, fetchFn, options?.prefetch, queryClient])

  return { ref, inView }
}

// Debounced search hook
export function useDebouncedSearch(
  searchFn: (query: string) => void,
  delay: number = 300
) {
  const timeoutRef = useRef<NodeJS.Timeout>()

  const debouncedSearch = useCallback((query: string) => {
    if (timeoutRef.current) {
      clearTimeout(timeoutRef.current)
    }

    timeoutRef.current = setTimeout(() => {
      searchFn(query)
    }, delay)
  }, [searchFn, delay])

  useEffect(() => {
    return () => {
      if (timeoutRef.current) {
        clearTimeout(timeoutRef.current)
      }
    }
  }, [])

  return debouncedSearch
}

// Image optimization hook
export function useOptimizedImage(
  src: string,
  options?: {
    width?: number
    height?: number
    quality?: number
    format?: 'webp' | 'avif' | 'auto'
  }
) {
  const optimizedSrc = useCallback(() => {
    if (!src) return ''
    
    // If using Next.js image optimization
    const params = new URLSearchParams()
    if (options?.width) params.append('w', options.width.toString())
    if (options?.height) params.append('h', options.height.toString())
    if (options?.quality) params.append('q', options.quality.toString())
    if (options?.format) params.append('fm', options.format)

    return `/_next/image?url=${encodeURIComponent(src)}&${params.toString()}`
  }, [src, options])

  const isLoading = useRef(true)
  const error = useRef<Error | null>(null)

  const handleLoad = useCallback(() => {
    isLoading.current = false
  }, [])

  const handleError = useCallback((e: Error) => {
    isLoading.current = false
    error.current = e
  }, [])

  return {
    src: optimizedSrc(),
    isLoading: isLoading.current,
    error: error.current,
    onLoad: handleLoad,
    onError: handleError,
  }
}

// Performance monitoring hook
export function usePerformanceMonitor(componentName: string) {
  const renderCount = useRef(0)
  const renderTime = useRef<number>(0)
  const mountTime = useRef<number>(0)

  useEffect(() => {
    mountTime.current = performance.now()
    
    return () => {
      const unmountTime = performance.now()
      const totalLifetime = unmountTime - mountTime.current
      
      // Log performance metrics
      if (process.env.NODE_ENV === 'development') {
        console.log(`[Performance] ${componentName}:`, {
          renders: renderCount.current,
          avgRenderTime: renderTime.current / renderCount.current,
          totalLifetime,
        })
      }

      // Send to analytics (commented out for now)
      // if (typeof window !== 'undefined' && (window as any).gtag) {
      //   (window as any).gtag('event', 'performance', {
      //     component: componentName,
      //     renders: renderCount.current,
      //     lifetime: totalLifetime,
      //   })
      // }
    }
  }, [componentName])

  useEffect(() => {
    renderCount.current++
    const startTime = performance.now()
    
    return () => {
      renderTime.current += performance.now() - startTime
    }
  })
}

// Request batching hook
export function useBatchRequests<T>(
  batchSize: number = 10,
  delay: number = 100
) {
  const queue = useRef<Array<{
    request: () => Promise<T>
    resolve: (value: T) => void
    reject: (error: any) => void
  }>>([])
  const timeoutRef = useRef<NodeJS.Timeout>()

  const processBatch = useCallback(async () => {
    const batch = queue.current.splice(0, batchSize)
    if (batch.length === 0) return

    try {
      const results = await Promise.all(batch.map(item => item.request()))
      batch.forEach((item, index) => item.resolve(results[index]))
    } catch (error) {
      batch.forEach(item => item.reject(error))
    }
  }, [batchSize])

  const addRequest = useCallback((request: () => Promise<T>): Promise<T> => {
    return new Promise((resolve, reject) => {
      queue.current.push({ request, resolve, reject })

      if (timeoutRef.current) {
        clearTimeout(timeoutRef.current)
      }

      timeoutRef.current = setTimeout(processBatch, delay)
    })
  }, [processBatch, delay])

  return { addRequest }
}

// Memory leak prevention hook
export function useMemoryCleanup() {
  const subscriptions = useRef<Set<() => void>>(new Set())
  const intervals = useRef<Set<NodeJS.Timeout>>(new Set())
  const timeouts = useRef<Set<NodeJS.Timeout>>(new Set())

  const addSubscription = useCallback((cleanup: () => void) => {
    subscriptions.current.add(cleanup)
    return () => subscriptions.current.delete(cleanup)
  }, [])

  const addInterval = useCallback((id: NodeJS.Timeout) => {
    intervals.current.add(id)
    return () => {
      clearInterval(id)
      intervals.current.delete(id)
    }
  }, [])

  const addTimeout = useCallback((id: NodeJS.Timeout) => {
    timeouts.current.add(id)
    return () => {
      clearTimeout(id)
      timeouts.current.delete(id)
    }
  }, [])

  useEffect(() => {
    return () => {
      // Clean up all subscriptions
      subscriptions.current.forEach(cleanup => cleanup())
      subscriptions.current.clear()

      // Clear all intervals
      intervals.current.forEach(id => clearInterval(id))
      intervals.current.clear()

      // Clear all timeouts
      timeouts.current.forEach(id => clearTimeout(id))
      timeouts.current.clear()
    }
  }, [])

  return { addSubscription, addInterval, addTimeout }
}

// Resource preloading hook
export function useResourcePreloader() {
  const preloadImage = useCallback((src: string) => {
    const img = new Image()
    img.src = src
    memoryCache.set(`preload_img_${src}`, true, CACHE_TIME.VERY_LONG)
  }, [])

  const preloadScript = useCallback((src: string) => {
    const link = document.createElement('link')
    link.rel = 'preload'
    link.as = 'script'
    link.href = src
    document.head.appendChild(link)
  }, [])

  const preloadStyle = useCallback((src: string) => {
    const link = document.createElement('link')
    link.rel = 'preload'
    link.as = 'style'
    link.href = src
    document.head.appendChild(link)
  }, [])

  const preloadFont = useCallback((src: string) => {
    const link = document.createElement('link')
    link.rel = 'preload'
    link.as = 'font'
    link.type = 'font/woff2'
    link.href = src
    link.crossOrigin = 'anonymous'
    document.head.appendChild(link)
  }, [])

  return { preloadImage, preloadScript, preloadStyle, preloadFont }
}
