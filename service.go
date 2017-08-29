// Copyright 2015 Daniel Theophanes.
// Use of this source code is governed by a zlib-style
// license that can be found in the LICENSE file.

// simple does nothing except block while running the service.
package main

import (
	"log"
	"os"
	"path"
	"strings"

	"github.com/creamdog/gonfig"
	"github.com/fsnotify/fsnotify"
	"github.com/kardianos/service"
)

var logger service.Logger
var Collectionpath, Workingpath string

type program struct{}

func (p *program) Start(s service.Service) error {
	// Start should not block. Do the actual work async.
	go p.run()
	return nil
}
func (p *program) run() {
	// Do work here.
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
				//log.Println("event:", event)
				if event.Op&fsnotify.Create == fsnotify.Create {
					fileName := getFilename(event.Name)
					log.Println("File found : ", fileName)
					if len(fileName) != 30 {
						log.Println("Incorrect file pattern : ", fileName)
						continue
					}
					fileDate := getDate(fileName)
					outputPath := path.Join(Collectionpath, fileDate)
					if !exists(outputPath) {
						err := os.MkdirAll(outputPath, os.ModePerm)
						log.Println("Make dir : ", outputPath)
						if err != nil {
							log.Fatal("Error : ", err)
						}
					}
					err := os.Rename(event.Name, path.Join(outputPath, fileName))
					log.Println("Move file : ", fileName, " >>> ", fileDate)
					if err != nil {
						log.Fatal("Error : ", err)
					}
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
func (p *program) Stop(s service.Service) error {
	// Stop should not block. Return with a few seconds.
	return nil
}

func init() {
	f, err := os.Open("settings.json")
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

func main() {
	svcConfig := &service.Config{
		Name:        "GoServiceExampleSimple",
		DisplayName: "Go Service Example",
		Description: "This is an example Go service.",
	}

	prg := &program{}
	s, err := service.New(prg, svcConfig)
	if err != nil {
		log.Fatal(err)
	}
	logger, err = s.Logger(nil)
	if err != nil {
		log.Fatal(err)
	}
	err = s.Run()
	if err != nil {
		logger.Error(err)
	}
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