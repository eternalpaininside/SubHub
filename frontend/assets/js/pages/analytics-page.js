import { api } from '../api/backend-api-client.js';
import { buildLayout, initLayoutUI } from '../ui/app-layout.js';
import { escapeAttr } from '../ui/formatters.js';
import { buildPageHeader } from '../ui/shared-page-sections.js';

const renderAnalyticsPage = (title, subtitle, body) => `
  <main>
    ${buildPageHeader({ title, subtitle })}
    ${body}
  </main>
`;

const renderLegend = (legend) => legend.length
  ? legend.map((item) => `
    <div class="legend-item">
      <div class="legend-main">
        <div class="legend-left">
          <span class="dot" style="background:${item.color}"></span>
          <span>${item.name}</span>
        </div>
        <strong>${item.amountLabel} <span class="legend-share">${item.share}%</span></strong>
      </div>
      <div class="legend-meter">
        <span style="width:${item.meterWidth}%; background:${item.color};"></span>
      </div>
    </div>
  `).join('')
  : '<div class="legend-empty">Пока нет данных по категориям</div>';

const renderBarChart = (chart) => `
  <div class="bar-plot-wrap">
    <div class="axis-grid">
      ${chart.yTicks.map((tick, index) => `
        <div class="axis-line ${index === chart.yTicks.length - 1 ? 'is-baseline' : ''}">
          <span class="axis-note">${tick.label}</span>
        </div>
      `).join('')}
    </div>
    <div class="bar-chart expenses-chart" style="grid-template-columns:repeat(${chart.bars.length},minmax(0,1fr));">
      ${chart.bars.map((bar) => `
        <div class="bar-column">
          <div
            class="bar bar-expense"
            style="--target-height:${bar.heightPercent.toFixed(2)}%; --bar-delay:${bar.delayMs}ms;"
            title="${bar.label}: ${bar.amountLabel}"
            aria-label="${bar.label}: ${bar.amountLabel}"
          >
            <span class="bar-tooltip">${bar.amountLabel}</span>
          </div>
        </div>
      `).join('')}
    </div>
  </div>
  <div class="bar-months" style="grid-template-columns:repeat(${chart.bars.length},minmax(0,1fr));">
    ${chart.bars.map((bar) => `<div class="bar-label">${bar.label}</div>`).join('')}
  </div>
`;

const renderKpiCard = ({ badge, title, valueLabel, meta }) => `
  <article class="card analytics-kpi-card">
    <div class="analytics-kpi-top">
      <span class="analytics-kpi-badge">${badge}</span>
    </div>
    <div class="analytics-kpi-title">${title}</div>
    <div class="metric-value analytics-kpi-value">${valueLabel}</div>
    <div class="analytics-kpi-meta-row">
      ${meta.map((item) => `<span class="analytics-kpi-meta-pill${item.className ? ` ${item.className}` : ''}">${item.text}</span>`).join('')}
    </div>
  </article>
`;

const renderFactCard = (title, body) => `
  <article class="card fact-card">
    <div class="fact-kicker">${title}</div>
    ${body}
  </article>
`;

const renderUpcomingCharges = (charges) => charges.length
  ? charges.map((subscription) => `
    <div class="fact-list-row">
      <span>${subscription.name}</span>
      <strong>${subscription.daysLeftNumber} дн.</strong>
    </div>
  `).join('')
  : '<div class="fact-muted">На этой неделе списаний нет</div>';

