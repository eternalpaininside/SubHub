export const buildPageHeader = ({ title, subtitle = '', actionButtonMarkup = '' }) => `
  <section class="page-header">
    <div>
      <h1 class="page-title">${title}</h1>
      ${subtitle ? `<p class="page-subtitle">${subtitle}</p>` : ''}
    </div>
    ${actionButtonMarkup}
  </section>
`;
