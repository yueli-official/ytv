<script setup lang="ts">
import type { useTitleView } from '~/composables/useTitleView';
const props=defineProps<{view:ReturnType<typeof useTitleView>}>();
const { isSingleFeature, titleLines, watchPath, searchReturnPath, browserProgress, mergeViewerProgress, ytvFailureMessage, route, api, loggedIn, returnTo, playbackLink, sourceKey, externalId, data, pending, error, refresh, selectedLine, rangeStart, lines, line, movie, episodes, ranges, history, posterFailed, summaryExpanded, toggleSummary, lineChosen, historyRequest, selectLine, applyHistory, revealResume, readHistory, resume, startKey, resumeTime }=toRefs(props.view);
</script>
<template>
  <div class="ytv-detail min-h-[80dvh] text-white">
    <img v-if="data?.title.imageUrl && !posterFailed" :src="data.title.imageUrl" alt="" aria-hidden="true" class="ytv-art-backdrop" referrerpolicy="no-referrer">
    <div class="ytv-container">
      <div v-if="pending" class="grid gap-6 py-8 sm:grid-cols-[16rem_1fr]"><USkeleton class="aspect-[2/3] rounded-xl" /><USkeleton class="h-48" /></div>
      <UAlert v-else-if="error" color="error" variant="soft" class="mt-8" title="影片详情暂时不可用" :description="ytvFailureMessage(error, '请重试，或返回搜索选择其他来源。')"><template #actions><UButton label="重试" @click="() => refresh()" /><UButton :to="returnTo || '/search'" label="返回搜索" color="neutral" variant="soft" /></template></UAlert>
      <template v-else-if="data">
        <section class="ytv-title-intro">
          <div class="ytv-title-poster relative overflow-hidden rounded-md bg-black/20">
            <img v-if="data.title.imageUrl && !posterFailed" :src="data.title.imageUrl" :alt="`${data.title.name} 海报`" class="block h-auto w-full" referrerpolicy="no-referrer" @error="posterFailed = true">
            <div v-else class="grid aspect-[2/3] place-items-center"><UIcon name="i-tabler-movie" class="size-12 text-white/30" /></div>
          </div>
          <div class="ytv-title-facts min-w-0 self-end">
            <h1>{{ data.title.name }}</h1>
            <p class="mt-3 text-sm leading-6 sm:mt-6 sm:text-base sm:leading-8 text-white/90">{{ [data.title.year, data.title.category, data.title.area, data.title.language, data.title.remarks].filter(Boolean).join(' · ') }}</p>
            <p v-if="Number(data.title.score) > 0" class="mt-3 text-amber-300">{{ Number(data.title.score).toFixed(1) }} <span class="text-sm text-white/75">来源评分</span></p>
          </div>
            <div class="ytv-title-actions">
              <UButton v-if="startKey !== undefined" :aria-label="resume ? '继续观看' : movie ? '播放正片' : '开始播放'" icon="i-tabler-player-play-filled" :to="playbackLink(startKey)" color="neutral" variant="ghost" class="ytv-primary">
                {{ resume ? '继续观看' : movie ? '播放正片' : '开始播放' }}<span v-if="resume" class="ml-2 hidden text-sm font-semibold sm:inline">{{ resume.episodeTitle || '上次观看' }} · {{ resumeTime }}</span>
              </UButton>
              <UButton label="查找其他来源" icon="i-tabler-search" :to="{ path: '/search', query: { q: data.title.name } }" color="neutral" variant="ghost" class="ytv-secondary" />
              <p v-if="resume" class="basis-full text-xs text-white/70 sm:hidden">{{ resume.episodeTitle || '上次观看' }} · {{ resumeTime }}</p>
            </div>
          <div class="ytv-title-synopsis min-w-0">
            <h2 class="sr-only">剧情简介</h2>
            <p id="title-synopsis" class="max-w-4xl whitespace-pre-line break-words text-sm leading-7 sm:text-base sm:leading-8 text-white/90" :class="summaryExpanded ? '' : 'line-clamp-2'">{{ data.title.summary || '当前来源暂未提供剧情简介。' }}</p>
            <UButton v-if="data.title.summary" :label="summaryExpanded ? '收起简介' : '展开简介'" :icon="summaryExpanded ? 'i-tabler-chevron-up' : 'i-tabler-chevron-down'" :aria-expanded="summaryExpanded" aria-controls="title-synopsis" color="neutral" variant="ghost" size="sm" class="mt-1" @click="toggleSummary" />
          </div>
        </section>
        <section class="ytv-episodes" aria-labelledby="playback-options-title">
          <div class="ytv-episode-heading">
            <h2 id="playback-options-title">{{ movie ? '选择版本与线路' : '选集' }}</h2>
            <p class="text-sm text-white/75">当前来源：<span class="text-primary">{{ data.title.sourceName }}</span><span v-if="lines.length === 1">　|　播放线路：<span class="text-primary">{{ line?.name }}</span></span></p>
            <USelect v-if="ranges.length > 1" v-model="rangeStart" :items="ranges" aria-label="选集范围" class="w-36" />
            <NuxtLink :to="returnTo || '/'" class="ml-auto inline-flex items-center gap-2 text-base text-primary no-underline"><UIcon name="i-tabler-arrow-left" />{{ returnTo ? '返回搜索结果' : '发现影片' }}</NuxtLink>
          </div>
          <div v-if="lines.length > 1" class="mb-4 flex flex-wrap gap-2" role="group" aria-label="播放线路"><UButton v-for="item in lines" :key="item.key" :label="item.name" :color="line?.key === item.key ? 'primary' : 'neutral'" :variant="line?.key === item.key ? 'solid' : 'ghost'" :aria-pressed="line?.key === item.key" @click="selectLine(item.key)" /></div>
          <div v-if="episodes.length" class="ytv-episode-grid">
            <NuxtLink v-for="episode in episodes.slice(rangeStart, rangeStart + 50)" :key="episode.key" :to="playbackLink(episode.key)" :aria-label="episode.key === resume?.episodeKey ? `${episode.title}，上次观看` : episode.title" class="ytv-episode" :class="{ selected: episode.key === resume?.episodeKey }"><UIcon v-if="movie || episode.key === resume?.episodeKey" name="i-tabler-player-play-filled" class="size-5 shrink-0" />{{ movie ? episode.title.replace(/^第\s*1\s*集$/u, '正片') || '正片' : episode.title }}</NuxtLink>
          </div>
          <p v-else class="text-sm text-white/75">当前没有可播放的条目，可以尝试其他来源。</p>
        </section>
      </template>
    </div>
  </div>
</template>
