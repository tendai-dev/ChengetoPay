import { QueryClient } from '@tanstack/react-query'

// Cache time constants
export const CACHE_TIME = {
  INSTANT: 0,
  SHORT: 1000 * 30, // 30 seconds
  MEDIUM: 1000 * 60 * 5, // 5 minutes
  LONG: 1000 * 60 * 30, // 30 minutes
  VERY_LONG: 1000 * 60 * 60, // 1 hour
  PERMANENT: Infinity,
}

export const STALE_TIME = {
  INSTANT: 0,
  SHORT: 1000 * 10, // 10 seconds
  MEDIUM: 1000 * 60 * 2, // 2 minutes
  LONG: 1000 * 60 * 10, // 10 minutes
  VERY_LONG: 1000 * 60 * 30, // 30 minutes
}

// Query keys factory
export const queryKeys = {
  all: ['app'] as const,
  auth: {
    all: ['auth'] as const,
    user: () => [...queryKeys.auth.all, 'user'] as const,
    session: () => [...queryKeys.auth.all, 'session'] as const,
  },
  transactions: {
    all: ['transactions'] as const,
    list: (filters?: any) => [...queryKeys.transactions.all, 'list', filters] as const,
    detail: (id: string) => [...queryKeys.transactions.all, 'detail', id] as const,
  },
  wallet: {
    all: ['wallet'] as const,
    balance: () => [...queryKeys.wallet.all, 'balance'] as const,
    history: () => [...queryKeys.wallet.all, 'history'] as const,
  },
  payments: {
    all: ['payments'] as const,
    list: (filters?: any) => [...queryKeys.payments.all, 'list', filters] as const,
    detail: (id: string) => [...queryKeys.payments.all, 'detail', id] as const,
  },
  escrow: {
    all: ['escrow'] as const,
    list: (status?: string) => [...queryKeys.escrow.all, 'list', status] as const,
    detail: (id: string) => [...queryKeys.escrow.all, 'detail', id] as const,
  },
  merchants: {
    all: ['merchants'] as const,
    list: () => [...queryKeys.merchants.all, 'list'] as const,
    detail: (id: string) => [...queryKeys.merchants.all, 'detail', id] as const,
  },
  analytics: {
    all: ['analytics'] as const,
    dashboard: () => [...queryKeys.analytics.all, 'dashboard'] as const,
    revenue: (range?: string) => [...queryKeys.analytics.all, 'revenue', range] as const,
    transactions: (range?: string) => [...queryKeys.analytics.all, 'transactions', range] as const,
  },
}

// Optimistic update helpers
export const optimisticUpdate = {
  addToList: <T>(
    queryClient: QueryClient,
    queryKey: readonly unknown[],
    newItem: T
  ) => {
    queryClient.setQueryData(queryKey, (old: T[] | undefined) => {
      if (!old) return [newItem]
      return [newItem, ...old]
    })
  },

  updateInList: <T extends { id: string }>(
    queryClient: QueryClient,
    queryKey: readonly unknown[],
    id: string,
    updates: Partial<T>
  ) => {
    queryClient.setQueryData(queryKey, (old: T[] | undefined) => {
      if (!old) return old
      return old.map(item => 
        item.id === id ? { ...item, ...updates } : item
      )
    })
  },

  removeFromList: <T extends { id: string }>(
    queryClient: QueryClient,
    queryKey: readonly unknown[],
    id: string
  ) => {
    queryClient.setQueryData(queryKey, (old: T[] | undefined) => {
      if (!old) return old
      return old.filter(item => item.id !== id)
    })
  },
}

// Prefetch helpers
export const prefetchHelpers = {
  prefetchNextPage: async (
    queryClient: QueryClient,
    queryKey: readonly unknown[],
    fetchFn: () => Promise<any>,
    page: number
  ) => {
    await queryClient.prefetchQuery({
      queryKey: [...queryKey, page + 1],
      queryFn: fetchFn,
      staleTime: STALE_TIME.MEDIUM,
    })
  },

  prefetchRelated: async (
    queryClient: QueryClient,
    queries: Array<{
      queryKey: readonly unknown[]
      queryFn: () => Promise<any>
    }>
  ) => {
    await Promise.all(
      queries.map(({ queryKey, queryFn }) =>
        queryClient.prefetchQuery({
          queryKey,
          queryFn,
          staleTime: STALE_TIME.LONG,
        })
      )
    )
  },
}

