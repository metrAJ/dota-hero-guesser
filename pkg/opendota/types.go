package opendota

// A subset of fields that are required for the downstream logic.
type Hero struct {
	ID       uint   `json:"id"`
	Name     string `json:"localized_name"`
	ImageURL string `json:"img"`
	Type     string `json:"primary_attr"`
}
type Item struct {
	ID       uint   `json:"id"`
	Name     string `json:"dname"`
	ImageURL string `json:"img"`
}

type FetchHeroesResponse map[string]Hero // id -> hero]
type FetchItemsResponse map[string]Item  // id -> item]
