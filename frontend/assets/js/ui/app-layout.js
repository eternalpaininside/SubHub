import { api } from '../api/backend-api-client.js';
import { STORAGE_KEYS } from './constants.js';
import { t } from './i18n.js';
import { clearCurrentUser, getCurrentUser, isAuthenticated, setCurrentUser } from './session.js';
import { createModalManager } from './modal-manager.js';
import { editContextStore } from './ui-store.js';
import { initHelpSystem } from './help-system.js';

const subscriptionCategories = ['Стриминг', 'Музыка', 'Облако', 'Продуктивность', 'AI', 'Комплекс'];
const commonPeriods = ['мес', 'год'];
const planTypes = ['Индивидуальный', 'Групповой'];
const groupTypes = ['Семейная', 'Друзья', 'Команда'];
const fieldFullClass = 'field-label field-full';
const subscriptionFieldSelector = (field) => `[data-edit-subscription-field="${field}"]`;

const navItems = [
  { href: 'index.html', labelKey: 'nav_home', key: 'home' },
  { href: 'analytics.html', labelKey: 'nav_analytics', key: 'analytics' },
  { href: 'groups.html', labelKey: 'nav_groups', key: 'groups' }
];

const authModes = {
  login: {
    titleKey: 'auth_login_title',
    subtitleKey: 'auth_login_subtitle',
    submitKey: 'button_login',
    fields: [
      { id: 'login-email', label: t('label_email'),
        type: 'email', placeholder: 'you@example.com' },
      { id: 'login-password', label: t('label_password'),
         type: 'password', placeholder: 'Введите пароль' }
    ]
  },
  register: {
    titleKey: 'auth_register_title',
    subtitleKey: 'auth_register_subtitle',
    submitKey: 'button_register',
    fields: [
      { id: 'register-name', label: t('label_name'),
        placeholder: 'Ваше имя' },
      { id: 'register-email', label: t('label_email'),
        type: 'email', placeholder: 'you@example.com' },
      { id: 'register-password', label: t('label_password'),
        type: 'password', placeholder: 'Придумайте пароль' }
    ]
  }
};

const subscriptionEditableFields = {
  name: { type: 'text' },
  price: { type: 'number' },
  status: { type: 'select', fallback: 'true' }
};

const renderInputField = ({ id = '', label, type = 'text',
                            placeholder = '', dataField = '',
                            full = false }) => `
  <label class="${full ? fieldFullClass : 'field-label'}">${label}
    <input
      ${id ? `id="${id}"` : ''}
      ${dataField ? `data-edit-subscription-field="${dataField}"` : ''}
      class="field-input"
      type="${type}"
      ${placeholder ? `placeholder="${placeholder}"` : ''}
    />
  </label>
`;

const renderSelectField = ({ id = '', label, options, dataField = '', full = false }) => `
  <label class="${full ? fieldFullClass : 'field-label'}">${label}
    <select
      ${id ? `id="${id}"` : ''}
      ${dataField ? `data-edit-subscription-field="${dataField}"` : ''}
      class="field-input field-select"
      >
      ${options.map((option) => {
        const normalized = typeof option === 'string' ? { label: option, value: option } : option;
        return `<option value="${normalized.value}">${normalized.label}</option>`;
      }).join('')}
    </select>
  </label>
`;

const renderTextareaField = ({ id = '', label, placeholder = '', dataField = '', full = false }) => `
  <label class="${full ? fieldFullClass : 'field-label'}">${label}
    <textarea
      ${id ? `id="${id}"` : ''}
      ${dataField ? `data-edit-subscription-field="${dataField}"` : ''}
      class="field-input field-textarea"
      ${placeholder ? `placeholder="${placeholder}"` : ''}
    ></textarea>
  </label>
`;

const renderCheckboxListField = ({ id = '', label, full = false }) => `
  <label class="${full ? fieldFullClass : 'field-label'}">${label}
    <div id="${id}" class="field-checkbox-list"></div>
  </label>
`;

const renderField = (field) => {
  if (field.kind === 'multiselect')
    return renderCheckboxListField(field);
  if (field.kind === 'select')
    return renderSelectField(field);
  if (field.kind === 'textarea')
    return renderTextareaField(field);
  return renderInputField(field);
};

