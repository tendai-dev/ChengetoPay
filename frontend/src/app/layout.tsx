import type { Metadata } from 'next'
import { Inter } from 'next/font/google'
import '@/styles/globals.css'
import { ThemeProvider } from '@/providers/theme-provider'
import { Toaster } from 'react-hot-toast'
import { QueryProvider } from '@/providers/query-provider'
import { AuthProvider } from '@/providers/auth-provider'

const inter = Inter({ subsets: ['latin'] })

export const metadata: Metadata = {
  title: 'ChengetoPay - Next-Generation Payment Platform',
  description: 'Revolutionary payment infrastructure for the digital economy',
  keywords: 'payments, fintech, escrow, digital wallet, payment gateway',
  authors: [{ name: 'ChengetoPay' }],
  viewport: 'width=device-width, initial-scale=1, maximum-scale=1',
  themeColor: [
    { media: '(prefers-color-scheme: light)', color: '#ffffff' },
    { media: '(prefers-color-scheme: dark)', color: '#0f172a' },
  ],
  manifest: '/manifest.json',
  icons: {
    icon: '/favicon.ico',
    shortcut: '/favicon-16x16.png',
    apple: '/apple-touch-icon.png',
  },
  openGraph: {
    type: 'website',
    locale: 'en_US',
    url: 'https://chengetopay.com',
    title: 'ChengetoPay - Next-Generation Payment Platform',
    description: 'Revolutionary payment infrastructure for the digital economy',
    siteName: 'ChengetoPay',
  },
  twitter: {
    card: 'summary_large_image',
    title: 'ChengetoPay',
    description: 'Revolutionary payment infrastructure',
  },
}

export default function RootLayout({
  children,
}: {
  children: React.ReactNode
}) {
  return (
    <html lang="en" suppressHydrationWarning>
      <body className={inter.className}>
        <ThemeProvider
          attribute="class"
          defaultTheme="system"
          enableSystem
          disableTransitionOnChange
        >
          <QueryProvider>
            <AuthProvider>
              <div className="relative min-h-screen bg-background">
                {/* Futuristic background effects */}
                <div className="fixed inset-0 -z-10">
                  <div className="absolute inset-0 bg-gradient-to-br from-primary/5 via-transparent to-purple-600/5" />
                  <div className="absolute inset-0 cyber-grid opacity-[0.02]" />
                  <div className="absolute top-0 -left-4 w-96 h-96 bg-purple-500 rounded-full mix-blend-multiply filter blur-3xl opacity-10 animate-float" />
                  <div className="absolute top-0 -right-4 w-96 h-96 bg-yellow-500 rounded-full mix-blend-multiply filter blur-3xl opacity-10 animate-float animation-delay-2000" />
                  <div className="absolute -bottom-8 left-20 w-96 h-96 bg-pink-500 rounded-full mix-blend-multiply filter blur-3xl opacity-10 animate-float animation-delay-4000" />
                </div>
                {children}
              </div>
              <Toaster
                position="top-right"
                toastOptions={{
                  duration: 4000,
                  style: {
                    background: 'hsl(var(--background))',
                    color: 'hsl(var(--foreground))',
                    border: '1px solid hsl(var(--border))',
                    borderRadius: '0.75rem',
                    backdropFilter: 'blur(10px)',
                  },
                  success: {
                    iconTheme: {
                      primary: '#10b981',
                      secondary: '#ffffff',
                    },
                  },
                  error: {
                    iconTheme: {
                      primary: '#ef4444',
                      secondary: '#ffffff',
                    },
                  },
                }}
              />
            </AuthProvider>
          </QueryProvider>
        </ThemeProvider>
      </body>
    </html>
  )
}
