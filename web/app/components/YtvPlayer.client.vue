<script setup lang="ts">
import "~/utils/video-player-element";
import type { YtvVideoPlayerElement } from "~/utils/video-player-element";
import "@videojs/html/video/neutral-skin";
import "@videojs/html/media/hlsjs-video";
import "@videojs/html/i18n";
import "@videojs/html/i18n/locales/zh/register";
import { createPlaybackSession, type PlaybackMedia, type PlaybackSelection, type PlaybackSnapshot } from "~/utils/playback-session";
import type { LocalPlaybackProgress } from "~/utils/local-progress";
import { bindPlayerPreferences } from "~/utils/player-preferences";
import { bindHoldSpeed } from "~/utils/hold-speed";
import { ytvFailureMessage } from "~/utils/ytv-failure";
import { mergeViewerProgress } from "~/utils/viewer-progress";

const props = defineProps<{ sourceKey: string; externalId: string; episodeKey: string; title: string; episodeLabel: string; poster?: string }>();
const emit = defineEmits<{ ended: []; started: [selection: PlaybackSelection]; progress: [record: LocalPlaybackProgress] }>();
const api = useYtvApi();
const { loggedIn } = useAuth();
const skin = ref<HTMLElement | null>(null);
const player = ref<YtvVideoPlayerElement | null>(null);
const fullscreenMessage = ref("");
const holdingSpeed = ref(false);
const fullscreenActive = ref(false);
const snapshot = ref<PlaybackSnapshot>({ state: "idle", message: "" });
let session: ReturnType<typeof createPlaybackSession> | null = null;
const selection = () => ({ sourceKey: props.sourceKey, externalId: props.externalId, episodeKey: props.episodeKey, titleName: props.title, episodeTitle: props.episodeLabel });
function persist() { session?.save(); }
function retryPlayback() { void session?.retry(); }
function play() { void session?.play(); }
function updateFullscreen() { fullscreenActive.value = Boolean(document.fullscreenElement); }
async function enterFullscreen() {
  fullscreenMessage.value = "";
  const store = player.value?.store;
  if (!store || store.state.fullscreenAvailability !== "available") {
    fullscreenMessage.value = "当前浏览器不支持全屏，可横放手机观看。";
    return;
  }
  try { await store.requestFullscreen(); }
  catch { fullscreenMessage.value = "未能进入全屏，请再点一次全屏观看。"; }
}
defineExpose({ enterFullscreen });

onMounted(() => {
  session = createPlaybackSession({
    describeError: cause => ytvFailureMessage(cause, "没有取得可播放的地址，请重试或选择其他线路。"),
    resolve: (value) => api.playback(value.sourceKey, value.externalId, value.episodeKey),
    loadProgress: async (value) => {
      if (!loggedIn.value) return undefined;
      return mergeViewerProgress([], (await api.listProgress()).items).find(item => item.sourceKey === value.sourceKey && item.externalId === value.externalId && item.episodeKey === value.episodeKey);
    },
    mount: (manifest) => {
      const element = document.createElement(manifest.engine === "hls" ? "hlsjs-video" : "video") as HTMLElement & PlaybackMedia & { src: string; playbackRate: number; load(): void };
      element.setAttribute("playsinline", "");
      element.setAttribute("preload", "metadata");
      element.setAttribute("aria-label", `${props.title} ${props.episodeLabel}`);
      element.style.cssText = "display:block;width:100%;height:100%;object-fit:contain";
      skin.value!.append(element);
      const hold = bindHoldSpeed(skin.value!, element, (active) => { holdingSpeed.value = active; });
      let releasePreferences = () => {};
      try { releasePreferences = bindPlayerPreferences(element as unknown as HTMLMediaElement, localStorage, hold.preferredRate); } catch { /* Playback works without browser storage. */ }
      queueMicrotask(() => { if (element.isConnected) element.src = manifest.url; });
      return { media: element, destroy: () => { hold.dispose(); releasePreferences(); element.pause(); element.removeAttribute("src"); element.remove(); if (element.tagName === "VIDEO") element.load(); } };
    },
    storage: () => localStorage,
    onState: (value) => { snapshot.value = value; },
    onStarted: (value) => emit("started", value),
    onProgress: (value) => emit("progress", value),
    onEnded: () => emit("ended")
  });
  void session.open(selection());
  document.addEventListener("visibilitychange", persist);
  window.addEventListener("pagehide", persist);
  document.addEventListener("fullscreenchange", updateFullscreen);
});
watch(() => [props.sourceKey, props.externalId, props.episodeKey], () => { void session?.open(selection()); });
onBeforeUnmount(() => {
  document.removeEventListener("visibilitychange", persist);
  window.removeEventListener("pagehide", persist);
  document.removeEventListener("fullscreenchange", updateFullscreen);
  session?.dispose();
});
</script>

