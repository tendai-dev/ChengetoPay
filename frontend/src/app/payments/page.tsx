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
  Send, CreditCard, DollarSign, Users, ArrowRight, QrCode, 
  Wallet, Globe, Shield, Clock, TrendingUp, AlertCircle
} from 'lucide-react'
import { toast } from 'react-hot-toast'
import { motion } from 'framer-motion'
import { PaymentService } from '@/services/api/services'

const paymentSchema = z.object({
  amount: z.string().regex(/^\d+(\.\d{1,2})?$/, 'Invalid amount format'),
  currency: z.string().min(3, 'Select a currency'),
  recipient: z.string().email('Invalid recipient email'),
  description: z.string().min(5, 'Description must be at least 5 characters'),
  paymentMethod: z.enum(['card', 'wallet', 'bank', 'crypto']),
})

type PaymentFormData = z.infer<typeof paymentSchema>

interface PaymentMethod {
  id: string
  type: 'card' | 'wallet' | 'bank' | 'crypto'
  name: string
  last4?: string
  balance?: number
  icon: React.ReactNode
}

export default function PaymentsPage() {
  const router = useRouter()
  const [isProcessing, setIsProcessing] = useState(false)
  const [selectedMethod, setSelectedMethod] = useState<string>('card')
  const [showQR, setShowQR] = useState(false)
  const paymentService = new PaymentService()

  const {
    register,
    handleSubmit,
    formState: { errors },
    watch,
    setValue,
  } = useForm<PaymentFormData>({
    resolver: zodResolver(paymentSchema),
    defaultValues: {
      currency: 'USD',
      paymentMethod: 'card',
    },
  })

  const amount = watch('amount')
  const currency = watch('currency')

  const paymentMethods: PaymentMethod[] = [
    { id: '1', type: 'card', name: 'Visa', last4: '4242', icon: <CreditCard className="w-5 h-5" /> },
    { id: '2', type: 'wallet', name: 'ChengetoPay Wallet', balance: 5420.50, icon: <Wallet className="w-5 h-5" /> },
    { id: '3', type: 'bank', name: 'Chase Bank', last4: '1234', icon: <Globe className="w-5 h-5" /> },
    { id: '4', type: 'crypto', name: 'Bitcoin Wallet', balance: 0.045, icon: <Shield className="w-5 h-5" /> },
  ]

  const currencies = [
    { code: 'USD', symbol: '$', name: 'US Dollar' },
    { code: 'EUR', symbol: '€', name: 'Euro' },
    { code: 'GBP', symbol: '£', name: 'British Pound' },
    { code: 'JPY', symbol: '¥', name: 'Japanese Yen' },
    { code: 'BTC', symbol: '₿', name: 'Bitcoin' },
  ]

  const quickAmounts = [10, 25, 50, 100, 250, 500]

  const onSubmit = async (data: PaymentFormData) => {
    setIsProcessing(true)
    try {
      const response = await paymentService.createPayment({
        amount: parseFloat(data.amount),
        currency: data.currency,
        recipient: data.recipient,
        description: data.description,
        payment_method: data.paymentMethod,
      })
      
      toast.success('Payment initiated successfully!')
      router.push(`/payments/${(response.data as any)?.id || 'new'}`)
    } catch (error: any) {
      toast.error(error.response?.data?.message || 'Payment failed')
    } finally {
      setIsProcessing(false)
    }
  }

  const calculateFee = (amount: string) => {
    const value = parseFloat(amount) || 0
    return (value * 0.029 + 0.30).toFixed(2)
  }

  return (
    <div className="min-h-screen bg-background p-6">
      <div className="max-w-6xl mx-auto">
        {/* Header */}
        <div className="mb-8">
          <h1 className="text-3xl font-bold mb-2">Send Payment</h1>
          <p className="text-muted-foreground">Transfer funds instantly and securely</p>
        </div>

        <div className="grid lg:grid-cols-3 gap-6">
          {/* Payment Form */}
          <div className="lg:col-span-2 space-y-6">
            <Card variant="glass">
              <CardHeader>
                <CardTitle>Payment Details</CardTitle>
                <CardDescription>Enter the payment information below</CardDescription>
              </CardHeader>
              <CardContent>
                <form onSubmit={handleSubmit(onSubmit)} className="space-y-6">
                  {/* Amount Input */}
                  <div>
                    <label className="text-sm font-medium mb-2 block">Amount</label>
                    <div className="relative">
                      <Input
                        {...register('amount')}
                        type="text"
                        placeholder="0.00"
                        variant="futuristic"
                        className="text-2xl font-bold pl-12"
                        icon={<DollarSign className="w-6 h-6" />}
                        error={errors.amount?.message}
                      />
                    </div>
                    {/* Quick Amount Buttons */}
                    <div className="flex flex-wrap gap-2 mt-3">
                      {quickAmounts.map((quickAmount) => (
                        <Button
                          key={quickAmount}
                          type="button"
                          variant="outline"
                          size="sm"
                          onClick={() => setValue('amount', quickAmount.toString())}
                        >
                          ${quickAmount}
                        </Button>
                      ))}
                    </div>
                  </div>

                  {/* Currency Selection */}
                  <div>
                    <label className="text-sm font-medium mb-2 block">Currency</label>
                    <div className="grid grid-cols-3 gap-2">
                      {currencies.map((curr) => (
                        <Button
                          key={curr.code}
                          type="button"
                          variant={currency === curr.code ? 'gradient' : 'outline'}
                          onClick={() => setValue('currency', curr.code)}
                        >
                          {curr.symbol} {curr.code}
                        </Button>
                      ))}
                    </div>
                  </div>

                  {/* Recipient */}
                  <div>
                    <label className="text-sm font-medium mb-2 block">Recipient Email</label>
                    <Input
                      {...register('recipient')}
                      type="email"
                      placeholder="recipient@example.com"
                      variant="futuristic"
                      icon={<Users className="w-4 h-4" />}
                      error={errors.recipient?.message}
                    />
                  </div>

                  {/* Description */}
                  <div>
                    <label className="text-sm font-medium mb-2 block">Description</label>
                    <Input
                      {...register('description')}
                      type="text"
                      placeholder="Payment for services"
                      variant="futuristic"
                      error={errors.description?.message}
                    />
                  </div>

                  {/* Payment Method Selection */}
                  <div>
                    <label className="text-sm font-medium mb-3 block">Payment Method</label>
                    <div className="space-y-2">
                      {paymentMethods.map((method) => (
                        <motion.div
                          key={method.id}
                          whileHover={{ scale: 1.02 }}
                          whileTap={{ scale: 0.98 }}
                        >
                          <Card
                            variant={selectedMethod === method.type ? 'gradient' : 'glass'}
                            className={`p-4 cursor-pointer transition-all ${
                              selectedMethod === method.type ? 'ring-2 ring-primary' : ''
                            }`}
                            onClick={() => {
                              setSelectedMethod(method.type)
                              setValue('paymentMethod', method.type)
                            }}
                          >
                            <div className="flex items-center justify-between">
                              <div className="flex items-center space-x-3">
                                <div className="w-10 h-10 rounded-lg bg-primary/10 flex items-center justify-center">
                                  {method.icon}
                                </div>
                                <div>
                                  <p className="font-medium">{method.name}</p>
                                  {method.last4 && (
                                    <p className="text-sm text-muted-foreground">•••• {method.last4}</p>
                                  )}
                                  {method.balance !== undefined && (
                                    <p className="text-sm text-muted-foreground">
                                      Balance: {method.type === 'crypto' ? method.balance : `$${method.balance.toFixed(2)}`}
                                    </p>
                                  )}
                                </div>
                              </div>
                              <div className={`w-5 h-5 rounded-full border-2 ${
                                selectedMethod === method.type
                                  ? 'border-primary bg-primary'
                                  : 'border-muted-foreground'
                              }`}>
                                {selectedMethod === method.type && (
                                  <div className="w-full h-full flex items-center justify-center">
                                    <div className="w-2 h-2 bg-white rounded-full" />
                                  </div>
                                )}
                              </div>
                            </div>
                          </Card>
                        </motion.div>
                      ))}
                    </div>
                  </div>

                  {/* Submit Button */}
                  <Button
                    type="submit"
                    className="w-full"
                    variant="gradient"
                    size="lg"
                    loading={isProcessing}
                  >
                    {isProcessing ? 'Processing...' : 'Send Payment'}
                    <Send className="ml-2 w-4 h-4" />
                  </Button>
                </form>
              </CardContent>
            </Card>

            {/* Security Notice */}
            <Card variant="glass" className="border-yellow-500/50">
              <CardContent className="flex items-start space-x-3 pt-6">
                <AlertCircle className="w-5 h-5 text-yellow-500 mt-0.5" />
                <div>
                  <p className="font-medium">Security Notice</p>
                  <p className="text-sm text-muted-foreground mt-1">
                    All payments are protected by 256-bit encryption and multi-factor authentication.
                    Your financial information is never stored on our servers.
                  </p>
                </div>
              </CardContent>
            </Card>
          </div>

          {/* Payment Summary */}
          <div className="space-y-6">
            <Card variant="futuristic">
              <CardHeader>
                <CardTitle>Payment Summary</CardTitle>
              </CardHeader>
              <CardContent className="space-y-4">
                <div className="flex justify-between">
                  <span className="text-muted-foreground">Amount</span>
                  <span className="font-medium">
                    {currency} {amount || '0.00'}
                  </span>
                </div>
                <div className="flex justify-between">
                  <span className="text-muted-foreground">Processing Fee</span>
                  <span className="font-medium">
                    {currency} {amount ? calculateFee(amount) : '0.00'}
                  </span>
                </div>
                <div className="border-t pt-4">
                  <div className="flex justify-between">
                    <span className="font-medium">Total</span>
                    <span className="text-xl font-bold gradient-text">
                      {currency} {amount ? (parseFloat(amount) + parseFloat(calculateFee(amount))).toFixed(2) : '0.00'}
                    </span>
                  </div>
                </div>
              </CardContent>
            </Card>

            {/* QR Code */}
            <Card variant="glass">
              <CardHeader>
                <CardTitle className="flex items-center justify-between">
                  <span>Quick Pay QR</span>
                  <Button
                    variant="ghost"
                    size="sm"
                    onClick={() => setShowQR(!showQR)}
                  >
                    <QrCode className="w-4 h-4" />
                  </Button>
                </CardTitle>
              </CardHeader>
              {showQR && (
                <CardContent>
                  <div className="w-full aspect-square bg-white rounded-lg flex items-center justify-center">
                    <div className="w-48 h-48 bg-black/10 rounded" />
                  </div>
                  <p className="text-xs text-center text-muted-foreground mt-4">
                    Scan to pay with ChengetoPay mobile app
                  </p>
                </CardContent>
              )}
            </Card>

            {/* Recent Activity */}
            <Card variant="glass">
              <CardHeader>
                <CardTitle>Recent Payments</CardTitle>
              </CardHeader>
              <CardContent className="space-y-3">
                {[
                  { name: 'John Doe', amount: 150, time: '2 hours ago' },
                  { name: 'Acme Corp', amount: 2500, time: '5 hours ago' },
                  { name: 'Sarah Smith', amount: 75, time: 'Yesterday' },
                ].map((payment, index) => (
                  <div key={index} className="flex items-center justify-between">
                    <div className="flex items-center space-x-3">
                      <div className="w-8 h-8 rounded-full bg-primary/10 flex items-center justify-center">
                        <Send className="w-4 h-4" />
                      </div>
                      <div>
                        <p className="text-sm font-medium">{payment.name}</p>
                        <p className="text-xs text-muted-foreground">{payment.time}</p>
                      </div>
                    </div>
                    <span className="text-sm font-medium">${payment.amount}</span>
                  </div>
                ))}
              </CardContent>
            </Card>
          </div>
        </div>
      </div>
    </div>
  )
}
