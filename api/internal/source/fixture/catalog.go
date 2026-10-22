package fixture

import (
	"fmt"

	"github.com/yueli-official/ytv/api/internal/source"
)

func Titles() []source.Title {
	return []source.Title{
		{
			Locator: source.TitleLocator{ExternalID: "ring-sea"}, Name: "环海之门", Year: "2026", Category: "科幻",
			Area: "合成地区", Language: "中文", Remarks: "演示内容",
			Summary:  "用于本地验收的合成条目。正式页面只展示当前获准来源返回的事实。",
			ImageURL: "/demo/hero-gate.png",
			Episodes: []source.Episode{
				{Key: "0", Title: "第 1 集", PlaybackURL: "/demo/hls/index.m3u8"},
				{Key: "1", Title: "第 2 集", PlaybackURL: "/demo/hls/index.m3u8"},
			},
			EpisodeCount: 2,
		},
		{
			Locator: source.TitleLocator{ExternalID: "frost-line"}, Name: "霜线", Year: "2025", Category: "悬疑",
			Area: "合成地区", Language: "中文", Remarks: "演示内容", ImageURL: "/demo/card-frost.png",
			Episodes: []source.Episode{
				{Key: "0", Title: "正片", PlaybackURL: "/demo/hls/index.m3u8", LineKey: "0", LineName: "主线路"},
				{Key: "line1-0", Title: "正片", PlaybackURL: "/demo/hls/index.m3u8", LineKey: "1", LineName: "备用线路"},
			},
			EpisodeCount: 1,
		},
		fixtureTitle("orbit-beyond", "轨道之外", "2026", "科幻", "/demo/card-orbit.png"),
		fixtureTitle("night-line", "夜行线", "2026", "剧情", "/demo/card-city.png"),
		fixtureTitle("desert-archive", "砂原档案", "2024", "纪录", "/demo/card-desert.png"),
		fixtureTitle("abyss-archive", "深水档案", "2025", "悬疑", "/demo/card-abyss.png"),
		longVoyageTitle(),
	}
}

func longVoyageTitle() source.Title {
	episodes := make([]source.Episode, 0, 125)
	for index := 0; index < 125; index++ {
		episodes = append(episodes, source.Episode{
			Key: indexTitle(index), Title: episodeTitle(index), PlaybackURL: "/demo/hls/index.m3u8",
		})
	}
	return source.Title{
		Locator: source.TitleLocator{ExternalID: "long-voyage"}, Name: "长航季", Year: "2026", Category: "动画",
		Area: "合成地区", Language: "中文", Remarks: "全125集", Summary: "用于长剧集播放器范围与导航验收的合成条目。",
		ImageURL: "/demo/card-orbit.png", Episodes: episodes, EpisodeCount: len(episodes),
	}
}

func indexTitle(index int) string {
	return fmt.Sprintf("%d", index)
}

func episodeTitle(index int) string {
	return fmt.Sprintf("第%d集 航线记录", index+1)
}

func fixtureTitle(id, name, year, category, imageURL string) source.Title {
	return source.Title{
		Locator: source.TitleLocator{ExternalID: id}, Name: name, Year: year, Category: category,
		Area: "合成地区", Language: "中文", Remarks: "演示内容", ImageURL: imageURL,
		Episodes:     []source.Episode{{Key: "0", Title: "第 1 集", PlaybackURL: "/demo/hls/index.m3u8"}},
		EpisodeCount: 1,
	}
}
