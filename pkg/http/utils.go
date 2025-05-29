package http

// normalizeURL объединяет baseURL и path, убирая лишние слеши
func normalizeURL(baseURL, path string) string {
	// Remove trailing slash from baseURL if exists
	if len(baseURL) > 0 && baseURL[len(baseURL)-1] == '/' {
		baseURL = baseURL[:len(baseURL)-1]
	}
	// Remove leading slash from path if exists
	if len(path) > 0 && path[0] == '/' {
		path = path[1:]
	}
	return baseURL + "/" + path
}