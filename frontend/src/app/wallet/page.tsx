'use client'

import { useState } from 'react'
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import {
  Wallet, Plus, ArrowUpRight, ArrowDownLeft, CreditCard,
  Bitcoin, DollarSign, Euro, PoundSterling, TrendingUp,
  Shield, Bell, Settings, Copy, QrCode, Eye, EyeOff
} from 'lucide-react'
import { motion, AnimatePresence } from 'framer-motion'
import { toast } from 'react-hot-toast'
import { useForm } from 'react-hook-form'
import { zodResolver } from '@hookform/resolvers/zod'
import * as z from 'zod'
// Mock wallet service for now
const WalletService = class {
  constructor() {}
  
  async fundWallet(data: any) {
    return Promise.resolve({ success: true, data: { id: Date.now().toString() } })
  }
  
  async withdrawFromWallet(data: any) {
    return Promise.resolve({ success: true, data: { id: Date.now().toString() } })
  }
  
  async convertCurrency(data: any) {
    return Promise.resolve({ success: true, data: { id: Date.now().toString() } })
  }
}

const fundSchema = z.object({
  amount: z.string().regex(/^\d+(\.\d{1,2})?$/, 'Invalid amount format'),
  currency: z.string().min(3, 'Select a currency'),
  paymentMethod: z.enum(['card', 'bank', 'crypto']),
})

const withdrawSchema = z.object({
  amount: z.string().regex(/^\d+(\.\d{1,2})?$/, 'Invalid amount format'),
  currency: z.string().min(3, 'Select a currency'),
  destination: z.string().min(10, 'Enter destination details'),
})

type FundFormData = z.infer<typeof fundSchema>
type WithdrawFormData = z.infer<typeof withdrawSchema>

interface WalletBalance {
  currency: string
  symbol: string
  balance: number
  available: number
  pending: number
  icon: React.ReactNode
  change24h: number
}

