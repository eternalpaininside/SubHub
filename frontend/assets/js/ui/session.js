import { STORAGE_KEYS } from './constants.js';

export const getCurrentUser = () => {
  const raw = window.localStorage.getItem(STORAGE_KEYS.currentUser);
  if (!raw) return null;

  try {
    return JSON.parse(raw);
  } catch {
    return null;
  }
};

export const getCurrentUserId = () => {
  const user = getCurrentUser();
  const id = Number(user?.id);
  return Number.isFinite(id) && id > 0 ? id : null;
};

export const setCurrentUser = (user) => {
  window.localStorage.setItem(STORAGE_KEYS.auth, 'true');
  window.localStorage.setItem(STORAGE_KEYS.currentUser, JSON.stringify(user));
};

export const clearCurrentUser = () => {
  window.localStorage.removeItem(STORAGE_KEYS.auth);
  window.localStorage.removeItem(STORAGE_KEYS.currentUser);
};

export const isAuthenticated = () => getCurrentUserId() !== null;
