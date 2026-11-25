<script setup lang="ts">
interface HelpAction { label: string; content: string; keys?: string[] }
const { mobile } = usePublicLayout();
const touch = {
  key: 'touch', title: '手机手势', icon: 'i-tabler-hand-finger', intro: '',
  actions: [
    { label: '双击画面', content: '暂停或继续播放。' },
    { label: '左右拖动', content: '预览进度，松开后跳转。' },
    { label: '左侧上下', content: '调整画面亮度，不改变设备背光。' },
    { label: '右侧上下', content: '调整视频音量；无变化时使用设备音量键。' },
    { label: '全屏与返回', content: '点“全屏观看”放大画面，左上角返回退出；未自动横屏时可横放手机。' }
  ] as HelpAction[]
};
const keyboard = {
  key: 'keyboard', title: '电脑操作', icon: 'i-tabler-keyboard', intro: '先点击播放器，再使用快捷键。',
  actions: [
    { label: '双击画面', content: '左侧快退 10 秒，右侧快进 10 秒，中央切换全屏。' },
    { label: '前后跳转', keys: ['←', '→'], content: '快退或快进 10 秒。' },
    { label: '播放暂停', keys: ['空格'], content: '播放或暂停。' },
    { label: '全屏', keys: ['F'], content: '进入或退出全屏。' },
    { label: '静音', keys: ['M'], content: '静音或恢复声音。' }
  ] as HelpAction[]
};
const sections = computed(() => mobile.value ? [touch, keyboard] : [keyboard, touch]);
const questions = [
  { value: 'unavailable', label: '线路无法播放怎么办？', content: '出现播放错误后，点“重新连接”重试；仍失败时切换播放线路，或点“换源”选择其他来源。' },
  { value: 'buffering', label: '播放一直卡顿怎么办？', content: '先检查网络，并尝试恢复 1 倍速播放；仍卡顿时，可以换一条线路或来源比较。' }
];
</script>

<template>
  <div class="playback-help">
    <div class="help-gestures">
      <p><UIcon name="i-tabler-player-track-next" class="size-4 shrink-0" /><span><strong>按住加速</strong>：播放时按住画面临时 2 倍速，松开恢复原倍速；触屏和鼠标都可使用。</span></p>
    </div>
    <div class="help-device-grid">
      <section v-for="section in sections" :key="section.key" :aria-labelledby="`help-${section.key}`">
        <h3 :id="`help-${section.key}`"><UIcon :name="section.icon" class="size-4" />{{ section.title }}</h3>
        <p v-if="section.intro" class="help-intro">{{ section.intro }}</p>
        <dl class="help-actions"><div v-for="action in section.actions" :key="action.label"><dt><template v-if="action.keys"><kbd v-for="key in action.keys" :key="key">{{ key }}</kbd><span class="sr-only">{{ action.label }}</span></template><template v-else>{{ action.label }}</template></dt><dd>{{ action.content }}</dd></div></dl>
      </section>
    </div>
    <section class="help-settings" aria-labelledby="help-settings-title">
      <h3 id="help-settings-title"><UIcon name="i-tabler-settings" class="size-4" />选集与设置</h3>
      <dl class="help-actions">
        <div><dt>选集</dt><dd>多集影片可在播放器内切集，全屏时无需退出画面。</dd></div>
        <div><dt>齿轮设置</dt><dd>选择播放倍速；片源提供可选字幕和音轨时，也能在这里切换。</dd></div>
        <div><dt>画面与播放</dt><dd>调整亮度、比例、镜像和关灯模式，也可设置播完暂停、自动下一集或单集循环。填满画面可能裁切边缘。</dd></div>
        <div><dt>声音与画中画</dt><dd>画中画需要浏览器支持；音量均衡取决于线路，切集后需重新开启。</dd></div>
      </dl>
    </section>
    <section class="help-questions" aria-labelledby="help-questions-title">
      <h3 id="help-questions-title"><UIcon name="i-tabler-help-circle" class="size-4" />常见播放问题</h3>
      <UAccordion :items="questions" type="multiple" trailing-icon="i-tabler-chevron-down" :ui="{ item: 'border-white/8', trigger: 'gap-3 py-3 text-sm font-medium text-[#dce2ea]', body: 'pb-4 text-sm leading-7 text-[#9aa5b5]', trailingIcon: 'text-[#9aa5b5]' }" />
    </section>
  </div>
</template>

<style scoped>
.playback-help{display:grid;gap:24px;color:var(--ui-text-muted);font-size:14px;line-height:1.8}
.playback-help h3{display:flex;align-items:center;gap:8px;margin:0 0 12px;color:var(--ui-text-highlighted);font-size:14px;font-weight:600}
.playback-help h3>.iconify{color:var(--ui-primary)}
.help-gestures{display:grid;gap:10px;padding:14px 16px;border-radius:8px;background:#ffffff04}
.help-gestures p{display:flex;align-items:baseline;gap:8px;margin:0}.help-gestures .iconify{align-self:flex-start;margin-top:4px;color:var(--ui-primary)}
.help-gestures strong{color:var(--ui-text);font-weight:500}
.help-device-grid{display:grid;grid-template-columns:repeat(2,minmax(0,1fr));gap:24px 32px}
.help-intro{margin:-4px 0 12px;font-size:12px;color:var(--ui-text-muted)}
.help-actions{display:grid;gap:12px;margin:0}.help-actions>div{display:grid;grid-template-columns:100px minmax(0,1fr);gap:12px;align-items:baseline}
.help-actions dt{color:var(--ui-text);font-size:13px;font-weight:500}.help-actions dd{margin:0;overflow-wrap:anywhere}
kbd{display:inline-block;min-width:24px;margin-right:4px;padding:0 6px;border:1px solid var(--ytv-line);border-bottom-width:2px;border-radius:5px;background:#ffffff04;color:var(--ui-text);font:12px/24px var(--font-sans);text-align:center}
.help-settings,.help-questions{padding-top:20px;border-top:1px solid var(--ytv-line)}
.help-questions h3{margin-bottom:4px}
@media(max-width:767px){.playback-help{gap:20px;font-size:13px}.help-device-grid{grid-template-columns:1fr;gap:20px}.help-actions>div{grid-template-columns:84px minmax(0,1fr);gap:10px}.help-gestures{padding:12px}.help-settings,.help-questions{padding-top:18px}}
</style>
