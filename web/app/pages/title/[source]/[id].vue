<script setup lang="ts">
const route = useRoute();
const api = useYtvApi();
const sourceKey = computed(() => String(route.params.source || ""));
const externalId = computed(() => String(route.params.id || ""));
const episodeRangeStart = ref(0);
const episodePageSize = 50;
const { data, pending, error, refresh } = await useAsyncData(
  () => `ytv-title-${sourceKey.value}-${externalId.value}`,
  () => api.title(sourceKey.value, externalId.value),
  { watch: [sourceKey, externalId] }
);

watch([sourceKey, externalId], () => { episodeRangeStart.value = 0; });
const visibleEpisodes = computed(() => data.value?.episodes.slice(episodeRangeStart.value, episodeRangeStart.value + episodePageSize) || []);
const episodeRanges = computed(() => {
  const total = data.value?.episodes.length || 0;
  return Array.from({ length: Math.ceil(total / episodePageSize) }, (_, index) => {
    const start = index * episodePageSize;
    return { label: `${start + 1}–${Math.min(start + episodePageSize, total)} 集`, value: start };
  });
});
const firstEpisode = computed(() => data.value?.episodes[0] || null);

async function retry() {
  await refresh();
}

useSeoMeta({ title: () => data.value?.title.name || "影片详情" });
</script>

<template>
  <div class="min-h-[calc(100dvh-4rem)] bg-[#080b10] text-white">
    <div class="mx-auto max-w-6xl space-y-6 px-4 py-7 sm:px-6 lg:px-8">
      <NuxtLink to="/" class="inline-flex items-center gap-2 text-sm text-white/60 no-underline hover:text-sky-300"><UIcon name="i-tabler-arrow-left" />返回浏览</NuxtLink>

      <div v-if="pending" class="grid items-start gap-6 md:grid-cols-[15rem_minmax(0,1fr)]">
        <USkeleton class="aspect-[2/3] rounded-xl bg-white/10" />
        <div class="space-y-4 rounded-xl border border-white/10 bg-[#121821] p-6"><USkeleton class="h-10 w-2/3 bg-white/10" /><USkeleton class="h-5 w-1/2 bg-white/10" /><USkeleton class="h-24 w-full bg-white/10" /></div>
      </div>
      <UAlert v-else-if="error" color="error" title="影片详情暂时不可用" description="当前来源可能已失效。请返回搜索选择其他来源，或稍后重试。">
        <template #actions><UButton label="重试" color="error" variant="soft" @click="retry" /></template>
      </UAlert>

      <template v-else-if="data">
        <div class="grid items-start gap-6 md:grid-cols-[15rem_minmax(0,1fr)] lg:gap-8">
          <div class="mx-auto aspect-[2/3] w-full max-w-[13rem] self-start overflow-hidden rounded-xl border border-white/10 bg-[#121821] shadow-xl md:mx-0 md:max-w-none">
            <img v-if="data.title.imageUrl" :src="data.title.imageUrl" :alt="`${data.title.name} 海报`" class="size-full object-cover" referrerpolicy="no-referrer">
            <div v-else class="grid size-full place-items-center"><UIcon name="i-tabler-photo-off" class="size-10 text-white/20" /></div>
          </div>

          <section class="space-y-5 rounded-xl border border-white/10 bg-[#121821] p-5 sm:p-7">
            <div class="flex flex-wrap items-center gap-2">
              <UBadge :label="data.title.sourceName" color="primary" variant="subtle" />
              <UBadge v-if="data.title.score && Number(data.title.score) > 0" :label="`★ ${Number(data.title.score).toFixed(1)}`" color="warning" variant="subtle" />
              <UBadge v-if="data.title.remarks" :label="data.title.remarks" color="error" variant="subtle" />
            </div>
            <div><h1 class="text-3xl font-black tracking-[-0.035em] text-white sm:text-4xl">{{ data.title.name }}</h1><p class="mt-3 text-sm text-white/55">{{ data.title.year || '年份未知' }} · {{ data.title.category || '分类未知' }} · {{ data.title.area || data.title.language || '地区未知' }}</p></div>
            <p v-if="data.title.summary" class="max-w-3xl leading-7 text-white/72">{{ data.title.summary }}</p>
            <p v-else class="text-sm text-white/45">当前来源没有提供剧情简介。</p>
            <div class="flex flex-wrap gap-3">
              <UButton v-if="firstEpisode" label="开始播放" icon="i-tabler-player-play-filled" size="lg" :to="`/watch/${sourceKey}/${externalId}/${firstEpisode.key}`" />
              <UButton label="返回搜索" icon="i-tabler-search" color="neutral" variant="soft" class="min-h-11" :to="{ path: '/search', query: { q: data.title.name } }" />
            </div>
          </section>
        </div>

        <section class="space-y-4 rounded-xl border border-white/10 bg-[#10151d] p-4 sm:p-6" aria-labelledby="episode-heading">
          <div class="flex flex-wrap items-center justify-between gap-3">
            <div><h2 id="episode-heading" class="text-lg font-bold">选择剧集</h2><p class="mt-1 text-xs text-white/50">共 {{ data.episodes.length }} 集，每次显示 {{ episodePageSize }} 集</p></div>
            <USelect v-if="episodeRanges.length > 1" v-model="episodeRangeStart" :items="episodeRanges" value-key="value" aria-label="选集范围" class="w-36" :ui="{ base: 'min-h-11' }" />
          </div>
          <div class="grid grid-cols-4 gap-2 sm:grid-cols-6 md:grid-cols-8 lg:grid-cols-10 xl:grid-cols-12">
            <UButton v-for="episode in visibleEpisodes" :key="episode.key" :label="episode.title || `第 ${Number(episode.key) + 1} 集`" :to="`/watch/${sourceKey}/${externalId}/${episode.key}`" color="neutral" variant="outline" size="sm" class="min-h-11 min-w-0 justify-center" />
          </div>
        </section>
      </template>
    </div>
  </div>
</template>
