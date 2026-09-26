<script setup lang="ts">

import { ytvFailureMessage } from "~/utils/ytv-failure";
import { ManagePage } from "@yueli/ui/admin";
import { AdminRowActions } from "@yueli/ui/admin";
import type { AdminRowActionItem } from "@yueli/ui/admin";
import { CollectionPanel, type CollectionPanelMessages, type CollectionPanelState } from "@yueli/ui/collection/pattern";
import { createJsonCollectionQueryPolicy, type CollectionControl, type CollectionControlValue, type CollectionWorkflow } from "@yueli/ui/collection";
import { useVueCollectionWorkflow } from "@yueli/ui/collection/vue";
import type { AdminSource } from "~/types/ytv";

definePageMeta({ layout: "manage", middleware: ["auth", "manage"] });
useSeoMeta({ title: "数据源" });

interface SourceDraft {
  name: string;
  enabled: boolean;
  homeVisible: boolean;
  priority: number;
  revision: number;
}

type TriStateFilter = "all" | "yes" | "no";
type BulkSourceAction = "enable" | "disable" | "show" | "hide" | "probe" | "delete";
interface SourceCollectionQuery {
  q: string;
  enabled: TriStateFilter;
  homeVisible: TriStateFilter;
  page: number;
  size: number;
}

const api = useYtvApi();
const { data: catalog, refresh: refreshCatalog } = await useAsyncData("ytv-admin-source-catalog", () => api.availableSources());
const drafts = reactive<Record<string, SourceDraft>>({});
const probes = reactive<Record<string, { available: boolean; latencyMs: number; detail: string }>>({});
const saveState = reactive<Record<string, "pending" | "success" | "error" | undefined>>({});
const saveMessages = reactive<Record<string, string>>({});
const saveRequested = reactive<Record<string, boolean>>({});
const saving = reactive<Record<string, boolean>>({});
const probing = reactive<Record<string, boolean>>({});
const sourceQuery = ref("");
const allSources = ref<AdminSource[]>([]);
const probingAll = ref(false);
const toolsPending = ref(false);
const probeSummary = ref("");
const addOpen = ref(false);
const selectedAvailableKeys = ref<string[]>([]);
const bulkPending = ref(false);
const bulkAction = ref<BulkSourceAction>();
const installing = ref(false);
const deleteOpen = ref(false);
const deleting = ref(false);
const deleteCandidates = ref<AdminSource[]>([]);
const operationError = ref("");
const saveTimers = new Map<string, ReturnType<typeof setTimeout>>();

const availableSources = computed(() => catalog.value?.sources || []);
const allAvailableSelected = computed(() => availableSources.value.length > 0 && availableSources.value.every((source) => selectedAvailableKeys.value.includes(source.key)));
const defaultQuery: SourceCollectionQuery = { q: "", enabled: "all", homeVisible: "all", page: 1, size: 10 };
const queryPolicy = createJsonCollectionQueryPolicy<SourceCollectionQuery>();

async function loadSources(query: Readonly<SourceCollectionQuery>, activeWorkflow: CollectionWorkflow<AdminSource, string, SourceCollectionQuery>) {
  const token = activeWorkflow.beginLoad();
  try {
    const response = await api.adminSources();
    allSources.value = response.sources;
    const needle = query.q.trim().toLocaleLowerCase("zh-CN");
    const filtered = response.sources.filter((source) => {
      const matchesText = !needle || [source.name, source.key, source.kind, source.catalog].some((value) => value.toLocaleLowerCase("zh-CN").includes(needle));
      const matchesEnabled = query.enabled === "all" || source.enabled === (query.enabled === "yes");
      const matchesHome = query.homeVisible === "all" || source.homeVisible === (query.homeVisible === "yes");
      return matchesText && matchesEnabled && matchesHome;
    });
    const start = (query.page - 1) * query.size;
    activeWorkflow.resolveLoad(token, { items: filtered.slice(start, start + query.size), total: filtered.length });
  } catch {
    activeWorkflow.rejectLoad(token, { key: "ytv.sources.load_failed" });
  }
}

