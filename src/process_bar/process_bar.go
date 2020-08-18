package process_bar

import (
	"fmt"
	"strings"
	"time"
)

type ProcessBar struct {
	current, total, length       int
	Filling, EmptyFilling        byte
	Title, Format, SuccessString string
	startTime                    time.Time
}

func New(total, length int) (processBar *ProcessBar) {
	return &ProcessBar{
		total:         total,
		length:        length,
		Filling:       '#',
		EmptyFilling:  ' ',
		Title:         "内容爬取中",
		SuccessString: "Success!",
		Format:        "%title% %process% %bar% %time% %status%",
	}
}

func (p *ProcessBar) Start() {
	if p.startTime.IsZero() {
		p.startTime = time.Now()
		p.Print()
	}
}

func (p *ProcessBar) Advance() (err error) {
	if p.current+1 > p.total {
		return p.Error()
	} else if p.current+1 == p.total {
		p.Complete()
		return
	} else {
		if p.current == 0 {
			p.Start()
		}
		p.current++
		p.Print()
		return
	}
}

func (p *ProcessBar) Complete() {
	if p.current < p.total {
		if p.current == 0 {
			p.Start()
		}
		p.current = p.total
		p.PrintComplete()
	}
}

func (p *ProcessBar) Process() (process float64) {
	return float64(p.current) / float64(p.total) * 100
}

func (p *ProcessBar) getBarString() (bar string) {
	process := p.Process()
	if process < 10 {
		bar = strings.ReplaceAll(p.Format, "%process%", fmt.Sprintf("  %.2f%%", process))
	} else if process < 100 {
		bar = strings.ReplaceAll(p.Format, "%process%", fmt.Sprintf(" %.2f%%", process))
	} else {
		bar = strings.ReplaceAll(p.Format, "%process%", fmt.Sprintf("%.2f%%", process))
	}
	bar = strings.ReplaceAll(bar, "%title%", fmt.Sprintf("%s", p.Title))
	bar = strings.ReplaceAll(bar, "%bar%", fmt.Sprintf("[%s]", p.getBar()))
	bar = strings.ReplaceAll(bar, "%time%", fmt.Sprintf("%.1fs", time.Since(p.startTime).Seconds()))
	if process == 100 {
		bar = strings.ReplaceAll(bar, "%status%", p.SuccessString)
	} else {
		bar = strings.ReplaceAll(bar, "%status%", "")
	}
	return
}

func (p *ProcessBar) Print() {
	fmt.Printf("%s\r", p.getBarString())
}

func (p *ProcessBar) PrintComplete() {
	fmt.Printf("%s\n", p.getBarString())
}

func (p *ProcessBar) getBar() (bar string) {
	process := p.Process() / 100 * float64(p.length)
	for i := 0; float64(i) < process; i++ {
		bar += string(p.Filling)
	}
	for j := process + 1; j < float64(p.length); j++ {
		bar += string(p.EmptyFilling)
	}
	return
}

func (p *ProcessBar) Error() (err error) {
	if p.current == p.total {
		return NewError("该进程已完成")
	} else if p.current+1 > p.total {
		return NewError("进度条进度溢出")
	}
	return
}
