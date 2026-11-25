<script setup lang="ts">
import { ManagePage } from "@yueli/ui/admin";

definePageMeta({ layout: "manage", middleware: ["auth", "manage"] });
useSeoMeta({ title: "控制台" });
const trafficSiteUrl = useRuntimeConfig().public.trafficSiteUrl;
const searchReportUrl = computed(() => {
  if (!trafficSiteUrl) return "";
  const url = new URL(trafficSiteUrl);
  url.searchParams.set("tab", "search");
  return url.href;
});
</script>

<template>
  <ManagePage id="ytv-dashboard" title="控制台" icon="i-tabler-layout-dashboard">
    <template v-if="searchReportUrl" #actions>
      <UButton :to="searchReportUrl" external label="查看搜索分析" icon="i-tabler-chart-line" variant="outline" />
    </template>
    <div class="flex flex-wrap gap-3">
      <UButton to="/manage/sources" label="管理数据源" icon="i-tabler-plug-connected" variant="soft" />
      <UButton to="/manage/settings" label="站点设置" icon="i-tabler-settings" variant="soft" />
    </div>
  </ManagePage>
</template>
