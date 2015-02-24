// Copyright 2010 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package main

import (
	"flag"
	"html/template"
	"io/ioutil"
	"log"
	"net"
	"net/http"
	"regexp"
)

var (
	addr = flag.Bool("addr", false, "find open address and print to final-port.txt")
)

type Page struct {
	Title string
	Body  []byte
	Blogs []Blog
}
type Blog struct {
	Content string
	User    string
}

var Blogs []Blog

func (p *Page) save() error {
	log.Println("save start")
	filename := p.Title + ".txt"
	log.Println("save end")
	return ioutil.WriteFile(filename, p.Body, 0600)
}

func loadPage(title string) (*Page, error) {
	filename := title + ".html"
	log.Printf("loadPage start \"%s\"\n", filename)

	body, err := ioutil.ReadFile(filename)
	if err != nil {
		return nil, err
	}
	log.Println("loadPage end")
	return &Page{Title: title, Body: body}, nil
}

func viewHandler(w http.ResponseWriter, r *http.Request, title string) {
	log.Println("viewHandler start")
	p, err := loadPage(title)
	if err != nil {
		http.Redirect(w, r, "/edit/"+title, http.StatusFound)
		return
	}
	renderTemplate(w, "view", p)
	log.Println("viewHandler end")
}

func indexHandler(w http.ResponseWriter, r *http.Request, title string) {
	log.Println("indexHandler start")
	p, err := loadPage("index")
	if err != nil {
		log.Println("indexHandler err end")
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	renderTemplate(w, "index", p)
	log.Println("indexHandler end")
}

func addHandler(w http.ResponseWriter, r *http.Request, title string) {
	log.Println("addHandler start")

	p, err := loadPage("add")
	log.Printf("p:%s", p)
	var blog Blog
	blog.Content = r.FormValue("Content")
	blog.User = r.FormValue("User")
	Blogs = append(Blogs, blog)
	p.Blogs = Blogs
	log.Printf("blogs:%s", blog)
	log.Printf("p:%s", p.Blogs)

	if err != nil {
		log.Println("addHandler err end")
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	renderTemplate(w, "index", p)
	log.Println("addHandler end")
}

var templates = template.Must(template.ParseFiles("index.html", "view.html"))

func renderTemplate(w http.ResponseWriter, tmpl string, p *Page) {
	log.Println("renderTemplate start")
	err := templates.ExecuteTemplate(w, tmpl+".html", p)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
	log.Println("renderTemplate end")
}

var validPath = regexp.MustCompile("^/(edit|save|view)/([a-zA-Z0-9]+)$")

func makeHandler(fn func(w http.ResponseWriter, r *http.Request, title string)) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		/*m := validPath.FindStringSubmatch(r.URL.Path)
		if m == nil {
			http.NotFound(w, r)
			return
		}*/
		log.Println("makeHandler hadling:" + r.URL.Path + " start")
		fn(w, r, r.URL.Path)
		log.Println("makeHandler hadling:" + r.URL.Path + " end")
	}
}

func main() {
	log.Println("start server")
	flag.Parse()
	http.HandleFunc("/", makeHandler(indexHandler))
	http.HandleFunc("/view/", makeHandler(viewHandler))
	http.HandleFunc("/add/", makeHandler(addHandler))
	Blogs = make([]Blog, 0)

	if *addr {
		l, err := net.Listen("tcp", "127.0.0.1:0")
		if err != nil {
			log.Fatal(err)
		}
		err = ioutil.WriteFile("final-port.txt", []byte(l.Addr().String()), 0644)
		if err != nil {
			log.Fatal(err)
		}
		s := &http.Server{}
		s.Serve(l)
		return
	}

	http.ListenAndServe(":8080", nil)
	log.Println("end server")
}
