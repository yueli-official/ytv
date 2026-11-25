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
import { bindVideoEnhancement, enhancementPreferenceKey, type EnhancementState } from "~/utils/video-enhancement";
import { bindPlayerGestures, type PlayerGestureFeedback } from "~/utils/player-gestures";
import { ytvFailureMessage } from "~/utils/ytv-failure";
import { mergeViewerProgress } from "~/utils/viewer-progress";
import Playlist from '~/components/public/shared/Playlist.vue';
import type { TitleResponse } from '~/types/ytv';
import type { RouteLocationRaw } from 'vue-router';
import { bindPlayerPicture, defaultPlayerSettings, playerSettingsKey, readPlayerSettings } from '~/utils/player-picture';
import { audioBalanceTarget, createAudioBalance, type AudioBalanceMode } from '~/utils/audio-balance';

const props = withDefaults(defineProps<{ sourceKey: string; externalId: string; episodeKey: string; title: string; episodeLabel: string; poster?: string; playlistTitle?: TitleResponse['title']; episodes?: TitleResponse['episodes']; lineOptions?: {label: string; value: string}[]; lineKey?: string; movie?: boolean; playbackLink?: (key: string) => RouteLocationRaw }>(), { episodes: () => [], lineOptions: () => [] });
const emit = defineEmits<{ ended: []; started: [selection: PlaybackSelection]; progress: [record: LocalPlaybackProgress]; line: [key: string]; lights: [off: boolean] }>();
const { mobile } = usePublicLayout();
const settings = ref(defaultPlayerSettings());
const storageError = ref('');
const panelOpen = ref(false), panelMode = ref<'settings' | 'episodes'>('settings');
const panelUI = computed(() => ({
  overlay: 'absolute inset-0 bg-black/50',
  content: panelMode.value === 'episodes'
    ? 'absolute inset-y-0 right-0 h-full max-h-full w-[min(20rem,88%)] divide-y-0 rounded-none bg-[#10151d] text-white'
    : 'absolute inset-y-0 right-0 h-full max-h-full w-[min(22rem,92%)] rounded-none bg-[#141e28] text-white',
  header: panelMode.value === 'episodes' ? 'min-h-0 shrink-0 gap-3 p-3 sm:p-3' : 'min-h-0 shrink-0 p-3',
  title: panelMode.value === 'episodes' ? 'flex items-center gap-2 text-base text-white' : 'text-sm text-white',
  body: panelMode.value === 'episodes' ? 'min-h-0 overflow-y-auto overscroll-contain p-4 pt-0 sm:p-4 sm:pt-0' : 'min-h-0 overflow-y-auto overscroll-contain p-3'
}));
const rangeStart = ref(0);
const line = computed({ get: () => props.lineKey || '0', set: value => emit('line', value) });
const ranges = computed(() => Array.from({ length: Math.ceil(props.episodes.length / 50) }, (_, index) => ({ label: `${index * 50 + 1}–${Math.min((index + 1) * 50, props.episodes.length)} 集`, value: index * 50 })));
const visibleEpisodes = computed(() => props.episodes.slice(rangeStart.value, rangeStart.value + 50));
const volume = ref(1), volumeSupported = ref(true), pipAvailable = ref(false);
const audioMode = ref<AudioBalanceMode>('off'), audioAvailable = ref(false), audioBusy = ref(false), audioError = ref('');
type PlayerMedia = HTMLElement & PlaybackMedia & { playbackRate: number; volume: number; muted: boolean; loop: boolean; engine?: { media: HTMLMediaElement | null } };
let currentMedia: PlayerMedia | null = null;
let picture: ReturnType<typeof bindPlayerPicture> | null = null;
let audio: ReturnType<typeof createAudioBalance> | null = null;
let pendingAutoplay = false;
let saveTimer: ReturnType<typeof setTimeout> | undefined, panelTimer: ReturnType<typeof setTimeout> | undefined;
function saveSettings() { clearTimeout(saveTimer);try { localStorage.setItem(playerSettingsKey, JSON.stringify(settings.value));storageError.value = ''; } catch { storageError.value = '设置未能保存，当前播放仍有效。'; } }
function openPanel(mode: 'settings' | 'episodes') { clearTimeout(panelTimer);panelTimer = setTimeout(() => { panelMode.value = mode;panelOpen.value = true; }, 0); }
function openEpisodes() { openPanel('episodes'); }
function selectEpisode(key: string) { panelOpen.value = false;if (key === props.episodeKey) play(); }
function updateMediaControls() { volume.value = currentMedia?.muted ? 0 : currentMedia?.volume ?? 1;audioAvailable.value = Boolean(currentMedia && audioBalanceTarget(currentMedia));pipAvailable.value = player.value?.store?.state.pictureInPictureAvailability === 'available'; }
function setVolume(value: number) { if (!currentMedia) return;currentMedia.volume = value;currentMedia.muted = value === 0;volumeSupported.value = Math.abs(currentMedia.volume - value) < 0.01;updateMediaControls(); }
async function setAudioMode(mode: AudioBalanceMode) {
  const media = currentMedia, target = media && audioBalanceTarget(media);
  if (!target || audioBusy.value) return;
  audioBusy.value = true;audioError.value = '';
  const controller = audio ||= createAudioBalance(target);
  try { await controller.setMode(mode);if (currentMedia === media) audioMode.value = mode; }
  catch { if (currentMedia === media) audioError.value = '未能开启音量均衡，请重新选择。'; }
  finally { if (currentMedia === media) audioBusy.value = false; }
}
async function enterPiP() { panelOpen.value = false;try { await player.value?.store?.requestPictureInPicture(); } catch { fullscreenMessage.value = '未能开启画中画，请再试一次。'; } }
async function exitFullscreen() {
  fullscreenMessage.value = '';
  try {
    // The fullscreen container survives episode changes, including the interval
    // when the player has released its old media and cannot resolve a target.
    if (document.fullscreenElement) await document.exitFullscreen();
    else await player.value?.store?.exitFullscreen();
  } catch { fullscreenMessage.value = '未能退出全屏，请使用浏览器返回。'; }
}
watch(settings, () => { picture?.apply();emit('lights', settings.value.lightsOff);clearTimeout(saveTimer);saveTimer = setTimeout(saveSettings, 300); }, { deep: true });
watch(() => props.episodeKey, () => { const index = props.episodes.findIndex(item => item.key === props.episodeKey);rangeStart.value = Math.max(0, Math.floor(index / 50)) * 50; }, { immediate: true });
const api = useYtvApi();
const { loggedIn } = useAuth();
const skin = ref<HTMLElement | null>(null);
const player = ref<YtvVideoPlayerElement | null>(null);
const fullscreenMessage = ref("");
const gestureFeedback = ref<PlayerGestureFeedback | null>(null);
function gestureTime(seconds: number) {
  const whole = Math.max(0, Math.floor(seconds));
  const minutes = Math.floor(whole / 60), tail = String(whole % 60).padStart(2, '0');
  return whole >= 3600 ? `${Math.floor(whole / 3600)}:${String(minutes % 60).padStart(2, '0')}:${tail}` : `${minutes}:${tail}`;
}
const gestureLabel = computed(() => {
  const feedback = gestureFeedback.value;
  if (!feedback) return '';
  if (feedback.kind === 'speed') return '2 倍速播放 · 松开恢复';
  if (feedback.kind === 'seek') return `${feedback.offset >= 0 ? '快进' : '快退'} ${gestureTime(Math.abs(feedback.offset))}`;
  if (feedback.kind === 'brightness') return `画面亮度 ${Math.round(feedback.value * 100)}%`;
  return feedback.supported ? `音量 ${Math.round(feedback.value * 100)}%` : '请使用设备音量键';
});
// Retain the implementation, but suspend the feature on every device until
// its per-frame GPU cost is suitable for playback, including mobile.
const enhancementAvailable = false;
const enhancementEnabled = ref(enhancementAvailable);
const enhancementState = ref<EnhancementState>('waiting');
const preferenceError = ref('');
const enhancementFailure = ref('');
const enhancementHelpOpen = ref(false);
let enhancement: ReturnType<typeof bindVideoEnhancement> | null = null;
const enhancementLabel = computed(() => enhancementEnabled.value && enhancementState.value === 'unavailable' ? '超分辨率失败' : `超分辨率：${enhancementEnabled.value ? '开启' : '关闭'}`);
function activateEnhancement() {
  if (enhancementEnabled.value && enhancementState.value === 'unavailable') { enhancementHelpOpen.value = !enhancementHelpOpen.value; return; }
  toggleEnhancement();
}
function toggleEnhancement() {
  if (!enhancementAvailable) return;
  enhancementHelpOpen.value = false;
  enhancementEnabled.value = !enhancementEnabled.value;
  preferenceError.value = '';
  try { localStorage.setItem(enhancementPreferenceKey, String(enhancementEnabled.value)); }
  catch { preferenceError.value = '浏览器未能保存设置，本次观看仍有效。'; }
  enhancement?.setEnabled(enhancementEnabled.value);
}
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
defineExpose({ enterFullscreen, openEpisodes, turnLightsOn: () => { settings.value.lightsOff = false; } });

