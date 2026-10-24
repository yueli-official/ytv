# 线上全部来源五片审核数据表（2026-10-24）

线上公开来源接口实取 **21 个启用来源**：10个综合、11个动漫（包括 aafun）。每源五个样本槽，共105条最终审核记录；修复补测后实际出画 **64条**，人工查看 **191张** 画面；41条未出画，不能宣称已完成画质审核。实际清晰度达到本轮接受标准的有 **44条**。

补测更新：已发布 `ytv-playback-policy-20261024-06`，删除播放地址域名/IP/HTTP授权限制。原先拦截的8条全部出画，补查24帧清晰且未见博彩广告。两源与8条明细已更新；初次失败证据保留。

## 审核口径

- 通过正式站 `https://tv.yuelili.com` 的来源、搜索、详情、播放解析接口取数据，再在本机 Chromium 中播放线上接口返回的地址；没有绕开线上失败直取上游，也没有用本地开发版冒充线上结果。媒体播放使用隔离 video/Hls.js 页面，所以验证的是线上接口链与媒体可播性，不等同于完整产品播放器、手机端或所有浏览器验收。
- 电影优先选择《流浪地球2》《飞驰人生2》《头脑特工队2》《蜘蛛侠：纵横宇宙》《哪吒之魔童降世》；动漫优先《葬送的芙莉莲》《鬼灭之刃》《咒术回战》《间谍过家家》《链锯人/电锯人》。每个片目采用详情接口返回的第一条集数/线路。
- 搜索不到精确片名时补查别名或替换：电影天堂补《沙丘2》；猫眼补《独行月球》《流浪地球》；akianime补《进击的巨人 最终季 Part.2》。OmoFun实际匹配《咒术回战0》《间谍过家家 第二季》，属于不同作品/季，不能当同版本横评。所有实际标题见明细。
- 每条出画样本检查约35秒、总时长45%、85%三个点。猫眼《独行月球》第三个点加载失败，仅两帧。广告只记录抽到的画面，不把“未见”写成全片无广告；不推断广告总时长。
- 人工检查画面细节、字幕/线条、模糊与广告遮挡，结合视频尺寸判定。1920宽的电影裁黑边版本可列1080档候选；720p及以下不达本轮目标。3840宽不证明原生4K，截图也不构成母版分辨率鉴定。本轮仅判断是否达到可接受清晰度。
- API单次请求超时65秒、开始出画等待23秒、每次拖动等待18秒。失败结论只代表本次时间、网络、默认线路；不能推断来源永久不可用。未出画的来源没有广告或画质结论。
- 首轮仅审核；随后按用户要求单独部署播放地址限制修复。没有修改生产数据库、来源配置、推荐或用户评价，没有commit/push。

## 来源汇总

“清晰”列只计实际看过画面且达到本轮目标的样本。分母均为5；失败样本保留在分母中。

| 来源 | 出画/5 | 清晰/5 | 画质与广告/故障 | 结论 |
|---|---:|---:|---|---|
| 非凡资源 (`ffzy`) | 5/5 | 5/5 | 修复后五片清晰，15点未见博彩广告 | 本轮优先保留 |
| 电影天堂 (`dytt`) | 5/5 | 5/5 | 修复后五片清晰，15点未见博彩广告 | 本轮优先保留 |
| 猫眼资源 (`maoyan`) | 5/5 | 3/5 | 三片清晰，两片低清；四片见广告，其中一片整屏广告 | 不作为推荐源；低清带广告样本淘汰 |
| 爱奇艺资源 (`iqiyi`) | 2/5 | 2/5 | 已出画两片清晰；两片见广告；另三片接口失败 | 暂不推荐；可用片保留，先修可用性 |
| 无尽资源 (`wujin`) | 5/5 | 0/5 | 五片均低于本轮1080档目标；采样未见博彩广告 | 不列高清推荐，可作低清备选 |
| 索尼资源 (`qgsuoni`) | 0/5 | 0/5 | 未审；五片搜索503 | 暂停推荐，查搜索适配 |
| 樱花资源 (`yinghua`) | 0/5 | 0/5 | 未审；五片搜索503 | 暂停推荐，查搜索适配 |
| 最大资源 (`zuid`) | 4/5 | 1/5 | 一片清晰；两片低清，一片竖屏套黑框；飞驰人生2见VPN叠字；一片无法加载 | 不推荐，逐片淘汰问题资源 |
| 魔都资源 (`mdzy`) | 5/5 | 5/5 | 五片画面清晰，达到本轮接受标准；五片开头见博彩叠字；蜘蛛侠中段仍有 | 可保留，明确有广告，不宣传无广告或原生4K |
| 360资源 (`zy360`) | 5/5 | 1/5 | 一片清晰，四片低于目标；五片开头见博彩广告 | 不推荐；四条低清带广告样本淘汰 |
| 7sefun (`kazumi-7sefun`) | 5/5 | 5/5 | 五片清晰，均1920×1080；15个采样点未见博彩广告 | 本轮优先保留 |
| aafun (`kazumi-aafun`) | 0/5 | 0/5 | 未审；五片解析503 | 暂停推荐，查解析适配 |
| akianime (`kazumi-akianime`) | 2/5 | 2/5 | 已出画两片清晰；采样未见广告；两片解析失败，一片加载超时 | 保留观察，修复失败片后再推荐 |
| moonci (`kazumi-moonci`) | 5/5 | 5/5 | 五片清晰，均1920×1080；15个采样点未见博彩广告 | 本轮优先保留 |
| MXdm (`kazumi-mxdm`) | 5/5 | 0/5 | 五片720p或更低；三片后段见博彩横幅 | 不推荐；低清带广告样本淘汰 |
| xfdmneo (`kazumi-xfdmneo`) | 0/5 | 0/5 | 未审；五片搜索503 | 暂停推荐，查搜索适配 |
| giriGiriLove (`kazumi-girigirilove`) | 5/5 | 5/5 | 五片清晰，均1920×1080；15个采样点未见博彩广告 | 本轮优先保留 |
| mutefun (`kazumi-mutefun`) | 0/5 | 0/5 | 未审；五片解析504 | 暂停推荐，查超时 |
| OmoFun (`kazumi-omofun`) | 3/5 | 2/5 | 两片清晰；咒术回战0为720p；三条出画样本均见博彩广告 | 不推荐整源；低清带广告样本淘汰 |
| 番茄动漫 (`kazumi-fqdm`) | 2/5 | 2/5 | 已出画两片清晰；两片中段见博彩横幅；三片加载超时 | 保留观察，不作为稳定推荐 |
| 动漫共和国 (`kazumi-dmghg`) | 1/5 | 1/5 | 已出画一片清晰；该片采样未见广告；另外四片搜索/解析失败 | 暂不推荐，先修可用性 |

## 关键结论

1. 本轮优先保留的动漫来源是 **7sefun、moonci、giriGiriLove**：五片均出画、细节清楚，15点未见博彩广告。这是本轮样本结论，不能外推整个片库或长期稳定性。
2. **魔都**五片画面达到接受标准，但存在博彩叠字，按“有广告但高清可接受”的要求可保留。不要把它宣传成无广告来源。
3. **猫眼、360、MXdm**有低清与广告同时存在的样本，不适合作为整源推荐。猫眼并非每片都差：三片清晰，但《飞驰人生2》720p加整屏广告、《独行月球》960×402加广告；用户此前反馈的《食神》不重复计入本次五片。
4. **无尽**五片均能出画、抽样未见博彩广告，但都低于本轮1080档目标。不能因无广告而给高清标签。**最大**另有竖画布嵌小电影的问题，1080这个数字没有代表有效画面。
5. **非凡五片、电影天堂三片**首轮被本站拦截，现已发布修复并全部出画。两源各五片达到本轮清晰度标准，采样未见博彩广告，可优先保留。
6. 索尼、樱花、xfdmneo五次搜索503；aafun五次解析503；mutefun五次解析504。其余零散失败见下表，均未凭猜测标成低清或广告多。

## 逐片明细（105条）

播放链接是当时实际取得的线上片目；抽帧链接指向本机保留的证据。动漫条目是该作品所选的一集，并非看完五部全剧。

### 非凡资源 / ffzy

