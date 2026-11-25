interface GestureMedia extends EventTarget {
  paused: boolean;
  ended: boolean;
  playbackRate: number;
  currentTime: number;
  duration: number;
  volume: number;
  muted: boolean;
  style: CSSStyleDeclaration;
}

export type PlayerGestureFeedback =
  | { kind: 'speed' }
  | { kind: 'seek'; target: number; duration: number; offset: number }
  | { kind: 'brightness'; value: number }
  | { kind: 'volume'; value: number; supported: boolean };

/** One direction lock owns touch scrubbing, brightness, volume and temporary speed. */
export function bindPlayerGestures(surface: HTMLElement, media: GestureMedia, onChange: (feedback: PlayerGestureFeedback | null) => void, picture: { brightness(): number; setBrightness(value: number): void }) {
  const events = new AbortController();
  const capture = { capture: true, signal: events.signal };
  let originalRate: number | null = null;
  let timer: ReturnType<typeof setTimeout> | undefined;
  let feedbackTimer: ReturnType<typeof setTimeout> | undefined;
  let ignoredClick: { x: number; y: number; until: number } | null = null;
  const consumedPointers = new Set<number>();
  const dragThreshold = 24;
  const beyondThreshold = (distance: number) => Math.sign(distance) * Math.max(0, Math.abs(distance) - dragThreshold);
  let pointer: {
    id: number; touch: boolean; x: number; y: number; width: number; height: number;
    mode: 'pending' | 'speed' | 'seek' | 'brightness' | 'volume'; left: boolean;
    time: number; duration: number; target: number; brightness: number; volume: number;
  } | null = null;
  const clamp = (value: number, min: number, max: number) => Math.min(max, Math.max(min, value));
  // SDK custom controls expose some roles through ElementInternals, so role
  // attribute selectors alone miss their slider hit areas and play buttons.
  const interactive = (event: Event) => event.composedPath().some(node => node instanceof Element && node.matches('button, a, input, select, textarea, media-controls, media-play-button, media-menu, media-error-dialog, [role="button"], [role="slider"], [role="menu"], [role="menuitem"], [role="dialog"], [contenteditable="true"]'));
  const restoreRate = () => {
    if (originalRate === null) return;
    const rate = originalRate;
    originalRate = null;
    media.playbackRate = rate;
  };
  const capturePointer = () => {
    if (!pointer) return;
    consumedPointers.add(pointer.id);
    if (!surface.hasPointerCapture(pointer.id)) surface.setPointerCapture(pointer.id);
  };
  const finish = (commit = false, event?: PointerEvent) => {
    const previous = pointer;
    pointer = null;
    clearTimeout(timer);
    clearTimeout(feedbackTimer);
    restoreRate();
    if (previous) {
      if (!commit || previous.mode !== 'pending') consumedPointers.add(previous.id);
      if (previous.mode !== 'pending' && event) ignoredClick = { x: event.clientX, y: event.clientY, until: performance.now() + 700 };
      if (surface.hasPointerCapture(previous.id)) surface.releasePointerCapture(previous.id);
      if (commit && previous.mode === 'seek') media.currentTime = previous.target;
    }
    if (commit && previous && previous.mode !== 'pending' && previous.mode !== 'speed') feedbackTimer = setTimeout(() => onChange(null), 650);
    else onChange(null);
  };
  window.addEventListener('pointerdown', event => {
    if (pointer && pointer.id !== event.pointerId) { consumedPointers.add(event.pointerId); finish(false, event); }
  }, capture);
  surface.addEventListener('pointerdown', event => {
    if (pointer) return;
    if (!event.isPrimary || event.button !== 0 || interactive(event)) return;
    consumedPointers.delete(event.pointerId);
    clearTimeout(feedbackTimer);
    onChange(null);
    const box = surface.getBoundingClientRect();
    pointer = {
      id: event.pointerId, touch: event.pointerType === 'touch', x: event.clientX, y: event.clientY,
      width: Math.max(1, box.width), height: Math.max(1, box.height), mode: 'pending',
      left: event.clientX < box.left + box.width / 2,
      time: media.currentTime, duration: media.duration, target: media.currentTime,
      brightness: picture.brightness(), volume: media.muted ? 0 : media.volume
    };
    if (!media.paused && !media.ended) timer = setTimeout(() => {
      if (!pointer || pointer.mode !== 'pending' || media.paused || media.ended) return;
      pointer.mode = 'speed';
      originalRate = media.playbackRate;
      media.playbackRate = 2;
      capturePointer();
      onChange({ kind: 'speed' });
    }, 400);
  }, capture);
  window.addEventListener('pointermove', event => {
    if (!pointer || event.pointerId !== pointer.id) return;
    const dx = event.clientX - pointer.x, dy = event.clientY - pointer.y;
    if (pointer.mode === 'pending' || pointer.mode === 'speed') {
      if (Math.hypot(dx, dy) < dragThreshold) return;
      clearTimeout(timer);
      // A hold may have advanced playback for minutes before becoming a swipe.
      if (pointer.mode === 'speed') { pointer.time = media.currentTime; pointer.target = pointer.time; }
      restoreRate();
      if (!pointer.touch) { finish(false, event); return; }
      pointer.mode = 'pending';
      // Wait for a clear direction rather than guessing from a diagonal jitter.
      if (Math.abs(dx) >= dragThreshold && Math.abs(dx) >= Math.abs(dy) * 1.3) {
        if (!Number.isFinite(pointer.duration) || pointer.duration <= 0 || !Number.isFinite(pointer.time)) { finish(false, event); return; }
        pointer.mode = 'seek';
      } else if (Math.abs(dy) >= dragThreshold && Math.abs(dy) >= Math.abs(dx) * 1.3) pointer.mode = pointer.left ? 'brightness' : 'volume';
      else return;
      capturePointer();
    }
    event.preventDefault();
    if (pointer.mode === 'seek') {
      pointer.target = clamp(pointer.time + beyondThreshold(dx) / pointer.width * Math.min(60, pointer.duration), 0, pointer.duration);
      onChange({ kind: 'seek', target: pointer.target, duration: pointer.duration, offset: pointer.target - pointer.time });
    } else if (pointer.mode === 'brightness') {
      const brightness = clamp(pointer.brightness - beyondThreshold(dy) / pointer.height * 1.5, 0.2, 2);
      picture.setBrightness(brightness);
      onChange({ kind: 'brightness', value: brightness });
    } else if (pointer.mode === 'volume') {
      const volume = clamp(pointer.volume - beyondThreshold(dy) / pointer.height, 0, 1);
      media.volume = volume;
      media.muted = volume === 0;
      const supported = Math.abs(media.volume - volume) < 0.01;
      onChange({ kind: 'volume', value: media.muted ? 0 : media.volume, supported });
    }
  }, { ...capture, passive: false });
  const release = (event: PointerEvent, commit: boolean) => {
    if (pointer?.id === event.pointerId) finish(commit, event);
    if (consumedPointers.delete(event.pointerId)) { event.preventDefault(); event.stopPropagation(); }
  };
  window.addEventListener('pointerup', event => release(event, true), capture);
  window.addEventListener('pointercancel', event => release(event, false), capture);
  surface.addEventListener('pointerleave', () => { if (pointer && !surface.hasPointerCapture(pointer.id)) finish(); }, { signal: events.signal });
  // Touch starts with implicit capture on the picture/control under the finger.
  // Its loss bubbles here when we take capture; only losing our own cancels us.
  surface.addEventListener('lostpointercapture', event => { if (event.composedPath()[0] === surface && pointer?.id === event.pointerId) finish(); }, { signal: events.signal });
  for (const name of ['click', 'dblclick']) surface.addEventListener(name, event => {
    const click = event as MouseEvent;
    if (!ignoredClick || performance.now() > ignoredClick.until || click.detail === 0 || interactive(event)) return;
    if (Math.hypot(click.clientX - ignoredClick.x, click.clientY - ignoredClick.y) > 32) return;
    event.preventDefault();
    event.stopImmediatePropagation();
  }, capture);
  surface.addEventListener('contextmenu', event => { if (pointer || originalRate !== null) event.preventDefault(); }, capture);
  window.addEventListener('blur', () => finish(), { signal: events.signal });
  document.addEventListener('visibilitychange', () => { if (document.hidden) finish(); }, { signal: events.signal });
  for (const name of ['pause', 'ended', 'error', 'emptied']) media.addEventListener(name, () => finish(), { signal: events.signal });
  return {
    preferredRate: () => originalRate ?? media.playbackRate,
    dispose() { finish(); events.abort(); }
  };
}
