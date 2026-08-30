<script setup lang="ts">
import { zh_cn } from "@nuxt/ui/locale";
import { FeedbackToastRegion } from "@yueli/ui/feedback/pattern";

useHead({ htmlAttrs: { lang: "zh-CN" } });
</script>

<template>
  <UApp :locale="zh_cn" :toaster="null">
    <NuxtLayout>
      <NuxtPage />
    </NuxtLayout>
    <FeedbackToastRegion />
  </UApp>
</template>
