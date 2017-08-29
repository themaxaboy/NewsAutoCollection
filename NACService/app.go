// Copyright 2012 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// +build !plan9

package main

import (
	"log"
	"os"
	"path"
	"path/filepath"
	"strings"

	"github.com/creamdog/gonfig"
	"github.com/fsnotify/fsnotify"
)

var Collectionpath, Workingpath string

func loadSetting() {
	dir, err := filepath.Abs(filepath.Dir(os.Args[0]))
	if err != nil {
		log.Fatal("Error : ", err)
	}

	f, err := os.Open(path.Join(dir, "settings.json"))
	if err != nil {
		log.Fatal("Error : ", err)
	}
	defer f.Close()
	config, err := gonfig.FromJson(f)
	if err != nil {
		log.Fatal("Error : ", err)
	}

	APPName, err := config.GetString("APPName", nil)
	if err != nil {
		log.Fatal("Error : ", err)
	}
	Collectionpath, err = config.GetString("Path/Collectionpath", nil)
	if err != nil {
		log.Fatal("errErroror : ", err)
	}
	Workingpath, err = config.GetString("Path/Workingpath", nil)
	if err != nil {
		log.Fatal("Error : ", err)
	}

	log.Println(APPName, "is working.")
	log.Println("Collectionpath :", Collectionpath)
	log.Println("Workingpath :", Workingpath)
}

func app() {
	loadSetting()
	moveOldfile()

	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		log.Fatal(err)
	}
	defer watcher.Close()

	done := make(chan bool)
	go func() {
		for {
			select {
			case event := <-watcher.Events:
				if event.Op&fsnotify.Create == fsnotify.Create {
					fileName := getFilename(event.Name)
					if !validPattern(fileName) {
						continue
					}

					moveFile(event.Name)
				}
			case err := <-watcher.Errors:
				log.Println("Error : ", err)
			}
		}
	}()

	err = watcher.Add(Workingpath)
	if err != nil {
		log.Fatal("Error : ", err)
	}
	<-done
}

func getFilename(fullName string) string {
	Replacer := strings.NewReplacer(Workingpath, "", "\\", "")
	return Replacer.Replace(fullName)
}

func getDate(fileName string) string {
	return convertDate(fileName[7:15])
}

func convertDate(fileDate string) string {
	return fileDate[4:] + fileDate[2:4] + fileDate[:2]
}

func exists(filePath string) (exists bool) {
	exists = true
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		exists = false
	}
	return
}

func validPattern(fileName string) bool {
	if len(fileName) != 30 {
		log.Println("Incorrect file pattern : ", fileName)
		return false
	}
	return true
}

func moveFile(evenFile string) {
	fileName := getFilename(evenFile)
	log.Println("File found : ", fileName)

	fileDate := getDate(fileName)
	outputPath := path.Join(Collectionpath, fileDate)
	if !exists(outputPath) {
		err := os.MkdirAll(outputPath, os.ModePerm)
		log.Println("Make dir : ", outputPath)
		if err != nil {
			log.Fatal("Error : ", err)
		}
	}
	err := os.Rename(evenFile, path.Join(outputPath, fileName))
	log.Println("Move file : ", fileName, " >>> ", fileDate)
	if err != nil {
		log.Fatal("Error : ", err)
	}
}

func moveOldfile() {
	fileList := []string{}
	err := filepath.Walk(Workingpath, func(path string, f os.FileInfo, err error) error {
		fileList = append(fileList, path)
		return nil
	})
	if err != nil {
		log.Fatal("Error : ", err)
	}

	for _, file := range fileList {
		log.Println("Old file : ", file)
		fileName := getFilename(file)
		if !validPattern(fileName) {
			continue
		}
		moveFile(file)
	}
}
