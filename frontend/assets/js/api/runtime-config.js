const config = window.__APP_CONFIG__ || {};

const toPositiveNumber = (value, fallback) => {
  const parsed = Number(value);
  return Number.isFinite(parsed) && parsed > 0 ? parsed : fallback;
};

const getDefaultApiBaseUrl = () => {
  const { protocol, hostname, port } = window.location;

  if (protocol === 'file:') {
    return 'http://localhost:8080/api';
  }

  if (port === '8080') {
    return '/api';
  }

  if (hostname === 'localhost' || hostname === '127.0.0.1') {
    return 'http://localhost:8080/api';
  }

  return '/api';
};

export const API_BASE_URL = String(
  config.API_BASE_URL ?? window.__API_BASE_URL__ ?? getDefaultApiBaseUrl()
).replace(/\/+$/, '');

export const REQUEST_TIMEOUT_MS = toPositiveNumber(
  config.REQUEST_TIMEOUT_MS ?? window.__REQUEST_TIMEOUT_MS__,
  15000
);
