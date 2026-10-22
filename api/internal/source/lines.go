package source

// EpisodeTotal counts content within a line, not alternative delivery lines.
func EpisodeTotal(episodes []Episode) int {
	counts := make(map[string]int)
	maximum := 0
	for _, episode := range episodes {
		counts[episode.LineKey]++
		if counts[episode.LineKey] > maximum {
			maximum = counts[episode.LineKey]
		}
	}
	return maximum
}
