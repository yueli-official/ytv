<script setup lang="ts">
interface TrendPoint {
  day: string;
  value: number;
}

const props = withDefaults(defineProps<{
  points: TrendPoint[];
  label: string;
  unit: string;
  compact?: boolean;
}>(), { compact: false });

const width = 720;
const height = computed(() => props.compact ? 140 : 220);
const inset = computed(() => ({ top: 16, right: 10, bottom: props.compact ? 16 : 24, left: 10 }));
const plotWidth = computed(() => width - inset.value.left - inset.value.right);
const plotHeight = computed(() => height.value - inset.value.top - inset.value.bottom);
const peak = computed(() => Math.max(1, ...props.points.map((point) => point.value)));
const coordinates = computed(() => props.points.map((point, index) => ({
  ...point,
  x: inset.value.left + (props.points.length <= 1 ? plotWidth.value / 2 : index / (props.points.length - 1) * plotWidth.value),
  y: inset.value.top + plotHeight.value - point.value / peak.value * plotHeight.value
})));
const linePath = computed(() => coordinates.value.map((point, index) => `${index ? "L" : "M"} ${point.x} ${point.y}`).join(" "));
const areaPath = computed(() => {
  if (!coordinates.value.length) return "";
  const first = coordinates.value[0]!;
  const last = coordinates.value.at(-1)!;
  const baseline = inset.value.top + plotHeight.value;
  return `${linePath.value} L ${last.x} ${baseline} L ${first.x} ${baseline} Z`;
});
const labels = computed(() => {
  if (!props.points.length) return [];
  const indexes = new Set([0, Math.floor((props.points.length - 1) / 2), props.points.length - 1]);
  return [...indexes].sort((left, right) => left - right).map((index) => props.points[index]!);
});

function shortDay(value: string) {
  const [, month = "", day = ""] = value.split("-");
  return `${month}/${day}`;
}
</script>

<template>
  <figure class="space-y-3" data-admin-trend-chart>
    <div v-if="points.length" class="w-full" :class="compact ? 'h-28 sm:h-32' : 'h-52 sm:h-60'">
      <svg class="size-full overflow-visible" :viewBox.attr="`0 0 ${width} ${height}`" preserveAspectRatio="none" role="img" :aria-label="`最近 ${points.length} 天${label}趋势，峰值 ${peak} ${unit}`">
        <title>最近 {{ points.length }} 天{{ label }}趋势，峰值 {{ peak }} {{ unit }}</title>
        <g class="text-accented" stroke="currentColor" stroke-width="1" vector-effect="non-scaling-stroke">
          <path v-for="ratio in [0, 0.5, 1]" :key="ratio" :d="`M ${inset.left} ${inset.top + plotHeight * ratio} H ${width - inset.right}`" />
        </g>
        <path :d="areaPath" class="fill-primary/10" />
        <path :d="linePath" fill="none" class="stroke-primary" stroke-width="3" stroke-linecap="round" stroke-linejoin="round" vector-effect="non-scaling-stroke" />
        <circle v-for="point in coordinates" :key="point.day" :cx.attr="point.x" :cy.attr="point.y" r="3.5" class="fill-default stroke-primary" stroke-width="2" vector-effect="non-scaling-stroke">
          <title>{{ point.day }} · {{ point.value }} {{ unit }}</title>
        </circle>
      </svg>
    </div>
    <div v-else class="grid place-items-center text-sm text-muted" :class="compact ? 'h-28 sm:h-32' : 'h-52 sm:h-60'">当前周期还没有{{ label }}记录</div>
    <div v-if="labels.length" class="flex justify-between text-xs tabular-nums text-dimmed">
      <span v-for="point in labels" :key="point.day">{{ shortDay(point.day) }}</span>
    </div>
    <figcaption class="sr-only"><span v-for="point in points" :key="point.day">{{ point.day }}：{{ point.value }} {{ unit }}。</span></figcaption>
  </figure>
</template>
