import { Lock, BookOpen, ArrowRight } from 'lucide-react';
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from './card';
import { Button } from './button';

interface FeatureLockedProps {
  title: string;
  description: string;
  stage: number;
  hint?: string;
  icon?: React.ReactNode;
}

export function FeatureLocked({ 
  title, 
  description, 
  stage, 
  hint,
  icon 
}: FeatureLockedProps) {
  return (
    <Card className="border-dashed border-2 border-muted-foreground/25 bg-muted/5">
      <CardHeader className="text-center pb-2">
        <div className="mx-auto mb-4 flex h-16 w-16 items-center justify-center rounded-full bg-muted">
          {icon || <Lock className="h-8 w-8 text-muted-foreground" />}
        </div>
        <CardTitle className="text-xl">{title}</CardTitle>
        <CardDescription className="text-base">
          {description}
        </CardDescription>
      </CardHeader>
      <CardContent className="text-center space-y-4">
        <div className="inline-flex items-center gap-2 bg-primary/10 text-primary px-4 py-2 rounded-full text-sm font-medium">
          <BookOpen className="h-4 w-4" />
          <span>Смотри Этап {stage} в DETAILED_STAGES.md</span>
        </div>
        
        {hint && (
          <p className="text-sm text-muted-foreground max-w-md mx-auto">
            💡 {hint}
          </p>
        )}

        <div className="pt-4">
          <Button variant="outline" disabled className="gap-2">
            <span>Скоро будет доступно</span>
            <ArrowRight className="h-4 w-4" />
          </Button>
        </div>
      </CardContent>
    </Card>
  );
}

interface FeatureErrorProps {
  title: string;
  error?: Error | null;
  onRetry?: () => void;
}

export function FeatureError({ title, error, onRetry }: FeatureErrorProps) {
  const isNetworkError = error?.message?.includes('Network Error') || 
                         error?.message?.includes('ERR_CONNECTION_REFUSED');
  
  return (
    <Card className="border-destructive/50 bg-destructive/5">
      <CardHeader className="text-center pb-2">
        <div className="mx-auto mb-4 flex h-16 w-16 items-center justify-center rounded-full bg-destructive/10">
          <span className="text-3xl">⚠️</span>
        </div>
        <CardTitle className="text-xl text-destructive">{title}</CardTitle>
        <CardDescription className="text-base">
          {isNetworkError 
            ? 'Не удалось подключиться к серверу. Убедись, что backend запущен.'
            : 'Произошла ошибка при загрузке данных.'
          }
        </CardDescription>
      </CardHeader>
      <CardContent className="text-center space-y-4">
        {isNetworkError && (
          <div className="bg-muted p-4 rounded-lg text-left max-w-md mx-auto">
            <p className="text-sm font-mono text-muted-foreground">
              $ go run ./cmd/server
            </p>
          </div>
        )}
        
        {onRetry && (
          <Button onClick={onRetry} variant="outline">
            Попробовать снова
          </Button>
        )}
      </CardContent>
    </Card>
  );
}
