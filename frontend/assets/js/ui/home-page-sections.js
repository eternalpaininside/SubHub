import { escapeAttr } from './formatters.js';

const buildFamilyMemberRow = (memberName) => {
  const initial = memberName === 'Участников пока нет'
    ? '•'
    : (memberName.trim().charAt(0).toUpperCase() || '•');

  return `
    <div class="family-member">
      <span class="family-badge">${initial}</span>
      <span class="family-member-name">${memberName}</span>
    </div>
  `;
};

const buildFamilyGroupWindow = (group, index) => {
  const members = Array.isArray(group.members) && group.members.length
    ? group.members.map((member) => member.name || 'Участник')
    : ['Участников пока нет'];

  const membersList = members.map(buildFamilyMemberRow).join('');
  const groupNotes = group.notes || (Array.isArray(group.members)
    ? group.members.map((member) => member.name || '').filter(Boolean).join(', ')
    : '');

  return `
    <article class="family-group-window ${index === 0 ? 'is-active' : ''}" data-group-index="${index}">
      <div class="family-group-window-head">
        <div class="family-group-window-title">${group.name}</div>
        <button
          class="ghost-btn family-group-settings-btn js-open-edit-group"
          type="button"
          aria-label="Изменить группу"
          data-group-id="${escapeAttr(group.id || '')}"
          data-group-name="${escapeAttr(group.name)}"
          data-group-type="${escapeAttr(group.type || 'Семейная')}"
          data-group-members-count="${escapeAttr(group.members?.length || '')}"
          data-group-budget="${escapeAttr(group.price || '')}"
          data-group-period="${escapeAttr(group.period || 'мес')}"
          data-group-invite="${escapeAttr(group.inviteUrl || '')}"
          data-group-notes="${escapeAttr(groupNotes)}"
          data-group-subscription-ids="${escapeAttr((group.subscriptionIds || []).join(','))}"
        ><span class="mirrored-edit-icon">✎</span></button>
      </div>
      <div class="family-badges">${membersList}</div>
    </article>
  `;
};

export const buildHomeDashboardGrid = (home) => {
  return `
    <section class="dashboard-grid">
      <section class="subs-pair" aria-label="Активные и заканчивающиеся подписки">
        <article class="card subs-active-card">
          <div class="subs-active-title">Активные подписки</div>
          <div class="subs-active-value">${home.activeSubscriptions}</div>
        </article>
        <div class="subs-active-tail">
          <button class="ghost-btn add-subscription-btn js-open-add-subscription" type="button">Добавить подписку</button>
        </div>
        <article class="subs-expiring-card">
          <div class="value">${home.expiringSoon}</div>
          <div class="label">заканчиваются</div>
        </article>
      </section>

      <section class="home-left-stack" aria-label="Расходы и краткая аналитика">
        <article class="card home-expenses-card">
          <div class="card-title">Расходы за ${home.monthLabel}</div>
          <div class="metric-value">${home.monthExpensesLabel}</div>
        </article>

        <article class="card quick-analytics-card">
          <div class="service-name">Краткая аналитика</div>
          <div class="quick-analytics-hero">
            <div class="quick-analytics-trend">
              <span class="quick-analytics-trend-arrow quick-analytics-trend-arrow-${home.trendDirection}">${home.trendDirection === 'up' ? '↗' : '↘'}</span>
              <div>
                <div class="quick-analytics-trend-value">${home.trendPercentLabel}</div>
                <div class="quick-analytics-trend-label">${home.trendLabel}</div>
              </div>
            </div>
          </div>
          <div class="quick-analytics-categories">
            <div class="quick-analytics-categories-title">Лидеры по затратам</div>
            ${home.topCategories.length ? home.topCategories.map((category) => `
                <div class="quick-analytics-category-row">
                  <span class="quick-analytics-category-name"><span class="quick-analytics-dot" style="background:${category.color}"></span>${category.name}</span>
                  <span class="quick-analytics-category-meta">${category.share}% • ${category.amountLabel}</span>
                </div>
              `).join('') : '<div class="quick-analytics-empty">Нет данных по категориям</div>'}
          </div>
        </article>
      </section>

      <article class="card family-card">
        <div class="service-name">${home.familyGroupTitle}</div>
        <div class="family-group-slider" data-group-slider>
          <div class="family-group-window-wrap">
            ${home.familyGroups.map(buildFamilyGroupWindow).join('')}
            <button class="ghost-btn family-slider-arrow family-slider-arrow-left" type="button" data-group-nav="prev" aria-label="Предыдущая группа">‹</button>
            <button class="ghost-btn family-slider-arrow family-slider-arrow-right" type="button" data-group-nav="next" aria-label="Следующая группа">›</button>
          </div>
        </div>
        <button class="ghost-btn add-group-btn js-open-add-group" type="button">Добавить группу</button>
      </article>
    </section>
  `;
};

export const initFamilyGroupSlider = () => {
  const slider = document.querySelector('[data-group-slider]');
  if (!slider) return;

  const windows = Array.from(slider.querySelectorAll('.family-group-window'));
  if (!windows.length) return;

  const navButtons = Array.from(slider.querySelectorAll('[data-group-nav]'));
  const animationDuration = 340;
  let activeIndex = Math.max(0, windows.findIndex((node) => node.classList.contains('is-active')));
  let isAnimating = false;

  if (windows.length < 2) {
    navButtons.forEach((button) => {
      button.style.visibility = 'hidden';
      button.setAttribute('aria-hidden', 'true');
      button.tabIndex = -1;
    });
  }

  const renderActiveWindow = () => {
    windows.forEach((windowNode, index) => {
      windowNode.classList.toggle('is-active', index === activeIndex);
    });
  };

  navButtons.forEach((button) => {
    button.addEventListener('click', () => {
      if (isAnimating || windows.length < 2) return;
      isAnimating = true;

      activeIndex = button.dataset.groupNav === 'next'
        ? (activeIndex + 1) % windows.length
        : (activeIndex - 1 + windows.length) % windows.length;

      renderActiveWindow();

      window.setTimeout(() => {
        isAnimating = false;
      }, animationDuration);
    });
  });

  renderActiveWindow();
};
