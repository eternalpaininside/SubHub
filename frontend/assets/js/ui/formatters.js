export const formatRubles = (value) => new Intl.NumberFormat('ru-RU').format(Math.round(value));

export const toNumber = (value) => Number(
  String(value ?? '')
    .replace(',', '.')
    .replace(/[^\d.-]/g, '')
) || 0;

export const escapeAttr = (value) => String(value ?? '')
  .replace(/&/g, '&amp;')
  .replace(/"/g, '&quot;')
  .replace(/</g, '&lt;')
  .replace(/>/g, '&gt;');
