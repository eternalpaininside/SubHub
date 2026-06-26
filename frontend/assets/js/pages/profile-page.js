import { api } from '../api/backend-api-client.js';
import { t } from '../ui/i18n.js';
import { buildLayout, initLayoutUI } from '../ui/app-layout.js';
import { buildPageHeader } from '../ui/shared-page-sections.js';
import { isAuthenticated } from '../ui/session.js';

const renderProfileTile = (item) => `
  <div class="profile-summary-tile ${item.tileClass}">
    <div class="muted small">${item.label}</div>
    <div class="profile-summary-value">${item.value}</div>
  </div>
`;

async function renderProfile() {
  if (!isAuthenticated()) {
    window.location.href = 'index.html';
    return;
  }

  let data;
  try {
    data = await api.getProfile();
  } catch {
    window.location.href = 'index.html';
    return;
  }

  const profileTiles = [
    { label: 'Email', value: data.user.email, tileClass: 'profile-tile-neutral' },
    ...data.stats.map((item) =>
        ({ ...item, tileClass: 'profile-tile-accent' }))
  ];

  document.body.innerHTML = buildLayout('profile', `
    <main>
      ${buildPageHeader({
        title: 'Профиль',
        subtitle: 'Основные данные аккаунта и сводная статистика',
        actionButtonMarkup: '<button class="primary-btn js-logout" type="button">Выйти из профиля</button>'
      })}

      <section class="list-stack">
        <article class="card profile-main-card">
          <div class="profile-main-head">
            <img class="profile-main-photo" src="../assets/images/profile.svg" alt="Фото профиля" />
            <div>
              <div class="profile-main-name">${data.user.name}</div>
              <div class="muted">${data.user.handle}</div>
            </div>
            
            <div class="profile-language-switcher">
                <label for="languageSwitcher">${t('language_mode')}</label>
                
                <select id="languageSwitcher" class="language-select">
                    <option value="ru">Русский</option>
                    <option value="en">English</option>
                </select>
            </div>
          </div>
          <div class="profile-summary-grid">${profileTiles.map(renderProfileTile).join('')}</div>
        </article>
      </section>
    </main>
  `);

  initLayoutUI();
}

renderProfile();