onMounted(() => {
  try { settings.value = readPlayerSettings(localStorage); } catch { /* Use the initial settings when storage is unavailable. */ }
  emit('lights', settings.value.lightsOff);
  if (enhancementAvailable) {
    try { enhancementEnabled.value = localStorage.getItem(enhancementPreferenceKey) !== 'false'; } catch { /* Keep the default preference. */ }
  }
  session = createPlaybackSession({
    describeError: cause => ytvFailureMessage(cause, "没有取得可播放的地址，请重试或选择其他线路。"),
    resolve: (value) => api.playback(value.sourceKey, value.externalId, value.episodeKey),
    loadProgress: async (value) => {
      if (!loggedIn.value) return undefined;
      return mergeViewerProgress([], (await api.listProgress()).items).find(item => item.sourceKey === value.sourceKey && item.externalId === value.externalId && item.episodeKey === value.episodeKey);
    },
    mount: (manifest) => {
      const element = document.createElement(manifest.engine === "hls" ? "hlsjs-video" : "video") as PlayerMedia & { src: string; source?: { src: string; type: string; preferPlayback: "mse" } | null; load(): void };
      const playlistURL = manifest.playlistText ? URL.createObjectURL(new Blob([manifest.playlistText], { type: "application/vnd.apple.mpegurl" })) : undefined;
      element.setAttribute("playsinline", "");
      element.setAttribute("preload", "metadata");
      element.setAttribute("aria-label", `${props.title} ${props.episodeLabel}`);
      element.style.cssText = "display:block;width:100%;height:100%;object-fit:contain";
      skin.value!.append(element);
      currentMedia = element;
      const currentPicture = bindPlayerPicture(element, skin.value!, () => settings.value);
      picture = currentPicture;
      element.addEventListener('volumechange', updateMediaControls);
      const currentEnhancement = enhancementAvailable ? bindVideoEnhancement(element, skin.value!, (state, reason) => { enhancementState.value = state; enhancementFailure.value = reason || ''; if (state !== 'unavailable') enhancementHelpOpen.value = false; }) : null;
      enhancement = currentEnhancement;
      currentEnhancement?.setEnabled(enhancementEnabled.value);
      const gestures = bindPlayerGestures(skin.value!, element, feedback => { gestureFeedback.value = feedback;if (feedback?.kind === 'volume') volumeSupported.value = feedback.supported; }, { brightness: () => settings.value.brightness, setBrightness: value => { settings.value.brightness = value; } });
      let releasePreferences = () => {};
      try { releasePreferences = bindPlayerPreferences(element as unknown as HTMLMediaElement, localStorage, gestures.preferredRate); } catch { /* Playback works without browser storage. */ }
      const mediaRequest = new AbortController();
      queueMicrotask(async () => {
        // Direct video can play without CORS yet remain unreadable to WebGPU.
        // Probe only headers/a byte range; never proxy or retain the video stream.
        if (enhancementAvailable && enhancementEnabled.value && manifest.engine === 'direct' && navigator.gpu) {
          const timeout = setTimeout(() => mediaRequest.abort(), 3000);
          try {
            const response = await fetch(manifest.url, { headers: { Range: 'bytes=0-0' }, credentials: 'omit', signal: mediaRequest.signal });
            if (response.ok && element.isConnected) element.setAttribute('crossorigin', 'anonymous');
            await response.body?.cancel();
          } catch { /* Approved: keep original playback when frames cannot be read. */ }
          finally { clearTimeout(timeout); }
        }
        if (element.isConnected) {
          if (playlistURL) element.source = { src: playlistURL, type: "application/vnd.apple.mpegurl", preferPlayback: "mse" };
          else element.src = manifest.url;
        }
      });
      return { media: element, destroy: () => { mediaRequest.abort();currentPicture.dispose();picture = null;audio?.dispose();audio = null;audioMode.value = 'off';audioBusy.value = false;audioAvailable.value = false;volumeSupported.value = true;currentMedia = null;element.removeEventListener('volumechange', updateMediaControls);currentEnhancement?.dispose(); if (enhancement === currentEnhancement) enhancement = null; gestures.dispose(); releasePreferences(); element.pause(); if (manifest.engine === "hls") element.source = null; element.removeAttribute("src"); element.remove(); if (element.tagName === "VIDEO") element.load(); if (playlistURL) URL.revokeObjectURL(playlistURL); } };
    },
    storage: () => localStorage,
    onState: (value) => { snapshot.value = value;updateMediaControls();if (value.state === 'ready' && pendingAutoplay) { pendingAutoplay = false;void session?.play(); } },
    onStarted: (value) => emit("started", value),
    onProgress: (value) => emit("progress", value),
    onEnded: () => { if (settings.value.endAction === 'next') emit('ended'); }
  });
  void session.open(selection());
  document.addEventListener("visibilitychange", persist);
  window.addEventListener("pagehide", persist);
  document.addEventListener("fullscreenchange", updateFullscreen);
});
watch(() => [props.sourceKey, props.externalId, props.episodeKey], () => { panelOpen.value = false;pendingAutoplay = true;void session?.open(selection()); });
onBeforeUnmount(() => {
  document.removeEventListener("visibilitychange", persist);
  window.removeEventListener("pagehide", persist);
  document.removeEventListener("fullscreenchange", updateFullscreen);
  session?.dispose();
  clearTimeout(panelTimer);saveSettings();emit('lights', false);
});
</script>

