import { api } from '../api/backend-api-client.js';
import { API_BASE_URL } from '../api/runtime-config.js';
import { buildLayout, initLayoutUI } from '../ui/app-layout.js';
import { buildHomeDashboardGrid, initFamilyGroupSlider } from '../ui/home-page-sections.js';
import { buildSubscriptionsSectionMarkup, createSubscriptionsSectionController } from '../ui/subscriptions-list-section.js';

const fallbackHome = {
  activeSubscriptions: '0',
  expiringSoon: '0',
  familyGroupTitle: 'Мои группы',
  familyGroups: [],
  monthLabel: 'текущий месяц',
  monthExpensesLabel: '0 ₽',
  monthBudgetLabel: '0 ₽',
  budgetProgress: 0,
  trendDirection: 'up',
  trendPercentLabel: '0%',
  trendLabel: 'после входа появится живая аналитика',
  topCategories: []
};

let subscriptionsSectionController = null;

const getHomeLoadIssue = (error) => {
  if (!error) return null;

  if (error.message === 'Требуется вход в систему') {
    return {
      type: 'auth',
      title: 'Данные не загружены: нужен вход в аккаунт',
      description: 'Главная страница показывает реальные подписки, ' +
          'расходы и группы только для авторизованного пользователя.',
      actionLabel: 'Открыть вход',
      actionClassName: 'primary-btn js-open-auth'
    };
  }

  if (error.name === 'AbortError' || error.message === 'Failed to fetch') {
    return {
      type: 'backend',
      title: 'Данные не загружены: backend недоступен',
      description: `Фронтенд не смог получить ответ от API на ${API_BASE_URL}.
       Проверь, что Go-сервер запущен и адрес API настроен корректно.`,
      actionLabel: 'Открыть профиль',
      actionClassName: 'ghost-btn'
    };
  }

  return {
    type: 'generic',
    title: 'Данные не загружены',
    description: error.message || 'Во время загрузки главной страницы произошла ошибка.',
    actionLabel: 'Открыть профиль',
    actionClassName: 'ghost-btn'
  };
};

const buildHomeLoadIssueMarkup = (issue) => {
  if (!issue) return '';

  const actionHref = issue.type === 'auth' ? '' : 'index.html';
  const actionMarkup = issue.type === 'auth'
    ? `<button class="${issue.actionClassName}" type="button">${issue.actionLabel}</button>`
    : `<a class="${issue.actionClassName}" href="${actionHref}">${issue.actionLabel}</a>`;

  return `
    <section class="card home-status-card home-status-card-${issue.type}" aria-live="polite">
      <div class="home-status-kicker">Проблема загрузки</div>
      <h2 class="home-status-title">${issue.title}</h2>
      <p class="home-status-text">${issue.description}</p>
      <div class="home-status-actions">${actionMarkup}</div>
    </section>
  `;
};

const toggleSectionReveal = (section) => {
  const rect = section.getBoundingClientRect();
  section.classList.toggle('is-visible', rect.top <= window.innerHeight * 0.9
      && rect.bottom >= window.innerHeight * 0.15);
};

const initHomeScrollReveal = () => {
  const heroGrid = document.querySelector('.dashboard-grid');
  const subscriptionsSection = document.querySelector('.home-subscriptions-section');
  if (!heroGrid || !subscriptionsSection) return;

  const syncFoldOffset = () => {
    subscriptionsSection.style.marginTop = '0px';
    const offset = Math.max(0,
        Math.ceil(window.innerHeight - heroGrid.getBoundingClientRect().bottom + 12));
    subscriptionsSection.style.marginTop = `${offset}px`;
  };

  syncFoldOffset();
  window.addEventListener('resize', syncFoldOffset, { passive: true });

  if ('IntersectionObserver' in window) {
    const observer = new IntersectionObserver((entries) => {
      entries.forEach((entry) => {
        subscriptionsSection.classList.toggle('is-visible',
            entry.isIntersecting && entry.intersectionRatio > 0.1);
      });
    }, { threshold: [0, 0.1, 0.2], rootMargin: '0px 0px -4% 0px' });
    observer.observe(subscriptionsSection);
    return;
  }

  const onScroll = () => toggleSectionReveal(subscriptionsSection);
  window.addEventListener('scroll', onScroll, { passive: true });
  onScroll();
};

const initHomeScrollHint = () => {
  document.querySelector('.scroll-hint-home')?.remove();

  const hint = document.createElement('button');
  hint.className = 'scroll-hint-home';
  hint.type = 'button';
  hint.setAttribute('aria-label', 'Прокрутить страницу вниз');
  hint.innerHTML = '<span>⌄</span>';
  document.body.append(hint);

  const toggleHint = () => {
    hint.classList.toggle('is-visible', window.scrollY < 100);
  };

  hint.addEventListener('click', () => {
    window.scrollBy({ top: Math.max(window.innerHeight * 0.6, 360),
      behavior: 'smooth' });
  });

  ['scroll', 'resize'].forEach((eventName) =>
      window.addEventListener(eventName, toggleHint, { passive: true }));
  window.addEventListener('load', toggleHint, { once: true });
  toggleHint();
  requestAnimationFrame(toggleHint);
  window.setTimeout(toggleHint, 120);
};

async function renderHome() {
  let home;
  let loadIssue = null;
  try {
    home = await api.getHomePage();
  } catch (error) {
    home = fallbackHome;
    loadIssue = getHomeLoadIssue(error);
  }

  document.body.innerHTML = buildLayout('home', `
    <main class="home-main">
      ${buildHomeLoadIssueMarkup(loadIssue)}
      ${buildHomeDashboardGrid(home)}
      ${buildSubscriptionsSectionMarkup({
        sectionClass: 'home-subscriptions-section',
        titleTag: 'h2',
        titleClass: 'section-title',
        subtitleClass: 'page-subtitle',
        actionButtonMarkup: '<button class="primary-btn js-open-add-subscription">＋ Добавить</button>'
      })}
    </main>
  `);

  initLayoutUI();
  subscriptionsSectionController = createSubscriptionsSectionController(document);
  await subscriptionsSectionController.init();
  initHomeScrollReveal();
  initHomeScrollHint();
  initFamilyGroupSlider();
}

window.addEventListener('subscription:changed', () => subscriptionsSectionController?.loadSubscriptions());

renderHome();
