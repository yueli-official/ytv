# 聚合播放产品与来源验证调研

调研日期：2026-10-23。目标是决定 YTV 如何处理来源验证及失败恢复，不是复制其他项目，也不是将全部来源标成可用。依据官方仓库、具体源码及浏览器平台文档；没有对第三方实例作流量、稳定性或市场份额测量。“代表性”指覆盖主要架构，非排名。

## 结论

YTV 应保持 Web 聚合播放主线：优先服务端可稳定检索、浏览器能直放的来源；原生客户端的浏览器会话能力不能靠网页弹窗替代。验证码不是单个 UI 控件，而是检测、用户操作、会话保持、原请求恢复与过期清理的完整链路。

推荐按来源能力选择路径：直接接口、静态页面、需交互验证、仅客户端可解析。来源“规则声明需要验证”和“本次请求确实遇到验证”是两个不同事实，不能互相替代。

## 代表性产品比较

| 产品 | 已核实的路线 | YTV 借鉴点 | 不直接移植的部分 |
| --- | --- | --- | --- |
| Kazumi | Flutter 客户端，规则解析；验证码 WebView 完成后保存 Cookie、UA 和结果 HTML；搜索状态区分 captcha/noResult/error | 每来源状态、验证结束后恢复原搜索、验证生命周期 | 原生 WebView 权限不能当作普通浏览器权限 |
| LibreTV | Web 聚合站，Apple CMS 接口；部分来源使用详情网页；README 提供源管理和代理安全说明 | 来源管理、结构化接口优先、播放与搜索分层 | 其代理能力不改变 YTV 不代理第三方视频的合同 |
| MoonTV/LunaTV | Web 聚合，CMS V10，来源配置、详情、收藏和多端进度；播放器为 ArtPlayer/HLS.js | 连贯观看旅程、收藏/续播、统一资料展示 | 官方项目包含视频代理相关实现，不能据此承诺 YTV 同等兼容 |
| Cloudstream | Android 扩展/提取器体系，WebView 和 CookieManager，网络请求复用浏览器 UA/Cookie | 能力按平台区分、会话与请求绑定 | 客户端本机行为不等于多用户服务器可安全共享会话 |
| Stremio | Addon 协议分离 catalog/meta/stream；stream 支持直接 URL 与 externalUrl | 目录可用和播放可用分开，明确外部打开与站内播放 | Addon 协议不是通用验证码求解器；不同流类型取决于客户端支持 |