<template>
  <div class="absolute inset-0 bg-black" :data-player-state="snapshot.state">
    <media-i18n lang="zh" class="block size-full">
      <component :is="'ytv-video-player'" ref="player" orientation-lock-type="landscape" :poster="poster || undefined" class="block size-full">
        <ytv-video-skin ref="skin" class="ytv-player-skin" :style="{ touchAction: panelOpen ? 'pan-y' : 'none' }">
          <component :is="'media-menu-item'" slot="picture-settings" class="ytv-settings-entry" @select="openPanel('settings')">画面与播放<UIcon name="i-tabler-chevron-right" class="ml-auto size-4" /></component>
          <button v-if="episodes.length > 1 && (mobile || fullscreenActive)" slot="episodes-control" type="button" class="ytv-player-button" aria-label="选集" @click="openEpisodes"><UIcon name="i-tabler-list" class="size-4" /><span>选集</span></button>
          <div v-if="mobile && fullscreenActive" class="ytv-fullscreen-header"><button type="button" class="ytv-player-button" aria-label="退出全屏" @click="exitFullscreen"><UIcon name="i-tabler-chevron-left" class="size-5" /></button><span>{{ title }} · {{ episodeLabel }}</span></div>
          <button v-if="enhancementAvailable" slot="enhancement" type="button" class="ytv-enhancement-button" :aria-pressed="enhancementEnabled" :data-enhancement-state="enhancementState" :aria-label="enhancementState === 'unavailable' && enhancementEnabled ? '超分辨率失败，查看原因' : enhancementLabel" @click.stop="activateEnhancement">{{ enhancementLabel }}</button>
          <div class="ytv-player-panels" :class="{ 'is-open': panelOpen }" @keydown.stop>
          <USlideover v-model:open="panelOpen" :portal="false" :transition="false" :close="false" side="right" :title="panelMode === 'settings' ? '画面与播放' : '选集'" :ui="panelUI">
            <template #title>{{ panelMode === 'settings' ? '画面与播放' : '选集' }}<span v-if="panelMode === 'episodes'" class="ytv-panel-count">{{ episodes.length }} 集</span></template>
            <template #actions><button type="button" aria-label="关闭面板" :class="panelMode === 'episodes' ? 'ytv-panel-close' : 'ytv-player-button ml-auto'" @click="panelOpen = false"><UIcon name="i-tabler-x" class="size-4" /><span v-if="panelMode === 'settings'">关闭</span></button></template>
            <template #body>
              <PlayerSettingsContent v-if="panelMode === 'settings'" v-model="settings" :volume="volume" :volume-supported="volumeSupported" :audio-mode="audioMode" :audio-available="audioAvailable" :audio-busy="audioBusy" :audio-error="audioError" :storage-error="storageError" :pip-available="pipAvailable" @volume="setVolume" @audio="setAudioMode" @pip="enterPiP" />
              <Playlist v-else-if="playlistTitle && playbackLink" compact contained v-model:line="line" v-model:range="rangeStart" :title="playlistTitle" :movie="movie || false" :episodes="visibleEpisodes" :total="episodes.length" :line-options="lineOptions" :line-name="lineOptions.find(item => item.value === line)?.label" :ranges="ranges" :episode-key="episodeKey" :playback-link="playbackLink" @select="selectEpisode" />
            </template>
          </USlideover>
          </div>
          <div class="ytv-gesture-feedback" :class="{ 'is-fullscreen': fullscreenActive }" role="status" aria-live="polite" aria-atomic="true">
            <span v-if="gestureFeedback" :data-gesture="gestureFeedback.kind" :data-hold-speed="gestureFeedback.kind === 'speed' ? '' : undefined">
              {{ gestureLabel }}
              <small v-if="gestureFeedback.kind === 'seek'">{{ gestureTime(gestureFeedback.target) }} / {{ gestureTime(gestureFeedback.duration) }} · 松开跳转</small>
            </span>
          </div>
        </ytv-video-skin>
      </component>
    </media-i18n>
    <div v-if="enhancementAvailable && enhancementHelpOpen" role="status" class="absolute bottom-14 right-3 z-30 max-w-[min(22rem,calc(100%-1.5rem))] rounded-lg border border-white/20 bg-black/95 p-3 text-xs leading-5 text-white">
      <p>{{ enhancementFailure }}</p>
      <div class="mt-2 flex justify-end gap-4">
        <button type="button" class="text-white/70" @click="enhancementHelpOpen = false">收起</button>
        <button type="button" class="text-[var(--ui-primary)]" @click="toggleEnhancement">关闭超分辨率</button>
      </div>
    </div>
    <p v-if="enhancementAvailable && preferenceError" role="status" class="absolute inset-x-3 top-3 z-30 rounded-lg bg-black/80 px-3 py-2 text-center text-xs text-white">{{ preferenceError }}</p>
    <p v-if="fullscreenMessage" role="status" class="absolute inset-x-3 top-3 z-30 rounded-lg bg-black/80 px-3 py-2 text-center text-xs text-white">{{ fullscreenMessage }}</p>
    <button v-if="['ready', 'paused', 'ended'].includes(snapshot.state)" type="button" data-primary-play :aria-label="snapshot.state === 'ended' ? '重新播放' : '播放'" class="absolute left-1/2 top-1/2 grid size-14 -translate-x-1/2 -translate-y-1/2 place-items-center rounded-2xl border border-white/20 bg-black/65 text-white shadow-lg transition hover:border-white/50 hover:bg-black/85 focus-visible:outline-2 focus-visible:outline-offset-4 focus-visible:outline-[var(--ui-primary)] sm:size-16" @click="play"><UIcon :name="snapshot.state === 'ended' ? 'i-tabler-refresh' : 'i-tabler-player-play-filled'" class="size-5 sm:size-6" /></button>
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
.ytv-player-skin { display: block; width: 100%; height: 100%; touch-action: none; user-select: none; -webkit-touch-callout: none; --media-accent-color: var(--ui-primary); --media-accent-text-color: #10212e; --media-object-fit: contain; --media-border-radius: 0; }
.ytv-player-button{display:inline-flex;align-items:center;justify-content:center;gap:4px;min-width:36px;min-height:36px;padding:6px;border:0;border-radius:6px;background:rgb(0 0 0 / 60%);color:white;font:500 12px/1.4 var(--font-sans);cursor:pointer;white-space:nowrap}
.ytv-player-button:focus-visible{outline:2px solid var(--ui-primary);outline-offset:2px}
.ytv-panel-count{font-size:12px;font-weight:400;color:#9aa5b5;font-variant-numeric:tabular-nums}
.ytv-panel-close{display:grid;place-items:center;flex-shrink:0;width:40px;height:40px;margin-left:auto;border:0;border-radius:9px;background:transparent;color:#9aa5b5;cursor:pointer}
.ytv-panel-close:hover{background:#ffffff08;color:white}.ytv-panel-close:focus-visible{outline:2px solid var(--ui-primary);outline-offset:2px}
.ytv-player-panels{position:absolute;inset:0;z-index:80;pointer-events:none}.ytv-player-panels.is-open{pointer-events:auto}
.ytv-settings-entry{display:flex;align-items:center;gap:16px;padding:10px 12px;color:white;cursor:pointer;font-size:13px}
.ytv-settings-entry:hover,.ytv-settings-entry:focus-visible{background:rgb(255 255 255 / 15%)}
.ytv-fullscreen-header{position:absolute;top:max(8px,env(safe-area-inset-top));left:max(8px,env(safe-area-inset-left));right:12px;z-index:25;display:flex;align-items:center;gap:10px;pointer-events:none;color:white;font-size:13px;text-shadow:0 1px 3px black}
.ytv-fullscreen-header button{pointer-events:auto}.ytv-fullscreen-header>span{overflow:hidden;text-overflow:ellipsis;white-space:nowrap}
.ytv-enhancement-button { flex-shrink:0; padding:6px; border:0; border-radius:6px; background:rgb(0 0 0 / 65%); color:white; font:500 12px/1.4 var(--font-sans); cursor:pointer; white-space:nowrap; }
.ytv-enhancement-button[aria-pressed="true"]:not([data-enhancement-state="unavailable"]) { color:var(--ui-primary); }
.ytv-enhancement-button:focus-visible { outline:2px solid var(--ui-primary); outline-offset:2px; }
.ytv-gesture-feedback { position: absolute; inset: 0; z-index: 20; pointer-events: none; display: flex; justify-content: center; align-items: flex-start; padding-top: 1rem; }
.ytv-gesture-feedback span { border-radius: 12px; background: rgb(0 0 0 / 75%); padding: 0.5rem 1rem; color: white; font-size: 0.75rem; text-align: center; font-variant-numeric: tabular-nums; }
.ytv-gesture-feedback small { display: block; margin-top: 4px; color: rgb(255 255 255 / 75%); font-size: 0.6875rem; }
.ytv-gesture-feedback.is-fullscreen { padding-top: 4rem; }
.ytv-gesture-feedback.is-fullscreen span { font-size: 0.875rem; }
@media (pointer: coarse) {
  .ytv-player-skin { --media-scale-unit: 20px; }
  .ytv-gesture-feedback { padding-top: max(12px, env(safe-area-inset-top)); }
}
</style>
