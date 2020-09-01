package biquge

import (
	"errors"
	"fmt"
	"net/http"
	"regexp"
	"strings"

	"github.com/PuerkitoBio/goquery"
	"github.com/codeyifei/fiction/src/processor"
)

const (
	fictionTitleElement string = ".box_con #maininfo #info h1"
	menuElement                = "#list dl dd a"
	contentElement             = "#content"
)

type drive struct {
	config               *processor.Config
	maxGoroutineQuantity uint8
}

func New(seed string, maxGoroutineQuantity uint8) *drive {
	if len(seed) == 0 {
		panic(errors.New("种子路径不能为空"))
	}
	return &drive{
		config: &processor.Config{
			Host:   host,
			Seed:   seed,
			Scheme: scheme,
		},
		maxGoroutineQuantity: maxGoroutineQuantity,
	}
}

func (d *drive) GetMaxGoroutineQuantity() int {
	return int(d.maxGoroutineQuantity)
}

func (d *drive) LoadMenuPage() (*goquery.Document, error) {
	u := fmt.Sprintf("%s://%s%s", d.config.Scheme, d.config.Host, d.config.Seed)
	resp, err := http.Get(u)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	return goquery.NewDocumentFromReader(resp.Body)
}

func (d *drive) LoadContentPage(path string) (*goquery.Document, error) {
	u := fmt.Sprintf("%s://%s%s", d.config.Scheme, d.config.Host, path)
	resp, err := http.Get(u)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	return goquery.NewDocumentFromReader(resp.Body)
}

func (d *drive) GetFictionTitle(dom *goquery.Document) string {
	return dom.Find(fictionTitleElement).Text()
}

func (d *drive) GetMenu(dom *goquery.Document) <-chan *processor.MenuMeta {
	nodes := dom.Find(menuElement)
	menus := make(chan *processor.MenuMeta)
	go func() {
		nodes.Each(func(i int, selection *goquery.Selection) {
			path, _ := selection.Attr("href")
			menus <- &processor.MenuMeta{
				Index: i,
				Title: selection.Text(),
				Path:  path,
			}
		})
		close(menus)
	}()
	return menus
}

func (d *drive) GetContent(dom *goquery.Document) (ret []string, err error) {
	content, err := dom.Find(contentElement).Html()
	if err != nil {
		return
	}
	reg := regexp.MustCompile(`(<!--\w+-->)|(<script>\w+</script>)`)
	content = reg.ReplaceAllString(content, "")
	ret = make([]string, 10)
	for _, r := range strings.Split(content, "<br/>") {
		r = strings.TrimSpace(r)
		if len(r) > 0 {
			ret = append(ret, r)
		}
	}
	return
}
