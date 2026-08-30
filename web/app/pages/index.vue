<script setup lang="ts">
const api = useYtvApi();
const sourcePage = ref(1);
const trendingPage = ref(1);
const trendingType = ref("movie");
const trendingTag = ref("热门");
const failedTrendingPosters = ref(new Set<string>());
const tags = ["热门", "美剧", "英剧", "日剧", "韩剧", "港剧", "综艺", "纪录片"];
const viewerSourceKeys = useState<string[]>("ytv-selected-source-keys", () => []);

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
  title: "热门与搜索",
  description: "浏览豆瓣热门，随后从已启用的互联网来源搜索和播放。"
});
</script>

<template>
  <!--
  THESIS: YTV is a dense internet-source film portal, not a cinematic marketing page.
  OWN-WORLD: Near-black ink fields, compact slate panels, portrait posters and one restrained cold-blue focus signal.
  STORY: Search chosen sources or revisit history, browse operator-selected source updates, then use Douban as an attributed idea list.
  FIRST VIEWPORT: A slim dark navigation bar leads into multi-source search and history; source updates own the first poster grid.
  FORM: User-directed restoration of the legacy YTV portal grammar after seed bd4f323d's cinematic Hero was explicitly rejected.
  FINISH: unreviewed and undocumented is unfinished; this correction ends with Playwright screenshots and updated DESIGN.md.
  -->
  <div class="min-h-[calc(100dvh-4rem)] bg-[#070b10] text-white">
    <div class="mx-auto max-w-[90rem] space-y-7 px-4 py-6 sm:px-6 lg:px-8 lg:py-7">
      <section class="space-y-4 rounded-xl border border-sky-400/30 bg-[#0f1720] p-4" aria-labelledby="search-title">
        <div class="flex flex-wrap items-center gap-3">
          <span class="grid size-9 place-items-center rounded-lg bg-sky-400/10 text-sky-300"><UIcon name="i-tabler-movie" class="size-5" /></span>
          <div class="min-w-0 flex-1"><h1 id="search-title" class="text-xl font-bold tracking-[-0.025em] sm:text-2xl">影视搜索</h1><p class="mt-1 text-xs text-white/55">选择来源，同时搜索多个互联网片库</p></div>
        </div>

        <SearchCommand compact />
      </section>

      <HomeContinueWatching />

      <section class="space-y-5" aria-labelledby="source-updates-title">
        <div class="flex flex-wrap items-end justify-between gap-3">
          <div><h2 id="source-updates-title" class="text-xl font-bold sm:text-2xl">来源更新</h2><p class="mt-1 text-xs text-white/45">由管理员选择在首页展示的真实来源条目</p></div>
          <div class="flex flex-wrap items-center justify-end gap-2"><UBadge v-for="source in activeSources" :key="source.key" :label="source.name" color="neutral" variant="subtle" /><UButton label="换一批资源" icon="i-tabler-refresh" color="neutral" variant="soft" :loading="sourcePending" @click="changeSourceBatch" /></div>
        </div>
        <UAlert v-if="sourceError" color="error" variant="subtle" title="来源更新暂时不可用"><template #actions><UButton label="重试" color="error" variant="soft" @click="() => refreshSources()" /></template></UAlert>
        <UAlert v-else-if="sourceData && !sourceData.complete" color="warning" variant="subtle" title="部分来源没有返回" :description="`${sourceData.succeeded} 个来源可用，${sourceData.failed} 个来源失败。`" />
        <div v-if="sourcePending" class="grid grid-cols-2 gap-3 sm:grid-cols-3 md:grid-cols-4 lg:grid-cols-5 xl:grid-cols-6"><USkeleton v-for="item in 12" :key="item" class="aspect-[2/3] rounded-lg bg-white/10" /></div>
        <PosterGrid v-else-if="sourceData?.groups.length" :groups="sourceData.groups.slice(0, 12)" />
        <UEmpty v-else-if="!sourceError" icon="i-tabler-database-off" title="当前批次没有来源内容" description="换一批资源，或直接搜索片名。" class="rounded-xl bg-[#0f1720]" />
      </section>

      <section class="space-y-4 rounded-xl border border-white/10 bg-[#0f1720] p-4 sm:p-5" aria-labelledby="douban-hot-title">
        <div class="flex flex-wrap items-center justify-between gap-3">
          <div><h2 id="douban-hot-title" class="text-xl font-bold sm:text-2xl">豆瓣热门</h2><p class="mt-1 text-xs text-white/55">外部热门标题与评分，点击后使用你选择的来源搜索</p></div>
          <UButton label="换一批热门" icon="i-tabler-refresh" color="neutral" variant="soft" :loading="trendingPending" @click="changeTrendingBatch" />
        </div>
        <div class="flex flex-wrap items-center gap-2">
          <UButton label="电影" color="primary" :variant="trendingType === 'movie' ? 'solid' : 'ghost'" size="sm" @click="selectType('movie')" /><UButton label="电视剧" color="neutral" :variant="trendingType === 'tv' ? 'solid' : 'ghost'" size="sm" @click="selectType('tv')" />
          <span class="hidden h-7 w-px bg-white/10 sm:block" aria-hidden="true" />
          <UButton v-for="item in tags" :key="item" :label="item" :color="trendingTag === item ? 'primary' : 'neutral'" :variant="trendingTag === item ? 'solid' : 'outline'" size="sm" @click="selectTag(item)" />
        </div>
        <div v-if="trendingData?.items.length" data-douban-grid class="flex snap-x gap-3 overflow-x-auto border-t border-white/10 pt-4 pb-2" aria-label="豆瓣热门搜索">
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
        <p v-else-if="trendingError" class="border-t border-white/10 pt-4 text-xs text-white/45">热门列表暂时不可用，来源浏览与搜索不受影响。</p>
      </section>
    </div>
  </div>
</template>
