package main

import (
	"encoding/base64"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"

	"github.com/sirupsen/logrus"
)

const (
	assetDir = "./webui_assets"
)

func main() {
	var (
		log *logrus.Entry
	)

	// Initialize logging
	log = logrus.WithFields(logrus.Fields{
		"module": "main",
	})

	fs, err := os.Stat(assetDir)
	if err != nil {
		log.Fatalf("%s does not exist", assetDir)
	}
	if !fs.IsDir() {
		log.Fatalf("%s: not a directory", assetDir)
	}

	fmt.Printf(`package webui

/*
 * WARNING: this file is auto generated, do not modify
 */

type assetData struct {
  Data        string
  ContentType string
}

var webuiAssets = map[string]assetData{
`)

	err = filepath.Walk(assetDir, func(path string, info os.FileInfo, err error) error {
		if info.IsDir() {
			return nil
		}

		name := strings.ReplaceAll(path, "/", "_")
		name = strings.ReplaceAll(name, ".", "_")
		name = strings.ReplaceAll(name, "webui_assets_", "")
		contents, err := ioutil.ReadFile(path)
		if err != nil {
			log.Fatalf("ioutil.ReadFile: %v", err)
		}
		encoded := base64.StdEncoding.EncodeToString(contents)

		contentType := "text/plain"
		if strings.HasSuffix(path, ".html") {
			contentType = "text/html"
		} else if strings.HasSuffix(path, "css") {
			contentType = "text/css"
		} else if strings.HasSuffix(path, ".js") {
			contentType = "text/javascript"
		} else if strings.HasSuffix(path, ".ico") {
			contentType = "image/x-icon"
		} else if strings.HasSuffix(path, ".png") {
			contentType = "image/png"
		}

		fmt.Printf("  \"%s\": assetData{\n    Data: \"%s\",\n    ContentType: \"%s\",\n  },\n", name, encoded, contentType)
		return nil
	})

	fmt.Printf("}\n")
}
