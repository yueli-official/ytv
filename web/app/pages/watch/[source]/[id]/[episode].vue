<script setup lang="ts">
definePageMeta({ colorMode: "dark" });
import { ytvFailureMessage } from "~/utils/ytv-failure";
import { createBrowserUUID } from "@yueli/nuxt-runtime/runtime";
import { browserVisitorID } from "~/utils/visitor";
import { isSingleFeature, titleLines, watchPath, searchReturnPath } from "~/utils/title-lines";
import type { PlaybackSelection } from "~/utils/playback-session";
import type { LocalPlaybackProgress } from "~/utils/local-progress";
const route = useRoute();
const api = useYtvApi();
const { loggedIn } = useAuth();
const sourceKey = computed(() => String(route.params.source || ""));
const externalId = computed(() => String(route.params.id || ""));
const episodeKey = computed(() => String(route.params.episode || ""));
const contextQuery = computed(() => { const from = searchReturnPath(route.query.from); return from ? { from } : {}; });
const titleLink = computed(() => ({ path: `/title/${encodeURIComponent(sourceKey.value)}/${encodeURIComponent(externalId.value)}`, query: contextQuery.value }));
const playbackLink = (episode: string) => ({ path: watchPath(sourceKey.value, externalId.value, episode), query: contextQuery.value });
const { data, pending, error, refresh } = await useAsyncData(
  () => `ytv-watch-title-${sourceKey.value}-${externalId.value}`,
  () => api.title(sourceKey.value, externalId.value)
);
const lines = computed(() => titleLines(data.value));
const currentLine = computed(() => lines.value.find((line) => line.episodes.some((item) => item.key === episodeKey.value)) || lines.value[0]);
const episodes = computed(() => currentLine.value?.episodes || []);
const episodeIndex = computed(() => episodes.value.findIndex((item) => item.key === episodeKey.value));
const movie = computed(() => isSingleFeature(data.value));
const currentEpisode = computed(() => episodes.value[episodeIndex.value]);
const currentEpisodeTitle = computed(() => movie.value ? (currentEpisode.value?.title.replace(/^第\s*1\s*集$/u, "正片") || "正片") : currentEpisode.value?.title || "播放");
const nextEpisode = computed(() => !movie.value && episodeIndex.value >= 0 ? episodes.value[episodeIndex.value + 1] : undefined);
const previousEpisode = computed(() => !movie.value && episodeIndex.value > 0 ? episodes.value[episodeIndex.value - 1] : undefined);
const rangeStart = ref(0);
const ranges = computed(() => Array.from({ length: Math.ceil(episodes.value.length / 50) }, (_, index) => ({ label: `${index * 50 + 1}–${Math.min((index + 1) * 50, episodes.value.length)} 集`, value: index * 50 })));
const visibleEpisodes = computed(() => episodes.value.slice(rangeStart.value, rangeStart.value + 50));
watch(episodeIndex, (index) => { rangeStart.value = Math.max(0, Math.floor(index / 50)) * 50; }, { immediate: true });
const lineOptions = computed(() => lines.value.map((line) => ({ label: line.name, value: line.key })));
const selectedLine = computed({ get: () => currentLine.value?.key || "0", set: (key: string) => {
  const first = lines.value.find((line) => line.key === key)?.episodes[0];
  if (first) void navigateTo(playbackLink(first.key));
} });
const theatre = ref(false);
function toggleTheatre() { theatre.value = !theatre.value; }
function showEpisodes() { theatre.value = false; }
async function recordPlaybackStart(value: PlaybackSelection) {
  await api.recordPlaybackStart({ eventId: createBrowserUUID(), sourceKey: value.sourceKey, externalId: value.externalId,
    episodeKey: value.episodeKey, visitorId: browserVisitorID(), occurredAt: new Date().toISOString() }).catch(() => undefined);
}
async function syncProgress(record: LocalPlaybackProgress) {
  if (!loggedIn.value) return;
  await api.saveProgress(record.sourceKey, record.externalId, record.episodeKey, { titleName: record.titleName, episodeTitle: record.episodeTitle,
    positionSeconds: record.position, durationSeconds: record.duration, completed: record.completed, updatedAt: new Date(record.updatedAt).toISOString() }).catch(() => undefined);
}
useSeoMeta({ title: () => data.value ? `${data.value.title.name} · ${currentEpisodeTitle.value}` : "播放" });
</script>

