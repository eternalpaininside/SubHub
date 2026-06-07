const translations = {
  ru: {
    brand: 'SubHub',
    nav_home: 'Главная',
    nav_analytics: 'Аналитика',
    nav_groups: 'Группы',
    auth_open_profile: 'Открыть профиль или окно входа и регистрации',
    auth_login_tab: 'Вход',
    auth_register_tab: 'Регистрация',
    auth_login_title: 'Вход в аккаунт',
    auth_login_subtitle: 'Вход в систему с сохранением текущего пользователя в браузере.',
    auth_register_title: 'Регистрация',
    auth_register_subtitle: 'Минимальная регистрация пользователя для личного кабинета и подписок.',
    label_name: 'Имя',
    label_email: 'Email',
    label_password: 'Пароль',
    button_login: 'Войти',
    button_register: 'Создать аккаунт'
  },
  en: {
    brand: 'SubHub',
    nav_home: 'Home',
    nav_analytics: 'Analytics',
    nav_groups: 'Groups',
    auth_open_profile: 'Open profile or authentication modal',
    auth_login_tab: 'Login',
    auth_register_tab: 'Register',
    auth_login_title: 'Sign in',
    auth_login_subtitle: 'Sign in stores the current user in the browser.',
    auth_register_title: 'Register',
    auth_register_subtitle: 'Minimal registration for profile and subscription tracking.',
    label_name: 'Name',
    label_email: 'Email',
    label_password: 'Password',
    button_login: 'Sign in',
    button_register: 'Create account'
  }
};

export const t = (key) => translations.ru[key] || key;
