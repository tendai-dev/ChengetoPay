'use client'

import { useState } from 'react'
import { useRouter } from 'next/navigation'
import Link from 'next/link'
import { useForm } from 'react-hook-form'
import { zodResolver } from '@hookform/resolvers/zod'
import * as z from 'zod'
import useAuthStore from '@/store/auth.store'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { Card, CardContent, CardDescription, CardFooter, CardHeader, CardTitle } from '@/components/ui/card'
import { Mail, Lock, User, Building, ArrowRight, Sparkles, Phone, Globe } from 'lucide-react'
import { toast } from 'react-hot-toast'

const registerSchema = z.object({
  firstName: z.string().min(2, 'First name must be at least 2 characters'),
  lastName: z.string().min(2, 'Last name must be at least 2 characters'),
  email: z.string().email('Invalid email address'),
  phone: z.string().min(10, 'Invalid phone number'),
  company: z.string().optional(),
  country: z.string().min(2, 'Please select your country'),
  password: z.string().min(8, 'Password must be at least 8 characters')
    .regex(/[A-Z]/, 'Password must contain at least one uppercase letter')
    .regex(/[a-z]/, 'Password must contain at least one lowercase letter')
    .regex(/[0-9]/, 'Password must contain at least one number')
    .regex(/[^A-Za-z0-9]/, 'Password must contain at least one special character'),
  confirmPassword: z.string(),
  acceptTerms: z.boolean().refine(val => val === true, 'You must accept the terms and conditions'),
}).refine(data => data.password === data.confirmPassword, {
  message: "Passwords don't match",
  path: ["confirmPassword"],
})

type RegisterFormData = z.infer<typeof registerSchema>

