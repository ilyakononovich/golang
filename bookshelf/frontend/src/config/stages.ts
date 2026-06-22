/**
 * Маппинг функций на этапы проекта
 * 
 * Каждая функция становится доступной после реализации соответствующего этапа.
 * Frontend автоматически определяет доступность по ответам API.
 */

export interface StageInfo {
  stage: number;
  name: string;
  description: string;
  hint: string;
  icon: string;
}

export const FEATURE_STAGES: Record<string, StageInfo> = {
  health: {
    stage: 3,
    name: 'Health Check',
    description: 'Проверка работоспособности сервера',
    hint: 'Реализуйте GET /health endpoint',
    icon: '🟢',
  },
  auth: {
    stage: 20,
    name: 'Авторизация',
    description: 'Регистрация и вход в систему',
    hint: 'Реализуйте POST /api/v1/auth/register и POST /api/v1/auth/login',
    icon: '👤',
  },
  books: {
    stage: 21,
    name: 'Каталог книг',
    description: 'Просмотр, создание и редактирование книг',
    hint: 'Реализуйте GET/POST /api/v1/books и CRUD для /api/v1/books/{id}',
    icon: '📚',
  },
  reviews: {
    stage: 22,
    name: 'Рецензии',
    description: 'Просмотр и написание рецензий на книги',
    hint: 'Реализуйте GET/POST /api/v1/books/{id}/reviews',
    icon: '⭐',
  },
};

/**
 * Определяет, является ли ошибка признаком нереализованного endpoint
 */
export function isFeatureNotImplemented(error: unknown): boolean {
  if (!error || typeof error !== 'object') return false;
  
  const axiosError = error as { response?: { status?: number }; code?: string };
  
  // 404 - endpoint не существует
  if (axiosError.response?.status === 404) return true;
  
  // Network error - сервер не запущен или endpoint не существует
  if (axiosError.code === 'ERR_NETWORK') return true;
  if (axiosError.code === 'ERR_CONNECTION_REFUSED') return true;
  
  return false;
}

/**
 * Определяет, является ли ошибка сетевой (сервер не запущен)
 */
export function isNetworkError(error: unknown): boolean {
  if (!error || typeof error !== 'object') return false;
  
  const axiosError = error as { code?: string; message?: string };
  
  if (axiosError.code === 'ERR_NETWORK') return true;
  if (axiosError.code === 'ERR_CONNECTION_REFUSED') return true;
  if (axiosError.message?.includes('Network Error')) return true;
  
  return false;
}
