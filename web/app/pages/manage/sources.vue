<script setup lang="ts">
import { ManagePage } from "@yueli/ui/admin";
import { CollectionPanel, type CollectionPanelMessages, type CollectionPanelState } from "@yueli/ui/collection/pattern";
import type { AdminSource, AvailableSource } from "~/types/ytv";

definePageMeta({ layout: "manage", middleware: ["auth", "manage"] });
useSeoMeta({ title: "数据源" });

interface SourceDraft {
  name: string;
  enabled: boolean;
  homeVisible: boolean;
  priority: number;
  revision: number;
}

const api = useYtvApi();
const { data, error, refresh, status } = await useAsyncData("ytv-admin-sources", () => api.adminSources());
const { data: catalog, refresh: refreshCatalog } = await useAsyncData("ytv-admin-source-catalog", () => api.availableSources());
const drafts = reactive<Record<string, SourceDraft>>({});
const probes = reactive<Record<string, { available: boolean; latencyMs: number; detail: string }>>({});
const saveState = reactive<Record<string, "pending" | "success" | "error" | undefined>>({});
const saveRequested = reactive<Record<string, boolean>>({});
const saving = reactive<Record<string, boolean>>({});
const probing = reactive<Record<string, boolean>>({});
const sourceQuery = ref("");
const page = ref(1);
const pageSize = ref(10);
const probingAll = ref(false);
const probeSummary = ref("");
const addOpen = ref(false);
const installingKey = ref("");
const deleteOpen = ref(false);
const deleting = ref(false);
const deleteCandidate = ref<AdminSource>();
const operationError = ref("");
const saveTimers = new Map<string, ReturnType<typeof setTimeout>>();

const filteredSources = computed(() => {
  const query = sourceQuery.value.trim().toLocaleLowerCase("zh-CN");
  if (!query) return data.value?.sources || [];
  return (data.value?.sources || []).filter((source) => [source.name, source.key, source.kind].some((value) => value.toLocaleLowerCase("zh-CN").includes(query)));
});
const visibleSources = computed(() => filteredSources.value.slice((page.value - 1) * pageSize.value, page.value * pageSize.value));
const availableSources = computed(() => catalog.value?.sources || []);
const panelState = computed<CollectionPanelState>(() => status.value === "pending" ? "loading" : error.value ? "error" : "ready");
const messages: CollectionPanelMessages = {
  searchPlaceholder: "搜索来源名称、标识或类型", searchAction: "搜索来源", filtersAction: "筛选",
  activeFilters: (count) => `${count} 个筛选条件`, clearFilters: "清除筛选", selectPage: "选择本页来源",
  selectItem: (label) => `选择 ${label}`, bulkRegion: "来源批量操作", selected: (count) => `已选择 ${count} 个来源`,
  selectAllResults: "选择全部结果", clearSelection: "清除选择", emptyTitle: "没有匹配的数据源",
  emptyDescription: "调整搜索词，或从已批准来源中添加一个数据源。", errorTitle: "来源设置暂时不可用", retry: "重试",
  showing: (first, last, total) => `显示 ${first}–${last}，共 ${total} 个来源`, pageSize: "每页", pagination: "来源分页",
  pageSizeControl: "每页来源数", pageSizeOption: (size) => `${size} 个`
};

watch(sourceQuery, () => { page.value = 1; });
watch(data, (value) => {
  const active = new Set((value?.sources || []).map((source) => source.key));
  for (const key of Object.keys(drafts)) {
    if (active.has(key)) continue;
    delete drafts[key];
    delete saveState[key];
    delete probes[key];
  }
  for (const source of value?.sources || []) {
    drafts[source.key] = { name: source.name, enabled: source.enabled, homeVisible: source.homeVisible, priority: source.priority, revision: source.revision };
  }
}, { immediate: true });

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
    } catch {
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
  for (const timer of saveTimers.values()) clearTimeout(timer);
  saveTimers.clear();
});

