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

    <copy-files :list="currentProgresses" />
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
      max: 2,
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
        speed: this.cep.list[name]?.speed,
        remain: this.cep.list[name]?.remain,
        lastLoad: this.cep.list[name]?.lastLoad,
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
    // 通道被占用的数量就是当前list中process>0的，sort（A,B=>b.process-a.process）
    // 排序之后自然就会先请求已经占用通道的，就不会导致同时salve下载的文件过多，之后就是递归的逻辑
    async checkLoadList() {
      try {
        let flag = false;
        // 判断在local是否存有未完成请求
        let tempList = localStorage.getItem("list");
        // 若有
        if (tempList && tempList !== "{}") {
          tempList = JSON.parse(tempList);
          // 存入vuex
          store.commit("cep/setList", tempList);
          // 循环请求每一个
          let values = Object.values(this.cep.list).sort(
            (a, b) => b.process - a.process
          );
          const download = async () => {
            values = Object.values(this.cep.list).sort(
              (a, b) => b.process - a.process
            );
            // 有未下载的，且通道大于0
            // 只有一个未下载完成的情况下，会请求两次，因为满足通道，直到2被终止
            // while (values.length > 0 && this.cep.max > 0) {
            //   values = Object.values(this.cep.list).sort(
            //     (a, b) => b.process - a.process
            //   );
            store.commit("cep/setCanStop", true);
            for (let i = 0; i < values.length && this.cep.max > 0; i++) {
              flag = false;
              const item = values[i];

              if (this.cep.list[item.name] && !this.cep.list[item.name].sse) {
                store.commit("cep/changeMax", this.cep.max - 1);
                store.commit("cep/setListProgressAdd1", {
                  name: item.name,
                  value: 0.01,
                });
                const res = cepApi.fetchDownload({
                  // const res = await cepApi.fetchDownload({
                  md5: item.md5,
                  target: item.target,
                  filename: item.name,
                });

                if (this.cep.list[values[i].name]) {
                  // 正常201完成 设置进度到100
                  if ((await res).status === 201) {
                    store.commit("cep/changeMax", this.cep.max + 1);
                    store.commit("cep/setListProgressAdd1", {
                      name: item.name,
                      value: 100,
                    });
                    // setTimeout(() => {
                    store.commit("cep/deleteListItem", item.name);
                    local.deleteListItem(item.name);
                    // this.max++;
                    let values = Object.values(this.cep.list).filter(
                      (item) => item.process == 0
                    );

                    if (JSON.stringify(this.cep.list) === "{}") {
                      // 关闭进度对象展示
                      store.commit("cep/setCanStop", false);
                      // 清空VUEX中的进度条对象
                      store.commit("cep/setList", {});

                      // 提示成功
                      // flag用作控制提示是否弹出成功 当出现404时且该次请求是最后一条 则不现实成功提示 避免歧义
                      // 每次请求会将flag赋值为false 404将赋值为true
                      if (!flag)
                        this.$showSuccess(this.$t("success.filesCopied"));
                    }
                    if (values.length > 0 && this.cep.max == 1) {
                      download();
                    }
                    // }, 100);
                  } else if ((await res).status === 302) {
                    local.changeListItemstatus(item.name);
                    let speedBox = [];
                    // 开启获取进度
                    // await new Promise((resolve) => {
                    let sseClient = this.$sse.create({
                      url: `${baseURL}/api/cd/download/size?stream=${item.md5}`,
                      format: "json",
                      withCredentials: true,
                      polyfill: true,
                    });
                    let timer = setInterval(() => {
                      // 设置一个定时循环检测此进度条是否被删除，如果传输中被删除，则结束并进行下一个
                      // 为什么一定要在这里disconnect：如果在CopyFiles断开连接，则这边的onMessage不会触发
                      // 则无法resolve，会卡在这里，不会执行下一次循环
                      if (!this.cep.list[item.name]) {
                        store.commit("cep/changeMax", this.cep.max + 1);
                        if (values.length > 0 && this.cep.max == 1) {
                          download();
                        }
                        sseClient.disconnect();
                        clearInterval(timer);
                        // resolve();
                      }
                    }, 2000);

                    sseClient.connect().then(() => {
                      store.commit("cep/setListItemSSE", {
                        name: item.name,
                        sse: sseClient,
                      });

                      // 建立连接 onmessage
                      sseClient.on("message", async (msg) => {
                        let speed =
                          Math.abs(msg - this.cep.list[item.name]?.lastLoad) /
                          2;
                        if (speedBox.length == 5) speedBox.shift();
                        speedBox.push(speed);
                        let showSpeed;
                        let sum = 0;
                        speedBox.forEach((item) => {
                          sum += item;
                        });
                        showSpeed = sum / speedBox.length;
                        let percent = msg / item.size;
                        let remain =
                          (item.size - msg) / showSpeed < 0
                            ? 0
                            : (item.size - msg) / showSpeed;
                        store.commit("cep/setListLastLoad", {
                          name: item.name,
                          value: msg,
                        });
                        store.commit("cep/setListProgressAdd1", {
                          name: item.name,
                          value: percent * 100,
                        });

                        store.commit("cep/setListSpeed", {
                          name: item.name,
                          value: (showSpeed / 1024 / 1024).toFixed(2),
                        });

                        store.commit("cep/setListRemain", {
                          name: item.name,
                          value: remain,
                        });

                        if (percent === 1 || msg == -2) {
                          // sseClient.disconnect();
                          store.commit("cep/disconnectSSE", item.name);
                          // 再次下载请求
                          await cepApi.fetchDownload({
                            md5: item.md5,
                            target: item.target,
                            filename: item.name,
                          });
                          // 然后从loadList中删除
                          setTimeout(() => {
                            store.commit("cep/deleteListItem", item.name);
                            clearInterval(timer);
                            local.deleteListItem(item.name);
                            store.commit("cep/changeMax", this.cep.max + 1);
                            let values = Object.values(this.cep.list).filter(
                              (item) => item.process == 0
                            );

                            if (JSON.stringify(this.cep.list) === "{}") {
                              // 关闭进度对象展示
                              store.commit("cep/setCanStop", false);
                              // 清空VUEX中的进度条对象
                              store.commit("cep/setList", {});

                              // 提示成功
                              // flag用作控制提示是否弹出成功 当出现404时且该次请求是最后一条 则不现实成功提示 避免歧义
                              // 每次请求会将flag赋值为false 404将赋值为true
                              if (!flag)
                                this.$showSuccess(
                                  this.$t("success.filesCopied")
                                );
                            }
                            if (values.length > 0 && this.cep.max == 1) {
                              download();
                            }
                          }, 100);

                          // resolve();
                        }
                      });
                    });
                    // });
                  } else if (res.status === 404) {
                    this.$showError(
                      {
                        message:
                          item.name + ": " + res.status + " " + res.statusText,
                      },
                      false
                    );
                    flag = true;
                  }
                }
              }
            }
            // }
          };

          download();
        }
      } catch (e) {
        console.log(e);
      }
    },
  },
};
</script>
