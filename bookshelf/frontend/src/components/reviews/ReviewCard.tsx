import { useState } from 'react';
import { Edit, Trash2, X, Check } from 'lucide-react';
import { formatDistanceToNow } from '@/lib/date';
import { Card, CardContent, CardHeader } from '@/components/ui/card';
import { Avatar, AvatarFallback } from '@/components/ui/avatar';
import { Button } from '@/components/ui/button';
import { Input } from '@/components/ui/input';
import { Textarea } from '@/components/ui/textarea';
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle,
  DialogTrigger,
} from '@/components/ui/dialog';
import { StarRating } from './StarRating';
import { useUpdateReview, useDeleteReview } from '@/api/reviews';
import type { Review } from '@/types/api';

interface ReviewCardProps {
  review: Review;
  bookId: string;
  currentUserId?: string;
}

export function ReviewCard({ review, bookId, currentUserId }: ReviewCardProps) {
  const [isEditing, setIsEditing] = useState(false);
  const [editRating, setEditRating] = useState(review.rating);
  const [editTitle, setEditTitle] = useState(review.title || '');
  const [editContent, setEditContent] = useState(review.content);
  const [deleteOpen, setDeleteOpen] = useState(false);

  const updateReview = useUpdateReview(review.id, bookId);
  const deleteReview = useDeleteReview(bookId);

  const isOwner = currentUserId && currentUserId === review.user_id;

  const initials = review.user.username
    .split(' ')
    .map((n) => n[0])
    .join('')
    .toUpperCase()
    .slice(0, 2);

  const handleStartEdit = () => {
    setEditRating(review.rating);
    setEditTitle(review.title || '');
    setEditContent(review.content);
    setIsEditing(true);
  };

  const handleCancelEdit = () => {
    setIsEditing(false);
  };

  const handleSaveEdit = async () => {
    await updateReview.mutateAsync({
      rating: editRating,
      title: editTitle || undefined,
      content: editContent,
    });
    setIsEditing(false);
  };

  const handleDelete = async () => {
    await deleteReview.mutateAsync(review.id);
    setDeleteOpen(false);
  };

  if (isEditing) {
    return (
      <Card>
        <CardHeader className="pb-3">
          <div className="flex items-start gap-3">
            <Avatar className="h-9 w-9">
              <AvatarFallback className="bg-primary/10 text-primary text-sm">
                {initials}
              </AvatarFallback>
            </Avatar>
            <div className="flex-1 min-w-0 space-y-3">
              <div className="flex items-center justify-between gap-2">
                <p className="font-medium text-sm truncate">
                  {review.user.username}
                </p>
                <div className="flex gap-1">
                  <Button
                    variant="ghost"
                    size="icon"
                    className="h-7 w-7"
                    onClick={handleCancelEdit}
                  >
                    <X className="h-4 w-4" />
                  </Button>
                  <Button
                    variant="ghost"
                    size="icon"
                    className="h-7 w-7 text-primary"
                    onClick={handleSaveEdit}
                    disabled={updateReview.isPending || editContent.length < 10}
                  >
                    <Check className="h-4 w-4" />
                  </Button>
                </div>
              </div>
              <StarRating
                value={editRating}
                onChange={setEditRating}
                size="sm"
              />
            </div>
          </div>
        </CardHeader>
        <CardContent className="pt-0 space-y-3">
          <Input
            placeholder="Заголовок (необязательно)"
            value={editTitle}
            onChange={(e) => setEditTitle(e.target.value)}
          />
          <Textarea
            placeholder="Рецензия (минимум 10 символов)"
            value={editContent}
            onChange={(e) => setEditContent(e.target.value)}
            className="min-h-[80px]"
          />
        </CardContent>
      </Card>
    );
  }

  return (
    <Card>
      <CardHeader className="pb-3">
        <div className="flex items-start gap-3">
          <Avatar className="h-9 w-9">
            <AvatarFallback className="bg-primary/10 text-primary text-sm">
              {initials}
            </AvatarFallback>
          </Avatar>
          <div className="flex-1 min-w-0">
            <div className="flex items-center justify-between gap-2">
              <p className="font-medium text-sm truncate">
                {review.user.username}
              </p>
              <div className="flex items-center gap-1">
                {isOwner && (
                  <>
                    <Button
                      variant="ghost"
                      size="icon"
                      className="h-7 w-7"
                      onClick={handleStartEdit}
                    >
                      <Edit className="h-3.5 w-3.5" />
                    </Button>
                    <Dialog open={deleteOpen} onOpenChange={setDeleteOpen}>
                      <DialogTrigger asChild>
                        <Button variant="ghost" size="icon" className="h-7 w-7">
                          <Trash2 className="h-3.5 w-3.5" />
                        </Button>
                      </DialogTrigger>
                      <DialogContent>
                        <DialogHeader>
                          <DialogTitle>Удалить рецензию?</DialogTitle>
                          <DialogDescription>
                            Это действие нельзя отменить.
                          </DialogDescription>
                        </DialogHeader>
                        <DialogFooter>
                          <Button
                            variant="outline"
                            onClick={() => setDeleteOpen(false)}
                          >
                            Отмена
                          </Button>
                          <Button
                            variant="destructive"
                            onClick={handleDelete}
                            disabled={deleteReview.isPending}
                          >
                            {deleteReview.isPending ? 'Удаление...' : 'Удалить'}
                          </Button>
                        </DialogFooter>
                      </DialogContent>
                    </Dialog>
                  </>
                )}
                <time className="text-xs text-muted-foreground whitespace-nowrap ml-1">
                  {formatDistanceToNow(review.created_at)}
                </time>
              </div>
            </div>
            <StarRating value={review.rating} readonly size="sm" />
          </div>
        </div>
      </CardHeader>
      <CardContent className="pt-0">
        {review.title && (
          <h4 className="font-medium mb-1">{review.title}</h4>
        )}
        <p className="text-sm text-muted-foreground whitespace-pre-wrap">
          {review.content}
        </p>
      </CardContent>
    </Card>
  );
}
