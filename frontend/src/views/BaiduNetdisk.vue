<template>
  <div class="dashboard">
    <header-bar v-if="showHeader" showMenu showLogo />

    <template v-if="user">
      <div class="netdisk-info">
        <img src="@/assets/img/baidu-netdisk-icon.png" />
        <p>
          <span>{{ user.name }}</span>
          <span>{{ $t(user.type) }}</span>
          <span>{{ user.hasUsed }} / {{ user.totolCap }}</span>
        </p>
        <button class="action" @click="logout">
          <img src="@/assets/img/unbind.png" />
          <span>{{ $t("baiduNetdisk.unbind") }}</span>
        </button>
      </div>
      <progress-bar :val="user.usedPercent" size="small" />

      <breadcrumbs base="/baidu-netdisk" />
    </template>

    <errors v-if="error" :errorCode="error.status" />
    <component v-else-if="currentView" :is="currentView"></component>

    <div v-else>
      <h2 class="message delayed">
        <div class="spinner">
          <div class="bounce1"></div>
          <div class="bounce2"></div>
          <div class="bounce3"></div>
        </div>
        <span>{{ $t("files.loading") }}</span>
      </h2>
    </div>
  </div>
</template>

<script>
import { bdApi } from "@/api";
import { mapState, mapMutations } from "vuex";

import HeaderBar from "@/components/header/HeaderBar.vue";
import Breadcrumbs from "@/components/Breadcrumbs.vue";
import Errors from "@/views/Errors.vue";
import AuthLogin from "@/views/baiduNetdisk/AuthLogin.vue";
import Listing from "@/views/baiduNetdisk/Listing.vue";
import ProgressBar from "vue-simple-progress";

/**
 * 1. 登录的情况
 * 2. 返回当前的目录：store.state.bd.req
 * 3. 现有：相关文件（夹）的操作
 *  (1) 文件夹的前进、后退
 *  (2) 文件的预览
 *  (3) 终端、视图、上传/下载、信息、多选
 *  (4) 选中的情况：分享、重命名、复制、移动、删除
 * 4. 基于第 3 点改版
 *  (1) 文件夹的前进、后退
 *  (2) 文件的预览（同 3.2）
 *  (3) 视图、文件信息、多选
 *  (4) 选中的情况：复制（下载到 My Files）
 * （5）获取下载进度
 */

export default {
  name: "baidu-netdisk",
  components: {
    HeaderBar,
    Breadcrumbs,
    Errors,
    AuthLogin,
    Listing,
    ProgressBar,
  },
  data: function () {
    return {
      error: null,
      width: window.innerWidth,
    };
  },
  computed: {
    ...mapState(["loading", "reload"]),
    ...mapState("bd", ["user", "req"]),
    currentView() {
      if (this.loading) {
        return null;
      } else if (this.user) {
        return "listing";
      } else {
        return "auth-login";
      }
    },
    showHeader() {
      return (
        this.error || this.req.type === null || this.currentView !== "listing"
      );
    },
  },
  async created() {
    try {
      this.setLoading(true);
      this.$store.commit("setHandlingType", "BaiduNetdisk");
      const at = bdApi.getToken();
      if (at) {
        await bdApi.fetchUserInfo();
      }
    } catch {
    } finally {
      this.setLoading(false);
    }
  },
  destroyed() {
    this.$store.commit("setHandlingType", "");
    this.$store.commit("bd/updateReq", {});
  },
  watch: {
    $route: "fetchData",
    reload: async function (value) {
      if (value === true) {
        await this.fetchData();
        // todo
        await bdApi.fetchProgress();
      }
    },
  },
  methods: {
    ...mapMutations(["setLoading"]),
    async fetchData() {
      // Reset view information.
      this.$store.commit("setReload", false);
      this.$store.commit("resetSelected");
      this.$store.commit("multiple", false);
      this.$store.commit("closeHovers");

      // Set loading to true and reset the error.
      this.setLoading(true);
      this.error = null;

      let url = this.$route.path;
      if (url === "") url = "/";
      if (url[0] !== "/") url = "/" + url;
      url = decodeURIComponent(url);

      try {
        await bdApi.fetchDir(url);
      } catch (e) {
        this.error = e;
      } finally {
        this.setLoading(false);
      }
    },
    logout: bdApi.logout,
  },
};
</script>
