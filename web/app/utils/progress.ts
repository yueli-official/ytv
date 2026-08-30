export interface PlaybackProgressRecord {
  position: number;
  duration: number;
  updatedAt: number;
  completed: boolean;
}

export function completedPlayback(position: number, duration: number) {
  if (!Number.isFinite(position) || !Number.isFinite(duration) || duration <= 0) return false;
  return duration - position <= 30 || position / duration >= 0.9;
}

export function restorablePosition(record: PlaybackProgressRecord | null, duration: number, seekable: TimeRanges | null) {
  if (!record || record.completed || record.position <= 5 || !Number.isFinite(duration) || duration <= 0 || !seekable?.length) return 0;
  const target = Math.min(record.position, Math.max(0, duration - 1));
  for (let index = 0; index < seekable.length; index += 1) {
    const start = seekable.start(index);
    const end = seekable.end(index);
    if (target >= start && target <= end) return target;
  }
  return 0;
}
