<script setup lang="ts">
import { ManagePage } from "@yueli/ui/admin";
import AdminTrendChart from "~/components/AdminTrendChart.vue";

definePageMeta({ layout: "manage", middleware: ["auth", "manage"] });
useSeoMeta({ title: "控制台" });

const api = useYtvApi();
const period = ref(14);
const periodItems = [
  { label: "7 天", value: 7 },
  { label: "14 天", value: 14 },
  { label: "30 天", value: 30 }
];
const { data: traffic, error, status } = await useAsyncData("ytv-admin-traffic", () => api.adminTraffic(period.value), { watch: [period] });
const formatter = new Intl.NumberFormat("zh-CN");
const showSkeleton = useMinimumLoading(computed(() => status.value === "pending" && !traffic.value?.pages.series.length));
const pageTrend = computed(() => (traffic.value?.pages.series || []).map((point) => ({ day: point.day, value: point.views })));
const visitorTrend = computed(() => (traffic.value?.pages.series || []).map((point) => ({ day: point.day, value: point.uniqueVisitorDays })));
const visitorAverage = computed(() => (traffic.value?.pages.recentUniqueVisitorDays || 0) / period.value);
const visitorPeak = computed(() => Math.max(0, ...(traffic.value?.pages.series || []).map((point) => point.uniqueVisitorDays)));
const metrics = computed(() => [
  { label: "累计浏览", value: traffic.value?.pages.allTimeViews || 0, detail: `${formatter.format(traffic.value?.pages.allTimeUniqueVisitorDays || 0)} 访客日`, icon: "i-tabler-eye" },
  { label: `近 ${period.value} 天浏览`, value: traffic.value?.pages.recentViews || 0, detail: `${formatter.format(traffic.value?.pages.recentUniqueVisitorDays || 0)} 访客日`, icon: "i-tabler-chart-line" },
  { label: `近 ${period.value} 天访客`, value: traffic.value?.pages.recentUniqueVisitorDays || 0, detail: `日均 ${visitorAverage.value.toFixed(1)} 访客日`, icon: "i-tabler-users" },
  { label: "累计播放启动", value: traffic.value?.playback.allTimeViews || 0, detail: `近 ${period.value} 天 ${formatter.format(traffic.value?.playback.recentViews || 0)} 次`, icon: "i-tabler-player-play" }
]);
</script>

<template>
  <ManagePage id="ytv-dashboard" title="控制台" icon="i-tabler-layout-dashboard">
    <UAlert v-if="error" color="error" variant="subtle" title="流量统计暂时不可用" description="数据源与设置仍可使用；请稍后刷新或检查 Traffic 数据库。" />
    <div v-if="showSkeleton" class="grid grid-cols-4 gap-4 max-[1020px]:grid-cols-2 max-[620px]:grid-cols-1">
      <USkeleton v-for="item in 4" :key="item" class="h-24 rounded-xl" />
    </div>
    <div v-else class="grid grid-cols-4 gap-4 max-[1020px]:grid-cols-2 max-[620px]:grid-cols-1">
      <div v-for="metric in metrics" :key="metric.label" class="yueli-card flex min-w-0 items-center gap-3.5 p-4">
        <span class="grid size-11 shrink-0 place-items-center rounded-xl bg-primary/10 text-primary"><UIcon :name="metric.icon" class="size-5" /></span>
        <div class="grid min-w-0 flex-1"><strong class="text-2xl leading-none tabular-nums text-highlighted">{{ formatter.format(metric.value) }}</strong><small class="mt-1 text-muted">{{ metric.label }}</small><span class="mt-1 truncate text-xs text-dimmed">{{ metric.detail }}</span></div>
      </div>
    </div>

    <div class="grid items-start gap-4 xl:grid-cols-[minmax(0,2fr)_minmax(18rem,1fr)]">
      <UCard variant="soft" class="yueli-card divide-y-0" data-dashboard-trend :aria-busy="status === 'pending'" :ui="{ header: 'p-5 sm:p-6', body: 'px-5 pb-5 pt-0 sm:px-6 sm:pb-6 sm:pt-0' }">
        <template #header><div class="flex flex-col gap-4 sm:flex-row sm:items-start sm:justify-between"><h2 class="text-base font-semibold text-highlighted">访问趋势</h2><UTabs v-model="period" :items="periodItems" :content="false" :disabled="status === 'pending'" size="sm" class="w-fit" aria-label="统计时间范围" /></div></template>
        <AdminTrendChart :points="pageTrend" label="访问" unit="次" />
      </UCard>
      <UCard title="访客趋势" variant="soft" class="yueli-card divide-y-0" data-dashboard-visitors :ui="{ header: 'p-5 sm:p-6', body: 'space-y-5 px-5 pb-5 pt-0 sm:px-6 sm:pb-6 sm:pt-0' }">
        <div class="flex items-end justify-between gap-4"><div><p class="text-3xl font-semibold tabular-nums tracking-tight text-highlighted">{{ formatter.format(traffic?.pages.recentUniqueVisitorDays || 0) }}</p><p class="mt-1 text-xs text-muted">近 {{ period }} 天访客日</p></div><UBadge color="neutral" variant="soft" :label="`播放访客日 ${formatter.format(traffic?.playback.recentUniqueVisitorDays || 0)}`" /></div>
        <AdminTrendChart :points="visitorTrend" label="访客" unit="访客日" compact />
        <dl class="grid grid-cols-2 gap-4"><div><dt class="text-xs text-muted">日均访客日</dt><dd class="mt-1 text-lg font-semibold tabular-nums text-highlighted">{{ visitorAverage.toFixed(1) }}</dd></div><div><dt class="text-xs text-muted">单日峰值</dt><dd class="mt-1 text-lg font-semibold tabular-nums text-highlighted">{{ formatter.format(visitorPeak) }}</dd></div></dl>
      </UCard>
    </div>
  </ManagePage>
</template>