const renderModalActions = (buttons, full = false) => `
  <div class="modal-actions${full ? ' field-full' : ''}">
    ${buttons.map((button) => `
      <button
        class="${button.className}"
        type="button"
        ${button.id ? `id="${button.id}"` : ''}
        ${button.attributes || ''}
      >${button.label}</button>
    `).join('')}
  </div>
`;

const renderFormModal = ({ id, cardClass, title, subtitle, formClass = 'modal-form', fields = [], actions = [], extraMarkup = '' }) => `
  <div class="modal-backdrop" id="${id}" aria-hidden="true">
    <div class="modal-card ${cardClass}">
      <button class="modal-close js-close-modal" type="button" aria-label="Закрыть">×</button>
      <h2 class="modal-title">${title}</h2>
      ${subtitle ? `<p class="modal-subtitle">${subtitle}</p>` : ''}
      ${fields.length || actions.length ? `
        <form class="${formClass}">
          ${fields.map(renderField).join('')}
          ${actions.length ? renderModalActions(actions, formClass.includes('two-columns')) : ''}
        </form>
      ` : ''}
      ${extraMarkup}
    </div>
  </div>
`;

const renderAuthModal = () => `
  <div class="modal-backdrop" id="auth-modal" aria-hidden="true">
    <div class="modal-card auth-modal-card">
      <button class="modal-close js-close-modal" type="button" aria-label="Закрыть">×</button>
      <div class="auth-switcher">
        <button class="auth-tab is-active" type="button" data-auth-tab="login">${t('auth_login_tab')}</button>
        <button class="auth-tab" type="button" data-auth-tab="register">${t('auth_register_tab')}</button>
      </div>
      ${Object.entries(authModes).map(([mode, config], index) => `
        <div class="auth-panel ${index === 0 ? 'is-active' : ''}" data-auth-panel="${mode}">
          <h2 class="modal-title">${t(config.titleKey)}</h2>
          <p class="modal-subtitle">${t(config.subtitleKey)}</p>
          <form class="modal-form" id="${mode}-form">
            ${config.fields.map(renderField).join('')}
            <button class="primary-btn modal-submit" 
            type="button" 
            data-auth-submit="${mode}">${t(config.submitKey)}</button>
          </form>
        </div>
      `).join('')}
    </div>
  </div>
`;

const renderLayoutHeader = (activePage) => {
  const profileLabel = getCurrentUser()?.name || 'Войти';

  return `
    <header class="topbar">
      <div class="branding">
        <a class="logo" href="index.html">${t('brand')}</a>
        <nav class="nav">
          ${navItems.map((item) => `
            <a class="nav-link${item.key === activePage ? 'is-active' : ''}" 
            href="${item.href}">${t(item.labelKey)}</a>
          `).join('')}
        </nav>
      </div>
      <div class="topbar-actions">
        <button class="help-button js-open-help ${activePage === 'help' ? 'is-active' : ''}" 
        type="button">Помощь</button>
        <button class="profile-chip ${activePage === 'profile' ? 'is-active' : ''}" 
        type="button" aria-label="${t('auth_open_profile')}">
          <img class="profile-photo" 
          src="../assets/images/profile.svg" 
          alt="Фото профиля"
          />
          <span>${profileLabel}</span>
        </button>
      </div>
    </header>
  `;
};

const subscriptionModalMarkup = () => renderFormModal({
  id: 'subscription-modal',
  cardClass: 'subscription-modal-card',
  title: 'Добавить подписку',
  subtitle: 'Заполните данные подписки, чтобы сохранить её в базе.',
  formClass: 'modal-form two-columns',
  fields: [
    { id: 'sub-name', label: 'Название сервиса', placeholder: 'Например, Netflix' },
    { id: 'sub-category', label: 'Категория', kind: 'select', options: subscriptionCategories },
    { id: 'sub-price', label: 'Стоимость', type: 'number', placeholder: '799' },
    { id: 'sub-period', label: 'Период оплаты', kind: 'select', options: commonPeriods },
    { id: 'sub-planType', label: 'Тип тарифа', kind: 'select', options: planTypes },
    { id: 'sub-nextPayment', label: 'Дата следующего платежа', type: 'date' },
    { id: 'sub-link', label: 'Ссылка на сайт', type: 'url', placeholder: 'https://...' },
    { id: 'sub-comment', label: 'Комментарий', kind: 'textarea', full: true, placeholder: 'Например, семейный тариф на 4 человека' }
  ],
  actions: [
    { label: 'Отмена', className: 'ghost-btn js-close-modal' },
    { label: 'Сохранить', className: 'primary-btn', id: 'add-subscription-button' }
  ]
});

