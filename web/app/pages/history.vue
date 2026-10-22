<script setup lang="ts">
definePageMeta({ colorMode: "dark" });
import type { PlaybackProgress } from "~/types/ytv";
import type { LocalPlaybackProgress } from "~/utils/local-progress";
import { browserProgress, mergeViewerProgress } from "~/utils/viewer-progress";
import { watchPath } from "~/utils/title-lines";

useSeoMeta({ title: "播放历史" });

const api = useYtvApi();
const { loggedIn, login } = useAuth();
const localItems = ref<LocalPlaybackProgress[]>([]);
const failedPosters = ref(new Set<string>());
const { data: remote, error, refresh } = await useAsyncData(
  "ytv-playback-history",
  () => loggedIn.value ? api.listProgress() : Promise.resolve({ items: [] as PlaybackProgress[] }),
  { watch: [loggedIn] }
);

onMounted(() => {
  localItems.value = browserProgress();
});

const rawItems = computed(() => mergeViewerProgress(localItems.value, loggedIn.value ? remote.value?.items : []));

const items = computed(() => {
  const seen = new Set<string>();
  return [...rawItems.value]
    .sort((left, right) => right.updatedAt - left.updatedAt)
    .filter((item) => {
      const key = `${item.sourceKey}\u0000${item.externalId}`;
      if (seen.has(key)) return false;
      seen.add(key);
      return true;
    })
    .slice(0, 24);
});

const locators = computed(() => items.value.map(({ sourceKey, externalId }) => ({ sourceKey, externalId })));
const { metaFor } = useProgressPosters(locators, 24);

const dateFormatter = new Intl.DateTimeFormat("zh-CN", { month: "2-digit", day: "2-digit", hour: "2-digit", minute: "2-digit" });
const percent = (position: number, duration: number) => duration > 0 ? Math.min(100, Math.round((position / duration) * 100)) : 0;
function startLogin() { void login("/history"); }
function posterKey(item: { sourceKey: string; externalId: string }) { return `${item.sourceKey}\u0000${item.externalId}`; }
function posterFor(item: { sourceKey: string; externalId: string }) {
  const image = metaFor(item.sourceKey, item.externalId)?.imageUrl;
  return image && !failedPosters.value.has(posterKey(item)) ? image : "";
}
function sourceName(item: { sourceKey: string; externalId: string }) { return metaFor(item.sourceKey, item.externalId)?.sourceName || item.sourceKey; }
function failPoster(item: { sourceKey: string; externalId: string }) { failedPosters.value = new Set([...failedPosters.value, posterKey(item)]); }
</script>

<template>
  <div class="ytv-container min-h-[70dvh] space-y-6 py-8">
    <div class="flex flex-wrap items-end justify-between gap-4">
      <div><h1 class="text-2xl font-semibold tracking-tight text-highlighted">播放历史</h1><p class="mt-2 max-w-2xl text-sm leading-6 text-muted">{{ loggedIn ? '找到看过的影片，接着上次的进度观看。' : '记录保存在当前浏览器，登录可同步最近 30 天的观看进度。' }}</p></div>
      <UButton v-if="!loggedIn" label="登录并同步" icon="i-tabler-login-2" color="neutral" variant="soft" @click="startLogin" />
    </div>

    <UAlert v-if="error" color="error" variant="subtle" title="账号历史暂时不可用" description="当前浏览器的本地进度仍然保留；请稍后重试。">
      <template #actions><UButton label="重试" color="error" variant="soft" @click="() => refresh()" /></template>
    </UAlert>

    <ul v-if="items.length" data-history-poster-grid class="ytv-poster-grid">
      <li v-for="item in items" :key="posterKey(item)" class="min-w-0">
        <NuxtLink :to="watchPath(item.sourceKey, item.externalId, item.episodeKey)" class="group block text-white no-underline">
          <div class="relative aspect-[2/3] overflow-hidden rounded-xl border border-white/10 bg-[#0f1720] shadow-sm transition group-hover:-translate-y-0.5 group-hover:border-sky-400/45">
            <img v-if="posterFor(item)" :src="posterFor(item)" :alt="`${item.titleName || item.externalId} 海报`" loading="lazy" decoding="async" referrerpolicy="no-referrer" class="size-full object-cover" @error="failPoster(item)">
            <span v-else class="grid size-full place-items-center text-white/25"><UIcon name="i-tabler-photo" class="size-9" aria-hidden="true" /></span>
            <span class="absolute right-2 top-2 rounded-md bg-black/75 px-2 py-1 text-[11px] font-semibold text-sky-200">{{ item.completed ? '已看完' : `${percent(item.position, item.duration)}%` }}</span>
            <span class="absolute inset-x-0 bottom-0 bg-gradient-to-t from-black/95 via-black/60 to-transparent px-3 pb-3 pt-10 text-xs font-medium">{{ item.episodeTitle || `第 ${Number(item.episodeKey) + 1} 集` }}</span>
            <span class="absolute inset-x-0 bottom-0 h-1 bg-white/15"><span class="block h-full bg-sky-500" :style="{ width: `${percent(item.position, item.duration)}%` }" /></span>
          </div>
          <h2 class="mt-2 line-clamp-2 text-sm font-semibold leading-5 group-hover:text-sky-300">{{ item.titleName || item.externalId }}</h2>
          <p class="mt-1 truncate text-xs text-white/45">{{ sourceName(item) }} · {{ dateFormatter.format(new Date(item.updatedAt)) }}</p>
        </NuxtLink>
      </li>
    </ul>
    <UEmpty v-else icon="i-tabler-history-off" title="还没有播放记录" description="看过的影片会出现在这里，方便下次接着看。"><template #actions><UButton to="/" label="去找一部电影" icon="i-tabler-search" /></template></UEmpty>
  </div>
</template>
