// genpage is a tool for generating pages containing photos.
package main

import (
	"flag"
	"html/template"
	"io"
	"log"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/mewkiz/pkg/errutil"
	"github.com/mewkiz/pkg/imgutil"
	"github.com/nfnt/resize"
)

// Verbose output.
var verbose bool

func main() {
	// Command line flags.
	var (
		// Maximum thumbnail dimensions (width or height).
		max uint
		// Output directory.
		outputDir string
		// Page title.
		title string
	)
	flag.UintVar(&max, "max", 320, "maximum thumbnail dimensions (width or height).")
	flag.StringVar(&outputDir, "o", "output", "output directory.")
	flag.StringVar(&title, "title", "unknown", "page title.")
	flag.BoolVar(&verbose, "v", false, "verbose output.")
	flag.Parse()
	paths := flag.Args()
	page := NewPage(title, paths)
	if err := dumpPage(outputDir, page, max); err != nil {
		log.Fatal(err)
	}
}

// dumpPage creates an index file for the given page, and stores its photos and
// thumbnails of the given maximum dimension in prescribed directory structures.
//
//    output/
//    output/index.html
//    output/img/
//    output/img/foo.jpg
//    output/img/bar.jpg
//    output/thumbs/
//    output/thumbs/foo.jpg
//    output/thumbs/bar.jpg
func dumpPage(outputDir string, page *Page, max uint) error {
	imgDir := filepath.Join(outputDir, "img")
	thumbsDir := filepath.Join(outputDir, "thumbs")
	// Create directory structure.
	if verbose {
		log.Printf("Creating directory %q", imgDir)
	}
	if err := os.MkdirAll(imgDir, 0755); err != nil {
		return errutil.Err(err)
	}
	if verbose {
		log.Printf("Creating directory %q", thumbsDir)
	}
	if err := os.MkdirAll(thumbsDir, 0755); err != nil {
		return errutil.Err(err)
	}

	// Store images.
	for _, photo := range page.Photos {
		src := photo.Path
		name := filepath.Base(src)
		dst := filepath.Join(imgDir, name)
		if verbose {
			log.Printf("Storing image %q", dst)
		}
		if err := CopyFile(dst, src); err != nil {
			return errutil.Err(err)
		}
	}

	// Create thumbnails.
	for _, photo := range page.Photos {
		src := photo.Path
		name := filepath.Base(src)
		dst := filepath.Join(thumbsDir, name)
		if verbose {
			log.Printf("Storing thumbnail %q", dst)
		}
		if err := createThumbnail(dst, src, max); err != nil {
			return errutil.Err(err)
		}
	}

	// Create index.html.
	indexPath := filepath.Join(outputDir, "index.html")
	if verbose {
		log.Printf("Creating %q", indexPath)
	}
	if err := dumpIndex(indexPath, page); err != nil {
		return errutil.Err(err)
	}
	return nil
}

// createThumbnail creates a thumbnail of the source image and stores it at the
// given destination, based on the given maximum dimension.
func createThumbnail(dstPath, srcPath string, max uint) error {
	src, err := imgutil.ReadFile(srcPath)
	if err != nil {
		return errutil.Err(err)
	}
	thumb := resize.Thumbnail(max, max, src, resize.Bicubic)
	if err := imgutil.WriteFile(dstPath, thumb); err != nil {
		return errutil.Err(err)
	}
	return nil
}

// dumpPage creates an index file for the given page, and stores its photos and
// thumbnails in prescribed directory structures.
func dumpIndex(indexPath string, page *Page) error {
	t := template.New("index")
	funcMap := template.FuncMap{
		"base": filepath.Base,
	}
	t = t.Funcs(funcMap)
	const index = `
<!DOCTYPE html>
<html>
	<head>
		<meta charset='utf-8'>
		<title>{{.Title}}</title>
		<style>
			#m {
				width: 800px;
				margin: 20px auto;
			}
			a img {
				border: 1px solid #000000;
			}
			h1 {
				font-size: 22px;
				border-bottom: 1px solid #CCCCCC;
			}
		</style>
	</head>
	<body>
		<div id='m'>
			<h1>{{.Title}}</h1>
{{range .Photos -}}
			<a href='img/{{base .Path}}' title='{{.Desc}}'><img src='thumbs/{{base .Path}}' alt='{{.Desc}}'></a>
{{end}}
		</div>
	</body>
</html>
`
	if _, err := t.Parse(index[1:]); err != nil {
		return errutil.Err(err)
	}
	fw, err := os.Create(indexPath)
	if err != nil {
		return errutil.Err(err)
	}
	defer fw.Close()
	if err := t.Execute(fw, page); err != nil {
		return errutil.Err(err)
	}
	return nil
}

// A Page represents a page containing photos.
type Page struct {
	// Page title.
	Title string
	// Photos.
	Photos []*Photo
}

// A Photo represents a photo with a description.
type Photo struct {
	// File path.
	Path string
	// Image description.
	Desc string
}

// NewPage returns a new page with the given title and set of photos.
func NewPage(title string, paths []string) *Page {
	sort.Strings(paths)
	page := &Page{
		Title: title,
	}
	for _, path := range paths {
		desc := title
		if tags := getTags(path); len(tags) > 0 {
			desc = title + " " + tags
		}
		photo := &Photo{
			Path: path,
			Desc: desc,
		}
		page.Photos = append(page.Photos, photo)
	}
	return page
}

// getTags returns the tags within a given file name. E.g.
//
//    Input:  "2013-03-26 - 0001 [Daniel] [Pat] [Rita].jpg"
//    Output: "[Daniel] [Pat] [Rita]"
func getTags(path string) string {
	start := strings.IndexByte(path, '[')
	if start == -1 {
		return ""
	}
	end := strings.LastIndexByte(path, ']')
	if start == -1 {
		return ""
	}
	return path[start : end+1]
}

// CopyFile copies the source file to the given destination.
func CopyFile(dst, src string) error {
	in, err := os.Open(src)
	if err != nil {
		return errutil.Err(err)
	}
	defer in.Close()
	out, err := os.Create(dst)
	if err != nil {
		return errutil.Err(err)
	}
	if _, err := io.Copy(out, in); err != nil {
		if err := out.Close(); err != nil {
			return errutil.Err(err)
		}
		return errutil.Err(err)
	}
	if err := out.Close(); err != nil {
		return errutil.Err(err)
	}
	return nil
}
