# YTV 电影站升级方案

调研日期：2026-10-22（Asia/Shanghai）。状态：实施前方案快照。用户随后已授权接口、播放器与 UI 本地实现；当前交付见[统一影视与播放接口](playback-integration.md)。下文未实施描述对应调研阶段。

## 建议与目标

沿用现有 Nuxt + Go 重建版，围绕“找到电影 → 选可播放版本 → 开始观看 → 中断后恢复”升级。复用已有片源和 Kazumi 适配基础，把线路、版本、字幕、进度和播放恢复做成完整产品能力。播放器优先验证 **Video.js 10 + hls.js 适配器**，Artplayer 为备选；确定性媒体样本与真实获准来源验收后再最终选定。

统一的是 YTV 的片源、解析和播放会话接口。业务页面不直接操作某个播放器库，也不根据某个来源写特殊分支。先在 YTV 仓内形成有真实消费者的模块；出现第二个产品消费者后，再决定是否发布独立 SDK。

本方案不把当前上线旧版当成新版代码，不把调研结论当成播放验证。实现顺序是建议，尚未得到用户选择；本轮不安装依赖、不改业务代码、不操作线上环境。

## 当前基础与缺口

源码基线：`3cb7dd154bb1ea383ca200752c37cc450d51c9f3`，目录 `E:/projects/organizations/yueli-official/ytv`。工作树已有 UI 测试清理和 vendor 文件，本轮保留。线上旧版状态依据相邻 Workspace 的 [部署记录](../../../workspace/docs/deployments.md)，本轮没有访问生产验证。

| 模块 | 已有能力 | 本轮静态检查发现的升级点 |
| --- | --- | --- |
| 片源 | `source.Adapter` 已统一 Browse/Search/ResolveTitle/ResolveEpisode/Probe，支持 JSON、HTML、declarative | 扩展已有接口的表达能力，不再包一层相同方法 |
| Kazumi | compiler → `rules.Plan` → declarative adapter；固定规则目录、来源专用 extractor、Chromium capture 均有代码 | 旧调研写着“不实现 WebView”，现有代码已超出该历史结论；需要逐能力核对，不能宣称完整兼容所有规则 |
| 聚合搜索 | 已有多源查询、部分失败与展示分组 | 展示组不能作为影片唯一身份；同名、重拍、不同剪辑版本要保留区别 |
| 影片/线路 | `Title.Episodes` 是单层列表；JSON adapter 仅按 `#`/`$` 拆分；declarative 多路结果压平 | 缺少明确的线路、版本及其与播放内容的关系；多线路 MacCMS 结构需要独立样本回归 |
| 播放解析 | Resolver 返回 URL、engine、timeline；URL/host 校验已有实现 | 类型主要由 URL 后缀推断；没有短效链接的有效期、字幕、音轨和恢复条件表达 |
| 播放器 | Plyr 3.8.4 + hls.js 1.7.1；基础状态、手动清晰度、解码恢复、进度持久化 | 引擎、控件、进度与事件处理集中在一个 Vue 文件；组件重试复用原 URL；原生 media error 未统一接入 |
| 电影体验 | 当前观看页统一显示集数、选集与前后集 | 单部电影应展示“播放正片/继续观看”和可用版本、线路，不能用换播放器掩盖内容模型问题 |
| 用户数据 | 匿名本地历史、账号进度接口已有实现 | 需要验证账号历史恢复、跨来源时间轴匹配、存储失败不阻塞观看，保留现有历史标识 |

以上是源码观察，不是已复现故障。相关入口：

- [Source 模型和接口](../../api/internal/source/types.go)、[JSON adapter](../../api/internal/source/jsoncatalog/adapter.go)、[declarative adapter](../../api/internal/source/declarative/adapter.go)。
- [Kazumi compiler](../../api/internal/source/rulecompiler/kazumi.go)、[规则目录](../../api/internal/source/rulecatalog/kazumi.go)、[浏览器解析](../../api/internal/source/browsercapture/resolver.go)。
- [播放 Resolver](../../api/internal/playback/resolver.go)、[播放器](../../web/app/components/VideoPlayer.client.vue)、[DTO](../../web/app/types/ytv.ts)、[依赖基线](../../web/package.json)。

