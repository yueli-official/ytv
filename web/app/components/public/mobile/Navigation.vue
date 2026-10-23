<script setup lang="ts">
const route=useRoute();
const items=[{to:'/',label:'发现',icon:'i-tabler-compass'},{to:'/search',label:'搜索',icon:'i-tabler-search'},{to:'/history',label:'历史',icon:'i-tabler-history'}];
</script>
<template><nav class="mp-navigation" aria-label="移动端主导航"><NuxtLink v-for="item in items" :key="item.to" :to="item.to" :aria-current="route.path===item.to?'page':undefined"><UIcon :name="item.icon"/><span>{{item.label}}</span></NuxtLink></nav></template>
<style scoped>
.mp-navigation{position:fixed;z-index:30;bottom:0;inset-inline:0;display:flex;justify-content:space-around;padding:6px 12px calc(6px + env(safe-area-inset-bottom));border-top:1px solid var(--ytv-line);background:#0d141bf5;backdrop-filter:blur(16px)}
a{display:flex;flex-direction:column;gap:4px;align-items:center;justify-content:center;min-width:64px;min-height:48px;color:#aebcca;font-size:11px;text-decoration:none}a>span:first-child{width:22px;height:22px}a[aria-current]{color:#f494b7}
</style>
