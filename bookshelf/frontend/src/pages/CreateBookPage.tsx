import { useNavigate, Link } from 'react-router-dom';
import { useForm } from 'react-hook-form';
import { zodResolver } from '@hookform/resolvers/zod';
import { z } from 'zod';
import { ArrowLeft } from 'lucide-react';
import { motion } from 'framer-motion';
import { Button } from '@/components/ui/button';
import { Input } from '@/components/ui/input';
import { Textarea } from '@/components/ui/textarea';
import { Card, CardContent, CardHeader, CardTitle, CardDescription } from '@/components/ui/card';
import {
  Form,
  FormControl,
  FormField,
  FormItem,
  FormLabel,
  FormMessage,
  FormDescription,
} from '@/components/ui/form';
import { useCreateBook } from '@/api/books';

const bookSchema = z.object({
  title: z.string().min(1, 'Обязательное поле').max(255),
  author: z.string().min(1, 'Обязательное поле').max(255),
  description: z.string().max(2000).optional(),
  isbn: z.string().optional(),
  published_year: z.coerce.number().min(1000).max(2100).optional().or(z.literal('')),
});

type BookFormData = z.infer<typeof bookSchema>;

export function CreateBookPage() {
  const navigate = useNavigate();
  const createBook = useCreateBook();
  
  const form = useForm<BookFormData>({
    resolver: zodResolver(bookSchema),
    defaultValues: {
      title: '',
      author: '',
      description: '',
      isbn: '',
      published_year: '',
    },
  });

  const onSubmit = async (data: BookFormData) => {
    const book = await createBook.mutateAsync({
      ...data,
      published_year: data.published_year ? Number(data.published_year) : undefined,
    });
    navigate(`/books/${book.id}`);
  };

  return (
    <motion.div
      initial={{ opacity: 0, y: 20 }}
      animate={{ opacity: 1, y: 0 }}
      className="max-w-2xl mx-auto space-y-6"
    >
      <Button variant="ghost" asChild>
        <Link to="/">
          <ArrowLeft className="h-4 w-4 mr-2" />
          Назад
        </Link>
      </Button>

      <Card>
        <CardHeader>
          <CardTitle className="font-display text-2xl">Добавить книгу</CardTitle>
          <CardDescription>
            Заполните информацию о книге для добавления в каталог
          </CardDescription>
        </CardHeader>
        <CardContent>
          <Form {...form}>
            <form onSubmit={form.handleSubmit(onSubmit)} className="space-y-6">
              <FormField
                control={form.control}
                name="title"
                render={({ field }) => (
                  <FormItem>
                    <FormLabel>Название *</FormLabel>
                    <FormControl>
                      <Input placeholder="Чистая архитектура" {...field} />
                    </FormControl>
                    <FormMessage />
                  </FormItem>
                )}
              />

              <FormField
                control={form.control}
                name="author"
                render={({ field }) => (
                  <FormItem>
                    <FormLabel>Автор *</FormLabel>
                    <FormControl>
                      <Input placeholder="Роберт Мартин" {...field} />
                    </FormControl>
                    <FormMessage />
                  </FormItem>
                )}
              />

              <FormField
                control={form.control}
                name="description"
                render={({ field }) => (
                  <FormItem>
                    <FormLabel>Описание</FormLabel>
                    <FormControl>
                      <Textarea 
                        placeholder="Краткое описание книги..."
                        className="min-h-[120px]"
                        {...field} 
                      />
                    </FormControl>
                    <FormDescription>Максимум 2000 символов</FormDescription>
                    <FormMessage />
                  </FormItem>
                )}
              />

              <div className="grid grid-cols-2 gap-4">
                <FormField
                  control={form.control}
                  name="isbn"
                  render={({ field }) => (
                    <FormItem>
                      <FormLabel>ISBN</FormLabel>
                      <FormControl>
                        <Input placeholder="978-5-4461-0772-8" {...field} />
                      </FormControl>
                      <FormMessage />
                    </FormItem>
                  )}
                />

                <FormField
                  control={form.control}
                  name="published_year"
                  render={({ field }) => (
                    <FormItem>
                      <FormLabel>Год публикации</FormLabel>
                      <FormControl>
                        <Input 
                          type="number" 
                          placeholder="2018" 
                          {...field} 
                        />
                      </FormControl>
                      <FormMessage />
                    </FormItem>
                  )}
                />
              </div>

              <div className="flex gap-4">
                <Button 
                  type="submit" 
                  disabled={createBook.isPending}
                >
                  {createBook.isPending ? 'Добавление...' : 'Добавить книгу'}
                </Button>
                <Button type="button" variant="outline" onClick={() => navigate(-1)}>
                  Отмена
                </Button>
              </div>
            </form>
          </Form>
        </CardContent>
      </Card>
    </motion.div>
  );
}