## Kazumi 借鉴范围

Kazumi 是 Flutter 客户端。适合借鉴的是“规则描述来源 → 搜索/详情提取 → 解析播放地址 → 播放”的分层，以及规则能力与播放线路的区分。它使用原生客户端和 WebView 的环境，不能等同于网页播放器。[官方仓库](https://github.com/Predidit/Kazumi)、[依赖声明](https://github.com/Predidit/Kazumi/blob/main/pubspec.yaml)、[规则仓库](https://github.com/Predidit/KazumiRules)。

YTV 继续将外部规则编译成自己的格式无关 Plan。电影网站可使用 JSON/API、静态页面规则或已有来源专用解析，业务页面消费相同结果。不要为了兼容动漫规则而让电影首页变成番剧追更页。

上游还有值得补入差距清单的新能力：Kazumi 官方文档已支持 API 规则、JSONPath 和 API/XPath 混合搜索/选集，API level 8 提供基础能力，level 9 增加 `sourceTemplate`。YTV 现有 Kazumi compiler 仍要求传统 XPath 字段，不能因为本地有 MacCMS JSON adapter 就宣称已支持这套通用 API 规则。建议在 A 阶段选真实电影 API 样本判断是否补这个格式 Adapter，而非直接导入上游全部规则。[Kazumi API 规则文档](https://kazumi.app/docs/rules/develop-api-rules/)。

浏览器捕获媒体 URL 成功只证明解析成功；用户设备还需要通过 CORS、媒体格式、Cookie 和请求条件检验。原生客户端可设置的请求条件不一定能由网页表达；不能承诺更换 JavaScript 播放器就解决这些问题。hls.js 官方也明确要求相关 HLS 资源可跨域访问。[hls.js 官方说明](https://github.com/video-dev/hls.js)。

沿用项目已有合同：部署拥有片源定义和解析行为，管理后台维护安装、启停、名称、排序和检测；浏览器播放媒体，YTV 不新增第三方视频代理、下载、转码或持久化。这是现有实现边界，不是本次新增审批流程。

历史 [Kazumi 调研](kazumi-integration.md) 保留为历史证据；其中的实现清单、否定性结论与规则数量不能替代当前代码检查。

## 电影站完整体验

| 旅程 | 建议行为 | 关键完成条件 |
| --- | --- | --- |
| 找电影 | 保留紧凑冷蓝门户；搜索、继续观看、影片列表为主体；年份/地区/类型按来源实际支持提供 | 筛选、结果数量和海报真实；慢源不吞掉成功源结果 |
| 辨认影片 | 展示名称、年份、地区、语言、简介和来源；同名低置信结果不强行合并 | 重拍片、同名片不会误选；元数据缺失按缺失展示 |
| 选择播放 | 主动作“播放正片/继续观看”；可选线路、配音、字幕或剪辑版本分开呈现 | 不把“高清”“国语”“导演剪辑版”都当成第 1/2/3 集 |
| 开始观看 | 统一加载、首帧、播放失败与恢复；移动端主要控件直接可操作 | HLS 与 MP4 路径均可用；自动播放被拒后可手动开始 |
| 调整观看 | 倍速、清晰度、字幕、音轨、画中画和全屏依据能力显示 | 菜单来自实际媒体能力；没有字幕/音轨时不展示虚假选项 |
| 断线恢复 | 原地址可恢复则有限重试；地址失效则重新解析；仍失败给可用线路选择 | 不反复加载失效 URL，不无限重试，不丢失本次位置 |
| 换线续播 | 同一内容时间轴可恢复位置；不同剪辑版不自动沿用相同秒数 | 不能只凭同名或集序号跨源同步进度 |
| 再次访问 | 匿名历史继续可用；登录后按明确冲突规则同步 | 换播放器、换新版站点都不会清空原有历史 |
| 运营维护 | 区分目录查询、播放地址解析、真实开始播放三个阶段 | “搜索可用”不再被误读为“每部影片均可播放” |

电影是默认体验；已有剧集内容继续使用同一底层接口和真正的集数列表，不删除既有数据。评论、社区、离线下载、AI 超分、弹幕不作为本次升级前置；后续是否需要按实际产品价值讨论。

## 统一接口设计

### 1. 片源接口：保留现有 Source Adapter

仍由 `source.Adapter` 隔离来源差异。必要的变化是增加结构化线路、来源能力与更准确的内容定位，而非再次创建功能相同的 Provider 服务。

建议领域关系：

```text
展示影片（可聚合多个来源；不是播放权威 ID）
  └─ 来源影片（sourceKey + externalID）
       ├─ 内容项（电影正片，或某一集）
       └─ 播放线路 / 版本
            └─ 对应内容项的稳定播放定位
                 └─ 临时 PlaybackManifest
```

线路代表同一内容的不同交付路径；版本可能代表不同配音、字幕或剪辑。清晰度档位通常属于媒体变体，不能强行等同于来源或线路。来源只提供一个地址时允许单线路，不虚构缺少的属性。

现有 `sourceKey + externalID + episodeKey` 仍可解析；新增结构不直接重排旧 key。旧数据无法证明线路身份时保留 legacy locator，运行时映射到来源默认线路。是否需要数据库迁移，应在正式接口设计时逐项决定，本方案不授权或创建迁移。

### 2. 播放解析接口：返回可执行播放描述

扩展现有 `PlaybackManifest`，以下为设计草案，不是已发布的 DTO：

| 字段/概念 | 含义与约束 |
| --- | --- |
| locator | 来源影片、内容项、线路或版本的稳定引用；不用临时 URL 作为影片 ID |
| url + mediaType | 当前媒体地址及明确类型；优先来源证据，后缀只做 fallback；类型未知不能伪装成可播放 |
| expiresAt? | 仅在上游提供可信有效期时填入；未知保留未知 |
| timelineKey? | 可靠确认相同剪辑和时间轴时才可共用；现有 source/title 字符串不能证明跨源等价 |
| tracks? | 来源实际提供的外部字幕/音轨描述；播放引擎解析出的内嵌轨道在运行时补充 |
| constraints | 来源的实际播放限制；不能通过下发任意请求头承诺浏览器可实现 |

解析错误归一为无法找到内容、来源超时、解析失败或当前环境无法播放等业务类别，按 Foundation 的现有错误目录生成合同。前端不解析来源错误字符串。字段演进同步 OpenAPI、Go DTO、TypeScript 和消费者；不引入另一套响应 Envelope。

### 3. 播放会话接口：业务页面只面向 YTV

建议由一个 YTV 播放会话 Module 对外提供小接口，内部使用成熟播放器自身的状态和能力，避免复制整个播放器框架：

```ts
// 行为草案；类型和实际方法需与选定播放器共同验证。
interface YtvPlaybackSession {
  open(selection: PlaybackSelection): Promise<void>;
  play(): Promise<void>;
  pause(): void;
  seek(seconds: number): void;
  retry(): Promise<void>;
  subscribe(listener: (snapshot: PlaybackSnapshot) => void): () => void;
  dispose(): void;
}
```

`open` 接受稳定定位并负责解析与载入。换线仍使用 `open`，不要求页面知道 Hls/Plyr/Video.js 实例。`retry` 根据错误阶段选择恢复媒体或重新解析地址；不能只是调用同 URL 的 initialize。控件专有操作由内部 Adapter 完成；字幕、音轨等先按真实消费需求接入，不做通用插件注册中心。

会话模块必须拥有以下行为约束：

- 新 open 取消或淘汰旧解析结果；快速切换影片时旧请求、旧事件不能覆盖新会话。
- 保存进度使用会话创建时的 locator，避免 props 已变化却把旧媒体时间写到新影片。
- 销毁幂等，释放播放引擎、订阅、定时器与媒体资源；不能有残留声音。
- 捕获原生媒体与 HLS 引擎错误；限制恢复次数，保留错误阶段和可执行操作。
- 暂停、切换、离开与周期保存按同一会话处理；localStorage 不可写、账号同步失败均不阻塞播放。
- 用户同意播放与浏览器允许播放分开处理；不以自动播放策略拒绝判断片源失效。
- 页面只使用状态、进度和能力快照；UI 控件样式与媒体引擎可分别替换。

实际状态建议归一为 resolving/loading/ready/playing/paused/buffering/ended/error/disposed；seeking 等播放器内部细节可作为快照属性，避免重复建立冲突的状态机。

## JavaScript 播放器选型

截至本次核对，Video.js 官方 latest 为 **10.2.0**。Plyr 官方已标记 deprecated/security-only，Vidstack 1.x 进入仅安全维护并明确截至 2028-01。新能力不宜继续绑定 Plyr 或改投处于同一维护方向的 Vidstack。[Video.js 发布](https://github.com/videojs/video.js/releases/tag/video.js%4010.2.0)、[Plyr 官方说明](https://github.com/sampotts/plyr)、[Vidstack 维护说明](https://vidstack.io/docs/player/)。

候选版本快照（GitHub 发布与 npm 信息交叉核对，仅供后续固定版本，不自动改依赖）：

| 项目 | 本次核实版本 | 官方版本依据 |
| --- | --- | --- |
| Video.js HTML / hlsjs-video | 10.2.0，正式发布 | [Release](https://github.com/videojs/video.js/releases/tag/video.js%4010.2.0)、[HTML npm 元数据](https://registry.npmjs.org/@videojs%2fhtml/latest) |
| Artplayer | 5.4.0 | [Release](https://github.com/zhw2590582/ArtPlayer/releases/tag/5.4.0) |
| hls.js | 1.7.3；YTV 当前固定 1.7.1 | [Release](https://github.com/video-dev/hls.js/releases/tag/v1.7.3) |
| Shaka Player | 5.3.1 | [Release](https://github.com/shaka-project/shaka-player/releases/tag/v5.3.1) |
| Plyr | npm latest 3.8.5；YTV 当前固定 3.8.4 | [npm 元数据](https://registry.npmjs.org/plyr/latest) |

Video.js 核心稳定版不意味着其所有扩展均稳定：当前官方导航仍将 `dash-video`、`shaka-video` 标为 Unstable。此次先验证已有 MP4/HLS 路径。[Shaka 适配组件](https://videojs.org/docs/framework/html/reference/components/shaka-video)。

| 方案 | 在 YTV 的角色 | 判断 |
| --- | --- | --- |
| Video.js 10 | 新控件与播放状态基础；Nuxt 使用 HTML/Web Components 路线 | 优先验证；已有官方 Plyr 迁移和 Vue 接入指南 |
| Artplayer + hls.js | 另一套控件与插件体系 | 备选，适合以字幕、播放菜单及快速定制为重点的比较 |
| hls.js | HLS 加载、解复用及自适应播放引擎 | 继续保留此能力；它本身不是完整播放器 UI |
| Shaka Player | 复杂 DASH/HLS、DRM 媒体场景的引擎与可选 UI | 当前没有真实 DASH/DRM 输入需求，不并行引入第二套大引擎 |
| 现有 Plyr | 本地对照与迁移基线 | 保留现有能力作为比较对象，不以仍能运行代表适合继续扩展 |

Video.js 使用 `@videojs/html` 和 `hlsjs-video` 路线，HLS 引擎仍由 hls.js 承担；不能把轻量 `hls-video` 名称相近视为完整等价。客户端组件承载浏览器依赖，SSR 只渲染稳定占位；对照当前 CSP、worker 路径、Nuxt 生产构建和路由卸载实测。[Vue 指南](https://videojs.org/docs/framework/html/guides/vue)、[hlsjs-video](https://videojs.org/docs/framework/html/reference/components/hlsjs-video)、[hls-video](https://videojs.org/docs/framework/html/reference/components/hls-video)。

迁移不是替换包名：现有倍速/音量偏好、MP4 多地址清晰度、网页全屏回退、进度恢复及快捷键必须逐项对照。官方迁移页有已知缺口，试验期间以 YTV 实际用到的行为决定工作量。[Plyr 迁移指南](https://videojs.org/docs/framework/html/guides/migrate-from-plyr)。

Artplayer 官方提供字幕、音轨、HLS 控制等插件和流媒体引擎接入文档，仍需测试插件与目标设备，而不能从插件列表推出所有片源都能提供这些能力。[Artplayer 文档](https://artplayer.org/document/en/)。Shaka 的候选价值来自其 DASH/HLS/EME 领域，本轮不将其作为必须引入的依赖。[Shaka 官方仓库](https://github.com/shaka-project/shaka-player)。

建议用同一组 MP4、HLS 单码率/多码率、字幕、音轨及错误媒体样本，比较现有 Plyr、Video.js 与必要时的 Artplayer。记录实际首帧、路由切换资源释放、产物体积和移动端操作；不把官网体积数字当 YTV 实测。

## 建议实施顺序与验收

以下顺序等待用户讨论确定，不是执行授权。

| 阶段 | 交付物 | 完成判断 |
| --- | --- | --- |
| A：模型与样本 | 电影/剧集/线路/版本的具体示例、现有 locator 兼容方案、统一 Manifest 草案 | 同一电影多线、不同剪辑、剧集多线、旧历史均能清晰表达；确定真正需要的字段和迁移 |
| B：播放器比较 | 同一会话接口下的 Video.js 候选，必要时 Artplayer 对照 | 现有 HLS/MP4、中文控件、SSR、CSP、卸载、Safari/手机限制有实际证据；完成最终选型 |
| C：统一播放与恢复 | 接入正式 Resolver、会话和错误映射；地址重解析、有限重试、可靠进度 | 失效 URL 可重解析；旧异步请求不会覆盖新片；进度不串片；无法恢复时有正确操作 |
| D：电影页面 | 电影详情与观看页、来源/线路/版本选择、字幕与音轨菜单 | 单电影无伪选集；多集保留选集；能力与媒体一致；实际截图人工检查 |
| E：发现与运营 | 筛选/搜索聚合改善、分阶段健康检测、首帧与卡顿指标 | 错误归因到查询/解析/播放阶段；不记录带 token 媒体 URL；指标对改善有帮助 |
| F：上线准备 | 当前新版依赖与环境复核、旧站历史迁移验证、独立发布清单 | 本地新版完整旅程通过；线上旧站和新制品准确对应；部署另获明确授权 |

如果希望更早看到效果，可以在 A 的必要字段确认后先做 B 和电影播放页样例；不要同时大改片源模型、播放器库和全站视觉，导致无法定位回归。

### 验证范围

自动检查针对行为：规则解析不串线路；搜索部分失败；播放地址重新解析；快速换片竞态；原生/HLS 错误；销毁清理；时间轴不匹配；存储不可用；旧 locator/历史恢复；同名不同电影不错误合并。测试跨模块接口，不只复述内部实现。

真实浏览器通过 Workspace 组合与仓库 CLI Playwright 进行：桌面与手机视口、匿名与登录、MP4/HLS、暂停/跳转/倍速、换线、重试、历史恢复、后台刷新。WebKit 自动化不能代替真实 iPhone 的全屏、AirPlay 或系统媒体能力验证；没有实机证据时明确记为未验收。

界面改动必须查看实际运行截图，人工检查布局、文字、对齐、点击区域、间距、控件状态、横竖屏与遮挡。自动截图、控件存在、可点击和测试通过不能作为视觉合格结论。

持久化如需变化，沿用项目 migration 不可变与用户数据保留规则；本轮不提高任何 writer 或格式版本。公共 DTO/错误变更遵循 Foundation 正式合同，依赖升级遵循 Workspace 基线。

## 本轮交付与下一步选择

本轮完成源码盘点、官方播放器维护状态核对和方案整理；没有安装或运行候选播放器，没有新增播放成功率、性能或视觉验收结论。没有 commit、push 或部署。

建议下一轮先确认两个产品取舍：电影为默认浏览/观看模式但保留现有剧集能力；Video.js 10 为第一验证候选。然后从 A 的内容模型样本和 B 的播放器比较开始，后续按实际验证结果调整顺序。
