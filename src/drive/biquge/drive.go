package biquge

import (
	"fmt"
	"net/http"

	"github.com/PuerkitoBio/goquery"
	"github.com/codeyifei/fiction/src/processor"
)

const (
	fictionTitleElement string = ".box_con #maininfo #info h1"
	menuElement                = "#list dl dd a"
	contentElement             = "#content"
)

type Drive struct {
	Config               *processor.Config
	MaxGoroutineQuantity int
	// MenuChan
}

func New(seed string) (drive *Drive) {
	return &Drive{Config: &processor.Config{
		Host:   host,
		Seed:   seed,
		Scheme: scheme,
	}}
}

func (d *Drive) LoadMenuPage() (*goquery.Document, error) {
	u := fmt.Sprintf("%s://%s%s", d.Config.Scheme, d.Config.Host, d.Config.Seed)
	resp, err := http.Get(u)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	return goquery.NewDocumentFromReader(resp.Body)
}

func (d *Drive) LoadContentPage(path string) (*goquery.Document, error) {
	u := fmt.Sprintf("%s://%s%s", d.Config.Scheme, d.Config.Host, path)
	resp, err := http.Get(u)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	return goquery.NewDocumentFromReader(resp.Body)
}

func (d *Drive) GetFictionTitle(dom *goquery.Document) string {
	return dom.Find(fictionTitleElement).Text()
}

// func (d *Drive) GetMenu(dom *goquery.Document) ([]string, error) {
// 	 dom.Find(menuElement).
// }

// func (d *Drive) getContent(dom *goquery.Document) ([]string, error) {
//
// }
