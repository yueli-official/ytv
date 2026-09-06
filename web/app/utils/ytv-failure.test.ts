import { describe, expect, it } from "vitest";
import { ytvFailureFeedback, ytvFailureMessage } from "./ytv-failure";
import { ytvBffProblem } from "../../server/utils/ytv-bff";

describe("YTV structured feedback", () => {
  it("keeps field errors, unmapped errors and the request trace", () => {
    const result = ytvFailureFeedback({ data: { failure: {
      kind: "remote", code: "common.validation_failed", status: 400, params: {}, reauth: "not-attempted", traceId: "ytv-test",
      violations: [{ pointer: "/name", code: "validation.required", params: {} }, { pointer: "/unknown", code: "validation.invalid", params: {} }],
    } } }, "保存失败", { "/name": "name" });
    expect(result.fieldErrors.name).toEqual(["请填写此项。"]);
    expect(result.summary).toHaveLength(1);
    expect(result.technical.traceId).toBe("ytv-test");
  });
  it("does not expose arbitrary provider or thrown messages", () => {
    const cause = Object.assign(new Error("private-provider-marker"), { statusCode: 502 });
    expect(ytvFailureMessage(cause, "请求失败")).toBe("请求失败");
    expect(ytvBffProblem(cause, "ytv-test")).toMatchObject({ code: "ytv.gateway_unavailable", status: 502, traceId: "ytv-test" });
    expect(JSON.stringify(ytvBffProblem(cause))).not.toContain("private-provider-marker");
    expect(ytvBffProblem({ statusCode: 799 })).toMatchObject({ code: "common.internal", status: 500 });
  });
});
