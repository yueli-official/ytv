<script setup lang="ts">
import { ytvFailureMessage } from "~/utils/ytv-failure";
import { rememberSearch } from "~/utils/search-preferences";

const route = useRoute();
const api = useYtvApi();
const query = computed(() => String(route.query.q || "").trim());
const page = computed(() => Math.max(1, Number(route.query.page) || 1));
const sourceKeys = computed(() => String(route.query.sources || "").split(",").map((value) => value.trim()).filter(Boolean));
const { data, pending, error, refresh } = await useAsyncData(
  () => `ytv-search-${query.value}-${page.value}-${sourceKeys.value.join("-")}`,
  () => query.value ? api.search(query.value, page.value, sourceKeys.value) : Promise.resolve(null),
  { watch: [query, page, sourceKeys] }
);

watch(query, (value) => {
  if (import.meta.client && value) rememberSearch(value);
}, { immediate: true });

async function retry() {
  await refresh();
}

useSeoMeta({ title: () => query.value ? `搜索：${query.value}` : "搜索" });
</script>

<template>
  <div class="min-h-[calc(100dvh-4rem)] bg-[#070b10] text-white">
    <div data-search-workspace class="mx-auto max-w-[90rem] space-y-7 px-4 py-8 sm:px-6 lg:px-8">
      <div class="space-y-4">
        <h1 class="text-2xl font-bold tracking-[-0.03em] sm:text-3xl">搜索影视</h1>
        <SearchCommand :initial-value="query" :initial-sources="sourceKeys" />
      </div>

      <UAlert v-if="data && !data.complete" color="warning" variant="subtle" title="部分来源暂时不可用" :description="`${data.succeeded} 个来源已返回，${data.failed} 个来源失败；当前结果仍可浏览。`" />

      <div v-if="data" data-search-summary class="flex flex-wrap items-center justify-between gap-3 rounded-xl border border-white/10 bg-[#0e151d] px-4 py-3">
        <div class="flex items-center gap-3">
          <span class="grid size-9 place-items-center rounded-lg bg-sky-400/10 text-sky-300"><UIcon name="i-tabler-clipboard-list" /></span>
          <div><p class="text-xs text-sky-300">搜索结果</p><p class="font-bold">{{ data.groups.length }} 个影片</p></div>
        </div>
        <p class="text-sm text-white/50">“{{ query }}” · {{ data.succeeded }} 个来源 · 第 {{ page }} 页</p>
      </div>

      <div v-if="pending" class="grid gap-4 md:grid-cols-2 xl:grid-cols-4">
        <USkeleton v-for="item in 12" :key="item" class="h-40 rounded-xl bg-white/10" />
      </div>
      <UAlert v-else-if="error" color="error" title="搜索暂时不可用" :description="ytvFailureMessage(error, '搜索暂时不可用，请稍后重试。')">
        <template #actions><UButton label="重试" color="error" variant="soft" @click="retry" /></template>
      </UAlert>
      <SearchResultGrid v-else-if="data?.groups.length" :groups="data.groups" />
      <UEmpty v-else-if="query" icon="i-tabler-search-off" title="没有找到可播放结果" description="换一个关键词，或稍后重试当前来源。" class="rounded-xl bg-[#0e151d]" />
      <UEmpty v-else icon="i-tabler-search" title="输入想看的影视" description="YTV 会查询当前启用来源，并明确显示部分失败。" class="rounded-xl bg-[#0e151d]" />

      <div v-if="data?.groups.length" class="flex justify-center gap-3">
        <UButton label="上一页" icon="i-tabler-chevron-left" color="neutral" variant="soft" :disabled="page <= 1" :to="page > 1 ? { path: '/search', query: { q: query, page: page - 1, ...(sourceKeys.length ? { sources: sourceKeys.join(',') } : {}) } } : undefined" />
        <UButton label="下一页" icon="i-tabler-chevron-right" trailing color="neutral" variant="soft" :to="{ path: '/search', query: { q: query, page: page + 1, ...(sourceKeys.length ? { sources: sourceKeys.join(',') } : {}) } }" />
      </div>
    </div>
  </div>
</template>
