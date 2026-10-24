import { isSingleFeature, titleLines, watchPath, searchReturnPath } from "~/utils/title-lines";
import type { LocalPlaybackProgress } from "~/utils/local-progress";
import { browserProgress, mergeViewerProgress } from "~/utils/viewer-progress";
import { ytvFailureMessage } from "~/utils/ytv-failure";
export function useTitleView() {
const route = useRoute();
const api = useYtvApi();
const { loggedIn } = useAuth();
const returnTo = computed(() => searchReturnPath(route.query.from));
const playbackLink = (episode: string) => ({ path: watchPath(sourceKey.value, externalId.value, episode), query: returnTo.value ? { from: returnTo.value } : {} });
const sourceKey = computed(() => String(route.params.source || ""));
const externalId = computed(() => String(route.params.id || ""));
const { data, pending, error, refresh } = useAsyncData(() => `ytv-title-${sourceKey.value}-${externalId.value}`, () => api.title(sourceKey.value, externalId.value));
const selectedLine = ref("");
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
watch([sourceKey, externalId], () => { historyRequest++; lineChosen = false; selectedLine.value = ""; rangeStart.value = 0; posterFailed.value = false; history.value = null; summaryExpanded.value = false; });
watch(data, () => { if (import.meta.client) readHistory(); });
watch(selectedLine, revealResume);
const resume = computed(() => history.value && episodes.value.some((item) => item.key === history.value?.episodeKey) ? history.value : null);
const startKey = computed(() => resume.value?.episodeKey || episodes.value[0]?.key);
const resumeTime = computed(() => resume.value ? `${Math.floor(resume.value.position / 60)}:${String(Math.floor(resume.value.position) % 60).padStart(2, "0")}` : "");
useSeoMeta({ title: () => data.value?.title.name || "影片详情", description: () => data.value?.title.summary || "查看影片信息和播放线路" });
return reactive({ isSingleFeature, titleLines, watchPath, searchReturnPath, browserProgress, mergeViewerProgress, ytvFailureMessage, route, api, loggedIn, returnTo, playbackLink, sourceKey, externalId, data, pending, error, refresh, selectedLine, rangeStart, lines, line, movie, episodes, ranges, history, posterFailed, summaryExpanded, toggleSummary, lineChosen, historyRequest, selectLine, applyHistory, revealResume, readHistory, resume, startKey, resumeTime });
}
