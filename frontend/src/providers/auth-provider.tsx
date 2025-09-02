"use client"

import { useEffect } from 'react'
import { useRouter } from 'next/navigation'
import useAuthStore from '@/store/auth.store'
import wsClient from '@/services/websocket/client'

export function AuthProvider({ children }: { children: React.ReactNode }) {
  const router = useRouter()
  const { checkAuth, sessionExpiry, refreshSession } = useAuthStore()

  useEffect(() => {
    checkAuth()
  }, [])

  useEffect(() => {
    if (!sessionExpiry) return

    const checkSession = setInterval(() => {
      const timeLeft = sessionExpiry - Date.now()
      
      if (timeLeft < 5 * 60 * 1000 && timeLeft > 0) {
        // Refresh session 5 minutes before expiry
        refreshSession()
      } else if (timeLeft <= 0) {
        // Session expired
        router.push('/login')
      }
    }, 60000) // Check every minute

    return () => clearInterval(checkSession)
  }, [sessionExpiry, refreshSession, router])

  useEffect(() => {
    // Set up WebSocket reconnection on page visibility change
    const handleVisibilityChange = () => {
      if (document.visibilityState === 'visible' && !wsClient.isConnected()) {
        const token = localStorage.getItem('accessToken')
        if (token) {
          wsClient.connect(token)
        }
      }
    }

    document.addEventListener('visibilitychange', handleVisibilityChange)
    return () => document.removeEventListener('visibilitychange', handleVisibilityChange)
  }, [])

  return <>{children}</>
}
