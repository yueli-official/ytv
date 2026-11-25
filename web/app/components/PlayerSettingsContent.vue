<script setup lang="ts">
import type { PlayerSettings } from '~/utils/player-picture';
import type { AudioBalanceMode } from '~/utils/audio-balance';
const settings = defineModel<PlayerSettings>({ required: true });
defineProps<{ volume: number; volumeSupported: boolean; audioMode: AudioBalanceMode; audioAvailable: boolean; audioBusy: boolean; audioError: string; storageError: string; pipAvailable: boolean }>();
const emit = defineEmits<{ volume: [value: number]; audio: [mode: AudioBalanceMode]; pip: [] }>();
const ends = [{ value: 'next', label: '自动下一集' }, { value: 'pause', label: '暂停' }, { value: 'loop', label: '单集循环' }] as const;
const ratios = [{ value: 'auto', label: '自动' }, { value: '4:3', label: '4:3' }, { value: '16:9', label: '16:9' }] as const;
const audioModes = [{ value: 'off', label: '关闭' }, { value: 'standard', label: '标准' }, { value: 'dynamic', label: '高动态' }] as const;
function resetPicture() { Object.assign(settings.value, { brightness: 1, ratio: 'auto', fill: false, mirror: false }); }
</script>

<template>
  <div class="player-settings-content">
    <fieldset><legend>画面亮度 <output>{{ Math.round(settings.brightness * 100) }}%</output></legend><input aria-label="画面亮度" type="range" min="20" max="200" step="1" :value="settings.brightness * 100" @input="settings.brightness = Number(($event.target as HTMLInputElement).value) / 100" /></fieldset>
    <fieldset><legend>音量 <output>{{ Math.round(volume * 100) }}%</output></legend><input aria-label="视频音量" type="range" min="0" max="100" step="1" :value="volume * 100" @input="emit('volume', Number(($event.target as HTMLInputElement).value) / 100)" /><p v-if="!volumeSupported">请使用设备音量键。</p></fieldset>
    <fieldset><legend>视频比例</legend><div class="player-setting-options"><label v-for="option in ratios" :key="option.value"><input v-model="settings.ratio" type="radio" name="picture-ratio" :value="option.value" /><span>{{ option.label }}</span></label></div></fieldset>
    <fieldset><legend>播完后</legend><div class="player-setting-options"><label v-for="option in ends" :key="option.value"><input v-model="settings.endAction" type="radio" name="end-action" :value="option.value" /><span>{{ option.label }}</span></label></div></fieldset>
    <fieldset :aria-busy="audioBusy"><legend>音量均衡</legend><div class="player-setting-options"><label v-for="option in audioModes" :key="option.value"><input type="radio" name="audio-balance" :value="option.value" :checked="audioMode === option.value" :disabled="audioBusy || (option.value !== 'off' && !audioAvailable)" @change="emit('audio', option.value)" /><span>{{ option.label }}</span></label></div><p>{{ audioAvailable ? '仅当前视频；高动态保留更多声音强弱变化。' : '当前线路不支持音量均衡。' }}</p><p v-if="audioError" role="status">{{ audioError }}</p></fieldset>
    <fieldset class="player-setting-checks"><legend>其他设置</legend><label><input v-model="settings.fill" type="checkbox" />填满画面</label><label><input v-model="settings.mirror" type="checkbox" />水平镜像</label><label><input v-model="settings.lightsOff" type="checkbox" />关灯模式</label></fieldset>
    <div class="player-settings-actions"><UButton label="还原画面" size="xs" variant="outline" color="neutral" @click="resetPicture" /><UButton v-if="pipAvailable" label="画中画" icon="i-tabler-picture-in-picture" size="xs" variant="outline" color="neutral" @click="emit('pip')" /></div>
    <p v-if="storageError" role="status">{{ storageError }}</p>
  </div>
</template>

<style scoped>
.player-settings-content{display:grid;gap:18px;color:#fff;font-size:13px}
fieldset{min-width:0;border:0;padding:0;margin:0}legend{width:100%;margin-bottom:9px;font-weight:600}output{float:right;font-variant-numeric:tabular-nums;color:var(--ui-primary)}
input[type=range]{display:block;width:100%;accent-color:var(--ui-primary);min-height:24px;cursor:pointer;touch-action:pan-y}
.player-setting-options{display:flex;gap:6px}.player-setting-options label{position:relative;flex:1;text-align:center}.player-setting-options input{position:absolute;opacity:0;width:1px;height:1px}.player-setting-options span{display:block;padding:7px 3px;border:1px solid #ffffff25;border-radius:6px;background:#ffffff08;cursor:pointer;white-space:nowrap}
.player-setting-options input:checked+span{background:var(--ui-primary);border-color:var(--ui-primary);color:#10212e}.player-setting-options input:focus-visible+span{outline:2px solid white;outline-offset:2px}.player-setting-options input:disabled+span{opacity:.4;cursor:default}
.player-setting-checks{display:grid;gap:10px}.player-setting-checks label{display:flex;align-items:center;gap:8px}.player-setting-checks input{accent-color:var(--ui-primary)}
p{margin:7px 0 0;color:#ffffffa6;font-size:11px;line-height:1.65}.player-settings-actions{display:flex;gap:8px;flex-wrap:wrap}
</style>
