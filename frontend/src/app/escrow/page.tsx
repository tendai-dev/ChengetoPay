'use client'

import { useState } from 'react'
import { useRouter } from 'next/navigation'
import { useForm } from 'react-hook-form'
import { zodResolver } from '@hookform/resolvers/zod'
import * as z from 'zod'
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { 
  Shield, Lock, Users, DollarSign, Clock, CheckCircle,
  AlertTriangle, FileText, ArrowRight, TrendingUp, Package
} from 'lucide-react'
import { toast } from 'react-hot-toast'
import { motion, AnimatePresence } from 'framer-motion'
import { EscrowService } from '@/services/api/services'
import { format } from 'date-fns'

const escrowSchema = z.object({
  amount: z.string().regex(/^\d+(\.\d{1,2})?$/, 'Invalid amount format'),
  currency: z.string().min(3, 'Select a currency'),
  buyerEmail: z.string().email('Invalid buyer email'),
  sellerEmail: z.string().email('Invalid seller email'),
  description: z.string().min(10, 'Description must be at least 10 characters'),
  releaseConditions: z.string().min(20, 'Release conditions must be at least 20 characters'),
  duration: z.string().regex(/^\d+$/, 'Duration must be a number'),
})

type EscrowFormData = z.infer<typeof escrowSchema>

interface EscrowTransaction {
  id: string
  amount: number
  currency: string
  buyer: string
  seller: string
  status: 'pending' | 'funded' | 'released' | 'disputed' | 'cancelled'
  createdAt: Date
  releaseDate?: Date
}

