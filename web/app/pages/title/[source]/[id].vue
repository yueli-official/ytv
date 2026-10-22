<script setup lang="ts">
definePageMeta({ colorMode: "dark" });
import { isSingleFeature, titleLines, watchPath, searchReturnPath } from "~/utils/title-lines";
import type { LocalPlaybackProgress } from "~/utils/local-progress";
import { browserProgress, mergeViewerProgress } from "~/utils/viewer-progress";
const route = useRoute();
const api = useYtvApi();
const { loggedIn } = useAuth();
const returnTo = computed(() => searchReturnPath(route.query.from));
const playbackLink = (episode: string) => ({ path: watchPath(sourceKey.value, externalId.value, episode), query: returnTo.value ? { from: returnTo.value } : {} });
const sourceKey = computed(() => String(route.params.source || ""));
const externalId = computed(() => String(route.params.id || ""));
const { data, pending, error, refresh } = await useAsyncData(() => `ytv-title-${sourceKey.value}-${externalId.value}`, () => api.title(sourceKey.value, externalId.value));
const selectedLine = ref("0");
const rangeStart = ref(0);
const lines = computed(() => titleLines(data.value));
const line = computed(() => lines.value.find((item) => item.key === selectedLine.value) || lines.value[0]);
const movie = computed(() => isSingleFeature(data.value));
const episodes = computed(() => line.value?.episodes || []);
const ranges = computed(() => Array.from({ length: Math.ceil(episodes.value.length / 50) }, (_, index) => ({ label: `${index * 50 + 1}–${Math.min((index + 1) * 50, episodes.value.length)} 集`, value: index * 50 })));
const history = ref<LocalPlaybackProgress | null>(null);
const posterFailed = ref(false);
const summaryExpanded = ref(false);
function toggleSummary() { summaryExpanded.value = !summaryExpanded.value; }
let lineChosen = false;
let historyRequest = 0;
function selectLine(key: string) { lineChosen = true; selectedLine.value = key; }
function applyHistory(records: LocalPlaybackProgress[]) {
  const latest = records.find(item => item.sourceKey === sourceKey.value && item.externalId === externalId.value && data.value?.episodes.some(episode => episode.key === item.episodeKey));
  history.value = latest && !latest.completed ? latest : null;
  if (!lineChosen && history.value) {
    const previous = lines.value.find(item => item.episodes.some(episode => episode.key === history.value?.episodeKey));
    if (previous) selectedLine.value = previous.key;
    revealResume();
  }
}
function revealResume() {
  const index = episodes.value.findIndex(item => item.key === history.value?.episodeKey);
  rangeStart.value = Math.max(0, Math.floor(index / 50)) * 50;
}
async function readHistory() {
  const request = ++historyRequest;
  const local = browserProgress();
  applyHistory(local);
  if (!loggedIn.value) return;
  const account = await api.listProgress().catch(() => null);
  if (request === historyRequest && account) applyHistory(mergeViewerProgress(local, account.items));
}
onMounted(readHistory);
onBeforeUnmount(() => { historyRequest++; });
watch([sourceKey, externalId], () => { historyRequest++; lineChosen = false; selectedLine.value = "0"; rangeStart.value = 0; posterFailed.value = false; history.value = null; summaryExpanded.value = false; });
watch(data, () => { if (import.meta.client) readHistory(); });
watch(selectedLine, revealResume);
const resume = computed(() => history.value && episodes.value.some((item) => item.key === history.value?.episodeKey) ? history.value : null);
const startKey = computed(() => resume.value?.episodeKey || episodes.value[0]?.key);
const resumeTime = computed(() => resume.value ? `${Math.floor(resume.value.position / 60)}:${String(resume.value.position % 60).padStart(2, "0")}` : "");
useSeoMeta({ title: () => data.value?.title.name || "影片详情", description: () => data.value?.title.summary || "查看影片信息和播放线路" });
</script>

