<script setup lang="ts">
import Hls, { ErrorTypes, Events, type ErrorData, type Level } from "hls.js";
import Plyr from "plyr";
import "plyr/dist/plyr.css";
import type { PlaybackManifest } from "~/types/ytv";
import { completedPlayback, restorablePosition, type PlaybackProgressRecord } from "~/utils/progress";
import { progressStorageKey, type LocalPlaybackProgress } from "~/utils/local-progress";

const props = defineProps<{
  manifest: PlaybackManifest;
  sourceKey: string;
  externalId: string;
  title: string;
  episodeLabel: string;
  poster?: string;
}>();
const emit = defineEmits<{ ended: []; started: []; progress: [record: LocalPlaybackProgress] }>();

type PlayerState = "idle" | "loading" | "ready" | "playing" | "paused" | "seeking" | "rebuffering" | "ended" | "error";

const video = ref<HTMLVideoElement | null>(null);
const state = ref<PlayerState>("idle");
const message = ref("");
const selectedLevel = ref(-1);
const levels = ref<Array<{ label: string; value: number }>>([{ label: "自动", value: -1 }]);
let hls: Hls | null = null;
let controls: Plyr | null = null;
let generation = 0;
let progressTimer: number | null = null;
let lastMediaRecoveryAt = 0;
let started = false;

const storageKey = computed(() => progressStorageKey(props.sourceKey, props.externalId, props.manifest.episodeKey));

function readProgress(): PlaybackProgressRecord | null {
  try {
    const raw = localStorage.getItem(storageKey.value);
    return raw ? JSON.parse(raw) as PlaybackProgressRecord : null;
  } catch {
    return null;
  }
}

function flushProgress() {
  const element = video.value;
  if (!element || !Number.isFinite(element.duration) || element.duration <= 0) return;
  const record: LocalPlaybackProgress = {
    sourceKey: props.sourceKey,
    externalId: props.externalId,
    episodeKey: props.manifest.episodeKey,
    titleName: props.title,
    episodeTitle: props.episodeLabel,
    position: Math.floor(element.currentTime),
    duration: Math.floor(element.duration),
    updatedAt: Date.now(),
    completed: completedPlayback(element.currentTime, element.duration)
  };
  localStorage.setItem(storageKey.value, JSON.stringify(record));
  emit("progress", record);
}

function restoreProgress() {
  const element = video.value;
  if (!element) return;
  const target = restorablePosition(readProgress(), element.duration, element.seekable);
  if (target > 0) element.currentTime = target;
}

function updateLevels(values: Level[]) {
  const seen = new Set<number>();
  levels.value = [{ label: "自动", value: -1 }];
  values.forEach((level, index) => {
    if (!level.height || seen.has(level.height)) return;
    seen.add(level.height);
    levels.value.push({ label: `${level.height}p`, value: index });
  });
  selectedLevel.value = -1;
}

function handleFatalError(data: ErrorData, currentGeneration: number) {
  if (currentGeneration !== generation || !data.fatal) return;
  if (data.type === ErrorTypes.MEDIA_ERROR && Date.now() - lastMediaRecoveryAt >= 5000) {
    lastMediaRecoveryAt = Date.now();
    message.value = "正在恢复媒体解码";
    hls?.recoverMediaError();
    return;
  }
  state.value = "error";
  message.value = data.type === ErrorTypes.NETWORK_ERROR
    ? "当前来源网络请求失败，请重试或切换来源。"
    : "当前媒体无法继续播放，请切换来源。";
}

function bindMediaEvents(element: HTMLVideoElement, currentGeneration: number) {
  element.onloadedmetadata = () => {
    if (currentGeneration !== generation) return;
    restoreProgress();
    state.value = "ready";
  };
  element.onplaying = () => {
    if (currentGeneration !== generation) return;
    state.value = "playing";
    message.value = "";
    if (!started) {
      started = true;
      emit("started");
    }
  };
  element.onpause = () => {
    if (currentGeneration !== generation || element.ended) return;
    state.value = "paused";
    flushProgress();
  };
  element.onwaiting = () => {
    if (currentGeneration !== generation || element.seeking) return;
    state.value = "rebuffering";
    message.value = "正在缓冲";
  };
  element.onseeking = () => { if (currentGeneration === generation) state.value = "seeking"; };
  element.onseeked = () => { if (currentGeneration === generation) flushProgress(); };
  element.onended = () => {
    if (currentGeneration !== generation) return;
    state.value = "ended";
    flushProgress();
    emit("ended");
  };
}

function destroyEngine() {
  flushProgress();
  if (progressTimer !== null) window.clearInterval(progressTimer);
  progressTimer = null;
  hls?.destroy();
  hls = null;
  controls?.destroy();
  controls = null;
  const element = video.value;
  if (element) {
    element.onloadedmetadata = null;
    element.onplaying = null;
    element.onpause = null;
    element.onwaiting = null;
    element.onseeking = null;
    element.onseeked = null;
    element.onended = null;
    element.removeAttribute("src");
    element.load();
  }
}

