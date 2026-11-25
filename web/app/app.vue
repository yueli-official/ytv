<script setup lang="ts">
import { zh_cn } from "@nuxt/ui/locale";
import { FeedbackToastRegion } from "@yueli/ui/feedback/pattern";

const api = useYtvApi();
const { data: branding } = await useAsyncData("ytv-branding", () => api.siteBranding());
useHead(() => ({
  htmlAttrs: { lang: "zh-CN" },
  link: [{ rel: "icon", href: branding.value?.faviconUrl || "/favicon.svg" }],
}));
</script>

<template>
  <UApp :locale="zh_cn" :toaster="null">
    <NuxtLayout>
      <NuxtPage />
    </NuxtLayout>
    <FeedbackToastRegion />
  </UApp>
</template>
