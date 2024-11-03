package models

// RepositorySearchResponse is the response from the GitHub API for the search repositories endpoint
// We do not use all fields from the response, only few ones, but adding them would be straightforward
type RepositorySearchResponse struct {
	TotalCount        int          `json:"total_count"`
	Count             int          `json:"count"`
	PerPage           string       `json:"per_page"`
	Page              string       `json:"page"`
	IncompleteResults bool         `json:"incomplete_results"`
	Items             []Repository `json:"items"`
}

// Repository is a single repository from the GitHub API response
type Repository struct {
	FullName    string    `json:"full_name"`
	Name        string    `json:"name"`
	Description string    `json:"description"`
	Languages   Languages `json:"languages"`
	Owner       Owner     `json:"owner"`
}

// Owner is the owner of a repository
type Owner struct {
	Login     string `json:"login"`
	ID        int    `json:"id"`
	NodeID    string `json:"node_id"`
	AvatarURL string `json:"avatar_url"`
}

// Languages is a map of languages to their usage in a repository
type Languages map[string]int

// RepositorySearchParams are the parameters for functions used to search repositories
type RepositorySearchParams struct {
	Query    string
	PerPage  string
	Page     string
	Header   string
	Language string
}
