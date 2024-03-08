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
      console.log(this.selected);
      try {
        buttons.loading("copy");
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
        // let temp = [];
        console.log(items);
        for (let item of items) {
          await cepApi.fetchDownload(item);
        }

        // 由于调下载接口再查询 progress 有延迟
        // window.setTimeout(() => {
        //   this.$store.commit("bd/setRefreshCopy", true);
        // }, 1000);
      } catch (e) {
        buttons.done("copy");
        console.log("cep download error:", e);
      } finally {
        buttons.success("copy");
        this.$store.commit("resetSelected");
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