const groupModalMarkup = () => renderFormModal({
  id: 'group-modal',
  cardClass: 'group-modal-card',
  title: 'Создать группу',
  subtitle: 'Выберите подписки для группы. Её стоимость посчитается автоматически.',
  formClass: 'modal-form two-columns',
  fields: [
    { id: 'group-name', label: 'Название группы', placeholder: 'Например, Семья Ивановых' },
    { id: 'group-type', label: 'Тип группы', kind: 'select', options: groupTypes },
    { id: 'group-subscription-ids', label: 'Подписки группы', kind: 'multiselect', full: true },
    { id: 'group-notes', label: 'Примечания', kind: 'textarea', full: true, placeholder: 'Групповые/Семейные подписки' }
  ],
  actions: [
    { label: 'Отмена', className: 'ghost-btn js-close-modal' },
    { label: 'Создать группу', className: 'primary-btn', id: 'create-group-button' }
  ]
});

const groupJoinModalMarkup = () => renderFormModal({
  id: 'group-join-modal',
  cardClass: 'group-modal-card',
  title: 'Присоединиться к группе',
  subtitle: 'Вставьте ссылку-приглашение, чтобы присоединиться к группе.',
  fields: [
    { id: 'group-join-invite', label: 'Ссылка-приглашение', type: 'url', placeholder: 'https://subhub.local/invite/...' }
  ],
  actions: [
    { label: 'Отмена', className: 'ghost-btn js-close-modal' },
    { label: 'Присоединиться', className: 'primary-btn', id: 'join-group-button' }
  ]
});

const subscriptionEditModalMarkup = () => renderFormModal({
  id: 'subscription-edit-modal',
  cardClass: 'subscription-modal-card',
  title: 'Изменить подписку',
  subtitle: 'Измените название, стоимость или статус подписки.',
  formClass: 'modal-form',
  fields: [
    { label: 'Название сервиса', dataField: 'name' },
    { label: 'Стоимость', type: 'number', dataField: 'price' },
    { label: 'Статус', kind: 'select', dataField: 'status', options: [{ label: 'Активна', value: 'true' }, { label: 'Приостановлена', value: 'false' }] },
  ],
  actions: [
    { label: 'Полностью удалить', className: 'ghost-btn', id: 'delete-subscription-button' },
    { label: 'Отмена', className: 'ghost-btn js-close-modal' },
    { label: 'Сохранить изменения', className: 'primary-btn', id: 'save-subscription-button' }
  ]
});

const groupEditModalMarkup = () => renderFormModal({
  id: 'group-edit-modal',
  cardClass: 'group-modal-card',
  title: 'Изменить группу',
  subtitle: 'Измените название, тип и набор подписок в группе.',
  formClass: 'modal-form two-columns',
  fields: [
    { id: 'group-edit-name', label: 'Название группы' },
    { id: 'group-edit-type', label: 'Тип группы', kind: 'select', options: groupTypes },
    { id: 'group-edit-subscription-ids', label: 'Подписки группы', kind: 'multiselect', full: true },
    { id: 'group-edit-notes', label: 'Примечания', kind: 'textarea', full: true }
  ],
  actions: [
    { label: 'Полностью удалить', className: 'ghost-btn', id: 'delete-group-button' },
    { label: 'Отмена', className: 'ghost-btn js-close-modal' },
    { label: 'Сохранить изменения', className: 'primary-btn', id: 'save-group-button' }
  ]
});

