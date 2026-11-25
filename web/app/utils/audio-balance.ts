export type AudioBalanceMode = 'off' | 'standard' | 'dynamic';
/** Only an active hls.js MediaSource is eligible; never capture cross-origin direct media. */
export function audioBalanceTarget(media: { engine?: { media: HTMLMediaElement | null } }): HTMLMediaElement | null {
  const target = media.engine?.media;
  return typeof AudioContext !== 'undefined' && target instanceof HTMLMediaElement && target.currentSrc.startsWith('blob:') ? target : null;
}
export function createAudioBalance(target: HTMLMediaElement) {
  let graph: { context: AudioContext; source: MediaElementAudioSourceNode; compressor: DynamicsCompressorNode; gain: GainNode } | null = null;
  let starting: Promise<void> | null = null;
  let requested: AudioBalanceMode = 'off';
  let disposed = false;
  const configure = () => {
    if (!graph) return;
    const { context, source, compressor, gain } = graph;
    source.disconnect();
    if (requested === 'off') { gain.gain.value = 1; source.connect(gain); }
    else {
      compressor.threshold.value = requested === 'standard' ? -24 : -16;
      compressor.knee.value = 24;
      compressor.ratio.value = requested === 'standard' ? 4 : 2;
      compressor.attack.value = 0.003;compressor.release.value = 0.25;
      gain.gain.setTargetAtTime(requested === 'standard' ? 1.4 : 1.1, context.currentTime, 0.02);
      source.connect(compressor);
    }
  };
  return {
    async setMode(mode: AudioBalanceMode) {
      requested = mode;
      if (mode === 'off' && !graph && !starting) return;
      if (!graph && !starting) starting = (async () => {
        const context = new AudioContext();
        try {
          // Unlock before rerouting audio so a refused context never silences playback.
          await context.resume();
          if (disposed || requested === 'off') { await context.close(); return; }
          if (context.state !== 'running') throw new Error('Audio context did not start');
          const compressor = context.createDynamicsCompressor(), gain = context.createGain();
          compressor.connect(gain);gain.connect(context.destination);
          const source = context.createMediaElementSource(target);
          graph = { context, source, compressor, gain };
          configure();
        } catch (error) { if (!graph) await context.close();throw error; }
      })().finally(() => { starting = null; });
      if (starting) await starting;
      if (!disposed) configure();
    },
    dispose() { disposed = true;requested = 'off';if (graph) { graph.source.disconnect();void graph.context.close();graph = null; } }
  };
}
