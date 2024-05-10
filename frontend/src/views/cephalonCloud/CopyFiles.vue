<template>
  <div
    v-if="canStop"
    class="upload-files netdisk-copy-files"
    v-bind:class="{ closed: !open }"
  >
    <div class="card floating">
      <div class="card-title">
        <h2>{{ $t("prompts.copyFiles", { files: list.length }) }}</h2>
        <div class="upload-info">
          <!-- <div class="upload-speed">{{ speed.toFixed(2) }} MB/s</div> -->
          <!-- <div class="upload-eta">{{ formattedETA }} remaining</div> -->
        </div>
        <button
          v-if="false"
          class="action"
          @click="abortAll"
          aria-label="Abort upload"
          title="Abort upload"
        >
          <i class="material-icons">{{ "cancel" }}</i>
        </button>
        <button
          class="action"
          @click="toggle"
          aria-label="Toggle file upload list"
          title="Toggle file upload list"
        >
          <i class="material-icons">{{
            open ? "keyboard_arrow_down" : "keyboard_arrow_up"
          }}</i>
        </button>
      </div>

      <div class="card-content file-icons">
        <div
          class="file"
          v-for="file in Object.values(list)"
          :key="file.name"
          :data-dir="file.isDir"
          :data-type="file.type"
          :aria-label="file.name"
        >
          <div class="file-name">
            <div>
              <i class="material-icons"></i> <span>{{ file.name }}</span>
              <div class="download-status" v-show="file.process != 0">
                {{ file.speed }}MB/s
                <span>{{ formattedETA(file.remain) }} remaining</span>
              </div>
            </div>
            <div class="close" @click="cancel(file)">✖️</div>
          </div>
          <div class="file-progress">
            <div v-bind:style="{ width: file.process + '%' }"></div>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>

<script>
import buttons from "@/utils/buttons";
import { mapState } from "vuex";
import store from "@/store";
import * as local from "@/utils/local";

export default {
  name: "uploadFiles",
  props: ["list", "speed", "eta"],
  data: function () {
    return {
      open: true,
    };
  },
  computed: {
    ...mapState(["cep"]),
    ...mapState("cep", ["canStop"]),
    // ...mapGetters("cep", ["isAllFin"]),
  },
  methods: {
    formattedETA(time) {
      if (!time || time === Infinity) {
        return "--:--:--";
      }

      let totalSeconds = time;
      const hours = Math.floor(totalSeconds / 3600);
      totalSeconds %= 3600;
      const minutes = Math.floor(totalSeconds / 60);
      const seconds = Math.round(totalSeconds % 60);

      return `${hours.toString().padStart(2, "0")}:${minutes
        .toString()
        .padStart(2, "0")}:${seconds.toString().padStart(2, "0")}`;
    },

    toggle: function () {
      this.open = !this.open;
    },
    abortAll() {
      if (confirm(this.$t("upload.abortUpload"))) {
        buttons.done("copy");
        this.open = false;
      }
    },
    cancel(file) {
      store.commit("cep/deleteListItem", file.name);
      local.deleteListItem(file.name);
      if (JSON.stringify(this.cep.list) == "{}") {
        store.commit("cep/setCanStop", false);
      }
    },
  },
};
</script>
