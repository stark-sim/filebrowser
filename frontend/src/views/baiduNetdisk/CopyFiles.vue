<template>
  <div
    v-if="list && list.length > 0"
    class="upload-files netdisk-copy-files"
    v-bind:class="{ closed: !open }"
  >
    <div class="card floating">
      <div class="card-title">
        <h2>{{ $t("prompts.copyFiles", { files: list.length }) }}</h2>
        <div class="upload-info">
          <div class="upload-speed">{{ speed.toFixed(2) }} MB/s</div>
          <div class="upload-eta">{{ formattedETA }} remaining</div>
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
          v-for="file in list"
          :key="file.name"
          :data-dir="file.isDir"
          :data-type="file.type"
          :aria-label="file.name"
        >
          <div class="content">
            <div class="file-name" :title="file.name">
              <i class="material-icons"></i> {{ file.name }}
            </div>
            <div class="file-progress">
              <div v-bind:style="{ width: file.progress + '%' }"></div>
            </div>
          </div>
          <div class="operate">
            <button
              class="action"
              @click="abortUpload(file)"
              aria-label="Abort File Upload"
              title="Abort File Upload"
            >
              <i class="material-icons">{{ "cancel" }}</i>
            </button>
            <button
              v-if="file.stopped"
              class="action"
              @click="continueUpload(file)"
              aria-label="Continue File Upload"
              title="Continue File Upload"
            >
              <i class="material-icons">{{ "play_circle" }}</i>
            </button>
            <button
              v-else
              class="action"
              @click="stopUpload(file)"
              aria-label="Stop File Upload"
              title="Stop File Upload"
            >
              <i class="material-icons">{{ "pause_circle" }}</i>
            </button>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>

<script>
import { bdApi } from "@/api";
import buttons from "@/utils/buttons";

export default {
  name: "uploadFiles",
  props: ["list", "speed", "eta"],
  data: function () {
    return {
      open: false,
    };
  },
  computed: {
    formattedETA() {
      if (!this.eta || this.eta === Infinity) {
        return "--:--:--";
      }

      let totalSeconds = this.eta;
      const hours = Math.floor(totalSeconds / 3600);
      totalSeconds %= 3600;
      const minutes = Math.floor(totalSeconds / 60);
      const seconds = Math.round(totalSeconds % 60);

      return `${hours.toString().padStart(2, "0")}:${minutes
        .toString()
        .padStart(2, "0")}:${seconds.toString().padStart(2, "0")}`;
    },
  },
  methods: {
    toggle: function () {
      this.open = !this.open;
    },
    async abortUpload({ path }) {
      if (confirm(this.$t("upload.abortUpload"))) {
        try {
          await bdApi.cancelProgress({ file_name: path });
          this.$showSuccess(this.$t("success.uploadAborted"));
          this.$emit("fetchProgress");
        } catch (e) {
          this.$showError(e?.message || JSON.stringify(e), false, 1500);
        }
      }
    },
    async continueUpload({ path }) {
      try {
        await bdApi.continueProgress({ file_name: path });
        console.log("HHH");
        this.$showSuccess(this.$t("success.uploadContinued"));
        this.$emit("fetchProgress");
      } catch (e) {
        this.$showError(e?.message || JSON.stringify(e), false, 1500);
      }
    },
    async stopUpload({ path }) {
      try {
        await bdApi.stopProgress({ file_name: path });
        this.$showSuccess(this.$t("success.uploadStopped"));
        this.$emit("fetchProgress");
      } catch (e) {
        this.$showError(e?.message || JSON.stringify(e), false, 1500);
      }
    },
    abortAll() {
      if (confirm(this.$t("upload.abortUpload"))) {
        buttons.done("copy");
        this.open = false;
      }
    },
  },
};
</script>

<style scoped>
.card-content .file {
  display: flex;
  gap: 10px;
}

.card-content .file .content {
  flex: 1;
  overflow: hidden;
}

.card-content .file .content .file-name {
  justify-content: flex-start;
}

.card-content .file .operate {
  display: flex;
  align-items: center;
}

.card-content .file .operate .action {
  display: inline-flex;
}

.card-content .file .operate .action:disabled i {
  color: gray;
}

.card-content .file .operate .action i {
  padding: 0;
}

.card-content .file .operate .action i::before {
  content: "";
}
</style>