const renderAnalyticsContent = (viewModel) => {
  const topCards = [
    {
      badge: 'Неделя',
      title: viewModel.topCards.week.title,
      valueLabel: viewModel.topCards.week.valueLabel,
      meta: [
        { text: viewModel.topCards.week.primaryText },
        { text: viewModel.topCards.week.secondaryText }
      ]
    },
    {
      badge: 'Месяц',
      title: 'Расходы за месяц',
      valueLabel: viewModel.topCards.month.valueLabel,
      meta: [
        { text: viewModel.topCards.month.trendText, className: `analytics-kpi-meta-pill-trend ${viewModel.topCards.month.trendClass}` },
        { text: viewModel.topCards.month.previousSpendText }
      ]
    },
    {
      badge: 'Год',
      title: 'Расходы за год',
      valueLabel: viewModel.topCards.year.valueLabel,
      meta: [
        { text: viewModel.topCards.year.averageText },
        { text: viewModel.topCards.year.projectionText }
      ]
    }
  ];

  const factCards = [
    renderFactCard('Топ категория', `
      <div class="fact-title">${viewModel.facts.topCategory.name}</div>
      <div class="fact-value">${viewModel.facts.topCategory.amountLabel}</div>
      <div class="fact-muted">Доля в месячных расходах: ${viewModel.facts.topCategory.share}%</div>
      <div class="fact-progress">
        <span style="width:${viewModel.facts.topCategory.progressWidth}%; background:${viewModel.facts.topCategory.color};"></span>
      </div>
    `),
    renderFactCard('Самая дорогая подписка', `
      <div class="fact-title">${viewModel.facts.mostExpensiveSubscription.name}</div>
      <div class="fact-value">${viewModel.facts.mostExpensiveSubscription.priceLabel}</div>
      <div class="fact-tags">
        <span class="fact-tag fact-tag-plan">${viewModel.facts.mostExpensiveSubscription.planType}</span>
        <span class="fact-tag fact-tag-category">${viewModel.facts.mostExpensiveSubscription.category}</span>
      </div>
    `),
    renderFactCard('Ближайшие списания', `<div class="fact-list">${renderUpcomingCharges(viewModel.facts.upcomingCharges)}</div>`),
    renderFactCard('Разбивка по тарифам', `
        <div class="fact-split">
          <div class="fact-split-row">
          <span>Групповые</span>
          <strong>${viewModel.facts.planSplit.familyLabel}</strong>
        </div>
        <div class="fact-split-row">
          <span>Индивидуальные</span>
          <strong>${viewModel.facts.planSplit.individualLabel}</strong>
        </div>
      </div>
    `)
  ];

  return renderAnalyticsPage('Аналитика', 'Подробный анализ расходов', `
    <section class="analytics-grid-top">
      ${topCards.map(renderKpiCard).join('')}
    </section>

    <section class="analytics-grid-bottom">
      <article class="card chart-box" style="position:relative;">
        <div class="service-name analytics-section-title">Расходы по месяцам</div>
        ${renderBarChart(viewModel.expensesChart)}
      </article>

      <article class="card">
        <div class="service-name analytics-section-title">По категориям</div>
        <div class="ring-wrap">
          <div class="ring-chart">
            <svg class="ring-svg" viewBox="0 0 120 120" aria-label="Категории расходов">
              ${viewModel.categoriesPie.segments.length
                ? viewModel.categoriesPie.segments.map((segment) => `
                  <path
                    class="ring-segment"
                    d="${segment.path}"
                    fill="${segment.color}"
                    data-name="${escapeAttr(segment.name)}"
                    data-amount="${escapeAttr(segment.amountLabel)}"
                    data-share="${segment.share}"
                    data-subscriptions="${escapeAttr(segment.subscriptionsPreview)}"
                  ></path>
                `).join('')
                : '<circle cx="60" cy="60" r="52" fill="rgba(255,255,255,0.1)"></circle>'}
            </svg>
            <div class="ring-tooltip"></div>
          </div>
        </div>
        <div class="legend-list">${renderLegend(viewModel.categoriesPie.legend)}</div>
      </article>
    </section>

    <section class="analytics-facts-grid">
      ${factCards.join('')}
    </section>
  `);
};

const initRingChartTooltip = () => {
  document.querySelectorAll('.ring-chart').forEach((chart) => {
    const tooltip = chart.querySelector('.ring-tooltip');
    const segments = chart.querySelectorAll('.ring-segment');
    if (!tooltip || !segments.length) return;

    const setTooltipPosition = (event) => {
      const rect = chart.getBoundingClientRect();
      const tooltipWidth = tooltip.offsetWidth || 268;
      const tooltipHeight = tooltip.offsetHeight || 120;
      const halfWidth = tooltipWidth / 2;
      const x = Math.min(Math.max(event.clientX - rect.left, halfWidth + 8), rect.width - halfWidth - 8);
      const y = Math.min(Math.max(event.clientY - rect.top - 14, tooltipHeight + 8), rect.height - 8);
      tooltip.style.left = `${x}px`;
      tooltip.style.top = `${y}px`;
    };

    const hideTooltip = () => {
      tooltip.classList.remove('is-visible');
    };

    segments.forEach((segment) => {
      const showTooltip = (event) => {
        tooltip.textContent = `${segment.dataset.name || '—'}: ${segment.dataset.amount || '0'} • ${segment.dataset.share || '0'}%\nПодписки: ${segment.dataset.subscriptions || 'Нет подписок'}`;
        setTooltipPosition(event);
        tooltip.classList.add('is-visible');
      };

      segment.addEventListener('mouseenter', showTooltip);
      segment.addEventListener('mousemove', showTooltip);
      segment.addEventListener('mouseleave', hideTooltip);
    });

    chart.addEventListener('mouseleave', hideTooltip);
  });
};

async function renderAnalytics() {
  try {
    const viewModel = await api.getAnalyticsPage();
    document.body.innerHTML = buildLayout('analytics', renderAnalyticsContent(viewModel));
    initLayoutUI();
    initRingChartTooltip();
  } catch (error) {
    document.body.innerHTML = buildLayout('analytics', renderAnalyticsPage(
      'Аналитика',
      'Живые данные появятся после входа в аккаунт',
      `<section class="list-stack"><article class="card">${error.message || 'Не удалось загрузить аналитику.'}</article></section>`
    ));
    initLayoutUI();
  }
}

renderAnalytics();
