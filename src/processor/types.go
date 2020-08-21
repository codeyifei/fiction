package processor

import "github.com/PuerkitoBio/goquery"

type Scheme string

const (
	HttpsScheme Scheme = "https"
	HttpScheme         = "http"
)

type Config struct {
	Host   string
	Seed   string
	Scheme Scheme
}

type ReptileDriver interface {
	LoadMenuPage() (*goquery.Document, error)
	GetFictionTitle(*goquery.Document) string
	GetMenu(*goquery.Document) ([]string, error)
}

type MenuMeta struct {
	Title string
	Path  string
}
