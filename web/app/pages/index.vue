<script setup lang="ts">
definePageMeta({ colorMode: "dark" });
const api = useYtvApi();
const sourcePage = ref(1);
const trendingPage = ref(1);
const trendingType = ref("movie");
const trendingTag = ref("热门");
const failedTrendingPosters = ref(new Set<string>());
const tags = ["热门", "美剧", "英剧", "日剧", "韩剧", "港剧", "综艺", "纪录片"];
const viewerSourceKeys = useState<string[]>("ytv-selected-source-keys", () => []);
const { rail: trendingRail, canPrevious: canPreviousTrending, canNext: canNextTrending, previous: previousTrending, next: nextTrending } = useHorizontalRail();

const { data: sourceData, pending: sourcePending, error: sourceError, refresh: refreshSources } = await useAsyncData(
  () => `ytv-discover-${sourcePage.value}`,
  () => api.discover(sourcePage.value),
  { watch: [sourcePage] }
);
const { data: trendingData, pending: trendingPending, error: trendingError } = await useAsyncData(
  () => `ytv-trending-${trendingType.value}-${trendingTag.value}-${trendingPage.value}`,
  () => api.trending(trendingType.value, trendingTag.value, trendingPage.value, 12),
  { watch: [trendingType, trendingTag, trendingPage] }
);

const activeSources = computed(() => (sourceData.value?.sources || []).filter((item) => item.outcome === "success"));

function selectType(value: string) {
  trendingType.value = value;
  trendingTag.value = "热门";
  trendingPage.value = 1;
}

function selectTag(value: string) {
  trendingTag.value = value;
  trendingPage.value = 1;
}

function changeSourceBatch() {
  sourcePage.value += 1;
}

function changeTrendingBatch() {
  trendingPage.value += 1;
}

async function searchTrending(title: string) {
  await navigateTo({ path: "/search", query: { q: title, ...(viewerSourceKeys.value.length ? { sources: viewerSourceKeys.value.join(",") } : {}) } });
}

function trendingPoster(item: { id: string; coverUrl: string }) {
  return item.coverUrl && !failedTrendingPosters.value.has(item.id) ? item.coverUrl : "";
}

function failTrendingPoster(id: string) {
  failedTrendingPosters.value = new Set([...failedTrendingPosters.value, id]);
}

useSeoMeta({
  title: "发现好电影",
  description: "搜索想看的电影，浏览热门影片，接着上次的进度继续观看。"
});
</script>

