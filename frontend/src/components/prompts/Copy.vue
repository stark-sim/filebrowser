<template>
  <div class="card floating">
    <div class="card-title">
      <h2>{{ $t("prompts.copy") }}</h2>
    </div>

    <div class="card-content">
      <p>{{ $t("prompts.copyMessage") }}</p>
      <file-list ref="fileList" @update:selected="(val) => (dest = val)">
      </file-list>
    </div>

    <div
      class="card-action"
      :style="
        showNewFolder
          ? 'display: flex; align-items: center; justify-content: space-between'
          : ''
      "
    >
      <template v-if="showNewFolder">
        <button
          class="button button--flat"
          @click="$refs.fileList.createDir()"
          :aria-label="$t('sidebar.newFolder')"
          :title="$t('sidebar.newFolder')"
          style="justify-self: left"
        >
          <span>{{ $t("sidebar.newFolder") }}</span>
        </button>
      </template>
      <div>
        <button
          class="button button--flat button--grey"
          @click="$store.commit('closeHovers')"
          :aria-label="$t('buttons.cancel')"
          :title="$t('buttons.cancel')"
        >
          {{ $t("buttons.cancel") }}
        </button>
        <button
          class="button button--flat"
          @click="copyByType"
          :aria-label="$t('buttons.copy')"
          :title="$t('buttons.copy')"
        >
          {{ $t("buttons.copy") }}
        </button>
      </div>
    </div>
  </div>
</template>

<script>
import { mapState } from "vuex";
import FileList from "./FileList.vue";
import { files as api, bdApi, cepApi } from "@/api";
import buttons from "@/utils/buttons";
import * as upload from "@/utils/upload";
import { removePrefix } from "@/api/utils";
import store from "@/store";
export default {
  name: "copy",
  components: { FileList },
  data: function () {
    return {
      current: window.location.pathname,
      dest: null,
    };
  },
  computed: {
    ...mapState(["req", "selected", "user", "handlingType", "bd", "cep"]),
    showNewFolder() {
      return this.user.perm.create && this.$route.path.includes("/files");
    },
  },
  methods: {
    copyByType(event) {
      event.preventDefault();
      if (this.handlingType === "BaiduNetdisk") {
        this.bdDownload();
      } else if (this.handlingType === "CephalonCloud") {
        this.cepDownload();
      } else {
        this.copy();
      }
    },
    cepDownload: async function () {
      try {
        buttons.loading("copy");
        // 获取选中的文件信息，下载请求需要的参数
        let items = [],
          target_path = `.${removePrefix(decodeURIComponent(this.dest))}`;
        for (let item of this.selected) {
          const { name, md5 } = this.cep.req.items[item];
          items.push({
            md5,
            target: target_path,
            filename: name,
          });
        }

        // 声明进度条对象
        const loadList = {};
        // 初始化进度条对象
        for (let item of this.selected) {
          const { size, name, md5 } = this.cep.req.items[item];
          let temp = {
            size,
            name,
            process: 0,
            canStop: false,
            md5,
            target: target_path,
          };
          loadList[name] = temp;
        }
        let templist = loadList;
        const values = Object.values(templist);
        Object.assign(loadList, this.cep.list);

        // 将进度条对象放到vuex
        store.commit("cep/setList", loadList);
        // 关闭弹窗
        this.$store.commit("closeHovers");
        // 取消选中
        this.$store.commit("resetSelected");
        // 像进度条对象展示
        store.commit("cep/setCanStop", true);

        // 开始请求
        for (let i = 0; i < values.length; i++) {
          const item = values[i];
          const res = await cepApi.fetchDownload({
            md5: item.md5,
            target: target_path,
            filename: item.name,
          });

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
                url: `/api/cd/download/size?stream=${item.md5}`,
                format: "json",
                withCredentials: true,
                polyfill: true,
              });

              sseClient.connect().then((sse) => {
                console.log("linking", sse);
              });
              // 建立连接 onmessage
              sseClient.on("message", async (msg) => {
                let percent = msg / item.size;

                store.commit("cep/setListProgressAdd1", {
                  name: item.name,
                  value: percent * 100,
                });

                if (percent === 1) {
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
          }
          // 然后从loadList中删除？
          setTimeout(() => {
            store.commit("cep/deleteListItem", item.name);
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
      } catch (e) {
        if (e.status === 403)
          this.$showError(
            { message: this.$t("errors.forbidden") },
            false,
            1500
          );
        else if (e.status === 500)
          this.$showError({ message: this.$t("errors.internal") }, false, 1500);
        else this.$showError(e);
        // else this.$showError(e?.message || e);

        buttons.done("copy");
        console.log("cep download error:", e);
      } finally {
        buttons.success("copy");
        // this.$store.commit("resetSelected");
        this.$store.commit("closeHovers");
      }
    },
    bdDownload: async function () {
      try {
        buttons.loading("copy");
        let items = [],
          target_path = `.${removePrefix(decodeURIComponent(this.dest))}`;
        for (let item of this.selected) {
          const { path, isDir: is_dir, fsId: fs_id } = this.bd.req.items[item];
          items.push({
            path,
            is_dir,
            fs_id,
            target_path,
          });
        }
        for (let item of items) {
          await bdApi.fetchDownload(item);
        }

        // 由于调下载接口再查询 progress 有延迟
        window.setTimeout(() => {
          this.$store.commit("bd/setRefreshCopy", true);
        }, 1000);
      } catch (e) {
        buttons.done("copy");
        console.log("bd download error:", e);
      } finally {
        buttons.success("copy");
        this.$store.commit("resetSelected");
        this.$store.commit("closeHovers");
      }
    },
    copy: async function () {
      let items = [];

      // Create a new promise for each file.
      for (let item of this.selected) {
        items.push({
          from: this.req.items[item].url,
          to: this.dest + encodeURIComponent(this.req.items[item].name),
          name: this.req.items[item].name,
        });
      }

      let action = async (overwrite, rename) => {
        buttons.loading("copy");

        await api
          .copy(items, overwrite, rename)
          .then(() => {
            buttons.success("copy");

            if (this.$route.path === this.dest) {
              this.$store.commit("setReload", true);

              return;
            }

            this.$router.push({ path: this.dest });
          })
          .catch((e) => {
            buttons.done("copy");
            this.$showError(e);
          });
      };

      if (this.$route.path === this.dest) {
        this.$store.commit("closeHovers");
        action(false, true);

        return;
      }

      let dstItems = (await api.fetch(this.dest)).items;
      let conflict = upload.checkConflict(items, dstItems);

      let overwrite = false;
      let rename = false;

      if (conflict) {
        this.$store.commit("showHover", {
          prompt: "replace-rename",
          confirm: (event, option) => {
            overwrite = option == "overwrite";
            rename = option == "rename";

            event.preventDefault();
            this.$store.commit("closeHovers");
            action(overwrite, rename);
          },
        });
        return;
      }
      action(overwrite, rename);
    },
  },
};
</script>
