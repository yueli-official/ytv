<script setup lang="ts">
definePageMeta({ colorMode: "dark" });
import { ytvFailureMessage } from "~/utils/ytv-failure";
import { rememberSearch } from "~/utils/search-preferences";

const route = useRoute();
const api = useYtvApi();
const sourceSettingsOpen = useState("ytv-source-settings-open", () => false);
const query = computed(() => String(route.query.q || "").trim());
const page = computed(() => Math.max(1, Number(route.query.page) || 1));
const sourceKeys = computed(() => String(route.query.sources || "").split(",").map((value) => value.trim()).filter(Boolean));
const { data, pending, error, refresh } = await useAsyncData(
  () => `ytv-search-${query.value}-${page.value}-${sourceKeys.value.join("-")}`,
  () => query.value ? api.search(query.value, page.value, sourceKeys.value) : Promise.resolve(null),
  { watch: [query, page, sourceKeys] }
);
const showLoading = useMinimumLoading(pending);
const verificationSources = computed(() => data.value?.sources.filter(source => source.outcome === "verification_required") || []);
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
    <div data-search-workspace class="ytv-container space-y-4 py-9">
      <div class="ytv-search-heading">
        <h1>{{ query ? `“${query}” 的搜索结果` : '搜索影视' }}</h1>
        <div v-if="query" data-search-summary class="ytv-search-summary min-h-6" role="status" aria-live="polite">
          <p v-if="showLoading">正在搜索 · 第 {{ page }} 页</p>
          <p v-else-if="data">本页 {{ data.groups.length }} 个影片 · 第 {{ page }} 页</p>
          <p v-if="data && !showLoading && !data.complete" class="ytv-search-degraded"><UIcon name="i-tabler-info-circle" />部分来源暂时不可用</p>
        </div>
        <div class="flex flex-wrap items-center gap-3">
        <button class="ytv-source-scope" @click="sourceSettingsOpen = true"><UIcon name="i-tabler-database" />{{ sourceKeys.length ? `已选 ${sourceKeys.length} 个来源` : '全部已启用来源' }}<span class="mx-2 text-white/30">|</span><UIcon name="i-tabler-adjustments-horizontal" />来源设置</button>
        <UButton v-if="query" label="重新搜索" icon="i-tabler-refresh" color="neutral" variant="ghost" size="sm" class="mt-4" :loading="pending" :disabled="pending" @click="retry" />
        </div>
        <SearchCommand v-if="!query" history-only :initial-sources="sourceKeys" />
        <details v-if="data?.sources.length && !showLoading" :open="verificationSources.length > 0 && !data.groups.length" class="mt-3 text-sm text-white/70">
          <summary class="w-fit cursor-pointer rounded py-1">查看来源状态 · {{ data.succeeded }}/{{ data.attempted }} 正常响应</summary>
          <ul class="mt-2 grid gap-2 md:grid-cols-2 xl:grid-cols-3" aria-label="本次搜索来源状态">
            <li v-for="source in data.sources" :key="source.key" class="flex items-center justify-between gap-3 rounded-lg border border-white/10 px-3 py-2">
              <span class="min-w-0 break-words">{{ source.name }}</span>
              <span class="shrink-0 text-xs" :class="source.outcome === 'success' ? 'text-white/65' : 'text-amber-300'">{{ source.outcome === 'success' ? source.itemCount ? `${source.itemCount} 条结果` : '没有匹配结果' : source.outcome === 'verification_required' ? source.verificationKind === 'image' ? '需要验证码' : '需要浏览器验证' : source.outcome === 'timeout' ? '请求超时' : '暂时不可用' }}</span>
            </li>
          </ul>
        </details>
      </div>

      <div v-if="showLoading" class="grid gap-5" role="status" aria-label="正在搜索影片">
        <div v-for="item in 3" :key="item" class="ytv-result-card" aria-hidden="true">
          <USkeleton class="ytv-result-poster bg-white/10" />
          <div class="space-y-3"><USkeleton class="h-6 w-2/5 bg-white/10" /><USkeleton class="h-4 w-1/4 bg-white/10" /></div>
          <div class="ytv-result-sources grid gap-2 xl:grid-cols-3"><USkeleton v-for="source in 3" :key="source" class="h-16 rounded-lg bg-white/10" /></div>
        </div>
      </div>
      <UAlert v-else-if="error" color="error" title="搜索暂时不可用" :description="ytvFailureMessage(error, '搜索暂时不可用，请稍后重试。')">
        <template #actions><UButton label="重试" color="error" variant="soft" @click="retry" /></template>
      </UAlert>
      <SearchResultGrid v-else-if="data?.groups.length" :groups="data.groups" :return-to="route.fullPath" />
      <UEmpty v-else-if="verificationSources.length" icon="i-tabler-shield-lock" title="部分来源需要验证" description="这些来源尚未返回搜索结果。本站暂不能完成其验证，请选择其他来源继续搜索。" class="rounded-xl bg-[#141e28]">
        <template #actions><UButton v-if="sourceKeys.length" label="搜索全部来源" icon="i-tabler-search" :to="{ path: '/search', query: { q: query } }" /><UButton label="来源设置" color="neutral" variant="soft" @click="sourceSettingsOpen = true" /></template>
      </UEmpty>
      <UEmpty v-else-if="query" icon="i-tabler-search-off" :title="page > 1 ? '这一页没有更多结果' : '没有找到相关影片'" :description="page > 1 ? '返回上一页继续挑选，或换一个关键词。' : '试试更短的片名，或搜索全部来源。'" class="rounded-xl bg-[#141e28]">
        <template #actions><UButton v-if="sourceKeys.length" label="搜索全部来源" icon="i-tabler-search" :to="{ path: '/search', query: { q: query } }" /><UButton to="/" label="去首页看看" color="neutral" variant="soft" /></template>
      </UEmpty>
      <UEmpty v-else icon="i-tabler-search" title="输入想看的影视" description="可以输入片名，也可以点选上面的搜索历史。" class="rounded-xl bg-[#141e28]" />

      <nav v-if="query && (data?.groups.length || page > 1)" class="flex items-center justify-center gap-3" aria-label="搜索结果分页">
        <UButton label="上一页" icon="i-tabler-chevron-left" color="neutral" variant="soft" :disabled="page <= 1 || showLoading" :to="page > 1 ? pageLink(page - 1) : undefined" />
        <span class="text-xs text-white/60">第 {{ page }} 页</span>
        <UButton label="下一页" icon="i-tabler-chevron-right" trailing color="neutral" variant="soft" :disabled="!data?.groups.length || showLoading" :to="data?.groups.length ? pageLink(page + 1) : undefined" />
      </nav>
    </div>
  </div>
</template>