export function buildLayout(activePage, content) {
  return `
    <div class="page-shell" data-active-page="${activePage}">
      ${renderLayoutHeader(activePage)}
      ${content}
    </div>
    ${renderAuthModal()}
    ${subscriptionModalMarkup()}
    ${groupModalMarkup()}
    ${groupJoinModalMarkup()}
    ${subscriptionEditModalMarkup()}
    ${groupEditModalMarkup()}
  `;
}

const setFieldValue = (modal, selector, value) => {
  const field = modal?.querySelector(selector);
  if (field && value !== undefined) field.value = String(value);
};

const applySelectValue = (modal, selector, value, fallback) => {
  const select = modal?.querySelector(selector);
  if (!select) return;

  const normalizedValue = String(value || fallback || '');
  const hasOption = Array.from(select.options).some((option) => option.value === normalizedValue);
  select.value = hasOption ? normalizedValue : String(fallback || select.options[0]?.value || '');
};

const updateHeaderProfileLabel = () => {
  const label = document.querySelector('.profile-chip span');
  if (label) label.textContent = getCurrentUser()?.name || 'Войти';
};

const isForcedAuthActive = () =>
    window.localStorage.getItem(STORAGE_KEYS.forceAuth) === 'true';

const setForcedAuthState = (enabled) => {
  if (enabled) {
    window.localStorage.setItem(STORAGE_KEYS.forceAuth, 'true');
    return;
  }
  window.localStorage.removeItem(STORAGE_KEYS.forceAuth);
};

const readValue = (id) => document.getElementById(id)?.value || '';
const readTrimmedValue = (id) => readValue(id).trim();
const normalizePeriodValue = (value) => (value === 'год' ? 12 : 1);
const readSelectedIntValues = (id) => Array.from(document.querySelectorAll(`#${id} input[type="checkbox"]:checked`))
  .map((input) => Number(input.value))
  .filter((value) => Number.isFinite(value) && value > 0);

const requireAuthForModal = (modalManager, currentModalId) => {
  if (isAuthenticated())
    return true;

  alert('Сначала войдите в аккаунт.');
  modalManager.close(currentModalId);
  modalManager.open('auth-modal');

  return false;
};

const validateSubscriptionPayload = (payload, message) => {
  if (!payload.name || !payload.category
      || !payload.next_payment || !payload.link
      || !Number.isFinite(payload.price) || payload.price <= 0) {
    alert(message);
    return false;
  }
  return true;
};

const buildSubscriptionPayloadFromModal = () => ({
  name: readTrimmedValue('sub-name'),
  category: readTrimmedValue('sub-category'),
  price: Number(readValue('sub-price')),
  period: normalizePeriodValue(readValue('sub-period')),
  plan_type: readValue('sub-planType') || 'Индивидуальный',
  next_payment: readValue('sub-nextPayment'),
  link: readTrimmedValue('sub-link'),
  comment: readTrimmedValue('sub-comment'),
  status: true
});

const buildEditedSubscriptionPayload = (modal) => ({
  name: modal.querySelector(subscriptionFieldSelector('name'))?.value.trim() || '',
  category: editContextStore.getState().subscription?.category || 'Стриминг',
  price: Number(modal.querySelector(subscriptionFieldSelector('price'))?.value),
  period: Number(editContextStore.getState().subscription?.periodValue) || 1,
  plan_type: editContextStore.getState().subscription?.planType || 'Индивидуальный',
  next_payment: editContextStore.getState().subscription?.date || '',
  link: editContextStore.getState().subscription?.site || '',
  comment: editContextStore.getState().subscription?.comment || '',
  status: modal.querySelector(subscriptionFieldSelector('status'))?.value === 'true'
});

const initAddSubscriptionForm = (modalManager) => {
  const addButton = document.getElementById('add-subscription-button');
  const modal = document.getElementById('subscription-modal');
  const form = modal?.querySelector('form');
  if (!addButton || !modal || !form)
    return;

  addButton.addEventListener('click', async () => {
    if (!requireAuthForModal(modalManager, 'subscription-modal'))
      return;

    const payload = buildSubscriptionPayloadFromModal();
    if (!validateSubscriptionPayload(payload, 'Заполните все обязательные поля корректно.')) return;

    try {
      await api.createSubscription(payload);
      form.reset();
      modalManager.close('subscription-modal');
      window.dispatchEvent(new CustomEvent('subscription:changed'));
    } catch (error) {
      alert(error.message || 'Не удалось добавить подписку.');
    }
  });
};

