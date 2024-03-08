<template>
  <div>
    <header-bar v-if="error || req.type == null" showMenu showLogo />

    <breadcrumbs base="/cephalon-cloud" />

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
    <!-- <copy-files :list="currentProgresses" :speed="speedMbyte" :eta="eta" /> -->
  </div>
</template>

<script>
import { cepApi } from "@/api";
import { mapState, mapMutations } from "vuex";
import HeaderBar from "@/components/header/HeaderBar.vue";
import Breadcrumbs from "@/components/Breadcrumbs.vue";
import Errors from "@/views/Errors.vue";
import Listing from "@/views/cephalonCloud/Listing.vue";
// import CopyFiles from "@/views/cephalonCloud/CopyFiles.vue";
function clean(path) {
  return path.endsWith("/") ? path.slice(0, -1) : path;
}

export default {
  name: "files",
  components: {
    HeaderBar,
    Breadcrumbs,
    Errors,
    // CopyFiles,
    Listing,
  },
  data: function () {
    return {
      error: null,
      width: window.innerWidth,
      // speedMbyte: 0,
      // eta: 0,
    };
  },
  computed: {
    ...mapState(["loading", "reload"]),
    ...mapState("cep", ["req", "reload", "loading"]),
    currentView() {
      if (this.loading) {
        return null;
      } else if (this.req.isDir) {
        return "listing";
      }
    },
    currentProgresses() {
      return Object.keys(this.progresses).map((name) => ({
        name,
        progress: this.progresses[name].progress,
      }));
    },
  },
  created() {
    this.$store.commit("setHandlingType", "CephalonCloud");
    this.fetchData();
  },
  watch: {
    $route: "fetchData",
    reload: function (value) {
      if (value === true) {
        this.fetchData();
      }
    },
  },
  mounted() {
    window.addEventListener("keydown", this.keyEvent);
  },
  beforeDestroy() {
    window.removeEventListener("keydown", this.keyEvent);
  },
  destroyed() {
    if (this.$store.state.showShell) {
      this.$store.commit("toggleShell");
    }
    this.$store.commit("updateRequest", {});
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

      try {
        await cepApi.fetchDir(url);
      } catch (e) {
        this.error = e;
      } finally {
        this.setLoading(false);
      }
    },
    keyEvent(event) {
      // F1!
      if (event.keyCode === 112) {
        event.preventDefault();
        this.$store.commit("showHover", "help");
      }
    },

    // fetchProgress: async function () {
    //   // Reset view information.
    //   this.$store.commit("bd/setRefreshCopy", false);

    //   try {
    //     const res = await bdApi.fetchProgress();
    //     const names = Object.keys(res); // 文件或文件夹

    //     // 从百度网盘下载到 My Files 的进度
    //     let progresses = {},
    //       copyBytes = 0,
    //       totalBytes = 0;
    //     for (let n of names) {
    //       let { percentage, size_b: size } = res[n]; // 返回的 size 是后端以 固定大小如 10MB 对文件切片后的次数
    //       if (percentage >= 1) continue;

    //       let name = n.split("/").slice(-1)[0];
    //       progresses[name] = {
    //         progress: percentage * 100,
    //         size,
    //       };
    //       copyBytes += percentage * size;
    //       totalBytes += size;
    //     }

    //     this.progresses = progresses;

    //     if (Object.keys(progresses).length > 0) {
    //       this.hasProgress = true;
    //       if (!this.hasStarted) {
    //         this.lastTimestamp = Date.now();
    //         this.prevBytes = copyBytes;
    //         this.hasStarted = true;
    //       }

    //       window.clearTimeout(this.timer);
    //       this.timer = window.setTimeout(() => {
    //         this.fetchProgress();
    //         this.calcProgress(copyBytes, totalBytes);
    //       }, 1500);
    //     } else if (this.hasProgress) {
    //       this.speedMbyte = 0;
    //       this.eta = 0;
    //       window.clearTimeout(this.timer);
    //       this.hasProgress = false;
    //       this.hasStarted = false;
    //       this.recentSpeeds = [];
    //       this.lastTimestamp = 0;
    //       this.prevBytes = 0;

    //       this.$showSuccess(this.$t("success.filesCopied"));
    //     }
    //   } catch (e) {
    //     this.$showError(e);
    //   }
    // },
    // calcProgress(copyBytes, totalBytes) {
    //   let elapsedTime = (Date.now() - this.lastTimestamp) / 1000;
    //   let lastCopyBytes = copyBytes - this.prevBytes;
    //   let currentSpeed = lastCopyBytes / (1024 * 1024) / elapsedTime;

    //   if (this.recentSpeeds.length >= 10) {
    //     this.recentSpeeds.shift();
    //   }
    //   this.recentSpeeds.push(currentSpeed);

    //   let avgSpeed =
    //     this.recentSpeeds.reduce((sum, item) => sum + item) /
    //     this.recentSpeeds.length;

    //   this.speedMbyte = avgSpeed;
    //   this.eta =
    //     this.speedMbyte === 0
    //       ? Infinity
    //       : (totalBytes - copyBytes) / (1024 * 1024) / this.speedMbyte;

    //   this.prevBytes = copyBytes;
    //   this.lastTimestamp = Date.now();
    // },
    // windowsResize: throttle(function () {
    //   this.width = window.innerWidth;
    // }, 100),
  },
};
</script>
