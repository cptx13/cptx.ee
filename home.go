// Home page as an index (D-03): latest posts and the sections; the full list is /all/.

package main

// homeData returns the home page's template data (without "on this day").
func homeData(all []*Post, toPostList func([]*Post) []any, bySection map[string][]*Post, photoPlaces int) map[string]any {
	latest := all
	if len(latest) > 12 {
		latest = latest[:12]
	}
	return map[string]any{
		"latest":         toPostList(latest),
		"postCount":      len(all),
		"croquisCount":   len(bySection["croquis"]),
		"recuerdosCount": len(bySection["recuerdos"]),
		"photoPlaces":    photoPlaces,
	}
}
