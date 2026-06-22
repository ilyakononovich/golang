# Project 1: Монолит — Инфраструктура

## О проекте

В этом проекте вы создадите **монолитное REST API** для платформы книжных рецензий "Bookshelf".

**Что вы реализуете:**
- Регистрация и авторизация пользователей (JWT)
- CRUD операции для книг
- Система рецензий с рейтингами
- Clean Architecture (Handler → Service → Repository)

**Архитектура:**
```
Frontend (:5173) → Backend (:8080) → PostgreSQL (:5432)
```

## Содержимое

```
├── frontend/           # React-приложение (готовое)
├── migrations/         # SQL-схема и тестовые данные
├── docker-compose.yml  # PostgreSQL + Frontend
└── README.md           # Этот файл
```

## Запуск

```bash
docker compose up -d --build
```

После запуска:
- **Frontend**: http://localhost:5173
- **PostgreSQL**: localhost:5432

## Как это работает

Frontend уже готов и ждёт ваш backend. По мере реализации API — функции будут "оживать":

1. Реализовали `/health` → frontend подключился
2. Реализовали авторизацию → появилась возможность входа
3. Реализовали CRUD книг → появился каталог
4. Реализовали рецензии → можно оставлять отзывы

Недоступные функции показываются как "заблокированные" с подсказкой, на каком этапе они станут доступны.

## Подключение к базе данных

```
postgres://postgres:postgres@localhost:5432/bookshelf?sslmode=disable
```

| Параметр | Значение |
|----------|----------|
| Host | localhost |
| Port | 5432 |
| Database | bookshelf |
| User | postgres |
| Password | postgres |

## Тестовые пользователи

| Email | Password |
|-------|----------|
| admin@bookshelf.dev | password123 |
| john@example.com | password123 |
| maria@example.com | password123 |

## Команды

```bash
docker compose up -d --build  # Запустить
docker compose down           # Остановить
docker compose logs -f        # Логи
docker compose down -v        # Удалить всё (включая данные)
```

## Устранение неполадок

### Docker не запускается
```bash
docker ps  # Проверьте, что Docker daemon работает
```

### Frontend не видит backend
1. Убедитесь, что ваш Go-сервер запущен на порту 8080
2. Проверьте CORS middleware в вашем коде
3. Откройте DevTools → Network для диагностики

### Ошибка подключения к БД
```bash
docker compose ps             # Статус контейнеров
docker compose logs postgres  # Логи PostgreSQL
```

### Пересборка frontend
```bash
docker compose build --no-cache frontend
docker compose up -d frontend
```

## Инструкции по реализации

Подробное описание каждого этапа находится на странице курса:

**https://praxiscode.io**