export default function WalletPage() {
  const [activeTab, setActiveTab] = useState<'overview' | 'fund' | 'withdraw' | 'convert'>('overview')
  const [showBalance, setShowBalance] = useState(true)
  const [selectedCurrency, setSelectedCurrency] = useState('USD')
  const [isFunding, setIsFunding] = useState(false)
  const [isWithdrawing, setIsWithdrawing] = useState(false)
  
  const walletService = new WalletService()

  const walletBalances: WalletBalance[] = [
    {
      currency: 'USD',
      symbol: '$',
      balance: 12450.75,
      available: 11450.75,
      pending: 1000,
      icon: <DollarSign className="w-5 h-5" />,
      change24h: 2.5,
    },
    {
      currency: 'EUR',
      symbol: '€',
      balance: 8320.50,
      available: 8320.50,
      pending: 0,
      icon: <Euro className="w-5 h-5" />,
      change24h: -0.8,
    },
    {
      currency: 'GBP',
      symbol: '£',
      balance: 5200.00,
      available: 5200.00,
      pending: 0,
      icon: <PoundSterling className="w-5 h-5" />,
      change24h: 1.2,
    },
    {
      currency: 'BTC',
      symbol: '₿',
      balance: 0.45,
      available: 0.45,
      pending: 0,
      icon: <Bitcoin className="w-5 h-5" />,
      change24h: 5.3,
    },
  ]

  const totalBalanceUSD = walletBalances.reduce((sum, wallet) => {
    // Convert to USD (simplified conversion)
    const rate = wallet.currency === 'EUR' ? 1.1 : 
                 wallet.currency === 'GBP' ? 1.3 : 
                 wallet.currency === 'BTC' ? 45000 : 1
    return sum + (wallet.balance * rate)
  }, 0)

  const {
    register: registerFund,
    handleSubmit: handleFundSubmit,
    formState: { errors: fundErrors },
  } = useForm<FundFormData>({
    resolver: zodResolver(fundSchema),
    defaultValues: {
      currency: 'USD',
      paymentMethod: 'card',
    },
  })

  const {
    register: registerWithdraw,
    handleSubmit: handleWithdrawSubmit,
    formState: { errors: withdrawErrors },
  } = useForm<WithdrawFormData>({
    resolver: zodResolver(withdrawSchema),
    defaultValues: {
      currency: 'USD',
    },
  })

  const onFund = async (data: FundFormData) => {
    setIsFunding(true)
    try {
      await walletService.fundWallet({
        amount: parseFloat(data.amount),
        currency: data.currency,
        payment_method: data.paymentMethod,
      })
      toast.success('Wallet funded successfully!')
      setActiveTab('overview')
    } catch (error: any) {
      toast.error(error.response?.data?.message || 'Failed to fund wallet')
    } finally {
      setIsFunding(false)
    }
  }

  const onWithdraw = async (data: WithdrawFormData) => {
    setIsWithdrawing(true)
    try {
      await walletService.withdrawFromWallet({
        amount: parseFloat(data.amount),
        currency: data.currency,
        destination: data.destination,
      })
      toast.success('Withdrawal initiated!')
      setActiveTab('overview')
    } catch (error: any) {
      toast.error(error.response?.data?.message || 'Withdrawal failed')
    } finally {
      setIsWithdrawing(false)
    }
  }

  const copyWalletAddress = (currency: string) => {
    const address = `chengeto-${currency.toLowerCase()}-${Math.random().toString(36).substring(7)}`
    navigator.clipboard.writeText(address)
    toast.success('Wallet address copied!')
  }

  return (
    <div className="min-h-screen bg-background p-6">
      <div className="max-w-7xl mx-auto">
        {/* Header */}
        <div className="flex justify-between items-start mb-8">
          <div>
            <h1 className="text-3xl font-bold mb-2">Wallet</h1>
            <p className="text-muted-foreground">Manage your multi-currency digital wallet</p>
          </div>
          <div className="flex space-x-2">
            <Button variant="default" size="icon">
              <Bell className="w-4 h-4" />
            </Button>
            <Button variant="default" size="icon">
              <Settings className="w-4 h-4" />
            </Button>
          </div>
        </div>

        {/* Total Balance Card */}
        <Card variant="gradient" className="mb-8">
          <CardContent className="p-6">
            <div className="flex justify-between items-start">
              <div>
                <p className="text-sm text-white/70 mb-2">Total Balance</p>
                <div className="flex items-center space-x-3">
                  <h2 className="text-4xl font-bold text-white">
                    {showBalance ? `$${totalBalanceUSD.toFixed(2)}` : '••••••'}
                  </h2>
                  <Button
                    variant="ghost"
                    size="sm"
                    onClick={() => setShowBalance(!showBalance)}
                    className="text-white/70 hover:text-white"
                  >
                    {showBalance ? <EyeOff className="w-4 h-4" /> : <Eye className="w-4 h-4" />}
                  </Button>
                </div>
                <div className="flex items-center mt-2">
                  <TrendingUp className="w-4 h-4 text-green-400 mr-1" />
                  <span className="text-green-400 text-sm">+$1,234.56 (2.5%)</span>
                  <span className="text-white/50 text-sm ml-2">Today</span>
                </div>
              </div>
              <div className="flex space-x-2">
                <Button
                  variant="default"
                  className="text-white border-white/20"
                  onClick={() => setActiveTab('fund')}
                >
                  <Plus className="w-4 h-4 mr-2" />
                  Fund
                </Button>
                <Button
                  variant="default"
                  className="text-white border-white/20"
                  onClick={() => setActiveTab('withdraw')}
                >
                  <ArrowUpRight className="w-4 h-4 mr-2" />
                  Withdraw
                </Button>
              </div>
            </div>
          </CardContent>
        </Card>

        {/* Tabs */}
        <div className="flex space-x-4 mb-6">
          {['overview', 'fund', 'withdraw', 'convert'].map((tab) => (
            <Button
              key={tab}
              variant={activeTab === tab ? 'gradient' : 'outline'}
              onClick={() => setActiveTab(tab as any)}
              className="capitalize"
            >
              {tab === 'overview' && <Wallet className="w-4 h-4 mr-2" />}
              {tab === 'fund' && <ArrowDownLeft className="w-4 h-4 mr-2" />}
              {tab === 'withdraw' && <ArrowUpRight className="w-4 h-4 mr-2" />}
              {tab === 'convert' && <TrendingUp className="w-4 h-4 mr-2" />}
              {tab}
            </Button>
          ))}
        </div>

        <AnimatePresence mode="wait">
          {/* Overview Tab */}
          {activeTab === 'overview' && (
            <motion.div
              key="overview"
              initial={{ opacity: 0, y: 20 }}
              animate={{ opacity: 1, y: 0 }}
              exit={{ opacity: 0, y: -20 }}
            >
              <div className="grid lg:grid-cols-2 gap-6">
                {/* Currency Balances */}
                <Card variant="default">
                  <CardHeader>
                    <CardTitle>Currency Balances</CardTitle>
                    <CardDescription>Your multi-currency holdings</CardDescription>
                  </CardHeader>
                  <CardContent className="space-y-4">
                    {walletBalances.map((wallet) => (
                      <Card key={wallet.currency} variant="futuristic" className="p-4">
                        <div className="flex items-center justify-between">
                          <div className="flex items-center space-x-3">
                            <div className="w-10 h-10 rounded-lg bg-primary/10 flex items-center justify-center">
                              {wallet.icon}
                            </div>
                            <div>
                              <p className="font-medium">{wallet.currency}</p>
                              <p className="text-sm text-muted-foreground">
                                Available: {wallet.symbol}{wallet.available.toFixed(wallet.currency === 'BTC' ? 8 : 2)}
                              </p>
                            </div>
                          </div>
                          <div className="text-right">
                            <p className="font-bold">
                              {wallet.symbol}{wallet.balance.toFixed(wallet.currency === 'BTC' ? 8 : 2)}
                            </p>
                            <p className={`text-sm ${wallet.change24h >= 0 ? 'text-green-500' : 'text-red-500'}`}>
                              {wallet.change24h >= 0 ? '+' : ''}{wallet.change24h}%
                            </p>
                          </div>
                        </div>
                        {wallet.pending > 0 && (
                          <div className="mt-3 pt-3 border-t border-border/50">
                            <p className="text-xs text-yellow-500">
                              Pending: {wallet.symbol}{wallet.pending.toFixed(2)}
                            </p>
                          </div>
                        )}
                      </Card>
                    ))}
                  </CardContent>
                </Card>

                {/* Quick Actions & Recent Activity */}
                <div className="space-y-6">
                  {/* Quick Actions */}
                  <Card variant="default">
                    <CardHeader>
                      <CardTitle>Quick Actions</CardTitle>
                    </CardHeader>
                    <CardContent className="grid grid-cols-2 gap-3">
                      <Button variant="default" className="h-auto py-4 flex-col">
                        <QrCode className="w-5 h-5 mb-2" />
                        <span className="text-xs">Receive</span>
                      </Button>
                      <Button variant="default" className="h-auto py-4 flex-col">
                        <Copy className="w-5 h-5 mb-2" />
                        <span className="text-xs">Copy Address</span>
                      </Button>
                      <Button variant="default" className="h-auto py-4 flex-col">
                        <CreditCard className="w-5 h-5 mb-2" />
                        <span className="text-xs">Card Details</span>
                      </Button>
                      <Button variant="default" className="h-auto py-4 flex-col">
                        <Shield className="w-5 h-5 mb-2" />
                        <span className="text-xs">Security</span>
                      </Button>
                    </CardContent>
                  </Card>

                  {/* Recent Activity */}
                  <Card variant="default">
                    <CardHeader>
                      <CardTitle>Recent Activity</CardTitle>
                    </CardHeader>
                    <CardContent className="space-y-3">
                      {[
                        { type: 'in', amount: 500, currency: 'USD', desc: 'Received from John', time: '2h ago' },
                        { type: 'out', amount: 120, currency: 'EUR', desc: 'Sent to Alice', time: '5h ago' },
                        { type: 'in', amount: 0.01, currency: 'BTC', desc: 'Mining reward', time: '1d ago' },
                        { type: 'convert', amount: 1000, currency: 'USD', desc: 'USD → EUR', time: '2d ago' },
                      ].map((activity, index) => (
                        <div key={index} className="flex items-center justify-between">
                          <div className="flex items-center space-x-3">
                            <div className={`w-8 h-8 rounded-full flex items-center justify-center ${
                              activity.type === 'in' ? 'bg-green-500/10' :
                              activity.type === 'out' ? 'bg-red-500/10' :
                              'bg-blue-500/10'
                            }`}>
                              {activity.type === 'in' ? <ArrowDownLeft className="w-4 h-4 text-green-500" /> :
                               activity.type === 'out' ? <ArrowUpRight className="w-4 h-4 text-red-500" /> :
                               <TrendingUp className="w-4 h-4 text-blue-500" />}
                            </div>
                            <div>
                              <p className="text-sm font-medium">{activity.desc}</p>
                              <p className="text-xs text-muted-foreground">{activity.time}</p>
                            </div>
                          </div>
                          <span className={`text-sm font-medium ${
                            activity.type === 'in' ? 'text-green-500' : ''
                          }`}>
                            {activity.type === 'in' ? '+' : activity.type === 'out' ? '-' : ''}
                            {activity.amount} {activity.currency}
                          </span>
                        </div>
                      ))}
                    </CardContent>
                  </Card>
                </div>
              </div>
            </motion.div>
          )}

          {/* Fund Tab */}
          {activeTab === 'fund' && (
            <motion.div
              key="fund"
              initial={{ opacity: 0, y: 20 }}
              animate={{ opacity: 1, y: 0 }}
              exit={{ opacity: 0, y: -20 }}
            >
              <Card variant="default">
                <CardHeader>
                  <CardTitle>Fund Wallet</CardTitle>
                  <CardDescription>Add funds to your wallet</CardDescription>
                </CardHeader>
                <CardContent>
                  <form onSubmit={handleFundSubmit(onFund)} className="space-y-6 max-w-md">
                    <div>
                      <label className="text-sm font-medium mb-2 block">Amount</label>
                      <Input
                        {...registerFund('amount')}
                        type="text"
                        placeholder="0.00"
                        variant="futuristic"
                        icon={<DollarSign className="w-4 h-4" />}
                        error={fundErrors.amount?.message}
                      />
                    </div>

                    <div>
                      <label className="text-sm font-medium mb-2 block">Currency</label>
                      <select
                        {...registerFund('currency')}
                        className="w-full px-4 py-2 rounded-lg border bg-background"
                      >
                        <option value="USD">USD - US Dollar</option>
                        <option value="EUR">EUR - Euro</option>
                        <option value="GBP">GBP - British Pound</option>
                        <option value="BTC">BTC - Bitcoin</option>
                      </select>
                    </div>

                    <div>
                      <label className="text-sm font-medium mb-2 block">Payment Method</label>
                      <select
                        {...registerFund('paymentMethod')}
                        className="w-full px-4 py-2 rounded-lg border bg-background"
                      >
                        <option value="card">Credit/Debit Card</option>
                        <option value="bank">Bank Transfer</option>
                        <option value="crypto">Crypto Transfer</option>
                      </select>
                    </div>

                    <Button
                      type="submit"
                      className="w-full"
                      variant="gradient"
                      size="lg"
                      loading={isFunding}
                    >
                      {isFunding ? 'Processing...' : 'Fund Wallet'}
                      <ArrowDownLeft className="ml-2 w-4 h-4" />
                    </Button>
                  </form>
                </CardContent>
              </Card>
            </motion.div>
          )}

          {/* Withdraw Tab */}
          {activeTab === 'withdraw' && (
            <motion.div
              key="withdraw"
              initial={{ opacity: 0, y: 20 }}
              animate={{ opacity: 1, y: 0 }}
              exit={{ opacity: 0, y: -20 }}
            >
              <Card variant="default">
                <CardHeader>
                  <CardTitle>Withdraw Funds</CardTitle>
                  <CardDescription>Transfer funds from your wallet</CardDescription>
                </CardHeader>
                <CardContent>
                  <form onSubmit={handleWithdrawSubmit(onWithdraw)} className="space-y-6 max-w-md">
                    <div>
                      <label className="text-sm font-medium mb-2 block">Amount</label>
                      <Input
                        {...registerWithdraw('amount')}
                        type="text"
                        placeholder="0.00"
                        variant="futuristic"
                        icon={<DollarSign className="w-4 h-4" />}
                        error={withdrawErrors.amount?.message}
                      />
                    </div>

                    <div>
                      <label className="text-sm font-medium mb-2 block">Currency</label>
                      <select
                        {...registerWithdraw('currency')}
                        className="w-full px-4 py-2 rounded-lg border bg-background"
                      >
                        {walletBalances.map((wallet) => (
                          <option key={wallet.currency} value={wallet.currency}>
                            {wallet.currency} - Available: {wallet.symbol}{wallet.available.toFixed(2)}
                          </option>
                        ))}
                      </select>
                    </div>

                    <div>
                      <label className="text-sm font-medium mb-2 block">Destination</label>
                      <Input
                        {...registerWithdraw('destination')}
                        type="text"
                        placeholder="Bank account or wallet address"
                        variant="futuristic"
                        error={withdrawErrors.destination?.message}
                      />
                    </div>

                    <Button
                      type="submit"
                      className="w-full"
                      variant="gradient"
                      size="lg"
                      loading={isWithdrawing}
                    >
                      {isWithdrawing ? 'Processing...' : 'Withdraw'}
                      <ArrowUpRight className="ml-2 w-4 h-4" />
                    </Button>
                  </form>
                </CardContent>
              </Card>
            </motion.div>
          )}

          {/* Convert Tab */}
          {activeTab === 'convert' && (
            <motion.div
              key="convert"
              initial={{ opacity: 0, y: 20 }}
              animate={{ opacity: 1, y: 0 }}
              exit={{ opacity: 0, y: -20 }}
            >
              <Card variant="default">
                <CardHeader>
                  <CardTitle>Convert Currency</CardTitle>
                  <CardDescription>Exchange between currencies at competitive rates</CardDescription>
                </CardHeader>
                <CardContent>
                  <div className="max-w-md space-y-6">
                    <div>
                      <label className="text-sm font-medium mb-2 block">From</label>
                      <div className="flex space-x-2">
                        <Input
                          type="text"
                          placeholder="0.00"
                          variant="futuristic"
                          className="flex-1"
                        />
                        <select className="px-4 py-2 rounded-lg border bg-background">
                          <option value="USD">USD</option>
                          <option value="EUR">EUR</option>
                          <option value="GBP">GBP</option>
                          <option value="BTC">BTC</option>
                        </select>
                      </div>
                    </div>

                    <div className="flex justify-center">
                      <Button variant="default" size="icon" className="rounded-full">
                        <TrendingUp className="w-4 h-4" />
                      </Button>
                    </div>

                    <div>
                      <label className="text-sm font-medium mb-2 block">To</label>
                      <div className="flex space-x-2">
                        <Input
                          type="text"
                          placeholder="0.00"
                          variant="futuristic"
                          className="flex-1"
                          disabled
                        />
                        <select className="px-4 py-2 rounded-lg border bg-background">
                          <option value="EUR">EUR</option>
                          <option value="USD">USD</option>
                          <option value="GBP">GBP</option>
                          <option value="BTC">BTC</option>
                        </select>
                      </div>
                    </div>

                    <Card variant="futuristic" className="p-4">
                      <div className="space-y-2 text-sm">
                        <div className="flex justify-between">
                          <span className="text-muted-foreground">Exchange Rate</span>
                          <span>1 USD = 0.92 EUR</span>
                        </div>
                        <div className="flex justify-between">
                          <span className="text-muted-foreground">Fee</span>
                          <span>0.5%</span>
                        </div>
                        <div className="flex justify-between font-medium pt-2 border-t">
                          <span>You'll receive</span>
                          <span className="gradient-text">0.00 EUR</span>
                        </div>
                      </div>
                    </Card>

                    <Button className="w-full" variant="gradient" size="lg">
                      Convert
                      <TrendingUp className="ml-2 w-4 h-4" />
                    </Button>
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
