import { createPlayer, orientationLockFeature } from "@videojs/html";
import { NeutralVideoSkinElement, videoFeatures } from "@videojs/html/video";

// Keep the packaged controls and native iOS fullscreen fallback. The standard
// video preset does not include orientation locking.
const { PlayerElement } = createPlayer({ features: [...videoFeatures, orientationLockFeature] });
export type YtvVideoPlayerElement = InstanceType<typeof PlayerElement>;

if (!customElements.get("ytv-video-player")) customElements.define("ytv-video-player", PlayerElement);

// Customize the packaged skin through its public template extension point.
// Keep the video unchanged when controls appear. Set the token on the skin
// container because its theme overrides values inherited from the outer host.
class YtvVideoSkinElement extends NeutralVideoSkinElement {
  static override template = (() => {
    const template = NeutralVideoSkinElement.template!.cloneNode(true) as HTMLTemplateElement;
    const container = template.content.firstElementChild as HTMLElement;
    container.style.setProperty("--media-controls-gradient", "none");
    return template;
  })();
}
if (!customElements.get("ytv-video-skin")) customElements.define("ytv-video-skin", YtvVideoSkinElement);
