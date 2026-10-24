<script setup lang="ts">
import type { AccountMenuAction } from '@yueli/ui/account-menu/pattern';
defineProps<{siteName:string;accountActions:readonly AccountMenuAction[]}>();
const emit=defineEmits<{sources:[]}>();
const route=useRoute();
const navigation=[{label:'发现',to:'/'},{label:'播放历史',to:'/history'}];
function openSourceSettings(){emit('sources');}
</script>
<template><header class="mp-header"><NuxtLink to="/" :aria-label="`${siteName} 首页`">{{siteName}}</NuxtLink><div><UButton v-if="route.path !== '/search'" icon="i-tabler-adjustments-horizontal" aria-label="来源设置" variant="ghost" color="neutral" @click="openSourceSettings"/><ConsumerAccountControl :context-actions="accountActions"/></div></header></template>
<style scoped>.mp-header{height:60px;padding:8px 16px;display:flex;align-items:center;justify-content:space-between}.mp-header>a{font-size:24px;font-weight:800;color:var(--ui-primary);text-decoration:none}.mp-header>div{display:flex;align-items:center;gap:8px}</style>