function initialize() {
  const element = video.value;
  if (!element) return;
  destroyEngine();
  generation += 1;
  started = false;
  const currentGeneration = generation;
  state.value = "loading";
  message.value = "正在读取播放清单";
  element.poster = props.poster || "";
  bindMediaEvents(element, currentGeneration);
  controls = new Plyr(element, {
    autoplay: false,
    keyboard: { focused: true, global: false },
    controls: ["play", "progress", "current-time", "duration", "mute", "volume", "settings", "pip", "fullscreen"],
    settings: ["speed"]
  });

  if (Hls.isSupported() && props.manifest.engine === "hls") {
    hls = new Hls({ enableWorker: true, workerPath: "/vendor/hls.worker.js" });
    hls.attachMedia(element);
    hls.on(Events.MEDIA_ATTACHED, () => {
      if (currentGeneration === generation) hls?.loadSource(props.manifest.url);
    });
    hls.on(Events.MANIFEST_PARSED, (_event, data) => {
      if (currentGeneration !== generation) return;
      updateLevels(data.levels);
      state.value = "ready";
      message.value = "";
    });
    hls.on(Events.ERROR, (_event, data) => handleFatalError(data, currentGeneration));
  } else if (element.canPlayType("application/vnd.apple.mpegurl") && props.manifest.engine === "hls") {
    element.src = props.manifest.url;
  } else {
    element.src = props.manifest.url;
  }
  progressTimer = window.setInterval(flushProgress, 10_000);
}

function retry() {
  initialize();
}

async function playPrimary() {
  const element = video.value;
  if (!element) return;
  if (state.value === "ended") element.currentTime = 0;
  try {
    if (controls) await controls.play();
    else await element.play();
  } catch {
    state.value = "ready";
    message.value = "浏览器没有开始播放，请再次点击播放。";
  }
}

watch(selectedLevel, (value) => {
  if (hls) hls.loadLevel = value;
});
watch(() => props.manifest.url, initialize);

function onVisibilityChange() {
  if (document.visibilityState === "hidden") flushProgress();
}

onMounted(async () => {
  await nextTick();
  document.addEventListener("visibilitychange", onVisibilityChange);
  initialize();
});
onBeforeUnmount(() => {
  document.removeEventListener("visibilitychange", onVisibilityChange);
  destroyEngine();
});
</script>

<template>
  <div class="absolute inset-0 bg-black" :data-player-state="state">
    <video ref="video" class="size-full" playsinline preload="metadata" :aria-label="`${title} ${episodeLabel}`" />
    <div class="pointer-events-none absolute inset-x-3 top-3 z-10 flex items-start justify-between gap-3">
      <span class="rounded-lg bg-black/75 px-3 py-2 text-xs text-white/80">{{ episodeLabel }}</span>
      <USelect v-if="levels.length > 1" v-model="selectedLevel" :items="levels" value-key="value" aria-label="清晰度" class="pointer-events-auto w-28" />
    </div>
    <div v-if="state === 'loading' || state === 'rebuffering' || state === 'seeking'" class="pointer-events-none absolute inset-0 grid place-items-center bg-black/25">
      <div class="grid justify-items-center gap-2 rounded-xl bg-black/75 px-4 py-3"><UIcon name="i-tabler-loader-2" class="size-6 animate-spin text-primary motion-reduce:animate-none" /><span class="text-sm text-white">{{ message || '正在加载' }}</span></div>
    </div>
    <div v-else-if="state === 'error'" class="absolute inset-0 grid place-items-center bg-black/80 p-4">
      <div class="max-w-md space-y-4 text-center"><UIcon name="i-tabler-alert-triangle" class="mx-auto size-10 text-error" /><div><h2 class="text-xl font-semibold">播放失败</h2><p class="mt-1 text-sm text-white/65">{{ message }}</p></div><UButton label="重试当前来源" icon="i-tabler-refresh" @click="retry" /></div>
    </div>
    <button
      v-else-if="state === 'ready' || state === 'paused' || state === 'ended'"
      type="button"
      data-primary-play
      class="absolute left-1/2 top-1/2 z-20 grid size-16 -translate-x-1/2 -translate-y-1/2 place-items-center rounded-full border border-white/25 bg-black/75 text-white shadow-2xl transition hover:scale-105 hover:bg-sky-600 focus-visible:outline-2 focus-visible:outline-offset-4 focus-visible:outline-sky-300 sm:size-[4.5rem]"
      :aria-label="state === 'ended' ? `重播 ${title} ${episodeLabel}` : `播放 ${title} ${episodeLabel}`"
      @click="playPrimary"
    >
      <UIcon :name="state === 'ended' ? 'i-tabler-replay' : 'i-tabler-player-play-filled'" class="size-7 translate-x-0.5 sm:size-8" />
    </button>
  </div>
</template>

<style scoped>
:global(.plyr) {
  --plyr-color-main: #38bdf8;
}
</style>
