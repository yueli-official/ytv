<script setup lang="ts">
import type { TitleResponse } from '~/types/ytv';
import type { RouteLocationRaw } from 'vue-router';

const props = defineProps<{
  title: TitleResponse['title']; movie: boolean; total: number; episodes: TitleResponse['episodes'];
  lineOptions: { label: string; value: string }[]; lineName?: string;
  ranges: { label: string; value: number }[]; episodeKey: string;
  playbackLink: (key: string) => RouteLocationRaw; compact?: boolean; contained?: boolean;
}>();
const line = defineModel<string>('line', { required: true });
const range = defineModel<number>('range', { required: true });
const emit = defineEmits<{ select: [key: string] }>();

function episodeNumber(title: string) {
  return title.trim().match(/^(?:第\s*)?(\d{1,4})(?:\s*[集话期])?$/u)?.[1];
}
function episodeLabel(title: string) {
  if (props.movie) return title.replace(/^第\s*1\s*集$/u, '正片') || '正片';
  const number = episodeNumber(title);
  return number ? number.padStart(2, '0') : title;
}
const numbered = computed(() => !props.movie && props.episodes.every(episode => episodeNumber(episode.title)));
</script>

<template>
  <section class="ytv-playlist" :class="{ 'is-compact': compact }" :aria-label="compact ? '剧集与线路' : movie ? '观看选项' : '选集'">
    <header v-if="!compact" class="playlist-heading">
      <h2>{{ movie ? '观看选项' : '选集' }}</h2>
      <span class="playlist-count">{{ movie ? `${lineOptions.length} 条线路` : `${total} 集` }}</span>
    </header>
    <div class="playlist-meta">
      <div class="playlist-source-row">
        <span class="playlist-source"><UIcon name="i-tabler-broadcast" class="size-4 shrink-0" /><span>{{ title.sourceName }}</span></span>
        <NuxtLink :to="{ path: '/search', query: { q: title.name } }" class="playlist-change-source" aria-label="查找其他来源"><UIcon name="i-tabler-switch-horizontal" class="size-3.5" />换源</NuxtLink>
      </div>
      <USelect v-if="lineOptions.length > 1" v-model="line" :items="lineOptions" :portal="contained ? false : undefined" aria-label="播放线路" size="sm" variant="soft" color="neutral" class="w-full" />
      <p v-else-if="lineName" class="playlist-line"><span>线路</span><span>{{ lineName }}</span></p>
      <USelect v-if="ranges.length > 1" v-model="range" :items="ranges" :portal="contained ? false : undefined" aria-label="剧集范围" size="sm" variant="soft" color="neutral" class="w-full" />
    </div>
    <nav class="playlist-episodes" :class="{ 'is-numbered': numbered, 'is-feature': movie }" aria-label="选择剧集">
      <NuxtLink v-for="episode in episodes" :key="episode.key" :to="playbackLink(episode.key)" data-player-episode-row class="playlist-episode" :class="{ 'is-current': episode.key === episodeKey }" :aria-label="movie ? episodeLabel(episode.title) : episode.title" :aria-current="episode.key === episodeKey ? 'page' : undefined" @click="emit('select', episode.key)">
        <UIcon v-if="episode.key === episodeKey" name="i-tabler-player-play-filled" class="playlist-playing size-3" />
        <span>{{ episodeLabel(episode.title) }}</span>
      </NuxtLink>
    </nav>
  </section>
</template>

<style scoped>
.ytv-playlist{min-width:0;padding:18px 16px 16px;border:1px solid var(--ytv-line);border-radius:14px;background:var(--ytv-panel,#10151d);color:var(--ui-text,#dce2ea)}
.playlist-heading{display:flex;align-items:center;justify-content:space-between;gap:12px;margin-bottom:18px}
.playlist-heading h2{margin:0;color:var(--ui-text-highlighted,#f3f5f8);font-size:16px;font-weight:600;line-height:1.4}
.playlist-count{color:var(--ui-text-muted,#9aa5b5);font-size:12px;font-variant-numeric:tabular-nums}
.playlist-meta{display:grid;gap:10px;margin-bottom:18px}
.playlist-source-row{display:flex;align-items:center;justify-content:space-between;gap:12px;min-width:0}
.playlist-source{display:flex;align-items:center;gap:7px;min-width:0;font-size:13px;font-weight:500;line-height:1.6}
.playlist-source>span{overflow-wrap:anywhere}.playlist-source>.iconify{color:var(--ui-text-muted,#9aa5b5)}
.playlist-change-source{display:inline-flex;align-items:center;gap:4px;flex-shrink:0;border-radius:5px;color:var(--ui-text-muted,#9aa5b5);font-size:12px;line-height:28px;text-decoration:none}
.playlist-change-source:hover{color:var(--ui-primary)}
.playlist-line{display:flex;align-items:baseline;gap:8px;margin:0;color:var(--ui-text-muted,#9aa5b5);font-size:12px;line-height:1.6}
.playlist-line>span:first-child{flex-shrink:0;color:#7e8c9c}.playlist-line>span:last-child{overflow-wrap:anywhere}
.playlist-episodes{display:grid;grid-template-columns:repeat(2,minmax(0,1fr));align-content:start;gap:8px;max-height:min(27rem,60dvh);overflow-y:auto;overscroll-behavior:contain;padding:4px;margin:-4px;scrollbar-width:thin;scrollbar-gutter:stable}
.playlist-episodes.is-numbered{grid-template-columns:repeat(3,minmax(0,1fr))}.playlist-episodes.is-feature{grid-template-columns:1fr}
.playlist-episode{position:relative;display:flex;align-items:center;justify-content:center;min-width:0;min-height:48px;padding:10px 14px;border:1px solid #ffffff0e;border-radius:9px;background:#ffffff04;color:#c3ceda;font-size:13px;font-weight:500;line-height:1.5;text-align:center;text-decoration:none;transition:background-color .15s,border-color .15s,color .15s}
.playlist-episode>span{overflow-wrap:anywhere}.is-numbered .playlist-episode{min-height:46px;font-size:15px;font-variant-numeric:tabular-nums}
.playlist-episode:hover{border-color:color-mix(in srgb,var(--ui-primary) 40%,transparent);background:color-mix(in srgb,var(--ui-primary) 4%,transparent);color:var(--ytv-accent-hover)}
.playlist-episode.is-current{border-color:var(--ui-primary);background:var(--ui-primary);color:var(--ytv-on-accent)}
.playlist-playing{position:absolute;left:8px;top:50%;transform:translateY(-50%)}
.playlist-episodes:not(.is-numbered) .playlist-episode.is-current{padding-inline:24px}
.playlist-episode:focus-visible,.playlist-change-source:focus-visible{outline:2px solid var(--ui-primary);outline-offset:2px}
.ytv-playlist.is-compact{padding:0;border:0;border-radius:0;background:transparent}
.is-compact .playlist-meta{margin-bottom:20px}.is-compact .playlist-episodes{max-height:none;overflow:visible;scrollbar-gutter:auto}
@media(prefers-reduced-motion:reduce){.playlist-episode{transition:none}}
</style>
