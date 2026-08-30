import { createBrowserUUID } from "@yueli/nuxt-runtime/runtime";

const storageKey = "ytv-visitor-v1";

export function browserVisitorID() {
  const existing = localStorage.getItem(storageKey);
  if (existing) return existing;
  const created = createBrowserUUID();
  localStorage.setItem(storageKey, created);
  return created;
}
