## Структура

- `pages/` - HTML-страницы
- `assets/css/` - стили
- `assets/js/pages/` - логика конкретных страниц
- `assets/js/api/` - подключениие интерфейса и бэкенда, обработка некоторых данных с запросов
- `assets/js/ui/` - общие UI-компоненты, модалки, локальное состояние и UI-утилиты

## Страницы

- `pages/index.html` - главная
- `pages/analytics.html` - аналитика расходов
- `pages/groups.html` - группы подписок
- `pages/profile.html` - профиль пользователя

## Используемые API

- `GET /subscriptions`
- `GET /analytics`
- `GET /groups`
- `GET /profile`
- `POST /auth/register`
- `POST /auth/login`
- `POST /subscriptions`
- `PUT /subscriptions/{subscriptions_id}`
- `DELETE /subscriptions/{subscriptions_id}`
- `POST /groups`
- `POST /groups/join`