const { snapshot: collection, workflow, reload: refresh } = useVueCollectionWorkflow({
  initialQuery: defaultQuery, queryPolicy, keyOf: (source: AdminSource) => source.key, load: loadSources
});
const query = computed(() => collection.value.query);
const visibleSources = computed(() => collection.value.items);
const page = computed({ get: () => query.value.page, set: (value: number) => workflow.setQuery({ ...query.value, page: value }) });
const pageSize = computed({ get: () => query.value.size, set: (value: number) => workflow.setQuery({ ...query.value, size: value, page: 1 }) });
const selectionCount = computed(() => collection.value.selection.count);
const selectedSourceKeys = computed(() => collection.value.selection.mode === "keys" ? collection.value.selection.keys : []);
const panelState = computed<CollectionPanelState>(() => collection.value.issue ? "error" : ["idle", "loading", "refreshing"].includes(collection.value.loadState) ? "loading" : "ready");
const sourceControls = computed<CollectionControl[]>(() => [
  { kind: "select", id: "enabled", label: "参与搜索", value: query.value.enabled, options: [{ label: "全部", value: "all" }, { label: "已启用", value: "yes" }, { label: "未启用", value: "no" }], icon: "i-tabler-search" },
  { kind: "select", id: "homeVisible", label: "首页展示", value: query.value.homeVisible, options: [{ label: "全部", value: "all" }, { label: "已展示", value: "yes" }, { label: "未展示", value: "no" }], icon: "i-tabler-home" }
]);
const activeFilterCount = computed(() => Number(query.value.enabled !== "all") + Number(query.value.homeVisible !== "all"));
const activeFilters = computed(() => [
  query.value.enabled !== "all" ? `参与搜索：${query.value.enabled === "yes" ? "已启用" : "未启用"}` : "",
  query.value.homeVisible !== "all" ? `首页展示：${query.value.homeVisible === "yes" ? "已展示" : "未展示"}` : ""
].filter(Boolean));
const messages: CollectionPanelMessages = {
  searchPlaceholder: "搜索来源名称、标识或类型", searchAction: "搜索来源", filtersAction: "筛选",
  activeFilters: (count) => `${count} 个筛选条件`, clearFilters: "清除筛选", selectPage: "选择本页来源",
  selectItem: (label) => `选择 ${label}`, bulkRegion: "来源批量操作", selected: (count) => `已选择 ${count} 个来源`,
  selectAllResults: "选择全部结果", clearSelection: "清除选择", emptyTitle: "没有匹配的数据源",
  emptyDescription: "调整搜索词，或从已批准来源中添加一个数据源。", errorTitle: "来源设置暂时不可用", retry: "重试",
  showing: (first, last, total) => `显示 ${first}–${last}，共 ${total} 个来源`, pageSize: "每页", pagination: "来源分页",
  pageSizeControl: "每页来源数", pageSizeOption: (size) => `${size} 个`
};

let searchTimer: ReturnType<typeof setTimeout> | undefined;
watch(sourceQuery, (value) => {
  if (searchTimer) clearTimeout(searchTimer);
  searchTimer = setTimeout(() => workflow.setQuery({ ...query.value, q: value.trim(), page: 1 }), 300);
});
watch(addOpen, (open) => {
  if (open) {
    selectedAvailableKeys.value = [];
    operationError.value = "";
  }
});
watch(allSources, (value) => {
  const active = new Set(value.map((source) => source.key));
  for (const key of Object.keys(drafts)) {
    if (active.has(key)) continue;
    delete drafts[key];
    delete saveState[key];
    delete probes[key];
  }
  for (const source of value) {
    drafts[source.key] = { name: source.name, enabled: source.enabled, homeVisible: source.homeVisible, priority: source.priority, revision: source.revision };
  }
}, { immediate: true });
function toggleAllAvailable(selected: boolean) {
  selectedAvailableKeys.value = selected ? availableSources.value.map((source) => source.key) : [];
}
function allAvailableChanged(event: Event) { toggleAllAvailable((event.currentTarget as HTMLInputElement).checked); }
function isSourceSelected(key: string) { return workflow.isSelected(key); }
function toggleOne(key: string, selected?: boolean) { if (selected === undefined || selected !== workflow.isSelected(key)) workflow.toggleKey(key); }
function togglePage(selected?: boolean | "indeterminate") { workflow.togglePage(selected === true); }
function clearSelection() { workflow.clearSelection(); }
function changeControl(id: string, value: CollectionControlValue) { workflow.setQuery({ ...query.value, [id]: value, page: 1 }); }
function clearFilters() { workflow.setQuery({ ...query.value, enabled: "all", homeVisible: "all", page: 1 }); }

