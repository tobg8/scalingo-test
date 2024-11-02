package models

type RepositorySearchResponse struct {
	TotalCount        int          `json:"total_count"`
	Count             int          `json:"count"`
	IncompleteResults bool         `json:"incomplete_results"`
	Items             []Repository `json:"items"`
}

type Repository struct {
	FullName    string    `json:"full_name"`
	Name        string    `json:"name"`
	Description string    `json:"description"`
	Languages   Languages `json:"languages"`
	Owner       Owner     `json:"owner"`
}

type Owner struct {
	Login     string `json:"login"`
	ID        int    `json:"id"`
	NodeID    string `json:"node_id"`
	AvatarURL string `json:"avatar_url"`
}

type Languages map[string]int
