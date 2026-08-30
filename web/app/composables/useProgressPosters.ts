import type { Ref } from "vue";
import type { TitleResponse } from "~/types/ytv";

export interface ProgressTitleMeta {
  imageUrl: string;
  sourceName: string;
}

export interface ProgressTitleLocator {
  sourceKey: string;
  externalId: string;
}

function locatorKey(locator: ProgressTitleLocator) {
  return `${locator.sourceKey}\u0000${locator.externalId}`;
}

export function useProgressPosters(locators: Ref<ProgressTitleLocator[]>, limit = 24) {
  const api = useYtvApi();
  const metadata = ref<Record<string, ProgressTitleMeta>>({});
  const attempted = new Set<string>();
  let mounted = false;

  async function resolve(values: ProgressTitleLocator[]) {
    if (!mounted) return;
    const unique = values
      .filter((value, index, all) => all.findIndex((other) => locatorKey(other) === locatorKey(value)) === index)
      .slice(0, limit)
      .filter((value) => !attempted.has(locatorKey(value)));

    for (let offset = 0; offset < unique.length; offset += 4) {
      const batch = unique.slice(offset, offset + 4);
      batch.forEach((value) => attempted.add(locatorKey(value)));
      const results = await Promise.allSettled(batch.map((value) => api.title(value.sourceKey, value.externalId)));
      const next = { ...metadata.value };
      results.forEach((result, index) => {
        if (result.status !== "fulfilled") return;
        const response = result.value as TitleResponse;
        const value = batch[index];
        if (!value) return;
        next[locatorKey(value)] = {
          imageUrl: response.title.imageUrl,
          sourceName: response.title.sourceName
        };
      });
      metadata.value = next;
    }
  }

  onMounted(() => {
    mounted = true;
    void resolve(locators.value);
  });

  watch(locators, (values) => {
    void resolve(values);
  }, { deep: true });

  function metaFor(sourceKey: string, externalId: string) {
    return metadata.value[locatorKey({ sourceKey, externalId })];
  }

  return { metaFor };
}