<template>
  <div class="ytv-detail min-h-[calc(100dvh-4rem)] text-white">
    <div class="mx-auto max-w-[90rem] space-y-7 px-4 py-6 sm:px-6 lg:px-8">
      <NuxtLink :to="returnTo || '/'" class="inline-flex items-center gap-1.5 py-1 text-xs text-white/60 no-underline hover:text-white"><UIcon name="i-tabler-chevron-left" class="size-4" />{{ returnTo ? '返回搜索结果' : '发现影片' }}</NuxtLink>
      <div v-if="pending" class="grid gap-6 sm:grid-cols-[12rem_1fr]"><USkeleton class="aspect-[2/3] max-w-48 rounded-xl" /><div class="space-y-4"><USkeleton class="h-9 w-2/3" /><USkeleton class="h-24 w-full" /></div></div>
      <UAlert v-else-if="error" color="error" title="影片详情暂时不可用" description="请重试，或返回搜索选择其他来源。"><template #actions><UButton label="重试" @click="() => refresh()" /></template></UAlert>
      <template v-else-if="data">
        <section class="grid grid-cols-[6rem_minmax(0,1fr)] items-start gap-4 sm:grid-cols-[10rem_minmax(0,1fr)] sm:gap-6 md:grid-cols-[12rem_minmax(0,1fr)] lg:grid-cols-[14rem_minmax(0,1fr)] lg:gap-x-9">
          <div class="relative aspect-[2/3] w-full overflow-hidden rounded-xl border border-white/10 bg-[#121b26] shadow-2xl sm:row-span-2">
            <img v-if="data.title.imageUrl && !posterFailed" :src="data.title.imageUrl" :alt="`${data.title.name} 海报`" class="size-full object-cover" referrerpolicy="no-referrer" @error="posterFailed = true">
            <div v-else class="grid size-full place-items-center"><UIcon name="i-tabler-movie" class="size-9 text-white/20" /></div>
          </div>
          <div class="min-w-0 space-y-4 py-1">
            <div class="flex flex-wrap items-center gap-2 text-xs text-white/60"><span class="rounded-md border border-sky-400/20 bg-sky-400/5 px-2 py-1 text-sky-300">{{ movie ? '单部影片' : '剧集' }}</span><span>{{ data.title.sourceName }}</span><span v-if="data.title.remarks" class="text-white/60">· {{ data.title.remarks }}</span></div>
            <div><h1 class="break-words text-2xl font-bold leading-tight tracking-tight lg:text-4xl">{{ data.title.name }}</h1><p class="mt-3 text-sm text-white/60">{{ [data.title.year, data.title.category, data.title.area, data.title.language].filter(Boolean).join(' · ') || '影片资料待补充' }}</p></div>
            <div v-if="Number(data.title.score) > 0" class="inline-flex items-center gap-2 text-amber-300"><UIcon name="i-tabler-star-filled" class="size-4" /><span class="text-xl font-semibold tabular-nums">{{ Number(data.title.score).toFixed(1) }}</span><span class="text-xs text-white/60">来源评分</span></div>
            <div class="flex flex-wrap items-center gap-3 pt-1">
              <UButton v-if="startKey !== undefined" :label="resume ? '继续观看' : movie ? '播放正片' : '开始播放'" icon="i-tabler-player-play-filled" :to="playbackLink(startKey)" size="lg" />
              <UButton label="其他来源" icon="i-tabler-search" :to="{ path: '/search', query: { q: data.title.name } }" color="neutral" variant="outline" />
              <span v-if="resume" class="w-full text-xs text-white/60">{{ resume.episodeTitle || '上次观看' }} · {{ resumeTime }}</span>
            </div>
          </div>
          <div class="col-span-2 min-w-0 sm:col-span-1 sm:col-start-2">
            <h2 class="mb-2 text-xs font-medium text-white/75">剧情简介</h2>
            <p id="title-synopsis" class="max-w-3xl whitespace-pre-line break-words text-sm leading-7 text-white/60" :class="summaryExpanded ? '' : 'line-clamp-3'">{{ data.title.summary || '当前来源暂未提供剧情简介。' }}</p>
            <UButton v-if="data.title.summary?.length > 100" :label="summaryExpanded ? '收起简介' : '展开简介'" :aria-expanded="summaryExpanded" aria-controls="title-synopsis" color="neutral" variant="link" size="sm" class="mt-1 px-0" @click="toggleSummary" />
          </div>
        </section>
        <section class="overflow-hidden rounded-xl border border-white/10 bg-[#101720]" aria-labelledby="playback-options-title">
          <div class="flex flex-wrap items-center justify-between gap-3 border-b border-white/8 px-4 py-4 sm:px-5"><h2 id="playback-options-title" class="text-base font-semibold">{{ movie ? '播放线路' : '选择剧集' }}</h2><USelect v-if="ranges.length > 1" v-model="rangeStart" :items="ranges" aria-label="选集范围" class="w-36" /><span v-else class="text-xs text-white/60">{{ lines.length }} 条线路</span></div>
          <div class="space-y-4 p-4 sm:p-5">
            <div v-if="lines.length > 1" class="flex flex-wrap gap-2" role="group" aria-label="播放线路"><UButton v-for="item in lines" :key="item.key" :label="item.name" :color="line?.key === item.key ? 'primary' : 'neutral'" :variant="line?.key === item.key ? 'soft' : 'ghost'" :aria-pressed="line?.key === item.key" @click="selectLine(item.key)" /></div>
            <div v-if="episodes.length" class="grid gap-2" :class="movie ? 'sm:grid-cols-3 lg:grid-cols-4' : 'grid-cols-3 sm:grid-cols-5 lg:grid-cols-8'">
              <NuxtLink v-for="episode in episodes.slice(rangeStart, rangeStart + 50)" :key="episode.key" :to="playbackLink(episode.key)" :aria-label="episode.key === resume?.episodeKey ? `${episode.title}，上次观看` : episode.title" class="flex min-h-11 items-center justify-center gap-2 rounded-lg border px-3 py-3 text-center text-sm no-underline transition hover:border-sky-400/40 hover:bg-sky-400/5 hover:text-sky-200" :class="episode.key === resume?.episodeKey ? 'border-sky-400/40 bg-sky-400/10 text-sky-200' : 'border-white/10 text-white/65'"><UIcon v-if="movie || episode.key === resume?.episodeKey" name="i-tabler-player-play" class="size-4 shrink-0" />{{ movie ? episode.title.replace(/^第\s*1\s*集$/u, '正片') || '正片' : episode.title }}</NuxtLink>
            </div>
            <p v-else class="text-sm text-white/60">当前没有可播放的条目，可以尝试其他来源。</p>
          </div>
        </section>
      </template>
    </div>
  </div>
</template>
<style scoped>
.ytv-detail { background: radial-gradient(ellipse at 15% 0%, #14283b 0%, #070b10 55%); }
</style>
