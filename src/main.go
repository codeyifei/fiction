package main

import (
	"bufio"
	"fiction/src/process_bar"
	"fmt"
	"github.com/PuerkitoBio/goquery"
	"github.com/fatih/color"
	"log"
	"net/http"
	urlPkg "net/url"
	"os"
	"regexp"
	"strings"
	"sync"
	"time"
)

type fictionList struct {
	title   string
	content []string
}

type fiction struct {
	title string
	lists []fictionList
}

type charterContent struct {
	index   int
	title   string
	content []string
}

type meta struct {
	path  string
	index int
	title string
}

var charterChan = make(chan charterContent, 100)
var metaChan = make(chan meta, 100)
var hostname string
var scheme string
var charterMap = make(map[int]charterContent)
var processBar *process_bar.ProcessBar
var reg = regexp.MustCompile(`<!--\w+-->`)

var wg = sync.WaitGroup{}

func main() {
	//url := "https://www.biquge.com.cn/book/11029/" // 修真聊天群
	url := "https://www.biquge.com.cn/book/36681/" // 斗罗大陆四
	u, err := urlPkg.Parse(url)
	check(err)
	hostname = u.Hostname()
	scheme = u.Scheme
	startTime := time.Now()

	var received *http.Response
	for i := 0; i < 3; i++ {
		received, err = http.Get(url)
		if err == nil && received.StatusCode == http.StatusOK {
			break
		}
		if i == 2 {
			log.Panicln("请求失败")
		} else {
			log.Printf("请求失败，正在第%d次重试", i+1)
		}
	}
	defer func() {
		check(received.Body.Close())
	}()

	dom, err := goquery.NewDocumentFromReader(received.Body)
	check(err)

	var fiction = fiction{}

	for i := 0; i < 100; i++ {
		wg.Add(1)
		go loadContext()
	}

	fiction.title = dom.Find(".box_con #maininfo #info h1").Text()

	go func() {
		defer func() {
			close(metaChan)
			wg.Wait()
			processBar.Complete()
			fmt.Print(color.GreenString("爬取完成，继续写入文件中..."))
			close(charterChan)
		}()
		ret := dom.Find("#list dl dd a")
		processBar = process_bar.New(ret.Length(), 100)
		processBar.SuccessString = color.GreenString("Success!")
		processBar.Start()
		ret.Each(func(i int, selection *goquery.Selection) {
			itemUrl, _ := selection.Attr("href")

			metaChan <- meta{
				path:  itemUrl,
				index: i,
				title: selection.Text(),
			}
		})
	}()

	f, err := os.Create("results/" + fiction.title + "（爬虫）.txt")
	check(err)
	w := bufio.NewWriter(f)
	_, err = w.WriteString(fmt.Sprintf("【%s】\n", fiction.title))
	check(err)
	defer func() {
		check(f.Close())
	}()
	var next int
	for item := range charterChan {
		context := handleContent(next, item)
		if context != nil {
			check(writeFile(context, w))
			next++
		}
	}
	for len(charterMap) > 0 {
		context := charterMap[next]
		check(writeFile(&context, w))
		delete(charterMap, next)
		next++
	}

	color.Green("\r\033[K写入完成，共用时%.2fs", time.Since(startTime).Seconds())
}

func handleContent(next int, item charterContent) (context *charterContent) {
	if item.index == next {
		return &item
	} else {
		charterMap[item.index] = item

		if charter, ok := charterMap[next]; ok {
			delete(charterMap, next)
			return &charter
		} else {
			return nil
		}
	}
}

func writeFile(content *charterContent, w *bufio.Writer) (err error) {
	defer func() {
		check(w.Flush())
	}()
	_, err = w.WriteString("\n" + content.title + "\n")
	for _, c := range content.content {
		_, err = w.WriteString(c + "\n")
	}
	return
}

func loadContext() {
	defer wg.Done()
	for m := range metaChan {
		charterChan <- charterContent{
			index:   m.index,
			title:   m.title,
			content: getContext(m.path, hostname, scheme),
		}
		check(processBar.Advance())
	}
}

func getContext(path, hostname, scheme string) []string {
	var resp *http.Response
	var err error
	for {
		resp, err = http.Get(scheme + "://" + hostname + path)
		if err == nil {
			break
		}
		log.Println(err)
	}

	dom, err := goquery.NewDocumentFromReader(resp.Body)
	check(err)

	ret := make([]string, 0)
	content, err := dom.Find("#content").Html()
	check(err)
	content = reg.ReplaceAllString(content, "")
	for _, r := range strings.Split(content, "<br/>") {
		r = strings.TrimSpace(r)
		if len(r) > 0 {
			ret = append(ret, r)
		}
	}
	return ret
}

func check(err error) {
	//defer func() {
	//	if err := recover(); err != nil {
	//		switch err.(type) {
	//		case *process_bar.Error:
	//			log.Println("进度条异常", err)
	//		default:
	//			log.Panicln(reflect.TypeOf(err))
	//		}
	//	}
	//}()
	if err != nil {
		panic(err)
	}
}
