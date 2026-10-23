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
const filter = ref("");
const today = ref(0);
onMounted(() => { today.value = new Date().setHours(0, 0, 0, 0); });
function elapsed(position: number) {
  const seconds = Math.max(0, Math.floor(position));
  return `${Math.floor(seconds / 60)}:${String(seconds % 60).padStart(2, "0")}`;
}
const groups = computed(() => {
  const result: { label: string; records: LocalPlaybackProgress[] }[] = [];
  const yesterday = new Date(today.value); yesterday.setDate(yesterday.getDate() - 1);
  for (const item of items.value) {
    if (filter.value.trim() && !(item.titleName || item.externalId).toLocaleLowerCase().includes(filter.value.trim().toLocaleLowerCase())) continue;
    const label = item.updatedAt >= today.value ? "今天" : item.updatedAt >= yesterday.getTime() ? "昨天" : "更早";
    let group = result.find(value => value.label === label);
    if (!group) { group = { label, records: [] }; result.push(group); }
    group.records.push(item);
  }
  return result;
});
</script>

<template>
  <div class="ytv-container min-h-[70dvh] space-y-6 pb-10">
    <div class="ytv-history-head">
      <div><h1>播放历史</h1><p class="mt-1 text-sm sm:text-base text-white/90">继续观看你曾经看过的内容</p></div>
      <div class="ytv-history-tools">
        <UInput v-if="items.length" v-model="filter" icon="i-tabler-search" placeholder="在当前记录中搜索" aria-label="筛选已加载的记录" class="w-64" :ui="{ base: 'min-h-11 rounded-2xl bg-white/5 text-base ring-white/20' }" />
        <div v-if="!loggedIn" class="ytv-history-sync"><UIcon name="i-tabler-info-circle" class="size-4" /><span>记录保存在当前浏览器</span><UButton label="登录并同步" color="primary" variant="link" size="sm" @click="startLogin" /></div>
      </div>
    </div>
    <UAlert v-if="error" color="error" variant="subtle" title="账号历史暂时不可用" description="当前浏览器的本地进度仍然保留；请稍后重试。"><template #actions><UButton label="重试" color="error" variant="soft" @click="() => refresh()" /></template></UAlert>
    <div v-if="groups.length" data-history-records class="space-y-8">
      <section v-for="group in groups" :key="group.label" :aria-label="group.label" class="space-y-4">
        <h2 class="ytv-history-group-title">{{ group.label }}</h2>
        <ul class="space-y-3">
          <li v-for="item in group.records" :key="posterKey(item)" class="ytv-history-row">
            <NuxtLink :to="watchPath(item.sourceKey, item.externalId, item.episodeKey)" :aria-label="`观看 ${item.titleName || item.externalId}`" class="ytv-history-poster grid place-items-center">
              <img v-if="posterFor(item)" :src="posterFor(item)" :alt="`${item.titleName || item.externalId} 海报`" loading="lazy" decoding="async" referrerpolicy="no-referrer" class="size-full object-cover" @error="failPoster(item)">
              <UIcon v-else name="i-tabler-photo" class="size-7 text-white/30" />
            </NuxtLink>
            <div class="min-w-0 flex-1 space-y-2">
              <h3 class="ytv-history-title">{{ item.titleName || item.externalId }}</h3>
              <p class="ytv-history-meta"><UIcon name="i-tabler-player-play" class="size-5" /><span>{{ sourceName(item) }}</span><span class="text-white/35">|</span><span>{{ item.episodeTitle || '上次观看' }}</span><span class="text-white/35">|</span><span>{{ item.completed ? '已看完' : item.duration > 0 ? `${elapsed(item.position)} / ${elapsed(item.duration)}` : `看到 ${elapsed(item.position)}` }}</span></p>
              <div v-if="item.duration > 0" class="ytv-history-progress"><span class="ytv-progress" :aria-label="`已观看 ${percent(item.position, item.duration)}%`"><span :style="{ width: `${percent(item.position, item.duration)}%` }" /></span><span class="text-sm tabular-nums">{{ percent(item.position, item.duration) }}%</span></div>
              <time :datetime="new Date(item.updatedAt).toISOString()" class="sr-only">{{ dateFormatter.format(new Date(item.updatedAt)) }}</time>
            </div>
            <UButton :to="watchPath(item.sourceKey, item.externalId, item.episodeKey)" :label="item.completed ? '再次观看' : '继续观看'" icon="i-tabler-player-play-filled" color="neutral" variant="ghost" class="ytv-primary ytv-history-action" />
          </li>
        </ul>
      </section>
    </div>
    <UEmpty v-else-if="items.length" icon="i-tabler-search-off" title="没有匹配的记录" description="这里只筛选当前加载的观看记录。"><template #actions><UButton label="清除筛选" color="neutral" variant="outline" @click="filter = ''" /></template></UEmpty>
    <UEmpty v-else icon="i-tabler-history-off" title="还没有播放记录" description="看过的影片会出现在这里，方便下次接着看。"><template #actions><UButton to="/" label="去找一部电影" icon="i-tabler-search" /></template></UEmpty>
  </div>
</template>
