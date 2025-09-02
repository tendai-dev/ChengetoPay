import axios, { AxiosInstance, AxiosRequestConfig, AxiosResponse, InternalAxiosRequestConfig } from 'axios';
import { toast } from 'react-hot-toast';

// API Configuration
const API_BASE_URL = process.env.NEXT_PUBLIC_API_URL || 'http://localhost:8080';
const REQUEST_TIMEOUT = 30000;

// Create axios instance with optimized config
const apiClient: AxiosInstance = axios.create({
  baseURL: API_BASE_URL,
  timeout: REQUEST_TIMEOUT,
  headers: {
    'Content-Type': 'application/json',
    'X-Client-Version': '1.0.0',
    'X-Client-Platform': 'web',
  },
  withCredentials: true,
});

// Token management
let accessToken: string | null = null;
let refreshToken: string | null = null;
let isRefreshing = false;
let failedQueue: Array<{
  resolve: (token: string) => void;
  reject: (error: any) => void;
}> = [];

const processQueue = (error: any, token: string | null = null) => {
  failedQueue.forEach(prom => {
    if (error) {
      prom.reject(error);
    } else {
      prom.resolve(token!);
    }
  });
  failedQueue = [];
};

// Request interceptor
apiClient.interceptors.request.use(
  (config: InternalAxiosRequestConfig) => {
    // Add auth token
    if (accessToken && config.headers) {
      config.headers.Authorization = `Bearer ${accessToken}`;
    }

    // Add request ID for tracing
    config.headers['X-Request-ID'] = generateRequestId();
    
    // Add timestamp
    config.headers['X-Request-Timestamp'] = new Date().toISOString();

    return config;
  },
  (error) => {
    return Promise.reject(error);
  }
);

// Response interceptor with automatic retry and token refresh
apiClient.interceptors.response.use(
  (response: AxiosResponse) => {
    // Track performance metrics
    const responseTime = Date.now() - (response.config as any).startTime;
    if (responseTime > 3000) {
      console.warn(`Slow API response: ${response.config.url} took ${responseTime}ms`);
    }
    return response;
  },
  async (error) => {
    const originalRequest = error.config;

    // Handle 401 - Unauthorized
    if (error.response?.status === 401 && !originalRequest._retry) {
      if (isRefreshing) {
        return new Promise((resolve, reject) => {
          failedQueue.push({ resolve, reject });
        }).then(token => {
          originalRequest.headers.Authorization = `Bearer ${token}`;
          return apiClient(originalRequest);
        }).catch(err => {
          return Promise.reject(err);
        });
      }

      originalRequest._retry = true;
      isRefreshing = true;

      try {
        const response = await refreshAccessToken();
        const { accessToken: newAccessToken } = response.data;
        setTokens(newAccessToken, refreshToken);
        processQueue(null, newAccessToken);
        originalRequest.headers.Authorization = `Bearer ${newAccessToken}`;
        return apiClient(originalRequest);
      } catch (refreshError) {
        processQueue(refreshError, null);
        clearTokens();
        window.location.href = '/login';
        return Promise.reject(refreshError);
      } finally {
        isRefreshing = false;
      }
    }

    // Handle other errors
    handleApiError(error);
    return Promise.reject(error);
  }
);

// Token management functions
export const setTokens = (access: string | null, refresh: string | null) => {
  accessToken = access;
  refreshToken = refresh;
  if (access) {
    localStorage.setItem('accessToken', access);
  }
  if (refresh) {
    localStorage.setItem('refreshToken', refresh);
  }
};

export const getTokens = () => ({
  accessToken: accessToken || localStorage.getItem('accessToken'),
  refreshToken: refreshToken || localStorage.getItem('refreshToken'),
});

export const clearTokens = () => {
  accessToken = null;
  refreshToken = null;
  localStorage.removeItem('accessToken');
  localStorage.removeItem('refreshToken');
};

const refreshAccessToken = async () => {
  return apiClient.post('/auth/refresh', {
    refreshToken: refreshToken || localStorage.getItem('refreshToken'),
  });
};

// Error handling
const handleApiError = (error: any) => {
  if (error.response) {
    const { status, data } = error.response;
    const message = data?.message || 'An error occurred';

    switch (status) {
      case 400:
        toast.error(`Bad Request: ${message}`);
        break;
      case 403:
        toast.error('You do not have permission to perform this action');
        break;
      case 404:
        toast.error('Resource not found');
        break;
      case 429:
        toast.error('Too many requests. Please try again later');
        break;
      case 500:
        toast.error('Server error. Please try again later');
        break;
      default:
        toast.error(message);
    }
  } else if (error.request) {
    toast.error('Network error. Please check your connection');
  } else {
    toast.error('An unexpected error occurred');
  }
};

// Utility functions
const generateRequestId = (): string => {
  return `${Date.now()}-${Math.random().toString(36).substr(2, 9)}`;
};

// Performance monitoring
apiClient.interceptors.request.use(
  (config: any) => {
    config.startTime = Date.now();
    return config;
  },
  (error) => Promise.reject(error)
);

// Retry logic for failed requests
const retryRequest = async (config: AxiosRequestConfig, retries = 3): Promise<AxiosResponse> => {
  try {
    return await apiClient(config);
  } catch (error: any) {
    if (retries > 0 && error.response?.status >= 500) {
      await new Promise(resolve => setTimeout(resolve, 1000 * (4 - retries)));
      return retryRequest(config, retries - 1);
    }
    throw error;
  }
};

// Export configured client
export default apiClient;
export { retryRequest };