export default function EscrowPage() {
  const router = useRouter()
  const [isCreating, setIsCreating] = useState(false)
  const [activeTab, setActiveTab] = useState<'create' | 'active' | 'history'>('create')
  const [escrowTransactions, setEscrowTransactions] = useState<EscrowTransaction[]>([
    {
      id: '1',
      amount: 5000,
      currency: 'USD',
      buyer: 'john@example.com',
      seller: 'seller@example.com',
      status: 'funded',
      createdAt: new Date(Date.now() - 86400000),
    },
    {
      id: '2',
      amount: 2500,
      currency: 'EUR',
      buyer: 'alice@example.com',
      seller: 'bob@example.com',
      status: 'pending',
      createdAt: new Date(Date.now() - 172800000),
    },
    {
      id: '3',
      amount: 10000,
      currency: 'USD',
      buyer: 'company@example.com',
      seller: 'vendor@example.com',
      status: 'released',
      createdAt: new Date(Date.now() - 259200000),
      releaseDate: new Date(),
    },
  ])

  const escrowService = new EscrowService()

  const {
    register,
    handleSubmit,
    formState: { errors },
    watch,
  } = useForm<EscrowFormData>({
    resolver: zodResolver(escrowSchema),
    defaultValues: {
      currency: 'USD',
      duration: '30',
    },
  })

  const amount = watch('amount')
  const currency = watch('currency')
  const duration = watch('duration')

  const onSubmit = async (data: EscrowFormData) => {
    setIsCreating(true)
    try {
      const response = await escrowService.createEscrow({
        amount: parseFloat(data.amount),
        currency: data.currency,
        buyer_email: data.buyerEmail,
        seller_email: data.sellerEmail,
        description: data.description,
        release_conditions: data.releaseConditions,
        duration_days: parseInt(data.duration),
      })
      
      toast.success('Escrow created successfully!')
      setActiveTab('active')
      
      // Add to local state
      setEscrowTransactions(prev => [{
        id: (response.data as any)?.id || Date.now().toString(),
        amount: parseFloat(data.amount),
        currency: data.currency,
        buyer: data.buyerEmail,
        seller: data.sellerEmail,
        status: 'pending',
        createdAt: new Date(),
      }, ...prev])
    } catch (error: any) {
      toast.error(error.response?.data?.message || 'Failed to create escrow')
    } finally {
      setIsCreating(false)
    }
  }

  const getStatusColor = (status: string) => {
    switch (status) {
      case 'funded': return 'text-green-500 bg-green-500/10'
      case 'pending': return 'text-yellow-500 bg-yellow-500/10'
      case 'released': return 'text-blue-500 bg-blue-500/10'
      case 'disputed': return 'text-red-500 bg-red-500/10'
      case 'cancelled': return 'text-gray-500 bg-gray-500/10'
      default: return 'text-gray-500 bg-gray-500/10'
    }
  }

  const getStatusIcon = (status: string) => {
    switch (status) {
      case 'funded': return <CheckCircle className="w-4 h-4" />
      case 'pending': return <Clock className="w-4 h-4" />
      case 'released': return <Package className="w-4 h-4" />
      case 'disputed': return <AlertTriangle className="w-4 h-4" />
      default: return <Lock className="w-4 h-4" />
    }
  }

  const calculateFee = (amount: string) => {
    const value = parseFloat(amount) || 0
    return (value * 0.02).toFixed(2) // 2% escrow fee
  }

  return (
    <div className="min-h-screen bg-background p-6">
      <div className="max-w-7xl mx-auto">
        {/* Header */}
        <div className="mb-8">
          <h1 className="text-3xl font-bold mb-2">Escrow Service</h1>
          <p className="text-muted-foreground">Secure transactions with built-in protection</p>
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
                  <p className="text-sm text-muted-foreground">Total in Escrow</p>
                  <p className="text-2xl font-bold">$125,420</p>
                </div>
                <div className="w-12 h-12 rounded-xl bg-gradient-to-br from-blue-500 to-cyan-500 flex items-center justify-center text-white">
                  <Shield className="w-6 h-6" />
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
                  <p className="text-sm text-muted-foreground">Active Escrows</p>
                  <p className="text-2xl font-bold">24</p>
                </div>
                <div className="w-12 h-12 rounded-xl bg-gradient-to-br from-purple-500 to-pink-500 flex items-center justify-center text-white">
                  <Lock className="w-6 h-6" />
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
                  <p className="text-sm text-muted-foreground">Released Today</p>
                  <p className="text-2xl font-bold">$15,230</p>
                </div>
                <div className="w-12 h-12 rounded-xl bg-gradient-to-br from-green-500 to-emerald-500 flex items-center justify-center text-white">
                  <TrendingUp className="w-6 h-6" />
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
                  <p className="text-sm text-muted-foreground">Disputes</p>
                  <p className="text-2xl font-bold">2</p>
                </div>
                <div className="w-12 h-12 rounded-xl bg-gradient-to-br from-orange-500 to-red-500 flex items-center justify-center text-white">
                  <AlertTriangle className="w-6 h-6" />
                </div>
              </div>
            </Card>
          </motion.div>
        </div>

        {/* Tabs */}
        <div className="flex space-x-4 mb-6">
          {['create', 'active', 'history'].map((tab) => (
            <Button
              key={tab}
              variant={activeTab === tab ? 'gradient' : 'outline'}
              onClick={() => setActiveTab(tab as any)}
              className="capitalize"
            >
              {tab === 'create' && <Shield className="w-4 h-4 mr-2" />}
              {tab === 'active' && <Lock className="w-4 h-4 mr-2" />}
              {tab === 'history' && <FileText className="w-4 h-4 mr-2" />}
              {tab} {tab === 'active' && '(2)'}
            </Button>
          ))}
        </div>

        <AnimatePresence mode="wait">
          {/* Create Escrow Tab */}
          {activeTab === 'create' && (
            <motion.div
              key="create"
              initial={{ opacity: 0, x: 20 }}
              animate={{ opacity: 1, x: 0 }}
              exit={{ opacity: 0, x: -20 }}
              className="grid lg:grid-cols-3 gap-6"
            >
              <div className="lg:col-span-2">
                <Card variant="glass">
                  <CardHeader>
                    <CardTitle>Create New Escrow</CardTitle>
                    <CardDescription>Set up a secure escrow transaction</CardDescription>
                  </CardHeader>
                  <CardContent>
                    <form onSubmit={handleSubmit(onSubmit)} className="space-y-6">
                      <div className="grid grid-cols-2 gap-4">
                        <div>
                          <label className="text-sm font-medium mb-2 block">Amount</label>
                          <Input
                            {...register('amount')}
                            type="text"
                            placeholder="0.00"
                            variant="futuristic"
                            icon={<DollarSign className="w-4 h-4" />}
                            error={errors.amount?.message}
                          />
                        </div>
                        <div>
                          <label className="text-sm font-medium mb-2 block">Currency</label>
                          <select
                            {...register('currency')}
                            className="w-full px-4 py-2 rounded-lg border bg-background"
                          >
                            <option value="USD">USD</option>
                            <option value="EUR">EUR</option>
                            <option value="GBP">GBP</option>
                          </select>
                        </div>
                      </div>

                      <div className="grid grid-cols-2 gap-4">
                        <div>
                          <label className="text-sm font-medium mb-2 block">Buyer Email</label>
                          <Input
                            {...register('buyerEmail')}
                            type="email"
                            placeholder="buyer@example.com"
                            variant="futuristic"
                            icon={<Users className="w-4 h-4" />}
                            error={errors.buyerEmail?.message}
                          />
                        </div>
                        <div>
                          <label className="text-sm font-medium mb-2 block">Seller Email</label>
                          <Input
                            {...register('sellerEmail')}
                            type="email"
                            placeholder="seller@example.com"
                            variant="futuristic"
                            icon={<Users className="w-4 h-4" />}
                            error={errors.sellerEmail?.message}
                          />
                        </div>
                      </div>

                      <div>
                        <label className="text-sm font-medium mb-2 block">Description</label>
                        <textarea
                          {...register('description')}
                          className="w-full px-4 py-2 rounded-lg border bg-background min-h-[100px]"
                          placeholder="Describe the transaction..."
                        />
                        {errors.description && (
                          <p className="text-xs text-destructive mt-1">{errors.description.message}</p>
                        )}
                      </div>

                      <div>
                        <label className="text-sm font-medium mb-2 block">Release Conditions</label>
                        <textarea
                          {...register('releaseConditions')}
                          className="w-full px-4 py-2 rounded-lg border bg-background min-h-[100px]"
                          placeholder="Specify the conditions for releasing funds..."
                        />
                        {errors.releaseConditions && (
                          <p className="text-xs text-destructive mt-1">{errors.releaseConditions.message}</p>
                        )}
                      </div>

                      <div>
                        <label className="text-sm font-medium mb-2 block">Duration (days)</label>
                        <Input
                          {...register('duration')}
                          type="text"
                          placeholder="30"
                          variant="futuristic"
                          icon={<Clock className="w-4 h-4" />}
                          error={errors.duration?.message}
                        />
                      </div>

                      <Button
                        type="submit"
                        className="w-full"
                        variant="gradient"
                        size="lg"
                        loading={isCreating}
                      >
                        {isCreating ? 'Creating Escrow...' : 'Create Escrow'}
                        <Shield className="ml-2 w-4 h-4" />
                      </Button>
                    </form>
                  </CardContent>
                </Card>
              </div>

              {/* Summary */}
              <div className="space-y-6">
                <Card variant="futuristic">
                  <CardHeader>
                    <CardTitle>Escrow Summary</CardTitle>
                  </CardHeader>
                  <CardContent className="space-y-4">
                    <div className="flex justify-between">
                      <span className="text-muted-foreground">Amount</span>
                      <span className="font-medium">
                        {currency} {amount || '0.00'}
                      </span>
                    </div>
                    <div className="flex justify-between">
                      <span className="text-muted-foreground">Escrow Fee (2%)</span>
                      <span className="font-medium">
                        {currency} {amount ? calculateFee(amount) : '0.00'}
                      </span>
                    </div>
                    <div className="flex justify-between">
                      <span className="text-muted-foreground">Duration</span>
                      <span className="font-medium">{duration || '30'} days</span>
                    </div>
                    <div className="border-t pt-4">
                      <div className="flex justify-between">
                        <span className="font-medium">Total Locked</span>
                        <span className="text-xl font-bold gradient-text">
                          {currency} {amount || '0.00'}
                        </span>
                      </div>
                    </div>
                  </CardContent>
                </Card>

                <Card variant="glass" className="border-blue-500/50">
                  <CardContent className="pt-6">
                    <div className="flex items-start space-x-3">
                      <Shield className="w-5 h-5 text-blue-500 mt-0.5" />
                      <div>
                        <p className="font-medium">How Escrow Works</p>
                        <ul className="text-sm text-muted-foreground mt-2 space-y-1">
                          <li>• Funds are securely held by ChengetoPay</li>
                          <li>• Released only when conditions are met</li>
                          <li>• Dispute resolution available</li>
                          <li>• Full transaction history tracked</li>
                        </ul>
                      </div>
                    </div>
                  </CardContent>
                </Card>
              </div>
            </motion.div>
          )}

          {/* Active Escrows Tab */}
          {activeTab === 'active' && (
            <motion.div
              key="active"
              initial={{ opacity: 0, x: 20 }}
              animate={{ opacity: 1, x: 0 }}
              exit={{ opacity: 0, x: -20 }}
            >
              <Card variant="glass">
                <CardHeader>
                  <CardTitle>Active Escrows</CardTitle>
                  <CardDescription>Manage your ongoing escrow transactions</CardDescription>
                </CardHeader>
                <CardContent>
                  <div className="space-y-4">
                    {escrowTransactions
                      .filter(t => ['pending', 'funded'].includes(t.status))
                      .map((transaction) => (
                        <Card key={transaction.id} variant="futuristic" className="p-4">
                          <div className="flex items-start justify-between">
                            <div className="flex-1">
                              <div className="flex items-center space-x-3 mb-2">
                                <span className={`inline-flex items-center px-2 py-1 rounded-full text-xs font-medium ${getStatusColor(transaction.status)}`}>
                                  {getStatusIcon(transaction.status)}
                                  <span className="ml-1 capitalize">{transaction.status}</span>
                                </span>
                                <span className="text-sm text-muted-foreground">
                                  ID: {transaction.id}
                                </span>
                              </div>
                              <p className="font-medium">
                                {transaction.currency} {transaction.amount.toFixed(2)}
                              </p>
                              <div className="flex items-center space-x-4 mt-2 text-sm text-muted-foreground">
                                <span>Buyer: {transaction.buyer}</span>
                                <span>•</span>
                                <span>Seller: {transaction.seller}</span>
                              </div>
                              <p className="text-sm text-muted-foreground mt-1">
                                Created: {format(transaction.createdAt, 'MMM dd, yyyy')}
                              </p>
                            </div>
                            <div className="flex space-x-2">
                              {transaction.status === 'funded' && (
                                <Button size="sm" variant="gradient">
                                  Release
                                </Button>
                              )}
                              <Button size="sm" variant="outline">
                                View
                              </Button>
                            </div>
                          </div>
                        </Card>
                      ))}
                  </div>
                </CardContent>
              </Card>
            </motion.div>
          )}

          {/* History Tab */}
          {activeTab === 'history' && (
            <motion.div
              key="history"
              initial={{ opacity: 0, x: 20 }}
              animate={{ opacity: 1, x: 0 }}
              exit={{ opacity: 0, x: -20 }}
            >
              <Card variant="glass">
                <CardHeader>
                  <CardTitle>Transaction History</CardTitle>
                  <CardDescription>View all your past escrow transactions</CardDescription>
                </CardHeader>
                <CardContent>
                  <div className="overflow-x-auto">
                    <table className="w-full">
                      <thead>
                        <tr className="border-b border-border/50">
                          <th className="text-left py-3 px-4 text-sm font-medium text-muted-foreground">ID</th>
                          <th className="text-left py-3 px-4 text-sm font-medium text-muted-foreground">Amount</th>
                          <th className="text-left py-3 px-4 text-sm font-medium text-muted-foreground">Parties</th>
                          <th className="text-left py-3 px-4 text-sm font-medium text-muted-foreground">Status</th>
                          <th className="text-left py-3 px-4 text-sm font-medium text-muted-foreground">Date</th>
                          <th className="text-left py-3 px-4 text-sm font-medium text-muted-foreground">Action</th>
                        </tr>
                      </thead>
                      <tbody>
                        {escrowTransactions.map((transaction) => (
                          <tr key={transaction.id} className="border-b border-border/30 hover:bg-muted/20 transition-colors">
                            <td className="py-3 px-4 text-sm">#{transaction.id}</td>
                            <td className="py-3 px-4 text-sm font-medium">
                              {transaction.currency} {transaction.amount.toFixed(2)}
                            </td>
                            <td className="py-3 px-4 text-sm">
                              <div>
                                <p className="text-xs">B: {transaction.buyer}</p>
                                <p className="text-xs text-muted-foreground">S: {transaction.seller}</p>
                              </div>
                            </td>
                            <td className="py-3 px-4">
                              <span className={`inline-flex items-center px-2 py-1 rounded-full text-xs font-medium ${getStatusColor(transaction.status)}`}>
                                {getStatusIcon(transaction.status)}
                                <span className="ml-1 capitalize">{transaction.status}</span>
                              </span>
                            </td>
                            <td className="py-3 px-4 text-sm text-muted-foreground">
                              {format(transaction.createdAt, 'MMM dd, yyyy')}
                            </td>
                            <td className="py-3 px-4">
                              <Button size="sm" variant="ghost">
                                View
                              </Button>
                            </td>
                          </tr>
                        ))}
                      </tbody>
                    </table>
                  </div>
                </CardContent>
              </Card>
            </motion.div>
          )}
        </AnimatePresence>
      </div>
    </div>
  )
}
