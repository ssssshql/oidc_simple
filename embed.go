package main

import (
	"embed"
	"io/fs"
)

//go:embed templates/* assets/*
var embeddedFiles embed.FS

// GetTemplatesFS 返回嵌入的模板文件系统
func GetTemplatesFS() (fs.FS, error) {
	return fs.Sub(embeddedFiles, "templates")
}

// GetAssetsFS 返回嵌入的资源文件系统
func GetAssetsFS() (fs.FS, error) {
	return fs.Sub(embeddedFiles, "assets")
}

// GetFavicon 返回 favicon.ico 内容
func GetFavicon() ([]byte, error) {
	return embeddedFiles.ReadFile("assets/favicon.ico")
}

// GetTemplate 返回指定模板内容
func GetTemplate(name string) ([]byte, error) {
	return embeddedFiles.ReadFile("templates/" + name)
}
