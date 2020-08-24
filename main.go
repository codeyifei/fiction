package main

import (
	"github.com/codeyifei/fiction/src/drive/biquge"
	"github.com/codeyifei/fiction/src/processor"
)

func main() {
	d := biquge.New("", 100)
	p := processor.New(d)
	p.Run()
}
