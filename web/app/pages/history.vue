<script setup lang="ts">
import DesktopView from '~/components/public/desktop/HistoryView.vue';
import MobileView from '~/components/public/mobile/HistoryView.vue';
definePageMeta({colorMode:'dark'});
const view=useHistoryView();
const {mobile,ready}=usePublicLayout();
</script>
<template><div class="ytv-surface" :data-ready="ready"><MobileView v-if="mobile" :view="view"/><DesktopView v-else :view="view"/></div></template>
