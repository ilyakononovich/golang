import { Star } from 'lucide-react';
import { cn } from '@/lib/utils';

interface StarRatingProps {
  value: number;
  onChange?: (value: number) => void;
  readonly?: boolean;
  size?: 'sm' | 'md' | 'lg';
}

const sizeClasses = {
  sm: 'h-3 w-3',
  md: 'h-4 w-4',
  lg: 'h-5 w-5',
};

export function StarRating({ 
  value, 
  onChange, 
  readonly = false,
  size = 'md' 
}: StarRatingProps) {
  const handleClick = (rating: number) => {
    if (!readonly && onChange) {
      onChange(rating);
    }
  };

  return (
    <div className={cn("star-rating flex gap-0.5", !readonly && "cursor-pointer")}>
      {[1, 2, 3, 4, 5].map((star) => (
        <Star
          key={star}
          className={cn(
            sizeClasses[size],
            "transition-colors",
            star <= value 
              ? "fill-primary text-primary" 
              : "fill-transparent text-muted-foreground"
          )}
          onClick={() => handleClick(star)}
        />
      ))}
    </div>
  );
}





