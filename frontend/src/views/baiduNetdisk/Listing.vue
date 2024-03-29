<template>
  <div>
    <header-bar showMenu showLogo>
      <title />

      <template #actions>
        <template v-if="!isMobile">
          <!-- 从百度网盘下载到 My Files 中 -->
          <action
            v-if="headerButtons.copy"
            id="copy-button"
            icon="content_copy"
            :label="$t('buttons.copyFile')"
            :counter="selectedCount"
            @action="copyToFiles"
          />
        </template>

        <action
          :icon="viewIcon"
          :label="$t('buttons.switchView')"
          @action="switchView"
        />
        <action icon="info" :label="$t('buttons.info')" show="info" />
        <action
          icon="check_circle"
          :label="$t('buttons.selectMultiple')"
          @action="toggleMultipleSelection"
        />
      </template>
    </header-bar>

    <div v-if="isMobile" id="file-selection">
      <span v-if="selectedCount > 0">{{ selectedCount }} selected</span>
      <action
        v-if="headerButtons.copy"
        icon="content_copy"
        :label="$t('buttons.copyFile')"
        @action="copyToFiles"
      />
    </div>

    <div v-if="loading">
      <h2 class="message delayed">
        <div class="spinner">
          <div class="bounce1"></div>
          <div class="bounce2"></div>
          <div class="bounce3"></div>
        </div>
        <span>{{ $t("files.loading") }}</span>
      </h2>
    </div>
    <template v-else>
      <div v-if="req.numDirs + req.numFiles == 0">
        <h2 class="message">
          <i class="material-icons">sentiment_dissatisfied</i>
          <span>{{ $t("files.lonely") }}</span>
        </h2>
      </div>

      <div
        v-else
        id="listing"
        ref="listing"
        :class="user.viewMode + ' file-icons'"
      >
        <div>
          <div class="item header">
            <div></div>
            <div>
              <p
                :class="{ active: nameSorted }"
                class="name"
                role="button"
                tabindex="0"
                @click="sort('name')"
                :title="$t('files.sortByName')"
                :aria-label="$t('files.sortByName')"
              >
                <span>{{ $t("files.name") }}</span>
                <i class="material-icons">{{ nameIcon }}</i>
              </p>

              <p
                :class="{ active: sizeSorted }"
                class="size"
                role="button"
                tabindex="0"
                @click="sort('size')"
                :title="$t('files.sortBySize')"
                :aria-label="$t('files.sortBySize')"
              >
                <span>{{ $t("files.size") }}</span>
                <i class="material-icons">{{ sizeIcon }}</i>
              </p>
              <p
                :class="{ active: modifiedSorted }"
                class="modified"
                role="button"
                tabindex="0"
                @click="sort('modified')"
                :title="$t('files.sortByLastModified')"
                :aria-label="$t('files.sortByLastModified')"
              >
                <span>{{ $t("files.lastModified") }}</span>
                <i class="material-icons">{{ modifiedIcon }}</i>
              </p>
            </div>
          </div>
        </div>

        <h2 v-if="req.numDirs > 0">{{ $t("files.folders") }}</h2>
        <div v-if="req.numDirs > 0">
          <item
            v-for="item in dirs"
            :key="base64(item.name)"
            v-bind:index="item.index"
            v-bind:name="item.name"
            v-bind:isDir="item.isDir"
            v-bind:url="item.url"
            v-bind:modified="item.modified"
            v-bind:type="item.type"
            v-bind:size="item.size"
            v-bind:path="item.path"
            read-only
          >
          </item>
        </div>

        <h2 v-if="req.numFiles > 0">{{ $t("files.files") }}</h2>
        <div v-if="req.numFiles > 0">
          <item
            v-for="item in files"
            :key="base64(item.name)"
            v-bind:index="item.index"
            v-bind:name="item.name"
            v-bind:isDir="item.isDir"
            v-bind:url="item.url"
            v-bind:modified="item.modified"
            v-bind:type="item.type"
            v-bind:size="item.size"
            v-bind:path="item.path"
            read-only
          >
          </item>
        </div>

        <div :class="{ active: $store.state.multiple }" id="multiple-selection">
          <p>{{ $t("files.multipleSelectionEnabled") }}</p>
          <div
            @click="$store.commit('multiple', false)"
            tabindex="0"
            role="button"
            :title="$t('files.clear')"
            :aria-label="$t('files.clear')"
            class="action"
          >
            <i class="material-icons">clear</i>
          </div>
        </div>
      </div>
    </template>
  </div>
