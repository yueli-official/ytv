<script setup lang="ts">
import { ManagePage } from "@yueli/ui/admin";
import AdminTrendChart from "~/components/AdminTrendChart.vue";

definePageMeta({ layout: "manage", middleware: ["auth", "manage"] });
useSeoMeta({ title: "搜索洞察" });

const api = useYtvApi();
const period = ref(14);
const periodItems = [{ label: "7 天", value: 7 }, { label: "14 天", value: 14 }, { label: "30 天", value: 30 }];
const { data, error, refresh, status } = await useAsyncData("ytv-admin-searches", () => api.adminSearches(period.value, 12), { watch: [period] });

const metrics = computed(() => {
  const summary = data.value?.summary;
  const searches = summary?.searches || 0;
  const percent = (value: number) => searches > 0 ? `${Math.round((value / searches) * 100)}%` : "0%";
  return [
    { label: `近 ${period.value} 天搜索`, value: searches, detail: "真实搜索请求", icon: "i-tabler-search" },
    { label: "零结果率", value: percent(summary?.zeroResultSearches || 0), detail: `${summary?.zeroResultSearches || 0} 次无结果`, icon: "i-tabler-search-off" },
    { label: "部分失败率", value: percent(summary?.incompleteSearches || 0), detail: "至少一个来源失败", icon: "i-tabler-plug-connected-x" },
    { label: "平均耗时", value: `${summary?.averageDurationMs || 0} ms`, detail: "全部来源聚合", icon: "i-tabler-clock-bolt" }
  ];
});
const searchTrend = computed(() => (data.value?.series || []).map((point) => ({ day: point.day, value: point.searches })));
const zeroResultTrend = computed(() => (data.value?.series || []).map((point) => ({ day: point.day, value: point.zeroResultSearches })));
const zeroResultPeak = computed(() => Math.max(0, ...(data.value?.series || []).map((point) => point.zeroResultSearches)));
const zeroResultAverage = computed(() => (data.value?.summary.zeroResultSearches || 0) / period.value);

const dateFormatter = new Intl.DateTimeFormat("zh-CN", { month: "2-digit", day: "2-digit", hour: "2-digit", minute: "2-digit" });
</script>

<template>
  <ManagePage id="ytv-search-insights" title="搜索洞察" icon="i-tabler-search">
    <UAlert v-if="error" color="error" variant="subtle" title="搜索统计暂时不可用" description="公开搜索仍可使用；请稍后刷新或检查 Search Analytics 数据库。">
      <template #actions><UButton label="重试" color="error" variant="soft" @click="() => refresh()" /></template>
    </UAlert>

    <div class="grid grid-cols-4 gap-4 max-[1020px]:grid-cols-2 max-[620px]:grid-cols-1">
      <div v-for="metric in metrics" :key="metric.label" class="yueli-card flex min-w-0 items-center gap-3.5 p-4">
        <span class="grid size-11 shrink-0 place-items-center rounded-xl bg-primary/10 text-primary"><UIcon :name="metric.icon" class="size-5" /></span>
        <div class="grid min-w-0 flex-1"><strong class="text-2xl leading-none tabular-nums text-highlighted">{{ metric.value }}</strong><small class="mt-1 text-muted">{{ metric.label }}</small><span class="mt-1 truncate text-xs text-dimmed">{{ metric.detail }}</span></div>
      </div>
    </div>

    <div class="grid items-start gap-4 xl:grid-cols-[minmax(0,2fr)_minmax(18rem,1fr)]">
      <UCard variant="soft" class="yueli-card divide-y-0" :aria-busy="status === 'pending'" :ui="{ header: 'p-5 sm:p-6', body: 'px-5 pb-5 pt-0 sm:px-6 sm:pb-6 sm:pt-0' }">
        <template #header><div class="flex flex-col gap-4 sm:flex-row sm:items-start sm:justify-between"><h2 class="text-base font-semibold text-highlighted">搜索趋势</h2><UTabs v-model="period" :items="periodItems" :content="false" :disabled="status === 'pending'" size="sm" class="w-fit" aria-label="搜索统计时间范围" /></div></template>
        <AdminTrendChart :points="searchTrend" label="搜索" unit="次" />
      </UCard>
      <UCard title="零结果趋势" variant="soft" class="yueli-card divide-y-0" :ui="{ header: 'p-5 sm:p-6', body: 'space-y-5 px-5 pb-5 pt-0 sm:px-6 sm:pb-6 sm:pt-0' }">
        <div><p class="text-3xl font-semibold tabular-nums tracking-tight text-highlighted">{{ data?.summary.zeroResultSearches || 0 }}</p><p class="mt-1 text-xs text-muted">近 {{ period }} 天零结果搜索</p></div>
        <AdminTrendChart :points="zeroResultTrend" label="零结果搜索" unit="次" compact />
        <dl class="grid grid-cols-2 gap-4"><div><dt class="text-xs text-muted">日均零结果</dt><dd class="mt-1 text-lg font-semibold tabular-nums text-highlighted">{{ zeroResultAverage.toFixed(1) }}</dd></div><div><dt class="text-xs text-muted">单日峰值</dt><dd class="mt-1 text-lg font-semibold tabular-nums text-highlighted">{{ zeroResultPeak }}</dd></div></dl>
      </UCard>
    </div>

    <UCard v-if="data?.topQueries.length" class="yueli-card" :ui="{ header: 'flex items-center justify-between gap-3', body: 'divide-y divide-default p-0' }">
      <template #header>
        <div><h2 class="font-semibold text-highlighted">热门搜索词</h2><p class="mt-1 text-xs text-muted">按近 14 天真实搜索次数排序</p></div>
        <UBadge color="neutral" variant="subtle">{{ data.topQueries.length }} 个词</UBadge>
      </template>
      <div v-for="(item, index) in data.topQueries" :key="item.query" class="grid gap-3 px-4 py-4 sm:grid-cols-[2rem_minmax(12rem,1fr)_6rem_7rem_8rem] sm:items-center">
        <span class="text-sm font-semibold tabular-nums text-dimmed">{{ String(index + 1).padStart(2, '0') }}</span>
        <div class="min-w-0"><p class="truncate font-medium text-highlighted">{{ item.query }}</p><p class="mt-0.5 text-xs text-muted">最近 {{ dateFormatter.format(new Date(item.lastSearchedAt)) }}</p></div>
        <p class="text-sm tabular-nums text-toned"><span class="sm:hidden">搜索 </span>{{ item.searches }} 次</p>
        <p class="text-sm tabular-nums text-toned"><span class="sm:hidden">平均结果 </span>{{ item.averageResultCount.toFixed(1) }} 条</p>
        <div class="flex items-center gap-2 sm:justify-end"><UBadge :color="item.zeroResultSearches ? 'warning' : 'success'" variant="subtle">零结果 {{ item.zeroResultSearches }}</UBadge></div>
      </div>
    </UCard>

    <UEmpty v-else-if="!error" icon="i-tabler-chart-bar-off" title="还没有搜索记录" description="用户完成第一次真实搜索后，这里会显示查询次数、零结果率和来源失败摘要。" />
  </ManagePage>
</template>
