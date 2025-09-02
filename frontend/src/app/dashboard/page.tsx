'use client'

import { useEffect, useState } from 'react'
import { useRouter } from 'next/navigation'
import useAuthStore from '@/store/auth.store'
import { Card } from '@/components/ui/card'
import { Button } from '@/components/ui/button'
import { 
  TrendingUp, DollarSign, CreditCard, Users, ArrowUpRight, ArrowDownRight,
  Activity, Shield, Bell, Settings, LogOut, Menu, X, Plus, Send, Receipt,
  Wallet, Globe, Clock, AlertCircle, CheckCircle, XCircle
} from 'lucide-react'
import CountUp from 'react-countup'
import { Line, Bar, Doughnut } from 'react-chartjs-2'
import {
  Chart as ChartJS,
  CategoryScale,
  LinearScale,
  PointElement,
  LineElement,
  BarElement,
  ArcElement,
  Title,
  Tooltip,
  Legend,
  Filler
} from 'chart.js'
import wsClient from '@/services/websocket/client'
import { format } from 'date-fns'
import { motion, AnimatePresence } from 'framer-motion'

ChartJS.register(
  CategoryScale,
  LinearScale,
  PointElement,
  LineElement,
  BarElement,
  ArcElement,
  Title,
  Tooltip,
  Legend,
  Filler
)

interface Transaction {
  id: string
  type: 'payment' | 'transfer' | 'escrow'
  amount: number
  currency: string
  status: 'pending' | 'completed' | 'failed'
  merchant: string
  timestamp: Date
}

interface Stats {
  totalRevenue: number
  totalTransactions: number
  activeUsers: number
  escrowBalance: number
  revenueChange: number
  transactionChange: number
  userChange: number
  escrowChange: number
}