// Cache invalidation patterns
export const cacheInvalidation = {
  invalidatePattern: (
    queryClient: QueryClient,
    pattern: string[]
  ) => {
    queryClient.invalidateQueries({ queryKey: pattern })
  },

  invalidateMultiple: (
    queryClient: QueryClient,
    patterns: string[][]
  ) => {
    patterns.forEach(pattern => {
      queryClient.invalidateQueries({ queryKey: pattern })
    })
  },

  smartInvalidate: (
    queryClient: QueryClient,
    entity: 'transaction' | 'payment' | 'wallet' | 'escrow',
    action: 'create' | 'update' | 'delete'
  ) => {
    const invalidationMap = {
      transaction: {
        create: [queryKeys.transactions.all, queryKeys.wallet.all, queryKeys.analytics.all],
        update: [queryKeys.transactions.all],
        delete: [queryKeys.transactions.all, queryKeys.wallet.all],
      },
      payment: {
        create: [queryKeys.payments.all, queryKeys.wallet.all, queryKeys.transactions.all],
        update: [queryKeys.payments.all],
        delete: [queryKeys.payments.all, queryKeys.wallet.all],
      },
      wallet: {
        create: [queryKeys.wallet.all],
        update: [queryKeys.wallet.all, queryKeys.analytics.all],
        delete: [queryKeys.wallet.all],
      },
      escrow: {
        create: [queryKeys.escrow.all, queryKeys.wallet.all],
        update: [queryKeys.escrow.all],
        delete: [queryKeys.escrow.all, queryKeys.wallet.all],
      },
    }

    const patterns = invalidationMap[entity][action]
    cacheInvalidation.invalidateMultiple(queryClient, patterns.map(pattern => [...pattern]))
  },
}

// Local storage cache
export const localCache = {
  set: (key: string, data: any, ttl?: number) => {
    const item = {
      data,
      timestamp: Date.now(),
      ttl,
    }
    localStorage.setItem(`cache_${key}`, JSON.stringify(item))
  },

  get: <T>(key: string): T | null => {
    const itemStr = localStorage.getItem(`cache_${key}`)
    if (!itemStr) return null

    const item = JSON.parse(itemStr)
    if (item.ttl && Date.now() - item.timestamp > item.ttl) {
      localStorage.removeItem(`cache_${key}`)
      return null
    }

    return item.data as T
  },

  remove: (key: string) => {
    localStorage.removeItem(`cache_${key}`)
  },

  clear: () => {
    Object.keys(localStorage)
      .filter(key => key.startsWith('cache_'))
      .forEach(key => localStorage.removeItem(key))
  },
}

// Session storage cache (for sensitive data)
export const sessionCache = {
  set: (key: string, data: any) => {
    sessionStorage.setItem(`session_${key}`, JSON.stringify(data))
  },

  get: <T>(key: string): T | null => {
    const itemStr = sessionStorage.getItem(`session_${key}`)
    if (!itemStr) return null
    return JSON.parse(itemStr) as T
  },

  remove: (key: string) => {
    sessionStorage.removeItem(`session_${key}`)
  },

  clear: () => {
    Object.keys(sessionStorage)
      .filter(key => key.startsWith('session_'))
      .forEach(key => sessionStorage.removeItem(key))
  },
}

// Memory cache for runtime data
class MemoryCache {
  private cache = new Map<string, { data: any; expiry: number }>()

  set(key: string, data: any, ttl: number = CACHE_TIME.MEDIUM) {
    const expiry = Date.now() + ttl
    this.cache.set(key, { data, expiry })
  }

  get<T>(key: string): T | null {
    const item = this.cache.get(key)
    if (!item) return null

    if (Date.now() > item.expiry) {
      this.cache.delete(key)
      return null
    }

    return item.data as T
  }

  delete(key: string) {
    this.cache.delete(key)
  }

  clear() {
    this.cache.clear()
  }

  cleanup() {
    const now = Date.now()
    for (const [key, item] of this.cache.entries()) {
      if (now > item.expiry) {
        this.cache.delete(key)
      }
    }
  }
}

export const memoryCache = new MemoryCache()

// Cleanup expired cache periodically
if (typeof window !== 'undefined') {
  setInterval(() => {
    memoryCache.cleanup()
  }, 60000) // Clean up every minute
}
