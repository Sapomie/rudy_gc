import { afterEach, beforeEach, describe, expect, test, vi } from "vitest";
import { useFeedbackToast } from "./useFeedbackToast";

describe("useFeedbackToast", () => {
  beforeEach(() => {
    vi.useFakeTimers();
    useFeedbackToast().clear();
  });

  afterEach(() => {
    useFeedbackToast().clear();
    vi.useRealTimers();
  });

  test("stacks notifications and dismisses one without changing the others", () => {
    const feedback = useFeedbackToast();
    const firstID = feedback.success("保存成功");
    feedback.error("保存失败");

    expect(feedback.toasts.value.map((toast) => [toast.message, toast.tone])).toEqual([
      ["保存成功", "success"],
      ["保存失败", "error"]
    ]);

    feedback.dismiss(firstID!);
    expect(feedback.toasts.value.map((toast) => toast.message)).toEqual(["保存失败"]);
  });

  test("automatically removes notifications", () => {
    const feedback = useFeedbackToast();
    feedback.success("操作完成", { label: "查看任务", href: "/tasks/1" });

    expect(feedback.toasts.value[0]?.action?.href).toBe("/tasks/1");
    vi.runAllTimers();
    expect(feedback.toasts.value).toEqual([]);
  });
});
