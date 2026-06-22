import { Link } from 'react-router-dom';
import { Card, CardContent } from '@/components/ui/card';
import { StarRating } from '@/components/reviews/StarRating';
import type { Book } from '@/types/api';

interface BookCardProps {
  book: Book;
}

export function BookCard({ book }: BookCardProps) {
  return (
    <Link to={`/books/${book.id}`}>
      <Card className="book-card h-full overflow-hidden hover:border-primary/50">
        <CardContent className="p-4">
          <h3 className="font-display font-semibold text-lg line-clamp-2 mb-1">
            {book.title}
          </h3>
          <p className="text-sm text-muted-foreground mb-2">
            {book.author}
          </p>
          <div className="flex items-center gap-2">
            <StarRating 
              value={book.average_rating || 0} 
              readonly 
              size="sm" 
            />
            <span className="text-xs text-muted-foreground">
              {book.reviews_count}
            </span>
          </div>
          {book.published_year && (
            <p className="text-xs text-muted-foreground mt-2">
              {book.published_year}
            </p>
          )}
        </CardContent>
      </Card>
    </Link>
  );
}



