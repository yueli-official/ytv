/// <reference types="@webgpu/types" />
import { getMediaElement } from '@videojs/media/dom';

export type EnhancementState = 'waiting' | 'active' | 'unavailable' | 'off';
export const enhancementPreferenceKey = 'ytv:video-enhancement';

// Own the renderer lifecycle: the package's demo renderer has no disposal API.
export function bindVideoEnhancement(media: HTMLElement, host: HTMLElement, report: (state: EnhancementState) => void) {
  let enabled = false;
  let disposed = false;
  let generation = 0;
  let stop = () => {};
  const restart = async () => {
    const current = ++generation;
    stop();
    stop = () => {};
    if (disposed || !enabled) { report('off'); return; }
    if (!navigator.gpu) { report('unavailable'); return; }
    report('waiting');
    const video = getMediaElement(media);
    if (!(video instanceof HTMLVideoElement) || video.readyState < 2) return;
    if (document.pictureInPictureElement === video) { report('unavailable'); return; }
    let device: GPUDevice | undefined;
    let context: GPUCanvasContext | null = null;
    let frame = 0;
    const canvas = document.createElement('canvas');
    canvas.className = 'ytv-enhanced-video';
    canvas.setAttribute('aria-hidden', 'true');
    canvas.style.cssText = 'position:absolute;inset:0;width:100%;height:100%;object-fit:contain;pointer-events:none;visibility:hidden';
    const cleanup = () => {
      video.cancelVideoFrameCallback(frame);
      canvas.remove();
      context?.unconfigure();
      device?.destroy();
    };
    stop = cleanup;
    const valid = () => !disposed && enabled && current === generation;
    const fail = () => { if (valid()) { ++generation; cleanup(); report('unavailable'); } };
    try {
      if (!navigator.gpu || !video.requestVideoFrameCallback) throw new Error('GPU unavailable');
      const adapter = await navigator.gpu.requestAdapter();
      if (!valid()) return;
      if (!adapter) throw new Error('GPU unavailable');
      device = await adapter.requestDevice();
      if (!valid()) { cleanup(); return; }
      device.addEventListener('uncapturederror', fail);
      void device.lost.then(fail);
      const { CNNx2M } = await import('anime4k-webgpu');
      if (!valid()) { cleanup(); return; }
      const width = video.videoWidth, height = video.videoHeight;
      if (!width || !height || width * 2 > device.limits.maxTextureDimension2D || height * 2 > device.limits.maxTextureDimension2D) throw new Error('Unsupported size');
      canvas.width = width * 2;
      canvas.height = height * 2;
      context = canvas.getContext('webgpu');
      if (!context) throw new Error('GPU canvas unavailable');
      const format = navigator.gpu.getPreferredCanvasFormat();
      context.configure({ device, format, alphaMode: 'opaque' });
      const input = device.createTexture({ size: [width, height], format: 'rgba16float', usage: GPUTextureUsage.TEXTURE_BINDING | GPUTextureUsage.COPY_DST | GPUTextureUsage.RENDER_ATTACHMENT });
      const enhance = new CNNx2M({ device, inputTexture: input });
      const module = device.createShaderModule({ code: `
        struct Vertex { @builtin(position) position: vec4f, @location(0) uv: vec2f }
        @vertex fn vertex(@builtin(vertex_index) i: u32) -> Vertex {
          let uv = vec2f(f32((i << 1u) & 2u), f32(i & 2u));
          return Vertex(vec4f(uv * vec2f(2., -2.) + vec2f(-1., 1.), 0., 1.), uv);
        }
        @group(0) @binding(0) var image: texture_2d<f32>;
        @fragment fn fragment(v: Vertex) -> @location(0) vec4f {
          let size = vec2f(textureDimensions(image));
          return textureLoad(image, vec2i(clamp(v.uv * size, vec2f(0.), size - 1.)), 0);
        }` });
      const pipeline = device.createRenderPipeline({ layout: 'auto', vertex: { module, entryPoint: 'vertex' }, fragment: { module, entryPoint: 'fragment', targets: [{ format }] }, primitive: { topology: 'triangle-list' } });
      const group = device.createBindGroup({ layout: pipeline.getBindGroupLayout(0), entries: [{ binding: 0, resource: enhance.getOutputTexture().createView() }] });
      host.append(canvas);
      let first = true;
      const draw = () => {
        if (!valid()) return;
        try {
          device!.queue.copyExternalImageToTexture({ source: video }, { texture: input }, [width, height]);
          const encoder = device!.createCommandEncoder();
          enhance.pass(encoder);
          const pass = encoder.beginRenderPass({ colorAttachments: [{ view: context!.getCurrentTexture().createView(), loadOp: 'clear', storeOp: 'store', clearValue: { r: 0, g: 0, b: 0, a: 1 } }] });
          pass.setPipeline(pipeline); pass.setBindGroup(0, group); pass.draw(3); pass.end();
          device!.queue.submit([encoder.finish()]);
          // Keep at most one GPU frame in flight on slower/mobile devices.
          void device!.queue.onSubmittedWorkDone().then(() => {
            if (!valid()) return;
            if (first) { first = false; canvas.style.visibility = 'visible'; report('active'); }
            frame = video.requestVideoFrameCallback(draw);
          }, fail);
        } catch { fail(); }
      };
      draw();
    } catch { fail(); }
  };
  const onReady = () => { void restart(); };
  media.addEventListener('loadeddata', onReady);
  media.addEventListener('resize', onReady);
  media.addEventListener('enterpictureinpicture', onReady);
  media.addEventListener('leavepictureinpicture', onReady);
  return {
    setEnabled(value: boolean) { enabled = value; void restart(); },
    dispose() { disposed = true; ++generation; stop(); media.removeEventListener('loadeddata', onReady); media.removeEventListener('resize', onReady); media.removeEventListener('enterpictureinpicture', onReady); media.removeEventListener('leavepictureinpicture', onReady); },
  };
}
