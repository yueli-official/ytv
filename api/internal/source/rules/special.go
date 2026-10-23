package rules

// ApplySpecialPlayback attaches deployment-reviewed playback behavior that is
// intentionally outside portable Kazumi XPath documents.
func ApplySpecialPlayback(plan *Plan) {
	if plan == nil {
		return
	}
	switch plan.Name {
	case "mutefun":
		// Observed 2026-10-23: HTTP 200 MacCMS search verification page.
		plan.VerificationXPath = []string{
			"//title[normalize-space(.)='系统安全验证']",
			"//input[@name='verify' and contains(concat(' ', normalize-space(@class), ' '), ' mac_verify ')]",
			"//input[contains(concat(' ', normalize-space(@class), ' '), ' verify_submit ')]",
		}
	case "7sefun":
		plan.Playback.Extractor = "maccms_ecplayer"
		plan.Playback.ParserURL = "https://dp.no3acg.com/player/ec.php?code=qw&if=1"
		plan.Playback.PlaybackHosts = []string{"*.svp.tencent-cloud.com", "v1.adkwai.com"}
		plan.Title.NameXPath = "//div[@class='video-p-name']/text()[1]"
	case "aafun":
		plan.Title = TitlePlan{
			RoadsXPath:    "//div[contains(@class, 'hl-play-source')]//ul[contains(@class, 'hl-plays-list')]",
			EpisodesXPath: ".//a[contains(@href, '/play/')]",
			MultipleRoads: true,
		}
		plan.Playback.Extractor = "maccms_direct"
		plan.Playback.PlaybackHosts = []string{"dl.playxf.top"}
	case "moonci":
		plan.Playback.Extractor = "maccms_direct"
		plan.Playback.PlaybackHosts = []string{"dl.playxf.top"}
	case "akianime":
		plan.Playback.Extractor = "aniplayer"
		plan.Playback.ParserURL = "https://aniplayer.xn--gmqr9gevarqk8t.cn/"
		plan.Playback.PlaybackHosts = []string{"*.cmecloud.cn"}
	case "xfdmneo":
		plan.Playback.Extractor = "maccms_direct"
		plan.Playback.PlaybackHosts = []string{"play.xfvod.pro"}
	case "MXdm":
		plan.Playback.Extractor = "html_media"
		plan.Playback.PlaybackHosts = []string{"*.play-cdn19.com"}
	}
}
