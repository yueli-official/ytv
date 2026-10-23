<script setup lang="ts">
definePageMeta({ colorMode: "dark" });
const api = useYtvApi();
const sourcePage = ref(1);
const trendingPage = ref(1);
const trendingType = ref("movie");
const trendingTag = ref("热门");
const failedFeaturePosters = ref(new Set<string>());
const failedTrendingPosters = ref(new Set<string>());
const tags = ["热门", "美剧", "英剧", "日剧", "韩剧", "港剧", "综艺", "纪录片"];
const viewerSourceKeys = useState<string[]>("ytv-selected-source-keys", () => []);
const { rail: trendingRail, canPrevious: canPreviousTrending, canNext: canNextTrending, previous: previousTrending, next: nextTrending } = useHorizontalRail();

const { data: sourceData, pending: sourcePending, error: sourceError, refresh: refreshSources } = await useAsyncData(
  () => `ytv-discover-${sourcePage.value}`,
  () => api.discover(sourcePage.value),
  { watch: [sourcePage] }
);
const { data: trendingData, pending: trendingPending, error: trendingError, refresh: refreshTrending } = await useAsyncData(
  () => `ytv-trending-${trendingType.value}-${trendingTag.value}-${trendingPage.value}`,
  () => api.trending(trendingType.value, trendingTag.value, trendingPage.value, 12),
  { watch: [trendingType, trendingTag, trendingPage] }
);

const featureArtwork = ref<HTMLImageElement>();
const activeFeatureIndex = ref(0);
const features = computed(() => {
  const items = trendingData.value?.items || [];
  const selected = items.filter(item => Number(item.score) >= 7).slice(0, 3);
  return [...selected, ...items.filter(item => !selected.includes(item))].slice(0, 3);
});
const activeFeature = computed(() => features.value[activeFeatureIndex.value] || features.value[0]);
const remainingTrending = computed(() => trendingData.value?.items || []);
watch(features, () => { activeFeatureIndex.value = 0; });
onMounted(() => {
  const image = featureArtwork.value;
  if (image?.complete && !image.naturalWidth && image.dataset.id) failFeaturePoster(image.dataset.id);
});

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

function failFeaturePoster(id: string) {
  if (failedFeaturePosters.value.has(id)) failTrendingPoster(id);
  else failedFeaturePosters.value = new Set([...failedFeaturePosters.value, id]);
}

useSeoMeta({
  title: "发现好电影",
  description: "搜索想看的电影，浏览热门影片，接着上次的进度继续观看。"
});
</script>

