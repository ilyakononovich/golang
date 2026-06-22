import { BookOpen } from "lucide-react";

export function Footer() {
  return (
    <footer className="border-t py-6 mt-auto">
      <div className="container flex flex-col sm:flex-row items-center justify-between gap-4">
        <div className="flex items-center gap-2 text-muted-foreground">
          <BookOpen className="h-4 w-4" />
          <span className="text-sm">Bookshelf by Praxis © 2026</span>
        </div>
        <p className="text-sm text-muted-foreground">
          Проект 1: Clean Architecture Monolith
        </p>
      </div>
    </footer>
  );
}
