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
        auth_login_subtitle: 'Введите корректные данные ниже для входа.',
        auth_register_title: 'Регистрация',
        auth_register_subtitle: 'Заполните все поля ниже для успешной регистрации.',
        label_name: 'Имя',
        label_email: 'Email',
        label_password: 'Пароль',
        button_login: 'Войти',
        button_register: 'Создать аккаунт',
        language_mode: 'Выберите язык'
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
        button_register: 'Create account',
        language_mode: 'Select language'
    }
};

function setLanguage(lang) {
    localStorage.setItem('lang', lang);
    document.querySelectorAll('[data-lang]').forEach((el) => {
        const key = el.getAttribute('data-lang');

        if (translations[lang][key]) {
            el.textContent = translations[lang][key];
        }
    });

    const languageSelect = document.getElementById('language-select');
    if (languageSelect) {
        languageSelect.value = lang;
    }
}

const savedLanguage = localStorage.getItem('lang') || 'ru';

function initLanguage() {

    setLanguage(savedLanguage);

    const languageSelect = document.getElementById('language-select');
    if (languageSelect) {
        languageSelect.addEventListener('change', (event) => {
            setLanguage(event.target.value);
        });
    }
}

document.addEventListener('DOMContentLoaded', initLanguage);

export const t = (key) => translations[savedLanguage][key] || key;
