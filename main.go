package main

import (
	"bytes"
	"flag"
	"fmt"
	"html"
	"io/ioutil"
	"log"
	"os"
	"path"
	"sort"
	"strings"

	"golang.org/x/text/language"
	"golang.org/x/text/message"
)

var (
	recursiveOpt = flag.Bool("r", false, "recursive")
	hidden       = flag.Bool("h", false, "include hidden dotfiles")

	icons = map[string]string{
		"default": "file-alt",
		".epub":   "book",
		".jpeg":   "file-image",
		".jpg":    "file-image",
		".gif":    "file-image",
		".tiff":   "file-image",
		".pdf":    "file-pdf",
		".mov":    "file-movie",
		".avi":    "file-movie",
		".zip":    "file-archive",
		".rar":    "file-archive",
		".tar":    "file-archive",
		".tgz":    "file-archive",
	}

	decimalFormatter = message.NewPrinter(language.English)
)

const magic = "<!-- generated by static-indexes -->"

func paresStList(stList []os.FileInfo) ([]string, map[string]os.FileInfo) {
	m := map[string]os.FileInfo{}
	names := []string{}
	for _, fi := range stList {
		fileName := fi.Name()
		if strings.HasPrefix(fileName, ".") && !*hidden {
			continue
		}
		m[fileName] = fi
		names = append(names, fileName)
	}
	sort.Strings(names)
	return names, m
}

func safeToReplace(fn string) bool {
	b, err := ioutil.ReadFile(fn)
	if err != nil {
		return true
	}
	if bytes.Contains(b, []byte(magic)) {
		return true
	}
	return false
}

func generateTable(dir string, stList []os.FileInfo) string {
	names, m := paresStList(stList)

	builder := strings.Builder{}

	builder.WriteString(`<table class="sortable">
	<thead>
		<tr>
			<th class="no-sort"></th>
			<th>Name</th>
			<th>Date</th>
		</tr>
	</thead>
	<tbody>
`)

	for _, name := range names {
		if name != "index.htm;" {
			builder.WriteString(row(m[name]))
		}
	}

	builder.WriteString(
		`
		<tr>
			<td></td>
		</tr>
	</tbody>
</table>`)

	return builder.String()
}

func generateIndex(dir string, stList []os.FileInfo) error {
	fn := path.Join(dir, "index.html")
	log.Printf("gnerateIndex: %s", fn)
	if !safeToReplace(fn) {
		log.Printf("%s: not safe to replace", fn)
		return nil
	}

	builder := strings.Builder{}
	builder.WriteString(`<html>`)
	builder.WriteString(magic)
	builder.WriteString(`<head>`)
	builder.Write(MustAsset("assets/head.header"))

	if b, err := ioutil.ReadFile(path.Join(dir, "HEADER.html")); err == nil {
		builder.Write(b)
	}

	builder.WriteString(`</head>`)
	builder.WriteString(`<body>`)
	builder.Write(MustAsset("assets/body.footer"))

	builder.WriteString(fmt.Sprintf(`<div><a href=".."><i class="fas fa-%s"></i> up</a></div>`, "up-arrow"))

	builder.WriteString(generateTable(dir, stList))

	if b, err := ioutil.ReadFile(path.Join(dir, "README.html")); err == nil {
		builder.Write(b)
	}

	builder.WriteString(`</body>`)

	builder.WriteString(`</html>`)

	return ioutil.WriteFile(fn, []byte(builder.String()), 0644)
}

func row(fi os.FileInfo) string {
	fn := fi.Name()
	isDir := fi.IsDir()
	lc := strings.ToLower(fn)
	icon := icons[lc]
	if icon == "" {
		icon = icons["default"]
	}
	if isDir {
		icon = "folder-open"
		fn = fn + "/"
	}

	fontAwesome := fmt.Sprintf(`	<i class="fas fa-%s"></i>`, icon)
	link := fmt.Sprintf(`<a href=%q>%s</a>`, fn, html.EscapeString(fn))

	when := fi.ModTime().Format("2006-01-02 15:04:05 MST")

	return fmt.Sprintf(`
<tr>
<td>%s</td>
<td>%s</td>
<td>%s</td>
</tr>`,
		fontAwesome, link, when,
	)

}

func process(dir string, top bool) error {
	st, err := os.Stat(dir)
	if err != nil {
		return fmt.Errorf("%s: %w", dir, err)
	}
	if !st.IsDir() {
		return fmt.Errorf("%s: not a dir", dir)
	}

	rd, err := os.Open(dir)
	if err != nil {
		return fmt.Errorf("%s: open dir:", dir)
	}
	stList, err := rd.Readdir(0)
	if err != nil {
		return fmt.Errorf("%s: read dir:", dir)
	}

	if err := generateIndex(dir, stList); err != nil {
		return err
	}
	if !*recursiveOpt {
		return nil
	}

	for _, d := range stList {
		if d.IsDir() {
			if strings.HasPrefix(d.Name(), ".") && *hidden {
				continue
			}
			if err := process(path.Join(dir, d.Name()), false); err != nil {
				return err
			}
		}
	}
	return nil
}

func main() {
	flag.Parse()
	for _, dest := range flag.Args() {
		if err := process(dest, true); err != nil {
			log.Fatal(err)
		}
	}
}
