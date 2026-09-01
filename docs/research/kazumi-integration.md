# Kazumi 规则接入 YTV 的可行性调研

> 调研日期：2026-09-01。资料仅使用 Predidit 官方 GitHub 仓库、规则仓库、源代码与许可证。本文不是法律意见。

## 结论

**技术上可行，但只能“借鉴/适配经审核的 Kazumi 规则子集”，不能把 KazumiRules 当作可在线安装的开放规则市场。**

Kazumi 本身是 Flutter 客户端，不是内容 API。它用规则中的搜索 URL 和 XPath 抓取搜索结果、详情页播放线路与集数链接，再用 WebView 观察页面发出的媒体请求或 iframe，得到 M3U8/MP4 地址交给播放器。官方 README 也将其定义为“基于自定义规则的番剧采集与在线观看程序”，并说明规则主要由最多五行 XPath 选择器组成。[Kazumi README](https://github.com/Predidit/Kazumi/blob/main/README.md)

YTV 可以实现一个独立、受限的兼容适配器，但必须满足：规则、endpoint、抓取行为全部随代码或部署制品发布；域名和重定向逐跳校验；管理端只能安装已批准的定义并调整展示/启用策略；浏览器直接播放第三方流，YTV 后端不得下载、转码、代理或持久化第三方视频。**上线前还必须逐来源确认内容授权、站点条款及适用法；开源许可证不授予第三方动漫或流媒体内容权利。**

## 规则格式与执行链路

典型规则是一个 JSON 对象。官方 `AGE.json` 展示了核心字段：`api`、`type`、`name`、`version`、`muliSources`（上游拼写如此）、`useWebview`、`useNativePlayer`、`userAgent`、`baseURL`、含 `@keyword` 的 `searchURL`，以及五个 XPath 字段 `searchList`、`searchName`、`searchResult`、`chapterRoads`、`chapterResult`。[AGE.json](https://github.com/Predidit/KazumiRules/blob/main/AGE.json)

规则仓库给出的 API level 能力矩阵是：level 2 增加 POST 搜索；level 3 增加 Referer 和 legacy iframe 解析；level 5 增加 HLS 广告过滤。[KazumiRules README](https://github.com/Predidit/KazumiRules/blob/main/README.md) 当前规则还能出现 `usePost`、`useLegacyParser`、`adBlocker`、`referer`、`deprecated`，以及包含检测值、DOM selector、甚至 JavaScript 的 `antiCrawlerConfig`；官方 `LMM.json` 是脚本型反爬配置实例。[LMM.json](https://github.com/Predidit/KazumiRules/blob/main/LMM.json)

推导出的主链路如下（字段语义可由上述规则实例及代码交叉确认）：

1. 搜索：将关键词编码后替换 `searchURL` 的 `@keyword`，按规则选择 GET/POST；解析 HTML，以 `searchList` 找结果节点，再以相对 XPath `searchName` 和 `searchResult` 取标题与详情链接。搜索结果模型只有 `name` 和 `src`。[搜索结果模型](https://github.com/Predidit/Kazumi/blob/main/lib/modules/search/plugin_search_module.dart)
2. 详情/线路/集数：请求详情链接，以 `chapterRoads` 找线路，以 `chapterResult` 提取每条线路的集数链接；`muliSources` 表示页面可能有多线路。规则本身通常不提供结构化海报、简介或版权元数据。
3. 播放解析：把集数页面 URL 交给视频源解析服务。其接口明确接收 `episodeUrl` 和 `useLegacyParser`，返回 M3U8、MP4 或本地路径；在线实现是 WebView，legacy 模式监听 iframe。[视频源接口](https://github.com/Predidit/Kazumi/blob/main/lib/services/video_source/video_source_service.dart) [WebView 实现](https://github.com/Predidit/Kazumi/blob/main/lib/services/video_source/webview_video_source_service.dart)
4. 播放：解析出的第三方媒体 URL 被交给本地播放器；这不是一个稳定的上游 API，链接可能依赖 Referer、User-Agent、Cookie、短效 token、重定向或浏览器执行环境。

Kazumi 明确依赖 `xpath_selector`、HTML XPath parser、WebView 和 fork 的 `media_kit`，说明完整兼容不是“读取 JSON 后直接喂给现有播放器”，而是要重建 HTML 选择器、浏览器解析和播放兼容层。[pubspec.yaml](https://github.com/Predidit/Kazumi/blob/main/pubspec.yaml)

## YTV 支持矩阵

| 能力 | 技术可行性 | YTV 建议 |
|---|---:|---|
| 固定域名 GET 搜索 + HTML/XPath | 高 | 首期支持；服务端抓取，严格 allowlist、超时、大小和并发限制 |
| 固定域名 POST 搜索 | 中 | 后续逐规则批准固定 method/body；禁止管理员自定义 body/URL |
| 搜索结果、线路、集数解析 | 高 | 支持 Kazumi 核心 XPath 子集；编译期校验并生成规范化定义 |
| Referer / User-Agent | 中 | 只能取定义内固定值；不得从用户输入透传 |
| 浏览器直接播放明确的 HLS/MP4 | 中 | 可试点；需验证 CORS、CSP、Referer、Cookie 和浏览器 codec |
| WebView JavaScript hook 抓媒体 URL | 低 | Web 端没有 Kazumi 原生 WebView 的等价可靠能力；不作为首期承诺 |
| legacy iframe 解析 | 低 | 易受跨域、CSP、反嵌套影响；首期拒绝 |
| `antiCrawlerConfig`/任意脚本 | 不接受 | 等同远程代码执行与浏览器自动化，违反静态抓取行为合同 |
| HLS 广告过滤/manifest 改写 | 不接受 | 往往需要代理或改写第三方流，且增加规避措施与版权风险 |
| 下载、缓存、转码、投屏中继 | 不接受 | 与 YTV“不代理/存储第三方视频”约束冲突 |
| 从 URL/文件导入任意规则 | 不接受 | 形成 SSRF、内网探测、远程脚本和供应链入口 |

## 安全边界

Kazumi 的开放规则设计适用于最终用户自己控制的客户端，却不适合直接搬到多用户 Web 服务：`baseURL`、`searchURL`、详情链接和播放链接均来自外部，XPath 结果还能继续产生新 URL；脚本型验证码配置甚至包含可执行 JavaScript。因此兼容器至少需要以下强制控制：

- 规则以 YTV 自有 schema/manifest 固定在代码或签名部署包中，记录来源、上游 commit、reviewer、启用域名、能力和到期复审日；不在运行时自动追踪 KazumiRules `main` 或 `index.json`。
- 对 search、detail、episode、redirect、media 每一跳重新做 `https`、精确 hostname/port allowlist 和 DNS/IP 校验；拒绝 loopback、私网、链路本地、云 metadata、非 HTTP(S)、userinfo、非常规端口及 DNS rebinding。
- 限制响应字节、压缩比、重定向次数、请求时间、并发、频率；不带 YTV 用户 Cookie/Authorization，不共享来源 Cookie jar，不写响应正文或媒体 URL 到长期日志。
- XPath 只解析静态 HTML；禁用任意 JS、任意 header、任意 POST body、文件访问和浏览器下载。管理端 probe 也必须走同一 allowlist 和网络策略。
- 返回播放 URL 前确认 scheme/host；短效 URL 只在内存/短 TTL 中使用，前端直接访问第三方 origin。若来源要求 YTV 代理流、Cookie 注入或 manifest 改写，则标记不兼容而不是绕开限制。

## 许可证与法律风险

- Kazumi 程序代码使用 **GPL-3.0**。复制、修改或组合其实现并分发时会触发 GPL 对完整对应源代码和同许可证分发等要求；独立重写一个只读取同类 JSON 数据的适配器通常比复制 Dart 源码更易隔离许可证风险，但边界应由法律顾问确认。[Kazumi LICENSE](https://github.com/Predidit/Kazumi/blob/main/LICENSE)
- KazumiRules 仓库标注 **MIT**，可在保留版权与许可声明的前提下复制/修改规则文件。[KazumiRules LICENSE](https://github.com/Predidit/KazumiRules/blob/main/LICENSE) 但规则作者是否有权授权其中指向的网站、页面结构、商标或内容，是另一个问题。
- 两份软件许可证都不授予动漫作品、字幕、封面、站点数据库或第三方视频流的版权。聚合搜索、深链播放、绕过广告/反爬、对外提供规模化访问，可能涉及版权、信息网络传播权、反规避、站点服务条款、数据库权益及不正当竞争风险；责任不会因“不存储视频”自动消失。
- 来源常变域名、使用验证码或反爬本身是高风险信号。YTV 应要求每个来源有明确的授权/法律审查状态、下架入口和快速停用开关；未知或无法证明合规的来源不得在生产环境启用。

## 推荐的分阶段方案

### Phase 0：离线兼容性验证

实现一个不联网的 Kazumi JSON lint/转换器，只接受仓库内测试夹具，输出 YTV 自有只读 source definition。先支持 API level 1 的核心字段和 XPath 子集，建立黄金 HTML fixtures、URL allowlist 测试、重定向/SSRF 测试和许可证 NOTICE。不要复制 Kazumi Dart 实现。

### Phase 1：单个获准来源试点

选择一个具有明确书面授权、无需登录/验证码/脚本、直接输出浏览器可播 HLS/MP4 的来源。实现 search → detail → roads/episodes → resolve → browser-direct playback；后端只抓 HTML，不碰媒体字节。管理端只允许安装/移除该已批准定义、改展示名及 enabled/home-visible/priority，并提供受限 probe。

### Phase 2：受控扩展

在安全与法务逐来源批准后增加固定 POST、Referer/User-Agent、健康度与降级策略。每个定义锁定上游 commit 和内容哈希，通过 PR 更新；规则失效时自动降级并提示来源不可用，不接受管理员临时粘贴替代域名。

### 明确不做

不做远端规则市场、自动同步 `main`、任意 URL 导入、验证码/反爬脚本、WebView 通用浏览器机器人、HLS 广告剔除、媒体代理/缓存/下载/转码。若“能播放”的前提是其中任一项，该来源对 YTV 即不可接入。

## 最终判断

Kazumi 最值得融入 YTV 的是**规则数据模型与搜索/剧集抽取思路**，而不是完整运行时。建议立项一个“受审查的 Kazumi-compatible source adapter”，目标是兼容静态 XPath 规则子集；先拿一个有授权、可浏览器直播的来源做端到端 PoC。若没有可证明合规的内容来源，即使技术 PoC 成功也不应上线“在线动漫”能力。

## 已实现的接入基础（2026-09-01）

YTV 已增加格式无关的 `rules.Plan`、可注册格式 Adapter 的 rule compiler、Kazumi compiler 和统一 declarative Source Adapter。Kazumi JSON 目前可以编译搜索、详情线路、剧集与播放提取计划；以后增加其他规则格式时，只需实现同一 compiler Interface，不需要修改 Discovery、Source Control 或播放器。

运行时支持静态 XPath、固定 GET/POST、固定 Referer/User-Agent、相对 URL、直链 HLS/MP4/WebM，以及部署定义提供的 `ytv.mediaXPath`。WebView、iframe、反爬脚本或 manifest 改写等能力会生成兼容提示并随 Source API 返回，不会阻止管理员安装来源；YTV 只是不执行违反当前产品合同或尚无 Adapter 的行为。后台数据源列表与安装窗口会原位展示首条兼容提示。

YTV 现还会从部署固定的 KazumiRules commit `5fea5eb84768a2290ddedb156583ef20d9b43561` 加载官方 `index.json`，四并发获取并编译目录规则。目录或单条规则失败只会降级该批候选，不阻塞既有来源；fixture 验收不访问远端目录。2026-09-01 live smoke 中，索引的 16 条规则有 15 条成功编译；`sorani` 属于缺少传统 `chapterRoads` 的新 API 型规则，等待独立格式 Adapter。