<template>
  <div class="ytv-watch min-h-[calc(100dvh-4rem)] text-white">
    <div class="ytv-container space-y-5 py-5">
      <div class="flex items-center justify-between gap-3 text-xs text-white/60">
        <NuxtLink :to="titleLink" class="inline-flex items-center gap-1.5 py-2 text-inherit no-underline hover:text-white"><UIcon name="i-tabler-chevron-left" class="size-4" />影片详情</NuxtLink>
        <span v-if="data" class="truncate">{{ data.title.sourceName }}<span class="px-2 text-white/20">/</span>{{ currentLine?.name }}</span>
      </div>
      <div v-if="pending" class="aspect-video max-h-[70vh] animate-pulse rounded-xl bg-white/5" role="status" aria-label="正在载入影片" />
      <UAlert v-else-if="error" color="error" title="影片暂时无法打开" :description="ytvFailureMessage(error, '请重试或返回搜索选择其他来源。')"><template #actions><UButton label="重试" @click="() => refresh()" /></template></UAlert>
      <div v-else-if="data" class="grid items-start gap-5" :class="theatre ? '' : 'lg:grid-cols-[minmax(0,1fr)_17rem]'">
        <div class="min-w-0 space-y-4">
          <section class="relative aspect-video overflow-hidden rounded-xl border border-white/10 bg-black shadow-2xl" aria-label="播放器区域">
            <YtvPlayer v-if="currentEpisode" :source-key="sourceKey" :external-id="externalId" :episode-key="episodeKey" :title="data.title.name" :episode-label="currentEpisodeTitle" :poster="data.title.imageUrl" @started="recordPlaybackStart" @progress="syncProgress" />
            <div v-else class="grid h-full place-items-center p-5 text-sm text-white/60">这个播放条目已不可用，请选择下方线路。</div>
          </section>
          <div class="flex flex-wrap items-start justify-between gap-4">
            <div class="min-w-0 space-y-1.5"><h1 class="text-xl font-semibold tracking-tight sm:text-2xl">{{ data.title.name }}</h1><p class="text-xs text-white/60">{{ [data.title.year, data.title.category, data.title.language, currentEpisodeTitle].filter(Boolean).join(' · ') }}</p></div>
            <div class="flex shrink-0 gap-1">
              <UButton v-if="previousEpisode" label="上一集" icon="i-tabler-player-skip-back" :to="playbackLink(previousEpisode.key)" color="neutral" variant="ghost" size="sm" />
              <UButton v-if="nextEpisode" label="下一集" trailing-icon="i-tabler-player-skip-forward" :to="playbackLink(nextEpisode.key)" color="neutral" variant="soft" size="sm" />
              <UButton :label="theatre ? '退出宽屏' : '宽屏'" icon="i-tabler-layout-bottombar-collapse" color="neutral" variant="ghost" size="sm" class="hidden lg:inline-flex" :aria-pressed="theatre" @click="toggleTheatre" />
            </div>
          </div>
          <div v-if="theatre" class="flex flex-wrap items-center gap-2 rounded-xl border border-white/10 p-3"><span class="text-xs text-white/60">线路</span><USelect v-model="selectedLine" :items="lineOptions" aria-label="播放线路" class="w-44" /><UButton v-if="!movie" label="展开选集" color="neutral" variant="ghost" @click="showEpisodes" /></div>
          <details class="rounded-xl border border-white/10 bg-white/[0.02] p-4"><summary class="cursor-pointer text-sm text-white/65">播放操作</summary><p class="mt-3 text-sm leading-7 text-white/65">播放时按住画面可临时 2 倍速，松开恢复原倍速。双击画面左侧快退 10 秒，右侧快进 10 秒；也可拖动进度条。电脑上先点击播放器，再用 ← / → 跳转，空格播放或暂停，F 全屏。倍速、字幕和音轨可在播放器设置中选择，具体选项取决于片源。</p></details>
          <details class="rounded-xl border border-white/10 bg-white/[0.02] p-4"><summary class="cursor-pointer text-sm text-white/65">剧情简介</summary><p class="mt-3 max-w-4xl text-sm leading-7 text-white/60">{{ data.title.summary || '当前来源暂未提供简介。' }}</p></details>
        </div>
        <aside v-if="!theatre" class="overflow-hidden rounded-xl border border-white/10 bg-[#101720]" :aria-label="movie ? '播放线路' : '播放列表'">
          <div class="space-y-4 border-b border-white/8 p-4">
            <div class="flex items-center justify-between"><h2 class="text-sm font-semibold">{{ movie ? '观看选项' : '选集' }}</h2><span class="text-[11px] text-white/60">{{ movie ? `${lines.length} 条线路` : `${episodes.length} 集` }}</span></div>
            <USelect v-model="selectedLine" :items="lineOptions" aria-label="播放线路" class="w-full" />
            <USelect v-if="ranges.length > 1" v-model="rangeStart" :items="ranges" aria-label="剧集范围" class="w-full" />
          </div>
          <div class="max-h-[28rem] space-y-1 overflow-y-auto p-2">
            <NuxtLink v-for="episode in visibleEpisodes" :key="episode.key" :to="playbackLink(episode.key)" data-player-episode-row class="flex min-h-11 items-center gap-3 rounded-lg px-3 py-2 text-sm no-underline transition" :class="episode.key === episodeKey ? 'bg-sky-400/10 text-sky-300' : 'text-white/55 hover:bg-white/5 hover:text-white'" :aria-current="episode.key === episodeKey ? 'page' : undefined">
              <UIcon :name="episode.key === episodeKey ? 'i-tabler-player-play-filled' : 'i-tabler-player-play'" class="size-4 shrink-0" /><span class="min-w-0 break-words">{{ movie ? episode.title.replace(/^第\s*1\s*集$/u, '正片') || '正片' : episode.title }}</span>
            </NuxtLink>
          </div>
          <div class="border-t border-white/8 p-4"><NuxtLink :to="{ path: '/search', query: { q: data.title.name } }" class="inline-flex items-center gap-2 text-xs text-white/60 no-underline hover:text-sky-300"><UIcon name="i-tabler-search" class="size-4" />查找其他来源</NuxtLink><NuxtLink v-if="contextQuery.from" :to="contextQuery.from" class="mt-3 block text-xs text-white/60 no-underline hover:text-sky-300">返回搜索结果</NuxtLink></div>
        </aside>
      </div>
    </div>
  </div>
</template>
