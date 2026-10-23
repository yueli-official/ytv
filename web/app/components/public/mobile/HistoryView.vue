<script setup lang="ts">
import type { useHistoryView } from '~/composables/useHistoryView';
defineProps<{view:ReturnType<typeof useHistoryView>}>();
const toolsOpen=ref(false);
</script>
<template>
 <div class="mp-page" data-mobile-page="history">
  <div class="mp-heading"><h1>播放历史</h1><UButton label="筛选" size="sm" variant="ghost" color="neutral" :aria-expanded="toolsOpen" @click="toolsOpen=!toolsOpen"/></div>
  <UInput v-if="toolsOpen" v-model="view.filter" placeholder="搜索当前记录" aria-label="筛选已加载的记录" class="w-full"/>
  <details v-if="!view.loggedIn" class="mp-notice"><summary>记录保存在当前浏览器</summary><UButton label="登录并同步" variant="link" size="sm" @click="view.startLogin"/></details>
  <UAlert v-if="view.error" title="账号历史暂时不可用" description="本地进度仍然保留。" variant="soft" color="error"/>
  <div v-if="view.groups.length" data-history-records><section v-for="group in view.groups" :key="group.label" class="mp-history-group"><h2>{{group.label}}</h2><ul>
   <li v-for="item in group.records" :key="view.posterKey(item)" class="mp-history-row"><NuxtLink :to="view.watchPath(item.sourceKey,item.externalId,item.episodeKey)" :aria-label="`观看 ${item.titleName}`" class="mp-history-poster"><img v-if="view.posterFor(item)" :src="view.posterFor(item)" :alt="`${item.titleName} 海报`" referrerpolicy="no-referrer" @error="view.failPoster(item)"><UIcon v-else name="i-tabler-photo"/></NuxtLink><div class="min-w-0"><h3>{{item.titleName||item.externalId}}</h3><p class="mp-muted">{{item.episodeTitle||'上次观看'}}</p><span v-if="item.duration>0" class="ytv-progress mt-3" :aria-label="`已观看 ${view.percent(item.position,item.duration)}%`"><span :style="{width:`${view.percent(item.position,item.duration)}%`}"/></span></div><UButton :to="view.watchPath(item.sourceKey,item.externalId,item.episodeKey)" :label="item.completed?'重看':'继续'" :aria-label="item.completed?'再次观看':'继续观看'" size="sm" class="mp-primary"/></li>
  </ul></section></div>
  <UEmpty v-else :title="view.items.length?'没有匹配的记录':'还没有播放记录'"><template #actions><UButton v-if="view.items.length" label="清除筛选" @click="view.filter=''"/><UButton v-else to="/" label="去发现影片"/></template></UEmpty>
 </div>
</template>
