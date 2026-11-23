<script setup lang="ts">
import { ManagePage, AdminOverview, AdminMetricCard, AdminTrendChart } from "@yueli/ui/admin";

definePageMeta({ layout: "manage", middleware: ["auth", "manage"] });
useSeoMeta({ title: "控制台" });

const api = useYtvApi();
const period = ref(14);
const periodItems = [
  { label: "7 天", value: 7 },
  { label: "14 天", value: 14 },
  { label: "30 天", value: 30 }
];
const trafficMetric = ref<"visitors" | "views">("visitors");
const trafficMetricItems = [
  { label: "独立访客", value: "visitors" },
  { label: "浏览量", value: "views" }
];
const { data: traffic, error, status } = await useAsyncData("ytv-admin-traffic", () => api.adminTraffic(period.value), { watch: [period] });
const { data: searches, error: searchError, status: searchStatus } = await useAsyncData("ytv-dashboard-searches", () => api.adminSearches(period.value, 10), { watch: [period] });
const formatter = new Intl.NumberFormat("zh-CN");
const showSkeleton = useMinimumLoading(computed(() => status.value === "pending" && !traffic.value?.pages.series.length));
const showVisitors = computed(() => trafficMetric.value === "visitors");
const trafficTrend = computed(() => (traffic.value?.pages.series ?? []).map((point) => ({
  label: shortDay(point.day),
  value: showVisitors.value ? point.uniqueVisitorDays : point.views
})));
const trafficDays = computed(() => traffic.value?.days ?? period.value);
const visitorAverage = computed(() => (traffic.value?.pages.recentUniqueVisitorDays || 0) / trafficDays.value);
const metrics = computed(() => [
  { label: `近 ${trafficDays.value} 天访客日`, value: traffic.value?.pages.recentUniqueVisitorDays || 0, detail: `日均 ${visitorAverage.value.toFixed(1)}，按日去重后相加`, icon: "i-tabler-users" },
  { label: `近 ${trafficDays.value} 天浏览`, value: traffic.value?.pages.recentViews || 0, icon: "i-tabler-chart-line" },
  { label: "累计浏览", value: traffic.value?.pages.allTimeViews || 0, detail: `${formatter.format(traffic.value?.pages.allTimeUniqueVisitorDays || 0)} 访客日`, icon: "i-tabler-eye" },
  { label: "累计播放启动", value: traffic.value?.playback.allTimeViews || 0, detail: `近 ${trafficDays.value} 天 ${formatter.format(traffic.value?.playback.recentViews || 0)} 次`, icon: "i-tabler-player-play" }
]);
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
const searchTrend = computed(() => (searches.value?.series ?? []).map((point) => ({ label: shortDay(point.day), value: point.searches })));
const dateFormatter = new Intl.DateTimeFormat("zh-CN", { month: "2-digit", day: "2-digit", hour: "2-digit", minute: "2-digit" });

function shortDay(day: string) {
  return day.slice(5).replace("-", "/");
}
</script>

<template>
  <ManagePage id="ytv-dashboard" title="控制台" icon="i-tabler-layout-dashboard">
    <template #tools>
      <AdminOverview>
        <template #artwork><ManageOverviewArtwork /></template>
        <div v-if="showSkeleton" data-admin-metrics><USkeleton v-for="n in 4" :key="n" class="h-24 rounded-xl" /></div>
        <div v-else data-admin-metrics><AdminMetricCard v-for="metric in metrics" :key="metric.label" :label="metric.label" :value="formatter.format(metric.value)" :icon="metric.icon" :detail="metric.detail" /></div>
      </AdminOverview>
    </template>
    <UAlert v-if="error" color="error" variant="subtle" title="流量统计暂时不可用" description="数据源与设置仍可使用；请稍后刷新或检查 Traffic 数据库。" />
    <UCard variant="soft" class="yueli-card divide-y-0" data-dashboard-trend :aria-busy="status === 'pending'" :ui="{ header: 'p-5 sm:p-6', body: 'px-5 pb-5 pt-0 sm:px-6 sm:pb-6 sm:pt-0' }">
      <template #header>
        <div class="flex flex-wrap items-center justify-between gap-3">
          <h2 class="text-base font-semibold text-highlighted">站点流量</h2>
          <UTabs v-model="period" :items="periodItems" :content="false" :disabled="status === 'pending'" size="sm" class="w-fit" aria-label="统计时间范围" />
        </div>
      </template>
      <UTabs v-model="trafficMetric" :items="trafficMetricItems" :content="false" size="sm" class="mb-4 w-fit max-w-full" aria-label="流量指标" />
      <USkeleton v-if="showSkeleton" class="h-60 rounded-lg" />
      <AdminTrendChart
        v-else-if="!error"
        :points="trafficTrend"
        :metric-label="showVisitors ? '独立访客' : '浏览量'"
        :unit="showVisitors ? ' 人' : ' 次'"
      />
    </UCard>

    <UAlert v-if="searchError" color="error" variant="subtle" title="搜索统计暂时不可用" description="公开搜索仍可使用，请稍后刷新。" />
    <section v-else aria-labelledby="dashboard-search-title" class="space-y-4">
      <div class="flex items-center justify-between gap-3"><h2 id="dashboard-search-title" class="text-lg font-semibold text-highlighted">搜索洞察</h2><span class="text-xs text-muted">近 {{ period }} 天</span></div>
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
