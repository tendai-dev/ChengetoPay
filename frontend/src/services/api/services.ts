import apiClient from './client';
import { AxiosResponse } from 'axios';

// Type definitions
export interface ServiceResponse<T = any> {
  data: T;
  success: boolean;
  message?: string;
  metadata?: {
    timestamp: string;
    requestId: string;
    version: string;
  };
}

// Base service class for common functionality
class BaseService {
  protected endpoint: string;

  constructor(endpoint: string) {
    this.endpoint = endpoint;
  }

  protected async get<T>(path: string, params?: any): Promise<ServiceResponse<T>> {
    const response = await apiClient.get<ServiceResponse<T>>(`${this.endpoint}${path}`, { params });
    return response.data;
  }

  protected async post<T>(path: string, data?: any): Promise<ServiceResponse<T>> {
    const response = await apiClient.post<ServiceResponse<T>>(`${this.endpoint}${path}`, data);
    return response.data;
  }

  protected async put<T>(path: string, data?: any): Promise<ServiceResponse<T>> {
    const response = await apiClient.put<ServiceResponse<T>>(`${this.endpoint}${path}`, data);
    return response.data;
  }

  protected async delete<T>(path: string): Promise<ServiceResponse<T>> {
    const response = await apiClient.delete<ServiceResponse<T>>(`${this.endpoint}${path}`);
    return response.data;
  }

  protected async patch<T>(path: string, data?: any): Promise<ServiceResponse<T>> {
    const response = await apiClient.patch<ServiceResponse<T>>(`${this.endpoint}${path}`, data);
    return response.data;
  }
}

// Authentication Service
export class AuthService extends BaseService {
  constructor() {
    super('/auth');
  }

  async login(email: string, password: string) {
    return this.post('/login', { email, password });
  }

  async register(userData: any) {
    return this.post('/register', userData);
  }

  async logout() {
    return this.post('/logout');
  }

  async refreshToken(refreshToken: string) {
    return this.post('/refresh', { refreshToken });
  }

  async verifyEmail(token: string) {
    return this.post('/verify-email', { token });
  }

  async resetPassword(email: string) {
    return this.post('/reset-password', { email });
  }

  async updatePassword(token: string, newPassword: string) {
    return this.post('/update-password', { token, newPassword });
  }

  async enable2FA() {
    return this.post('/2fa/enable');
  }

  async verify2FA(code: string) {
    return this.post('/2fa/verify', { code });
  }
}

// User Service
export class UserService extends BaseService {
  constructor() {
    super('/users');
  }

  async getCurrentUser() {
    return this.get('/me');
  }

  async updateProfile(data: any) {
    return this.put('/me', data);
  }

  async uploadAvatar(file: File) {
    const formData = new FormData();
    formData.append('avatar', file);
    return this.post('/me/avatar', formData);
  }

  async getNotifications() {
    return this.get('/notifications');
  }

  async markNotificationRead(id: string) {
    return this.put(`/notifications/${id}/read`);
  }
}

// Payment Service
export class PaymentService extends BaseService {
  constructor() {
    super('/payments');
  }

  async createPayment(data: any) {
    return this.post('/create', data);
  }

  async getPayments(filters?: any) {
    return this.get('/', filters);
  }

  async getPaymentById(id: string) {
    return this.get(`/${id}`);
  }

  async confirmPayment(id: string, data: any) {
    return this.post(`/${id}/confirm`, data);
  }

  async cancelPayment(id: string) {
    return this.post(`/${id}/cancel`);
  }

  async refundPayment(id: string, amount?: number) {
    return this.post(`/${id}/refund`, { amount });
  }
}

// Transaction Service
export class TransactionService extends BaseService {
  constructor() {
    super('/transactions');
  }

  async getTransactions(filters?: any) {
    return this.get('/', filters);
  }

  async getTransactionById(id: string) {
    return this.get(`/${id}`);
  }

