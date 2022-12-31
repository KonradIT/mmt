package gopro

import "time"

func getImportanceName(tags []int, videoDuration int, names []string) string {
	if videoDuration < 20 || len(names) < 3 {
		return ""
	}
	var lastSeconds = 10 // Last 10 seconds
	rangeToFind := (videoDuration * int(time.Microsecond)) - lastSeconds*int(time.Microsecond)

	howMany := 0
	for _, tag := range tags {
		if tag > rangeToFind {
			howMany++
		}
	}

	if howMany == 0 {
		return ""
	}

	if howMany > len(names) {
		howMany = len(names)
	}
	return names[howMany-1]
}
