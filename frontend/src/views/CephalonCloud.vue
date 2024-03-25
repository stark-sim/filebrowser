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

    <copy-files :list="currentProgresses" :progress="progress" />
  </div>
</template>

<script>
import { cepApi } from "@/api";
import { mapState, mapMutations } from "vuex";
import HeaderBar from "@/components/header/HeaderBar.vue";
import Breadcrumbs from "@/components/Breadcrumbs.vue";
import Errors from "@/views/Errors.vue";
import Listing from "@/views/cephalonCloud/Listing.vue";
import CopyFiles from "@/views/cephalonCloud/CopyFiles.vue";
import { nextTick } from "vue";
import store from "@/store";
import * as local from "@/utils/local";
import { baseURL } from "@/utils/constants";

export default {
  name: "files",
  components: {
    HeaderBar,
    Breadcrumbs,
    Errors,
    CopyFiles,
    Listing,
  },
  data: function () {
    return {
      error: null,
      // speedMbyte: 0,
      // eta: 0,
    };
  },
  computed: {
    ...mapState(["loading", "reload", "cep"]),
    ...mapState("cep", ["req", "progress", "list"]),
    currentView() {
      if (this.req) {
        return "listing";
      } else {
        // this.loading
        return null;
      }
    },
    currentProgresses() {
      let list = Object.keys(this.cep.list).map((name) => ({
        name,
        process: this.cep.list[name]?.process,
      }));
      return list;
    },
  },
  async created() {
    await nextTick();
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
    list: {
      handler: function (value) {
        console.log(2, value);
      },
      immediate: true,
      deep: true,
    },
  },

  mounted() {
    window.addEventListener("keydown", this.keyEvent);
    this.checkLoadList();
  },
  beforeDestroy() {
    window.removeEventListener("keydown", this.keyEvent);
  },
  destroyed() {
    if (this.$store.state.showShell) {
      this.$store.commit("toggleShell");
    }
    this.$store.commit("setHandlingType", "");
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
    // 刷新重新下载未完成的
    async checkLoadList() {
      try {
        // 判断在local是否存有未完成请求
        let tempList = localStorage.getItem("list");
        // 若有
        if (tempList && tempList !== "{}") {
          tempList = JSON.parse(tempList);
          // 存入vuex
          store.commit("cep/setList", tempList);
          // 循环请求每一个
          const values = Object.values(tempList);
          store.commit("cep/setCanStop", true);
          for (let i = 0; i < values.length; i++) {
            const item = values[i];
            if (this.cep.list[values[i].name]) {
              const res = await cepApi.fetchDownload({
                md5: item.md5,
                target: item.target,
                filename: item.name,
              });

              if (this.cep.list[values[i].name]) {
                // 正常201完成 设置进度到100
                if (res.status === 201)
                  store.commit("cep/setListProgressAdd1", {
                    name: item.name,
                    value: 100,
                  });
                else if (res.status === 302) {
                  // 开启获取进度
                  await new Promise((resolve) => {
                    let sseClient = this.$sse.create({
                      url: `${baseURL}/api/cd/download/size?stream=${item.md5}`,
                      format: "json",
                      withCredentials: true,
                      polyfill: true,
                    });
                    sseClient.connect().then((sse) => {
                      store.commit("cep/setListItemSSE", {
                        name: item.name,
                        sse: sseClient,
                      });

                      // 建立连接 onmessage
                      sseClient.on("message", async (msg) => {
                        let percent = msg / item.size;

                        store.commit("cep/setListProgressAdd1", {
                          name: item.name,
                          value: percent * 100,
                        });

                        if (percent === 1 || msg == -2) {
                          sseClient.disconnect();
                          // 再次下载请求
                          await cepApi.fetchDownload({
                            md5: item.md5,
                            target: target_path,
                            filename: item.name,
                          });
                          resolve();
                        }
                      });
                    });
                  });
                }

                // 然后从loadList中删除
                setTimeout(() => {
                  store.commit("cep/deleteListItem", item.name);
                  local.deleteListItem(item.name);
                  if (JSON.stringify(this.cep.list) === "{}") {
                    // 关闭进度对象展示
                    store.commit("cep/setCanStop", false);
                    // // 清空VUEX中的进度条对象
                    store.commit("cep/setList", {});

                    // // 提示成功
                    this.$showSuccess(this.$t("success.filesCopied"));
                  }
                }, 100);
              }
            }
          }
        }
      } catch (e) {
        console.log(e);
      }
    },
  },
};
</script>