const initEditSubscriptionForm = (modalManager) => {
  const modal = document.getElementById('subscription-edit-modal');
  const saveButton = document.getElementById('save-subscription-button');
  const deleteButton = document.getElementById('delete-subscription-button');
  if (!modal || !saveButton || !deleteButton)
    return;

  saveButton.addEventListener('click', async () => {
    const id = Number(editContextStore.getState().subscription?.id);
    if (!Number.isFinite(id) || id <= 0) {
      alert('Не удалось определить подписку для обновления.');
      return;
    }

    const payload = buildEditedSubscriptionPayload(modal);
    if (!validateSubscriptionPayload(payload, 'Заполните поля подписки корректно.'))
      return;

    try {
      await api.updateSubscription(id, payload);
      modalManager.close('subscription-edit-modal');
      window.dispatchEvent(new CustomEvent('subscription:changed'));
    } catch (error) {
      alert(error.message || 'Не удалось обновить подписку.');
    }
  });

  deleteButton.addEventListener('click', async () => {
    const id = Number(editContextStore.getState().subscription?.id);
    if (!Number.isFinite(id) || id <= 0) {
      alert('Не удалось определить подписку для удаления.');
      return;
    }

    try {
      await api.deleteSubscription(id);
      modalManager.close('subscription-edit-modal');
      window.dispatchEvent(new CustomEvent('subscription:changed'));
    } catch (error) {
      alert(error.message || 'Не удалось удалить подписку.');
    }
  });
};

const initGroupCreateForm = (modalManager) => {
  const createButton = document.getElementById('create-group-button');
  if (!createButton)
    return;

  createButton.addEventListener('click', async () => {
    if (!requireAuthForModal(modalManager, 'group-modal'))
      return;

    const payload = {
      name: readTrimmedValue('group-name'),
      type: readValue('group-type') || 'Семейная',
      notes: readTrimmedValue('group-notes'),
      subscription_ids: readSelectedIntValues('group-subscription-ids')
    };

    if (!payload.name) {
      alert('Введите название группы.');
      return;
    }

    try {
      await api.createGroup(payload);
      document.getElementById('group-modal')?.querySelector('form')?.reset();
      modalManager.close('group-modal');
      window.dispatchEvent(new CustomEvent('group:changed'));
    } catch (error) {
      alert(error.message || 'Не удалось создать группу.');
    }
  });
};

const initGroupEditForm = (modalManager) => {
  const saveButton = document.getElementById('save-group-button');
  const deleteButton = document.getElementById('delete-group-button');
  if (!saveButton || !deleteButton)
    return;

  saveButton.addEventListener('click', async () => {
    const id = Number(editContextStore.getState().group?.id);
    if (!Number.isFinite(id) || id <= 0) {
      alert('Не удалось определить группу для обновления.');
      return;
    }

    const payload = {
      name: readTrimmedValue('group-edit-name'),
      type: readValue('group-edit-type') || 'Семейная',
      notes: readTrimmedValue('group-edit-notes'),
      subscription_ids: readSelectedIntValues('group-edit-subscription-ids')
    };

    if (!payload.name) {
      alert('Введите название группы.');
      return;
    }

    try {
      await api.updateGroup(id, payload);
      modalManager.close('group-edit-modal');
      window.dispatchEvent(new CustomEvent('group:changed'));
    } catch (error) {
      alert(error.message || 'Не удалось обновить группу.');
    }
  });

  deleteButton.addEventListener('click', async () => {
    const id = Number(editContextStore.getState().group?.id);
    if (!Number.isFinite(id) || id <= 0) {
      alert('Не удалось определить группу для удаления.');
      return;
    }

    try {
      await api.deleteGroup(id);
      modalManager.close('group-edit-modal');
      window.dispatchEvent(new CustomEvent('group:changed'));
    } catch (error) {
      alert(error.message || 'Не удалось удалить группу.');
    }
  });
};