function sameDraft(left: SourceDraft, right: SourceDraft) {
  return left.name === right.name && left.enabled === right.enabled && left.homeVisible === right.homeVisible && left.priority === right.priority;
}

async function persist(sourceKey: string) {
  const draft = drafts[sourceKey];
  if (!draft) return;
  draft.name = draft.name.trim();
  if (!draft.name) { saveState[sourceKey] = "error"; return; }
  saveRequested[sourceKey] = true;
  if (saving[sourceKey]) return;
  saving[sourceKey] = true;
  while (saveRequested[sourceKey]) {
    saveRequested[sourceKey] = false;
    const current = drafts[sourceKey];
    if (!current) break;
    const snapshot = { ...current };
    saveState[sourceKey] = "pending";
    try {
      const result = await api.updateSource(sourceKey, snapshot);
      const latest = drafts[sourceKey];
      if (!latest) break;
      latest.revision = result.source.revision;
      if (!saveRequested[sourceKey] && sameDraft(latest, snapshot)) {
        Object.assign(latest, { name: result.source.name, enabled: result.source.enabled, homeVisible: result.source.homeVisible, priority: result.source.priority, revision: result.source.revision });
      }
      saveState[sourceKey] = "success";
    } catch (error) {
      saveMessages[sourceKey] = ytvFailureMessage(error, "自动保存失败，请稍后重试。");
      saveRequested[sourceKey] = false;
      saveState[sourceKey] = "error";
    }
  }
  saving[sourceKey] = false;
}

function updateEnabled(sourceKey: string, value: boolean) {
  const draft = drafts[sourceKey];
  if (!draft) return;
  draft.enabled = value;
  if (!value) draft.homeVisible = false;
  void persist(sourceKey);
}
function updateHomeVisible(sourceKey: string, value: boolean) {
  const draft = drafts[sourceKey];
  if (!draft) return;
  draft.homeVisible = value;
  void persist(sourceKey);
}
function normalizePriority(sourceKey: string) {
  const draft = drafts[sourceKey];
  if (!draft) return;
  draft.priority = Math.max(0, Math.min(100000, Math.round(Number(draft.priority) || 0)));
  void persist(sourceKey);
}
function schedulePersist(sourceKey: string, normalize = false) {
  const current = saveTimers.get(sourceKey);
  if (current) clearTimeout(current);
  saveTimers.set(sourceKey, setTimeout(() => {
    saveTimers.delete(sourceKey);
    if (normalize) normalizePriority(sourceKey);
    else void persist(sourceKey);
  }, 350));
}
onBeforeUnmount(() => {
  if (searchTimer) clearTimeout(searchTimer);
  for (const timer of saveTimers.values()) clearTimeout(timer);
  saveTimers.clear();
});

