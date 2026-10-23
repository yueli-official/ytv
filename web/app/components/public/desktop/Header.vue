<script setup lang="ts">
import type { AccountMenuAction } from '@yueli/ui/account-menu/pattern';
defineProps<{siteName:string;accountActions:readonly AccountMenuAction[]}>();
const emit=defineEmits<{sources:[]}>();
const route=useRoute();
const navigation=[{label:'发现',to:'/'},{label:'播放历史',to:'/history'}];
function openSourceSettings(){emit('sources');}
</script>
<template>
      <header class="ytv-header">
        <div class="ytv-container ytv-topbar">
          <NuxtLink to="/" class="ytv-wordmark" :aria-label="`${siteName} 首页`">{{ siteName }}</NuxtLink>
          <nav class="ytv-nav" aria-label="主导航">
            <NuxtLink v-for="item in navigation" :key="item.to" :to="item.to" :aria-current="route.path === item.to ? 'page' : undefined" class="ytv-nav-link" :class="{ active: route.path === item.to }">{{ item.label }}</NuxtLink>
          </nav>
          <div class="ytv-top-search"><SearchCommand compact history-popover :show-submit="route.path === '/search'" :initial-value="route.path === '/search' ? String(route.query.q || '') : ''" :initial-sources="route.path === '/search' ? String(route.query.sources || '').split(',').filter(Boolean) : undefined" /></div>
          <div class="ytv-top-actions">
            <button type="button" aria-label="来源设置" class="ytv-source-trigger" @click="openSourceSettings"><UIcon name="i-tabler-adjustments-horizontal" class="size-5" /><span>来源设置</span></button>
            <ConsumerAccountControl :context-actions="accountActions" />
          </div>
        </div>
      </header>
</template>