<template>
  <div class="absolute inset-0 bg-black" :data-player-state="snapshot.state">
    <media-i18n lang="zh" class="block size-full">
      <component :is="'ytv-video-player'" ref="player" orientation-lock-type="landscape" :poster="poster || undefined" class="block size-full">
        <ytv-video-skin ref="skin" class="ytv-player-skin">
          <div class="ytv-gesture-feedback" :class="{ 'is-fullscreen': fullscreenActive }" aria-live="polite" aria-atomic="true"><span v-if="holdingSpeed" data-hold-speed>2 倍速播放 · 松开恢复</span></div>
        </ytv-video-skin>
      </component>
    </media-i18n>
    <p v-if="fullscreenMessage" role="status" class="absolute inset-x-3 top-3 z-30 rounded-lg bg-black/80 px-3 py-2 text-center text-xs text-white">{{ fullscreenMessage }}</p>
    <button v-if="['ready', 'paused', 'ended'].includes(snapshot.state)" type="button" data-primary-play :aria-label="snapshot.state === 'ended' ? '重新播放' : '播放'" class="absolute left-1/2 top-1/2 grid size-14 -translate-x-1/2 -translate-y-1/2 place-items-center rounded-full border border-white/25 bg-black/55 text-white shadow-xl backdrop-blur-sm transition hover:scale-105 hover:bg-[var(--ui-primary)] focus-visible:outline-2 focus-visible:outline-offset-4 focus-visible:outline-[var(--ui-primary)] sm:size-16" @click="play"><UIcon :name="snapshot.state === 'ended' ? 'i-tabler-refresh' : 'i-tabler-player-play-filled'" class="size-6 sm:size-7" /></button>
    <div v-if="snapshot.state === 'resolving' || snapshot.state === 'loading'" class="pointer-events-none absolute inset-0 grid place-items-center bg-black/60">
      <div class="grid justify-items-center gap-3 text-sm text-white/70" role="status"><UIcon name="i-tabler-loader-2" class="size-6 animate-spin text-[var(--ui-primary)] motion-reduce:animate-none" />{{ snapshot.message }}</div>
    </div>
    <div v-else-if="snapshot.state === 'error'" class="absolute inset-0 z-20 grid place-items-center bg-[#0d141b]/95 p-5">
      <div class="max-w-sm space-y-4 text-center" role="alert">
        <UIcon name="i-tabler-player-pause" class="mx-auto size-8 text-[var(--ui-primary)]" />
        <div><h2 class="text-lg font-semibold">播放暂时中断</h2><p class="mt-2 text-sm leading-6 text-white/75">{{ snapshot.message }}</p></div>
        <div class="flex flex-wrap justify-center gap-2">
          <UButton label="重新连接" icon="i-tabler-refresh" @click="retryPlayback" />
          <slot name="recovery" />
        </div>
      </div>
    </div>
  </div>
</template>

<style scoped>
.ytv-player-skin { display: block; width: 100%; height: 100%; --media-accent-color: #9dd5f5; --media-accent-text-color: #10212e; --media-object-fit: contain; --media-border-radius: 0; }
.ytv-gesture-feedback { position: absolute; inset: 0; z-index: 20; pointer-events: none; display: flex; justify-content: center; align-items: flex-start; padding-top: 1rem; }
.ytv-gesture-feedback span { border-radius: 999px; background: rgb(0 0 0 / 75%); padding: 0.5rem 1rem; color: white; font-size: 0.75rem; }
.ytv-gesture-feedback.is-fullscreen { padding-top: 4rem; }
.ytv-gesture-feedback.is-fullscreen span { font-size: 0.875rem; }
@media (pointer: coarse) {
  .ytv-player-skin { --media-scale-unit: 20px; touch-action: manipulation; }
  .ytv-gesture-feedback { padding-top: max(12px, env(safe-area-inset-top)); }
}
</style>
