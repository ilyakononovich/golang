import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import { toast } from 'sonner';
import { api } from './client';
import type { 
  Review,
  ReviewListResponse,
  ReviewsParams,
  CreateReviewRequest,
  UpdateReviewRequest
} from '@/types/api';

export function useBookReviews(bookId: string, params: ReviewsParams = {}) {
  return useQuery({
    queryKey: ['reviews', bookId, params],
    queryFn: async () => {
      const response = await api.get<ReviewListResponse>(
        `/api/v1/books/${bookId}/reviews`, 
        { params }
      );
      return response.data;
    },
    enabled: !!bookId,
    retry: false, // Не повторять если endpoint не реализован
  });
}

export function useReview(id: string) {
  return useQuery({
    queryKey: ['reviews', 'single', id],
    queryFn: async () => {
      const response = await api.get<Review>(`/api/v1/reviews/${id}`);
      return response.data;
    },
    enabled: !!id,
    retry: false,
  });
}

export function useCreateReview(bookId: string) {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: async (data: CreateReviewRequest) => {
      const response = await api.post<Review>(
        `/api/v1/books/${bookId}/reviews`, 
        data
      );
      return response.data;
    },
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['reviews', bookId] });
      queryClient.invalidateQueries({ queryKey: ['books', bookId] });
      queryClient.invalidateQueries({ queryKey: ['books'] });
      toast.success('Рецензия добавлена!');
    },
    onError: (error: any) => {
      const message = error.response?.status === 409 
        ? 'Вы уже оставили рецензию на эту книгу'
        : 'Ошибка добавления рецензии';
      toast.error(message);
    },
  });
}

export function useUpdateReview(reviewId: string, bookId: string) {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: async (data: UpdateReviewRequest) => {
      const response = await api.put<Review>(
        `/api/v1/reviews/${reviewId}`, 
        data
      );
      return response.data;
    },
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['reviews', bookId] });
      queryClient.invalidateQueries({ queryKey: ['reviews', 'single', reviewId] });
      queryClient.invalidateQueries({ queryKey: ['books', bookId] });
      queryClient.invalidateQueries({ queryKey: ['books'] });
      toast.success('Рецензия обновлена');
    },
    onError: () => {
      toast.error('Ошибка обновления рецензии');
    },
  });
}

export function useDeleteReview(bookId: string) {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: async (reviewId: string) => {
      await api.delete(`/api/v1/reviews/${reviewId}`);
    },
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['reviews', bookId] });
      queryClient.invalidateQueries({ queryKey: ['books', bookId] });
      queryClient.invalidateQueries({ queryKey: ['books'] });
      toast.success('Рецензия удалена');
    },
    onError: () => {
      toast.error('Ошибка удаления рецензии');
    },
  });
}



