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
          <span v-if="!isMobile">{{ $t("baiduNetdisk.unbind") }}</span>
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

    <copy-files :list="currentProgresses" :speed="speedMbyte" :eta="eta" />
  </div>
</template>

<script>
import { bdApi } from "@/api";
import { mapState, mapMutations } from "vuex";
import throttle from "lodash.throttle";

import HeaderBar from "@/components/header/HeaderBar.vue";
import Breadcrumbs from "@/components/Breadcrumbs.vue";
import Errors from "@/views/Errors.vue";
import AuthLogin from "@/views/baiduNetdisk/AuthLogin.vue";
import Listing from "@/views/baiduNetdisk/Listing.vue";
import ProgressBar from "vue-simple-progress";
import CopyFiles from "@/views/baiduNetdisk/CopyFiles.vue";

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
    CopyFiles,
  },
  data: function () {
    return {
      error: null,
      width: window.innerWidth,
      progresses: {},
      speedMbyte: 0,
      eta: 0,
      // tmp
      timer: null,
      hasProgress: false,
      hasStarted: false,
      recentSpeeds: [],
      lastTimestamp: 0,
      prevBytes: 0,
    };
  },
  computed: {
    ...mapState(["loading", "reload"]),
    ...mapState("bd", ["user", "req", "refreshCopy"]),
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
    currentProgresses() {
      return Object.keys(this.progresses).map((name) => ({
        name,
        progress: this.progresses[name].progress,
      }));
    },
    isMobile() {
      return this.width <= 736;
    },
  },
  async created() {
    try {
      this.setLoading(true);
      this.$store.commit("setHandlingType", "BaiduNetdisk");
      await bdApi.getAccessToken();
      const at = bdApi.getToken();
      if (at) {
        await bdApi.fetchUserInfo();
      }
    } catch (e) {
      console.log(e);
    } finally {
      this.setLoading(false);
    }
  },
  destroyed() {
    this.$store.commit("setHandlingType", "");
    this.$store.commit("bd/updateReq", {});
    window.clearTimeout(this.timer);
  },
  watch: {
    $route: "fetchData",
    reload: async function (value) {
      if (value === true) {
        await this.fetchData();
        await this.fetchProgress();
      }
    },
    refreshCopy: async function (value) {
      if (value === true) {
        this.hasProgress = true;
        await this.fetchProgress();
      }
    },
  },
  mounted: function () {
    // Add the needed event listeners to the window and document.
    window.addEventListener("resize", this.windowsResize);
  },
  beforeDestroy() {
    // Remove event listeners before destroying this page.
    window.removeEventListener("resize", this.windowsResize);
  },
  methods: {
    ...mapMutations(["setLoading"]),
    async fetchData() {
      if (!this.user) return;
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
        if (e.status === 401) {
          bdApi.logout();
          this.$showError(this.$t("baiduNetdisk.authExpired"), false, 1500);
        } else {
          this.error = e;
        }
      } finally {
        this.setLoading(false);
      }
    },
    logout: bdApi.logout,
    fetchProgress: async function () {
      // Reset view information.
      this.$store.commit("bd/setRefreshCopy", false);

      try {
        const res = await bdApi.fetchProgress();
        const names = Object.keys(res); // 文件或文件夹

        // 从百度网盘下载到 My Files 的进度
        let progresses = {},
          copyBytes = 0,
          totalBytes = 0;
        for (let n of names) {
          let { percentage, size_b: size } = res[n]; // 返回的 size 是后端以 固定大小如 10MB 对文件切片后的次数
          if (percentage >= 1) continue;

          let name = n.split("/").slice(-1)[0];
          progresses[name] = {
            progress: percentage * 100,
            size,
          };
          copyBytes += percentage * size;
          totalBytes += size;
        }

        this.progresses = progresses;

        if (Object.keys(progresses).length > 0) {
          this.hasProgress = true;
          if (!this.hasStarted) {
            this.lastTimestamp = Date.now();
            this.prevBytes = copyBytes;
            this.hasStarted = true;
          }

          window.clearTimeout(this.timer);
          this.timer = window.setTimeout(() => {
            this.fetchProgress();
            this.calcProgress(copyBytes, totalBytes);
          }, 1500);
        } else if (this.hasProgress) {
          this.speedMbyte = 0;
          this.eta = 0;
          window.clearTimeout(this.timer);
          this.hasProgress = false;
          this.hasStarted = false;
          this.recentSpeeds = [];
          this.lastTimestamp = 0;
          this.prevBytes = 0;

          this.$showSuccess(this.$t("success.filesCopied"));
        }
      } catch (e) {
        this.$showError(e);
      }
    },
    calcProgress(copyBytes, totalBytes) {
      let elapsedTime = (Date.now() - this.lastTimestamp) / 1000;
      let lastCopyBytes = copyBytes - this.prevBytes;
      let currentSpeed = lastCopyBytes / (1024 * 1024) / elapsedTime;

      if (this.recentSpeeds.length >= 10) {
        this.recentSpeeds.shift();
      }
      this.recentSpeeds.push(currentSpeed);

      let avgSpeed =
        this.recentSpeeds.reduce((sum, item) => sum + item) /
        this.recentSpeeds.length;

      this.speedMbyte = avgSpeed;
      this.eta =
        this.speedMbyte === 0
          ? Infinity
          : (totalBytes - copyBytes) / (1024 * 1024) / this.speedMbyte;

      this.prevBytes = copyBytes;
      this.lastTimestamp = Date.now();
    },
    windowsResize: throttle(function () {
      this.width = window.innerWidth;
    }, 100),
  },
};
</script>
