<template>
  <div>
    <router-view></router-view>
  </div>
</template>

<script>
// eslint-disable-next-line no-undef
// __webpack_public_path__ = window.FileBrowser.StaticURL + "/";

export default {
  name: "app",
  mounted() {
    console.log(process.env);
    const loading = document.getElementById("loading");
    loading.classList.add("done");

    setTimeout(function () {
      loading.parentNode.removeChild(loading);
    }, 200);
  },
};
</script>

<style>
@import "./css/styles.css";
</style>
