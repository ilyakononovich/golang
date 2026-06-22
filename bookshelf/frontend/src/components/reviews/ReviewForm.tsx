import { useForm } from 'react-hook-form';
import { zodResolver } from '@hookform/resolvers/zod';
import { z } from 'zod';
import { Button } from '@/components/ui/button';
import { Textarea } from '@/components/ui/textarea';
import { Input } from '@/components/ui/input';
import {
  Form,
  FormControl,
  FormField,
  FormItem,
  FormLabel,
  FormMessage,
} from '@/components/ui/form';
import { StarRating } from './StarRating';
import { useCreateReview } from '@/api/reviews';

const reviewSchema = z.object({
  rating: z.number().min(1, 'Выберите оценку').max(5),
  title: z.string().max(255).optional(),
  content: z.string().min(10, 'Минимум 10 символов').max(5000),
});

type ReviewFormData = z.infer<typeof reviewSchema>;

interface ReviewFormProps {
  bookId: string;
  onSuccess?: () => void;
}

export function ReviewForm({ bookId, onSuccess }: ReviewFormProps) {
  const createReview = useCreateReview(bookId);
  
  const form = useForm<ReviewFormData>({
    resolver: zodResolver(reviewSchema),
    defaultValues: {
      rating: 0,
      title: '',
      content: '',
    },
  });

  const onSubmit = async (data: ReviewFormData) => {
    await createReview.mutateAsync(data);
    form.reset();
    onSuccess?.();
  };

  return (
    <Form {...form}>
      <form onSubmit={form.handleSubmit(onSubmit)} className="space-y-4">
        <FormField
          control={form.control}
          name="rating"
          render={({ field }) => (
            <FormItem>
              <FormLabel>Оценка</FormLabel>
              <FormControl>
                <StarRating
                  value={field.value}
                  onChange={field.onChange}
                  size="lg"
                />
              </FormControl>
              <FormMessage />
            </FormItem>
          )}
        />

        <FormField
          control={form.control}
          name="title"
          render={({ field }) => (
            <FormItem>
              <FormLabel>Заголовок (необязательно)</FormLabel>
              <FormControl>
                <Input placeholder="Краткое резюме" {...field} />
              </FormControl>
              <FormMessage />
            </FormItem>
          )}
        />

        <FormField
          control={form.control}
          name="content"
          render={({ field }) => (
            <FormItem>
              <FormLabel>Рецензия</FormLabel>
              <FormControl>
                <Textarea 
                  placeholder="Поделитесь своими впечатлениями о книге..."
                  className="min-h-[120px]"
                  {...field} 
                />
              </FormControl>
              <FormMessage />
            </FormItem>
          )}
        />

        <Button 
          type="submit" 
          disabled={createReview.isPending}
        >
          {createReview.isPending ? 'Отправка...' : 'Отправить рецензию'}
        </Button>
      </form>
    </Form>
  );
}