</template>

<script>
import Vue from "vue";
import { mapState, mapGetters, mapMutations } from "vuex";
import { users, files as api } from "@/api";
import css from "@/utils/css";
import throttle from "lodash.throttle";

import HeaderBar from "@/components/header/HeaderBar.vue";
import Action from "@/components/header/Action.vue";
import Item from "@/components/files/ListingItem.vue";

export default {
  name: "listing",
  components: {
    HeaderBar,
    Action,
    Item,
  },
  data: function () {
    return {
      showLimit: 50,
      columnWidth: 280,
      width: window.innerWidth,
      itemWeight: 0,
    };
  },
  computed: {
    ...mapState(["selected", "user", "multiple", "loading"]),
    ...mapState("bd", ["req"]),
    ...mapGetters(["selectedCount", "currentPrompt"]),
    nameSorted() {
      return this.req.sorting?.by === "name";
    },
    sizeSorted() {
      return this.req.sorting?.by === "size";
    },
    modifiedSorted() {
      return this.req.sorting?.by === "modified";
    },
    ascOrdered() {
      return this.req.sorting?.asc;
    },
    items() {
      const dirs = [];
      const files = [];

      this.req.items.forEach((item) => {
        if (item.isDir) {
          dirs.push(item);
        } else {
          files.push(item);
        }
      });

      return { dirs, files };
    },
    dirs() {
      return this.items.dirs.slice(0, this.showLimit);
    },
    files() {
      let showLimit = this.showLimit - this.items.dirs.length;

      if (showLimit < 0) showLimit = 0;

      return this.items.files.slice(0, showLimit);
    },
    nameIcon() {
      if (this.nameSorted && !this.ascOrdered) {
        return "arrow_upward";
      }

      return "arrow_downward";
    },
    sizeIcon() {
      if (this.sizeSorted && this.ascOrdered) {
        return "arrow_downward";
      }

      return "arrow_upward";
    },
    modifiedIcon() {
      if (this.modifiedSorted && this.ascOrdered) {
        return "arrow_downward";
      }

      return "arrow_upward";
    },
    viewIcon() {
      const icons = {
        list: "view_module",
        mosaic: "grid_view",
        "mosaic gallery": "view_list",
      };
      return icons[this.user.viewMode];
    },
    headerButtons() {
      return {
        copy: this.selectedCount > 0 && this.user.perm.create,
      };
    },
    isMobile() {
      return this.width <= 736;
    },
  },
  watch: {
    req: function () {
      // Reset the show value
      this.showLimit = 50;

      // Ensures that the listing is displayed
      Vue.nextTick(() => {
        // How much every listing item affects the window height
        this.setItemWeight();

        // Fill and fit the window with listing items
        this.fillWindow(true);
      });
    },
  },
  mounted: function () {
    // Check the columns size for the first time.
    this.colunmsResize();

    // How much every listing item affects the window height
    this.setItemWeight();

    // Fill and fit the window with listing items
    this.fillWindow(true);

    // Add the needed event listeners to the window and document.
    window.addEventListener("keydown", this.keyEvent);
    window.addEventListener("scroll", this.scrollEvent);
    window.addEventListener("resize", this.windowsResize);
  },
  beforeDestroy() {
    // Remove event listeners before destroying this page.
    window.removeEventListener("keydown", this.keyEvent);
    window.removeEventListener("scroll", this.scrollEvent);
    window.removeEventListener("resize", this.windowsResize);
  },
  methods: {
    ...mapMutations(["updateUser", "addSelected"]),
    base64: function (name) {
      return window.btoa(unescape(encodeURIComponent(name)));
    },
    keyEvent(event) {
      // No prompts are shown
      if (this.currentPrompt !== null) {
        return;
      }

      // Esc!
      if (event.keyCode === 27) {
        // Reset files selection.
        this.$store.commit("resetSelected");
      }

      // Ctrl is pressed
      if (!event.ctrlKey && !event.metaKey) {
        return;
      }

      let key = String.fromCharCode(event.which).toLowerCase();

      switch (key) {
        case "a":
          event.preventDefault();
          for (let file of this.items.files) {
            if (this.$store.state.selected.indexOf(file.index) === -1) {
              this.addSelected(file.index);
            }
          }
          for (let dir of this.items.dirs) {
            if (this.$store.state.selected.indexOf(dir.index) === -1) {
              this.addSelected(dir.index);
            }
          }
          break;
      }
    },
    preventDefault(event) {
      // Wrapper around prevent default.
      event.preventDefault();
    },
    colunmsResize() {
      // Update the columns size based on the window width.
      let items = css(["#listing.mosaic .item", ".mosaic#listing .item"]);
      if (!items) return;

      let columns = Math.floor(
        document.querySelector("main").offsetWidth / this.columnWidth
      );
      if (columns === 0) columns = 1;
      items.style.width = `calc(${100 / columns}% - 1em)`;
    },
    scrollEvent: throttle(function () {
      const totalItems = this.req.numDirs + this.req.numFiles;

      // All items are displayed
      if (this.showLimit >= totalItems) return;

      const currentPos = window.innerHeight + window.scrollY;

      // Trigger at the 75% of the window height
      const triggerPos = document.body.offsetHeight - window.innerHeight * 0.25;

      if (currentPos > triggerPos) {
        // Quantity of items needed to fill 2x of the window height
        const showQuantity = Math.ceil(
          (window.innerHeight * 2) / this.itemWeight
        );

        // Increase the number of displayed items
        this.showLimit += showQuantity;
      }
    }, 100),
    async sort(by) {
      let asc = false;

      if (by === "name") {
        if (this.nameIcon === "arrow_upward") {
          asc = true;
        }
      } else if (by === "size") {
        if (this.sizeIcon === "arrow_upward") {
          asc = true;
        }
      } else if (by === "modified") {
        if (this.modifiedIcon === "arrow_upward") {
          asc = true;
        }
      }

      try {
        await users.update({ id: this.user.id, sorting: { by, asc } }, [
          "sorting",
        ]);
      } catch (e) {
        this.$showError(e?.message || e, false);
      }

      this.$store.commit("setReload", true);
    },
    toggleMultipleSelection() {
      this.$store.commit("multiple", !this.multiple);
      this.$store.commit("closeHovers");
    },
    windowsResize: throttle(function () {
      this.colunmsResize();
      this.width = window.innerWidth;

      // Listing element is not displayed
      if (this.$refs.listing == null) return;

      // How much every listing item affects the window height
      this.setItemWeight();

      // Fill but not fit the window
      this.fillWindow();
    }, 100),
    switchView: async function () {
      this.$store.commit("closeHovers");

      const modes = {
        list: "mosaic",
        mosaic: "mosaic gallery",
        "mosaic gallery": "list",
      };

      const data = {
        id: this.user.id,
        viewMode: modes[this.user.viewMode] || "list",
      };

      users.update(data, ["viewMode"]).catch(this.$showError);

      // Await ensures correct value for setItemWeight()
      await this.$store.commit("updateUser", data);

      this.setItemWeight();
      this.fillWindow();
    },
    setItemWeight() {
      // Listing element is not displayed
      if (this.$refs.listing == null) return;

      let itemQuantity = this.req.numDirs + this.req.numFiles;
      if (itemQuantity > this.showLimit) itemQuantity = this.showLimit;

      // How much every listing item affects the window height
      this.itemWeight = this.$refs.listing.offsetHeight / itemQuantity;
    },
    fillWindow(fit = false) {
      const totalItems = this.req.numDirs + this.req.numFiles;

      // More items are displayed than the total
      if (this.showLimit >= totalItems && !fit) return;

      const windowHeight = window.innerHeight;

      // Quantity of items needed to fill 2x of the window height
      const showQuantity = Math.ceil(
        (windowHeight + windowHeight * 2) / this.itemWeight
      );

      // Less items to display than current
      if (this.showLimit > showQuantity && !fit) return;

      // Set the number of displayed items
      this.showLimit = showQuantity > totalItems ? totalItems : showQuantity;
    },
    async copyToFiles() {
      // init state.req
      try {
        const res = await api.fetch("/");
        this.$store.commit("updateReq", res);
        this.$store.commit("showHover", "copy");
      } catch (e) {
        this.$showError(e?.message || e, false);
      }
    },
  },
};
</script>