export default function RegisterPage() {
  const router = useRouter()
  const { register: registerUser } = useAuthStore()
  const [isLoading, setIsLoading] = useState(false)
  const [step, setStep] = useState(1)

  const {
    register,
    handleSubmit,
    formState: { errors },
    watch,
  } = useForm<RegisterFormData>({
    resolver: zodResolver(registerSchema),
  })

  const onSubmit = async (data: RegisterFormData) => {
    setIsLoading(true)
    try {
      await registerUser(data)
      toast.success('Account created successfully!')
      router.push('/dashboard')
    } catch (error: any) {
      toast.error(error.response?.data?.message || 'Registration failed')
    } finally {
      setIsLoading(false)
    }
  }

  const passwordStrength = (password: string) => {
    let strength = 0
    if (password.length >= 8) strength++
    if (/[A-Z]/.test(password)) strength++
    if (/[a-z]/.test(password)) strength++
    if (/[0-9]/.test(password)) strength++
    if (/[^A-Za-z0-9]/.test(password)) strength++
    return strength
  }

  const password = watch('password', '')
  const strength = passwordStrength(password)

  return (
    <div className="min-h-screen flex items-center justify-center px-4 py-12">
      {/* Animated background */}
      <div className="absolute inset-0 -z-10 overflow-hidden">
        <div className="absolute top-1/3 left-1/3 w-96 h-96 bg-gradient-to-r from-purple-500/20 to-pink-500/20 rounded-full filter blur-3xl animate-float" />
        <div className="absolute bottom-1/3 right-1/3 w-96 h-96 bg-gradient-to-r from-blue-500/20 to-cyan-500/20 rounded-full filter blur-3xl animate-float animation-delay-2000" />
      </div>

      <div className="w-full max-w-lg">
        {/* Logo */}
        <div className="text-center mb-8">
          <div className="inline-flex items-center space-x-2 mb-4">
            <div className="w-12 h-12 rounded-xl bg-gradient-to-br from-primary to-purple-600 flex items-center justify-center">
              <Sparkles className="w-6 h-6 text-white" />
            </div>
            <span className="text-2xl font-bold gradient-text">ChengetoPay</span>
          </div>
          <p className="text-muted-foreground">Join the payment revolution</p>
        </div>

        <Card variant="glass" className="backdrop-blur-2xl">
          <CardHeader>
            <CardTitle className="text-2xl">Create Account</CardTitle>
            <CardDescription>Step {step} of 2 - {step === 1 ? 'Personal Information' : 'Security'}</CardDescription>
            <div className="flex space-x-2 mt-4">
              <div className={`flex-1 h-2 rounded-full transition-all ${step >= 1 ? 'bg-primary' : 'bg-muted'}`} />
              <div className={`flex-1 h-2 rounded-full transition-all ${step >= 2 ? 'bg-primary' : 'bg-muted'}`} />
            </div>
          </CardHeader>
          <CardContent>
            <form onSubmit={handleSubmit(onSubmit)} className="space-y-4">
              {step === 1 && (
                <>
                  <div className="grid grid-cols-2 gap-4">
                    <div>
                      <label className="text-sm font-medium mb-2 block">First Name</label>
                      <Input
                        {...register('firstName')}
                        placeholder="John"
                        variant="futuristic"
                        icon={<User className="w-4 h-4" />}
                        error={errors.firstName?.message}
                      />
                    </div>
                    <div>
                      <label className="text-sm font-medium mb-2 block">Last Name</label>
                      <Input
                        {...register('lastName')}
                        placeholder="Doe"
                        variant="futuristic"
                        error={errors.lastName?.message}
                      />
                    </div>
                  </div>
                  <div>
                    <label className="text-sm font-medium mb-2 block">Email</label>
                    <Input
                      {...register('email')}
                      type="email"
                      placeholder="john@example.com"
                      variant="futuristic"
                      icon={<Mail className="w-4 h-4" />}
                      error={errors.email?.message}
                    />
                  </div>
                  <div>
                    <label className="text-sm font-medium mb-2 block">Phone</label>
                    <Input
                      {...register('phone')}
                      type="tel"
                      placeholder="+1 234 567 8900"
                      variant="futuristic"
                      icon={<Phone className="w-4 h-4" />}
                      error={errors.phone?.message}
                    />
                  </div>
                  <div>
                    <label className="text-sm font-medium mb-2 block">Company (Optional)</label>
                    <Input
                      {...register('company')}
                      placeholder="Acme Corp"
                      variant="futuristic"
                      icon={<Building className="w-4 h-4" />}
                      error={errors.company?.message}
                    />
                  </div>
                  <div>
                    <label className="text-sm font-medium mb-2 block">Country</label>
                    <Input
                      {...register('country')}
                      placeholder="United States"
                      variant="futuristic"
                      icon={<Globe className="w-4 h-4" />}
                      error={errors.country?.message}
                    />
                  </div>
                  <Button
                    type="button"
                    className="w-full"
                    variant="gradient"
                    size="lg"
                    onClick={() => setStep(2)}
                  >
                    Continue
                    <ArrowRight className="ml-2 w-4 h-4" />
                  </Button>
                </>
              )}

              {step === 2 && (
                <>
                  <div>
                    <label className="text-sm font-medium mb-2 block">Password</label>
                    <Input
                      {...register('password')}
                      type="password"
                      placeholder="••••••••"
                      variant="futuristic"
                      icon={<Lock className="w-4 h-4" />}
                      error={errors.password?.message}
                    />
                    {password && (
                      <div className="mt-2">
                        <div className="flex space-x-1">
                          {[1, 2, 3, 4, 5].map((i) => (
                            <div
                              key={i}
                              className={`flex-1 h-1 rounded-full transition-all ${
                                i <= strength
                                  ? strength <= 2
                                    ? 'bg-red-500'
                                    : strength <= 3
                                    ? 'bg-yellow-500'
                                    : 'bg-green-500'
                                  : 'bg-muted'
                              }`}
                            />
                          ))}
                        </div>
                        <p className="text-xs mt-1 text-muted-foreground">
                          Password strength: {strength <= 2 ? 'Weak' : strength <= 3 ? 'Medium' : 'Strong'}
                        </p>
                      </div>
                    )}
                  </div>
                  <div>
                    <label className="text-sm font-medium mb-2 block">Confirm Password</label>
                    <Input
                      {...register('confirmPassword')}
                      type="password"
                      placeholder="••••••••"
                      variant="futuristic"
                      icon={<Lock className="w-4 h-4" />}
                      error={errors.confirmPassword?.message}
                    />
                  </div>
                  <div className="flex items-start space-x-2">
                    <input
                      {...register('acceptTerms')}
                      type="checkbox"
                      className="mt-1 rounded border-gray-300"
                    />
                    <label className="text-sm text-muted-foreground">
                      I agree to the{' '}
                      <Link href="/terms" className="text-primary hover:underline">
                        Terms of Service
                      </Link>{' '}
                      and{' '}
                      <Link href="/privacy" className="text-primary hover:underline">
                        Privacy Policy
                      </Link>
                    </label>
                  </div>
                  {errors.acceptTerms && (
                    <p className="text-xs text-destructive">{errors.acceptTerms.message}</p>
                  )}
                  <div className="flex space-x-4">
                    <Button
                      type="button"
                      variant="outline"
                      className="flex-1"
                      onClick={() => setStep(1)}
                    >
                      Back
                    </Button>
                    <Button
                      type="submit"
                      className="flex-1"
                      variant="gradient"
                      size="lg"
                      loading={isLoading}
                    >
                      Create Account
                      <ArrowRight className="ml-2 w-4 h-4" />
                    </Button>
                  </div>
                </>
              )}
            </form>
          </CardContent>
          <CardFooter>
            <div className="text-center text-sm w-full">
              Already have an account?{' '}
              <Link href="/login" className="text-primary hover:underline font-medium">
                Sign in
              </Link>
            </div>
          </CardFooter>
        </Card>
      </div>
    </div>
  )
}
