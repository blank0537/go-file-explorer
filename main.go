package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"html/template"
	"math"
	"net/http"
	"os"
	"slices"
	"strings"
)

type Explorer struct {
	Path   string
	Prev   string
	IsRoot bool
	Files  []File
}

type File struct {
	Name    string
	Path    string
	Size    string
	Mode    os.FileMode
	ModTime string
	IsDir   bool
}

var ignoreFiles = []string{".dmg", ".iso"}

func renderFile(path *string, w http.ResponseWriter) error {
	for _, val := range ignoreFiles {
		if strings.Contains(*path, val) {
			return errors.New("unsupported file to open in browser")
		}
	}
	file, err := os.ReadFile(*path)
	if err != nil {
		return err
	}

	if strings.Contains(*path, ".pdf") {
		splitStr := strings.Split(*path, "/")
		fileName := splitStr[len(splitStr)-1]
		w.Header().Set("Content-Type", "application/pdf")
		w.Header().Set("Content-Disposition", fmt.Sprintf("inline;filename=%s", fileName))
		w.Header().Set("Content-Transfer-Encoding", "binary")
		w.Header().Set("Accept-Ranges", "bytes")
	} else {
		w.Header().Set("Content-Type", "text/plain")
	}

	w.Write(file)
	return nil
}

var explorer Explorer

func main() {
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		path := r.URL.Path[0:]
		if strings.Contains(path, "favicon.ico") {
			return
		}

		files, err := os.ReadDir(path)
		if err != nil {
			fileErr := renderFile(&path, w)
			if fileErr != nil {
				http.Error(w, fileErr.Error(), http.StatusInternalServerError)
				return
			}
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		if path != "/" {
			explorer.IsRoot = true
			splitPath := strings.Split(path, "/")
			splitPath = splitPath[:len(splitPath)-1]
			explorer.Prev = strings.Join(splitPath, "/")
			if explorer.Prev == "/" {
				explorer.IsRoot = false
			}
			if explorer.Prev == "" {
				explorer.Prev = "/"
			}
		} else {
			explorer.IsRoot = false
		}

		var fileList []File
		for _, file := range files {
			fileInfo, err := file.Info()
			if err != nil {
				continue
			}

			var filePath string
			if path == "/" {
				filePath = path + file.Name()
			} else {
				filePath = path + "/" + file.Name()
			}
			sizeType := "KB"
			size := float64(fileInfo.Size()) / (1000)
			if size > 999 {
				sizeType = "MB"
				size = size / 1000
			}
			if size > 999 {
				sizeType = "GB"
				size = size / 1000
			}
			size = math.Ceil(size*100) / 100

			fileList = append(fileList, File{
				Name:    file.Name(),
				Path:    filePath,
				Size:    fmt.Sprintf("%.1f %s", size, sizeType),
				Mode:    fileInfo.Mode(),
				ModTime: fileInfo.ModTime().Format("02 Jan 2006 15:04"),
				IsDir:   file.IsDir(),
			})
		}

		explorer.Files = fileList
		explorer.Path = path

		tmpl := template.Must(template.ParseFiles("./index.html"))
		if err := tmpl.Execute(w, explorer); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
	})

	http.HandleFunc("/create/folder", func(w http.ResponseWriter, r *http.Request) {
		var data map[string]interface{}
		if err := json.NewDecoder(r.Body).Decode(&data); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
		}
		if err := createFolder(data["path"].(string), data["name"].(string)); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		json.NewEncoder(w).Encode(map[string]interface{}{"success": true})
	})

	http.HandleFunc("/create/file", func(w http.ResponseWriter, r *http.Request) {
		var data map[string]interface{}
		if err := json.NewDecoder(r.Body).Decode(&data); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
		}
		if err := createFile(data["path"].(string), data["name"].(string)); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		json.NewEncoder(w).Encode(map[string]interface{}{"success": true})
	})

	http.HandleFunc("/rename", func(w http.ResponseWriter, r *http.Request) {
		var data map[string]interface{}
		if err := json.NewDecoder(r.Body).Decode(&data); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
		}

		if isExist := slices.ContainsFunc(explorer.Files, func(val File) bool {
			return val.Name == data["newName"].(string)
		}); isExist {
			http.Error(w, "name already exist", http.StatusBadRequest)
		}

		if err := renameFileOrDir(data["path"].(string), data["name"].(string), data["newName"].(string)); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		json.NewEncoder(w).Encode(map[string]interface{}{"success": true})
	})

	http.HandleFunc("/delete", func(w http.ResponseWriter, r *http.Request) {
		var data map[string]interface{}
		if err := json.NewDecoder(r.Body).Decode(&data); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
		}
		if err := deleteFileOrDir(data["path"].(string), data["name"].(string)); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		json.NewEncoder(w).Encode(map[string]interface{}{"success": true})
	})

	fmt.Println("Starting server on :8800")
	http.ListenAndServe(":8800", nil)
}
