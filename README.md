# МедДок — Система документооборота IT-заявок

Веб-приложение для управления IT-заявками в медицинском учреждении.

## Стек
- **Backend**: Go 1.22, Gin, pgx, JWT, cron
- **Frontend**: Vue 3, Vite, Pinia, Vue Router, Axios
- **БД**: PostgreSQL 16
- **Деплой**: Docker Compose + Nginx

---

## Быстрый старт (Docker)

```bash
# 1. Клонируй репозиторий
git clone <repo> && cd meddoc

# 2. Запусти через Docker Compose
docker-compose up --build

# Фронтенд: http://localhost:3000
# API:       http://localhost:8080/api
```

---

## Локальная разработка

### База данных
```bash
docker run -d \
  --name meddoc-pg \
  -e POSTGRES_USER=meddoc_user \
  -e POSTGRES_PASSWORD=supersecret \
  -e POSTGRES_DB=meddoc_db \
  -p 5432:5432 \
  postgres:16-alpine
```

### Backend
```bash
cd backend
cp .env.example .env
# Отредактируй .env при необходимости
go mod tidy
go run ./cmd/server
# Сервер запустится на :8080
```

### Frontend
```bash
cd frontend
npm install
npm run dev
# Откроется на http://localhost:5173
```

---

## Создание первого администратора (SQL)

```sql
-- Подключись к БД и выполни:
INSERT INTO admin_users (login, password_hash)
VALUES (
  'admin',
  '$2a$10$...'  -- bcrypt hash вашего пароля
);
```

Или используй вспомогательный скрипт:
```bash
cd backend
go run ./cmd/create_admin -login admin -password MyPassword123 -secret admin_secret_8ch
```

---

## Структура проекта

```
meddoc/
├── backend/
│   ├── cmd/server/main.go          # Точка входа
│   ├── internal/
│   │   ├── auth/jwt.go             # JWT токены
│   │   ├── handlers/               # HTTP обработчики
│   │   │   ├── auth.go             # Регистрация/вход
│   │   │   └── tickets.go          # CRUD заявок
│   │   ├── middleware/auth.go      # JWT middleware
│   │   ├── models/models.go        # Модели данных
│   │   ├── repository/             # Работа с БД
│   │   │   ├── db.go              # Подключение + миграции
│   │   │   ├── user.go            # Пользователи
│   │   │   └── ticket.go          # Заявки
│   │   └── scheduler/scheduler.go  # Cron: автоэскалация
│   ├── Dockerfile
│   └── go.mod
│
└── frontend/
    ├── src/
    │   ├── api/index.js            # Axios клиент
    │   ├── stores/auth.js          # Pinia: авторизация
    │   ├── router/index.js         # Маршрутизация
    │   ├── views/
    │   │   ├── LoginView.vue       # Страница входа (3 формы)
    │   │   ├── DashboardView.vue   # Кабинет пользователя
    │   │   ├── NewTicketView.vue   # Подача заявления
    │   │   ├── EditTicketView.vue  # Редактирование
    │   │   ├── ProfileView.vue     # Личный кабинет
    │   │   └── AdminView.vue       # Панель администратора
    │   ├── components/shared/
    │   │   └── AppNav.vue          # Навигация
    │   └── assets/styles/main.css  # Глобальные стили
    ├── Dockerfile
    ├── nginx.conf
    └── vite.config.js
```

---

## API эндпоинты

### Авторизация (публичные)
| Метод | URL | Описание |
|-------|-----|----------|
| POST | /api/auth/register | Регистрация пользователя |
| POST | /api/auth/login | Вход пользователя |
| POST | /api/auth/admin/login | Вход администратора |

### Пользователь (JWT required)
| Метод | URL | Описание |
|-------|-----|----------|
| GET | /api/user/profile | Получить профиль |
| PUT | /api/user/profile | Обновить профиль |
| GET | /api/tickets | Список своих заявок |
| POST | /api/tickets | Создать заявку |
| GET | /api/tickets/:id | Получить заявку |
| PUT | /api/tickets/:id | Изменить заявку |
| PATCH | /api/tickets/:id/cancel | Отменить |
| PATCH | /api/tickets/:id/close | Закрыть |

### Администратор (JWT + role=admin)
| Метод | URL | Описание |
|-------|-----|----------|
| GET | /api/admin/tickets | Все заявки с фильтрами |
| DELETE | /api/admin/tickets/:id | Удалить заявку |
| PATCH | /api/admin/tickets/:id/close | Закрыть заявку |
| POST | /api/admin/tickets/:id/comment | Добавить комментарий |
| GET | /api/admin/tickets/export?format=xml | Экспорт XML |
| GET | /api/admin/ip-logs | Журнал входов |

### Параметры фильтрации (/api/admin/tickets)
- `division` — подразделение
- `priority` — low / medium / high
- `status` — open / in_progress / closed / cancelled
- `date_from`, `date_to` — YYYY-MM-DD
- `sort_by` — created_at / priority / division
- `sort_order` — ASC / DESC

---

## Автоэскалация приоритета

Cron-задача запускается **ежедневно в 08:00**.
Заявления со статусом `open` или `in_progress`, не имеющие приоритета `high`
и созданные более **7 дней** назад (≈ 5 рабочих дней), автоматически
получают `priority = high` и флаг `auto_escalated = true`.
