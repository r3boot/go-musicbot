package albumart

const (
	SEARCH_URL   = "https://api.discogs.com/database/search?q="
	NOTFOUND_URI = "/img/notfound.png"
)

type dgPaginationUrls struct {
	Last string `json:"last"`
	Next string `json:"next"`
}

type dgPagination struct {
	PerPage int              `json:"per_page"`
	Pages   int              `json:"pages"`
	Page    int              `json:"page"`
	Urls    dgPaginationUrls `json:"urls"`
}

type dgResultCommunity struct {
	Want int `json:"want"`
	Have int `json:"have"`
}

type dgResult struct {
	Style       []string          `json:"style"`
	Thumb       string            `json:"thumb"`
	Format      []string          `json:"format"`
	Country     string            `json:"country"`
	Barcode     []string          `json:"barcode"`
	Uri         string            `json:"uri"`
	Community   dgResultCommunity `json:"community"`
	Label       []string          `json:"label"`
	CatNo       string            `json:"catno"`
	Year        string            `json:"year"`
	Genre       []string          `json:"genre"`
	Title       string            `json:"title"`
	ResourceUrl string            `json:"resource_url"`
	Type        string            `json:"type"`
	Id          int               `json:"id"`
}

type dgSearchResult struct {
	Pagination dgPagination `json:"pagination"`
	Items      int          `json:"items"`
	Results    []dgResult   `json:"results"`
}

type AlbumArt struct {
	cacheDir string
}
