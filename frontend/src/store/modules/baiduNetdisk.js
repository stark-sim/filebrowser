const state = {
  at: "",
  user: null,
  req: {},
};

const mutations = {
  setAt: (state, value) => (state.at = value),
  setUser: (state, value) => (state.user = value),
  updateReq: (state, value) => (state.req = value),
};

const actions = {};

export default { state, mutations, actions, namespaced: true };
