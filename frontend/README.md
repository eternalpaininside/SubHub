## Структура

- `pages/` - HTML-страницы
- `assets/css/` - стили
- `assets/js/pages/` - логика конкретных страниц
- `assets/js/api/` - слой доступа к backend API
- `assets/js/ui/` - общие UI-компоненты, модалки, локальное состояние и UI-утилиты

## Страницы

- `pages/index.html` - главная
- `pages/analytics.html` - аналитика расходов
- `pages/groups.html` - группы подписок
- `pages/profile.html` - профиль пользователя

## Используемые API

- `GET /api/subscriptions`
- `GET /api/analytics`
- `GET /api/groups`
- `GET /api/profile`
- `POST /api/auth/register`
- `POST /api/auth/login`
- `POST /api/subscriptions`
- `PUT /api/subscriptions/{id}`
- `DELETE /api/subscriptions/{id}`
- `POST /api/groups`
- `POST /api/groups/join`
