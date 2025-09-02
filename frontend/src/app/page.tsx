'use client'

import { useEffect } from 'react'
import { useRouter } from 'next/navigation'
import useAuthStore from '@/store/auth.store'
import { Button } from '@/components/ui/button'
import { Card } from '@/components/ui/card'
import { ArrowRight, Shield, Zap, Globe, Lock, TrendingUp, CreditCard } from 'lucide-react'
import Link from 'next/link'
import CountUp from 'react-countup'
import { useInView } from 'react-intersection-observer'

export default function HomePage() {
  const router = useRouter()
  const { isAuthenticated } = useAuthStore()
  const { ref, inView } = useInView({ triggerOnce: true })

  useEffect(() => {
    if (isAuthenticated) {
      router.push('/dashboard')
    }
  }, [isAuthenticated, router])

  const features = [
    {
      icon: <Shield className="w-8 h-8" />,
      title: 'Bank-Grade Security',
      description: 'Advanced encryption and multi-factor authentication protect every transaction',
      gradient: 'from-blue-500 to-cyan-500',
    },
    {
      icon: <Zap className="w-8 h-8" />,
      title: 'Lightning Fast',
      description: 'Process payments in milliseconds with our optimized infrastructure',
      gradient: 'from-purple-500 to-pink-500',
    },
    {
      icon: <Globe className="w-8 h-8" />,
      title: 'Global Reach',
      description: 'Accept payments from anywhere in the world, in any currency',
      gradient: 'from-green-500 to-emerald-500',
    },
    {
      icon: <Lock className="w-8 h-8" />,
      title: 'Secure Escrow',
      description: 'Built-in escrow service for secure transactions between parties',
      gradient: 'from-orange-500 to-red-500',
    },
    {
      icon: <TrendingUp className="w-8 h-8" />,
      title: 'Real-time Analytics',
      description: 'Monitor your business performance with advanced analytics dashboard',
      gradient: 'from-indigo-500 to-purple-500',
    },
    {
      icon: <CreditCard className="w-8 h-8" />,
      title: 'Multiple Payment Methods',
      description: 'Support for cards, wallets, bank transfers, and cryptocurrencies',
      gradient: 'from-pink-500 to-rose-500',
    },
  ]

  const stats = [
    { value: 99.99, suffix: '%', label: 'Uptime' },
    { value: 500, suffix: 'M+', label: 'Transactions' },
    { value: 150, suffix: '+', label: 'Countries' },
    { value: 2.5, suffix: 's', label: 'Avg Response Time' },
  ]

  return (
    <div className="min-h-screen">
      {/* Navigation */}
      <nav className="fixed top-0 left-0 right-0 z-50 backdrop-blur-xl bg-background/80 border-b border-border/50">
        <div className="container mx-auto px-4">
          <div className="flex items-center justify-between h-16">
            <div className="flex items-center space-x-2">
              <div className="w-10 h-10 rounded-lg bg-gradient-to-br from-primary to-purple-600 flex items-center justify-center">
                <span className="text-white font-bold text-xl">C</span>
              </div>
              <span className="text-xl font-bold gradient-text">ChengetoPay</span>
            </div>
            <div className="flex items-center space-x-4">
              <Link href="/login">
                <Button variant="ghost">Sign In</Button>
              </Link>
              <Link href="/register">
                <Button variant="gradient">Get Started</Button>
              </Link>
            </div>
          </div>
        </div>
      </nav>

      {/* Hero Section */}
      <section className="pt-32 pb-20 px-4">
        <div className="container mx-auto max-w-6xl">
          <div className="text-center space-y-6">
            <h1 className="text-5xl md:text-7xl font-bold">
              <span className="gradient-text">Next-Generation</span>
              <br />
              Payment Infrastructure
            </h1>
            <p className="text-xl text-muted-foreground max-w-2xl mx-auto">
              Revolutionary payment platform built for the digital economy. 
              Process payments, manage escrows, and scale your business globally.
            </p>
            <div className="flex flex-col sm:flex-row gap-4 justify-center pt-8">
              <Link href="/register">
                <Button size="xl" variant="futuristic" className="group">
                  Start Building
                  <ArrowRight className="ml-2 w-5 h-5 group-hover:translate-x-1 transition-transform" />
                </Button>
              </Link>
              <Link href="/docs">
                <Button size="xl" variant="outline">
                  View Documentation
                </Button>
              </Link>
            </div>
          </div>
        </div>
      </section>

      {/* Stats Section */}
      <section className="py-20 px-4" ref={ref}>
        <div className="container mx-auto max-w-6xl">
          <div className="grid grid-cols-2 md:grid-cols-4 gap-8">
            {stats.map((stat, index) => (
              <div key={index} className="text-center">
                <div className="text-4xl font-bold gradient-text">
                  {inView && (
                    <CountUp
                      end={stat.value}
                      duration={2.5}
                      decimals={stat.suffix === '%' ? 2 : 0}
                      suffix={stat.suffix}
                    />
                  )}
                </div>
                <p className="text-muted-foreground mt-2">{stat.label}</p>
              </div>
            ))}
          </div>
        </div>
      </section>

      {/* Features Grid */}
      <section className="py-20 px-4">
        <div className="container mx-auto max-w-6xl">
          <div className="text-center mb-12">
            <h2 className="text-4xl font-bold mb-4">Cutting-Edge Features</h2>
            <p className="text-muted-foreground text-lg">
              Everything you need to build and scale your payment infrastructure
            </p>
          </div>
          <div className="grid md:grid-cols-2 lg:grid-cols-3 gap-6">
            {features.map((feature, index) => (
              <Card
                key={index}
                variant="futuristic"
                className="p-6 hover:scale-105 transition-transform duration-300"
              >
                <div className={`w-16 h-16 rounded-xl bg-gradient-to-br ${feature.gradient} flex items-center justify-center text-white mb-4`}>
                  {feature.icon}
                </div>
                <h3 className="text-xl font-semibold mb-2">{feature.title}</h3>
                <p className="text-muted-foreground">{feature.description}</p>
              </Card>
            ))}
          </div>
        </div>
      </section>

      {/* CTA Section */}
      <section className="py-20 px-4">
        <div className="container mx-auto max-w-4xl">
          <Card variant="gradient" className="p-12 text-center">
            <h2 className="text-4xl font-bold mb-4">Ready to Transform Your Business?</h2>
            <p className="text-lg text-muted-foreground mb-8">
              Join thousands of businesses already using ChengetoPay
            </p>
            <Link href="/register">
              <Button size="xl" variant="glow" className="shadow-2xl">
                Get Started For Free
                <ArrowRight className="ml-2 w-5 h-5" />
              </Button>
            </Link>
          </Card>
        </div>
      </section>

      {/* Footer */}
      <footer className="py-12 px-4 border-t border-border/50">
        <div className="container mx-auto max-w-6xl">
          <div className="text-center text-muted-foreground">
            <p>&copy; 2024 ChengetoPay. All rights reserved.</p>
          </div>
        </div>
      </footer>
    </div>
  )
}
