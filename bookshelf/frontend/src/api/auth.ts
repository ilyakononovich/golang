import { useMutation, useQuery } from '@tanstack/react-query';
import { useNavigate } from 'react-router-dom';
import { toast } from 'sonner';
import { api } from './client';
import { useAuthStore } from '@/stores/auth';
import type { 
  AuthResponse, 
  LoginRequest, 
  RegisterRequest,
  User,
  UpdateUserRequest
} from '@/types/api';

export function useLogin() {
  const navigate = useNavigate();
  const setAuth = useAuthStore((state) => state.setAuth);

  return useMutation({
    mutationFn: async (data: LoginRequest) => {
      const response = await api.post<AuthResponse>('/api/v1/auth/login', data);
      return response.data;
    },
    onSuccess: (data) => {
      setAuth(data.access_token, data.user);
      toast.success('Добро пожаловать!');
      navigate('/');
    },
    onError: () => {
      toast.error('Неверный email или пароль');
    },
  });
}

export function useRegister() {
  const navigate = useNavigate();
  const setAuth = useAuthStore((state) => state.setAuth);

  return useMutation({
    mutationFn: async (data: RegisterRequest) => {
      const response = await api.post<AuthResponse>('/api/v1/auth/register', data);
      return response.data;
    },
    onSuccess: (data) => {
      setAuth(data.access_token, data.user);
      toast.success('Регистрация успешна!');
      navigate('/');
    },
    onError: (error: any) => {
      const message = error.response?.data?.message || 'Ошибка регистрации';
      toast.error(message);
    },
  });
}

export function useCurrentUser() {
  const token = useAuthStore((state) => state.token);
  
  return useQuery({
    queryKey: ['currentUser'],
    queryFn: async () => {
      const response = await api.get<User>('/api/v1/users/me');
      return response.data;
    },
    enabled: !!token,
  });
}

export function useUpdateProfile() {
  const setUser = useAuthStore((state) => state.setUser);

  return useMutation({
    mutationFn: async (data: UpdateUserRequest) => {
      const response = await api.put<User>('/api/v1/users/me', data);
      return response.data;
    },
    onSuccess: (data) => {
      setUser(data);
      toast.success('Профиль обновлён');
    },
    onError: () => {
      toast.error('Ошибка обновления профиля');
    },
  });
}



