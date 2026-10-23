<script setup lang="ts">
import type { useHomeView } from '~/composables/useHomeView';
defineProps<{view:ReturnType<typeof useHomeView>}>();
</script>
<template>
 <div class="mp-page mp-home" data-mobile-page="home">
  <h1 class="sr-only">发现好电影</h1>
  <section v-if="view.activeFeature" class="mp-feature" aria-label="热门影片">
   <img v-if="view.trendingPoster(view.activeFeature)" :src="view.trendingPoster(view.activeFeature)" :alt="`${view.activeFeature.title} 海报`" referrerpolicy="no-referrer" @error="view.failTrendingPoster(view.activeFeature.id)">
   <div class="mp-feature-copy"><span class="mp-muted">豆瓣热门</span><h2>{{view.activeFeature.title}}</h2><UButton label="查找片源" size="sm" class="mp-primary" @click="view.searchTrending(view.activeFeature.title)"/></div>
  </section>
  <USkeleton v-else-if="view.trendingPending" class="h-56 bg-white/5"/>
  <p v-else class="mp-muted">热门暂不可用，试试<NuxtLink to="/search">搜索片名</NuxtLink>。</p>
  <HomeContinueWatching compact/>
  <section class="mp-section">
   <div class="mp-heading"><h2>豆瓣热门</h2><UButton label="换一批热门" icon="i-tabler-refresh" size="xs" variant="ghost" color="neutral" :loading="view.trendingPending" @click="view.changeTrendingBatch"/></div>
   <div class="mp-tools" role="group" aria-label="热门类型"><UButton v-for="type in [{key:'movie',label:'电影'},{key:'tv',label:'电视剧'}]" :key="type.key" :label="type.label" size="xs" :variant="view.trendingType===type.key?'soft':'ghost'" :aria-pressed="view.trendingType===type.key" @click="view.selectType(type.key)"/></div>
   <div v-if="view.trendingPending" class="mp-poster-rail"><USkeleton v-for="n in 3" :key="n" class="aspect-[2/3] bg-white/5"/></div>
   <div v-else-if="view.remainingTrending.length" class="mp-poster-rail" data-douban-grid>
    <NuxtLink v-for="item in view.remainingTrending" :key="item.id" :to="{path:'/search',query:{q:item.title}}" @click.prevent="view.searchTrending(item.title)"><img v-if="view.trendingPoster(item)" :src="view.trendingPoster(item)" :alt="`${item.title} 海报`" loading="lazy" @error="view.failTrendingPoster(item.id)"><div v-else class="mp-poster-empty"><UIcon name="i-tabler-photo"/></div><h3>{{item.title}}</h3></NuxtLink>
   </div>
   <p v-else class="mp-muted">暂无热门影片</p>
  </section>
  <section class="mp-section"><div class="mp-heading"><h2>来源更新</h2><UButton label="换一批资源" icon="i-tabler-refresh" size="xs" color="neutral" variant="ghost" :loading="view.sourcePending" @click="view.changeSourceBatch"/></div>
   <p v-if="view.sourceError" class="mp-muted">来源暂时不可用，请重试。</p>
   <div v-else-if="view.sourceData?.groups.length" class="mp-poster-rail"><NuxtLink v-for="group in view.sourceData.groups.slice(0,12)" :key="group.key" :to="`/title/${encodeURIComponent(group.candidates[0]!.sourceKey)}/${encodeURIComponent(group.candidates[0]!.externalId)}`"><img v-if="group.candidates[0]?.imageUrl" :src="group.candidates[0].imageUrl" :alt="`${group.title} 海报`" loading="lazy" referrerpolicy="no-referrer"><div v-else class="mp-poster-empty"><UIcon name="i-tabler-photo"/></div><h3>{{group.title}}</h3></NuxtLink></div>
   <p v-else class="mp-muted">{{view.sourcePending?'正在加载来源…':'本批暂无内容，试试换一批。'}}</p>
  </section>
 </div>
</template>
