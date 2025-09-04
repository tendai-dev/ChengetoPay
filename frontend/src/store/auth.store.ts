import { create } from 'zustand';
import { persist, createJSONStorage } from 'zustand/middleware';
import { authService } from '@/services/api/services';
import { setTokens, clearTokens } from '@/services/api/client';
import wsClient from '@/services/websocket/client';
import { toast } from 'react-hot-toast';

interface User {
  id: string;
  email: string;
  firstName: string;
  lastName: string;
  role: 'admin' | 'merchant' | 'user';
  avatar?: string;
  verified: boolean;
  twoFactorEnabled: boolean;
  permissions: string[];
  metadata: {
    lastLogin: string;
    createdAt: string;
    updatedAt: string;
  };
}

interface AuthState {
  user: User | null;
  isAuthenticated: boolean;
  isLoading: boolean;
  error: string | null;
  sessionExpiry: number | null;
  
  // Actions
  login: (email: string, password: string) => Promise<void>;
  register: (userData: any) => Promise<void>;
  logout: () => Promise<void>;
  refreshSession: () => Promise<void>;
  updateUser: (updates: Partial<User>) => void;
  checkAuth: () => Promise<void>;
  enable2FA: () => Promise<void>;
  verify2FA: (code: string) => Promise<void>;
  resetPassword: (email: string) => Promise<void>;
  updatePassword: (token: string, newPassword: string) => Promise<void>;
}

const useAuthStore = create<AuthState>()(
  persist(
    (set, get) => ({
      user: null,
      isAuthenticated: false,
      isLoading: false,
      error: null,
      sessionExpiry: null,

      login: async (email: string, password: string) => {
        set({ isLoading: true, error: null });
        
        try {
          const response = await authService.login(email, password);
          const { user, accessToken, refreshToken, expiresIn } = (response.data as any);
          
          setTokens(accessToken, refreshToken);
          await wsClient.connect(accessToken);
          
          set({
            user,
            isAuthenticated: true,
            isLoading: false,
            sessionExpiry: Date.now() + expiresIn * 1000,
          });
          
          toast.success(`Welcome back, ${user.firstName}!`);
        } catch (error: any) {
          set({
            isLoading: false,
            error: error.response?.data?.message || 'Login failed',
          });
          throw error;
        }
      },

      register: async (userData: any) => {
        set({ isLoading: true, error: null });
        
        try {
          const response = await authService.register(userData);
          const { user, accessToken, refreshToken, expiresIn } = (response.data as any);
          
          setTokens(accessToken, refreshToken);
          await wsClient.connect(accessToken);
          
          set({
            user,
            isAuthenticated: true,
            isLoading: false,
            sessionExpiry: Date.now() + expiresIn * 1000,
          });
          
          toast.success('Account created successfully!');
        } catch (error: any) {
          set({
            isLoading: false,
            error: error.response?.data?.message || 'Registration failed',
          });
          throw error;
        }
      },

      logout: async () => {
        set({ isLoading: true });
        
        try {
          await authService.logout();
        } catch (error) {
          console.error('Logout error:', error);
        } finally {
          clearTokens();
          wsClient.disconnect();
          
          set({
            user: null,
            isAuthenticated: false,
            isLoading: false,
            sessionExpiry: null,
          });
          
          toast.success('Logged out successfully');
        }
      },

      refreshSession: async () => {
        try {
          const refreshToken = localStorage.getItem('refreshToken');
          if (!refreshToken) throw new Error('No refresh token');
          
          const response = await authService.refreshToken(refreshToken);
          const { accessToken, expiresIn } = (response.data as any);
          
          setTokens(accessToken, refreshToken);
          
          set({
            sessionExpiry: Date.now() + expiresIn * 1000,
          });
        } catch (error) {
          get().logout();
          throw error;
        }
      },

      updateUser: (updates: Partial<User>) => {
        set((state) => ({
          user: state.user ? { ...state.user, ...updates } : null,
        }));
      },

      checkAuth: async () => {
        const accessToken = localStorage.getItem('accessToken');
        if (!accessToken) {
          set({ isAuthenticated: false, user: null });
          return;
        }

        set({ isLoading: true });
        
        try {
          // Note: getCurrentUser is not available in AuthService
          // This would need to be implemented or moved to UserService
          console.log('getCurrentUser not implemented in AuthService');
          
          set({
            isLoading: false,
          });
        } catch (error) {
          clearTokens();
          set({
            user: null,
            isAuthenticated: false,
            isLoading: false,
          });
        }
      },

      enable2FA: async () => {
        set({ isLoading: true, error: null });
        
        try {
          const response = await authService.enable2FA();
          toast.success('2FA enabled successfully');
          
          set((state) => ({
            user: state.user ? { ...state.user, twoFactorEnabled: true } : null,
            isLoading: false,
          }));
          
          // return response.data;
        } catch (error: any) {
          set({
            isLoading: false,
            error: error.response?.data?.message || '2FA setup failed',
          });
          throw error;
        }
      },

      verify2FA: async (code: string) => {
        set({ isLoading: true, error: null });
        
        try {
          await authService.verify2FA(code);
          toast.success('2FA verification successful');
          set({ isLoading: false });
        } catch (error: any) {
          set({
            isLoading: false,
            error: error.response?.data?.message || '2FA verification failed',
          });
          throw error;
        }
      },

      resetPassword: async (email: string) => {
        set({ isLoading: true, error: null });
        
        try {
          await authService.resetPassword(email);
          toast.success('Password reset email sent');
          set({ isLoading: false });
        } catch (error: any) {
          set({
            isLoading: false,
            error: error.response?.data?.message || 'Password reset failed',
          });
          throw error;
        }
      },

      updatePassword: async (token: string, newPassword: string) => {
        set({ isLoading: true, error: null });
        
        try {
          await authService.updatePassword(token, newPassword);
          toast.success('Password updated successfully');
          set({ isLoading: false });
        } catch (error: any) {
          set({
            isLoading: false,
            error: error.response?.data?.message || 'Password update failed',
          });
          throw error;
        }
      },
    }),
    {
      name: 'auth-storage',
      storage: createJSONStorage(() => localStorage),
      partialize: (state) => ({
        user: state.user,
        isAuthenticated: state.isAuthenticated,
        sessionExpiry: state.sessionExpiry,
      }),
    }
  )
);

export default useAuthStore;
