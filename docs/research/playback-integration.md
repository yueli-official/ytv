# 统一影视与播放接口

适用于 YTV 页面、网站导航入口和将来独立客户端。源码合同为 [OpenAPI](../../api/contracts/openapi/ytv.json)，页面通过 `useYtvApi` 调用同域 `/api/v1`；Nuxt 服务端将请求转发给 YTV API，账号沿用 Identity/Foundation。网页不认识各采集站的 URL 格式，也不直接执行 Kazumi 规则。

```mermaid
flowchart LR
  Site[网站 / YTV 页面] --> API[YTV HTTP API]
  API --> Sources[来源适配器与规则]
  API --> Resolver[播放地址解析]
  Resolver --> Session[YTV 播放会话]
  Session --> Player[Video.js / 原生媒体]
```

## 请求链

| 操作 | HTTP | 结果 |
| --- | --- | --- |
| 可用来源 | `GET /api/v1/sources` | `items` 集合，包含来源状态与显示名称 |
| 搜索 | `GET /api/v1/search?q=影片名&sources=sourceA,sourceB&page=1` | 来源分组结果及部分失败情况 |
| 来源更新 | `GET /api/v1/discover?page=1` | 片库内容；不代表精选推荐 |
| 影片详情 | `GET /api/v1/titles/{sourceKey}/{externalId}` | `title` 元数据和 `episodes` |
| 解析播放 | `GET /api/v1/playback/{sourceKey}/{externalId}/{episodeKey}` | 可播放地址、引擎提示、媒体类型和时间轴标识 |
| 观看历史 | `GET /api/v1/progress` | 账号历史；匿名进度留在浏览器 |
| 更新进度 | `PUT /api/v1/progress/{sourceKey}/{externalId}/{episodeKey}` | 已认证用户的进度记录 |

各 path segment 必须独立 URL 编码。错误使用 `application/problem+json`，沿用公共 `code` 与 `traceId`。详情只返回 locator，不把媒体地址提前散落在页面数据里。

## 电影与线路

`episodes[]` 增补 `lineKey`、`lineName`；电影的“正片”和备用线路分别表示内容和播放入口。客户端按 `lineKey` 分组，而不是把所有入口排成连续剧。`episodeCount` 取各线路中的最大内容数，避免两条电影线路显示成两集。

MacCMS `vod_play_from` / `vod_play_url` 的 `$$$` 分隔线路、`#` 分隔条目、`$` 分隔名称与地址。首线路保持原数字 episodeKey；额外线路为 `line{index}-{episodeIndex}`。声明式规则保留原 URL 编码 key，按规则的 MultipleRoads 行为保留线路。

字段是兼容性增补。老客户端可继续使用原列表和定位符；新客户端对无线路字段的旧数据按默认线路显示。线路 key 来自来源顺序，无法保证上游重新编排后的语义不变，后续稳定内容身份工作仍需独立处理。

目前每条线路均只有一个条目的内容显示“播放正片”；这是一种观看界面判断，不能当作权威电影分类。长剧集按 50 集分页展示。

## 播放合同

解析返回原有 `url`、`engine`、`episodeKey`、`sourceName`、`timeline`，另加可选 `mediaType`。`hls` 接入 Video.js 的 hls.js 媒体组件；`direct` 使用浏览器媒体元素。MP4/WebM 的 MIME 根据路径提示，空值表示无法从地址可靠推断。未增加任意 URL 的解析、代理或转码入口。

`createPlaybackSession` 统一管理 resolve、mount、播放状态、手动重新解析、进度保存和销毁。页面不依赖具体播放器实例。旧 resolve 和媒体事件在换片后被丢弃；重新连接会重新调用解析接口，不重复使用已经失败的地址。进度使用原 `ytv-progress-v2` key，倍速/音量继续读取原 `plyr` 偏好记录。

换线路从目标线路的第一条目进入，只恢复该条目自己的进度；不会把不同片源或不同剪辑的秒数直接套用。当前 `timeline` 是历史兼容字段，不足以证明跨线路时间轴一致。跨设备历史仍由现有账号服务同步；不承诺不同线路无缝续播。

已登录播放器在解析地址时并行读取账号最近 100 条历史，同一个 locator 按更新时间比较账号与本机记录后恢复。账号读取最多等待 2.5 秒，失败或超时后使用本机记录，迟到响应不会在播放中突然跳转。历史页与首页合并两端记录，账号接口失败时仍保留本机入口；匿名历史同步在页面挂载后执行，不阻塞页面启动。

详情页同样合并本机与账号历史，默认选择最近记录对应的线路和 50 集范围，主操作显示继续观看与具体时间。用户主动选择线路后，迟到的账号记录不再切换所选线路。长简介默认折叠，播放操作保持在简介之前。

## 接入现有网站

同一 Nuxt 站点直接复用 `useYtvApi()`，卡片链接使用 `/title/{source}/{id}`，观看链接使用 `/watch/{source}/{id}/{episode}`。其他网站可先链接 YTV 页面；若需要把搜索/播放内嵌到另一站点，应由该站的服务端正式接入 YTV API，并配置对应网关、Origin 与 Identity 客户端。当前未发布 npm SDK 或通用跨域 iframe 播放器，不能把开发地址直接写进生产站。

搜索结果进入详情时携带 `from` 查询参数，详情与观看页传递该参数供返回原结果；只接受站内 `/search` 路径及其查询，不接受任意返回地址。搜索词、页码和来源范围保存在 URL；搜索历史更新输入框，空页保留上一页入口，同一影片可展开全部候选来源。

接口返回来自获准来源的数据。规则目录、播放主机约束、来源启停和运营统计继续由服务端管理；网站无需复制采集逻辑。字幕、音轨、清晰度只有上游媒体实际提供时才由播放器显示，页面不生成虚假的版本选项。

## 验证与边界

本次验证使用 Workspace 真实组合、自产演示 HLS 和由同一合成片段重新封装的 `web/public/demo/feature.mp4`；MP4 浏览器测试替换解析响应以覆盖 native 引擎，不改变生产来源准入。为验证精确的 10 秒跳转，MP4 夹具循环到 60 秒：`ffmpeg -stream_loop 7 -i segment-000.ts -t 60 -c copy -movflags +faststart feature.mp4`。

已覆盖主/备用线路切换、HLS 播放、原生 MP4、播放错误重新解析、旧异步结果、卸载清理、存储不可用及旧播放器偏好。浏览器手机检查为 Chromium 移动视口；真实 iOS Safari、投屏、多音轨/字幕媒体与线上来源全量可用性不能由该测试推出。半商业运营的支付、会员权益、广告投放及内容采购不在此次实现中。

## 播放操作

播放中按住画面 400ms 临时切到 2 倍速，松开、移出、取消触摸、页面失焦、暂停或销毁后恢复按住前倍速；临时值不写入默认偏好。进度条、设置和其他控件不触发长按。键盘左右方向键及 J/L 跳转 10 秒，双击画面左右两侧分别快退/快进 10 秒，空格/K 播放暂停，F 全屏；这些操作沿用 Video.js 原生绑定，避免重复触发。长按提示放在播放器内容内，桌面全屏下也可见。手机验证是 Chromium 触摸模拟，真实 iOS 原生全屏仍需设备验证。
