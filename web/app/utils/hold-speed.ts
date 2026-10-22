interface HoldMedia extends EventTarget {
  paused: boolean;
  ended: boolean;
  playbackRate: number;
}

/** A temporary speed gesture; tap, double-tap and keyboard actions stay with the player. */
export function bindHoldSpeed(surface: HTMLElement, media: HoldMedia, onChange: (active: boolean) => void) {
  const events = new AbortController();
  const capture = { capture: true, signal: events.signal };
  let pointer: { id: number; x: number; y: number } | null = null;
  let timer: ReturnType<typeof setTimeout> | undefined;
  let originalRate: number | null = null;
  const cancel = () => {
    clearTimeout(timer);
    timer = undefined;
    pointer = null;
    if (originalRate !== null) {
      const rate = originalRate;
      originalRate = null;
      media.playbackRate = rate;
      onChange(false);
    }
  };
  const interactive = (event: Event) => event.composedPath().some((node) => node instanceof Element && node.matches('button, a, input, select, textarea, [role="button"], [role="slider"], [role="menu"], [role="dialog"], [contenteditable="true"]'));
  surface.addEventListener('pointerdown', (event) => {
    if (pointer) { cancel(); return; }
    if (!event.isPrimary || event.button !== 0 || media.paused || media.ended || interactive(event)) return;
    pointer = { id: event.pointerId, x: event.clientX, y: event.clientY };
    timer = setTimeout(() => {
      if (!pointer || media.paused || media.ended) return;
      originalRate = media.playbackRate;
      media.playbackRate = 2;
      onChange(true);
    }, 400);
  }, capture);
  window.addEventListener('pointermove', (event) => {
    if (pointer && event.pointerId === pointer.id && Math.hypot(event.clientX - pointer.x, event.clientY - pointer.y) > 14) cancel();
  }, capture);
  const release = (event: PointerEvent) => { if (pointer?.id === event.pointerId) cancel(); };
  window.addEventListener('pointerup', release, capture);
  window.addEventListener('pointercancel', release, capture);
  surface.addEventListener('pointerleave', cancel, { signal: events.signal });
  surface.addEventListener('lostpointercapture', cancel, { signal: events.signal });
  surface.addEventListener('contextmenu', (event) => { if (pointer || originalRate !== null) event.preventDefault(); }, capture);
  window.addEventListener('blur', cancel, { signal: events.signal });
  document.addEventListener('visibilitychange', () => { if (document.hidden) cancel(); }, { signal: events.signal });
  for (const name of ['pause', 'ended', 'error', 'emptied']) media.addEventListener(name, cancel, { signal: events.signal });
  return {
    preferredRate: () => originalRate ?? media.playbackRate,
    dispose() { cancel(); events.abort(); }
  };
}
