const state = {
  at: "",
  user: null,
  req: {},
  refreshCopy: false,
  max: 2,
  list: {},
  canStop: false,
};

const mutations = {
  setAt: (state, value) => (state.at = value),
  setUser: (state, value) => (state.user = value),
  updateReq: (state, value) => (state.req = value),
  setRefreshCopy: (state, value) => (state.refreshCopy = value),
  setListProgressAdd1: (state, value) => {
    if (state.list[value.name]) {
      state.list[value.name].process = value.value;
    }
  },
  setListLastLoad: (state, value) => {
    if (state.list[value.name]) {
      state.list[value.name].lastLoad = value.value;
    }
  },
  setListSpeed: (state, value) => {
    if (state.list[value.name]) {
      state.list[value.name].speed = value.value;
    }
  },
  setListRemain: (state, value) => {
    if (state.list[value.name]) {
      state.list[value.name].remain = value.value;
    }
  },
  setList: (state, value) => {
    state.list = value;
  },
  setCanStop: (state, value) => (state.canStop = value),
  setListItemSSE: (state, value) => {
    state.list[value.name].sse = value.sse;
  },
  deleteListItem: (state, value) => {
    delete state.list[value];
    state.list = Object.assign({}, state.list);
  },
  disconnectSSE: (state, value) => {
    console.log(Object.values(state.list), "12312312312");
    if (state.list[value]?.sse) state.list[value]?.sse?.disconnect();
  },
  changeMax: (state, value) => {
    state.max = value;
  },
};

const actions = {};

const getters = {
  isLogged: (state) => state.user !== null,
  isFiles: (state) => !state.loading && state.route.name === "Files",
  selectedCount: (state) => state.selected.length,
  // progress: (state) => {
  //   if (state.upload.progress.length === 0) {
  //     return 0;
  //   }

  //   let totalSize = state.upload.sizes.reduce((a, b) => a + b, 0);

  //   let sum = state.upload.progress.reduce((acc, val) => acc + val);
  //   return Math.ceil((sum / totalSize) * 100);
  // },
  filesInUploadCount: (state) => {
    return Object.keys(state.upload.uploads).length + state.upload.queue.length;
  },
  filesInUpload: (state) => {
    let files = [];

    for (let index in state.upload.uploads) {
      let upload = state.upload.uploads[index];
      let id = upload.id;
      let type = upload.type;
      let name = upload.file.name;
      let size = state.upload.sizes[id];
      let isDir = upload.file.isDir;
      let progress = isDir
        ? 100
        : Math.ceil((state.upload.progress[id] / size) * 100);

      files.push({
        id,
        name,
        progress,
        type,
        isDir,
      });
    }

    return files.sort((a, b) => a.progress - b.progress);
  },
  currentPrompt: (state) => {
    return state.prompts.length > 0
      ? state.prompts[state.prompts.length - 1]
      : null;
  },
  currentPromptName: (_, getters) => {
    return getters.currentPrompt?.prompt;
  },
  uploadSpeed: (state) => state.upload.speedMbyte,
  eta: (state) => state.upload.eta,
};

export default { state, mutations, actions, namespaced: true, getters };
