<script setup lang="ts">
import type { useSearchView } from '~/composables/useSearchView';
defineProps<{view:ReturnType<typeof useSearchView>}>();
</script>
<template>
 <div class="mp-page" data-mobile-page="search" data-search-workspace>
  <h1 class="sr-only">搜索影视</h1>
  <SearchCommand compact :show-history="false" :initial-value="view.query" :initial-sources="view.sourceKeys"/>
  <div class="mp-heading"><UButton label="来源设置" icon="i-tabler-adjustments-horizontal" color="neutral" variant="outline" size="sm" @click="() => { view.sourceSettingsOpen=true; }"/><UButton v-if="view.query" label="重新搜索" icon="i-tabler-refresh" color="neutral" variant="ghost" size="sm" :loading="view.pending" @click="view.retry"/></div>
  <p v-if="view.query" class="mp-muted" role="status" data-search-summary>{{view.showLoading?'正在搜索…':`本页 ${view.data?.groups.length||0} 个影片`}}</p>
  <details v-if="view.data && !view.data.complete" class="mp-notice"><summary>部分来源未返回结果</summary><ul><li v-for="s in view.data.sources.filter(s=>s.outcome!=='success')" :key="s.key">{{s.name}} · {{s.outcome==='verification_required'?'需要验证':s.outcome==='timeout'?'请求超时':'暂时不可用'}}</li></ul></details>
  <div v-if="view.showLoading" class="space-y-4"><USkeleton v-for="n in 3" :key="n" class="h-28 bg-white/5"/></div>
  <UAlert v-else-if="view.error" title="搜索暂时不可用" :description="view.ytvFailureMessage(view.error)" color="error" variant="soft"/>
  <SearchResultGrid v-else-if="view.data?.groups.length" :groups="view.data.groups" :return-to="view.route.fullPath" compact/>
  <UEmpty v-else-if="view.verificationSources.length" title="来源需要验证" description="本站暂不能完成这些来源的验证，请换个来源。"/>
  <UEmpty v-else-if="view.query" title="没有找到相关影片" description="试试更短的片名，或调整来源。"/>
  <SearchCommand v-else history-only/>
  <nav v-if="view.data?.groups.length || view.page>1" class="mp-pagination" aria-label="搜索结果分页"><UButton label="上一页" :disabled="view.page<=1||view.showLoading" :to="view.pageLink(view.page-1)" color="neutral" variant="soft"/><span>{{view.page}}</span><UButton label="下一页" :disabled="!view.data?.groups.length||view.showLoading" :to="view.pageLink(view.page+1)" color="neutral" variant="soft"/></nav>
 </div>
</template>
