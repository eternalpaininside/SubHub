const helpFilePath = 'help.html';

const pageHelpAnchors = {
  home: 'home',
  analytics: 'analytics',
  groups: 'groups',
  profile: 'profile',
  help: 'overview'
};

const modalHelpAnchors = {
  'auth-modal': 'auth',
  'subscription-modal': 'subscription-create',
  'subscription-edit-modal': 'subscription-edit',
  'group-modal': 'group-create',
  'group-edit-modal': 'group-edit',
  'group-join-modal': 'group-join'
};

const fieldHelpAnchors = {
  'login-email': 'auth',
  'login-password': 'auth',
  'register-name': 'auth',
  'register-email': 'auth',
  'register-password': 'auth',
  'sub-name': 'subscription-create',
  'sub-category': 'subscription-create',
  'sub-price': 'subscription-create',
  'sub-period': 'subscription-create',
  'sub-planType': 'subscription-create',
  'sub-nextPayment': 'subscription-create',
  'sub-link': 'subscription-create',
  'sub-comment': 'subscription-create',
  'group-name': 'group-create',
  'group-type': 'group-create',
  'group-subscription-ids': 'group-create',
  'group-notes': 'group-create',
  'group-join-invite': 'group-join',
  'group-edit-name': 'group-edit',
  'group-edit-type': 'group-edit',
  'group-edit-subscription-ids': 'group-edit',
  'group-edit-notes': 'group-edit'
};

let isKeyboardHelpBound = false;

const getActivePage = () => document.querySelector('.page-shell')?.dataset.activePage || 'home';

const getVisibleModalAnchor = () => {
  const visibleModal = Array.from(document.querySelectorAll('.modal-backdrop'))
    .find((modal) => modal.getAttribute('aria-hidden') === 'false');

  return visibleModal ? modalHelpAnchors[visibleModal.id] : null;
};

const getFocusedFieldAnchor = () => {
  const activeElement = document.activeElement;
  if (!activeElement)
    return null;

  if (activeElement.id && fieldHelpAnchors[activeElement.id]) {
    return fieldHelpAnchors[activeElement.id];
  }

  if (activeElement.dataset?.editSubscriptionField)
    return 'subscription-edit';

  const checkboxList = activeElement.closest?.
    ('#group-subscription-ids, #group-edit-subscription-ids');

  if (checkboxList?.id)
    return fieldHelpAnchors[checkboxList.id] || null;

  return null;
};

const openHelp = (anchor = 'overview') => {
  window.open(`${helpFilePath}#${anchor}`, 'subhub-help');
};

const resolveContextAnchor = () => (
  getFocusedFieldAnchor() ||
  getVisibleModalAnchor() ||
  pageHelpAnchors[getActivePage()] ||
  'overview'
);

export function initHelpSystem() {
  document.querySelectorAll('.js-open-help').forEach((button) => {
    button.addEventListener('click', () => openHelp('overview'));
  });

  if (isKeyboardHelpBound)
    return;
  isKeyboardHelpBound = true;

  window.addEventListener('keydown', (event) => {
    if (event.key !== 'F1')
      return;
    event.preventDefault();
    openHelp(resolveContextAnchor());
  });
}
