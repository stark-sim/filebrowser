import Vue from "vue";
import Vuex from "vuex";
import mutations from "./mutations";
import getters from "./getters";
import upload from "./modules/upload";
import bd from "./modules/baiduNetdisk";

Vue.use(Vuex);

const state = {
  user: null,
  req: {},
  oldReq: {},
  clipboard: {
    key: "",
    items: [],
  },
  jwt: "",
  progress: 0,
  loading: false,
  reload: false,
  selected: [],
  multiple: false,
  prompts: [],
  showShell: false,
  handlingType: "", // 识别我的文件（本地）或百度网盘
};

export default new Vuex.Store({
  strict: true,
  state,
  getters,
  mutations,
  modules: { upload, bd },
});
