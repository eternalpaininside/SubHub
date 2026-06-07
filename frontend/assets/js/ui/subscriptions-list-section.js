import { api } from '../api/backend-api-client.js';
import { SUBSCRIPTION_FILTERS } from './constants.js';
import { subscriptionListCard } from './subscription-group-cards.js';
import { subscriptionsStore } from './ui-store.js';

const emptyStateMarkup = '<article class="card">Нет подписок по выбранному фильтру.</article>';
const loadingMarkup = '<article class="card">Загрузка подписок...</article>';

const buildApiFilters = () => {
  const { activeFilter } = subscriptionsStore.getState();
  if (activeFilter === 'Все') return {};
  if (activeFilter === 'Групповой тариф') return { planType: 'Групповой' };
  if (activeFilter === 'Индивидуальный тариф') return { planType: 'Индивидуальный' };
  return { category: activeFilter };
};

const renderSubscriptionsList = (items) => items.length
  ? items.map(subscriptionListCard).join('')
  : emptyStateMarkup;

const buildSubscriptionFiltersMarkup = (activeFilter) => SUBSCRIPTION_FILTERS.map((filter) => `
  <button
    class="filter-chip ${filter === activeFilter ? 'is-active' : ''}"
    type="button"
    data-filter="${filter}"
  >${filter}</button>
`).join('');

const buildSubscriptionsSearchSection = (activeFilter) => `
  <section class="search-row" style="margin-bottom:18px;">
    <div class="filters-scroll-wrap">
      <div class="filters-row" data-filters-scroll>
        ${buildSubscriptionFiltersMarkup(activeFilter)}
      </div>
      <button class="ghost-btn filters-scroll-arrow" type="button" aria-label="Прокрутить фильтры" data-filters-scroll-next>
        <span aria-hidden="true">›</span>
      </button>
    </div>
  </section>
`;

export const buildSubscriptionsSectionMarkup = ({
  sectionClass = '',
  title = 'Мои подписки',
  titleTag = 'h1',
  titleClass = 'page-title',
  subtitle = 'Загрузка...',
  subtitleClass = 'page-subtitle',
  actionButtonMarkup = ''
} = {}) => `
  <section${sectionClass ? ` class="${sectionClass}"` : ''}>
    <div class="page-header">
      <div>
        <${titleTag} class="${titleClass}">${title}</${titleTag}>
        <p class="${subtitleClass} js-subscriptions-subtitle">${subtitle}</p>
      </div>
      ${actionButtonMarkup}
    </div>
    ${buildSubscriptionsSearchSection(subscriptionsStore.getState().activeFilter)}
    <section class="list-stack js-subscriptions-list">${loadingMarkup}</section>
  </section>
`;

export const createSubscriptionsSectionController = (root = document) => {
  const getSubtitleNode = () => root.querySelector('.js-subscriptions-subtitle');
  const getListNode = () => root.querySelector('.js-subscriptions-list');

  const updateSubscriptionsUI = (viewModel) => {
    const subtitleNode = getSubtitleNode();
    const listNode = getListNode();
    if (subtitleNode) subtitleNode.textContent = viewModel.subtitle;
    if (listNode) listNode.innerHTML = renderSubscriptionsList(viewModel.items);
  };

  const setActiveFilterChip = () => {
    const { activeFilter } = subscriptionsStore.getState();
    root.querySelectorAll('[data-filter]').forEach((chip) => {
      chip.classList.toggle('is-active', chip.dataset.filter === activeFilter);
    });
  };

  const initFiltersScroll = () => {
    const scrollNode = root.querySelector('[data-filters-scroll]');
    const nextButton = root.querySelector('[data-filters-scroll-next]');
    if (!scrollNode || !nextButton) return;

    nextButton.addEventListener('click', () => {
      scrollNode.scrollBy({ left: Math.max(220, Math.round(scrollNode.clientWidth * 0.55)), behavior: 'smooth' });
    });
  };

  const loadSubscriptions = async () => {
    try {
      updateSubscriptionsUI(await api.getSubscriptionsPage(buildApiFilters()));
    } catch (error) {
      const listNode = getListNode();
      if (listNode) {
        listNode.innerHTML = `<article class="card">${error.message || 'Не удалось загрузить подписки. Попробуйте позже.'}</article>`;
      }
    }
  };

  const initFilterActions = () => {
    root.querySelectorAll('[data-filter]').forEach((chip) => {
      chip.addEventListener('click', async () => {
        const nextFilter = chip.dataset.filter || 'Все';
        if (nextFilter === subscriptionsStore.getState().activeFilter) return;
        subscriptionsStore.setState({ activeFilter: nextFilter });
        setActiveFilterChip();
        await loadSubscriptions();
      });
    });
  };

  return {
    init: async () => {
      initFiltersScroll();
      initFilterActions();
      await loadSubscriptions();
    },
    loadSubscriptions
  };
};
