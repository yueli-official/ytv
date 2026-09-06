import { randomUUID } from "node:crypto";
import catalog from "../../../api/contracts/http-result/error-catalog.json";

const codes: Readonly<Record<number, string>> = {
  400: "ytv.invalid_input", 401: "ytv.not_authenticated", 403: "ytv.forbidden", 404: "ytv.not_found",
  405: "ytv.method_not_allowed", 413: "ytv.request_body_too_large", 415: "ytv.unsupported_media_type",
  502: "ytv.gateway_unavailable", 503: "ytv.source_unavailable", 504: "ytv.gateway_timeout",
};

export function ytvBffProblem(cause: unknown, traceId: string = randomUUID()) {
  const raw = Number((cause as { statusCode?: unknown })?.statusCode ?? 500);
  const code = codes[raw] ?? (raw === 429 ? "common.rate_limited" : "common.internal");
  const status = catalog.errors.find(entry => entry.code === code)?.status ?? (raw === 429 ? 429 : 500);
  return { type: `https://errors.yueli.dev/problems/${code}`, status, code, params: {}, violations: [], traceId };
}

export function withYtvBffErrors(handler: ReturnType<typeof defineEventHandler>) {
  return defineEventHandler(async event => {
    try { return await handler(event); }
    catch (error) {
      const failure = ytvBffProblem(error);
      setResponseStatus(event, failure.status);
      setResponseHeader(event, "content-type", "application/problem+json");
      setResponseHeader(event, "x-trace-id", failure.traceId);
      setResponseHeader(event, "cache-control", "no-store");
      return failure;
    }
  });
}
