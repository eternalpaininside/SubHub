import { api } from '../api/backend-api-client.js';
import { groupCard } from '../ui/subscription-group-cards.js';
import { buildLayout, initLayoutUI } from '../ui/app-layout.js';
import { buildPageHeader } from '../ui/shared-page-sections.js';

async function renderGroups() {
  let groups = [];
  try {
    groups = await api.getGroups();
  } catch {
    groups = [];
  }

  const content = `
    <main>
      ${buildPageHeader({
        title: 'Группы',
        subtitle: 'Семейные и групповые подписки',
        actionButtonMarkup: `
          <div style="display:flex; gap:12px; flex-wrap:wrap;">
            <button class="ghost-btn js-open-join-group" type="button">Присоединиться</button>
            <button class="primary-btn js-open-add-group" type="button">＋ Создать группу</button>
          </div>
        `
      })}

      <section class="groups-stack">
        ${groups.length ? groups.map(groupCard).join('') : '<article class="card">Пока нет групп. Создайте первую группу через кнопку сверху.</article>'}
      </section>
    </main>
  `;

  document.body.innerHTML = buildLayout('groups', content);
  initLayoutUI();
}

window.addEventListener('group:changed', renderGroups);

renderGroups();
