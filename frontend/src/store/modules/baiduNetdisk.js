const state = {
  at: "",
  user: null,
  req: {},
  refreshCopy: false,
};

const mutations = {
  setAt: (state, value) => (state.at = value),
  setUser: (state, value) => (state.user = value),
  updateReq: (state, value) => (state.req = value),
  setRefreshCopy: (state, value) => (state.refreshCopy = value),
};

const actions = {};

export default { state, mutations, actions, namespaced: true };
