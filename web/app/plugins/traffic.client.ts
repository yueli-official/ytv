import { createBrowserUUID } from "@yueli/nuxt-runtime/runtime";
import { browserVisitorID } from "~/utils/visitor";

export default defineNuxtPlugin(() => {
  const router = useRouter();
  const api = useYtvApi();
  let lastPath = "";

  async function record(path: string) {
    if (!path.startsWith("/") || path.startsWith("/manage") || path.startsWith("/auth") || path.startsWith("/api/")) return;
    if (lastPath === path) return;
    lastPath = path;
    await api.recordPageView({
      eventId: createBrowserUUID(),
      path,
      visitorId: browserVisitorID(),
      occurredAt: new Date().toISOString()
    }).catch(() => undefined);
  }

  onNuxtReady(() => { void record(router.currentRoute.value.path); });
  router.afterEach((to) => { void record(to.path); });
});
