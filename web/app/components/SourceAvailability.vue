<script setup lang="ts">
import type { SourceAvailability } from "~/types/ytv";

defineProps<{ sources: SourceAvailability[] }>();
const emit = defineEmits<{ select: [key: string] }>();

const stateLabel = {
  available: "可用",
  degraded: "连接不稳定",
  unavailable: "不可用"
} as const;

const dotClass = {
  available: "bg-success",
  degraded: "bg-warning",
  unavailable: "bg-error"
} as const;
</script>

<template>
  <ul class="flex min-w-0 gap-2 overflow-x-auto pb-1" aria-label="选择播放来源" role="list">
    <li
      v-for="source in sources"
      :key="source.key"
      class="min-w-32 shrink-0"
    >
      <button type="button" class="w-full rounded-xl border px-3 py-2.5 text-left transition-colors focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-primary" :class="source.selected ? 'border-primary bg-primary/15 text-white' : 'border-white/15 bg-black/35 text-white hover:bg-black/50'" :aria-pressed="source.selected || false" @click="emit('select', source.key)">
        <span class="flex items-center gap-2 text-sm font-semibold"><span class="size-2 rounded-full" :class="dotClass[source.state]" aria-hidden="true" />{{ source.name }}</span>
        <span class="mt-1 flex items-center justify-between gap-2 text-xs text-white/65"><span>{{ source.quality }}</span><span>{{ stateLabel[source.state] }}</span></span>
      </button>
    </li>
  </ul>
</template>
