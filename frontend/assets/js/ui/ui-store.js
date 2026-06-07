const createStore = (initialState) => {
  let state = structuredClone(initialState);
  const listeners = new Set();

  return {
    getState: () => state,
    setState: (patch) => {
      state = { ...state, ...patch };
      listeners.forEach((listener) => listener(state));
    }
  };
};

export const subscriptionsStore = createStore({
  activeFilter: 'Все'
});

export const editContextStore = createStore({
  subscription: null,
  group: null
});
