import { readSelectedSources, writeSelectedSources } from "~/utils/search-preferences";

export function useViewerSourcePreferences() {
  const api = useYtvApi();
  const { user } = useAuth();
  const selected = useState<string[]>("ytv-selected-source-keys", () => []);
  const revision = useState<number>("ytv-source-preference-revision", () => 0);
  const syncedUser = useState<string>("ytv-source-preference-user", () => "");
  const ready = useState<boolean>("ytv-source-preference-ready", () => false);
  const saving = ref(false);
  const { data } = useAsyncData("ytv-public-sources", () => api.publicSources());
  const sources = computed(() => data.value?.sources || []);
  const availableKeys = computed(() => sources.value.map((source) => source.key));
  const userKey = computed(() => user.value?.userKey || user.value?.sub || "");

  function normalize(values: string[]) {
    const available = new Set(availableKeys.value);
    const result = [...new Set(values)].filter((key) => available.has(key));
    return result.length ? result : availableKeys.value.slice(0, 1);
  }

  async function load() {
    const keys = availableKeys.value;
    if (!import.meta.client || !keys.length) return;
    const local = normalize(readSelectedSources(keys));
    if (!userKey.value) {
      selected.value = local;
      revision.value = 0;
      syncedUser.value = "";
      ready.value = true;
      return;
    }
    if (syncedUser.value === userKey.value && ready.value) {
      selected.value = normalize(selected.value);
      return;
    }
    ready.value = false;
    try {
      const response = await api.sourcePreference();
      revision.value = response.preference.revision;
      if (response.preference.exists) {
        const disabled = new Set(response.preference.disabledSourceKeys);
        selected.value = normalize(keys.filter((key) => !disabled.has(key)));
      } else {
        selected.value = local;
        const saved = await api.saveSourcePreference(keys.filter((key) => !selected.value.includes(key)), 0);
        revision.value = saved.preference.revision;
      }
      writeSelectedSources(selected.value, keys);
      syncedUser.value = userKey.value;
      ready.value = true;
    } catch {
      selected.value = local;
      ready.value = true;
    }
  }

  async function persist(values: string[]) {
    const keys = availableKeys.value;
    selected.value = normalize(values);
    writeSelectedSources(selected.value, keys);
    if (!userKey.value) return;
    saving.value = true;
    try {
      const response = await api.saveSourcePreference(keys.filter((key) => !selected.value.includes(key)), revision.value);
      revision.value = response.preference.revision;
      syncedUser.value = userKey.value;
    } catch {
      syncedUser.value = "";
      await load();
    } finally {
      saving.value = false;
    }
  }

  function toggle(key: string) {
    if (selected.value.includes(key) && selected.value.length === 1) return;
    void persist(selected.value.includes(key) ? selected.value.filter((item) => item !== key) : [...selected.value, key]);
  }
  function selectAll() { void persist([...availableKeys.value]); }

  onMounted(() => void load());
  watch([availableKeys, userKey], () => void load(), { flush: "post" });

  return { sources, selected, ready, saving, loggedIn: computed(() => Boolean(userKey.value)), toggle, selectAll };
}
