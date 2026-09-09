// TODO: может вынести в домен?
package api

type ShortenRequest struct {
	URL string `json:"url"`
}

type ShortenResponse struct {
	Code string `json:"code"`
}

type ResolveResponse struct {
	OriginalURL string `json:"original_url"`
}

type ErrorResponse struct {
	Error string `json:"error"`
}