export default function DashboardPage() {
  const router = useRouter()
  const { user, isAuthenticated, logout } = useAuthStore()
  const [sidebarOpen, setSidebarOpen] = useState(false)
  const [recentTransactions, setRecentTransactions] = useState<Transaction[]>([])
  const [stats, setStats] = useState<Stats>({
    totalRevenue: 1234567.89,
    totalTransactions: 45678,
    activeUsers: 12345,
    escrowBalance: 567890.12,
    revenueChange: 12.5,
    transactionChange: 8.3,
    userChange: 15.7,
    escrowChange: -3.2,
  })

  useEffect(() => {
    if (!isAuthenticated) {
      router.push('/login')
      return
    }

    // Connect WebSocket
    const token = localStorage.getItem('accessToken')
    if (token) {
      wsClient.connect(token)
      
      // Subscribe to real-time updates
      wsClient.subscribeToPaymentUpdates((data) => {
        setRecentTransactions(prev => [data, ...prev].slice(0, 10))
        setStats(prev => ({
          ...prev,
          totalRevenue: prev.totalRevenue + data.amount,
          totalTransactions: prev.totalTransactions + 1,
        }))
      })
    }

    return () => {
      wsClient.disconnect()
    }
  }, [isAuthenticated, router])

  // Mock recent transactions
  useEffect(() => {
    setRecentTransactions([
      { id: '1', type: 'payment', amount: 1250.00, currency: 'USD', status: 'completed', merchant: 'Acme Corp', timestamp: new Date() },
      { id: '2', type: 'escrow', amount: 5000.00, currency: 'USD', status: 'pending', merchant: 'Tech Solutions', timestamp: new Date() },
      { id: '3', type: 'transfer', amount: 750.50, currency: 'EUR', status: 'completed', merchant: 'Global Trade', timestamp: new Date() },
      { id: '4', type: 'payment', amount: 320.00, currency: 'GBP', status: 'failed', merchant: 'Digital Services', timestamp: new Date() },
      { id: '5', type: 'payment', amount: 890.25, currency: 'USD', status: 'completed', merchant: 'Cloud Systems', timestamp: new Date() },
    ])
  }, [])

  const chartData = {
    revenue: {
      labels: ['Jan', 'Feb', 'Mar', 'Apr', 'May', 'Jun'],
      datasets: [{
        label: 'Revenue',
        data: [65000, 72000, 68000, 85000, 92000, 98000],
        borderColor: 'rgb(99, 102, 241)',
        backgroundColor: 'rgba(99, 102, 241, 0.1)',
        tension: 0.4,
        fill: true,
      }]
    },
    transactions: {
      labels: ['Mon', 'Tue', 'Wed', 'Thu', 'Fri', 'Sat', 'Sun'],
      datasets: [{
        label: 'Transactions',
        data: [1200, 1900, 1500, 2100, 2300, 1800, 2500],
        backgroundColor: 'rgba(168, 85, 247, 0.8)',
        borderRadius: 8,
      }]
    },
    distribution: {
      labels: ['Payments', 'Transfers', 'Escrow'],
      datasets: [{
        data: [65, 25, 10],
        backgroundColor: ['rgb(99, 102, 241)', 'rgb(168, 85, 247)', 'rgb(236, 72, 153)'],
        borderWidth: 0,
      }]
    }
  }

  const chartOptions = {
    responsive: true,
    maintainAspectRatio: false,
    plugins: {
      legend: {
        display: false,
      },
      tooltip: {
        backgroundColor: 'rgba(0, 0, 0, 0.8)',
        padding: 12,
        borderRadius: 8,
      }
    },
    scales: {
      x: {
        grid: {
          display: false,
        },
        ticks: {
          color: 'rgba(156, 163, 175, 0.8)',
        }
      },
      y: {
        grid: {
          color: 'rgba(156, 163, 175, 0.1)',
        },
        ticks: {
          color: 'rgba(156, 163, 175, 0.8)',
        }
      }
    }
  }

  const statCards = [
    {
      title: 'Total Revenue',
      value: stats.totalRevenue,
      change: stats.revenueChange,
      icon: DollarSign,
      prefix: '$',
      gradient: 'from-blue-500 to-cyan-500',
    },
    {
      title: 'Transactions',
      value: stats.totalTransactions,
      change: stats.transactionChange,
      icon: CreditCard,
      gradient: 'from-purple-500 to-pink-500',
    },
    {
      title: 'Active Users',
      value: stats.activeUsers,
      change: stats.userChange,
      icon: Users,
      gradient: 'from-green-500 to-emerald-500',
    },
    {
      title: 'Escrow Balance',
      value: stats.escrowBalance,
      change: stats.escrowChange,
      icon: Shield,
      prefix: '$',
      gradient: 'from-orange-500 to-red-500',
    },
  ]

  const sidebarItems = [
    { icon: Activity, label: 'Dashboard', active: true, href: '/dashboard' },
    { icon: CreditCard, label: 'Payments', href: '/payments' },
    { icon: Send, label: 'Transfers', href: '/transfers' },
    { icon: Shield, label: 'Escrow', href: '/escrow' },
    { icon: Receipt, label: 'Transactions', href: '/transactions' },
    { icon: Wallet, label: 'Wallet', href: '/wallet' },
    { icon: Globe, label: 'API Keys', href: '/api-keys' },
    { icon: Settings, label: 'Settings', href: '/settings' },
  ]

  return (
    <div className="min-h-screen bg-background">
      {/* Sidebar */}
      <AnimatePresence>
        {(sidebarOpen || window.innerWidth >= 1024) && (
          <motion.aside
            initial={{ x: -320 }}
            animate={{ x: 0 }}
            exit={{ x: -320 }}
            className="fixed left-0 top-0 bottom-0 w-64 bg-card border-r border-border/50 z-40 lg:z-30"
          >
            <div className="p-6">
              <div className="flex items-center space-x-2 mb-8">
                <div className="w-10 h-10 rounded-lg bg-gradient-to-br from-primary to-purple-600 flex items-center justify-center">
                  <span className="text-white font-bold text-xl">C</span>
                </div>
                <span className="text-xl font-bold gradient-text">ChengetoPay</span>
              </div>
              <nav className="space-y-2">
                {sidebarItems.map((item) => (
                  <Button
                    key={item.label}
                    variant={item.active ? 'gradient' : 'ghost'}
                    className="w-full justify-start"
                    onClick={() => router.push(item.href)}
                  >
                    <item.icon className="w-4 h-4 mr-3" />
                    {item.label}
                  </Button>
                ))}
              </nav>
            </div>
            <div className="absolute bottom-0 left-0 right-0 p-6 border-t border-border/50">
              <Button
                variant="ghost"
                className="w-full justify-start text-destructive hover:text-destructive"
                onClick={logout}
              >
                <LogOut className="w-4 h-4 mr-3" />
                Logout
              </Button>
            </div>
          </motion.aside>
        )}
      </AnimatePresence>

      {/* Main Content */}
      <div className="lg:pl-64">
        {/* Header */}
        <header className="sticky top-0 z-30 backdrop-blur-xl bg-background/80 border-b border-border/50">
          <div className="flex items-center justify-between p-6">
            <div className="flex items-center space-x-4">
              <Button
                variant="ghost"
                size="icon"
                className="lg:hidden"
                onClick={() => setSidebarOpen(!sidebarOpen)}
              >
                {sidebarOpen ? <X className="w-5 h-5" /> : <Menu className="w-5 h-5" />}
              </Button>
              <div>
                <h1 className="text-2xl font-bold">Dashboard</h1>
                <p className="text-sm text-muted-foreground">
                  Welcome back, {user?.firstName || 'User'}
                </p>
              </div>
            </div>
            <div className="flex items-center space-x-4">
              <Button variant="ghost" size="icon" className="relative">
                <Bell className="w-5 h-5" />
                <span className="absolute top-1 right-1 w-2 h-2 bg-red-500 rounded-full" />
              </Button>
              <Button variant="gradient">
                <Plus className="w-4 h-4 mr-2" />
                New Payment
              </Button>
            </div>
          </div>
        </header>

        {/* Dashboard Content */}
        <div className="p-6 space-y-6">
          {/* Stats Grid */}
          <div className="grid grid-cols-1 md:grid-cols-2 xl:grid-cols-4 gap-6">
            {statCards.map((stat, index) => (
              <motion.div
                key={stat.title}
                initial={{ opacity: 0, y: 20 }}
                animate={{ opacity: 1, y: 0 }}
                transition={{ delay: index * 0.1 }}
              >
                <Card variant="futuristic" className="p-6">
                  <div className="flex items-center justify-between mb-4">
                    <div className={`w-12 h-12 rounded-xl bg-gradient-to-br ${stat.gradient} flex items-center justify-center text-white`}>
                      <stat.icon className="w-6 h-6" />
                    </div>
                    <div className={`flex items-center text-sm ${stat.change >= 0 ? 'text-green-500' : 'text-red-500'}`}>
                      {stat.change >= 0 ? <ArrowUpRight className="w-4 h-4" /> : <ArrowDownRight className="w-4 h-4" />}
                      {Math.abs(stat.change)}%
                    </div>
                  </div>
                  <p className="text-sm text-muted-foreground mb-1">{stat.title}</p>
                  <p className="text-2xl font-bold">
                    {stat.prefix}
                    <CountUp end={stat.value} duration={2} separator="," decimals={stat.prefix ? 2 : 0} />
                  </p>
                </Card>
              </motion.div>
            ))}
          </div>

          {/* Charts Row */}
          <div className="grid grid-cols-1 lg:grid-cols-3 gap-6">
            <Card variant="glass" className="lg:col-span-2 p-6">
              <h3 className="text-lg font-semibold mb-4">Revenue Overview</h3>
              <div className="h-64">
                <Line data={chartData.revenue} options={chartOptions} />
              </div>
            </Card>
            <Card variant="glass" className="p-6">
              <h3 className="text-lg font-semibold mb-4">Transaction Distribution</h3>
              <div className="h-64">
                <Doughnut data={chartData.distribution} options={{ ...chartOptions, cutout: '70%' }} />
              </div>
            </Card>
          </div>

          {/* Recent Transactions */}
          <Card variant="glass" className="p-6">
            <div className="flex items-center justify-between mb-6">
              <h3 className="text-lg font-semibold">Recent Transactions</h3>
              <Button variant="outline" size="sm">View All</Button>
            </div>
            <div className="overflow-x-auto">
              <table className="w-full">
                <thead>
                  <tr className="border-b border-border/50">
                    <th className="text-left py-3 px-4 text-sm font-medium text-muted-foreground">Type</th>
                    <th className="text-left py-3 px-4 text-sm font-medium text-muted-foreground">Merchant</th>
                    <th className="text-left py-3 px-4 text-sm font-medium text-muted-foreground">Amount</th>
                    <th className="text-left py-3 px-4 text-sm font-medium text-muted-foreground">Status</th>
                    <th className="text-left py-3 px-4 text-sm font-medium text-muted-foreground">Time</th>
                  </tr>
                </thead>
                <tbody>
                  {recentTransactions.map((transaction) => (
                    <tr key={transaction.id} className="border-b border-border/30 hover:bg-muted/20 transition-colors">
                      <td className="py-3 px-4">
                        <span className="capitalize text-sm">{transaction.type}</span>
                      </td>
                      <td className="py-3 px-4 text-sm">{transaction.merchant}</td>
                      <td className="py-3 px-4 text-sm font-medium">
                        {transaction.currency} {transaction.amount.toFixed(2)}
                      </td>
                      <td className="py-3 px-4">
                        <span className={`inline-flex items-center px-2 py-1 rounded-full text-xs font-medium ${
                          transaction.status === 'completed' ? 'bg-green-500/20 text-green-500' :
                          transaction.status === 'pending' ? 'bg-yellow-500/20 text-yellow-500' :
                          'bg-red-500/20 text-red-500'
                        }`}>
                          {transaction.status === 'completed' && <CheckCircle className="w-3 h-3 mr-1" />}
                          {transaction.status === 'pending' && <Clock className="w-3 h-3 mr-1" />}
                          {transaction.status === 'failed' && <XCircle className="w-3 h-3 mr-1" />}
                          {transaction.status}
                        </span>
                      </td>
                      <td className="py-3 px-4 text-sm text-muted-foreground">
                        {format(transaction.timestamp, 'HH:mm')}
                      </td>
                    </tr>
                  ))}
                </tbody>
              </table>
            </div>
          </Card>

          {/* Weekly Transactions Chart */}
          <Card variant="glass" className="p-6">
            <h3 className="text-lg font-semibold mb-4">Weekly Transactions</h3>
            <div className="h-64">
              <Bar data={chartData.transactions} options={chartOptions} />
            </div>
          </Card>
        </div>
      </div>

      {/* Mobile sidebar overlay */}
      {sidebarOpen && (
        <div
          className="fixed inset-0 bg-black/50 z-30 lg:hidden"
          onClick={() => setSidebarOpen(false)}
        />
      )}
    </div>
  )
}