async function probe(sourceKey: string) {
  probing[sourceKey] = true;
  try { probes[sourceKey] = (await api.probeSource(sourceKey)).health; }
  catch { probes[sourceKey] = { available: false, latencyMs: 0, detail: "检测请求失败" }; }
  finally { probing[sourceKey] = false; }
}
async function probeAll() {
  probingAll.value = true;
  probeSummary.value = "";
  try {
    const results = (await api.probeAllSources()).results;
    for (const item of results) probes[item.source.key] = item.health;
    probeSummary.value = `${results.filter((item) => item.health.available).length}/${results.length} 个来源可用`;
  } catch { probeSummary.value = "全部检测失败"; }
  finally { probingAll.value = false; }
}
async function install(source: AvailableSource) {
  installingKey.value = source.key;
  operationError.value = "";
  try {
    await api.installSource(source.key, source.revision);
    await Promise.all([refresh(), refreshCatalog()]);
    addOpen.value = false;
  } catch { operationError.value = "添加失败；来源状态可能已被其他管理员修改，请刷新后重试。"; }
  finally { installingKey.value = ""; }
}
function requestDelete(source: AdminSource) {
  deleteCandidate.value = source;
  operationError.value = "";
  deleteOpen.value = true;
}
async function removeSource() {
  const source = deleteCandidate.value;
  if (!source) return;
  deleting.value = true;
  operationError.value = "";
  try {
    await api.deleteSource(source.key, drafts[source.key]?.revision ?? source.revision);
    deleteOpen.value = false;
    deleteCandidate.value = undefined;
    await Promise.all([refresh(), refreshCatalog()]);
  } catch { operationError.value = "删除失败；来源状态可能已被其他管理员修改，请刷新后重试。"; }
  finally { deleting.value = false; }
}
</script>

