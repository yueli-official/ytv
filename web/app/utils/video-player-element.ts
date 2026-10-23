import { createPlayer, orientationLockFeature } from "@videojs/html";
import { videoFeatures } from "@videojs/html/video";

// Keep the packaged controls and native iOS fullscreen fallback. The standard
// video preset does not include orientation locking.
const { PlayerElement } = createPlayer({ features: [...videoFeatures, orientationLockFeature] });
export type YtvVideoPlayerElement = InstanceType<typeof PlayerElement>;

if (!customElements.get("ytv-video-player")) customElements.define("ytv-video-player", PlayerElement);
