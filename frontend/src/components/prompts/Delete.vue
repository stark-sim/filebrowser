<template>
  <div class="card floating">
    <div class="card-content">
      <p v-if="selectedCount === 1">
        {{ $t("prompts.deleteMessageSingle") }}
      </p>
      <p v-else>
        {{ $t("prompts.deleteMessageMultiple", { count: selectedCount }) }}
      </p>
    </div>
    <div class="card-action">
      <button
        @click="$store.commit('closeHovers')"
        class="button button--flat button--grey"
        :aria-label="$t('buttons.cancel')"
        :title="$t('buttons.cancel')"
      >
        {{ $t("buttons.cancel") }}
      </button>
      <button
        :disabled="btnLoading"
        @click="submit"
        class="button button--flat button--red"
        :aria-label="$t('buttons.delete')"
        :title="$t('buttons.delete')"
      >
        {{ $t("buttons.delete") }}
      </button>
    </div>
  </div>
</template>

<script>
import { mapGetters, mapMutations, mapState } from "vuex";
import { files as api } from "@/api";
import buttons from "@/utils/buttons";

export default {
  name: "delete",
  data: () => {
    return {
      btnLoading: false,
    };
  },
  computed: {
    ...mapGetters(["isListing", "selectedCount", "currentPrompt"]),
    ...mapState(["req", "selected"]),
  },
  methods: {
    ...mapMutations(["closeHovers"]),
    submit: async function () {
      buttons.loading("delete");

      try {
        this.btnLoading = true;
        if (!this.isListing) {
          await api.remove(this.$route.path);
          buttons.success("delete");

          this.currentPrompt?.confirm();
          this.closeHovers();
          return;
        }

        this.closeHovers();
        this.btnLoading = false;
        if (this.selectedCount === 0) {
          return;
        }

        let promises = [];
        for (let index of this.selected) {
          promises.push(api.remove(this.req.items[index].url));
        }

        await Promise.all(promises);
        buttons.success("delete");
        this.$store.commit("setReload", true);
      } catch (e) {
        buttons.done("delete");
        this.$showError(e);
        if (this.isListing) this.$store.commit("setReload", true);
      }
    },
  },
};
</script>
