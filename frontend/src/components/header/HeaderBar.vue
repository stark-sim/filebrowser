<template>
  <header>
    <img v-if="showLogo !== undefined" :src="logoURL" />
    <span v-if="!isMobile && application?.length > 1" class="title"
      >「{{ application }}」 文件管理器</span
    >
    <action
      v-if="showMenu !== undefined"
      class="menu-button"
      icon="menu"
      :label="$t('buttons.toggleSidebar')"
      @action="openSidebar()"
    />

    <slot />

    <div id="dropdown" :class="{ active: this.currentPromptName === 'more' }">
      <slot name="actions" />
    </div>

    <action
      v-if="this.$slots.actions"
      id="more"
      icon="more_vert"
      :label="$t('buttons.more')"
      @action="$store.commit('showHover', 'more')"
    />

    <div
      class="overlay"
      v-show="this.currentPromptName == 'more'"
      @click="$store.commit('closeHovers')"
    />
  </header>
</template>

<script>
import { logoURL } from "@/utils/constants";
import throttle from "lodash.throttle";
import Action from "@/components/header/Action.vue";
import { mapGetters } from "vuex";

export default {
  name: "header-bar",
  props: ["showLogo", "showMenu"],
  components: {
    Action,
  },
  data: function () {
    return {
      logoURL,
      application: "",
      width: window.innerWidth,
    };
  },
  methods: {
    openSidebar() {
      this.$store.commit("showHover", "sidebar");
    },
    windowsResize: throttle(function () {
      this.width = window.innerWidth;
    }, 100),
  },
  computed: {
    ...mapGetters(["currentPromptName"]),
    isMobile() {
      return this.width <= 736;
    },
  },
  mounted() {
    if (!sessionStorage.getItem("application")) {
      let params =
        Object.fromEntries(new URLSearchParams(window.location.search)) || "";

      if (!params?.type) return;
      sessionStorage.setItem("application", params?.type);

      this.application = params?.type;
    } else {
      this.application = sessionStorage.getItem("application");
    }
    // Add the needed event listeners to the window and document.
    window.addEventListener("resize", this.windowsResize);
  },
  beforeDestroy() {
    // Remove event listeners before destroying this page.
    window.removeEventListener("resize", this.windowsResize);
  },
};
</script>

<style></style>
