package processor

import "github.com/PuerkitoBio/goquery"

type Scheme string

const (
	HttpsScheme Scheme = "https"
)

type Config struct {
	Host   string
	Seed   string
	Scheme Scheme
}

type ReptileDriver interface {
	// 获取最大goroutine数量
	GetMaxGoroutineQuantity() (quantity int)
	// 获取菜单页
	LoadMenuPage() (dom *goquery.Document, err error)
	// 获取正文页
	LoadContentPage(path string) (dom *goquery.Document, err error)
	// 获取小说标题
	GetFictionTitle(dom *goquery.Document) (title string)
	// 获取菜单
	GetMenu(dom *goquery.Document) (menus <-chan *MenuMeta)
	// 获取章节正文
	GetContent(dom *goquery.Document) (content []string, err error)
}

type MenuMeta struct {
	Index int
	Title string
	Path  string
}
