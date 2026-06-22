import { useState } from 'react';
import { useSearchParams } from 'react-router-dom';
import { ChevronLeft, ChevronRight, BookOpen } from 'lucide-react';
import { Button } from '@/components/ui/button';
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from '@/components/ui/select';
import { BookGrid } from '@/components/books/BookGrid';
import { useBooks } from '@/api/books';
import { FeatureLocked, FeatureError } from '@/components/ui/FeatureLocked';
import { FEATURE_STAGES, isFeatureNotImplemented, isNetworkError } from '@/config/stages';

export function HomePage() {
  const [searchParams, setSearchParams] = useSearchParams();
  const [sort, setSort] = useState<string>(searchParams.get('sort') || 'created_at');
  const [order, setOrder] = useState<string>(searchParams.get('order') || 'desc');
  
  const page = parseInt(searchParams.get('page') || '1');
  const search = searchParams.get('search') || undefined;
  
  const { data, isLoading, isError, error, refetch } = useBooks({
    page,
    limit: 20,
    search,
    sort: sort as any,
    order: order as any,
  });

  // Если endpoint не реализован — показываем FeatureLocked
  if (isError && isFeatureNotImplemented(error)) {
    return (
      <div className="max-w-lg mx-auto py-12">
        <FeatureLocked
          title="📚 Каталог книг"
          description="Список книг станет доступен после реализации BookHandler"
          stage={FEATURE_STAGES.books.stage}
          hint={FEATURE_STAGES.books.hint}
          icon={<BookOpen className="h-8 w-8 text-muted-foreground" />}
        />
      </div>
    );
  }

  // Если сетевая ошибка — показываем FeatureError
  if (isError && isNetworkError(error)) {
    return (
      <div className="max-w-lg mx-auto py-12">
        <FeatureError
          title="Не удалось загрузить каталог"
          error={error as Error}
          onRetry={() => refetch()}
        />
      </div>
    );
  }

  const handleSortChange = (value: string) => {
    setSort(value);
    setSearchParams((prev) => {
      prev.set('sort', value);
      prev.delete('page');
      return prev;
    });
  };

  const handleOrderChange = (value: string) => {
    setOrder(value);
    setSearchParams((prev) => {
      prev.set('order', value);
      prev.delete('page');
      return prev;
    });
  };

  const handlePageChange = (newPage: number) => {
    setSearchParams((prev) => {
      prev.set('page', String(newPage));
      return prev;
    });
    window.scrollTo({ top: 0, behavior: 'smooth' });
  };

  return (
    <div className="space-y-6">
      <div className="flex flex-col sm:flex-row gap-4 items-start sm:items-center justify-between">
        <div>
          <h1 className="font-display text-3xl font-bold">
            {search ? `Поиск: "${search}"` : 'Каталог книг'}
          </h1>
          {data && (
            <p className="text-muted-foreground mt-1">
              Найдено {data.pagination.total} книг
            </p>
          )}
        </div>

        <div className="flex gap-2">
          <Select value={sort} onValueChange={handleSortChange}>
            <SelectTrigger className="w-[160px]">
              <SelectValue placeholder="Сортировка" />
            </SelectTrigger>
            <SelectContent>
              <SelectItem value="created_at">По дате</SelectItem>
              <SelectItem value="title">По названию</SelectItem>
              <SelectItem value="author">По автору</SelectItem>
              <SelectItem value="rating">По рейтингу</SelectItem>
            </SelectContent>
          </Select>

          <Select value={order} onValueChange={handleOrderChange}>
            <SelectTrigger className="w-[140px]">
              <SelectValue placeholder="Порядок" />
            </SelectTrigger>
            <SelectContent>
              <SelectItem value="desc">По убыванию</SelectItem>
              <SelectItem value="asc">По возрастанию</SelectItem>
            </SelectContent>
          </Select>
        </div>
      </div>

      <BookGrid books={data?.data || []} loading={isLoading} />

      {data && data.pagination.total_pages > 1 && (
        <div className="flex items-center justify-center gap-2">
          <Button
            variant="outline"
            size="icon"
            disabled={page <= 1}
            onClick={() => handlePageChange(page - 1)}
          >
            <ChevronLeft className="h-4 w-4" />
          </Button>
          
          <span className="text-sm text-muted-foreground px-4">
            Страница {page} из {data.pagination.total_pages}
          </span>
          
          <Button
            variant="outline"
            size="icon"
            disabled={page >= data.pagination.total_pages}
            onClick={() => handlePageChange(page + 1)}
          >
            <ChevronRight className="h-4 w-4" />
          </Button>
        </div>
      )}
    </div>
  );
}





