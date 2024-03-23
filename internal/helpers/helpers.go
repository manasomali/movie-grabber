package helpers

func FilterMovieLinks(links []string) []string {
	filteredLinks := []string{}
	for _, link := range links {
		if len(link) >= 4 && link[len(link)-4:] == ".zip" {
			filteredLinks = append(filteredLinks, link)
		}
	}
	return filteredLinks
}
