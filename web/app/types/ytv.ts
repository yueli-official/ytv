export type SourceState = "available" | "degraded" | "unavailable";

export interface SourceAvailability {
  key: string;
  name: string;
  quality: string;
  state: SourceState;
  selected?: boolean;
}

export interface MediaTitle {
  id: string;
  sourceKey?: string;
  title: string;
  category: string;
  year: number;
  image: string;
  episodeLabel?: string;
  progress?: number;
  remaining?: string;
  to?: string;
}

export interface FeaturedTitle extends MediaTitle {
  summary: string;
  sources: SourceAvailability[];
}

export interface SearchCandidate {
  sourceKey: string;
  sourceName: string;
  externalId: string;
  name: string;
  year: string;
  category: string;
  area: string;
  language: string;
  score: string;
  remarks: string;
  summary: string;
  imageUrl: string;
  episodeCount: number;
}

export interface SearchGroup {
  key: string;
  title: string;
  year: string;
  category: string;
  candidates: SearchCandidate[];
}

export interface DiscoveryResponse {
  totalPages: number;
  hasNext: boolean;
  query: string;
  page: number;
  groups: SearchGroup[];
  sources: Array<{ key: string; name: string; outcome: "success" | "timeout" | "unavailable" | "verification_required"; verificationKind?: "image" | "browser"; durationMs: number; itemCount: number }>;
  attempted: number;
  succeeded: number;
  failed: number;
  complete: boolean;
}

export interface TitleResponse {
  title: SearchCandidate;
  episodes: Array<{ key: string; title: string; lineKey?: string; lineName?: string }>;
}

export interface PlaybackManifest {
  sourceName: string;
  episodeKey: string;
  url: string;
  engine: "hls" | "direct";
  timeline: string;
  mediaType?: string;
  playlistText?: string;
}

export interface SearchAnalyticsOverview {
  days: number;
  summary: {
    searches: number;
    zeroResultSearches: number;
    incompleteSearches: number;
    averageDurationMs: number;
  };
  topQueries: Array<{
    query: string;
    searches: number;
    zeroResultSearches: number;
    averageResultCount: number;
    averageDurationMs: number;
    lastSearchedAt: string;
  }>;
  series: Array<{
    day: string;
    searches: number;
    zeroResultSearches: number;
  }>;
}

export interface AdminSource {
  recommended: boolean;
  key: string;
  name: string;
  kind: string;
  catalog: string;
  catalogVersion: string;
  enabled: boolean;
  homeVisible: boolean;
  priority: number;
  revision: number;
  updatedAt: string;
  compatibility: SourceCompatibilityNotice[];
}

export interface SourceCompatibilityNotice {
  level: "notice" | "warning" | "unsupported";
  code: string;
  message: string;
}

export interface AvailableSource {
  recommended: boolean;
  key: string;
  name: string;
  kind: string;
  catalog: string;
  catalogVersion: string;
  revision: number;
  compatibility: SourceCompatibilityNotice[];
}

export interface SourceReviewSummary { total: number; adsFew: number; adsMany: number; hd: number; normal: number; blurry: number; }
export interface SourceReview { ads: string; quality: string; updatedAt: string; }

export interface PublicSource {
  reviews: SourceReviewSummary;
  recommended: boolean;
  key: string;
  name: string;
}

export interface SourcePreference {
  disabledSourceKeys: string[];
  revision: number;
  updatedAt: string;
  exists: boolean;
}

export interface SourceProbe {
  source: AdminSource;
  health: {
    available: boolean;
    checkedAt: string;
    latencyMs: number;
    detail: string;
  };
}

export interface PlaybackProgress {
  sourceKey: string;
  externalId: string;
  episodeKey: string;
  titleName: string;
  episodeTitle: string;
  positionSeconds: number;
  durationSeconds: number;
  completed: boolean;
  revision: number;
  updatedAt: string;
}

export interface SiteProfile {
  logoUrl?: string;
  faviconUrl?: string;
  siteKey: string;
  name: string;
  description: string;
  revision: number;
  updatedAt: string;
}

export interface TrendingItem {
  id: string;
  title: string;
  score: string;
  coverUrl: string;
  isNew: boolean;
}