const initGroupJoinForm = (modalManager) => {
  const joinButton = document.getElementById('join-group-button');
  const inviteField = document.getElementById('group-join-invite');
  if (!joinButton || !inviteField)
    return;

  joinButton.addEventListener('click', async () => {
    if (!requireAuthForModal(modalManager, 'group-join-modal'))
      return;

    const inviteURL = inviteField.value.trim();
    if (!inviteURL) {
      alert('Вставьте ссылку-приглашение.');
      return;
    }

    try {
      await api.joinGroup(inviteURL);
      inviteField.value = '';
      modalManager.close('group-join-modal');
      window.dispatchEvent(new CustomEvent('group:changed'));
    } catch (error) {
      alert(error.message || 'Не удалось присоединиться к группе.');
    }
  });
};

const enablePageMotion = () => {
  requestAnimationFrame(() => {
    document.body.classList.add('is-page-ready');
  });

  document.querySelectorAll('a[href]').forEach((link) => {
    const href = link.getAttribute('href');
    if (!href || href.startsWith('#') || href.startsWith('http') || link.target === '_blank')
      return;

    link.addEventListener('click', (event) => {
      if (event.metaKey || event.ctrlKey || event.shiftKey || event.altKey)
        return;

      event.preventDefault();
      document.body.classList.remove('is-page-ready');
      document.body.classList.add('is-page-exiting');
      window.setTimeout(() => {
        window.location.href = href;
      }, 220);
    });
  });
};

const initAuthTabs = () => {
  const tabs = document.querySelectorAll('[data-auth-tab]');
  const panels = document.querySelectorAll('[data-auth-panel]');

  tabs.forEach((tab) => {
    tab.addEventListener('click', () => {
      const key = tab.dataset.authTab;
      tabs.forEach((item) => item.classList.toggle('is-active', item === tab));
      panels.forEach((panel) => panel.classList.toggle('is-active', panel.dataset.authPanel === key));
    });
  });
};

const getAuthPayload = (mode) => mode === 'register'
  ? {
    name: readTrimmedValue('register-name'),
    email: readTrimmedValue('register-email'),
    password: readValue('register-password')
  }
  : {
    email: readTrimmedValue('login-email'),
    password: readValue('login-password')
  };

const initAuthActions = (modalManager) => {
  document.querySelectorAll('[data-auth-submit]').forEach((button) => {
    button.addEventListener('click', async () => {
      const mode = button.dataset.authSubmit;

      try {
        const response = await api[mode === 'register' ? 'register' : 'login'](getAuthPayload(mode));
        setCurrentUser(response.user);
        setForcedAuthState(false);
        const authModal = modalManager.getModal('auth-modal');

        if (authModal)
          authModal.dataset.locked = 'false';

        updateHeaderProfileLabel();
        modalManager.close('auth-modal');
        window.setTimeout(() => {
          window.location.href = 'profile.html';
        }, 220);
      } catch (error) {
        alert(error.message || 'Не удалось выполнить авторизацию.');
      }
    });
  });
};

const populateSubscriptionEditModal = (modal, trigger) => {
  if (!modal || !trigger)
    return;

    const payload = {
      id: trigger.dataset.subscriptionId || '',
      name: trigger.dataset.subscriptionName || '',
      category: trigger.dataset.subscriptionCategory || 'Стриминг',
      price: trigger.dataset.subscriptionPrice || '',
      period: trigger.dataset.subscriptionPeriod || 'мес',
      periodValue: trigger.dataset.subscriptionPeriod === 'год' ? 12 : 1,
      planType: trigger.dataset.subscriptionPlanType || 'Индивидуальный',
      date: trigger.dataset.subscriptionDate || '',
      site: trigger.dataset.subscriptionSite || '',
    comment: trigger.dataset.subscriptionComment || '',
    status: trigger.dataset.subscriptionStatus || 'true'
  };

  editContextStore.setState({ subscription: payload });
  Object.entries(subscriptionEditableFields).forEach(([field, config]) => {
    const selector = subscriptionFieldSelector(field);
    if (config.type === 'select') {
      applySelectValue(modal, selector, payload[field], config.fallback);
      return;
    }
    setFieldValue(modal, selector, payload[field]);
  });
};

