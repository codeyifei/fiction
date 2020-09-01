package processor

import (
	"log"
	"sync"
)

type processor struct {
	drive *ReptileDriver
}

type Content struct {
	Index   int
	Title   string
	Content []string
}

func New(d ReptileDriver) *processor {
	return &processor{drive: &d}
}

func (p *processor) Run() error {
	d := *p.drive
	maxGoroutineQuantity := d.GetMaxGoroutineQuantity()
	menuChan := make(chan *MenuMeta, maxGoroutineQuantity)
	contentChan := make(chan *Content)
	errChan := make(chan error)
	wg := sync.WaitGroup{}
	for i := 0; i < maxGoroutineQuantity; i++ {
		wg.Add(1)
		go func(menuChan <-chan *MenuMeta, conChan chan<- *Content, errChan chan<- error) {
			defer wg.Done()
			for menu := range menuChan {
				dom, err := d.LoadContentPage(menu.Path)
				if err != nil {
					errChan <- err
				}
				con, err := d.GetContent(dom)
				if err != nil {
					errChan <- err
				}
				conChan <- &Content{
					Index:   menu.Index,
					Title:   menu.Title,
					Content: con,
				}
			}
		}(menuChan, contentChan, errChan)
	}
	dom, err := d.LoadMenuPage()
	if err != nil {
		return err
	}
	go func() {
		for err = range errChan {
			log.Println(err)
		}
	}()
	go func(conChan <-chan *Content) {
		for range conChan {
			// fmt.Println(con.Index)
		}
	}(contentChan)
	for m := range d.GetMenu(dom) {
		menuChan <- m
	}
	close(menuChan)
	wg.Wait()
	close(contentChan)
	close(errChan)
	return nil
}