官方入口：[Kazumi](https://github.com/Predidit/Kazumi)、[LibreTV](https://github.com/LibreSpark/LibreTV/blob/main/README.md)、[LunaTV](https://github.com/MoonTechLab/LunaTV/blob/main/README.md)、[Cloudstream](https://github.com/recloudstream/cloudstream)、[Stremio Stream](https://github.com/Stremio/stremio-addon-sdk/blob/master/docs/api/responses/stream.md)。本次只借鉴架构与交互，不复制实现。

## Kazumi 验证到底做了什么

源码固定在 `db5202611902fe67c449e0a53bc05c3a75b4e8b0`。

- `anti_crawler_config.dart` 明确定义三种类型：图片输入、按钮验证、自定义 JavaScript。`enabled` 缺省为 false。
- `captcha_verification_service.dart` 管理验证、取消和释放；验证成功后读取 HTML、Cookie、UA，再交给后续搜索流程。不是简单打开第三方网址。
- `plugin_search_service.dart` 单独记录需要验证码、没有结果和一般错误，失败来源不应吞掉其他成功结果。
- 上游 issue #2578 报告过非 2xx challenge 提前变成网络错误的情况。它是问题报告，不代表所有当前版本都仍然存在；它提醒我们不能只在 HTTP 200 的空结果后识别验证。

源码：[配置](https://github.com/Predidit/Kazumi/blob/db5202611902fe67c449e0a53bc05c3a75b4e8b0/lib/plugins/anti_crawler_config.dart)、[验证生命周期](https://github.com/Predidit/Kazumi/blob/db5202611902fe67c449e0a53bc05c3a75b4e8b0/lib/services/plugin/captcha_verification_service.dart)、[搜索状态](https://github.com/Predidit/Kazumi/blob/db5202611902fe67c449e0a53bc05c3a75b4e8b0/lib/services/plugin/plugin_search_service.dart)、[非 2xx 问题报告](https://github.com/Predidit/Kazumi/issues/2578)。

Cloudstream 对照源码固定在 `a7ac454af8fbd64574c96847a936905855911720`，其网络实现直接使用 Android CookieManager 和 WebView UA，进一步证明这是一种客户端会话方案：[源码](https://github.com/recloudstream/cloudstream/blob/a7ac454af8fbd64574c96847a936905855911720/app/src/main/java/com/lagradost/cloudstream3/network/CloudflareKiller.kt)。

## Web 可行性边界

1. **普通图片验证码**：若来源允许、且验证与后续请求可以在同一服务端会话完成，可实现专用适配器。需要按用户/来源隔离短期 Cookie jar；验证码图片、提交端点、参数、重定向必须由部署定义固定。用户手动输入；成功后只恢复原搜索一次。不能用全站共用 Cookie，也不能让前端提交任意 URL。
2. **浏览器挑战**：普通网页不能读取另一个站点的 DOM 和 Cookie。新窗口验证只能改变该浏览器对应站点的状态，不能自然传回 Go 服务端。iframe 也可能受到 CSP/frame-ancestors 等限制。
3. **Cloudflare 等设备绑定状态**：官方说明 clearance 与访客及设备关联，跨机器复用并不可靠。官方可靠识别信号为 `cf-mitigated: challenge`；403/429/503 本身不足以判定验证码。
4. **原生客户端或本地伴随程序**：可让验证、搜索、解析留在同一用户设备上，但这是独立产品能力，不是加一个网页组件。需要单独管理安装、更新、会话和平台兼容。
5. **去源站打开**：可以作为明确的站外观看选择；不能标成“验证后本站即可播放”。更不能把返回本站当作验证成功。

依据：[同源策略](https://developer.mozilla.org/en-US/docs/Web/Security/Defenses/Same-origin_policy)、[Cloudflare clearance](https://developers.cloudflare.com/cloudflare-challenges/concepts/clearance/)、[Challenge 响应识别](https://developers.cloudflare.com/cloudflare-challenges/challenge-types/challenge-pages/detect-response/)。本节可行性方案是结合这些约束与 YTV 架构的工程推导。

## YTV 现状与本次落地

本地部署快照 `5fea5eb84768a2290ddedb156583ef20d9b43561` 有 16 条规则，其中 dalvdm、giriGiriLove、mgnacg、mutefun 共 4 条声明启用图片验证码；其余 12 条未启用。这不是当前线上安装列表，也不证明这 4 个来源每次请求都会验证。

初次调研时 Adapter 没有用户级验证码会话，也没有请求级识别。后续已实现下述识别层；用户提交与验证会话仍未实现。

已完成两项本地改进：

- 编译器按 enabled/type 区分关闭配置、交互验证与自定义脚本。关闭配置不再误报任意脚本，启用图片/按钮验证明确提示不支持交互，自定义脚本保持不执行。这里是兼容性诊断，并非本次网络检测；目录仍按现有批准列表与安装策略工作。
- 搜索页增加可折叠逐来源结果：结果条数、无匹配、请求超时、暂不可用。保留成功影片和重新搜索入口，不把一源失败当全局无片。

## 后续实施顺序与完成条件

**第一步：请求级识别。** 在适配器收取响应后、状态码与 XPath 解析前识别可靠 challenge 信号；图片验证码使用逐来源已批准的 DOM 组合特征，不能用全文包含“验证码”猜测。新增 typed cause、声明式错误目录及搜索来源 outcome，同时覆盖全部来源验证/部分来源验证两种情况。测试 200 challenge、403 challenge、普通 403、正常含验证码字样页面，避免误报和空结果掩盖。

**第二步：一种真实图片验证码的端到端试点。** 从上述 4 个来源选择实际仍可访问且可重复触发的站点，核对图片/提交/结果会话链，再实现短期、隔离的用户验证。验收包括答错、刷新图片、取消、过期、重复提交、两用户隔离和恢复原搜索。当前没有这些站点的已完成验证证据，不能宣称已支持。

**第三步：明确客户端专属来源。** 对必须依赖原生浏览器或设备绑定会话的来源，展示能力限制与换源；若产品决定提供本地伴随程序，再单独实施，不搭建一个冒充成功的验证码弹窗。

其他值得排期的体验：来源健康按阶段/时间记录、局部重试失败来源而不丢已返回结果、缓存保留过期与强制刷新的明确语义、同片不同版的保守分组、换源时保留集数并确认版本差异。已有换源/换线路入口不代表这些完整能力已实现。

## 验证记录

规则编译、目录和声明式 Adapter 三个 Go 包通过；前端类型检查与来源状态桌面交互/截图复查记录在当前 Flightdeck Work。未提交、未部署；开发 API 未重启，本轮编译器改进尚未加载到运行服务。


## 2026-10-23 请求级适配落地

`internal/source/verification` 提供 Detector、响应上下文、组合检测器和 typed Required cause。来源规则只声明经过部署审核的选择器；检测器不执行来源脚本，不向前端泄露 HTML、Cookie 或上游地址。HTTP challenge header 与 DOM 检测分阶段执行，普通 403 不自动判为验证码。

- 声明式 HTML 请求在结果解析前识别 Cloudflare 明确信号及完整 DOM 组合；特殊播放解析请求也识别该 header。其他 catalog 类型尚未统一接入 DOM 识别。
- 搜索返回 `verification_required` 和 `verificationKind`。全受阻时保留来源状态；部分受阻时保留正常影片。详情/播放通过 catalog 生成的 `ytv.source_verification_required`（409）传递。
- 搜索界面区分无匹配、图片验证码、浏览器验证、超时和不可用；详情/播放保留具体原因和返回搜索/换源入口。
- mutefun 实际返回 HTTP 200 MacCMS 验证页，旧快照 XPath 已失配。本次新增审核过的标题+输入框+提交按钮组合，真实 Go Adapter 搜索识别通过。它不在当前六源批准列表，独立测试直接编译快照，不修改批准列表或安装来源。其他探测：dalvdm 403、giriGiriLove 普通页面、mgnacg TLS 失败，均不能直接推定为验证码。

### 会话适配层的后续边界（设计，尚未实现）

来源专用 Challenge Adapter 应负责创建挑战、刷新图片、提交答案和读取明确验证结果；Session Store 负责用户/来源隔离、短期 Cookie jar、UA、过期与取消；恢复器保存原始操作参数并只恢复一次。页面只能提交不可猜测的会话 ID 与答案，不能传来源 URL、脚本或 Cookie。验证失败不能更新为成功，也不能把用户回到本站视为通过。

第一种图片验证码适配器必须先验证静态图片/提交端点、会话绑定及成功判据，再把接口固定下来；当前只从真实 HTML 确认了 verify_check 提交调用，验证码图片由源站脚本生成，尚未完成图像与会话链核对。因此没有添加无法工作的验证码输入框或空实现接口。

### 本轮验证和交付状态

7 个相关 Go 包通过；真实 mutefun 独立探测通过；28 个操作的生成合同校验通过；Nuxt 直接类型检查通过。桌面 Playwright 两项通过（搜索全/部分受阻、详情与播放），亲自检查搜索/详情/播放截图，并修正详情错误区过亮与贴导航的问题。

`pnpm typecheck` 的正式依赖门禁因既有 Foundation 源码覆盖而失败，单独 `pnpm exec nuxt typecheck` 通过不替代该门禁。浏览器测试使用 API fixture 响应；开发 API 未重启，线上未部署。当前是本地识别与反馈能力，不是已支持验证码完成或全来源稳定播放。

证据目录：`D:/coding/yueli-official/task-artifacts/ytv/verification-adapter-20261023-07/`。保存实际验证页面和桌面截图；临时合同生成程序已在检查完成后回收。


## 2026-10-23 图片验证码会话试点

已实现内部会话管理器 `verification.Sessions` 与 `declarative.NewMutefunSearchVerification`。未增加公开 HTTP 路由、未在页面显示验证码输入框、未扩大批准来源列表；该能力目前通过受控试点调用，不应宣称站点用户已经可以使用。

会话按调用方提供的可信用户标识隔离，随机 ID 不能替代所有权；每用户最多 2 个、全局最多 128 个，3 分钟有效。启动/读取/提交/取消时清理过期条目，运行时所有者须调用 Close；无后台清扫线程。Cookie jar 每会话独立，只在内存中使用；Flow 生命周期有超时与取消。答错可刷新后重试；提交网络结果不明确则结束会话。提交成功后只恢复最初捕获的搜索一次，重复请求不能再次恢复；恢复进行中仍受取消与关闭管理。

来源适配器只接受固定 mutefun 定义；检查图片 MIME、真实图片格式、尺寸和体积，拒绝 HTML 冒充图片及跨来源重定向。每阶段使用一致 UA 与 Cookie。配置与搜索参数在创建时复制，不接受前端提供图片/提交 URL、Cookie 或脚本。

真实协议核对：`/static/js/home.js` 的 MAC.Verify.Show 在输入框聚焦时生成 `/index.php/verify/index.html` 图片；POST `/index.php/ajax/verify_check`，query 参数为 type=search 和 verify。实测错误答案返回 HTTP 200 / code 1002；code 1 表示接受。其他/缺失业务码不是成功。

### 实际验证

- Go 7 个相关包通过；verification 和 declarative 两包 race 检查通过，之后补充恢复期间关闭的回归也通过 race。
- 测试覆盖两用户 Cookie 隔离、错误答案、刷新、取消、过期、容量、重复与并发提交、关闭时的恢复取消、模糊响应、跨域重定向和伪造图片。
- 首次人工试点未及时收到输入，明确标记 Skip 并关闭会话，没有作为成功证据。第二次由用户输入新图片答案，源站接受；实际 Go 适配器在同一 Cookie 会话恢复“海贼王”搜索，返回 1 条结果。没有下载视频，没有修改线上站点。
- 产物 `D:/coding/yueli-official/task-artifacts/ytv/verification-session-20261023-08/` 保存两次图片；人工答案临时文件已在测试退出后删除。会话 Cookie 从未落盘。

### 下一步产品接入

需要将已实现的 Sessions 由 API runtime 持有并关闭，入口通过现有 Identity principal 提供 owner，检查来源已安装/启用及部署批准，再开放创建、刷新、提交和取消操作。成功响应只返回恢复后的搜索结果，不能返回 Cookie。公开错误需进入声明式目录并生成合同；页面需要输入、刷新、取消、过期和答错状态，并进行 Playwright/人工视觉检查。当前 mutefun 仍未批准为完整播放来源，其搜索验证码试点不代表媒体播放合同通过。不得仅为展示验证码 UI 就绕开来源安装策略。