const populateSubscriptionsSelect = async (selectId, selectedIDs = []) => {
  const container = document.getElementById(selectId);
  if (!container)
    return;

  try {
    const subscriptions = await api.getSubscriptions();
    const selected = new Set(selectedIDs.map((id) => Number(id)));
    container.innerHTML = subscriptions.map((subscription) => `
      <label class="checkbox-option">
        <input
          type="checkbox"
          value="${subscription.id}"
          ${selected.has(Number(subscription.id)) ? 'checked' : ''}
        />
        <span>${subscription.name} • ${subscription.price} ₽/${subscription.period}</span>
      </label>
    `).join('');
  } catch {
    container.innerHTML = '';
  }
};

export function initLayoutUI() {
  const modalManager = createModalManager();
  const subscriptionEditModal = document.getElementById('subscription-edit-modal');
  const groupEditModal = document.getElementById('group-edit-modal');

  initHelpSystem();
  modalManager.bindTrigger({ selector: '.js-open-auth', modalId: 'auth-modal' });
  modalManager.bindTrigger({ selector: '.js-open-add-subscription', modalId: 'subscription-modal' });
  modalManager.bindTrigger({ selector: '.js-open-join-group', modalId: 'group-join-modal' });
  modalManager.bindTrigger({
    selector: '.js-open-edit-subscription',
    modalId: 'subscription-edit-modal',
    beforeOpen: (trigger) => populateSubscriptionEditModal(subscriptionEditModal, trigger)
  });

  document.querySelectorAll('.js-open-add-group').forEach((trigger) => {
    trigger.addEventListener('click', async () => {
      await populateSubscriptionsSelect('group-subscription-ids');
      document.getElementById('group-modal')?.querySelector('form')?.reset();
      modalManager.open('group-modal');
    });
  });

  document.querySelectorAll('.js-open-edit-group').forEach((trigger) => {
    trigger.addEventListener('click', async () => {
      const groupState = {
        id: trigger.dataset.groupId || '',
        name: trigger.dataset.groupName || '',
        type: trigger.dataset.groupType || 'Семейная',
        notes: trigger.dataset.groupNotes || '',
        subscriptionIDs: String(trigger.dataset.groupSubscriptionIds || '')
          .split(',')
          .map((value) => Number(value.trim()))
          .filter((value) => Number.isFinite(value) && value > 0)
      };

      editContextStore.setState({ group: groupState });
      setFieldValue(groupEditModal, '#group-edit-name', groupState.name);
      applySelectValue(groupEditModal, '#group-edit-type', groupState.type, 'Семейная');
      setFieldValue(groupEditModal, '#group-edit-notes', groupState.notes);

      await populateSubscriptionsSelect('group-edit-subscription-ids', groupState.subscriptionIDs);
      modalManager.open('group-edit-modal');
    });
  });

  modalManager.bindCloseHandlers();
  initAddSubscriptionForm(modalManager);
  initEditSubscriptionForm(modalManager);
  initGroupCreateForm(modalManager);
  initGroupEditForm(modalManager);
  initGroupJoinForm(modalManager);
  initAuthActions(modalManager);

  document.querySelectorAll('.profile-chip').forEach((button) => {
    button.addEventListener('click', () => {
      if (isAuthenticated()) {
        window.location.href = 'profile.html';
        return;
      }
      modalManager.open('auth-modal');
    });
  });

  document.querySelectorAll('.js-logout').forEach((button) => {
    button.addEventListener('click', () => {
      if (!isAuthenticated())
        return;

      clearCurrentUser();
      setForcedAuthState(true);
      updateHeaderProfileLabel();
      window.location.href = 'index.html';
    });
  });

  initAuthTabs();
  enablePageMotion();
  updateHeaderProfileLabel();

  if (!isAuthenticated() && isForcedAuthActive()) {
    const authModal = modalManager.getModal('auth-modal');
    if (authModal) {
      authModal.dataset.locked = 'true';
      authModal.querySelectorAll('.js-close-modal').forEach((button) => {
        button.style.display = 'none';
      });
    }
    modalManager.open('auth-modal');
  }
}
