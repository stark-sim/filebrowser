<template>
  <div class="row" id="netdisk-login">
    <div class="column">
      <div class="card">
        <div class="card-title">
          <h2>{{ $t("baiduNetdisk.addAuth") }}</h2>
        </div>
        <div class="card-content">
          <input
            class="input input--block"
            v-focus
            type="text"
            @keyup.enter="submit"
            v-model.trim="code"
            :placeholder="$t('baiduNetdisk.inputCode')"
          />
          <p class="link small">
            <a
              target="_blank"
              href="http://openapi.baidu.com/oauth/2.0/authorize?response_type=code&client_id=zNBhtXeLhZDRoxMI6trDohpVREC5AEFP&redirect_uri=oob&scope=basic,netdisk&device_id=39856593"
              class="link"
            >
              {{ $t("baiduNetdisk.getCode") }}
            </a>
            <a target="_blank" href="" class="link">
              {{ $t("baiduNetdisk.documentation") }}
            </a>
          </p>
        </div>
        <div class="card-action">
          <button
            class="button button--flat"
            @click="submit"
            :aria-label="$t('buttons.add')"
            :title="$t('buttons.add')"
          >
            <span>{{ $t("buttons.add") }}</span>
          </button>
        </div>
      </div>
    </div>
  </div>
</template>

<script>
import { bdApi } from "@/api";

export default {
  name: "auth-login",
  data: function () {
    return {
      code: "",
    };
  },
  methods: {
    async submit(event) {
      event.preventDefault();
      event.stopPropagation();

      if (!this.code) return;

      try {
        await bdApi.login(this.code);
        await bdApi.fetchUserInfo();
      } catch (e) {
        if (e.status === 500) {
          this.$showError(
            { message: this.$t("baiduNetdisk.bindFail") },
            false,
            1000
          );
        }
      }
    },
  },
};
</script>