<template>
  <div class="ytv-discover">
    <div class="ytv-home">
      <h1 class="sr-only">发现好电影</h1>
      <section v-if="activeFeature" class="ytv-feature" aria-label="热门影片">
        <img v-if="trendingPoster(activeFeature)" :src="trendingPoster(activeFeature)" alt="" aria-hidden="true" class="ytv-art-backdrop">
        <div class="ytv-container ytv-feature-inner"><div class="ytv-feature-copy">
          <p class="ytv-eyebrow">豆瓣热门 · {{ trendingType === 'movie' ? '电影' : '电视剧' }}</p>
          <h2>{{ activeFeature.title }}</h2>
          <p class="mt-4 flex items-baseline gap-3"><strong v-if="Number(activeFeature.score) > 0" class="text-3xl font-medium tabular-nums text-amber-300">{{ activeFeature.score }}</strong><span class="text-base text-white/80">{{ Number(activeFeature.score) > 0 ? '豆瓣评分' : '豆瓣热门' }}</span></p>
          <UButton :to="{ path: '/search', query: { q: activeFeature.title } }" label="查找片源" icon="i-tabler-search" color="neutral" variant="ghost" class="ytv-primary mt-6" @click.prevent="searchTrending(activeFeature.title)" />
          <div class="ytv-feature-selectors" role="group" aria-label="切换热门推荐">
            <button v-for="(item, index) in features" :key="item.id" :aria-pressed="index === activeFeatureIndex" :aria-label="`查看热门：${item.title}`" @click="activeFeatureIndex = index"><span aria-hidden="true">0{{ index + 1 }}</span><strong class="sr-only">{{ item.title }}</strong></button>
          </div>
        </div>
        <img v-if="trendingPoster(activeFeature)" :key="activeFeature.id" ref="featureArtwork" :data-id="activeFeature.id" :src="failedFeaturePosters.has(activeFeature.id) ? trendingPoster(activeFeature) : `${trendingPoster(activeFeature)}?size=feature`" :alt="`${activeFeature.title} 海报`" fetchpriority="high" class="ytv-feature-art" @error="failFeaturePoster(activeFeature.id)">
        <div v-else class="ytv-feature-fallback"><UIcon name="i-tabler-movie" class="size-10" /></div></div>
      </section>
      <USkeleton v-else-if="trendingPending" class="h-96 rounded-xl bg-white/5" aria-label="正在加载热门影片" />
      <div v-else class="ytv-quiet-notice">{{ trendingError ? '热门影片暂时未能加载' : '暂时没有热门影片' }}<NuxtLink to="/search" class="text-primary">搜索想看的影视</NuxtLink></div>

      <div class="ytv-container ytv-home-history"><HomeContinueWatching /></div>

      <section class="ytv-container ytv-section" aria-labelledby="douban-hot-title">
        <div class="ytv-section-heading">
          <div class="flex flex-wrap items-center gap-4"><h2 id="douban-hot-title">豆瓣热门</h2><div class="ytv-segment" role="group" aria-label="热门类型"><button :aria-pressed="trendingType === 'movie'" @click="selectType('movie')">电影</button><button :aria-pressed="trendingType === 'tv'" @click="selectType('tv')">电视剧</button></div></div>
          <UButton label="换一批热门" icon="i-tabler-refresh" color="neutral" variant="ghost" size="lg" :loading="trendingPending" @click="changeTrendingBatch" />
        </div>
        <div v-if="trendingType === 'tv'" class="flex flex-wrap gap-1"><UButton v-for="item in tags" :key="item" :label="item" :color="trendingTag === item ? 'primary' : 'neutral'" :variant="trendingTag === item ? 'soft' : 'ghost'" size="sm" @click="selectTag(item)" /></div>
        <div v-if="trendingPending" class="ytv-poster-grid" aria-label="正在更新热门列表"><USkeleton v-for="item in 6" :key="item" class="aspect-[2/3] bg-white/5" /></div>
        <div v-else-if="remainingTrending.length" class="relative">
          <UButton v-if="canPreviousTrending" data-carousel-previous icon="i-tabler-chevron-left" aria-label="上一组豆瓣热门" color="neutral" variant="outline" class="ytv-rail-arrow -left-4" @click="previousTrending" />
          <div ref="trendingRail" data-douban-grid class="ytv-trending-rail" aria-label="豆瓣热门搜索">
            <NuxtLink v-for="item in remainingTrending" :key="item.id" :to="{ path: '/search', query: { q: item.title } }" class="ytv-poster-card" @click.prevent="searchTrending(item.title)">
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
        <div v-else-if="trendingError" class="ytv-quiet-notice">热门列表暂时不可用，仍可搜索想看的片名。<button @click="() => refreshTrending()">重试</button></div>
        <p v-else class="ytv-section-note">本批热门已展示完，换一批看看。</p>
      </section>

      <section class="ytv-container ytv-section" aria-labelledby="source-updates-title">
        <div class="ytv-section-heading"><div><h2 id="source-updates-title">来源更新</h2></div><UButton label="换一批资源" icon="i-tabler-refresh" color="neutral" variant="ghost" size="lg" :loading="sourcePending" @click="changeSourceBatch" /></div>
        <UAlert v-if="sourceError" color="error" variant="subtle" title="来源更新暂时不可用"><template #actions><UButton label="重试" color="error" variant="soft" @click="() => refreshSources()" /></template></UAlert>
        <p v-else-if="sourceData && !sourceData.complete" class="ytv-quiet-notice"><UIcon name="i-tabler-info-circle" class="size-4 shrink-0" />部分来源暂未返回，先看看这些影片。<button @click="() => refreshSources()">重试</button></p>
        <div v-if="sourcePending" class="ytv-poster-grid"><USkeleton v-for="item in 12" :key="item" class="aspect-[2/3] rounded-lg bg-white/10" /></div>
        <PosterGrid v-else-if="sourceData?.groups.length" :groups="sourceData.groups.slice(0, 12)" />
        <UEmpty v-else-if="!sourceError" icon="i-tabler-database-off" title="当前批次没有来源内容" description="换一批资源，或直接搜索片名。" />
      </section>
    </div>
  </div>
</template>
