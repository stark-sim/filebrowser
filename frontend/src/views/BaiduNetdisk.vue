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

    <copy-files
      :list="currentProgresses"
      :speed="speedMbyte"
      :eta="eta"
      @resetPrevBytes="resetPrevBytes"
      @fetchProgress="fetchProgress"
    />
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
import { nextTick } from "vue";

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
      progressLoading: false,
      progresses: {}, // 所有文件的下载大小
      deleteProgresses: {}, // 删掉下载成功的文件
      speedMbyte: 0,
      eta: 0,
      // tmp
      timer: null,
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
      return Object.keys(this.progresses)
        .filter((name) => this.progresses[name].progress < 100)
        .map((name) => ({
          path: name,
          name: name.split("/").slice(-1)[0],
          ...this.progresses[name],
        }));
    },
    isMobile() {
      return this.width <= 736;
    },
  },
  async created() {
    try {
      this.setLoading(true);
      await nextTick();
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
        if (this.progressLoading) return;
        this.progressLoading = true;
        const res = await bdApi.fetchProgress();
        const names = Object.keys(res); // 文件或文件夹

        // 从百度网盘下载到 My Files 的进度
        let progresses = {},
          copyBytes = 0,
          totalBytes = 0,
          count = 0, // 是否有正在上传/暂停的文件
          successCount = 0, // 成功的文件
          stopCount = 0,
          errStopCount = 0; // 跟上一个存的暂停状态比较，上一个未暂停，但下一个错误暂停，需要提示
        for (let n of names) {
          /**
           * 返回字段说明：
           * size_b 文件大小，单位 Byte
           * is_err 判定 is_stop 是被动还是主动关闭
           * is_small 表示为小文件，只能中断，无法暂停/恢复
           */
          let {
            percentage,
            size_b: size,
            is_err: isError,
            is_stop: isStop,
          } = res[n];

          if (percentage >= 1) {
            if (!Object.keys(this.deleteProgresses).includes(n)) {
              this.deleteProgresses[n] = false;
            }
            successCount++;
          } else if (isStop) {
            const hasName = Object.keys(this.progresses).find((k) => k === n);
            if (
              isError &&
              ((hasName && !this.progresses[n].stopped) || !hasName)
            ) {
              errStopCount++; // 提示可以重新恢复上传
            }
            stopCount++;
            count++;
          } else {
            count++;
          }

          progresses[n] = {
            progress: percentage * 100,
            size,
            stopped: isStop,
          };
          copyBytes += percentage * size;
          totalBytes += size;
        }

        this.progresses = progresses;
        this.deleteProgress(); // 需要先展示再有删除

        if (successCount > 0) {
          this.$showSuccess(this.$t("success.filesCopied"));
        }

        if (count > 0) {
          if (!this.hasStarted) {
            this.lastTimestamp = Date.now();
            this.prevBytes = copyBytes;
            this.hasStarted = true;
          }

          if (errStopCount > 0) {
            this.$showError(this.$t("errors.uploadingError"), false);
          }

          if (count === stopCount) return;
          window.clearTimeout(this.timer);
          this.timer = window.setTimeout(() => {
            this.fetchProgress();
            this.calcProgress(copyBytes, totalBytes);
          }, 1500);
        } else {
          // 下载结束需要清除数据
          this.speedMbyte = 0;
          this.eta = 0;
          window.clearTimeout(this.timer);
          this.hasStarted = false;
          this.recentSpeeds = [];
          this.lastTimestamp = 0;
          this.prevBytes = 0;
        }
      } catch (e) {
        if (e.status === 502) {
          this.$showError(this.$t("errors.uploadingRetry"), false);
        } else {
          this.$showError(e?.message || JSON.stringify(e), false);
        }
      } finally {
        this.progressLoading = false;
      }
    },
    resetPrevBytes(size, progress) {
      /**
       * copyBytes 的计算方式有二：
       * 1. 累加的已复制字节数必须包括已经下载好的数据，
       *    否则会出现当前已复制字节数 < 上一个缓存已复制字节数的情况致使出现负数速度
       * 2. 将上一个缓存已复制字节数也减去已经下载好的数据的字节数（采用 √）
       */
      this.prevBytes -= (size * progress) / 100;
    },
    deleteProgress: async function () {
      const names = Object.keys(this.deleteProgresses);
      const key = names.find((n) => this.deleteProgresses[n] === false);
      if (!key) return;
      try {
        this.deleteProgresses[key] = true;
        await bdApi.deleteProgress({ file_name: key });
        const hasName = Object.keys(this.progresses).find((k) => k === key);
        if (hasName) {
          const { size, progress } = this.progresses[key];
          this.resetPrevBytes(size, progress);
        }
        delete this.deleteProgresses[key];
        this.deleteProgress();
      } catch (e) {
        this.$showError(e?.message || JSON.stringify(e), false);
      } finally {
        if (this.deleteProgresses[key] === true) {
          this.deleteProgresses[key] = false;
        }
      }
    },
    calcProgress(copyBytes, totalBytes) {
      let elapsedTime = (Date.now() - this.lastTimestamp) / 1000;
      let lastCopyBytes = copyBytes - this.prevBytes;
      let currentSpeed = lastCopyBytes / (1024 * 1024) / elapsedTime;

      if (this.recentSpeeds.length >= 5) {
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
