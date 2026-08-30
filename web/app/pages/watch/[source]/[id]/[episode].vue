<script setup lang="ts">
import { createBrowserUUID } from "@yueli/nuxt-runtime/runtime";
import { browserVisitorID } from "~/utils/visitor";
import type { LocalPlaybackProgress } from "~/utils/local-progress";
const route = useRoute();
const api = useYtvApi();
const { loggedIn } = useAuth();
const sourceKey = computed(() => String(route.params.source || ""));
const externalId = computed(() => String(route.params.id || ""));
const episodeKey = computed(() => String(route.params.episode || ""));
const { data: title } = await useAsyncData(
  () => `ytv-watch-title-${sourceKey.value}-${externalId.value}`,
  () => api.title(sourceKey.value, externalId.value)
);
const { data: manifest, pending, error, refresh } = await useAsyncData(
  () => `ytv-playback-${sourceKey.value}-${externalId.value}-${episodeKey.value}`,
  () => api.playback(sourceKey.value, externalId.value, episodeKey.value)
);
const episodeIndex = computed(() => title.value?.episodes.findIndex((item) => item.key === episodeKey.value) ?? -1);
const previousEpisode = computed(() => episodeIndex.value > 0 ? title.value?.episodes[episodeIndex.value - 1] : null);
const nextEpisode = computed(() => episodeIndex.value >= 0 ? title.value?.episodes[episodeIndex.value + 1] || null : null);
const episodeRangeStart = ref(0);
const episodePageSize = 50;
const episodeList = ref<HTMLElement | null>(null);
const episodeRanges = computed(() => {
  const total = title.value?.episodes.length || 0;
  return Array.from({ length: Math.ceil(total / episodePageSize) }, (_, index) => {
    const start = index * episodePageSize;
    return { label: `${start + 1}–${Math.min(start + episodePageSize, total)}集`, value: start };
  });
});
const playlistEpisodes = computed(() => (title.value?.episodes || []).slice(episodeRangeStart.value, episodeRangeStart.value + episodePageSize));
const currentEpisodeTitle = computed(() => title.value?.episodes[episodeIndex.value]?.title || `第 ${Number(episodeKey.value) + 1} 集`);

watch([episodeIndex, () => title.value?.episodes.length], ([index]) => {
  if (typeof index === "number" && index >= 0) episodeRangeStart.value = Math.floor(index / episodePageSize) * episodePageSize;
}, { immediate: true });

watch([episodeRangeStart, episodeKey, () => title.value?.episodes.length], async () => {
  await nextTick();
  if (!import.meta.client || !window.matchMedia("(min-width: 64rem)").matches) return;
  episodeList.value?.querySelector<HTMLElement>('[aria-current="page"]')?.scrollIntoView({ block: "center" });
}, { flush: "post" });

function episodeNumber(key: string) {
  const index = title.value?.episodes.findIndex((episode) => episode.key === key) ?? -1;
  return index >= 0 ? index + 1 : Number(key) + 1;
}

function episodeRowTitle(label: string, number: number) {
  const trimmed = label.trim();
  const withoutPrefix = trimmed.replace(new RegExp(`^第?\\s*${number}\\s*集?\\s*`), "").trim();
  return withoutPrefix && withoutPrefix !== trimmed ? withoutPrefix : /^第?\s*\d+\s*集?$/u.test(trimmed) ? "" : trimmed;
}

async function retry() {
  await refresh();
}

async function recordPlaybackStart() {
  await api.recordPlaybackStart({
    eventId: createBrowserUUID(),
    sourceKey: sourceKey.value,
    externalId: externalId.value,
    episodeKey: episodeKey.value,
    visitorId: browserVisitorID(),
    occurredAt: new Date().toISOString()
  }).catch(() => undefined);
}

async function syncProgress(record: LocalPlaybackProgress) {
  if (!loggedIn.value) return;
  await api.saveProgress(record.sourceKey, record.externalId, record.episodeKey, {
    titleName: record.titleName,
    episodeTitle: record.episodeTitle,
    positionSeconds: record.position,
    durationSeconds: record.duration,
    completed: record.completed,
    updatedAt: new Date(record.updatedAt).toISOString()
  }).catch(() => undefined);
}

useSeoMeta({ title: () => `${title.value?.title.name || '播放'} · ${episodeKey.value}` });
</script>