async function probe(sourceKey: string) {
  probing[sourceKey] = true;
  try { probes[sourceKey] = (await api.probeSource(sourceKey)).health; }
  catch (error) { probes[sourceKey] = { available: false, latencyMs: 0, detail: ytvFailureMessage(error, "检测请求失败") }; }
  finally { probing[sourceKey] = false; }
}
async function probeAll() {
  probingAll.value = true;
  probeSummary.value = "";
  try {
    const results = (await api.probeAllSources()).items;
    for (const item of results) probes[item.source.key] = item.health;
    probeSummary.value = `${results.filter((item) => item.health.available).length}/${results.length} 个来源搜索可用`;
  } catch (error) { probeSummary.value = ytvFailureMessage(error, "全部检测失败"); }
  finally { probingAll.value = false; }
}
async function reconcileAvailability(mode: "enable" | "disable") {
  toolsPending.value = true;
  probeSummary.value = "";
  try {
    const results = (await api.probeAllSources()).items;
    for (const item of results) probes[item.source.key] = item.health;
    const targets = results.filter((item) => mode === "enable" ? item.health.available && !item.source.enabled : !item.health.available && item.source.enabled);
    const updates = await Promise.allSettled(targets.map((item) => api.updateSource(item.source.key, {
      name: item.source.name, enabled: mode === "enable", homeVisible: mode === "enable" ? item.source.homeVisible : false,
      priority: item.source.priority, revision: item.source.revision
    })));
    await refresh();
    const changed = updates.filter((result) => result.status === "fulfilled").length;
    probeSummary.value = mode === "enable" ? `已启用 ${changed} 个搜索可用来源` : `已停用 ${changed} 个搜索不可用来源`;
  } catch { probeSummary.value = "来源工具执行失败，请重试"; }
  finally { toolsPending.value = false; }
}
async function probeSelected() {
  bulkPending.value = true;
  const keys = [...selectedSourceKeys.value];
  const results = await Promise.allSettled(keys.map((key) => api.probeSource(key)));
  results.forEach((result, index) => {
    probes[keys[index]!] = result.status === "fulfilled" ? result.value.health : { available: false, latencyMs: 0, detail: "检测失败" };
  });
  probeSummary.value = `${results.filter((result) => result.status === "fulfilled" && result.value.health.available).length}/${results.length} 个来源搜索可用`;
  bulkPending.value = false;
}
async function updateSelected(patch: "enable" | "disable" | "show" | "hide") {
  bulkPending.value = true;
  const keys = new Set(selectedSourceKeys.value);
  const sources = allSources.value.filter((source) => keys.has(source.key));
  const results = await Promise.allSettled(sources.map((source) => {
    const draft = drafts[source.key]!;
    const enabled = patch === "enable" || patch === "show" ? true : patch === "disable" ? false : draft.enabled;
    const homeVisible = patch === "show" ? true : patch === "hide" || patch === "disable" ? false : draft.homeVisible;
    return api.updateSource(source.key, { ...draft, enabled, homeVisible });
  }));
  await refresh();
  probeSummary.value = results.every((result) => result.status === "fulfilled") ? `已更新 ${results.length} 个来源` : "部分来源更新失败，请重试";
  if (results.every((result) => result.status === "fulfilled")) clearSelection();
  bulkPending.value = false;
}
async function applyBulkAction() {
  if (!bulkAction.value || selectionCount.value === 0) return;
  if (bulkAction.value === "delete") {
    requestDeleteSelected();
    return;
  }
  if (bulkAction.value === "probe") await probeSelected();
  else await updateSelected(bulkAction.value);
  bulkAction.value = undefined;
}
async function installSelected() {
  const selected = new Set(selectedAvailableKeys.value);
  const sources = availableSources.value.filter((source) => selected.has(source.key));
  if (!sources.length) return;
  installing.value = true;
  operationError.value = "";
  const results = await Promise.allSettled(sources.map((source) => api.installSource(source.key, source.revision)));
  await Promise.all([refresh(), refreshCatalog()]);
  const failed = results.filter((result) => result.status === "rejected").length;
  if (failed === 0) {
    addOpen.value = false;
  } else {
    operationError.value = `${sources.length - failed} 个添加成功，${failed} 个添加失败。 ` + results.filter(result => result.status === "rejected").map(result => ytvFailureMessage(result.reason)).join(" ");
    selectedAvailableKeys.value = sources.filter((_source, index) => results[index]?.status === "rejected").map((source) => source.key);
  }
  installing.value = false;
}
function requestDelete(source: AdminSource) {
	deleteCandidates.value = [source];
  operationError.value = "";
  deleteOpen.value = true;
}
function requestDeleteSelected() {
  const keys = new Set(selectedSourceKeys.value);
  deleteCandidates.value = allSources.value.filter((source) => keys.has(source.key));
  operationError.value = "";
  deleteOpen.value = deleteCandidates.value.length > 0;
}
async function removeSource() {
  const sources = [...deleteCandidates.value];
  if (!sources.length) return;
  deleting.value = true;
  operationError.value = "";
  const results = await Promise.allSettled(sources.map((source) => api.deleteSource(source.key, drafts[source.key]?.revision ?? source.revision)));
  await Promise.all([refresh(), refreshCatalog()]);
  const failed = results.filter((result) => result.status === "rejected").length;
  if (failed === 0) {
    deleteOpen.value = false;
    deleteCandidates.value = [];
    clearSelection();
  } else {
    operationError.value = `${sources.length - failed} 个删除成功，${failed} 个删除失败，请重试。`;
    deleteCandidates.value = sources.filter((_source, index) => results[index]?.status === "rejected");
  }
  deleting.value = false;
}

