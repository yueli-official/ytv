<script setup lang="ts">
defineProps<{siteName:string;logoUrl?:string}>();
</script>
<template><header class="mp-header"><NuxtLink to="/" :aria-label="`${siteName} 首页`"><img v-if="logoUrl" :src="logoUrl" alt="" class="mr-2 inline-block size-7 object-contain align-middle" />{{siteName}}</NuxtLink><div><YSiteLauncher /></div></header></template>
<style scoped>.mp-header{height:60px;padding:8px 16px;display:flex;align-items:center;justify-content:space-between}.mp-header>a{font-size:24px;font-weight:800;color:var(--ui-primary);text-decoration:none}.mp-header>div{display:flex;align-items:center;gap:8px}</style>
