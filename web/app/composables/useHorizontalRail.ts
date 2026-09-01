export function useHorizontalRail() {
  const rail = ref<HTMLElement | null>(null);
  const canPrevious = ref(false);
  const canNext = ref(false);

  function update() {
    const element = rail.value;
    if (!element) {
      canPrevious.value = false;
      canNext.value = false;
      return;
    }
    canPrevious.value = element.scrollLeft > 2;
    canNext.value = element.scrollLeft + element.clientWidth < element.scrollWidth - 2;
  }

  function move(direction: -1 | 1) {
    const element = rail.value;
    if (!element) return;
    element.scrollBy({
      left: direction * Math.max(240, element.clientWidth - 48),
      behavior: window.matchMedia("(prefers-reduced-motion: reduce)").matches ? "auto" : "smooth"
    });
  }

  watch(rail, (element, _previous, onCleanup) => {
    if (!element) return;
    const resize = new ResizeObserver(update);
    const mutation = new MutationObserver(update);
    resize.observe(element);
    mutation.observe(element, { childList: true });
    element.addEventListener("scroll", update, { passive: true });
    void nextTick(update);
    onCleanup(() => {
      resize.disconnect();
      mutation.disconnect();
      element.removeEventListener("scroll", update);
    });
  }, { flush: "post" });

  return {
    rail,
    canPrevious: readonly(canPrevious),
    canNext: readonly(canNext),
    previous: () => move(-1),
    next: () => move(1)
  };
}
