package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"strings"
	"text/template"
	"time"
)

const (
	ipaName   = ".ipa"
	plistName = ".plist"
	plist     = `<?xml version="1.0" encoding="UTF-8"?>
<!DOCTYPE plist PUBLIC "-//Apple//DTD PLIST 1.0//EN" "http://www.apple.com/DTDs/PropertyList-1.0.dtd">
<plist version="1.0">
<dict>
	<key>items</key>
	<array>
		<dict>
			<key>assets</key>
			<array>
				<dict>
					<key>kind</key>
					<string>software-package</string>
					<key>url</key>
					<string>{{.Ipa_url}}</string>
				</dict>
			</array>
			<key>metadata</key>
			<dict>
				<key>bundle-identifier</key>
				<string>{{.Bundle_id}}</string>
				<key>bundle-version</key>
				<string>{{.Version}}</string>
				<key>kind</key>
				<string>software</string>
				<key>title</key>
				<string>{{.AppName}}</string>
			</dict>
		</dict>
	</array>
</dict>
</plist>
`
)

const (
	Dir = iota
	File
)

var (
	scheme   string = "http"
	baseDir  string = "/"
	indexTem        = "index.html"
)

type Model struct {
	Ipa_url   string
	Bundle_id string
	Version   string
	AppName   string
}

type sampleInfo struct {
	Name       string
	CreateTime time.Time
}

type FileInfo struct {
	FileName   string
	Url        string
	CreateTime string
	Type       int
}

func getAppInfo(ipafile string) map[string]interface{} {
	sh := "." + "/genplist.sh"
	cmd := exec.Command("/bin/sh", sh, ipafile)
	var out bytes.Buffer
	cmd.Stdout = &out

	err := cmd.Run()
	if err != nil {
		fmt.Println("failed.")
	}

	data := make(map[string]interface{}, 20)
	json.Unmarshal(out.Bytes(), &data)

	return data
}

func makeplistfile(url, plistfile string, data map[string]interface{}) {
	model := Model{
		Ipa_url:   url,
		Bundle_id: data["CFBundleIdentifier"].(string),
		Version:   data["CFBundleVersion"].(string),
		AppName:   data["CFBundleName"].(string),
	}
	tem := template.New("plist")
	tem = template.Must(tem.Parse(plist))

	fd, _ := os.OpenFile(plistfile, os.O_RDWR|os.O_CREATE, 0644)
	tem.Execute(fd, model)
	fd.Close()
}

func handler(w http.ResponseWriter, r *http.Request) {
	url := r.URL.Path

	if url == "/favicon.ico" {
	} else {
		urlPath := baseDir + url
		f, err := os.Open(urlPath)
		if err != nil {
			reportError(w, err)
			return
		}

		defer f.Close()
		fi, err := f.Stat()
		if err != nil {
			reportError(w, err)
			return
		}

		switch mode := fi.Mode(); {
		case mode.IsDir():
			files, err := ioutil.ReadDir(urlPath)
			if err != nil {
				reportError(w, err)
				return
			}

			fileinfos := make([]FileInfo, 0, 10)
			if url != "/" {
				last, err := filepath.Abs(path.Join(url, ".."))
				if err != nil {
					reportError(w, err)
					return
				}

				fileinfos = append(fileinfos, FileInfo{
					FileName:   "..",
					Url:        last,
					CreateTime: "",
					Type:       Dir,
				})
			}

			headermap := make(map[string]*[]sampleInfo)
			for _, element := range files {
				if strings.HasPrefix(element.Name(), ".") {
					continue
				}
				fname := element.Name()
				if element.IsDir() {
					fileinfos = append(fileinfos, FileInfo{
						FileName:   fname,
						Url:        path.Join(url, fname),
						CreateTime: element.ModTime().Format("2006-01-02 15:04:05"),
						Type:       Dir,
					})
				} else {
					extension := filepath.Ext(fname)
					if extension == ipaName || extension == plistName {
						name := strings.Replace(fname, extension, "", 1)
						child := headermap[name]
						if child != nil {
							*child = append(*child, sampleInfo{Name: fname, CreateTime: element.ModTime()})
						} else {
							headermap[name] = &[]sampleInfo{sampleInfo{Name: fname, CreateTime: element.ModTime()}}
						}
						if extension == plistName {
							plisturl := scheme + "://" + filepath.Join(r.Host, url, fname)
							fileinfos = append(fileinfos, FileInfo{
								FileName:   fname,
								Url:        "itms-services://?action=download-manifest&url=" + plisturl,
								CreateTime: element.ModTime().Format("2006-01-02 15:04:05"),
								Type:       File,
							})
						} else {
							continue
						}
					} else {
						continue
					}

				}
			}

			for key, value := range headermap {
				if len(*value) < 2 {
					saminfo := (*value)[0]
					ext := filepath.Ext(saminfo.Name)
					if ext == ipaName {
						fpath := path.Join(baseDir, url, saminfo.Name)
						data := getAppInfo(fpath)
						plistfile := strings.Replace(fpath, ext, plistName, 1)

						ipaurl := scheme + "://" + filepath.Join(r.Host, url, saminfo.Name)
						makeplistfile(ipaurl, plistfile, data)
						fname := key + plistName

						fileinfos = append(fileinfos, FileInfo{
							FileName:   fname,
							Url:        "itms-services://?action=download-manifest&url=" + scheme + "://" + path.Join(r.Host, url, fname),
							CreateTime: saminfo.CreateTime.Format("2006-01-02 15:04:05"),
							Type:       File,
						})
					}

				}
			}

			if err != nil {
				reportError(w, err)
				return
			}

			jsonbyte, err := json.Marshal(fileinfos)
			data := make(map[string]interface{})
			data["fileinfos"] = string(jsonbyte)
			databuffer := parseTemplate(indexTem, data)
			w.Write(databuffer)
		case mode&os.ModeType == 0:
			http.ServeFile(w, r, urlPath)
		}
	}
}

func reportError(w http.ResponseWriter, err error) {
	fmt.Fprintf(w, "Error during operation: %s", err)
}

func parseTemplate(file string, data map[string]interface{}) []byte {
	var buf bytes.Buffer
	t := template.New(file)

	t, err := t.ParseFiles("templates/" + file)
	if err != nil {
		panic(err)
	}
	err = t.Execute(&buf, data)

	if err != nil {
		panic(err)
	}

	return buf.Bytes()
}

func init() {

}

func main() {
	port := "5555"
	for index, element := range os.Args {
		if index == 1 {
			scheme = element
		} else if index == 2 {
			port = element
		} else {
			baseDir = element
		}
	}
	http.Handle("/static/", http.FileServer(http.Dir(".")))
	http.HandleFunc("/", handler)
	http.ListenAndServe(":"+port, nil)
}
