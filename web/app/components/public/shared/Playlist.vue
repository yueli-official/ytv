<script setup lang="ts">
import type { TitleResponse } from '~/types/ytv';
import type { RouteLocationRaw } from 'vue-router';
defineProps<{title:TitleResponse['title'];movie:boolean;total:number;episodes:TitleResponse['episodes'];lineOptions:{label:string;value:string}[];lineName?:string;ranges:{label:string;value:number}[];episodeKey:string;playbackLink:(key:string)=>RouteLocationRaw;compact?:boolean}>();
const line=defineModel<string>('line',{required:true});
const range=defineModel<number>('range',{required:true});
</script>
<template><div :class="compact?'mobile-playlist':'ytv-playlist'">
          <div class="ytv-playlist-head">
            <div v-if="!compact" class="flex items-center justify-between gap-3"><h2>{{ movie ? '观看选项' : '选集' }}</h2><span class="text-white/65">{{ movie ? `${lineOptions.length} 条线路` : `${total} 集` }}</span></div>
            <p class="text-white/80">来源：{{ title.sourceName }}</p>
            <USelect v-if="lineOptions.length > 1" v-model="line" :items="lineOptions" aria-label="播放线路" class="w-full" /><p v-else class="break-all text-sm text-white/65">线路：{{ lineName }}</p>
            <NuxtLink :to="{ path: '/search', query: { q: title.name } }" class="inline-flex items-center gap-2 text-sm text-white/70 no-underline hover:text-white"><UIcon name="i-tabler-search" class="size-4" />查找其他来源</NuxtLink>
            <USelect v-if="ranges.length > 1" v-model="range" :items="ranges" aria-label="剧集范围" class="w-full" />
          </div>
          <div class="ytv-playlist-rows">
            <NuxtLink v-for="episode in episodes" :key="episode.key" :to="playbackLink(episode.key)" data-player-episode-row class="ytv-episode" :class="{ selected: episode.key === episodeKey }" :aria-current="episode.key === episodeKey ? 'page' : undefined">
              <UIcon v-if="episode.key === episodeKey" name="i-tabler-player-play-filled" class="ytv-episode-icon size-3 shrink-0" /><span class="min-w-0 break-words">{{ movie ? episode.title.replace(/^第\s*1\s*集$/u, '正片') || '正片' : episode.title }}</span>
            </NuxtLink>
          </div>
        </div></template>
<style scoped>
.mobile-playlist .ytv-playlist-head{padding:0 0 16px;font-size:13px}
.mobile-playlist .ytv-playlist-rows{display:grid;grid-template-columns:repeat(5,minmax(0,1fr));gap:8px;padding:0;max-height:none}
.mobile-playlist .ytv-episode{min-height:44px;justify-content:center;font-size:12px;padding:6px 3px}
.mobile-playlist .ytv-episode-icon{display:none}
</style>