<template>
  <div class="ytv-discover min-h-[calc(100dvh-4rem)] text-white">
    <div class="mx-auto max-w-[90rem] space-y-8 px-4 py-6 sm:px-6 lg:px-8 lg:py-8">
      <section class="grid gap-5 border-b border-white/10 pb-6 lg:grid-cols-[minmax(13rem,0.6fr)_minmax(0,1fr)] lg:items-center" aria-labelledby="search-title">
        <div class="space-y-2">
          <p class="text-[10px] font-semibold tracking-[0.24em] text-sky-300/80">YTV / CINEMA</p>
          <h1 id="search-title" class="text-2xl font-semibold tracking-tight sm:text-3xl">今晚，想看什么？</h1>
          <p class="text-xs text-white/45">找到好电影，接着上次的故事。</p>
        </div>

        <SearchCommand compact />
      </section>

      <HomeContinueWatching />

      <section class="space-y-4" aria-labelledby="douban-hot-title">
        <div class="flex flex-wrap items-center justify-between gap-3">
          <div class="flex items-center gap-3"><h2 id="douban-hot-title" class="text-lg font-semibold">大家在看</h2><span class="rounded border border-white/10 px-1.5 py-0.5 text-[10px] text-white/40">豆瓣热门</span></div>
          <UButton label="换一批热门" icon="i-tabler-refresh" color="neutral" variant="ghost" size="sm" :loading="trendingPending" @click="changeTrendingBatch" />
        </div>
        <div class="flex flex-wrap items-center gap-2">
          <UButton label="电影" color="primary" :variant="trendingType === 'movie' ? 'solid' : 'ghost'" size="sm" @click="selectType('movie')" /><UButton label="电视剧" color="neutral" :variant="trendingType === 'tv' ? 'solid' : 'ghost'" size="sm" @click="selectType('tv')" />
          <span class="hidden h-7 w-px bg-white/10 sm:block" aria-hidden="true" />
          <template v-if="trendingType === 'tv'"><UButton v-for="item in tags" :key="item" :label="item" :color="trendingTag === item ? 'primary' : 'neutral'" :variant="trendingTag === item ? 'soft' : 'ghost'" size="sm" @click="selectTag(item)" /></template>
        </div>
        <div v-if="trendingData?.items.length" class="relative border-t border-white/10 pt-4">
          <UButton v-if="canPreviousTrending" data-carousel-previous icon="i-tabler-chevron-left" aria-label="上一组豆瓣热门" color="neutral" variant="ghost" class="absolute -left-3 top-1/2 z-10 hidden min-h-11 min-w-11 -translate-y-1/2 justify-center rounded-full border border-white/15 bg-[#070b10]/95 text-white shadow-lg hover:bg-[#111a24] sm:inline-flex" @click="previousTrending" />
          <div ref="trendingRail" data-douban-grid class="douban-rail flex snap-x snap-mandatory gap-3 overflow-x-auto overscroll-x-contain pb-1" aria-label="豆瓣热门搜索">
            <NuxtLink v-for="item in trendingData.items" :key="item.id" :to="{ path: '/search', query: { q: item.title } }" class="group w-[8.75rem] shrink-0 snap-start text-white no-underline sm:w-40" @click.prevent="searchTrending(item.title)">
              <div class="relative aspect-[2/3] overflow-hidden rounded-xl border border-white/10 bg-[#111923] transition group-hover:-translate-y-0.5 group-hover:border-sky-400/45">
                <img v-if="trendingPoster(item)" :src="trendingPoster(item)" :alt="`${item.title} 海报`" loading="lazy" decoding="async" class="size-full object-cover" @error="failTrendingPoster(item.id)">
                <span v-else class="grid size-full place-items-center text-white/25"><UIcon name="i-tabler-photo" class="size-9" aria-hidden="true" /></span>
                <span v-if="item.score" class="absolute right-2 top-2 rounded-md bg-black/75 px-2 py-1 text-[11px] font-bold text-amber-300">★ {{ item.score }}</span>
                <span v-if="item.isNew" class="absolute left-2 top-2 rounded-md bg-sky-300 px-2 py-1 text-[10px] font-semibold text-sky-950">新片</span>
              </div>
              <h3 class="mt-2 truncate text-sm font-semibold group-hover:text-sky-300">{{ item.title }}</h3>
            </NuxtLink>
          </div>
          <UButton v-if="canNextTrending" data-carousel-next icon="i-tabler-chevron-right" aria-label="下一组豆瓣热门" color="neutral" variant="ghost" class="absolute -right-3 top-1/2 z-10 hidden min-h-11 min-w-11 -translate-y-1/2 justify-center rounded-full border border-white/15 bg-[#070b10]/95 text-white shadow-lg hover:bg-[#111a24] sm:inline-flex" @click="nextTrending" />
        </div>
        <p v-else-if="trendingError" class="border-t border-white/10 pt-4 text-xs text-white/45">热门列表暂时不可用，来源浏览与搜索不受影响。</p>
      </section>

      <section class="space-y-5" aria-labelledby="source-updates-title">
        <div class="flex flex-wrap items-end justify-between gap-3">
          <div><h2 id="source-updates-title" class="text-lg font-semibold">片库新上架</h2><p class="mt-1 text-xs text-white/40">来自当前片库的最新更新</p></div>
          <div class="flex flex-wrap items-center justify-end gap-2"><UBadge v-for="source in activeSources" :key="source.key" :label="source.name" color="neutral" variant="subtle" /><UButton label="换一批资源" icon="i-tabler-refresh" color="neutral" variant="soft" :loading="sourcePending" @click="changeSourceBatch" /></div>
        </div>
        <UAlert v-if="sourceError" color="error" variant="subtle" title="来源更新暂时不可用"><template #actions><UButton label="重试" color="error" variant="soft" @click="() => refreshSources()" /></template></UAlert>
        <UAlert v-else-if="sourceData && !sourceData.complete" color="warning" variant="subtle" title="部分来源没有返回" :description="`${sourceData.succeeded} 个来源可用，${sourceData.failed} 个来源失败。`" />
        <div v-if="sourcePending" class="grid grid-cols-2 gap-3 sm:grid-cols-3 md:grid-cols-4 lg:grid-cols-5 xl:grid-cols-6"><USkeleton v-for="item in 12" :key="item" class="aspect-[2/3] rounded-lg bg-white/10" /></div>
        <PosterGrid v-else-if="sourceData?.groups.length" :groups="sourceData.groups.slice(0, 12)" />
        <UEmpty v-else-if="!sourceError" icon="i-tabler-database-off" title="当前批次没有来源内容" description="换一批资源，或直接搜索片名。" class="rounded-xl bg-[#0f1720]" />
      </section>
    </div>
  </div>
</template>

<style scoped>
.ytv-discover { background: radial-gradient(ellipse at 85% 0%, #142337 0%, #070b10 45%); }
.douban-rail {
  scrollbar-width: none;
}

.douban-rail::-webkit-scrollbar {
  display: none;
}
</style>
