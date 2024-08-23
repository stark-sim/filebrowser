import Vue from "vue";
import Vuex from "vuex";
import mutations from "./mutations";
import getters from "./getters";
import upload from "./modules/upload";
import bd from "./modules/baiduNetdisk";
import cep from "./modules/cephalonCloud";

Vue.use(Vuex);

const state = {
  user: null,
  req: {}, // My Files 的数据 ？todo：del
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
  handlingType: "", // 识别我的文件（本地）或百度网盘或者端脑云空间
};

export default new Vuex.Store({
  strict: true,
  state,
  getters,
  mutations,
  modules: { upload, bd, cep },
});
