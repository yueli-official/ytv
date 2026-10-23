<script setup lang="ts">
import type { useTitleView } from '~/composables/useTitleView';
defineProps<{view:ReturnType<typeof useTitleView>}>();
</script>
<template>
 <div class="mp-page" data-mobile-page="title">
  <UButton :to="view.returnTo||'/'" label="返回" icon="i-tabler-chevron-left" variant="ghost" color="neutral" class="w-fit"/>
  <USkeleton v-if="view.pending" class="h-64 bg-white/5"/>
  <UAlert v-else-if="view.error" title="影片详情暂时不可用" :description="view.ytvFailureMessage(view.error)" color="error" variant="soft"><template #actions><UButton label="重试" @click="()=>view.refresh()"/></template></UAlert>
  <template v-else-if="view.data">
   <section class="mp-title-intro"><img v-if="view.data.title.imageUrl&&!view.posterFailed" :src="view.data.title.imageUrl" :alt="`${view.data.title.name} 海报`" referrerpolicy="no-referrer" @error="view.posterFailed=true"><div v-else class="mp-poster-empty"><UIcon name="i-tabler-photo"/></div><div class="min-w-0"><h1>{{view.data.title.name}}</h1><p class="mp-muted mt-2">{{[view.data.title.year,view.data.title.category].filter(Boolean).join(' · ')}}</p><UButton v-if="view.startKey!==undefined" :to="view.playbackLink(view.startKey)" :label="view.resume?'继续观看':view.movie?'播放正片':'开始播放'" icon="i-tabler-player-play-filled" class="mp-primary mt-4" size="sm"/><p v-if="view.resume" class="mp-muted mt-2">{{view.resume.episodeTitle}} · {{view.resumeTime}}</p></div></section>
   <section><p class="mp-muted leading-7" :class="view.summaryExpanded?'':'line-clamp-2'">{{view.data.title.summary||'当前来源暂未提供剧情简介。'}}</p><UButton v-if="view.data.title.summary" :label="view.summaryExpanded?'收起简介':'展开简介'" :aria-expanded="view.summaryExpanded" variant="ghost" color="neutral" size="xs" @click="view.toggleSummary"/></section>
   <section class="mp-section"><div class="mp-heading"><h2>{{view.movie?'观看选项':'选集'}}</h2><span class="mp-muted">{{view.episodes.length}} {{view.movie?'个条目':'集'}}</span></div><div class="mp-tools"><USelect v-if="view.lines.length>1" :model-value="view.line?.key" :items="view.lines.map(l=>({label:l.name,value:l.key}))" aria-label="播放线路" class="w-36" @update:model-value="value=>view.selectLine(String(value))"/><span v-else class="mp-muted">{{view.line?.name}}</span><USelect v-if="view.ranges.length>1" v-model="view.rangeStart" :items="view.ranges" aria-label="选集范围" class="w-32"/></div><p class="mp-muted">{{view.data.title.sourceName}}</p><div class="mp-episodes"><NuxtLink v-for="episode in view.episodes.slice(view.rangeStart,view.rangeStart+50)" :key="episode.key" :to="view.playbackLink(episode.key)" :aria-label="episode.title" :class="{selected:episode.key===view.resume?.episodeKey}">{{view.movie?'正片':episode.title}}</NuxtLink></div><p v-if="!view.episodes.length" class="mp-muted">当前没有可播放条目。</p><UButton :to="{path:'/search',query:{q:view.data.title.name}}" label="查找其他来源" icon="i-tabler-search" variant="ghost" color="neutral" size="sm" class="w-fit"/></section>
  </template>
 </div>
</template>
