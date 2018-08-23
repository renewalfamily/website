package main

import (
	"fmt"
	"html/template"
	"io"
	"log"
	"net/http"
	"os"
	"strings"
)

type tmpDetail struct {
	tmName      string
	tmFileNames []string
}

// lists of the template name and the files to parse together
var tmpDetails = []tmpDetail{
	{"index", []string{"template/baseof.html", "template/index.html"}},
	{"renewal-family", []string{"template/baseof.html", "template/renewal-family.html"}},
	{"city-prayer-alter", []string{"template/baseof.html", "template/city-prayer-alter.html"}},
	{"together", []string{"template/baseof.html", "template/together.html"}},
}

var supportedLangs = []string{"cn", "tw", "en"}

//var re = regexp.MustCompile("([[:alpha:]]*)-([[:alpha:]]{2}).html$")

var baseTemplate = "baseof.html"
var tmpls map[string]*template.Template

// parse all the templates
func init() {
	tmpls = make(map[string]*template.Template)
	for _, d := range tmpDetails {
		tmpls[d.tmName] = template.Must(template.New(d.tmName).ParseFiles(d.tmFileNames...))
	}
}

func main() {
	// for static files
	fs := http.StripPrefix("/static/", http.FileServer(http.Dir("./static/")))
	http.Handle("/static/", fs)

	// generate static files
	http.HandleFunc("/gen", hdGen)

	// catch all
	http.HandleFunc("/", hdIndex)
	log.Println("listening at :1112")
	log.Println(http.ListenAndServe(":1112", nil))
}

func hdIndex(w http.ResponseWriter, r *http.Request) {
	fmt.Println(r.URL)

	// language specific file using name-lang.html format to easy generate static files
	// accepted lang codes are: cn, tw, and en
	// example: renewal-family-cn.html, index-tw.html
	path := r.URL.Path

	err := hdTemplate(w, path)
	if err != nil {
		fmt.Println(err)
	}
}

func hdGen(w http.ResponseWriter, r *http.Request) {
	dst := "dst"
	staticFiles := []string{
		"index-cn.html", "index-tw.html", "index-en.html",
		"renewal-family-cn.html", "renewal-family-tw.html", "renewal-family-en.html",
		"city-prayer-alter-cn.html", "city-prayer-alter-tw.html", "city-prayer-alter-en.html",
		"together-cn.html", "together-tw.html", "together-en.html",
	}
	for _, fileName := range staticFiles {
		fullName := fmt.Sprintf("%v/%v", dst, fileName)
		fmt.Println(fullName)
		writeToFile(fullName)
	}
	fmt.Println("done.")
	fmt.Fprintf(w, "Done.", nil)
}

func writeToFile(fullName string) {

	f, err := os.Create(fullName)
	if err != nil {
		fmt.Println(" ...Failed.")
		log.Fatal(err)
		return
	}
	defer f.Close()

	hdTemplate(f, fullName)
}

func hdTemplate(w io.Writer, path string) error {

	//match := re.FindAllStringSubmatch(path, 1)
	//if match == nil {
	//	return fmt.Errorf("hdIndex: can't find match filename and land from path: %v\n", path)
	//}
	//
	//tmName := match[0][1]
	//_, ok := tmpls[tmName]
	//if !ok {
	//	fmt.Printf("hdIndex: can't find the template %v", tmName)
	//}
	//
	//lang := match[0][2]

	tmName := ""
	for _, name := range tmpDetails {
		if strings.Contains(path, name.tmName) {
			tmName = name.tmName
			break
		}
	}
	if tmName == "" {
		return fmt.Errorf("hdTemplate: can't find match filename and land from path: %v\n", path)
	}

	lang := ""
	for _, l := range supportedLangs {
		if strings.HasSuffix(path, l+".html") {
			lang = l
			break
		}
	}
	if lang == "" {
		return fmt.Errorf("hdTemplate: wrong Lang in url: %v", path)
	}
	return tmpls[tmName].ExecuteTemplate(w, baseTemplate, struct {
		Lang     string
		PageName string
	}{
		lang,
		tmName,
	})
}
