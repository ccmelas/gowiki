package main

import (
	"errors"
	"html/template"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

var templates = template.Must(template.ParseGlob("templates/*.html"))
var validPath = regexp.MustCompile("^/(edit|save|view|delete)/([a-zA-Z0-9 ]+)$")

type Page struct {
	Title string
	Body []byte
}

func (p *Page) getPath() string {
	return "data/" + p.Title + ".txt"
}

func (p *Page) save() error {
	return ioutil.WriteFile(p.getPath(), p.Body, 0600)
}

func (p *Page) delete() error {
	return os.Remove(p.getPath())
}


func loadPage(title string) (*Page, error) {
	filename := "data/" + title + ".txt"
	body, err := ioutil.ReadFile(filename)
	if err != nil {
		return nil, err
	}
	return &Page{ title, body }, nil
}

func renderTemplate(w http.ResponseWriter, tpl string, data interface{}) {
	err := templates.ExecuteTemplate(w, tpl + ".html", data)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

func getTitle(w http.ResponseWriter, r *http.Request) (string, error) {
	m := validPath.FindStringSubmatch(r.URL.Path)
	if m == nil {
		http.NotFound(w, r)
		return "", errors.New("invalid page title")
	}
	return m[2], nil
}

func makeHandler(fn func(http.ResponseWriter, *http.Request, string)) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		title, err := getTitle(w, r)
		if err != nil {
			return
		}
		fn(w, r, title)
	}
}

func indexHandler(w http.ResponseWriter, r *http.Request) {
	var pages []*Page
	wikis,err := ioutil.ReadDir("data")
	if err != nil {

	}
	for _, wiki := range wikis {
		filename := wiki.Name()
		if filename == ".gitkeep" {
			continue
		}
		content,_ := ioutil.ReadFile(filename)
		page := Page{ Title: strings.TrimSuffix(filename, filepath.Ext(filename)), Body: content }
		pages = append(pages, &page)
	}

	renderTemplate(w, "index", pages)
}

func createHandler(w http.ResponseWriter, r *http.Request) {
	renderTemplate(w, "create", nil)
}

func viewHandler(w http.ResponseWriter, r *http.Request, title string) {
	p, err := loadPage(title)
	if err != nil {
		http.Redirect(w, r,"/edit/" + title, http.StatusFound)
	}
	renderTemplate(w, "view", p)
}

func editHandler(w http.ResponseWriter, r *http.Request, title string) {
	p, err := loadPage(title)
	if err != nil {
		p = &Page{ Title: title }
	}
	renderTemplate(w, "edit", p)
}

func saveHandler(w http.ResponseWriter, r *http.Request, title string) {
	body := r.FormValue("body")
	p := &Page{ Title: title, Body: []byte(body) }
	err := p.save()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	http.Redirect(w, r, "/view/" + title, http.StatusFound)
}

func storeHandler(w http.ResponseWriter, r *http.Request) {
	body := r.FormValue("body")
	title := r.FormValue("title")
	p := &Page{ Title: title, Body: []byte(body) }
	err := p.save()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	http.Redirect(w, r, "/view/" + title, http.StatusFound)
}

func deleteHandler(w http.ResponseWriter, r *http.Request, title string) {
	p, err := loadPage(title)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError);
		return
	}
	err = p.delete();
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError);
		return
	}
	http.Redirect(w, r, "/", http.StatusFound)
}

func main() {
	fs := http.FileServer(http.Dir("./public"))
	http.Handle("/public/", http.StripPrefix("/public/", fs));
	http.HandleFunc("/", indexHandler)
	http.HandleFunc("/create", createHandler)
	http.HandleFunc("/store", storeHandler)
	http.HandleFunc("/view/", makeHandler(viewHandler))
	http.HandleFunc("/edit/", makeHandler(editHandler))
	http.HandleFunc("/save/", makeHandler(saveHandler))
	http.HandleFunc("/delete/", makeHandler(deleteHandler))
	log.Fatal(http.ListenAndServe(":8080", nil))
}