  async getTransactionHistory(days: number = 30) {
    return this.get('/history', { days });
  }

  async exportTransactions(format: 'csv' | 'pdf' | 'excel') {
    return this.get('/export', { format });
  }
}

// Merchant Service
export class MerchantService extends BaseService {
  constructor() {
    super('/merchants');
  }

  async getMerchantProfile() {
    return this.get('/profile');
  }

  async updateMerchantProfile(data: any) {
    return this.put('/profile', data);
  }

  async getMerchantStats() {
    return this.get('/stats');
  }

  async getMerchantSettlements() {
    return this.get('/settlements');
  }

  async createApiKey() {
    return this.post('/api-keys');
  }

  async revokeApiKey(keyId: string) {
    return this.delete(`/api-keys/${keyId}`);
  }
}

// Escrow Service
export class EscrowService extends BaseService {
  constructor() {
    super('/escrow');
  }

  async createEscrow(data: any) {
    return this.post('/create', data);
  }

  async getEscrows() {
    return this.get('/');
  }

  async releaseEscrow(id: string) {
    return this.post(`/${id}/release`);
  }

  async disputeEscrow(id: string, reason: string) {
    return this.post(`/${id}/dispute`, { reason });
  }
}

// Webhook Service
export class WebhookService extends BaseService {
  constructor() {
    super('/webhooks');
  }

  async getWebhooks() {
    return this.get('/');
  }

  async createWebhook(url: string, events: string[]) {
    return this.post('/', { url, events });
  }

  async updateWebhook(id: string, data: any) {
    return this.put(`/${id}`, data);
  }

  async deleteWebhook(id: string) {
    return this.delete(`/${id}`);
  }

  async testWebhook(id: string) {
    return this.post(`/${id}/test`);
  }
}

// Reporting Service
export class ReportingService extends BaseService {
  constructor() {
    super('/reports');
  }

  async getDashboardData() {
    return this.get('/dashboard');
  }

  async getAnalytics(period: string) {
    return this.get('/analytics', { period });
  }

  async generateReport(type: string, params: any) {
    return this.post('/generate', { type, ...params });
  }

  async getReportHistory() {
    return this.get('/history');
  }
}

// Compliance Service
export class ComplianceService extends BaseService {
  constructor() {
    super('/compliance');
  }

  async getComplianceStatus() {
    return this.get('/status');
  }

  async submitKYC(data: any) {
    return this.post('/kyc', data);
  }

  async submitKYB(data: any) {
    return this.post('/kyb', data);
  }

  async getAuditLogs(filters?: any) {
    return this.get('/audit-logs', filters);
  }
}

// Risk Service
export class RiskService extends BaseService {
  constructor() {
    super('/risk');
  }

  async assessRisk(transactionData: any) {
    return this.post('/assess', transactionData);
  }

  async getRiskScore(entityId: string) {
    return this.get(`/score/${entityId}`);
  }

  async getRiskRules() {
    return this.get('/rules');
  }

  async updateRiskRules(rules: any) {
    return this.put('/rules', rules);
  }
}

// Notification Service
export class NotificationService extends BaseService {
  constructor() {
    super('/notifications');
  }

  async getNotifications() {
    return this.get('/');
  }

  async markAsRead(id: string) {
    return this.put(`/${id}/read`);
  }

  async markAllAsRead() {
    return this.put('/read-all');
  }

  async updatePreferences(preferences: any) {
    return this.put('/preferences', preferences);
  }
}

// Export service instances
export const authService = new AuthService();
export const userService = new UserService();
export const paymentService = new PaymentService();
export const transactionService = new TransactionService();
export const merchantService = new MerchantService();
export const escrowService = new EscrowService();
export const webhookService = new WebhookService();
export const reportingService = new ReportingService();
export const complianceService = new ComplianceService();
export const riskService = new RiskService();
export const notificationService = new NotificationService();
