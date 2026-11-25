<script setup lang="ts">
import { ManagePage, AdminTrendChart } from "@yueli/ui/admin";

definePageMeta({ layout: "manage", middleware: ["auth", "manage"] });
useSeoMeta({ title: "控制台" });

const api = useYtvApi();
const trafficSiteUrl = useRuntimeConfig().public.trafficSiteUrl;
const period = ref(14);
const periodItems = [
  { label: "7 天", value: 7 },
  { label: "14 天", value: 14 },
  { label: "30 天", value: 30 }
];
const { data: searches, error: searchError, status: searchStatus } = await useAsyncData("ytv-dashboard-searches", () => api.adminSearches(period.value, 10), { watch: [period] });
const searchMetrics = computed(() => {
  const summary = searches.value?.summary; const count = summary?.searches || 0;
  const percent = (value: number) => count ? `${Math.round(value / count * 100)}%` : "0%";
  return [
    { label: `近 ${period.value} 天搜索`, value: count, detail: "真实搜索请求", icon: "i-tabler-search" },
    { label: "零结果率", value: percent(summary?.zeroResultSearches || 0), detail: `${summary?.zeroResultSearches || 0} 次无结果`, icon: "i-tabler-search-off" },
    { label: "部分失败率", value: percent(summary?.incompleteSearches || 0), detail: "至少一个来源失败", icon: "i-tabler-plug-connected-x" },
    { label: "平均耗时", value: `${summary?.averageDurationMs || 0} ms`, detail: "全部来源聚合", icon: "i-tabler-clock-bolt" }
  ];
});
const searchTrend = computed(() => (searches.value?.series ?? []).map((point) => ({
  label: point.day,
  axisLabel: shortDay(point.day),
  value: point.searches,
})));
const dateFormatter = new Intl.DateTimeFormat("zh-CN", { month: "2-digit", day: "2-digit", hour: "2-digit", minute: "2-digit" });

function shortDay(day: string) {
  return day.slice(5).replace("-", "/");
}
</script>

<template>
  <ManagePage id="ytv-dashboard" title="控制台" icon="i-tabler-layout-dashboard">
    <template v-if="trafficSiteUrl" #actions>
      <UButton :to="trafficSiteUrl" external label="查看流量统计" icon="i-tabler-chart-line" variant="outline" />
    </template>

    <UAlert v-if="searchError" color="error" variant="subtle" title="搜索统计暂时不可用" description="公开搜索仍可使用，请稍后刷新。" />
    <section v-else aria-labelledby="dashboard-search-title" class="space-y-4">
      <div class="flex flex-wrap items-center justify-between gap-3">
        <h2 id="dashboard-search-title" class="text-lg font-semibold text-highlighted">搜索洞察</h2>
        <UTabs v-model="period" :items="periodItems" :content="false" :disabled="searchStatus === 'pending'" size="sm" class="w-fit" aria-label="搜索统计时间范围" />
      </div>
      <div class="grid grid-cols-4 gap-4 max-[1020px]:grid-cols-2 max-[620px]:grid-cols-1">
        <div v-for="metric in searchMetrics" :key="metric.label" class="yueli-card flex min-w-0 items-center gap-3.5 p-4">
          <span class="grid size-11 shrink-0 place-items-center rounded-xl bg-primary/10 text-primary"><UIcon :name="metric.icon" class="size-5" /></span>
          <div class="grid min-w-0 flex-1"><strong class="text-2xl leading-none tabular-nums text-highlighted">{{ metric.value }}</strong><small class="mt-1 text-muted">{{ metric.label }}</small><span class="mt-1 truncate text-xs text-dimmed">{{ metric.detail }}</span></div>
        </div>
      </div>
      <div class="grid items-start gap-4 xl:grid-cols-[minmax(0,1.25fr)_minmax(22rem,1fr)]">
        <UCard title="搜索趋势" variant="soft" class="yueli-card divide-y-0" :aria-busy="searchStatus === 'pending'" :ui="{ header: 'p-5 sm:p-6', body: 'px-5 pb-5 pt-0 sm:px-6 sm:pb-6 sm:pt-0' }"><AdminTrendChart :points="searchTrend" metric-label="搜索" unit=" 次" /></UCard>
        <UCard class="yueli-card" :ui="{ header: 'p-5', body: 'divide-y divide-default p-0' }">
          <template #header><h3 class="font-semibold text-highlighted">热门搜索词</h3></template>
          <div v-for="item in searches?.topQueries || []" :key="item.query" class="flex items-center justify-between gap-3 px-5 py-3"><div class="min-w-0"><p class="truncate text-sm font-medium text-highlighted">{{ item.query }}</p><p class="mt-0.5 text-xs text-muted">{{ dateFormatter.format(new Date(item.lastSearchedAt)) }}</p></div><div class="shrink-0 text-right"><p class="text-sm tabular-nums text-toned">{{ item.searches }} 次</p><p class="text-xs text-muted">零结果 {{ item.zeroResultSearches }}</p></div></div>
          <UEmpty v-if="!searches?.topQueries.length" icon="i-tabler-search-off" title="还没有搜索记录" />
        </UCard>
      </div>
    </section>
  </ManagePage>
</template>
