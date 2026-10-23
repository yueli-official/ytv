import type { PlaybackProgress } from "~/types/ytv";
import type { LocalPlaybackProgress } from "~/utils/local-progress";
import { browserProgress, mergeViewerProgress } from "~/utils/viewer-progress";
import { watchPath } from "~/utils/title-lines";
export function useHistoryView() {
useSeoMeta({ title: "播放历史" });
const api = useYtvApi();
const { loggedIn, login } = useAuth();
const localItems = ref<LocalPlaybackProgress[]>([]);
const failedPosters = ref(new Set<string>());
const { data: remote, error, refresh } = useAsyncData(
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
return reactive({ browserProgress, mergeViewerProgress, watchPath, api, loggedIn, login, localItems, failedPosters, remote, error, refresh, rawItems, items, locators, metaFor, dateFormatter, percent, startLogin, posterKey, posterFor, sourceName, failPoster, filter, today, elapsed, groups });
}
