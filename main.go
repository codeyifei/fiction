package main

import (
	"log"

	"github.com/codeyifei/fiction/src/drive/biquge"
	"github.com/codeyifei/fiction/src/processor"
)

func main() {
	defer func() {
		if err := recover(); err != nil {
			log.Fatalln(err)
		}
	}()
	d := biquge.New("/16_16431/", 100)
	p := processor.New(d)
	if err := p.Run(); err != nil {
		panic(err)
	}
}
