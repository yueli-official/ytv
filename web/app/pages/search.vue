<script setup lang="ts">
definePageMeta({ colorMode: "dark" });
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
const showLoading = useMinimumLoading(pending);
const pageLink = (value: number) => ({ path: "/search", query: { q: query.value, ...(value > 1 ? { page: value } : {}), ...(sourceKeys.value.length ? { sources: sourceKeys.value.join(",") } : {}) } });

watch(query, (value) => {
  if (import.meta.client && value) rememberSearch(value);
}, { immediate: true });

async function retry() {
  await refresh();
}

useSeoMeta({ title: () => query.value ? `搜索：${query.value}` : "搜索" });
</script>

<template>
  <div class="min-h-[calc(100dvh-4rem)] text-white">
    <div data-search-workspace class="ytv-container space-y-7 py-8">
      <div class="max-w-2xl space-y-5">
        <div><p class="ytv-eyebrow">FIND YOUR NEXT FILM</p><h1 class="mt-2 text-2xl font-semibold tracking-tight">搜索影视</h1></div>
        <SearchCommand :initial-value="query" :initial-sources="sourceKeys" />
      </div>

      <div v-if="data && !showLoading" data-search-summary class="ytv-search-summary" role="status" aria-live="polite">
        <div><h2>“{{ query }}”<span>的搜索结果</span></h2><p>{{ data.groups.length }} 个影片 · 第 {{ page }} 页</p></div>
        <p v-if="!data.complete" class="ytv-search-degraded"><UIcon name="i-tabler-info-circle" />部分来源暂时不可用 <button @click="retry">重新搜索</button></p>
      </div>

      <div v-if="showLoading" class="grid gap-4 md:grid-cols-2 xl:grid-cols-3" role="status" aria-label="正在搜索影片">
        <USkeleton v-for="item in 12" :key="item" class="h-40 rounded-xl bg-white/10" />
      </div>
      <UAlert v-else-if="error" color="error" title="搜索暂时不可用" :description="ytvFailureMessage(error, '搜索暂时不可用，请稍后重试。')">
        <template #actions><UButton label="重试" color="error" variant="soft" @click="retry" /></template>
      </UAlert>
      <SearchResultGrid v-else-if="data?.groups.length" :groups="data.groups" :return-to="route.fullPath" />
      <UEmpty v-else-if="query" icon="i-tabler-search-off" :title="page > 1 ? '这一页没有更多结果' : '没有找到可播放结果'" :description="page > 1 ? '返回上一页继续挑选，或换一个关键词。' : '试试更短的片名，或搜索全部来源。'" class="rounded-xl bg-[#0e151d]">
        <template #actions><UButton v-if="sourceKeys.length" label="搜索全部来源" icon="i-tabler-search" :to="{ path: '/search', query: { q: query } }" /><UButton to="/" label="去首页看看" color="neutral" variant="soft" /></template>
      </UEmpty>
      <UEmpty v-else icon="i-tabler-search" title="输入想看的影视" description="可以输入片名，也可以点选上面的搜索历史。" class="rounded-xl bg-[#0e151d]" />

      <nav v-if="query && (data?.groups.length || page > 1)" class="flex items-center justify-center gap-3" aria-label="搜索结果分页">
        <UButton label="上一页" icon="i-tabler-chevron-left" color="neutral" variant="soft" :disabled="page <= 1 || showLoading" :to="page > 1 ? pageLink(page - 1) : undefined" />
        <span class="text-xs text-white/60">第 {{ page }} 页</span>
        <UButton label="下一页" icon="i-tabler-chevron-right" trailing color="neutral" variant="soft" :disabled="!data?.groups.length || showLoading" :to="data?.groups.length ? pageLink(page + 1) : undefined" />
      </nav>
    </div>
  </div>
</template>
