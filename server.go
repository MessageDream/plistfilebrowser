package main

import (
	"bytes"
	"encoding/json"
	"errors"
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
	apkname   = ".apk"
	//crtname   = ".crt"
	plist = `<?xml version="1.0" encoding="UTF-8"?>
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

func getAppInfo(ipafile string) (map[string]interface{}, error) {
	sh := "." + "/genplist.sh"
	cmd := exec.Command("/bin/sh", sh, ipafile)
	var out bytes.Buffer
	cmd.Stdout = &out

	err := cmd.Run()
	if err != nil {
		return nil, err
	}

	data := make(map[string]interface{}, 20)
	json.Unmarshal(out.Bytes(), &data)

	return data, nil
}

func makeplistfile(url, plistfile string, data map[string]interface{}) error {
	if data == nil {
		return errors.New("the map of data is nil")
	}
	bundleId := data["CFBundleIdentifier"]
	version := data["CFBundleVersion"]
	appName := data["CFBundleName"]

	if &bundleId == nil || &version == nil || &appName == nil {
		return errors.New("the map of data is nil")
	}
	model := Model{
		Ipa_url:   url,
		Bundle_id: bundleId.(string),
		Version:   version.(string),
		AppName:   appName.(string),
	}
	tem := template.New("plist")
	tem = template.Must(tem.Parse(plist))

	fd, err := os.OpenFile(plistfile, os.O_RDWR|os.O_CREATE, 0644)
	if err != nil {
		return err
	}
	err = tem.Execute(fd, model)
	if err != nil {
		return err
	}
	err = fd.Close()
	if err != nil {
		return err
	}
	return nil
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
					name := strings.Replace(fname, extension, "", 1)
					if extension == ipaName || extension == plistName {
						child := headermap[name]
						if child != nil {
							*child = append(*child, sampleInfo{Name: fname, CreateTime: element.ModTime()})
						} else {
							headermap[name] = &[]sampleInfo{sampleInfo{Name: fname, CreateTime: element.ModTime()}}
						}
						if extension == plistName {
							plisturl := scheme + "://" + filepath.Join(r.Host, url, fname)
							fileinfos = append(fileinfos, FileInfo{
								FileName:   name,
								Url:        "itms-services://?action=download-manifest&url=" + plisturl,
								CreateTime: element.ModTime().Format("2006-01-02 15:04:05"),
								Type:       File,
							})
						} else {
							continue
						}
					} else if extension == apkname {
						fileinfos = append(fileinfos, FileInfo{
							FileName:   name,
							Url:        path.Join(url, fname),
							CreateTime: element.ModTime().Format("2006-01-02 15:04:05"),
							Type:       File,
						})
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
						data, err := getAppInfo(fpath)
						if err != nil {
							reportError(w, err)
							return
						}
						plistfile := strings.Replace(fpath, ext, plistName, 1)

						ipaurl := scheme + "://" + filepath.Join(r.Host, url, saminfo.Name)
						err = makeplistfile(ipaurl, plistfile, data)

						if err != nil {
							reportError(w, err)
							return
						}
						fname := key + plistName

						fileinfos = append(fileinfos, FileInfo{
							FileName:   strings.Replace(fname, ext, "", 1),
							Url:        "itms-services://?action=download-manifest&url=" + scheme + "://" + path.Join(r.Host, url, fname),
							CreateTime: saminfo.CreateTime.Format("2006-01-02 15:04:05"),
							Type:       File,
						})
					}

				}
			}

			jsonbyte, err := json.Marshal(fileinfos)
			if err != nil {
				reportError(w, err)
				return
			}
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
			scheme = strings.ToLower(element)
		} else if index == 2 {
			port = element
		} else {
			baseDir = element
		}
	}
	http.Handle("/static/", http.FileServer(http.Dir(".")))
	http.HandleFunc("/", handler)
	if scheme == "https" {
		http.ListenAndServeTLS(":"+port, "server.crt", "server.key", nil)
	} else {
		http.ListenAndServe(":"+port, nil)
	}
}