function sourceRowActions(source: AdminSource): readonly (readonly AdminRowActionItem[])[] {
  return [[{
    id: "probe", label: `检测 ${source.name}`, icon: "i-tabler-heart-rate-monitor",
    loading: Boolean(probing[source.key]), onSelect: () => void probe(source.key)
  }], [{
    id: "delete", label: `删除 ${source.name}`, icon: "i-tabler-trash", tone: "danger",
    onSelect: () => requestDelete(source)
  }]];
}

const bulkItems = [
  { label: "启用", value: "enable" }, { label: "停用", value: "disable" },
  { label: "首页展示", value: "show" }, { label: "首页隐藏", value: "hide" },
  { label: "检测", value: "probe" }, { label: "删除", value: "delete" }
];
const sourceToolItems = computed(() => [[
  { label: "检测全部", icon: "i-tabler-heart-rate-monitor", onSelect: () => void probeAll() },
  { label: "启用搜索可用来源", icon: "i-tabler-circle-check", onSelect: () => void reconcileAvailability("enable") },
  { label: "停用搜索不可用来源", icon: "i-tabler-circle-x", onSelect: () => void reconcileAvailability("disable") }
]]);
</script>

<template>
  <ManagePage id="ytv-source-settings" title="数据源" icon="i-tabler-plug-connected">

    <template #subtitle>管理已批准的来源，字段修改后自动保存。</template>
    <template #actions>
      <div class="flex flex-wrap items-center justify-end gap-2">
        <span v-if="probeSummary" class="text-xs text-muted" role="status">{{ probeSummary }}</span>
        <UDropdownMenu :items="sourceToolItems"><UButton label="来源工具" icon="i-tabler-tool" color="neutral" variant="outline" :loading="probingAll || toolsPending" /></UDropdownMenu>
        <UButton label="添加数据源" icon="i-tabler-plus" @click="void (addOpen = true)" />
      </div>
    </template>

    <CollectionPanel
      v-model:search="sourceQuery" label="数据源" :items="visibleSources" :item-key="(source) => source.key" :item-label="(source) => source.name"
      :controls="sourceControls" :messages="messages" :state="panelState" error-message="来源设置暂时不可用，请重试。" :total="collection.total"
      :page="page" :page-size="pageSize" :page-sizes="[5, 10, 20]" :active-filter-count="activeFilterCount"
      selectable :selection-count="selectionCount" :page-selected="collection.isPageSelected" :page-indeterminate="collection.isPageIndeterminate"
      :is-selected="isSourceSelected" @search="sourceQuery = $event" @control-change="changeControl" @clear-filters="clearFilters" @retry="refresh"
      @toggle-page="togglePage" @toggle-item="toggleOne" @clear-selection="clearSelection"
      @page-change="page = $event" @page-size-change="pageSize = $event; page = 1"
    >
      <template #active-filters>
        <div class="flex flex-wrap items-center gap-2">
          <UBadge v-for="filter in activeFilters" :key="filter" :label="filter" color="neutral" variant="soft" size="sm" />
          <UButton label="清除筛选" color="neutral" variant="link" size="xs" @click="clearFilters" />
        </div>
      </template>
      <template #columns>
        <span class="text-xs font-medium text-muted xl:hidden">来源配置</span>
        <div data-source-columns class="hidden grid-cols-[minmax(18rem,1fr)_7rem_7rem_7rem_3rem] items-center gap-4 xl:grid">
          <span data-source-cell="source">来源</span><span data-source-cell="search">参与搜索</span><span data-source-cell="home">首页展示</span><span data-source-cell="priority">优先级</span><span data-source-cell="actions" class="text-end">操作</span>
        </div>
      </template>

      <template #bulk-actions>
        <USelect v-model="bulkAction" :items="bulkItems" placeholder="批量操作" aria-label="批量操作" size="xs" class="w-28" />
        <UButton label="应用" size="xs" color="primary" variant="soft" :disabled="!bulkAction" :loading="bulkPending" @click="applyBulkAction" />
      </template>

      <template #item="{ item: source }">
        <div data-source-row class="grid w-full min-w-0 gap-4 sm:grid-cols-2 sm:items-center xl:grid-cols-[minmax(18rem,1fr)_7rem_7rem_7rem_3rem]">
          <div data-source-cell="source" class="flex min-w-0 items-start gap-3 sm:col-span-2 xl:col-span-1">
            <span class="mt-0.5 grid size-10 shrink-0 place-items-center rounded-lg bg-primary/10 text-primary"><UIcon name="i-tabler-plug" class="size-5" /></span>
            <div v-if="drafts[source.key]" class="min-w-0 flex-1">
              <UInput v-model="drafts[source.key]!.name" class="w-full" maxlength="80" :aria-label="`${source.name} 名称`" @update:model-value="schedulePersist(source.key)" />
              <p class="mt-1 truncate text-xs text-muted">{{ source.catalog ? `${source.catalog} ${source.catalogVersion}` : source.kind }} · {{ source.key }}</p>
              <p v-if="saveState[source.key] === 'pending'" class="mt-1 text-xs text-muted" role="status">正在保存…</p>
              <p v-else-if="saveState[source.key] === 'success'" class="mt-1 text-xs text-success" role="status">已自动保存</p>
              <p v-else-if="saveState[source.key] === 'error'" class="mt-1 text-xs text-error" role="alert">{{ saveMessages[source.key] || "请填写来源名称。" }}</p>
              <p v-else-if="probes[source.key]" class="mt-1 truncate text-xs text-dimmed">{{ probes[source.key]?.latencyMs }} ms · {{ probes[source.key]?.detail }}</p>
            </div>
          </div>
          <div v-if="drafts[source.key]" data-source-cell="search" class="flex items-center justify-between gap-3 sm:justify-start">
            <span class="text-xs text-muted sm:sr-only">参与搜索</span>
            <USwitch :model-value="drafts[source.key]!.enabled" :aria-label="`参与搜索 ${source.name}`" @update:model-value="updateEnabled(source.key, Boolean($event))" />
            <UBadge v-if="probes[source.key]" :color="probes[source.key]?.available ? 'success' : 'error'" variant="subtle">{{ probes[source.key]?.available ? '搜索可用' : '搜索失败' }}</UBadge>
          </div>
          <div v-if="drafts[source.key]" data-source-cell="home" class="flex items-center justify-between gap-3 sm:justify-start">
            <span class="text-xs text-muted sm:sr-only">首页展示</span>
            <USwitch :model-value="drafts[source.key]!.homeVisible" :disabled="!drafts[source.key]!.enabled" :aria-label="`在首页展示 ${source.name}`" @update:model-value="updateHomeVisible(source.key, Boolean($event))" />
          </div>
          <div v-if="drafts[source.key]" data-source-cell="priority">
            <UInput v-model.number="drafts[source.key]!.priority" class="w-full" type="number" min="0" max="100000" :aria-label="`${source.name} 优先级`" @update:model-value="schedulePersist(source.key, true)" />
          </div>
          <AdminRowActions data-source-cell="actions" class="sm:col-span-2 xl:col-span-1" :label="`${source.name} 的操作`" :items="sourceRowActions(source)" presentation="overflow" />
        </div>
      </template>
    </CollectionPanel>

    <UModal v-model:open="addOpen" title="添加数据源" description="选择一个或多个来源后统一添加。" :ui="{ content: 'w-full max-w-xl bg-default' }">
      <template #body>
        <div class="space-y-3">
          <UAlert v-if="operationError" color="error" variant="subtle" title="添加失败" :description="operationError" />
          <div v-if="availableSources.length" class="max-h-[60vh] divide-y divide-default overflow-y-auto rounded-xl border border-default">
            <label class="flex cursor-pointer items-center gap-3 bg-elevated/50 px-4 py-3 text-sm font-medium text-highlighted">
              <input type="checkbox" class="size-4 accent-primary" :checked="allAvailableSelected" aria-label="选择全部可添加来源" @change="allAvailableChanged" />
              选择全部
            </label>
            <label v-for="source in availableSources" :key="source.key" class="flex cursor-pointer items-center gap-3 p-4 hover:bg-elevated/40">
              <input v-model="selectedAvailableKeys" type="checkbox" :value="source.key" class="size-4 accent-primary" :aria-label="`选择 ${source.name}`" />
              <div class="min-w-0 flex-1">
                <div class="flex flex-wrap items-center gap-2">
                  <p class="truncate text-sm font-medium text-highlighted">{{ source.name }}</p>
                  <UBadge v-if="source.catalog" color="primary" variant="subtle" size="sm">{{ source.catalog }}{{ source.catalogVersion ? ` ${source.catalogVersion}` : "" }}</UBadge>
                </div>
                <p class="mt-1 truncate text-xs text-muted">{{ source.key }}</p>
              </div>
            </label>
          </div>
          <UEmpty
            v-else icon="i-tabler-circle-check" title="全部来源已添加"
            :description="`目录中的 ${allSources.length} 个来源都已在当前数据源列表中，可关闭弹窗后直接启用、停用或检测。`"
          />
        </div>
      </template>
      <template #footer="{ close }">
        <div class="flex w-full justify-end gap-2">
          <UButton label="取消" color="neutral" variant="ghost" :disabled="installing" @click="close" />
          <UButton :label="selectedAvailableKeys.length ? `添加所选（${selectedAvailableKeys.length}）` : '添加所选'" icon="i-tabler-plus" :loading="installing" :disabled="selectedAvailableKeys.length === 0" @click="installSelected" />
        </div>
      </template>
    </UModal>

    <UModal v-model:open="deleteOpen" :title="deleteCandidates.length > 1 ? `删除 ${deleteCandidates.length} 个数据源？` : '删除数据源？'" :description="deleteCandidates.length === 1 ? `将移除「${deleteCandidates[0]?.name}」，之后仍可重新添加。` : '所选来源将从搜索和首页移除，之后仍可重新添加。'" :ui="{ content: 'w-full max-w-md bg-default' }">
      <template #body>
        <div class="space-y-4">
          <UAlert v-if="operationError" color="error" variant="subtle" title="删除失败" :description="operationError" />
          <div class="flex justify-end gap-2">
            <UButton label="取消" color="neutral" variant="ghost" :disabled="deleting" @click="void (deleteOpen = false)" />
            <UButton :label="deleteCandidates.length > 1 ? `删除所选（${deleteCandidates.length}）` : '删除'" color="error" icon="i-tabler-trash" :loading="deleting" @click="removeSource" />
          </div>
        </div>
      </template>
    </UModal>
  </ManagePage>
</template>
