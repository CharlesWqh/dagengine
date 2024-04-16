package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	"xxxx/dagengine/engine"
)

// WebRes response
type WebRes struct {
	Path string `json:",omitempty"`
	Err  string `json:",omitempty"`
}

func main() {
	log.Printf("Start web server")
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "edit.html")
	})
	http.Handle("/pngs/", http.StripPrefix("/pngs/", http.FileServer(http.Dir("./pngs"))))
	var cursor int64
	http.HandleFunc("/gen_png", func(w http.ResponseWriter, r *http.Request) {
		_ = r.ParseForm()
		ops := r.FormValue("ops")
		script := r.FormValue("script")
		// log.Printf("Receive form ops:%s", ops)
		// log.Printf("Receive form script:%s", script)
		dag, err := engine.NewDAGConfigByContent(ops, script)
		rs := &WebRes{}
		if nil != err {
			rs.Err = fmt.Sprintf("%v", err)
			b, err1 := json.Marshal(rs)
			if err1 != nil {
				return
			}
			w.WriteHeader(400)
			w.Header().Set("Content-Type", "application/json")
			_, _ = w.Write(b)
			return
		}
		path := fmt.Sprintf("/pngs/%d", cursor)
		err = dag.GenPng("." + path)
		if nil != err {
			log.Printf("Error:%v", err)
			rs.Err = fmt.Sprintf("%v", err)
			b, err1 := json.Marshal(rs)
			if err1 != nil {
				return
			}
			w.WriteHeader(400)
			w.Header().Set("Content-Type", "application/json")
			_, _ = w.Write(b)
			return
		}
		png := path + ".png"

		rs.Path = png
		// w.WriteHeader(200)
		b, err := json.Marshal(rs)
		if err != nil {
			return
		}

		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write(b)
		log.Printf("Response:%v", string(b))
		cursor++
		if 100 == cursor {
			cursor = 0
		}
	})

	log.Fatal(http.ListenAndServe(":8080", nil))

}
