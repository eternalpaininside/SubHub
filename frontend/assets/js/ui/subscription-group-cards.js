import { escapeAttr } from './formatters.js';

const parseDaysLeft = (value) => {
  const match = String(value || '').match(/\d+/);
  return match ? Number(match[0]) : null;
};

const getDaysLeftNumber = (item) => {
  const direct = Number(item.daysLeftNumber);
  return Number.isFinite(direct) ? direct : parseDaysLeft(item.daysLeft);
};

const pluralizeDays = (days) => {
  const mod10 = days % 10;
  const mod100 = days % 100;
  if (mod10 === 1 && mod100 !== 11) return 'день';
  if (mod10 >= 2 && mod10 <= 4 && (mod100 < 10 || mod100 >= 20)) return 'дня';
  return 'дней';
};

const buildDaysLeftText = (item) => {
  const days = getDaysLeftNumber(item);
  return days === null || days > 7 ? '' : `Осталось ${days} ${pluralizeDays(days)}`;
};

const getBillingDate = (item) => item.dateText || '—';
const buildBillingDateText = (item, prefix = 'Списание') => `${prefix}: ${getBillingDate(item)}`;

const buildPlanLabel = (planType) => {
  const value = String(planType || '').toLowerCase();
  if (value.includes('груп')) return 'Групповой тариф';
  if (value.includes('инд')) return 'Индивидуальный тариф';
  return '';
};

const buildSubscriptionTags = (item, compact = false) => {
  const modifier = compact ? ' subscription-tag-compact' : '';
  const tags = [
    item.category ? `<span class="subscription-tag subscription-tag-category${modifier}">${item.category}</span>` : '',
    buildPlanLabel(item.planType) ? `<span class="subscription-tag subscription-tag-plan${modifier}">${buildPlanLabel(item.planType)}</span>` : ''
  ].filter(Boolean);

  return tags.length
    ? `<div class="subscription-tags${compact ? ' subscription-tags-compact' : ''}">${tags.join('')}</div>`
    : '';
};

const buildSubscriptionEditAttrs = (item) => [
  ['id', item.id || ''],
  ['name', item.name],
  ['category', item.category || 'Стриминг'],
  ['price', item.price || ''],
  ['period', item.period || 'мес'],
  ['plan-type', item.planType || 'Индивидуальный'],
  ['date', item.nextPaymentDate || ''],
  ['site', item.siteUrl || ''],
  ['comment', item.comment || ''],
  ['status', item.status !== false ? 'true' : 'false']
].map(([key, value]) => `data-subscription-${key}="${escapeAttr(value)}"`).join(' ');

const buildEditButton = (attrs, classes, label) => `
  <button class="${classes}" type="button" aria-label="${label}" ${attrs}>
    <span class="mirrored-edit-icon">✎</span>
  </button>
`;

export const subscriptionListCard = (item) => {
  const daysLeftText = buildDaysLeftText(item);
  const daysLeftLabel = daysLeftText || 'Осталось 00 дней';
  const attrs = buildSubscriptionEditAttrs(item);
  const editButton = buildEditButton(attrs, 'ghost-btn card-edit-btn list-edit-btn js-open-edit-subscription', 'Изменить подписку');
  return `
      <article class="list-card">
        <div class="list-left">
          <span class="service-icon" style="background:${item.iconBg}">${item.icon}</span>
          <div>
            <div class="service-name">${item.name}</div>
            ${buildSubscriptionTags(item, true)}
            <div class="list-meta">
              <span class="list-next-payment">${buildBillingDateText(item, 'Следующее списание')}</span>
              <span class="pink list-days-left ${daysLeftText ? '' : 'is-empty'}">${daysLeftLabel}</span>
            </div>
          </div>
        </div>
        <div class="list-price">
          <span class="value">${item.price} ₽</span>
          <span class="period">/${item.period}</span>
        </div>
        ${editButton}
      </article>
  `;
};

const buildGroupAttrs = (group, groupMembers) => [
  ['id', group.id || ''],
  ['name', group.name],
  ['type', group.type || 'Семейная'],
  ['members-count', group.members?.length || ''],
  ['budget', group.price || ''],
  ['period', group.period || 'мес'],
  ['invite', group.inviteUrl || ''],
  ['notes', group.notes || groupMembers],
  ['subscription-ids', (group.subscriptionIds || []).join(',')]
].map(([key, value]) => `data-group-${key}="${escapeAttr(value)}"`).join(' ');

export function groupCard(group) {
  const members = Array.isArray(group.members) ? group.members : [];
  const services = Array.isArray(group.services) ? group.services : [];
  const membersCount = members.length;
  const groupMembers = members.map((member) => member.name).filter(Boolean).join(', ');

  return `
    <article class="card group-card">
      <div class="group-main">
        <div class="group-top">
          <div class="group-icon">👥</div>
          <div class="group-title-block">
            <div class="service-name">${group.name}</div>
            <div class="muted">${membersCount} участников</div>
          </div>
        </div>
        <div class="subscription-tags subscription-tags-compact group-meta-tags">
          <span class="subscription-tag subscription-tag-category subscription-tag-compact">${group.type || 'Семейная группа'}</span>
          <span class="subscription-tag subscription-tag-plan subscription-tag-compact">${membersCount} участников</span>
        </div>
        <div class="group-section">
          <div class="group-section-title">Участники</div>
          <div class="pills">
            ${members.map((member) => `<span class="pill ${member.owner ? 'is-owner' : ''}">${member.owner ? '♛ ' : ''}${member.name}</span>`).join('')}
          </div>
        </div>
        <div class="group-section">
          <div class="group-section-title">Подписки</div>
          <div class="pills pills-services">
            ${services.length ? services.map((service) => `<span class="pill">${service}</span>`).join('') : '<span class="muted small">Пока не привязаны</span>'}
          </div>
        </div>
        <div class="group-section">
          <div class="group-section-title">Ссылка-приглашение</div>
          <div class="muted small">
            <a href="${group.inviteUrl || '#'}" target="_blank" rel="noreferrer">${group.inviteUrl || 'Будет сгенерирована автоматически'}</a>
          </div>
        </div>
      </div>
      <div class="group-price">
        <div class="group-price-value">
          <div class="price">${group.price} ₽</div>
          <div class="period">/${group.period}</div>
        </div>
        ${buildEditButton(buildGroupAttrs(group, groupMembers), 'ghost-btn card-edit-btn js-open-edit-group', 'Изменить группу')}
      </div>
    </article>
  `;
}
