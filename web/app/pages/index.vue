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

const featuredIndex = ref(0);
const featured = computed(() => trendingData.value?.items[featuredIndex.value] || trendingData.value?.items[0]);
watch([trendingType, trendingTag, trendingPage], () => { featuredIndex.value = 0; });

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
  <div class="ytv-discover">
    <div class="ytv-container ytv-home">
      <div class="ytv-page-heading">
        <div><p class="ytv-eyebrow">THE MOVIE MOMENT</p><h1>发现好电影</h1></div>
        <div class="ytv-home-search"><SearchCommand compact /></div>
      </div>

      <section v-if="featured" class="ytv-spotlight" aria-label="热门影片推荐">
        <div class="ytv-feature">
          <img v-if="trendingPoster(featured)" :src="trendingPoster(featured)" alt="" class="ytv-feature-wash" aria-hidden="true">
          <div class="ytv-feature-copy">
            <p class="ytv-eyebrow"><span class="ytv-live-dot" /> 此刻热门 · {{ trendingType === 'movie' ? '电影' : '电视剧' }}</p>
            <h2>{{ featured.title }}</h2>
            <p class="ytv-feature-meta"><span v-if="featured.score"><UIcon name="i-tabler-star-filled" /> {{ featured.score }} <span>豆瓣评分</span></span><span v-if="featured.isNew">新近上榜</span></p>
            <UButton label="查找片源" icon="i-tabler-player-play-filled" :to="{ path: '/search', query: { q: featured.title } }" size="lg" class="ytv-feature-action" @click.prevent="searchTrending(featured.title)" />
            <div class="ytv-feature-select" role="group" aria-label="切换热门推荐">
              <button v-for="(item, index) in trendingData?.items.slice(0, 4)" :key="item.id" :aria-label="`推荐 ${index + 1}：${item.title}`" :aria-pressed="featuredIndex === index" @click="featuredIndex = index"><span>{{ String(index + 1).padStart(2, '0') }}</span></button>
            </div>
          </div>
          <NuxtLink :to="{ path: '/search', query: { q: featured.title } }" class="ytv-feature-poster" :aria-label="`查找 ${featured.title}`" @click.prevent="searchTrending(featured.title)">
            <img v-if="trendingPoster(featured)" :src="trendingPoster(featured)" :alt="`${featured.title} 海报`" fetchpriority="high" @error="failTrendingPoster(featured.id)">
            <UIcon v-else name="i-tabler-movie" class="size-10 text-white/25" />
          </NuxtLink>
        </div>
        <aside class="ytv-shortlist" aria-label="热门速览">
          <div class="ytv-shortlist-heading"><h2>热门速览</h2><span>豆瓣热门</span></div>
          <NuxtLink v-for="(item, index) in trendingData?.items.slice(1, 4)" :key="item.id" :to="{ path: '/search', query: { q: item.title } }" class="ytv-shortlist-item" @click.prevent="searchTrending(item.title)">
            <span class="ytv-shortlist-number">{{ String(index + 2).padStart(2, '0') }}</span>
            <img v-if="trendingPoster(item)" :src="trendingPoster(item)" :alt="`${item.title} 海报`" @error="failTrendingPoster(item.id)"><span v-else class="ytv-shortlist-placeholder"><UIcon name="i-tabler-movie" /></span>
            <div><h3>{{ item.title }}</h3><p><span v-if="item.score" class="text-amber-300">{{ item.score }}</span><span>{{ item.isNew ? '新近上榜' : '热门关注' }}</span></p></div>
            <UIcon name="i-tabler-chevron-right" class="size-4 shrink-0 text-white/30" />
          </NuxtLink>
        </aside>
      </section>

      <HomeContinueWatching />

      <section class="ytv-section" aria-labelledby="douban-hot-title">
        <div class="ytv-section-heading">
          <div class="flex flex-wrap items-center gap-4"><h2 id="douban-hot-title">大家在看</h2><div class="ytv-segment" role="group" aria-label="热门类型"><button :aria-pressed="trendingType === 'movie'" @click="selectType('movie')">电影</button><button :aria-pressed="trendingType === 'tv'" @click="selectType('tv')">电视剧</button></div></div>
          <UButton label="换一批热门" icon="i-tabler-refresh" color="neutral" variant="ghost" size="sm" :loading="trendingPending" @click="changeTrendingBatch" />
        </div>
        <div v-if="trendingType === 'tv'" class="flex flex-wrap gap-1"><UButton v-for="item in tags" :key="item" :label="item" :color="trendingTag === item ? 'primary' : 'neutral'" :variant="trendingTag === item ? 'soft' : 'ghost'" size="sm" @click="selectTag(item)" /></div>
        <div v-if="trendingData?.items.length" class="relative">
          <UButton v-if="canPreviousTrending" data-carousel-previous icon="i-tabler-chevron-left" aria-label="上一组豆瓣热门" color="neutral" variant="outline" class="ytv-rail-arrow -left-4" @click="previousTrending" />
          <div ref="trendingRail" data-douban-grid class="ytv-trending-rail" aria-label="豆瓣热门搜索">
            <NuxtLink v-for="item in trendingData.items" :key="item.id" :to="{ path: '/search', query: { q: item.title } }" class="ytv-poster-card" @click.prevent="searchTrending(item.title)">
              <div class="ytv-poster-image">
                <img v-if="trendingPoster(item)" :src="trendingPoster(item)" :alt="`${item.title} 海报`" loading="lazy" decoding="async" @error="failTrendingPoster(item.id)">
                <UIcon v-else name="i-tabler-photo" class="size-9 text-white/20" aria-hidden="true" />
                <span class="ytv-poster-play"><UIcon name="i-tabler-search" /></span>
                <span v-if="item.isNew" class="ytv-new-label">新片</span>
              </div>
              <div class="ytv-poster-caption"><h3>{{ item.title }}</h3><span v-if="item.score" class="ytv-poster-score">{{ item.score }}</span></div>
            </NuxtLink>
          </div>
          <UButton v-if="canNextTrending" data-carousel-next icon="i-tabler-chevron-right" aria-label="下一组豆瓣热门" color="neutral" variant="outline" class="ytv-rail-arrow -right-4" @click="nextTrending" />
        </div>
        <p v-else-if="trendingError" class="text-sm text-white/50">热门列表暂时不可用，仍可搜索想看的片名。</p>
      </section>

      <section class="ytv-section" aria-labelledby="source-updates-title">
        <div class="ytv-section-heading"><div><h2 id="source-updates-title">片库新上架</h2><p class="mt-1.5 text-xs text-white/45">电影、剧集与更多新故事</p></div><UButton label="换一批资源" icon="i-tabler-refresh" color="neutral" variant="ghost" size="sm" :loading="sourcePending" @click="changeSourceBatch" /></div>
        <UAlert v-if="sourceError" color="error" variant="subtle" title="来源更新暂时不可用"><template #actions><UButton label="重试" color="error" variant="soft" @click="() => refreshSources()" /></template></UAlert>
        <p v-else-if="sourceData && !sourceData.complete" class="ytv-quiet-notice"><UIcon name="i-tabler-info-circle" class="size-4 shrink-0" />部分来源暂未返回，先看看这些影片。<button @click="() => refreshSources()">重试</button></p>
        <div v-if="sourcePending" class="ytv-poster-grid"><USkeleton v-for="item in 12" :key="item" class="aspect-[2/3] rounded-lg bg-white/10" /></div>
        <PosterGrid v-else-if="sourceData?.groups.length" :groups="sourceData.groups.slice(0, 12)" />
        <UEmpty v-else-if="!sourceError" icon="i-tabler-database-off" title="当前批次没有来源内容" description="换一批资源，或直接搜索片名。" />
      </section>
    </div>
  </div>
</template>
