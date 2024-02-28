<template>
  <div class="dashboard">
    <header-bar showLogo />

    <input
      class="input input--block"
      type="text"
      @keyup.enter="submit"
      v-model.trim="md5"
      placeholder="md5"
    />
    <input
      class="input input--block"
      type="text"
      @keyup.enter="submit"
      v-model.trim="filename"
      placeholder="filename"
    />
    <button class="button" @click="submit">提交</button>
  </div>
</template>

<script>
import { test as api } from "@/api";
import HeaderBar from "@/components/header/HeaderBar.vue";

export default {
  name: "test",
  components: {
    HeaderBar,
  },
  data: function () {
    return {
      md5: "",
      filename: "",
    };
  },
  methods: {
    submit: async function (event) {
      event.preventDefault();

      try {
        await api.cdDownload({ md5: this.md5, filename: this.filename });
        // this.$showSuccess("success");
      } catch (e) {
        console.log(e);
        this.$showError(e);
      }
    },
  },
};
</script>
