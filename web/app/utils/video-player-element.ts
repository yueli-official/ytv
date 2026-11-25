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
    // Use the SDK recognizer so a double tap does not also trigger a single tap.
    // Mouse regions keep their packaged seek/fullscreen actions.
    for (const gesture of template.content.querySelectorAll('media-gesture[type="doubletap"]')) gesture.setAttribute('pointer', 'mouse');
    const touchPause = document.createElement('media-gesture');
    touchPause.setAttribute('type', 'doubletap');
    touchPause.setAttribute('action', 'togglePaused');
    touchPause.setAttribute('pointer', 'touch');
    container.append(touchPause);
    const enhancementSlot = document.createElement('slot');
    enhancementSlot.name = 'enhancement';
    template.content.querySelector('.video-controls-end')!.prepend(enhancementSlot);
    const settingsSlot = document.createElement('slot');
    settingsSlot.name = 'picture-settings';
    const settingsMenu = template.content.querySelector('media-menu')!;
    settingsMenu.classList.add('ytv-settings-menu');
    settingsMenu.querySelector('media-menu-content')!.prepend(settingsSlot);
    const popupLayout = document.createElement('style');
    popupLayout.textContent = '.ytv-settings-menu{min-width:220px;max-width:calc(100vw - 24px)}.media-container:fullscreen .ytv-settings-menu{position:fixed!important;inset:auto 12px 60px auto!important;max-height:calc(100dvh - 80px);overflow-y:auto}';
    template.content.append(popupLayout);
    const episodeSlot = document.createElement('slot');
    episodeSlot.name = 'episodes-control';
    template.content.querySelector('.video-controls-end')!.prepend(episodeSlot);
    return template;
  })();
}
if (!customElements.get("ytv-video-skin")) customElements.define("ytv-video-skin", YtvVideoSkinElement);
