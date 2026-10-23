import { resolveFailureFeedback } from "@yueli/http-runtime";
import { ytvFailurePresentation, type YtvFailureCode } from "../generated/ytvFailure";

const messages: Record<YtvFailureCode, string> = {
  "ytv.not_authenticated": "请先登录后继续。",
  "ytv.forbidden": "当前账户没有执行此操作的权限。",
  "ytv.not_found": "内容或来源不存在，请刷新后核对。",
  "ytv.conflict": "内容状态已变化，请刷新后重试。",
  "ytv.invalid_input": "填写的信息不符合要求，请检查后重试。",
  "ytv.source_unavailable": "来源暂时不可用，请稍后重试或选择其他来源。",
  "ytv.source_verification_required": "此来源要求完成验证，本站暂不能代为完成，请选择其他来源。",
  "ytv.playback_unsafe": "此来源返回的播放地址不受支持，请选择其他来源。",
  "ytv.revision_conflict": "设置已被更新，请刷新后再修改。",
  "ytv.source_already_installed": "该来源已添加，请刷新列表。",
  "ytv.source_not_installed": "该来源已移除，请刷新列表。",
  "ytv.gateway_unavailable": "服务暂时不可用，请稍后重试。",
  "ytv.gateway_timeout": "请求超时，请核对操作结果后再试。",
  "ytv.method_not_allowed": "请求方式不受支持，请刷新页面。",
  "ytv.request_body_too_large": "提交内容过大，请减少内容后重试。",
  "ytv.unsupported_media_type": "请求格式不受支持，请刷新页面。",
};

export function ytvFailureFeedback(reason: unknown, fallback = "请求未完成，请稍后重试。", fields?: Readonly<Record<string, string>>) {
  return resolveFailureFeedback(reason, { fallback, fields, resolveText(code) {
    if (Object.hasOwn(ytvFailurePresentation, code)) return { message: messages[code as YtvFailureCode] };
    if (code === "common.validation_failed") return { message: "请检查填写内容后重试。" };
    if (code === "common.rate_limited") return { message: "操作过于频繁，请稍后再试。" };
    if (code.startsWith("validation.")) return { message: code === "validation.required" ? "请填写此项。" : "此项内容不符合要求，请检查后重试。" };
  } });
}

export function ytvFailureMessage(reason: unknown, fallback?: string) {
  const feedback = ytvFailureFeedback(reason, fallback);
  return [feedback.message, ...feedback.summary].join(" ");
}