<template>
  <ManagePage id="ytv-source-settings" title="数据源" icon="i-tabler-plug-connected">
    <template #subtitle>管理已批准的来源，字段修改后自动保存。</template>
    <template #actions>
      <div class="flex flex-wrap items-center justify-end gap-2">
        <span v-if="probeSummary" class="text-xs text-muted" role="status">{{ probeSummary }}</span>
        <UButton label="检测全部" icon="i-tabler-heart-rate-monitor" color="neutral" variant="outline" :loading="probingAll" @click="probeAll" />
        <UButton label="添加数据源" icon="i-tabler-plus" @click="void (addOpen = true)" />
      </div>
    </template>

    <CollectionPanel
      v-model:search="sourceQuery" label="数据源" :items="visibleSources" :item-key="(source) => source.key" :item-label="(source) => source.name"
      :messages="messages" :state="panelState" error-message="公开站保持当前运行设置；请刷新或检查 API。" :total="filteredSources.length"
      :page="page" :page-size="pageSize" :page-sizes="[5, 10, 20]" @search="sourceQuery = $event" @retry="refresh"
      @page-change="page = $event" @page-size-change="pageSize = $event; page = 1"
    >
      <template #columns>
        <span class="text-xs font-medium text-muted xl:hidden">来源配置</span>
        <div data-source-columns class="hidden grid-cols-[minmax(18rem,1fr)_7rem_7rem_8rem_13rem] items-center gap-4 xl:grid">
          <span data-source-cell="source">来源</span><span data-source-cell="search">参与搜索</span><span data-source-cell="home">首页展示</span><span data-source-cell="priority">优先级</span><span data-source-cell="actions" class="text-end">操作</span>
        </div>
      </template>

      <template #item="{ item: source }">
        <div data-source-row class="grid min-w-0 gap-4 sm:grid-cols-2 sm:items-center xl:grid-cols-[minmax(18rem,1fr)_7rem_7rem_8rem_13rem]">
          <div data-source-cell="source" class="flex min-w-0 items-start gap-3 sm:col-span-2 xl:col-span-1">
            <span class="mt-0.5 grid size-10 shrink-0 place-items-center rounded-lg bg-primary/10 text-primary"><UIcon name="i-tabler-plug" class="size-5" /></span>
            <div v-if="drafts[source.key]" class="min-w-0 flex-1">
              <UInput v-model="drafts[source.key]!.name" class="w-full" maxlength="80" :aria-label="`${source.name} 名称`" @update:model-value="schedulePersist(source.key)" />
              <p class="mt-1 truncate text-xs text-muted">{{ source.kind }} · {{ source.key }}</p>
              <p v-if="saveState[source.key] === 'pending'" class="mt-1 text-xs text-muted" role="status">正在保存…</p>
              <p v-else-if="saveState[source.key] === 'success'" class="mt-1 text-xs text-success" role="status">已自动保存</p>
              <p v-else-if="saveState[source.key] === 'error'" class="mt-1 text-xs text-error" role="alert">自动保存失败，请刷新后重试。</p>
              <p v-else-if="probes[source.key]" class="mt-1 truncate text-xs text-dimmed">{{ probes[source.key]?.latencyMs }} ms · {{ probes[source.key]?.detail }}</p>
            </div>
          </div>
          <div v-if="drafts[source.key]" data-source-cell="search" class="flex items-center justify-between gap-3 sm:justify-start">
            <span class="text-xs text-muted sm:sr-only">参与搜索</span>
            <USwitch :model-value="drafts[source.key]!.enabled" :aria-label="`参与搜索 ${source.name}`" @update:model-value="updateEnabled(source.key, Boolean($event))" />
            <UBadge v-if="probes[source.key]" :color="probes[source.key]?.available ? 'success' : 'error'" variant="subtle">{{ probes[source.key]?.available ? '可用' : '不可用' }}</UBadge>
          </div>
          <div v-if="drafts[source.key]" data-source-cell="home" class="flex items-center justify-between gap-3 sm:justify-start">
            <span class="text-xs text-muted sm:sr-only">首页展示</span>
            <USwitch :model-value="drafts[source.key]!.homeVisible" :disabled="!drafts[source.key]!.enabled" :aria-label="`在首页展示 ${source.name}`" @update:model-value="updateHomeVisible(source.key, Boolean($event))" />
          </div>
          <div v-if="drafts[source.key]" data-source-cell="priority">
            <UInput v-model.number="drafts[source.key]!.priority" class="w-full" type="number" min="0" max="100000" :aria-label="`${source.name} 优先级`" @update:model-value="schedulePersist(source.key, true)" />
          </div>
          <div data-source-cell="actions" class="flex items-center justify-end gap-2 sm:col-span-2 xl:col-span-1">
            <UButton :label="`检测 ${source.name}`" icon="i-tabler-heart-rate-monitor" color="neutral" variant="outline" size="sm" :loading="probing[source.key]" @click="probe(source.key)" />
            <UButton :label="`删除 ${source.name}`" icon="i-tabler-trash" color="error" variant="ghost" size="sm" @click="requestDelete(source)" />
          </div>
        </div>
      </template>
    </CollectionPanel>

    <UModal v-model:open="addOpen" title="添加数据源" description="只显示部署合同中已批准、当前尚未安装的来源。" :ui="{ content: 'w-full max-w-xl bg-default' }">
      <template #body>
        <div class="space-y-3">
          <UAlert v-if="operationError" color="error" variant="subtle" title="添加失败" :description="operationError" />
          <div v-if="availableSources.length" class="divide-y divide-default overflow-hidden rounded-xl border border-default">
            <div v-for="source in availableSources" :key="source.key" class="flex items-center justify-between gap-3 p-4">
              <div class="min-w-0"><p class="truncate text-sm font-medium text-highlighted">{{ source.name }}</p><p class="mt-1 truncate text-xs text-muted">{{ source.kind }} · {{ source.key }}</p></div>
              <UButton :label="`添加 ${source.name}`" icon="i-tabler-plus" size="sm" :loading="installingKey === source.key" :disabled="Boolean(installingKey)" @click="install(source)" />
            </div>
          </div>
          <UEmpty v-else icon="i-tabler-database-off" title="暂无可添加来源" description="当前已安装全部经过部署批准的数据源。删除的来源会回到这里，可随时重新添加。" />
        </div>
      </template>
      <template #footer="{ close }"><div class="flex w-full justify-end"><UButton label="关闭" color="neutral" variant="outline" @click="close" /></div></template>
    </UModal>

    <UModal v-model:open="deleteOpen" title="删除数据源？" :description="deleteCandidate ? `将从搜索和首页移除「${deleteCandidate.name}」。历史记录不会删除，之后仍可重新添加。` : ''" :ui="{ content: 'w-full max-w-md bg-default' }">
      <template #body>
        <div class="space-y-4">
          <UAlert v-if="operationError" color="error" variant="subtle" title="删除失败" :description="operationError" />
          <div class="flex justify-end gap-2">
            <UButton label="取消" color="neutral" variant="ghost" :disabled="deleting" @click="void (deleteOpen = false)" />
            <UButton :label="deleteCandidate ? `删除 ${deleteCandidate.name}` : '删除数据源'" color="error" icon="i-tabler-trash" :loading="deleting" @click="removeSource" />
          </div>
        </div>
      </template>
    </UModal>
  </ManagePage>
</template>
