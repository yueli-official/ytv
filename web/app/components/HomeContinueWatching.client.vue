<script setup lang="ts">
import type { PlaybackProgress } from "~/types/ytv";
import { loadLocalProgress, migrateLegacyProgress, type LocalPlaybackProgress } from "~/utils/local-progress";

interface ContinueItem {
  sourceKey: string;
  externalId: string;
  episodeKey: string;
  title: string;
  episode: string;
  position: number;
  duration: number;
  updatedAt: number;
}

const items = ref<ContinueItem[]>([]);
const api = useYtvApi();
const { loggedIn } = useAuth();
const failedPosters = ref(new Set<string>());
const locators = computed(() => items.value.map(({ sourceKey, externalId }) => ({ sourceKey, externalId })));
const { metaFor } = useProgressPosters(locators, 6);
const { rail, canPrevious, canNext, previous, next } = useHorizontalRail();

function remaining(position: number, duration: number) {
  const seconds = Math.max(0, duration - position);
  return seconds >= 60 ? `剩余约 ${Math.ceil(seconds / 60)} 分钟` : `剩余 ${seconds} 秒`;
}

function dedupe(values: ContinueItem[]) {
  const seen = new Set<string>();
  return values.sort((left, right) => right.updatedAt - left.updatedAt).filter((item) => {
    const key = `${item.sourceKey}\u0000${item.externalId}`;
    if (seen.has(key)) return false;
    seen.add(key);
    return true;
  }).slice(0, 6);
}

function fromLocal(values: LocalPlaybackProgress[]) {
  return dedupe(values.filter((item) => !item.completed).map((item) => ({
    sourceKey: item.sourceKey, externalId: item.externalId, episodeKey: item.episodeKey,
    title: item.titleName || item.externalId,
    episode: item.episodeTitle || `第 ${Number(item.episodeKey) + 1} 集`,
    position: item.position, duration: item.duration, updatedAt: item.updatedAt
  })));
}

function fromAccount(values: PlaybackProgress[]) {
  return dedupe(values.filter((item) => !item.completed).map((item) => ({
    sourceKey: item.sourceKey, externalId: item.externalId, episodeKey: item.episodeKey,
    title: item.titleName || item.externalId,
    episode: item.episodeTitle || `第 ${Number(item.episodeKey) + 1} 集`,
    position: item.positionSeconds, duration: item.durationSeconds, updatedAt: new Date(item.updatedAt).getTime()
  })));
}

function posterKey(item: ContinueItem) {
  return `${item.sourceKey}\u0000${item.externalId}`;
}

function posterFor(item: ContinueItem) {
  const image = metaFor(item.sourceKey, item.externalId)?.imageUrl;
  return image && !failedPosters.value.has(posterKey(item)) ? image : "";
}

function failPoster(item: ContinueItem) {
  failedPosters.value = new Set([...failedPosters.value, posterKey(item)]);
}

function percent(item: ContinueItem) {
  return item.duration > 0 ? Math.min(100, Math.round((item.position / item.duration) * 100)) : 0;
}

onMounted(async () => {
  migrateLegacyProgress(localStorage);
  const local = loadLocalProgress(localStorage);
  if (!loggedIn.value) {
    items.value = fromLocal(local);
    return;
  }
  const account = await api.listProgress().catch(() => null);
  items.value = account ? fromAccount(account.items) : fromLocal(local);
});
</script>

<template>
  <section v-if="items.length" data-continue-watching class="space-y-3" aria-labelledby="continue-watching-title">
    <div class="flex items-center justify-between gap-3">
      <h2 id="continue-watching-title" class="text-lg font-bold sm:text-xl">继续观看</h2>
      <UButton to="/history" label="全部记录" trailing-icon="i-tabler-chevron-right" color="neutral" variant="link" size="sm" />
    </div>
    <div class="relative">
      <UButton v-if="canPrevious" data-carousel-previous icon="i-tabler-chevron-left" aria-label="上一组继续观看" color="neutral" variant="ghost" class="absolute -left-5 top-1/2 z-10 hidden min-h-11 min-w-11 -translate-y-1/2 justify-center rounded-full border border-white/15 bg-[#070b10]/95 text-white shadow-lg hover:bg-[#111a24] sm:inline-flex" @click="previous" />
      <ul ref="rail" class="continue-rail flex snap-x snap-mandatory gap-3 overflow-x-auto overscroll-x-contain pb-1" aria-label="继续观看列表">
        <li v-for="item in items" :key="posterKey(item)" class="w-[15rem] shrink-0 snap-start sm:w-[16rem]">
          <NuxtLink :to="`/watch/${item.sourceKey}/${item.externalId}/${item.episodeKey}`" class="group relative flex h-28 overflow-hidden rounded-xl border border-white/10 bg-[#0f1720] text-white no-underline transition hover:border-sky-400/45 hover:bg-[#121d28]">
            <div class="grid w-[4.7rem] shrink-0 place-items-center overflow-hidden bg-[#111923] text-white/25">
              <img v-if="posterFor(item)" :src="posterFor(item)" :alt="`${item.title} 海报`" loading="lazy" decoding="async" referrerpolicy="no-referrer" class="size-full object-cover" @error="failPoster(item)">
              <UIcon v-else name="i-tabler-photo" class="size-6" aria-hidden="true" />
            </div>
            <div class="min-w-0 flex-1 p-3 pb-4">
              <h3 class="line-clamp-2 text-sm font-semibold leading-5">{{ item.title }}</h3>
              <p class="mt-1 truncate text-xs text-sky-300">{{ item.episode }}</p>
              <p class="mt-2 truncate text-[11px] text-white/45">{{ remaining(item.position, item.duration) }}</p>
            </div>
            <span class="absolute inset-x-0 bottom-0 h-1 bg-white/10"><span class="block h-full bg-sky-500" :style="{ width: `${percent(item)}%` }" /></span>
          </NuxtLink>
        </li>
      </ul>
      <UButton v-if="canNext" data-carousel-next icon="i-tabler-chevron-right" aria-label="下一组继续观看" color="neutral" variant="ghost" class="absolute -right-5 top-1/2 z-10 hidden min-h-11 min-w-11 -translate-y-1/2 justify-center rounded-full border border-white/15 bg-[#070b10]/95 text-white shadow-lg hover:bg-[#111a24] sm:inline-flex" @click="next" />
    </div>
  </section>
</template>

<style scoped>
.continue-rail {
  scrollbar-width: none;
}

.continue-rail::-webkit-scrollbar {
  display: none;
}
</style>
