package main

import (
	"fmt"
	"html/template"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"
)

type tmpDetail struct {
	tmName      string
	tmFileNames []string
}

// lists of the template name and the files to parse together
var tmpDetails []tmpDetail

var supportedLangs = []string{"cn", "tw", "en"}

//var re = regexp.MustCompile("([[:alpha:]]*)-([[:alpha:]]{2}).html$")
var tmplDir = "template"
var baseTemplate = "baseof.html"
var tmplSuffix = ".html"
var tmpls map[string]*template.Template

// parse all the templates
func init() {
	tmpls = make(map[string]*template.Template)
	tempFolder, err := os.Open(tmplDir)
	if err != nil {
		log.Fatal(err)
	}
	teplFileInfos, err := tempFolder.Readdir(-1)
	if err != nil {
		log.Fatal(err)
	}

	for _, fi := range teplFileInfos {
		var d tmpDetail
		name := fi.Name()
		if name == baseTemplate {
			continue
		}
		d.tmName = strings.TrimSuffix(name, tmplSuffix)
		d.tmFileNames = []string{filepath.Join(tmplDir, baseTemplate), filepath.Join(tmplDir, name)}
		tmpDetails = append(tmpDetails, d)

	}

	fmt.Println(tmpDetails)

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

	fi, err := os.Stat(dst)
	if err != nil {
		if os.IsNotExist(err) {
			if err := os.MkdirAll(dst, 0777); err != nil {
				fmt.Printf("can't create folder %v\n", dst)
				fmt.Fprint(w, err)
				return
			}
		} else {
			fmt.Printf("os.Stat error: %v\n", dst)
			fmt.Fprint(w, err)
		}

	} else {
		if !fi.IsDir() {
			fmt.Fprint(w, "dst is not a folder")
		}
	}

	fmt.Print("Generating files... ")
	var staticFiles []string
	for _, t := range tmpDetails {
		for _, l := range []string{"en", "cn", "tw"} {
			staticFiles = append(staticFiles, fmt.Sprintf("%v-%v.html", t.tmName, l))
		}
	}
	//fmt.Println(staticFiles)

	for _, fileName := range staticFiles {
		fullName := fmt.Sprintf("%v/%v", dst, fileName)
		fmt.Println(fullName)
		if err := writeToFile(fullName); err != nil {
			fmt.Fprint(w, err)
			return
		}
	}
	fmt.Println("done.")

	fmt.Print("Copying static folder... ")
	if err := copyAll("static", filepath.Join(dst, "static")); err != nil {
		fmt.Print("failed: ", err)
		fmt.Fprintf(w, "internal error: failed to copy static folder: %v", err)
		return
	}
	fmt.Println("Done.")

	fmt.Fprint(w, "All Done.")
}

func writeToFile(fullName string) error {

	f, err := os.Create(fullName)
	if err != nil {
		fmt.Println(" ...Failed.", err)
		return err
	}
	defer f.Close()

	return hdTemplate(f, fullName)
}

func hdTemplate(w io.Writer, path string) error {

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

func copyAll(src, dst string) error {
	srcDir := src
	//srcDir := `c:\andrew\prg\sqlite`
	dstDir := dst

	c := 0
	err := filepath.Walk(srcDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return fmt.Errorf("walk err - path: %v: %v", path, err)
		}
		// normal file
		df := filepath.Join(dstDir, strings.TrimLeft(path, srcDir))

		if info.IsDir() {
			fmt.Println("in isdir", path)
			if err = os.MkdirAll(df, 0777); err != nil {
				if !os.IsExist(err) {
					return fmt.Errorf("mkdir err - path: %v: %v", df, err)
				}
				log.Println("created dir: ", path)
				return nil
			}
			fmt.Println("dir created", df)
			return nil
		}
		dh, err := os.Create(df)
		if err != nil {
			return fmt.Errorf("create file err - name: %v: %v", df, err)
		}
		defer dh.Close()

		sh, err := os.Open(path)
		if err != nil {
			return fmt.Errorf("open file err - name: %v: %v", path, err)
		}
		defer sh.Close()

		_, err = io.Copy(dh, sh)
		c++
		fmt.Println(df, "done.")
		return err
	})

	fmt.Println(c)
	return err
}
