export interface PlayerSettings {
  brightness: number;
  ratio: 'auto' | '4:3' | '16:9';
  fill: boolean;
  mirror: boolean;
  lightsOff: boolean;
  endAction: 'pause' | 'next' | 'loop';
}
export const playerSettingsKey = 'ytv:player-settings';
export const defaultPlayerSettings = (): PlayerSettings => ({ brightness: 1, ratio: 'auto', fill: false, mirror: false, lightsOff: false, endAction: 'pause' });
export function readPlayerSettings(storage: Pick<Storage, 'getItem'>): PlayerSettings {
  const settings = defaultPlayerSettings();
  const saved = JSON.parse(storage.getItem(playerSettingsKey) || 'null');
  if (!saved || typeof saved !== 'object') return settings;
  if (Number.isFinite(saved.brightness)) settings.brightness = Math.max(0.2, Math.min(2, saved.brightness));
  if (['auto', '4:3', '16:9'].includes(saved.ratio)) settings.ratio = saved.ratio;
  if (['pause', 'next', 'loop'].includes(saved.endAction)) settings.endAction = saved.endAction;
  for (const key of ['fill', 'mirror', 'lightsOff'] as const) if (typeof saved[key] === 'boolean') settings[key] = saved[key];
  return settings;
}
/** The settings panel and touch gestures share this picture state. */
export function bindPlayerPicture(media: HTMLElement & { loop: boolean }, surface: HTMLElement, settings: () => PlayerSettings) {
  const viewport = surface.shadowRoot?.querySelector('media-container') || surface;
  const apply = () => {
    const value = settings(), box = viewport.getBoundingClientRect();
    let width = box.width, height = box.height;
    const ratio = value.ratio === 'auto' ? 0 : value.ratio === '4:3' ? 4 / 3 : 16 / 9;
    if (ratio) {
      height = value.fill ? Math.max(box.height, box.width / ratio) : Math.min(box.height, box.width / ratio);
      width = height * ratio;
    }
    const fit = ratio ? 'fill' : value.fill ? 'cover' : 'contain';
    Object.assign(media.style, { position: 'absolute', left: '50%', top: '50%', width: `${width}px`, height: `${height}px`,
      transform: `translate(-50%, -50%) scaleX(${value.mirror ? -1 : 1})`, objectFit: fit,
      filter: value.brightness === 1 ? '' : `brightness(${value.brightness})` });
    media.style.setProperty('--media-object-fit', fit);
    media.loop = value.endAction === 'loop';
  };
  const observer = new ResizeObserver(apply);
  observer.observe(viewport);apply();
  return { apply, dispose: () => observer.disconnect() };
}
