<script setup lang="ts">
defineProps<{compact?:boolean}>();
import type { LocalPlaybackProgress } from "~/utils/local-progress";
import { browserProgress, mergeViewerProgress } from "~/utils/viewer-progress";
import { watchPath } from "~/utils/title-lines";

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

function elapsed(position: number) {
  const seconds = Math.max(0, Math.floor(position));
  return `${Math.floor(seconds / 60)}:${String(seconds % 60).padStart(2, "0")}`;
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
  const local = browserProgress();
  items.value = fromLocal(local);
  if (!loggedIn.value) {
    items.value = fromLocal(local);
    return;
  }
  const account = await api.listProgress().catch(() => null);
  items.value = fromLocal(mergeViewerProgress(local, account?.items));
});
</script>

<template>
  <section v-if="items.length" data-continue-watching class="space-y-3" :class="{'compact-continue':compact}" aria-labelledby="continue-watching-title">
    <div class="ytv-section-heading">
      <h2 id="continue-watching-title" class="text-xl font-bold sm:text-2xl">继续观看</h2>
      <UButton to="/history" label="全部记录" trailing-icon="i-tabler-chevron-right" color="neutral" variant="link" size="sm" />
    </div>
    <div class="relative">
      <UButton v-if="canPrevious" data-carousel-previous icon="i-tabler-chevron-left" aria-label="上一组继续观看" color="neutral" variant="ghost" class="absolute -left-5 top-1/2 z-10 hidden min-h-11 min-w-11 -translate-y-1/2 justify-center rounded-full border border-white/15 bg-[#0d141b]/95 text-white shadow-lg hover:bg-[#1d2a36] sm:inline-flex" @click="previous" />
      <ul ref="rail" class="continue-rail flex snap-x snap-mandatory gap-3 overflow-x-auto overscroll-x-contain pb-1" aria-label="继续观看列表">
        <li v-for="item in compact ? items.slice(0,1) : items" :key="posterKey(item)" class="w-[18rem] shrink-0 snap-start lg:w-[calc((100%-2rem)/3)]">
          <NuxtLink :to="watchPath(item.sourceKey, item.externalId, item.episodeKey)" class="group relative flex h-32 gap-3 sm:h-40 sm:gap-5 lg:h-32 lg:gap-4 overflow-hidden rounded-lg border border-white/10 bg-white/[0.02] p-1.5 text-white no-underline transition hover:border-[var(--ui-primary)]/45 hover:bg-[#243544]">
            <div class="grid w-20 sm:w-28 lg:w-20 rounded shrink-0 place-items-center overflow-hidden bg-[#202c38] text-white/25">
              <img v-if="posterFor(item)" :src="posterFor(item)" :alt="`${item.title} 海报`" loading="lazy" decoding="async" referrerpolicy="no-referrer" class="size-full object-cover" @error="failPoster(item)">
              <UIcon v-else name="i-tabler-photo" class="size-6" aria-hidden="true" />
            </div>
            <div class="min-w-0 flex-1 py-2 pr-2 sm:py-3 sm:pr-3 lg:py-1">
              <h3 class="line-clamp-2 text-base font-semibold leading-6 sm:text-lg lg:text-base lg:leading-5">{{ item.title }}</h3>
              <p v-if="!compact" class="mt-2 truncate text-xs sm:text-sm lg:mt-1 lg:text-xs text-white/70">{{ metaFor(item.sourceKey, item.externalId)?.sourceName || item.sourceKey }}</p>
              <span v-if="item.duration > 0" class="ytv-progress mt-3 lg:mt-2"><span :style="{ width: `${percent(item)}%` }" /></span>
              <p class="mt-2 truncate text-xs sm:text-sm lg:mt-1 lg:text-xs text-white/75">{{ item.episode }} · {{ elapsed(item.position) }}</p>
            </div>

          </NuxtLink>
        </li>
      </ul>
      <UButton v-if="canNext" data-carousel-next icon="i-tabler-chevron-right" aria-label="下一组继续观看" color="neutral" variant="ghost" class="absolute -right-5 top-1/2 z-10 hidden min-h-11 min-w-11 -translate-y-1/2 justify-center rounded-full border border-white/15 bg-[#0d141b]/95 text-white shadow-lg hover:bg-[#1d2a36] sm:inline-flex" @click="next" />
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
.compact-continue .continue-rail>li{width:100%}
.compact-continue .continue-rail>li>a{height:104px;gap:12px;padding:6px}
.compact-continue .continue-rail>li>a>div:first-child{width:62px}
.compact-continue .continue-rail>li>a>div:last-child{padding-top:2px;padding-bottom:2px}
.compact-continue .ytv-progress{margin-top:6px}
.compact-continue h2{font-size:17px}
.compact-continue h3{font-size:14px;line-height:1.5}
.compact-continue [data-carousel-next],.compact-continue [data-carousel-previous]{display:none}
</style>
