/** Keep the established Plyr preference record so existing viewers retain their settings. */
export function bindPlayerPreferences(media: HTMLMediaElement, storage: Pick<Storage, 'getItem' | 'setItem'>, preferredRate = () => media.playbackRate) {
  let saved: Record<string, unknown> = {};
  let applying = false;
  const restore = () => {
    applying = true;
    try {
      const parsed = JSON.parse(storage.getItem('plyr') || '{}');
      saved = parsed && typeof parsed === 'object' && !Array.isArray(parsed) ? parsed : {};
      if (typeof saved.volume === 'number' && saved.volume >= 0 && saved.volume <= 1) media.volume = saved.volume;
      if (typeof saved.muted === 'boolean') media.muted = saved.muted;
      if (typeof saved.speed === 'number' && saved.speed >= 0.25 && saved.speed <= 4) media.playbackRate = saved.speed;
    } catch { /* Browser preferences are optional. */ }
    finally { applying = false; }
  };
  const save = () => {
    if (applying) return;
    try { storage.setItem('plyr', JSON.stringify({ ...saved, volume: media.volume, muted: media.muted, speed: preferredRate() })); } catch { /* Full or disabled storage must not interrupt playback. */ }
  };
  media.addEventListener('loadedmetadata', restore, { once: true });
  media.addEventListener('volumechange', save);
  media.addEventListener('ratechange', save);
  return () => {
    media.removeEventListener('loadedmetadata', restore);
    media.removeEventListener('volumechange', save);
    media.removeEventListener('ratechange', save);
  };
}
