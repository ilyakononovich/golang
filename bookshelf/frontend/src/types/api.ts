// Types based on OpenAPI spec: project1-monolith-openapi.yaml

// ==================== AUTH ====================
export interface RegisterRequest {
  username: string;
  email: string;
  password: string;
}

export interface LoginRequest {
  email: string;
  password: string;
}

export interface AuthResponse {
  access_token: string;
  token_type: 'Bearer';
  expires_in: number;
  user: User;
}

// ==================== USER ====================
export interface User {
  id: string;
  username: string;
  email: string;
  created_at: string;
  updated_at: string;
}

export interface UpdateUserRequest {
  username?: string;
}

// ==================== BOOK ====================
export interface Book {
  id: string;
  title: string;
  author: string;
  description?: string;
  isbn?: string;
  published_year?: number;
  average_rating: number | null;
  reviews_count: number;
  created_by: string;
  created_at: string;
  updated_at: string;
}

export interface CreateBookRequest {
  title: string;
  author: string;
  description?: string;
  isbn?: string;
  published_year?: number;
}

export interface UpdateBookRequest {
  title?: string;
  author?: string;
  description?: string;
  isbn?: string;
  published_year?: number;
}

export interface BookListResponse {
  data: Book[];
  pagination: Pagination;
}

// ==================== REVIEW ====================
export interface Review {
  id: string;
  book_id: string;
  user_id: string;
  user: ReviewAuthor;
  rating: number;
  title?: string;
  content: string;
  created_at: string;
  updated_at: string;
}

export interface ReviewAuthor {
  id: string;
  username: string;
}

export interface CreateReviewRequest {
  rating: number;
  title?: string;
  content: string;
}

export interface UpdateReviewRequest {
  rating?: number;
  title?: string;
  content?: string;
}

export interface ReviewListResponse {
  data: Review[];
  pagination: Pagination;
}

// ==================== COMMON ====================
export interface Pagination {
  page: number;
  limit: number;
  total: number;
  total_pages: number;
}

export interface ApiError {
  code: string;
  message: string;
  details?: ErrorDetail[];
}

export interface ErrorDetail {
  field: string;
  message: string;
}

export interface HealthResponse {
  status: 'ok' | 'degraded' | 'unhealthy';
  version: string;
  timestamp: string;
}

// ==================== QUERY PARAMS ====================
export interface BooksParams {
  page?: number;
  limit?: number;
  search?: string;
  sort?: 'title' | 'author' | 'created_at' | 'rating';
  order?: 'asc' | 'desc';
}

export interface ReviewsParams {
  page?: number;
  limit?: number;
}