<template>
  <div class="min-h-[calc(100dvh-4rem)] bg-[#05080d] text-white">
    <div class="mx-auto max-w-[90rem] space-y-4 px-3 py-4 sm:px-6 lg:px-8 lg:py-6">
      <div class="flex min-w-0 flex-wrap items-center gap-x-4 gap-y-2">
        <NuxtLink :to="`/title/${sourceKey}/${externalId}`" class="inline-flex shrink-0 items-center gap-2 text-sm text-white/60 no-underline hover:text-sky-300"><UIcon name="i-tabler-arrow-left" />返回详情</NuxtLink>
        <div class="min-w-0 flex-1"><p class="truncate text-sm font-semibold text-white">{{ title?.title.name || '正在读取标题' }} · {{ currentEpisodeTitle }}</p></div>
        <UBadge v-if="title?.title.sourceName" :label="title.title.sourceName" color="primary" variant="subtle" />
      </div>

      <div class="player-watch-layout grid gap-3 lg:items-start">
        <div class="flex min-w-0 flex-col gap-3">
          <section class="relative grid aspect-video place-items-center overflow-hidden rounded-xl border border-white/10 bg-black shadow-2xl" aria-label="播放器区域">
            <div v-if="pending" class="grid justify-items-center gap-4"><UIcon name="i-tabler-loader-2" class="size-8 animate-spin text-primary" /><p class="text-sm text-white/60">正在解析播放清单</p></div>
            <UAlert v-else-if="error" color="error" title="当前来源无法播放" description="可以重试当前来源，或返回详情选择其他来源。"><template #actions><UButton label="重试" color="error" variant="soft" @click="retry" /></template></UAlert>
            <VideoPlayer v-else-if="manifest" :manifest="manifest" :source-key="sourceKey" :external-id="externalId" :title="title?.title.name || '演示内容'" :episode-label="currentEpisodeTitle" :poster="title?.title.imageUrl || '/demo/hero-gate.png'" @started="recordPlaybackStart" @progress="syncProgress" />
          </section>
          <div class="flex items-center justify-between gap-3 rounded-lg border border-white/10 bg-[#10151c] p-2">
            <UButton :disabled="!previousEpisode" :to="previousEpisode ? `/watch/${sourceKey}/${externalId}/${previousEpisode.key}` : undefined" label="上一集" icon="i-tabler-player-skip-back" color="neutral" variant="ghost" class="min-h-11" />
            <p class="truncate text-xs text-white/55">第 {{ episodeIndex + 1 }} / {{ title?.episodes.length || 0 }} 集</p>
            <UButton :disabled="!nextEpisode" :to="nextEpisode ? `/watch/${sourceKey}/${externalId}/${nextEpisode.key}` : undefined" label="下一集" icon="i-tabler-player-skip-forward" trailing color="neutral" variant="ghost" class="min-h-11" />
          </div>
        </div>

        <aside class="player-playlist flex min-h-0 flex-col overflow-visible rounded-xl border border-white/10 bg-[#10151c] lg:overflow-hidden" aria-label="播放列表">
          <div class="sticky top-16 z-10 flex min-h-14 items-center justify-between gap-3 border-b border-white/10 bg-[#10151c] p-3 lg:static">
            <div class="min-w-0"><h2 class="font-bold">选集</h2><p class="mt-0.5 text-xs text-white/45">当前第 {{ episodeIndex + 1 }} / {{ title?.episodes.length || 0 }} 集</p></div>
            <USelect v-if="episodeRanges.length > 1" v-model="episodeRangeStart" :items="episodeRanges" value-key="value" aria-label="剧集范围" class="w-32 shrink-0" :ui="{ base: 'min-h-11 text-xs' }" />
          </div>
          <div ref="episodeList" class="divide-y divide-white/8 lg:min-h-0 lg:flex-1 lg:overflow-y-auto">
            <NuxtLink
              v-for="episode in playlistEpisodes"
              :key="episode.key"
              :to="`/watch/${sourceKey}/${externalId}/${episode.key}`"
              data-player-episode-row
              class="group flex min-h-11 min-w-0 items-center gap-2 border-s-2 px-3 py-2 text-[13px] leading-5 text-white/72 no-underline transition hover:bg-white/5 hover:text-white"
              :class="episode.key === episodeKey ? 'border-s-sky-400 bg-sky-400/10 text-white' : 'border-s-transparent'"
              :aria-current="episode.key === episodeKey ? 'page' : undefined"
            >
              <span class="shrink-0 font-semibold tabular-nums" :class="[episodeRowTitle(episode.title, episodeNumber(episode.key)) ? 'w-10 text-center' : 'min-w-0 flex-1', episode.key === episodeKey ? 'text-sky-300' : 'text-white/45']">{{ episodeRowTitle(episode.title, episodeNumber(episode.key)) ? episodeNumber(episode.key) : `第${episodeNumber(episode.key)}集` }}</span>
              <span v-if="episodeRowTitle(episode.title, episodeNumber(episode.key))" class="min-w-0 flex-1 truncate">{{ episodeRowTitle(episode.title, episodeNumber(episode.key)) }}</span>
              <UIcon v-if="episode.key === episodeKey" name="i-tabler-player-play-filled" class="size-4 shrink-0 text-sky-300" aria-hidden="true" />
            </NuxtLink>
          </div>
        </aside>
      </div>
    </div>
  </div>
</template>

<style scoped>
@media (min-width: 64rem) {
  .player-watch-layout {
    --playlist-width: clamp(20rem, 24vw, 22rem);
    grid-template-columns: minmax(0, 1fr) var(--playlist-width);
  }

  .player-playlist {
    height: calc((min(90rem, 100vw) - 4rem - var(--playlist-width) - 0.75rem) * 0.5625 + 4.5rem);
  }
}
</style>