| 样本 | 实际片目 / 集数 / 线路 | 尺寸 / 时长 | 实播结果 | 画质观察 | 广告观察 | 证据 |
|---|---|---|---|---|---|---|
| 1 | [流浪地球2](https://tv.yuelili.com/watch/ffzy/32627/line1-0) / HD国语 / ffm3u8 | 1920×800 / 174.6分 | 部署修复后，三点出画 | 画面清晰，满足本轮目标 | 采样点未见明显广告 | [35秒](D:/coding/yueli-official/task-artifacts/ytv/deploy-playback-policy-20261024-06/ffzy-0-opening.jpg)、[4715秒](D:/coding/yueli-official/task-artifacts/ytv/deploy-playback-policy-20261024-06/ffzy-0-middle.jpg)、[8906秒](D:/coding/yueli-official/task-artifacts/ytv/deploy-playback-policy-20261024-06/ffzy-0-late.jpg) |
| 2 | [飞驰人生2](https://tv.yuelili.com/watch/ffzy/64816/line1-0) / HD国语 / ffm3u8 | 1920×1080 / 122.2分 | 部署修复后，三点出画 | 画面清晰，满足本轮目标 | 采样点未见明显广告 | [35秒](D:/coding/yueli-official/task-artifacts/ytv/deploy-playback-policy-20261024-06/ffzy-1-opening.jpg)、[3300秒](D:/coding/yueli-official/task-artifacts/ytv/deploy-playback-policy-20261024-06/ffzy-1-middle.jpg)、[6233秒](D:/coding/yueli-official/task-artifacts/ytv/deploy-playback-policy-20261024-06/ffzy-1-late.jpg) |
| 3 | [头脑特工队2](https://tv.yuelili.com/watch/ffzy/70551/line1-0) / HD中字 / ffm3u8 | 1918×802 / 97.6分 | 部署修复后，三点出画 | 画面清晰，满足本轮目标 | 采样点未见明显广告 | [35秒](D:/coding/yueli-official/task-artifacts/ytv/deploy-playback-policy-20261024-06/ffzy-2-opening.jpg)、[2634秒](D:/coding/yueli-official/task-artifacts/ytv/deploy-playback-policy-20261024-06/ffzy-2-middle.jpg)、[4976秒](D:/coding/yueli-official/task-artifacts/ytv/deploy-playback-policy-20261024-06/ffzy-2-late.jpg) |
| 4 | [蜘蛛侠：纵横宇宙](https://tv.yuelili.com/watch/ffzy/54978/line1-0) / HD中字 / ffm3u8 | 1920×1080 / 142.7分 | 部署修复后，三点出画 | 画面清晰，满足本轮目标 | 采样点未见明显广告 | [35秒](D:/coding/yueli-official/task-artifacts/ytv/deploy-playback-policy-20261024-06/ffzy-3-opening.jpg)、[3852秒](D:/coding/yueli-official/task-artifacts/ytv/deploy-playback-policy-20261024-06/ffzy-3-middle.jpg)、[7276秒](D:/coding/yueli-official/task-artifacts/ytv/deploy-playback-policy-20261024-06/ffzy-3-late.jpg) |
| 5 | [哪吒之魔童降世](https://tv.yuelili.com/watch/ffzy/11201/line1-0) / HD国语 / ffm3u8 | 1920×808 / 111.2分 | 部署修复后，三点出画 | 画面清晰，满足本轮目标 | 采样点未见明显广告 | [35秒](D:/coding/yueli-official/task-artifacts/ytv/deploy-playback-policy-20261024-06/ffzy-4-opening.jpg)、[3003秒](D:/coding/yueli-official/task-artifacts/ytv/deploy-playback-policy-20261024-06/ffzy-4-middle.jpg)、[5672秒](D:/coding/yueli-official/task-artifacts/ytv/deploy-playback-policy-20261024-06/ffzy-4-late.jpg) |

### 电影天堂 / dytt

| 样本 | 实际片目 / 集数 / 线路 | 尺寸 / 时长 | 实播结果 | 画质观察 | 广告观察 | 证据 |
|---|---|---|---|---|---|---|
| 1 | [流浪地球2](https://tv.yuelili.com/watch/dytt/40510/0) / HD国语 / dyttm3u8 | 1920×806 / 174.4分 | 部署修复后，三点出画 | 画面清晰，满足本轮目标 | 采样点未见明显广告 | [35秒](D:/coding/yueli-official/task-artifacts/ytv/deploy-playback-policy-20261024-06/dytt-0-opening.jpg)、[4707秒](D:/coding/yueli-official/task-artifacts/ytv/deploy-playback-policy-20261024-06/dytt-0-middle.jpg)、[8892秒](D:/coding/yueli-official/task-artifacts/ytv/deploy-playback-policy-20261024-06/dytt-0-late.jpg) |
| 2 | [飞驰人生2](https://tv.yuelili.com/watch/dytt/39412/0) / HD国语 / dyttm3u8 | 1920×814 / 122.2分 | 部署修复后，三点出画 | 画面清晰，满足本轮目标 | 采样点未见明显广告 | [35秒](D:/coding/yueli-official/task-artifacts/ytv/deploy-playback-policy-20261024-06/dytt-1-opening.jpg)、[3298秒](D:/coding/yueli-official/task-artifacts/ytv/deploy-playback-policy-20261024-06/dytt-1-middle.jpg)、[6230秒](D:/coding/yueli-official/task-artifacts/ytv/deploy-playback-policy-20261024-06/dytt-1-late.jpg) |
| 3 | [沙丘2](https://tv.yuelili.com/watch/dytt/14659/0) / HD中字 / dyttm3u8 | 1920×830 / 166.9分 | 三点出画 | 画面清晰，满足本轮目标 | 采样点未见明显广告 | [35秒](D:/coding/yueli-official/task-artifacts/ytv/online-five-20261024/dytt-2-opening.jpg)、[4506秒](D:/coding/yueli-official/task-artifacts/ytv/online-five-20261024/dytt-2-middle.jpg)、[8511秒](D:/coding/yueli-official/task-artifacts/ytv/online-five-20261024/dytt-2-late.jpg) |
| 4 | [蜘蛛侠：纵横宇宙](https://tv.yuelili.com/watch/dytt/6125/0) / HD中字 / dyttm3u8 | 1920×800 / 141.3分 | 部署修复后，三点出画 | 画面清晰，满足本轮目标 | 采样点未见明显广告 | [35秒](D:/coding/yueli-official/task-artifacts/ytv/deploy-playback-policy-20261024-06/dytt-3-opening.jpg)、[3813秒](D:/coding/yueli-official/task-artifacts/ytv/deploy-playback-policy-20261024-06/dytt-3-middle.jpg)、[7204秒](D:/coding/yueli-official/task-artifacts/ytv/deploy-playback-policy-20261024-06/dytt-3-late.jpg) |
| 5 | [哪吒之魔童降世](https://tv.yuelili.com/watch/dytt/2377/0) / HD国语 / dyttm3u8 | 1920×803 / 111.2分 | 三点出画 | 画面清晰，满足本轮目标 | 采样点未见明显广告 | [35秒](D:/coding/yueli-official/task-artifacts/ytv/online-five-20261024/dytt-4-opening.jpg)、[3001秒](D:/coding/yueli-official/task-artifacts/ytv/online-five-20261024/dytt-4-middle.jpg)、[5669秒](D:/coding/yueli-official/task-artifacts/ytv/online-five-20261024/dytt-4-late.jpg) |

### 猫眼资源 / maoyan

| 样本 | 实际片目 / 集数 / 线路 | 尺寸 / 时长 | 实播结果 | 画质观察 | 广告观察 | 证据 |
|---|---|---|---|---|---|---|
| 1 | [流浪地球2](https://tv.yuelili.com/watch/maoyan/137597/0) / 正片 / mym3u8 | 3840×1608 / 176.3分 | 三点出画 | 画面清晰，满足本轮目标 | 开头博彩叠字 | [35秒](D:/coding/yueli-official/task-artifacts/ytv/online-five-20261024/maoyan-0-opening.jpg)、[4759秒](D:/coding/yueli-official/task-artifacts/ytv/online-five-20261024/maoyan-0-middle.jpg)、[8989秒](D:/coding/yueli-official/task-artifacts/ytv/online-five-20261024/maoyan-0-late.jpg) |
| 2 | [飞驰人生2](https://tv.yuelili.com/watch/maoyan/135775/0) / 正片 / mym3u8 | 1280×720 / 123.4分 | 三点出画 | 画面低于本轮1080档目标 | 开头整屏博彩广告 | [35秒](D:/coding/yueli-official/task-artifacts/ytv/online-five-20261024/maoyan-1-opening.jpg)、[3332秒](D:/coding/yueli-official/task-artifacts/ytv/online-five-20261024/maoyan-1-middle.jpg)、[6294秒](D:/coding/yueli-official/task-artifacts/ytv/online-five-20261024/maoyan-1-late.jpg) |
| 3 | [独行月球](https://tv.yuelili.com/watch/maoyan/153961/0) / 正片 / mym3u8 | 960×402 / 123.9分 | 出画，抽帧不完整 | 画面低于本轮1080档目标；后段拖动超时，仅两帧 | 开头博彩叠字 | [35秒](D:/coding/yueli-official/task-artifacts/ytv/online-five-20261024/maoyan-2-opening.jpg)、[3344秒](D:/coding/yueli-official/task-artifacts/ytv/online-five-20261024/maoyan-2-middle.jpg) |
| 4 | [流浪地球](https://tv.yuelili.com/watch/maoyan/123736/0) / 正片 / mym3u8 | 3840×1504 / 127.7分 | 三点出画 | 画面清晰，满足本轮目标 | 开头博彩叠字 | [35秒](D:/coding/yueli-official/task-artifacts/ytv/online-five-20261024/maoyan-3-opening.jpg)、[3447秒](D:/coding/yueli-official/task-artifacts/ytv/online-five-20261024/maoyan-3-middle.jpg)、[6511秒](D:/coding/yueli-official/task-artifacts/ytv/online-five-20261024/maoyan-3-late.jpg) |
| 5 | [哪吒之魔童降世](https://tv.yuelili.com/watch/maoyan/125781/0) / HD国语 / mym3u8 | 1920×808 / 111.8分 | 三点出画 | 画面清晰，满足本轮目标 | 采样点未见明显广告 | [35秒](D:/coding/yueli-official/task-artifacts/ytv/online-five-20261024/maoyan-4-opening.jpg)、[3017秒](D:/coding/yueli-official/task-artifacts/ytv/online-five-20261024/maoyan-4-middle.jpg)、[5699秒](D:/coding/yueli-official/task-artifacts/ytv/online-five-20261024/maoyan-4-late.jpg) |

### 爱奇艺资源 / iqiyi

| 样本 | 实际片目 / 集数 / 线路 | 尺寸 / 时长 | 实播结果 | 画质观察 | 广告观察 | 证据 |
|---|---|---|---|---|---|---|
| 1 | [流浪地球2](https://tv.yuelili.com/watch/iqiyi/35125/0) / 正片 / iqym3u8 | — | 解析：503 ytv.source_unavailable | 未出画，画质未审 | 未审 | 无画面证据 |
| 2 | [飞驰人生2](https://tv.yuelili.com/watch/iqiyi/20246/0) / 正片 / iqym3u8 | 1920×822 / 121.5分 | 三点出画 | 画面清晰，满足本轮目标 | 开头及后段见博彩叠字 | [35秒](D:/coding/yueli-official/task-artifacts/ytv/online-five-20261024/iqiyi-1-opening.jpg)、[3280秒](D:/coding/yueli-official/task-artifacts/ytv/online-five-20261024/iqiyi-1-middle.jpg)、[6195秒](D:/coding/yueli-official/task-artifacts/ytv/online-five-20261024/iqiyi-1-late.jpg) |
| 3 | [头脑特工队2](https://tv.yuelili.com/watch/iqiyi/22485/0) / 正片 / iqym3u8 | 1920×1080 / 96.6分 | 三点出画 | 画面清晰，满足本轮目标 | 开头博彩叠字 | [35秒](D:/coding/yueli-official/task-artifacts/ytv/online-five-20261024/iqiyi-2-opening.jpg)、[2609秒](D:/coding/yueli-official/task-artifacts/ytv/online-five-20261024/iqiyi-2-middle.jpg)、[4928秒](D:/coding/yueli-official/task-artifacts/ytv/online-five-20261024/iqiyi-2-late.jpg) |
| 4 | 蜘蛛侠：纵横宇宙 | — | 搜索：503 ytv.source_unavailable | 未出画，画质未审 | 未审 | 无画面证据 |
| 5 | 哪吒之魔童降世 | — | 搜索：503 ytv.source_unavailable | 未出画，画质未审 | 未审 | 无画面证据 |

### 无尽资源 / wujin

| 样本 | 实际片目 / 集数 / 线路 | 尺寸 / 时长 | 实播结果 | 画质观察 | 广告观察 | 证据 |
|---|---|---|---|---|---|---|
| 1 | [流浪地球2](https://tv.yuelili.com/watch/wujin/49411/0) / HD中字 / wjm3u8 | 1280×536 / 173.2分 | 三点出画 | 画面低于本轮1080档目标 | 采样点未见明显广告 | [35秒](D:/coding/yueli-official/task-artifacts/ytv/online-five-20261024/wujin-0-opening.jpg)、[4676秒](D:/coding/yueli-official/task-artifacts/ytv/online-five-20261024/wujin-0-middle.jpg)、[8832秒](D:/coding/yueli-official/task-artifacts/ytv/online-five-20261024/wujin-0-late.jpg) |
| 2 | [飞驰人生2](https://tv.yuelili.com/watch/wujin/61126/0) / HD / wjm3u8 | 1280×720 / 121.0分 | 三点出画 | 画面低于本轮1080档目标 | 采样点未见明显广告 | [35秒](D:/coding/yueli-official/task-artifacts/ytv/online-five-20261024/wujin-1-opening.jpg)、[3266秒](D:/coding/yueli-official/task-artifacts/ytv/online-five-20261024/wujin-1-middle.jpg)、[6170秒](D:/coding/yueli-official/task-artifacts/ytv/online-five-20261024/wujin-1-late.jpg) |
| 3 | [头脑特工队2](https://tv.yuelili.com/watch/wujin/115400/0) / 国语 / wjm3u8 | 1720×720 / 96.4分 | 三点出画 | 画面低于本轮1080档目标 | 采样点未见明显广告 | [35秒](D:/coding/yueli-official/task-artifacts/ytv/online-five-20261024/wujin-2-opening.jpg)、[2603秒](D:/coding/yueli-official/task-artifacts/ytv/online-five-20261024/wujin-2-middle.jpg)、[4917秒](D:/coding/yueli-official/task-artifacts/ytv/online-five-20261024/wujin-2-late.jpg) |
| 4 | [蜘蛛侠：纵横宇宙](https://tv.yuelili.com/watch/wujin/87350/0) / HD中字 / wjm3u8 | 1280×536 / 133.0分 | 三点出画 | 画面低于本轮1080档目标 | 采样点未见明显广告 | [35秒](D:/coding/yueli-official/task-artifacts/ytv/online-five-20261024/wujin-3-opening.jpg)、[3590秒](D:/coding/yueli-official/task-artifacts/ytv/online-five-20261024/wujin-3-middle.jpg)、[6782秒](D:/coding/yueli-official/task-artifacts/ytv/online-five-20261024/wujin-3-late.jpg) |
| 5 | [哪吒之魔童降世](https://tv.yuelili.com/watch/wujin/89904/0) / HD中字 / wjm3u8 | 1710×720 / 110.0分 | 三点出画 | 画面低于本轮1080档目标 | 采样点未见明显广告 | [35秒](D:/coding/yueli-official/task-artifacts/ytv/online-five-20261024/wujin-4-opening.jpg)、[2969秒](D:/coding/yueli-official/task-artifacts/ytv/online-five-20261024/wujin-4-middle.jpg)、[5609秒](D:/coding/yueli-official/task-artifacts/ytv/online-five-20261024/wujin-4-late.jpg) |

### 索尼资源 / qgsuoni

| 样本 | 实际片目 / 集数 / 线路 | 尺寸 / 时长 | 实播结果 | 画质观察 | 广告观察 | 证据 |
|---|---|---|---|---|---|---|
| 1 | 流浪地球2 | — | 搜索：503 ytv.source_unavailable | 未出画，画质未审 | 未审 | 无画面证据 |
| 2 | 飞驰人生2 | — | 搜索：503 ytv.source_unavailable | 未出画，画质未审 | 未审 | 无画面证据 |
| 3 | 头脑特工队2 | — | 搜索：503 ytv.source_unavailable | 未出画，画质未审 | 未审 | 无画面证据 |
| 4 | 蜘蛛侠：纵横宇宙 | — | 搜索：503 ytv.source_unavailable | 未出画，画质未审 | 未审 | 无画面证据 |
| 5 | 哪吒之魔童降世 | — | 搜索：503 ytv.source_unavailable | 未出画，画质未审 | 未审 | 无画面证据 |

### 樱花资源 / yinghua

| 样本 | 实际片目 / 集数 / 线路 | 尺寸 / 时长 | 实播结果 | 画质观察 | 广告观察 | 证据 |
|---|---|---|---|---|---|---|
| 1 | 流浪地球2 | — | 搜索：503 ytv.source_unavailable | 未出画，画质未审 | 未审 | 无画面证据 |
| 2 | 飞驰人生2 | — | 搜索：503 ytv.source_unavailable | 未出画，画质未审 | 未审 | 无画面证据 |
| 3 | 头脑特工队2 | — | 搜索：503 ytv.source_unavailable | 未出画，画质未审 | 未审 | 无画面证据 |
| 4 | 蜘蛛侠：纵横宇宙 | — | 搜索：503 ytv.source_unavailable | 未出画，画质未审 | 未审 | 无画面证据 |
| 5 | 哪吒之魔童降世 | — | 搜索：503 ytv.source_unavailable | 未出画，画质未审 | 未审 | 无画面证据 |

### 最大资源 / zuid

| 样本 | 实际片目 / 集数 / 线路 | 尺寸 / 时长 | 实播结果 | 画质观察 | 广告观察 | 证据 |
|---|---|---|---|---|---|---|
| 1 | [流浪地球2](https://tv.yuelili.com/watch/zuid/7046/0) / HD中字 / zuidam3u8 | 1280×536 / 173.7分 | 三点出画 | 画面低于本轮1080档目标 | 采样点未见明显广告 | [35秒](D:/coding/yueli-official/task-artifacts/ytv/online-five-20261024/zuid-0-opening.jpg)、[4689秒](D:/coding/yueli-official/task-artifacts/ytv/online-five-20261024/zuid-0-middle.jpg)、[8857秒](D:/coding/yueli-official/task-artifacts/ytv/online-five-20261024/zuid-0-late.jpg) |
| 2 | [飞驰人生2](https://tv.yuelili.com/watch/zuid/63395/0) / TC / zuidam3u8 | 1280×568 / 116.5分 | 三点出画 | 画面低于本轮1080档目标 | 后段顶部VPN广告 | [35秒](D:/coding/yueli-official/task-artifacts/ytv/online-five-20261024/zuid-1-opening.jpg)、[3146秒](D:/coding/yueli-official/task-artifacts/ytv/online-five-20261024/zuid-1-middle.jpg)、[5943秒](D:/coding/yueli-official/task-artifacts/ytv/online-five-20261024/zuid-1-late.jpg) |
| 3 | [头脑特工队2](https://tv.yuelili.com/watch/zuid/77529/0) / HD / zuidam3u8 | — | 媒体加载：23秒未开始出画 | 未出画，画质未审 | 未审 | 无画面证据 |
| 4 | [蜘蛛侠：纵横宇宙](https://tv.yuelili.com/watch/zuid/29939/0) / HD中字 / zuidam3u8 | 1080×1921 / 141.4分 | 三点出画 | 1080×1921竖画布内嵌小幅横屏电影；有效画面不达标 | 采样点未见明显广告 | [35秒](D:/coding/yueli-official/task-artifacts/ytv/online-five-20261024/zuid-3-opening.jpg)、[3818秒](D:/coding/yueli-official/task-artifacts/ytv/online-five-20261024/zuid-3-middle.jpg)、[7213秒](D:/coding/yueli-official/task-artifacts/ytv/online-five-20261024/zuid-3-late.jpg) |
| 5 | [哪吒之魔童降世](https://tv.yuelili.com/watch/zuid/26373/0) / HD中字 / zuidam3u8 | 1920×808 / 110.6分 | 三点出画 | 画面清晰，满足本轮目标 | 采样点未见明显广告 | [35秒](D:/coding/yueli-official/task-artifacts/ytv/online-five-20261024/zuid-4-opening.jpg)、[2987秒](D:/coding/yueli-official/task-artifacts/ytv/online-five-20261024/zuid-4-middle.jpg)、[5642秒](D:/coding/yueli-official/task-artifacts/ytv/online-five-20261024/zuid-4-late.jpg) |

### 魔都资源 / mdzy

| 样本 | 实际片目 / 集数 / 线路 | 尺寸 / 时长 | 实播结果 | 画质观察 | 广告观察 | 证据 |
|---|---|---|---|---|---|---|
| 1 | [流浪地球2](https://tv.yuelili.com/watch/mdzy/35125/0) / 正片 / modum3u8 | 3840×1608 / 176.3分 | 三点出画 | 画面清晰，满足本轮目标 | 开头博彩叠字 | [35秒](D:/coding/yueli-official/task-artifacts/ytv/online-five-20261024/mdzy-0-opening.jpg)、[4760秒](D:/coding/yueli-official/task-artifacts/ytv/online-five-20261024/mdzy-0-middle.jpg)、[8992秒](D:/coding/yueli-official/task-artifacts/ytv/online-five-20261024/mdzy-0-late.jpg) |
| 2 | [飞驰人生2](https://tv.yuelili.com/watch/mdzy/20246/0) / 正片 / modum3u8 | 1920×822 / 123.8分 | 三点出画 | 画面清晰，满足本轮目标 | 开头博彩叠字 | [35秒](D:/coding/yueli-official/task-artifacts/ytv/online-five-20261024/mdzy-1-opening.jpg)、[3343秒](D:/coding/yueli-official/task-artifacts/ytv/online-five-20261024/mdzy-1-middle.jpg)、[6314秒](D:/coding/yueli-official/task-artifacts/ytv/online-five-20261024/mdzy-1-late.jpg) |
| 3 | [头脑特工队2](https://tv.yuelili.com/watch/mdzy/22485/0) / 正片 / modum3u8 | 1920×1080 / 98.1分 | 三点出画 | 画面清晰，满足本轮目标 | 开头博彩叠字 | [35秒](D:/coding/yueli-official/task-artifacts/ytv/online-five-20261024/mdzy-2-opening.jpg)、[2648秒](D:/coding/yueli-official/task-artifacts/ytv/online-five-20261024/mdzy-2-middle.jpg)、[5003秒](D:/coding/yueli-official/task-artifacts/ytv/online-five-20261024/mdzy-2-late.jpg) |
| 4 | [蜘蛛侠：纵横宇宙](https://tv.yuelili.com/watch/mdzy/51105/0) / 正片 / modum3u8 | 1920×1080 / 141.8分 | 三点出画 | 画面清晰，满足本轮目标 | 开头及中段博彩叠字 | [35秒](D:/coding/yueli-official/task-artifacts/ytv/online-five-20261024/mdzy-3-opening.jpg)、[3828秒](D:/coding/yueli-official/task-artifacts/ytv/online-five-20261024/mdzy-3-middle.jpg)、[7231秒](D:/coding/yueli-official/task-artifacts/ytv/online-five-20261024/mdzy-3-late.jpg) |
| 5 | [哪吒之魔童降世](https://tv.yuelili.com/watch/mdzy/6629/0) / HD高清 / modum3u8 | 1920×804 / 116.7分 | 三点出画 | 画面清晰，满足本轮目标 | 开头博彩叠字 | [35秒](D:/coding/yueli-official/task-artifacts/ytv/online-five-20261024/mdzy-4-opening.jpg)、[3151秒](D:/coding/yueli-official/task-artifacts/ytv/online-five-20261024/mdzy-4-middle.jpg)、[5953秒](D:/coding/yueli-official/task-artifacts/ytv/online-five-20261024/mdzy-4-late.jpg) |

### 360资源 / zy360

| 样本 | 实际片目 / 集数 / 线路 | 尺寸 / 时长 | 实播结果 | 画质观察 | 广告观察 | 证据 |
|---|---|---|---|---|---|---|
| 1 | [流浪地球2](https://tv.yuelili.com/watch/zy360/71142/0) / 流浪地球2 / 360zy | 1280×538 / 176.1分 | 三点出画 | 画面低于本轮1080档目标 | 开头博彩叠字 | [35秒](D:/coding/yueli-official/task-artifacts/ytv/online-five-20261024/zy360-0-opening.jpg)、[4754秒](D:/coding/yueli-official/task-artifacts/ytv/online-five-20261024/zy360-0-middle.jpg)、[8981秒](D:/coding/yueli-official/task-artifacts/ytv/online-five-20261024/zy360-0-late.jpg) |
| 2 | [飞驰人生2](https://tv.yuelili.com/watch/zy360/60384/0) / 飞驰人生2 / 360zy | 1920×800 / 123.6分 | 三点出画 | 画面清晰，满足本轮目标 | 开头博彩叠字 | [35秒](D:/coding/yueli-official/task-artifacts/ytv/online-five-20261024/zy360-1-opening.jpg)、[3338秒](D:/coding/yueli-official/task-artifacts/ytv/online-five-20261024/zy360-1-middle.jpg)、[6305秒](D:/coding/yueli-official/task-artifacts/ytv/online-five-20261024/zy360-1-late.jpg) |
| 3 | [头脑特工队2](https://tv.yuelili.com/watch/zy360/64433/0) / 头脑特工队2 / 360zy | 1282×536 / 98.4分 | 三点出画 | 画面低于本轮1080档目标 | 开头博彩叠字 | [35秒](D:/coding/yueli-official/task-artifacts/ytv/online-five-20261024/zy360-2-opening.jpg)、[2656秒](D:/coding/yueli-official/task-artifacts/ytv/online-five-20261024/zy360-2-middle.jpg)、[5016秒](D:/coding/yueli-official/task-artifacts/ytv/online-five-20261024/zy360-2-late.jpg) |
| 4 | [蜘蛛侠：纵横宇宙](https://tv.yuelili.com/watch/zy360/61217/0) / 蜘蛛侠：纵横宇宙 / 360zy | 1724×720 / 118.9分 | 三点出画 | 画面低于本轮1080档目标 | 开头博彩叠字 | [35秒](D:/coding/yueli-official/task-artifacts/ytv/online-five-20261024/zy360-3-opening.jpg)、[3209秒](D:/coding/yueli-official/task-artifacts/ytv/online-five-20261024/zy360-3-middle.jpg)、[6062秒](D:/coding/yueli-official/task-artifacts/ytv/online-five-20261024/zy360-3-late.jpg) |
| 5 | [哪吒之魔童降世](https://tv.yuelili.com/watch/zy360/12102/0) / 哪吒之魔童降世 / 360zy | 1024×428 / 111.5分 | 三点出画 | 画面低于本轮1080档目标 | 开头博彩叠字 | [35秒](D:/coding/yueli-official/task-artifacts/ytv/online-five-20261024/zy360-4-opening.jpg)、[3009秒](D:/coding/yueli-official/task-artifacts/ytv/online-five-20261024/zy360-4-middle.jpg)、[5684秒](D:/coding/yueli-official/task-artifacts/ytv/online-five-20261024/zy360-4-late.jpg) |

### 7sefun / kazumi-7sefun

| 样本 | 实际片目 / 集数 / 线路 | 尺寸 / 时长 | 实播结果 | 画质观察 | 广告观察 | 证据 |
|---|---|---|---|---|---|---|
| 1 | [葬送的芙莉莲](https://tv.yuelili.com/watch/kazumi-7sefun/aHR0cHM6Ly93d3cuN3NlZnVuLnRvcC92b2RkZXRhaWwvMzI2NDQuaHRtbA/aHR0cHM6Ly93d3cuN3NlZnVuLnRvcC92b2RwbGF5LzMyNjQ0LTEtMS5odG1s) / 第01集 / 线路 1 | 1920×1080 / 26.0分 | 三点出画 | 画面清晰，满足本轮目标 | 采样点未见明显广告 | [35秒](D:/coding/yueli-official/task-artifacts/ytv/online-five-20261024/kazumi-7sefun-0-opening.jpg)、[701秒](D:/coding/yueli-official/task-artifacts/ytv/online-five-20261024/kazumi-7sefun-0-middle.jpg)、[1325秒](D:/coding/yueli-official/task-artifacts/ytv/online-five-20261024/kazumi-7sefun-0-late.jpg) |
| 2 | [鬼灭之刃](https://tv.yuelili.com/watch/kazumi-7sefun/aHR0cHM6Ly93d3cuN3NlZnVuLnRvcC92b2RkZXRhaWwvMzAzMzMuaHRtbA/aHR0cHM6Ly93d3cuN3NlZnVuLnRvcC92b2RwbGF5LzMwMzMzLTEtMS5odG1s) / 第01集 / 线路 1 | 1920×1080 / 23.7分 | 三点出画 | 画面清晰，满足本轮目标 | 采样点未见明显广告 | [35秒](D:/coding/yueli-official/task-artifacts/ytv/online-five-20261024/kazumi-7sefun-1-opening.jpg)、[639秒](D:/coding/yueli-official/task-artifacts/ytv/online-five-20261024/kazumi-7sefun-1-middle.jpg)、[1207秒](D:/coding/yueli-official/task-artifacts/ytv/online-five-20261024/kazumi-7sefun-1-late.jpg) |
| 3 | [咒术回战](https://tv.yuelili.com/watch/kazumi-7sefun/aHR0cHM6Ly93d3cuN3NlZnVuLnRvcC92b2RkZXRhaWwvMzE4NzAuaHRtbA/aHR0cHM6Ly93d3cuN3NlZnVuLnRvcC92b2RwbGF5LzMxODcwLTEtMS5odG1s) / 第01集 / 线路 1 | 1920×1080 / 23.9分 | 三点出画 | 画面清晰，满足本轮目标 | 采样点未见明显广告 | [35秒](D:/coding/yueli-official/task-artifacts/ytv/online-five-20261024/kazumi-7sefun-2-opening.jpg)、[645秒](D:/coding/yueli-official/task-artifacts/ytv/online-five-20261024/kazumi-7sefun-2-middle.jpg)、[1219秒](D:/coding/yueli-official/task-artifacts/ytv/online-five-20261024/kazumi-7sefun-2-late.jpg) |
| 4 | [间谍过家家](https://tv.yuelili.com/watch/kazumi-7sefun/aHR0cHM6Ly93d3cuN3NlZnVuLnRvcC92b2RkZXRhaWwvMjk2MjQuaHRtbA/aHR0cHM6Ly93d3cuN3NlZnVuLnRvcC92b2RwbGF5LzI5NjI0LTEtMS5odG1s) / 第01集 / 线路 1 | 1920×1080 / 24.2分 | 三点出画 | 画面清晰，满足本轮目标 | 采样点未见明显广告 | [35秒](D:/coding/yueli-official/task-artifacts/ytv/online-five-20261024/kazumi-7sefun-3-opening.jpg)、[653秒](D:/coding/yueli-official/task-artifacts/ytv/online-five-20261024/kazumi-7sefun-3-middle.jpg)、[1234秒](D:/coding/yueli-official/task-artifacts/ytv/online-five-20261024/kazumi-7sefun-3-late.jpg) |
| 5 | [电锯人](https://tv.yuelili.com/watch/kazumi-7sefun/aHR0cHM6Ly93d3cuN3NlZnVuLnRvcC92b2RkZXRhaWwvMjk1MTcuaHRtbA/aHR0cHM6Ly93d3cuN3NlZnVuLnRvcC92b2RwbGF5LzI5NTE3LTEtMS5odG1s) / 第01集 / 线路 1 | 1920×1080 / 25.4分 | 三点出画 | 画面清晰，满足本轮目标 | 采样点未见明显广告 | [35秒](D:/coding/yueli-official/task-artifacts/ytv/online-five-20261024/kazumi-7sefun-4-opening.jpg)、[686秒](D:/coding/yueli-official/task-artifacts/ytv/online-five-20261024/kazumi-7sefun-4-middle.jpg)、[1296秒](D:/coding/yueli-official/task-artifacts/ytv/online-five-20261024/kazumi-7sefun-4-late.jpg) |

### aafun / kazumi-aafun

| 样本 | 实际片目 / 集数 / 线路 | 尺寸 / 时长 | 实播结果 | 画质观察 | 广告观察 | 证据 |
|---|---|---|---|---|---|---|
| 1 | [葬送的芙莉莲](https://tv.yuelili.com/watch/kazumi-aafun/aHR0cHM6Ly93d3cubW9vbmNpLmNvbS9hbmltZS8yMjguaHRtbA/aHR0cHM6Ly93d3cubW9vbmNpLmNvbS9hbmltZS8yMjgvcGxheS8yLTEuaHRtbA) / 第01话 / 线路 1 | — | 解析：503 ytv.source_unavailable | 未出画，画质未审 | 未审 | 无画面证据 |
| 2 | [鬼灭之刃](https://tv.yuelili.com/watch/kazumi-aafun/aHR0cHM6Ly93d3cubW9vbmNpLmNvbS9hbmltZS8yMjc0Lmh0bWw/aHR0cHM6Ly93d3cubW9vbmNpLmNvbS9hbmltZS8yMjc0L3BsYXkvMS0xLmh0bWw) / 第01话 / 线路 1 | — | 解析：503 ytv.source_unavailable | 未出画，画质未审 | 未审 | 无画面证据 |
| 3 | [咒术回战](https://tv.yuelili.com/watch/kazumi-aafun/aHR0cHM6Ly93d3cubW9vbmNpLmNvbS9hbmltZS8xMzc2Lmh0bWw/aHR0cHM6Ly93d3cubW9vbmNpLmNvbS9hbmltZS8xMzc2L3BsYXkvMS0xLmh0bWw) / 第01话 / 线路 1 | — | 解析：503 ytv.source_unavailable | 未出画，画质未审 | 未审 | 无画面证据 |
| 4 | [间谍过家家](https://tv.yuelili.com/watch/kazumi-aafun/aHR0cHM6Ly93d3cubW9vbmNpLmNvbS9hbmltZS8xMzY2Lmh0bWw/aHR0cHM6Ly93d3cubW9vbmNpLmNvbS9hbmltZS8xMzY2L3BsYXkvMS0xLmh0bWw) / 第01话 / 线路 1 | — | 解析：503 ytv.source_unavailable | 未出画，画质未审 | 未审 | 无画面证据 |
| 5 | [链锯人](https://tv.yuelili.com/watch/kazumi-aafun/aHR0cHM6Ly93d3cubW9vbmNpLmNvbS9hbmltZS8xNDAzLmh0bWw/aHR0cHM6Ly93d3cubW9vbmNpLmNvbS9hbmltZS8xNDAzL3BsYXkvMS0xLmh0bWw) / 第01话 / 线路 1 | — | 解析：503 ytv.source_unavailable | 未出画，画质未审 | 未审 | 无画面证据 |

### akianime / kazumi-akianime

| 样本 | 实际片目 / 集数 / 线路 | 尺寸 / 时长 | 实播结果 | 画质观察 | 广告观察 | 证据 |
|---|---|---|---|---|---|---|
| 1 | [葬送的芙莉莲](https://tv.yuelili.com/watch/kazumi-akianime/aHR0cHM6Ly93d3cuYWtpYW5pbWUuY2MvYmdtcGxheS9kUEREREUtMS0xLmh0bWw/aHR0cHM6Ly93d3cuYWtpYW5pbWUuY2MvYmdtcGxheS9kUEREREUtNS0xLmh0bWw) / 01 / 线路 1 | 1920×1080 / 26.0分 | 三点出画 | 画面清晰，满足本轮目标 | 采样点未见明显广告 | [35秒](D:/coding/yueli-official/task-artifacts/ytv/online-five-20261024/kazumi-akianime-0-opening.jpg)、[702秒](D:/coding/yueli-official/task-artifacts/ytv/online-five-20261024/kazumi-akianime-0-middle.jpg)、[1326秒](D:/coding/yueli-official/task-artifacts/ytv/online-five-20261024/kazumi-akianime-0-late.jpg) |
| 2 | [鬼灭之刃](https://tv.yuelili.com/watch/kazumi-akianime/aHR0cHM6Ly93d3cuYWtpYW5pbWUuY2MvYmdtcGxheS9BY0REREUtMS0xLmh0bWw/aHR0cHM6Ly93d3cuYWtpYW5pbWUuY2MvYmdtcGxheS9BY0REREUtNC0xLmh0bWw) / 第01集 / 线路 1 | — | 媒体加载：23秒未开始出画 | 未出画，画质未审 | 未审 | 无画面证据 |
| 3 | [咒术回战](https://tv.yuelili.com/watch/kazumi-akianime/aHR0cHM6Ly93d3cuYWtpYW5pbWUuY2MvYmdtcGxheS9JY0REREUtMS0xLmh0bWw/aHR0cHM6Ly93d3cuYWtpYW5pbWUuY2MvYmdtcGxheS9JY0REREUtMi0xLmh0bWw) / 第01集 / 线路 1 | — | 解析：503 ytv.source_unavailable | 未出画，画质未审 | 未审 | 无画面证据 |
| 4 | [间谍过家家](https://tv.yuelili.com/watch/kazumi-akianime/aHR0cHM6Ly93d3cuYWtpYW5pbWUuY2MvYmdtcGxheS83UkREREUtMS0xLmh0bWw/aHR0cHM6Ly93d3cuYWtpYW5pbWUuY2MvYmdtcGxheS83UkREREUtMy0xLmh0bWw) / 第1集 / 线路 1 | — | 解析：503 ytv.source_unavailable | 未出画，画质未审 | 未审 | 无画面证据 |
| 5 | [进击的巨人 最终季 Part.2](https://tv.yuelili.com/watch/kazumi-akianime/aHR0cHM6Ly93d3cuYWtpYW5pbWUuY2MvYmdtcGxheS9yaGNEREUtMS0xLmh0bWw/aHR0cHM6Ly93d3cuYWtpYW5pbWUuY2MvYmdtcGxheS9yaGNEREUtMS0xLmh0bWw) / 17 / 线路 1 | 1920×1080 / 23.9分 | 三点出画 | 画面清晰，满足本轮目标 | 采样点未见明显广告 | [35秒](D:/coding/yueli-official/task-artifacts/ytv/online-five-20261024/kazumi-akianime-4-opening.jpg)、[645秒](D:/coding/yueli-official/task-artifacts/ytv/online-five-20261024/kazumi-akianime-4-middle.jpg)、[1219秒](D:/coding/yueli-official/task-artifacts/ytv/online-five-20261024/kazumi-akianime-4-late.jpg) |

### moonci / kazumi-moonci

| 样本 | 实际片目 / 集数 / 线路 | 尺寸 / 时长 | 实播结果 | 画质观察 | 广告观察 | 证据 |
|---|---|---|---|---|---|---|
| 1 | [葬送的芙莉莲](https://tv.yuelili.com/watch/kazumi-moonci/aHR0cHM6Ly93d3cubW9vbmNpLmNvbS9hbmltZS8yMjguaHRtbA/aHR0cHM6Ly93d3cubW9vbmNpLmNvbS9hbmltZS8yMjgvcGxheS8yLTEuaHRtbA) / 第01话 / 线路 1 | 1920×1080 / 26.0分 | 三点出画 | 画面清晰，满足本轮目标 | 采样点未见明显广告 | [35秒](D:/coding/yueli-official/task-artifacts/ytv/online-five-20261024/kazumi-moonci-0-opening.jpg)、[701秒](D:/coding/yueli-official/task-artifacts/ytv/online-five-20261024/kazumi-moonci-0-middle.jpg)、[1325秒](D:/coding/yueli-official/task-artifacts/ytv/online-five-20261024/kazumi-moonci-0-late.jpg) |
| 2 | [鬼灭之刃](https://tv.yuelili.com/watch/kazumi-moonci/aHR0cHM6Ly93d3cubW9vbmNpLmNvbS9hbmltZS8yMjc0Lmh0bWw/aHR0cHM6Ly93d3cubW9vbmNpLmNvbS9hbmltZS8yMjc0L3BsYXkvMS0xLmh0bWw) / 第01话 / 线路 1 | 1920×1080 / 23.7分 | 三点出画 | 画面清晰，满足本轮目标 | 采样点未见明显广告 | [35秒](D:/coding/yueli-official/task-artifacts/ytv/online-five-20261024/kazumi-moonci-1-opening.jpg)、[639秒](D:/coding/yueli-official/task-artifacts/ytv/online-five-20261024/kazumi-moonci-1-middle.jpg)、[1207秒](D:/coding/yueli-official/task-artifacts/ytv/online-five-20261024/kazumi-moonci-1-late.jpg) |
| 3 | [咒术回战](https://tv.yuelili.com/watch/kazumi-moonci/aHR0cHM6Ly93d3cubW9vbmNpLmNvbS9hbmltZS8xMzc2Lmh0bWw/aHR0cHM6Ly93d3cubW9vbmNpLmNvbS9hbmltZS8xMzc2L3BsYXkvMS0xLmh0bWw) / 第01话 / 线路 1 | 1920×1080 / 23.9分 | 三点出画 | 画面清晰，满足本轮目标 | 采样点未见明显广告 | [35秒](D:/coding/yueli-official/task-artifacts/ytv/online-five-20261024/kazumi-moonci-2-opening.jpg)、[645秒](D:/coding/yueli-official/task-artifacts/ytv/online-five-20261024/kazumi-moonci-2-middle.jpg)、[1219秒](D:/coding/yueli-official/task-artifacts/ytv/online-five-20261024/kazumi-moonci-2-late.jpg) |
| 4 | [间谍过家家](https://tv.yuelili.com/watch/kazumi-moonci/aHR0cHM6Ly93d3cubW9vbmNpLmNvbS9hbmltZS8xMzY2Lmh0bWw/aHR0cHM6Ly93d3cubW9vbmNpLmNvbS9hbmltZS8xMzY2L3BsYXkvMS0xLmh0bWw) / 第01话 / 线路 1 | 1920×1080 / 24.2分 | 三点出画 | 画面清晰，满足本轮目标 | 采样点未见明显广告 | [35秒](D:/coding/yueli-official/task-artifacts/ytv/online-five-20261024/kazumi-moonci-3-opening.jpg)、[653秒](D:/coding/yueli-official/task-artifacts/ytv/online-five-20261024/kazumi-moonci-3-middle.jpg)、[1234秒](D:/coding/yueli-official/task-artifacts/ytv/online-five-20261024/kazumi-moonci-3-late.jpg) |
| 5 | [链锯人](https://tv.yuelili.com/watch/kazumi-moonci/aHR0cHM6Ly93d3cubW9vbmNpLmNvbS9hbmltZS8xNDAzLmh0bWw/aHR0cHM6Ly93d3cubW9vbmNpLmNvbS9hbmltZS8xNDAzL3BsYXkvMS0xLmh0bWw) / 第01话 / 线路 1 | 1920×1080 / 25.4分 | 三点出画 | 画面清晰，满足本轮目标 | 采样点未见明显广告 | [35秒](D:/coding/yueli-official/task-artifacts/ytv/online-five-20261024/kazumi-moonci-4-opening.jpg)、[686秒](D:/coding/yueli-official/task-artifacts/ytv/online-five-20261024/kazumi-moonci-4-middle.jpg)、[1296秒](D:/coding/yueli-official/task-artifacts/ytv/online-five-20261024/kazumi-moonci-4-late.jpg) |

### MXdm / kazumi-mxdm

| 样本 | 实际片目 / 集数 / 线路 | 尺寸 / 时长 | 实播结果 | 画质观察 | 广告观察 | 证据 |
|---|---|---|---|---|---|---|
| 1 | [葬送的芙莉莲](https://tv.yuelili.com/watch/kazumi-mxdm/aHR0cHM6Ly93d3cuZGNjMy5jb20vZGV0YWlsLzI3NDIv/aHR0cHM6Ly93d3cuZGNjMy5jb20vcGxheS8yNzQyLTEtMS8) / 第01集 / 线路 1 | 1280×720 / 26.6分 | 三点出画 | 画面低于本轮1080档目标 | 采样点未见明显广告 | [35秒](D:/coding/yueli-official/task-artifacts/ytv/online-five-20261024/kazumi-mxdm-0-opening.jpg)、[717秒](D:/coding/yueli-official/task-artifacts/ytv/online-five-20261024/kazumi-mxdm-0-middle.jpg)、[1354秒](D:/coding/yueli-official/task-artifacts/ytv/online-five-20261024/kazumi-mxdm-0-late.jpg) |
| 2 | [鬼灭之刃](https://tv.yuelili.com/watch/kazumi-mxdm/aHR0cHM6Ly93d3cuZGNjMy5jb20vZGV0YWlsLzg2MTAv/aHR0cHM6Ly93d3cuZGNjMy5jb20vcGxheS84NjEwLTQtMS8) / 第01集 / 线路 1 | 1080×608 / 24.6分 | 三点出画 | 画面低于本轮1080档目标 | 后段顶部博彩横幅 | [35秒](D:/coding/yueli-official/task-artifacts/ytv/online-five-20261024/kazumi-mxdm-1-opening.jpg)、[663秒](D:/coding/yueli-official/task-artifacts/ytv/online-five-20261024/kazumi-mxdm-1-middle.jpg)、[1252秒](D:/coding/yueli-official/task-artifacts/ytv/online-five-20261024/kazumi-mxdm-1-late.jpg) |
| 3 | [咒术回战](https://tv.yuelili.com/watch/kazumi-mxdm/aHR0cHM6Ly93d3cuZGNjMy5jb20vZGV0YWlsLzEzODQv/aHR0cHM6Ly93d3cuZGNjMy5jb20vcGxheS8xMzg0LTEtMS8) / 第01集 / 线路 1 | 1080×608 / 24.5分 | 三点出画 | 画面低于本轮1080档目标 | 后段顶部博彩横幅 | [35秒](D:/coding/yueli-official/task-artifacts/ytv/online-five-20261024/kazumi-mxdm-2-opening.jpg)、[661秒](D:/coding/yueli-official/task-artifacts/ytv/online-five-20261024/kazumi-mxdm-2-middle.jpg)、[1248秒](D:/coding/yueli-official/task-artifacts/ytv/online-five-20261024/kazumi-mxdm-2-late.jpg) |
| 4 | [间谍过家家](https://tv.yuelili.com/watch/kazumi-mxdm/aHR0cHM6Ly93d3cuZGNjMy5jb20vZGV0YWlsLzEyMDUv/aHR0cHM6Ly93d3cuZGNjMy5jb20vcGxheS8xMjA1LTEtMS8) / 第01集 / 线路 1 | 1280×720 / 24.7分 | 三点出画 | 画面低于本轮1080档目标 | 后段顶部博彩横幅 | [35秒](D:/coding/yueli-official/task-artifacts/ytv/online-five-20261024/kazumi-mxdm-3-opening.jpg)、[667秒](D:/coding/yueli-official/task-artifacts/ytv/online-five-20261024/kazumi-mxdm-3-middle.jpg)、[1261秒](D:/coding/yueli-official/task-artifacts/ytv/online-five-20261024/kazumi-mxdm-3-late.jpg) |
| 5 | [电锯人](https://tv.yuelili.com/watch/kazumi-mxdm/aHR0cHM6Ly93d3cuZGNjMy5jb20vZGV0YWlsLzE2MDIv/aHR0cHM6Ly93d3cuZGNjMy5jb20vcGxheS8xNjAyLTEtMS8) / 第01集 / 线路 1 | 1080×608 / 25.9分 | 三点出画 | 画面低于本轮1080档目标 | 采样点未见明显广告 | [35秒](D:/coding/yueli-official/task-artifacts/ytv/online-five-20261024/kazumi-mxdm-4-opening.jpg)、[699秒](D:/coding/yueli-official/task-artifacts/ytv/online-five-20261024/kazumi-mxdm-4-middle.jpg)、[1321秒](D:/coding/yueli-official/task-artifacts/ytv/online-five-20261024/kazumi-mxdm-4-late.jpg) |

### xfdmneo / kazumi-xfdmneo

| 样本 | 实际片目 / 集数 / 线路 | 尺寸 / 时长 | 实播结果 | 画质观察 | 广告观察 | 证据 |
|---|---|---|---|---|---|---|
| 1 | 葬送的芙莉莲 | — | 搜索：503 ytv.source_unavailable | 未出画，画质未审 | 未审 | 无画面证据 |
| 2 | 鬼灭之刃 | — | 搜索：503 ytv.source_unavailable | 未出画，画质未审 | 未审 | 无画面证据 |
| 3 | 咒术回战 | — | 搜索：503 ytv.source_unavailable | 未出画，画质未审 | 未审 | 无画面证据 |
| 4 | 间谍过家家 | — | 搜索：503 ytv.source_unavailable | 未出画，画质未审 | 未审 | 无画面证据 |
| 5 | 链锯人 | — | 搜索：503 ytv.source_unavailable | 未出画，画质未审 | 未审 | 无画面证据 |

### giriGiriLove / kazumi-girigirilove

| 样本 | 实际片目 / 集数 / 线路 | 尺寸 / 时长 | 实播结果 | 画质观察 | 广告观察 | 证据 |
|---|---|---|---|---|---|---|
| 1 | [葬送的芙莉莲](https://tv.yuelili.com/watch/kazumi-girigirilove/MjU1MjQ/aHR0cHM6Ly9hbmEuZ2lyaWdpcmlsb3ZlLmNvbS96aWppYW4vYW5pbWUvMjAyNC8wOC8wODAyL1NvdXNvdW5vRnJpZXJlbkNIVC8wMS5tcDQ) / 01 / cht | 1920×1080 / 26.0分 | 三点出画 | 画面清晰，满足本轮目标 | 采样点未见明显广告 | [35秒](D:/coding/yueli-official/task-artifacts/ytv/online-five-20261024/kazumi-girigirilove-0-opening.jpg)、[702秒](D:/coding/yueli-official/task-artifacts/ytv/online-five-20261024/kazumi-girigirilove-0-middle.jpg)、[1326秒](D:/coding/yueli-official/task-artifacts/ytv/online-five-20261024/kazumi-girigirilove-0-late.jpg) |
| 2 | [鬼灭之刃](https://tv.yuelili.com/watch/kazumi-girigirilove/MTk5/aHR0cHM6Ly9hbmEuZ2lyaWdpcmlsb3ZlLmNvbS96aWppYW4vYW5pbWUvMjAyNC8wMjA2L0tpbWV0c3Vub1lhaWJhLzAxLm1wNA) / 01 / chs | 1920×1080 / 23.7分 | 三点出画 | 画面清晰，满足本轮目标 | 采样点未见明显广告 | [35秒](D:/coding/yueli-official/task-artifacts/ytv/online-five-20261024/kazumi-girigirilove-1-opening.jpg)、[639秒](D:/coding/yueli-official/task-artifacts/ytv/online-five-20261024/kazumi-girigirilove-1-middle.jpg)、[1207秒](D:/coding/yueli-official/task-artifacts/ytv/online-five-20261024/kazumi-girigirilove-1-late.jpg) |
| 3 | [咒术回战](https://tv.yuelili.com/watch/kazumi-girigirilove/Nzc2/aHR0cHM6Ly9hbmEuZ2lyaWdpcmlsb3ZlLmNvbS96aWppYW4vYW5pbWUvMjAyNC8wNy8wNzA4L0p1anV0c3VLYWlzZW4vMDEubXA0) / 01 / chs | 1920×1080 / 23.9分 | 三点出画 | 画面清晰，满足本轮目标 | 采样点未见明显广告 | [35秒](D:/coding/yueli-official/task-artifacts/ytv/online-five-20261024/kazumi-girigirilove-2-opening.jpg)、[646秒](D:/coding/yueli-official/task-artifacts/ytv/online-five-20261024/kazumi-girigirilove-2-middle.jpg)、[1220秒](D:/coding/yueli-official/task-artifacts/ytv/online-five-20261024/kazumi-girigirilove-2-late.jpg) |
| 4 | [间谍过家家](https://tv.yuelili.com/watch/kazumi-girigirilove/MTA5/aHR0cHM6Ly9hbmEuZ2lyaWdpcmlsb3ZlLmNvbS96aWppYW4vYW5pbWUvMjAyNC8xMi8xMjI0L1NQWXhGQU1JTFkvMDEubXA0) / 01 / chs | 1920×1080 / 24.2分 | 三点出画 | 画面清晰，满足本轮目标 | 采样点未见明显广告 | [35秒](D:/coding/yueli-official/task-artifacts/ytv/online-five-20261024/kazumi-girigirilove-3-opening.jpg)、[653秒](D:/coding/yueli-official/task-artifacts/ytv/online-five-20261024/kazumi-girigirilove-3-middle.jpg)、[1234秒](D:/coding/yueli-official/task-artifacts/ytv/online-five-20261024/kazumi-girigirilove-3-late.jpg) |
| 5 | [电锯人](https://tv.yuelili.com/watch/kazumi-girigirilove/NzY4/aHR0cHM6Ly9hbmEuZ2lyaWdpcmlsb3ZlLmNvbS96aWppYW4vb2xkYW5pbWUvMjAyMi8wMS9DaGFpbnNhd01hbi8wMS5tcDQ) / 01 / chs | 1920×1080 / 25.5分 | 三点出画 | 画面清晰，满足本轮目标 | 采样点未见明显广告 | [35秒](D:/coding/yueli-official/task-artifacts/ytv/online-five-20261024/kazumi-girigirilove-4-opening.jpg)、[687秒](D:/coding/yueli-official/task-artifacts/ytv/online-five-20261024/kazumi-girigirilove-4-middle.jpg)、[1297秒](D:/coding/yueli-official/task-artifacts/ytv/online-five-20261024/kazumi-girigirilove-4-late.jpg) |

### mutefun / kazumi-mutefun

| 样本 | 实际片目 / 集数 / 线路 | 尺寸 / 时长 | 实播结果 | 画质观察 | 广告观察 | 证据 |
|---|---|---|---|---|---|---|
| 1 | [葬送的芙莉莲](https://tv.yuelili.com/watch/kazumi-mutefun/aHR0cHM6Ly93d3cuMmtkbS5jb20vdm9kZGV0YWlsLzIzODMuaHRtbA/aHR0cHM6Ly93d3cuMmtkbS5jb20vdm9kcGxheS8yMzgzLTItMS5odG1s) / 第01集 / 线路 1 | — | 解析：504 ytv.gateway_timeout | 未出画，画质未审 | 未审 | 无画面证据 |
| 2 | [鬼灭之刃](https://tv.yuelili.com/watch/kazumi-mutefun/aHR0cHM6Ly93d3cuMmtkbS5jb20vdm9kZGV0YWlsLzE4NzEuaHRtbA/aHR0cHM6Ly93d3cuMmtkbS5jb20vdm9kcGxheS8xODcxLTMtMS5odG1s) / 第01集 / 线路 1 | — | 解析：504 ytv.gateway_timeout | 未出画，画质未审 | 未审 | 无画面证据 |
| 3 | [咒术回战](https://tv.yuelili.com/watch/kazumi-mutefun/aHR0cHM6Ly93d3cuMmtkbS5jb20vdm9kZGV0YWlsLzEyMjYuaHRtbA/aHR0cHM6Ly93d3cuMmtkbS5jb20vdm9kcGxheS8xMjI2LTMtMS5odG1s) / 第01集 / 线路 1 | — | 解析：504 ytv.gateway_timeout | 未出画，画质未审 | 未审 | 无画面证据 |
| 4 | [间谍过家家](https://tv.yuelili.com/watch/kazumi-mutefun/aHR0cHM6Ly93d3cuMmtkbS5jb20vdm9kZGV0YWlsLzMyNC5odG1s/aHR0cHM6Ly93d3cuMmtkbS5jb20vdm9kcGxheS8zMjQtMy0xLmh0bWw) / 第01集 / 线路 1 | — | 解析：504 ytv.gateway_timeout | 未出画，画质未审 | 未审 | 无画面证据 |
| 5 | [电锯人](https://tv.yuelili.com/watch/kazumi-mutefun/aHR0cHM6Ly93d3cuMmtkbS5jb20vdm9kZGV0YWlsLzE3MjkuaHRtbA/aHR0cHM6Ly93d3cuMmtkbS5jb20vdm9kcGxheS8xNzI5LTMtMS5odG1s) / 第01集 / 线路 1 | — | 解析：504 ytv.gateway_timeout | 未出画，画质未审 | 未审 | 无画面证据 |

### OmoFun / kazumi-omofun

| 样本 | 实际片目 / 集数 / 线路 | 尺寸 / 时长 | 实播结果 | 画质观察 | 广告观察 | 证据 |
|---|---|---|---|---|---|---|
| 1 | [葬送的芙莉莲](https://tv.yuelili.com/watch/kazumi-omofun/aHR0cHM6Ly9vbW9mdW4uaW4vdm9kL2RldGFpbC8yMDIzMTAzMDkzLmh0bWw/aHR0cHM6Ly9vbW9mdW4uaW4vdm9kL3BsYXkvMjAyMzEwMzA5My9lcDEuaHRtbA) / 第01集 / 线路 1 | — | 解析：504 ytv.gateway_timeout | 未出画，画质未审 | 未审 | 无画面证据 |
| 2 | [鬼灭之刃](https://tv.yuelili.com/watch/kazumi-omofun/aHR0cHM6Ly9vbW9mdW4uaW4vdm9kL2RldGFpbC8yMDE5MjE5OTYzLmh0bWw/aHR0cHM6Ly9vbW9mdW4uaW4vdm9kL3BsYXkvMjAxOTIxOTk2My9lcDEuaHRtbA) / 第01集 / 线路 1 | 1920×1080 / 23.7分 | 三点出画 | 画面清晰，满足本轮目标 | 中段顶部博彩横幅 | [35秒](D:/coding/yueli-official/task-artifacts/ytv/online-five-20261024/kazumi-omofun-1-opening.jpg)、[641秒](D:/coding/yueli-official/task-artifacts/ytv/online-five-20261024/kazumi-omofun-1-middle.jpg)、[1211秒](D:/coding/yueli-official/task-artifacts/ytv/online-five-20261024/kazumi-omofun-1-late.jpg) |
| 3 | [咒术回战 0](https://tv.yuelili.com/watch/kazumi-omofun/aHR0cHM6Ly9vbW9mdW4uaW4vdm9kL2RldGFpbC8yMDIxNDI0MzY1Lmh0bWw/aHR0cHM6Ly9vbW9mdW4uaW4vdm9kL3BsYXkvMjAyMTQyNDM2NS9tb3ZpZS5odG1sP3ZvZD16aGVuZ19waWFu) / 正片 / 线路 1 | 1280×720 / 104.9分 | 三点出画 | 画面低于本轮1080档目标 | 中段右下博彩叠字 | [35秒](D:/coding/yueli-official/task-artifacts/ytv/online-five-20261024/kazumi-omofun-2-opening.jpg)、[2832秒](D:/coding/yueli-official/task-artifacts/ytv/online-five-20261024/kazumi-omofun-2-middle.jpg)、[5351秒](D:/coding/yueli-official/task-artifacts/ytv/online-five-20261024/kazumi-omofun-2-late.jpg) |
| 4 | [间谍过家家 第二季](https://tv.yuelili.com/watch/kazumi-omofun/aHR0cHM6Ly9vbW9mdW4uaW4vdm9kL2RldGFpbC8yMDIzMjAwNjMwLmh0bWw/aHR0cHM6Ly9vbW9mdW4uaW4vdm9kL3BsYXkvMjAyMzIwMDYzMC9lcDEuaHRtbA) / 第01集 / 线路 1 | 1920×1080 / 23.7分 | 三点出画 | 画面清晰，满足本轮目标 | 中段顶部博彩横幅 | [35秒](D:/coding/yueli-official/task-artifacts/ytv/online-five-20261024/kazumi-omofun-3-opening.jpg)、[639秒](D:/coding/yueli-official/task-artifacts/ytv/online-five-20261024/kazumi-omofun-3-middle.jpg)、[1207秒](D:/coding/yueli-official/task-artifacts/ytv/online-five-20261024/kazumi-omofun-3-late.jpg) |
| 5 | 链锯人：刺客篇 | — | 详情：503 ytv.source_unavailable | 未出画，画质未审 | 未审 | 无画面证据 |

### 番茄动漫 / kazumi-fqdm

| 样本 | 实际片目 / 集数 / 线路 | 尺寸 / 时长 | 实播结果 | 画质观察 | 广告观察 | 证据 |
|---|---|---|---|---|---|---|
| 1 | [葬送的芙莉莲](https://tv.yuelili.com/watch/kazumi-fqdm/aHR0cHM6Ly93d3cuZnFkbS5jYy9pbmRleC5waHAvdm9kL2RldGFpbC9pZC81NDguaHRtbA/aHR0cHM6Ly93d3cuZnFkbS5jYy9pbmRleC5waHAvdm9kL3BsYXkvaWQvNTQ4L3NpZC8yL25pZC8xLmh0bWw) / 第01集 / 线路 1 | — | 媒体加载：23秒未开始出画 | 未出画，画质未审 | 未审 | 无画面证据 |
| 2 | [鬼灭之刃](https://tv.yuelili.com/watch/kazumi-fqdm/aHR0cHM6Ly93d3cuZnFkbS5jYy9pbmRleC5waHAvdm9kL2RldGFpbC9pZC8zMDE5Lmh0bWw/aHR0cHM6Ly93d3cuZnFkbS5jYy9pbmRleC5waHAvdm9kL3BsYXkvaWQvMzAxOS9zaWQvMi9uaWQvMS5odG1s) / 第01集 / 线路 1 | — | 媒体加载：23秒未开始出画 | 未出画，画质未审 | 未审 | 无画面证据 |
| 3 | [咒术回战](https://tv.yuelili.com/watch/kazumi-fqdm/aHR0cHM6Ly93d3cuZnFkbS5jYy9pbmRleC5waHAvdm9kL2RldGFpbC9pZC8xNzA2Lmh0bWw/aHR0cHM6Ly93d3cuZnFkbS5jYy9pbmRleC5waHAvdm9kL3BsYXkvaWQvMTcwNi9zaWQvMi9uaWQvMS5odG1s) / 第01集 / 线路 1 | — | 媒体加载：23秒未开始出画 | 未出画，画质未审 | 未审 | 无画面证据 |
| 4 | [间谍过家家](https://tv.yuelili.com/watch/kazumi-fqdm/aHR0cHM6Ly93d3cuZnFkbS5jYy9pbmRleC5waHAvdm9kL2RldGFpbC9pZC8yNjY1Lmh0bWw/aHR0cHM6Ly93d3cuZnFkbS5jYy9pbmRleC5waHAvdm9kL3BsYXkvaWQvMjY2NS9zaWQvMS9uaWQvMS5odG1s) / 第1集 / 线路 1 | 1920×1080 / 24.1分 | 三点出画 | 画面清晰，满足本轮目标 | 中段顶部博彩横幅 | [35秒](D:/coding/yueli-official/task-artifacts/ytv/online-five-20261024/kazumi-fqdm-3-opening.jpg)、[650秒](D:/coding/yueli-official/task-artifacts/ytv/online-five-20261024/kazumi-fqdm-3-middle.jpg)、[1228秒](D:/coding/yueli-official/task-artifacts/ytv/online-five-20261024/kazumi-fqdm-3-late.jpg) |
| 5 | [电锯人](https://tv.yuelili.com/watch/kazumi-fqdm/aHR0cHM6Ly93d3cuZnFkbS5jYy9pbmRleC5waHAvdm9kL2RldGFpbC9pZC8yNTU2Lmh0bWw/aHR0cHM6Ly93d3cuZnFkbS5jYy9pbmRleC5waHAvdm9kL3BsYXkvaWQvMjU1Ni9zaWQvNC9uaWQvMS5odG1s) / 第01集 / 线路 1 | 1920×1080 / 25.4分 | 三点出画 | 画面清晰，满足本轮目标 | 中段顶部博彩横幅 | [35秒](D:/coding/yueli-official/task-artifacts/ytv/online-five-20261024/kazumi-fqdm-4-opening.jpg)、[686秒](D:/coding/yueli-official/task-artifacts/ytv/online-five-20261024/kazumi-fqdm-4-middle.jpg)、[1296秒](D:/coding/yueli-official/task-artifacts/ytv/online-five-20261024/kazumi-fqdm-4-late.jpg) |

### 动漫共和国 / kazumi-dmghg

| 样本 | 实际片目 / 集数 / 线路 | 尺寸 / 时长 | 实播结果 | 画质观察 | 广告观察 | 证据 |
|---|---|---|---|---|---|---|
| 1 | 葬送的芙莉莲 | — | 搜索：503 ytv.source_unavailable | 未出画，画质未审 | 未审 | 无画面证据 |
| 2 | 鬼灭之刃 | — | 搜索：503 ytv.source_unavailable | 未出画，画质未审 | 未审 | 无画面证据 |
| 3 | [咒术回战](https://tv.yuelili.com/watch/kazumi-dmghg/aHR0cHM6Ly93d3cuZG1naGcxLmNvbS9pbmRleC5waHAvdm9kL2RldGFpbC9pZC83NzE1Ni5odG1s/aHR0cHM6Ly93d3cuZG1naGcxLmNvbS9pbmRleC5waHAvdm9kL3BsYXkvaWQvNzcxNTYvc2lkLzQvbmlkLzEuaHRtbA) / 第01集 / 线路 1 | 1920×1080 / 23.9分 | 三点出画 | 画面清晰，满足本轮目标 | 采样点未见明显广告 | [35秒](D:/coding/yueli-official/task-artifacts/ytv/online-five-20261024/kazumi-dmghg-2-opening.jpg)、[645秒](D:/coding/yueli-official/task-artifacts/ytv/online-five-20261024/kazumi-dmghg-2-middle.jpg)、[1219秒](D:/coding/yueli-official/task-artifacts/ytv/online-five-20261024/kazumi-dmghg-2-late.jpg) |
| 4 | [间谍过家家](https://tv.yuelili.com/watch/kazumi-dmghg/aHR0cHM6Ly93d3cuZG1naGcxLmNvbS9pbmRleC5waHAvdm9kL2RldGFpbC9pZC85Mzc2Mi5odG1s/aHR0cHM6Ly93d3cuZG1naGcxLmNvbS9pbmRleC5waHAvdm9kL3BsYXkvaWQvOTM3NjIvc2lkLzEvbmlkLzEuaHRtbA) / 第01集 / 线路 1 | — | 解析：504 ytv.gateway_timeout | 未出画，画质未审 | 未审 | 无画面证据 |
| 5 | [电锯人](https://tv.yuelili.com/watch/kazumi-dmghg/aHR0cHM6Ly93d3cuZG1naGcxLmNvbS9pbmRleC5waHAvdm9kL2RldGFpbC9pZC81Njg2OS5odG1s/aHR0cHM6Ly93d3cuZG1naGcxLmNvbS9pbmRleC5waHAvdm9kL3BsYXkvaWQvNTY4Njkvc2lkLzIvbmlkLzEuaHRtbA) / 第01集 / 线路 1 | — | 解析：504 ytv.gateway_timeout | 未出画，画质未审 | 未审 | 无画面证据 |

## 证据与复核

- 来源快照时间（UTC）：2026-10-24T04:22:44.436Z；[线上来源接口](https://tv.yuelili.com/api/v1/sources)。
- 本机证据目录：`D:/coding/yueli-official/task-artifacts/ytv/online-five-20261024`。`audit.json`保存首次105条；`supplement.json`和`extra.json`保存补查；`merged.json`保存最终105槽；`reviewed.json`附人工画质/广告结论。初次失败没有删除。
- 每个来源的 `*-sheet.jpg` 是人工复看的画面拼图，单帧 JPG 保留。没有保存第三方完整视频，也没有保存带签名的媒体URL。
- 以上仅为抽样审核；未逐秒看片、未检查所有线路，未对音轨/字幕翻译、全片完整性或真实原生分辨率做鉴定。
- 与此前 [候选来源五片直测](source-five-sample-20261024.md) 的对象和路径不同：本表以当前线上公开清单与线上解析结果为准，不把未上线候选混入。

补测原始记录与24张画面：`D:/coding/yueli-official/task-artifacts/ytv/deploy-playback-policy-20261024-06/production-media.json`；同目录 production-release.json、production-checks.json 保存发布和正式站双端播放验证。
