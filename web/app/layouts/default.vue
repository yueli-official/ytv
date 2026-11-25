<script setup lang="ts">
import type { AccountMenuAction } from '@yueli/ui/account-menu/pattern';
import DesktopHeader from '~/components/public/desktop/Header.vue';
import MobileHeader from '~/components/public/mobile/Header.vue';
import MobileNavigation from '~/components/public/mobile/Navigation.vue';
const route=useRoute();
const {mobile}=usePublicLayout();
const immersive=computed(()=>route.path.startsWith('/watch/')||route.path.startsWith('/title/'));
const sourceSettingsOpen=useState('ytv-source-settings-open',()=>false);
const api=useYtvApi();
const {data:siteMeta}=await useAsyncData('ytv-site-meta',()=>api.meta());
const siteName=computed(()=>siteMeta.value?.name||'YTV');
const { data: branding } = useNuxtData<{ logoUrl: string }>("ytv-branding");
const logoUrl = computed(() => branding.value?.logoUrl === "/favicon.svg" ? "" : branding.value?.logoUrl);
const accountActions:readonly AccountMenuAction[]=[{label:'播放历史',icon:'i-tabler-history',to:'/history'},{label:'管理后台',icon:'i-tabler-shield-cog',to:'/manage'}];
</script>
<template>
 <div class="ytv-public min-h-dvh text-white" :class="{'mp-immersive':mobile&&immersive}">
  <a href="#page-content" class="ytv-skip">跳到内容</a>
  <main id="main-content" class="ytv-main">
   <DesktopHeader v-if="!mobile" :site-name="siteName" :logo-url="logoUrl" :account-actions="accountActions" @sources="sourceSettingsOpen=true"/>
   <MobileHeader v-else-if="!immersive" :site-name="siteName" :logo-url="logoUrl" />
   <div id="page-content" tabindex="-1"><slot/></div>
  </main>
  <MobileNavigation v-if="mobile&&!immersive"/>
  <footer v-if="!mobile" class="ytv-footer border-t border-white/8"><div class="ytv-container flex flex-wrap items-center justify-between gap-4 py-8 text-xs text-white/45"><p><span class="font-semibold tracking-wide text-white/65">{{siteName}}</span><span class="mx-3 text-white/15">/</span>电影与故事，在这里相遇。</p><div class="flex items-center gap-5"><NuxtLink to="/history" class="text-inherit no-underline hover:text-white">播放历史</NuxtLink><button @click="sourceSettingsOpen=true">选择来源</button></div></div></footer>
  <USlideover v-model:open="sourceSettingsOpen" :side="mobile?'bottom':'right'" title="来源设置" description="选择搜索和热门入口使用的数据源。" :ui="{content:mobile?'ytv-public max-h-[85dvh] bg-[#141e28] text-white':'ytv-public w-full max-w-md bg-[#141e28] text-white',body:'p-4 sm:p-5'}"><template #body><ViewerSourceSettings/></template></USlideover>
 </div>
</template>
