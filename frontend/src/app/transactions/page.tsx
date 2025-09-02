'use client'

import { useState, useEffect } from 'react'
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import {
  ArrowUpRight, ArrowDownLeft, Filter, Download, Search,
  Calendar, DollarSign, RefreshCw, CheckCircle, XCircle,
  Clock, TrendingUp, TrendingDown, MoreVertical
} from 'lucide-react'
import { motion, AnimatePresence } from 'framer-motion'
import { format } from 'date-fns'
import { TransactionService } from '@/services/api/services'
import { useQuery } from '@tanstack/react-query'

interface Transaction {
  id: string
  type: 'credit' | 'debit'
  amount: number
  currency: string
  description: string
  status: 'completed' | 'pending' | 'failed' | 'processing'
  category: string
  merchant?: string
  reference: string
  fee?: number
  createdAt: Date
  completedAt?: Date
}

export default function TransactionsPage() {
  const [searchTerm, setSearchTerm] = useState('')
  const [filterType, setFilterType] = useState<'all' | 'credit' | 'debit'>('all')
  const [filterStatus, setFilterStatus] = useState<'all' | 'completed' | 'pending' | 'failed'>('all')
  const [dateRange, setDateRange] = useState<'7d' | '30d' | '90d' | 'all'>('30d')
  const [showFilters, setShowFilters] = useState(false)
  
  const transactionService = new TransactionService()

  // Mock data for demonstration
  const mockTransactions: Transaction[] = [
    {
      id: 'TRX001',
      type: 'credit',
      amount: 5000,
      currency: 'USD',
      description: 'Payment from John Doe',
      status: 'completed',
      category: 'Payment',
      merchant: 'John Doe',
      reference: 'PAY-2024-001',
      fee: 25,
      createdAt: new Date(Date.now() - 3600000),
      completedAt: new Date(),
    },
    {
      id: 'TRX002',
      type: 'debit',
      amount: 1250,
      currency: 'USD',
      description: 'Transfer to Savings',
      status: 'completed',
      category: 'Transfer',
      reference: 'TRF-2024-002',
      fee: 0,
      createdAt: new Date(Date.now() - 7200000),
      completedAt: new Date(Date.now() - 7000000),
    },
    {
      id: 'TRX003',
      type: 'debit',
      amount: 89.99,
      currency: 'USD',
      description: 'AWS Services',
      status: 'processing',
      category: 'Subscription',
      merchant: 'Amazon Web Services',
      reference: 'SUB-2024-003',
      createdAt: new Date(Date.now() - 86400000),
    },
    {
      id: 'TRX004',
      type: 'credit',
      amount: 15000,
      currency: 'USD',
      description: 'Invoice Payment - Project X',
      status: 'completed',
      category: 'Invoice',
      merchant: 'Acme Corp',
      reference: 'INV-2024-004',
      fee: 150,
      createdAt: new Date(Date.now() - 172800000),
      completedAt: new Date(Date.now() - 172000000),
    },
    {
      id: 'TRX005',
      type: 'debit',
      amount: 500,
      currency: 'USD',
      description: 'Failed payment attempt',
      status: 'failed',
      category: 'Payment',
      reference: 'PAY-2024-005',
      createdAt: new Date(Date.now() - 259200000),
    },
  ]

  const { data: transactions = mockTransactions, isLoading, refetch } = useQuery({
    queryKey: ['transactions', filterType, filterStatus, dateRange],
    queryFn: async () => {
      // In production, this would call the API
      // const response = await transactionService.getTransactions({ type: filterType, status: filterStatus, range: dateRange })
      // return response.data
      return mockTransactions
    },
    refetchInterval: 30000, // Refetch every 30 seconds
  })

  const filteredTransactions = transactions.filter(transaction => {
    const matchesSearch = transaction.description.toLowerCase().includes(searchTerm.toLowerCase()) ||
                         transaction.reference.toLowerCase().includes(searchTerm.toLowerCase()) ||
                         transaction.merchant?.toLowerCase().includes(searchTerm.toLowerCase())
    const matchesType = filterType === 'all' || transaction.type === filterType
    const matchesStatus = filterStatus === 'all' || transaction.status === filterStatus
    return matchesSearch && matchesType && matchesStatus
  })

  const getStatusColor = (status: string) => {
    switch (status) {
      case 'completed': return 'text-green-500 bg-green-500/10'
      case 'pending': return 'text-yellow-500 bg-yellow-500/10'
      case 'processing': return 'text-blue-500 bg-blue-500/10'
      case 'failed': return 'text-red-500 bg-red-500/10'
      default: return 'text-gray-500 bg-gray-500/10'
    }
  }

  const getStatusIcon = (status: string) => {
    switch (status) {
      case 'completed': return <CheckCircle className="w-3 h-3" />
      case 'pending': return <Clock className="w-3 h-3" />
      case 'processing': return <RefreshCw className="w-3 h-3 animate-spin" />
      case 'failed': return <XCircle className="w-3 h-3" />
      default: return null
    }
  }

  const calculateTotals = () => {
    const credits = filteredTransactions
      .filter(t => t.type === 'credit' && t.status === 'completed')
      .reduce((sum, t) => sum + t.amount, 0)
    const debits = filteredTransactions
      .filter(t => t.type === 'debit' && t.status === 'completed')
      .reduce((sum, t) => sum + t.amount, 0)
    return { credits, debits, net: credits - debits }
  }

  const totals = calculateTotals()

  const exportTransactions = () => {
    const csv = [
      ['ID', 'Date', 'Type', 'Amount', 'Currency', 'Description', 'Status', 'Reference'],
      ...filteredTransactions.map(t => [
        t.id,
        format(t.createdAt, 'yyyy-MM-dd HH:mm:ss'),
        t.type,
        t.amount.toString(),
        t.currency,
        t.description,
        t.status,
        t.reference,
      ])
    ].map(row => row.join(',')).join('\n')

    const blob = new Blob([csv], { type: 'text/csv' })
    const url = window.URL.createObjectURL(blob)
    const a = document.createElement('a')
    a.href = url
    a.download = `transactions-${format(new Date(), 'yyyy-MM-dd')}.csv`
    a.click()
  }

  return (
    <div className="min-h-screen bg-background p-6">
      <div className="max-w-7xl mx-auto">
        {/* Header */}
        <div className="flex justify-between items-start mb-8">
          <div>
            <h1 className="text-3xl font-bold mb-2">Transaction History</h1>
            <p className="text-muted-foreground">View and manage all your transactions</p>
          </div>
          <div className="flex space-x-2">
            <Button variant="outline" onClick={() => refetch()}>
              <RefreshCw className="w-4 h-4 mr-2" />
              Refresh
            </Button>
            <Button variant="gradient" onClick={exportTransactions}>
              <Download className="w-4 h-4 mr-2" />
              Export
            </Button>
          </div>
        </div>

        {/* Stats Cards */}
        <div className="grid grid-cols-1 md:grid-cols-4 gap-4 mb-8">
          <motion.div
            initial={{ opacity: 0, y: 20 }}
            animate={{ opacity: 1, y: 0 }}
          >
            <Card variant="futuristic" className="p-4">
              <div className="flex items-center justify-between">
                <div>
                  <p className="text-sm text-muted-foreground">Total Received</p>
                  <p className="text-2xl font-bold text-green-500">
                    ${totals.credits.toFixed(2)}
                  </p>
                </div>
                <div className="w-10 h-10 rounded-lg bg-green-500/10 flex items-center justify-center">
                  <ArrowDownLeft className="w-5 h-5 text-green-500" />
                </div>
              </div>
            </Card>
          </motion.div>

          <motion.div
            initial={{ opacity: 0, y: 20 }}
            animate={{ opacity: 1, y: 0 }}
            transition={{ delay: 0.1 }}
          >
            <Card variant="futuristic" className="p-4">
              <div className="flex items-center justify-between">
                <div>
                  <p className="text-sm text-muted-foreground">Total Sent</p>
                  <p className="text-2xl font-bold text-red-500">
                    ${totals.debits.toFixed(2)}
                  </p>
                </div>
                <div className="w-10 h-10 rounded-lg bg-red-500/10 flex items-center justify-center">
                  <ArrowUpRight className="w-5 h-5 text-red-500" />
                </div>
              </div>
            </Card>
          </motion.div>

          <motion.div
            initial={{ opacity: 0, y: 20 }}
            animate={{ opacity: 1, y: 0 }}
            transition={{ delay: 0.2 }}
          >
            <Card variant="futuristic" className="p-4">
              <div className="flex items-center justify-between">
                <div>
                  <p className="text-sm text-muted-foreground">Net Balance</p>
                  <p className={`text-2xl font-bold ${totals.net >= 0 ? 'text-green-500' : 'text-red-500'}`}>
                    ${Math.abs(totals.net).toFixed(2)}
                  </p>
                </div>
                <div className={`w-10 h-10 rounded-lg ${totals.net >= 0 ? 'bg-green-500/10' : 'bg-red-500/10'} flex items-center justify-center`}>
                  {totals.net >= 0 ? (
                    <TrendingUp className="w-5 h-5 text-green-500" />
                  ) : (
                    <TrendingDown className="w-5 h-5 text-red-500" />
                  )}
                </div>
              </div>
            </Card>
          </motion.div>

          <motion.div
            initial={{ opacity: 0, y: 20 }}
            animate={{ opacity: 1, y: 0 }}
            transition={{ delay: 0.3 }}
          >
            <Card variant="futuristic" className="p-4">
              <div className="flex items-center justify-between">
                <div>
                  <p className="text-sm text-muted-foreground">Transactions</p>
                  <p className="text-2xl font-bold">{filteredTransactions.length}</p>
                </div>
                <div className="w-10 h-10 rounded-lg bg-primary/10 flex items-center justify-center">
                  <DollarSign className="w-5 h-5 text-primary" />
                </div>
              </div>
            </Card>
          </motion.div>
        </div>

        {/* Filters */}
        <Card variant="glass" className="mb-6">
          <CardContent className="p-4">
            <div className="flex flex-col lg:flex-row gap-4">
              {/* Search */}
              <div className="flex-1">
                <div className="relative">
                  <Search className="absolute left-3 top-1/2 -translate-y-1/2 w-4 h-4 text-muted-foreground" />
                  <Input
                    type="text"
                    placeholder="Search transactions..."
                    value={searchTerm}
                    onChange={(e) => setSearchTerm(e.target.value)}
                    className="pl-10"
                    variant="futuristic"
                  />
                </div>
              </div>

              {/* Filter Buttons */}
              <div className="flex space-x-2">
                <Button
                  variant={showFilters ? 'gradient' : 'outline'}
                  onClick={() => setShowFilters(!showFilters)}
                >
                  <Filter className="w-4 h-4 mr-2" />
                  Filters
                </Button>

                <select
                  value={dateRange}
                  onChange={(e) => setDateRange(e.target.value as any)}
                  className="px-4 py-2 rounded-lg border bg-background"
                >
                  <option value="7d">Last 7 days</option>
                  <option value="30d">Last 30 days</option>
                  <option value="90d">Last 90 days</option>
                  <option value="all">All time</option>
                </select>
              </div>
            </div>

            {/* Advanced Filters */}
            <AnimatePresence>
              {showFilters && (
                <motion.div
                  initial={{ height: 0, opacity: 0 }}
                  animate={{ height: 'auto', opacity: 1 }}
                  exit={{ height: 0, opacity: 0 }}
                  className="overflow-hidden"
                >
                  <div className="flex space-x-4 mt-4 pt-4 border-t">
                    <div className="flex space-x-2">
                      <Button
                        size="sm"
                        variant={filterType === 'all' ? 'gradient' : 'outline'}
                        onClick={() => setFilterType('all')}
                      >
                        All Types
                      </Button>
                      <Button
                        size="sm"
                        variant={filterType === 'credit' ? 'gradient' : 'outline'}
                        onClick={() => setFilterType('credit')}
                      >
                        Credits
                      </Button>
                      <Button
                        size="sm"
                        variant={filterType === 'debit' ? 'gradient' : 'outline'}
                        onClick={() => setFilterType('debit')}
                      >
                        Debits
                      </Button>
                    </div>

                    <div className="flex space-x-2">
                      <Button
                        size="sm"
                        variant={filterStatus === 'all' ? 'gradient' : 'outline'}
                        onClick={() => setFilterStatus('all')}
                      >
                        All Status
                      </Button>
                      <Button
                        size="sm"
                        variant={filterStatus === 'completed' ? 'gradient' : 'outline'}
                        onClick={() => setFilterStatus('completed')}
                      >
                        Completed
                      </Button>
                      <Button
                        size="sm"
                        variant={filterStatus === 'pending' ? 'gradient' : 'outline'}
                        onClick={() => setFilterStatus('pending')}
                      >
                        Pending
                      </Button>
                      <Button
                        size="sm"
                        variant={filterStatus === 'failed' ? 'gradient' : 'outline'}
                        onClick={() => setFilterStatus('failed')}
                      >
                        Failed
                      </Button>
                    </div>
                  </div>
                </motion.div>
              )}
            </AnimatePresence>
          </CardContent>
        </Card>

        {/* Transactions List */}
        <Card variant="glass">
          <CardHeader>
            <CardTitle>Transactions</CardTitle>
            <CardDescription>
              {filteredTransactions.length} transaction{filteredTransactions.length !== 1 ? 's' : ''} found
            </CardDescription>
          </CardHeader>
          <CardContent>
            {isLoading ? (
              <div className="flex justify-center py-8">
                <RefreshCw className="w-6 h-6 animate-spin text-primary" />
              </div>
            ) : filteredTransactions.length === 0 ? (
              <div className="text-center py-8 text-muted-foreground">
                No transactions found
              </div>
            ) : (
              <div className="space-y-2">
                {filteredTransactions.map((transaction, index) => (
                  <motion.div
                    key={transaction.id}
                    initial={{ opacity: 0, x: -20 }}
                    animate={{ opacity: 1, x: 0 }}
                    transition={{ delay: index * 0.05 }}
                  >
                    <Card variant="futuristic" className="p-4 hover:shadow-lg transition-shadow">
                      <div className="flex items-center justify-between">
                        <div className="flex items-center space-x-4">
                          <div className={`w-10 h-10 rounded-lg flex items-center justify-center ${
                            transaction.type === 'credit' ? 'bg-green-500/10' : 'bg-red-500/10'
                          }`}>
                            {transaction.type === 'credit' ? (
                              <ArrowDownLeft className="w-5 h-5 text-green-500" />
                            ) : (
                              <ArrowUpRight className="w-5 h-5 text-red-500" />
                            )}
                          </div>
                          <div>
                            <p className="font-medium">{transaction.description}</p>
                            <div className="flex items-center space-x-3 mt-1">
                              <span className="text-xs text-muted-foreground">
                                {format(transaction.createdAt, 'MMM dd, yyyy HH:mm')}
                              </span>
                              <span className="text-xs text-muted-foreground">
                                {transaction.reference}
                              </span>
                              {transaction.merchant && (
                                <span className="text-xs text-muted-foreground">
                                  {transaction.merchant}
                                </span>
                              )}
                            </div>
                          </div>
                        </div>

                        <div className="flex items-center space-x-4">
                          <div className="text-right">
                            <p className={`font-bold text-lg ${
                              transaction.type === 'credit' ? 'text-green-500' : 'text-red-500'
                            }`}>
                              {transaction.type === 'credit' ? '+' : '-'}
                              {transaction.currency} {transaction.amount.toFixed(2)}
                            </p>
                            {transaction.fee && transaction.fee > 0 && (
                              <p className="text-xs text-muted-foreground">
                                Fee: {transaction.currency} {transaction.fee.toFixed(2)}
                              </p>
                            )}
                          </div>

                          <span className={`inline-flex items-center px-2 py-1 rounded-full text-xs font-medium ${getStatusColor(transaction.status)}`}>
                            {getStatusIcon(transaction.status)}
                            <span className="ml-1 capitalize">{transaction.status}</span>
                          </span>

                          <Button variant="ghost" size="sm">
                            <MoreVertical className="w-4 h-4" />
                          </Button>
                        </div>
                      </div>
                    </Card>
                  </motion.div>
                ))}
              </div>
            )}
          </CardContent>
        </Card>
      </div>
    </div>
  )
}
