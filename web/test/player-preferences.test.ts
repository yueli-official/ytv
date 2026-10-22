import { expect, it } from 'vitest';
import { bindPlayerPreferences } from '../app/utils/player-preferences';
it('restores existing player preferences, preserves extra fields and detaches listeners', () => {
 const media = Object.assign(new EventTarget(), {volume:1,muted:false,playbackRate:1});
 let record = JSON.stringify({volume:0.4,muted:true,speed:1.5,quality:720});
 const detach = bindPlayerPreferences(media as unknown as HTMLMediaElement, {getItem:()=>record,setItem:(_,value)=>{record=value;}});
 media.dispatchEvent(new Event('loadedmetadata'));
 expect([media.volume,media.muted,media.playbackRate]).toEqual([0.4,true,1.5]);
 media.playbackRate=2;media.dispatchEvent(new Event('ratechange'));
 expect(JSON.parse(record)).toEqual({volume:0.4,muted:true,speed:2,quality:720});
 detach();media.volume=0.7;media.dispatchEvent(new Event('volumechange'));
 expect(JSON.parse(record).volume).toBe(0.4);
});
