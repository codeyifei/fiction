package main

import (
	"bufio"
	"errors"
	"fmt"
	"log"
	"net/http"
	urlPkg "net/url"
	"os"
	"regexp"
	"strings"
	"sync"
	"time"

	"fiction/src/process_bar"
	"github.com/PuerkitoBio/goquery"
	"github.com/fatih/color"
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
var reg = regexp.MustCompile(`(<!--\w+-->)|(<script>\w+</script>)`)

var wg = sync.WaitGroup{}

var ErrRequest = errors.New("请求失败")

func main() {
	defer func() {
		err := recover()
		if err != nil {
			switch err.(type) {
			default:
				log.Println(err)
			}
		}
	}()
	// url := "https://www.biquge.com.cn/book/11029/" // 修真聊天群
	// url := "https://www.biquge.com.cn/book/36681/" // 斗罗大陆四
	url := getUrl()
	fmt.Printf("您想要爬取的网站地址为 %s\n", color.BlueString(url))
	u, err := urlPkg.Parse(url)
	check(err)
	fmt.Print("即将开始爬取...\r")
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
			panic(ErrRequest)
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
		processBar = process_bar.New(ret.Length(), 60)
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

	f, err := os.Create("results/" + fiction.title + ".txt")
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

func getUrl() (url string) {
	if len(os.Args) > 1 {
		url = strings.TrimSpace(os.Args[1])
	} else {
		url = scanUrl()
	}
	return
}

func scanUrl() (url string) {
	fmt.Print("请输入你想要爬取的网站地址：")
	n, err := fmt.Scanln(&url)
	if n == 0 {
		color.Red("网站地址不能为空")
		return strings.TrimSpace(scanUrl())
	}
	check(err)
	return
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
			content: getContext(m.path, hostname, scheme, m.title),
		}
		check(processBar.Advance())
	}
}

func getContext(path, hostname, scheme, title string) []string {
	dom, err := getDomRecursion(scheme+"://"+hostname+path, 0)
	if err != nil {
		if err, ok := err.(*charterErr); ok {
			err.title = title
			fmt.Println(err.Error())
		} else {
			fmt.Println("出现未知错误，跳过该章节")
		}
		return nil
	}

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

type charterErr struct {
	title string
	url   string
}

func (e *charterErr) Error() string {
	if len(e.title) == 0 {
		return fmt.Sprintf("章节加载失败: [%s]", e.url)
	}
	return fmt.Sprintf("章节加载失败: %s[%s]", e.title, e.url)
}

func getDomRecursion(url string, count int) (dom *goquery.Document, err error) {
	resp, err := http.Get(url)
	if err != nil {
		if count < 100 {
			count++
			return getDomRecursion(url, count)
		} else {
			if e, ok := err.(*urlPkg.Error); ok {
				return nil, &charterErr{url: e.URL}
			}
		}
	}
	defer func() {
		check(resp.Body.Close())
	}()
	return goquery.NewDocumentFromReader(resp.Body)
}

type ErrHandleFunc func(error)

func check(err error) {
	handleCheck(err, func(err error) {
		panic(err)
	})
}

func handleCheck(err error, handleFunc ErrHandleFunc) {
	if err != nil {
		handleFunc(err)
	}
}
