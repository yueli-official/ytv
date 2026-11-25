import { createBrowserUUID, useApi as useFoundationApi } from "@yueli/nuxt-runtime/runtime";
import type { JsonValue } from "@yueli/http-runtime";
import type { SourceReview, AdminSource, AvailableSource, DiscoveryResponse, PlaybackManifest, PlaybackProgress, PublicSource, SearchAnalyticsOverview, SiteProfile, SourcePreference, SourceProbe, TitleResponse, TrendingItem } from "~/types/ytv";

export function useYtvApi() {
  const api = useFoundationApi("platform");
  async function sourceCollection<T>(path: `/${string}`, auth: "required" | "optional") {
    const result = await api.request<{ items: T[] }>(path, { auth });
    return { sources: result.items };
  }
  return {
    setup() {
      return api.request<{ claimed: boolean; canClaim: boolean }>("/api/v1/authorization/setup", { auth: "optional" });
    },
    claim() {
      return api.request<{ claimed: boolean; created: boolean }>("/api/v1/authorization/setup/claim", { method: "POST", auth: "required", replayAfterReauth: "never" });
    },
    meta() {
      return api.request<{ name: string; description: string; siteKey: string; capabilities: string[] }>("/api/v1/meta", { auth: "optional" });
    },
    search(query: string, page = 1, sources: string[] = []) {
      return api.request<DiscoveryResponse>("/api/v1/search", {
        query: { q: query, page, sources: sources.join(","), eventId: `search_${createBrowserUUID()}` },
        auth: "optional"
      });
    },
    sourceReview(key: string) {
      return api.request<{review: SourceReview}>(`/api/v1/me/source-reviews/${encodeURIComponent(key)}`, {auth: "optional"});
    },
    saveSourceReview(key: string, input: {ads: string; quality: string}) {
      return api.request<{review: SourceReview}>(`/api/v1/me/source-reviews/${encodeURIComponent(key)}`, {method: "PUT", body: input, auth: "optional", replayAfterReauth: "never"});
    },
    publicSources() {
      return sourceCollection<PublicSource>("/api/v1/sources", "optional");
    },
    discover(page = 1) {
      return api.request<DiscoveryResponse>("/api/v1/discover", { query: { page }, auth: "optional" });
    },
    trending(type = "movie", tag = "热门", page = 1, limit = 12) {
      return api.request<{ type: string; tag: string; page: number; items: TrendingItem[] }>("/api/v1/trending", {
        query: { type, tag, page, limit }, auth: "optional"
      });
    },
    title(sourceKey: string, externalId: string) {
      return api.request<TitleResponse>(`/api/v1/titles/${encodeURIComponent(sourceKey)}/${encodeURIComponent(externalId)}`, { auth: "optional" });
    },
    playback(sourceKey: string, externalId: string, episodeKey: string) {
      return api.request<PlaybackManifest>(`/api/v1/playback/${encodeURIComponent(sourceKey)}/${encodeURIComponent(externalId)}/${encodeURIComponent(episodeKey)}`, { auth: "optional" });
    },
    recordPageView(input: { eventId: string; path: string; referrer?: string; visitorId: string; occurredAt: string }) {
      return api.request<{ counted: boolean; replay: boolean }>("/api/v1/traffic/views", {
        method: "POST", body: input as unknown as JsonValue, auth: "optional"
      });
    },
    recordPlaybackStart(input: { eventId: string; sourceKey: string; externalId: string; episodeKey: string; visitorId: string; occurredAt: string }) {
      return api.request<{ counted: boolean; replay: boolean }>("/api/v1/traffic/playback", {
        method: "POST", body: input as unknown as JsonValue, auth: "optional"
      });
    },
    adminSearches(days = 14, limit = 10) {
      return api.request<SearchAnalyticsOverview>("/api/v1/admin/searches", {
        query: { days, limit }, auth: "required"
      });
    },
    async adminSources() {
      const result = await api.request<{ items: AdminSource[] }>("/api/v1/admin/sources", { auth: "required" });
      return { sources: result.items };
    },
    availableSources() {
      return sourceCollection<AvailableSource>("/api/v1/admin/source-catalog", "required");
    },
    installSource(sourceKey: string, revision: number) {
      return api.request<{ source: AdminSource }>("/api/v1/admin/sources", {
        method: "POST", body: { sourceKey, revision } as unknown as JsonValue, auth: "required", replayAfterReauth: "never"
      });
    },
    updateSource(sourceKey: string, input: { name: string; enabled: boolean; homeVisible: boolean; priority: number; revision: number; recommended?: boolean }) {
      return api.request<{ source: AdminSource }>(`/api/v1/admin/sources/${encodeURIComponent(sourceKey)}`, {
        method: "PATCH", body: input as unknown as JsonValue, auth: "required", replayAfterReauth: "never"
      });
    },
    deleteSource(sourceKey: string, revision: number) {
      return api.request<{ source: AvailableSource }>(`/api/v1/admin/sources/${encodeURIComponent(sourceKey)}`, {
        method: "DELETE", body: { revision } as unknown as JsonValue, auth: "required", replayAfterReauth: "never"
      });
    },
    probeSource(sourceKey: string) {
      return api.request<SourceProbe>(`/api/v1/admin/sources/${encodeURIComponent(sourceKey)}/probe`, {
        method: "POST", auth: "required"
      });
    },
    probeAllSources() {
      return api.request<{ items: SourceProbe[] }>("/api/v1/admin/sources/probe-all", {
        method: "POST", auth: "required"
      });
    },
    listProgress(limit = 100) {
      return api.request<{ items: PlaybackProgress[] }>("/api/v1/progress", {
        query: { limit }, auth: "required"
      });
    },
    saveProgress(sourceKey: string, externalId: string, episodeKey: string, input: {
      titleName: string;
      episodeTitle: string;
      positionSeconds: number;
      durationSeconds: number;
      completed: boolean;
      updatedAt: string;
    }) {
      return api.request<{ item: PlaybackProgress; applied: boolean }>(
        `/api/v1/progress/${encodeURIComponent(sourceKey)}/${encodeURIComponent(externalId)}/${encodeURIComponent(episodeKey)}`,
        { method: "PUT", body: input as unknown as JsonValue, auth: "required", replayAfterReauth: "never" }
      );
    },
    sourcePreference() {
      return api.request<{ preference: SourcePreference }>("/api/v1/preferences/sources", { auth: "required" });
    },
    saveSourcePreference(disabledSourceKeys: string[], revision: number) {
      return api.request<{ preference: SourcePreference }>("/api/v1/preferences/sources", {
        method: "PUT", body: { disabledSourceKeys, revision } as unknown as JsonValue, auth: "required", replayAfterReauth: "never"
      });
    },
    siteBranding() {
      return api.request<{ logoUrl: string; faviconUrl: string }>("/api/v1/site/branding", { auth: "optional" });
    },
    adminSiteProfile() {
      return api.request<{ profile: SiteProfile }>("/api/v1/admin/site-profile", { auth: "required" });
    },
    updateSiteProfile(input: { name: string; description: string; revision: number; logoUrl?: string; faviconUrl?: string }) {
      return api.request<{ profile: SiteProfile }>("/api/v1/admin/site-profile", {
        method: "PUT", body: input as unknown as JsonValue, auth: "required", replayAfterReauth: "never"
      });
    },
    adminAuthorization() {
      return api.request<{
        currentUserKey: string;
        administratorRole: string;
        capabilities: string[];
        initialClaimAvailable: boolean;
        bootstrapMode: boolean;
      }>("/api/v1/admin/authorization", { auth: "required" });
    },
    adminSession() {
      return api.request<{ allowed: boolean; userKey: string }>("/api/v1/admin/session", { auth: "required" });
    }
  };
}
