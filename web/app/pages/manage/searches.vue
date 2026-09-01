<script setup lang="ts">
definePageMeta({ layout: "manage", middleware: ["auth", "manage"] });
await navigateTo("/manage", { redirectCode: 302 });
</script>
<template><span /></template>
