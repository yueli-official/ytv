<script setup lang="ts">
import Playlist from "~/components/public/shared/Playlist.vue";
const {mobile}=usePublicLayout();
const episodesOpen=ref(false);
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
watch(episodeKey,()=>{episodesOpen.value=false;});
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
  const target = lines.value.find((line) => line.key === key);
  const episode = target?.episodes.find(item => item.title === currentEpisode.value?.title) || target?.episodes[0];
  if (episode) void navigateTo(playbackLink(episode.key));
} });
const alternateEpisode = computed(() => {
  const line = lines.value.find(item => item.key !== currentLine.value?.key);
  return line?.episodes.find(item => item.title === currentEpisode.value?.title) || line?.episodes[0];
});
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
    <div class="ytv-container py-7">
      <UButton v-if="mobile" :to="titleLink" label="影片详情" icon="i-tabler-chevron-left" variant="ghost" color="neutral" class="mb-3"/>
      <div v-if="pending" class="aspect-video max-h-[70vh] animate-pulse rounded-xl bg-white/5" role="status" aria-label="正在载入影片" />
      <UAlert v-else-if="error" color="error" title="影片暂时无法打开" :description="ytvFailureMessage(error, '请重试或返回搜索选择其他来源。')"><template #actions><UButton label="重试" @click="() => refresh()" /></template></UAlert>
      <div v-else-if="data" class="ytv-watch-grid" :class="{ wide: theatre }">
        <section class="ytv-watch-player relative aspect-video overflow-hidden rounded-xl border border-white/10 bg-black shadow-2xl" aria-label="播放器区域">
          <YtvPlayer v-if="currentEpisode" :source-key="sourceKey" :external-id="externalId" :episode-key="episodeKey" :title="data.title.name" :episode-label="currentEpisodeTitle" :poster="data.title.imageUrl" @started="recordPlaybackStart" @progress="syncProgress">
            <template #recovery>
              <UButton v-if="alternateEpisode" :to="playbackLink(alternateEpisode.key)" label="换条线路" icon="i-tabler-switch-horizontal" color="neutral" variant="outline" />
              <UButton :to="{ path: '/search', query: { q: data.title.name } }" label="换个来源" icon="i-tabler-search" color="neutral" variant="outline" />
            </template>
          </YtvPlayer>
          <div v-else class="grid h-full place-items-center p-5 text-white/60">这个播放条目已不可用，请选择下方线路。</div>
        </section>
        <Playlist v-if="!mobile && !theatre" v-model:line="selectedLine" v-model:range="rangeStart" :title="data.title" :movie="movie" :episodes="visibleEpisodes" :total="episodes.length" :line-options="lineOptions" :line-name="currentLine?.name" :ranges="ranges" :episode-key="episodeKey" :playback-link="playbackLink"/>
        <div class="ytv-watch-meta">
          <div class="min-w-0 space-y-3"><h1>{{ data.title.name }}<span class="ml-3 text-base font-normal text-white/75">{{ currentEpisodeTitle }}</span></h1><div class="flex flex-wrap gap-x-5 gap-y-2 text-sm text-white/70"><p>{{ [data.title.year, data.title.category, data.title.language].filter(Boolean).join(' · ') }}</p><NuxtLink :to="titleLink" class="text-inherit no-underline hover:text-white">影片详情</NuxtLink><NuxtLink v-if="contextQuery.from" :to="contextQuery.from" class="text-inherit no-underline hover:text-white">返回搜索结果</NuxtLink></div></div>
          <div class="flex flex-wrap gap-3">
            <UButton v-if="previousEpisode && !mobile" label="上一集" icon="i-tabler-player-skip-back" :to="playbackLink(previousEpisode.key)" color="neutral" variant="ghost" class="ytv-secondary" />
            <UButton v-if="nextEpisode" label="下一集" trailing-icon="i-tabler-player-skip-forward" :to="playbackLink(nextEpisode.key)" color="neutral" class="ytv-primary" />
            <UButton :label="theatre ? '退出宽屏' : '宽屏'" icon="i-tabler-layout-bottombar-collapse" color="neutral" variant="ghost" class="ytv-secondary hidden lg:inline-flex" :aria-pressed="theatre" @click="toggleTheatre" />
          </div>
        </div>
        <div v-if="mobile" class="mobile-watch-actions"><UButton label="选集" icon="i-tabler-list" color="neutral" variant="outline" size="sm" @click="episodesOpen=true"/><UButton :to="{path:'/search',query:{q:data.title.name}}" label="换源" icon="i-tabler-switch-horizontal" color="neutral" variant="outline" size="sm"/></div>
        <USlideover v-if="mobile" v-model:open="episodesOpen" side="bottom" title="选集" description="选择播放线路与集数" :ui="{content:'bg-[#141e28] text-white max-h-[80dvh]',body:'p-4'}"><template #body><Playlist compact v-model:line="selectedLine" v-model:range="rangeStart" :title="data.title" :movie="movie" :episodes="visibleEpisodes" :total="episodes.length" :line-options="lineOptions" :line-name="currentLine?.name" :ranges="ranges" :episode-key="episodeKey" :playback-link="playbackLink"/></template></USlideover>
        <div class="ytv-watch-help space-y-4">
          <div v-if="theatre" class="flex flex-wrap items-center gap-3 rounded-xl border border-white/10 p-4"><span class="text-white/70">线路</span><USelect v-if="lines.length > 1" v-model="selectedLine" :items="lineOptions" aria-label="播放线路" class="w-44" /><span v-else class="text-white/70">{{ currentLine?.name }}</span><UButton v-if="!movie" label="展开选集" color="neutral" variant="ghost" @click="showEpisodes" /></div>
          <details class="ytv-fold"><summary>播放操作</summary><p class="mt-3 leading-8 text-white/70">播放时按住画面可临时 2 倍速，松开恢复原倍速。双击画面左侧快退 10 秒，右侧快进 10 秒；也可拖动进度条。电脑上先点击播放器，再用 ← / → 跳转，空格播放或暂停，F 全屏。倍速、字幕和音轨可在播放器设置中选择，具体选项取决于片源。</p></details>
          <details class="ytv-fold"><summary>剧情简介</summary><p class="mt-3 max-w-5xl leading-8 text-white/70">{{ data.title.summary || '当前来源暂未提供简介。' }}</p></details>
        </div>
      </div>
    </div>
  </div>
</template>

<style scoped>
@media(max-width:767px){
.ytv-watch>.ytv-container{width:100%;padding:12px 0 24px}
.ytv-watch>.ytv-container>button,.ytv-watch>.ytv-container>a{margin-left:12px}
.ytv-watch-grid,.ytv-watch-grid.wide{display:flex;flex-direction:column;gap:16px}
.ytv-watch-player{border-radius:0;border-inline:0;box-shadow:none;width:100%}
.ytv-watch-meta{padding:0 16px;align-items:start;gap:12px}
.ytv-watch-meta h1{font-size:19px;line-height:1.5}
.ytv-watch-meta h1>span{display:block;font-size:12px;margin:4px 0 0}
.ytv-watch-meta>div:first-child>div{display:none}
.ytv-watch-meta>div:last-child{flex-shrink:0}
.mobile-watch-actions{display:flex;gap:10px;padding:0 16px}
.ytv-watch-help{padding:0 16px;width:100%}
.ytv-watch-meta,.mobile-watch-actions{width:100%}
.ytv-fold{padding:12px;font-size:13px}
.ytv-fold summary{font-size:13px}
.ytv-fold p{font-size:13px;line-height:1.8}
}
</style>
