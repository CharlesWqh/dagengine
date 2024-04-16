package main

import (
	"flag"
	"log"

	"xxxx/dagengine/engine"
)

func main() {
	meta := flag.String("meta", "", "Specify input op meta file")
	script := flag.String("toml", "", "Specify input toml script")
	flag.Parse()

	if len(*meta) == 0 || len(*script) == 0 {
		flag.Usage()
		return
	}
	cfg, err := engine.NewDAGConfigByFile(*meta, *script)
	if err != nil {
		log.Printf("%v", err)
		return
	}
	if err = cfg.GenPng(""); err != nil {
		log.Printf("%v", err)
		return
	}
}
