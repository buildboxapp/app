package lib

import (
	"archive/zip"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
)

// Создаем файл по указанному пути если его нет
func (c *Lib) CreateFile(path string) (err error) {

	// detect if file exists
	_, err = os.Stat(path)
	var file *os.File

	// delete old file if exists
	if !os.IsNotExist(err) {
		os.RemoveAll(path)
	}

	// create file
	file, err = os.Create(path)
	if isError(err) {
		c.Logger.Error(err, "Error creating directory")
		return err
	}
	defer file.Close()

	return err
}

// функция печати в лог ошибок (вспомогательная)
func (c *Lib) isError(err error) bool {
	if err != nil {
		fmt.Println(err.Error())
	}
	return (err != nil)
}

// пишем в файл по указанному пути
func (c *Lib) WriteFile(path string, data []byte) error {

	// detect if file exists and create
	c.CreateFile(path)

	// open file using READ & WRITE permission
	var file, err = os.OpenFile(path, os.O_RDWR, 0644)

	if isError(err) {
		return err
	}
	defer file.Close()

	// write into file
	_, err = file.Write(data)
	if isError(err) {
		return err
	}

	// save changes
	err = file.Sync()
	if isError(err) {
		return err
	}

	return nil
}

// читаем файл. (отключил: всегда в рамках рабочей диретории)
func (c *Lib) ReadFile(path string) (result string, err error) {
	// если не от корня, то подставляем текущую директорию
	//if path[:1] != "/" {
	//	path = CurrentDir() + "/" + path
	//} else {
	//	path = CurrentDir() + path
	//}

	file, err := os.Open(path)
	if err != nil {
		return "", err
	}

	b, err := ioutil.ReadAll(file)
	if err == nil {
		result = string(b)
	}
	defer file.Close()

	return result, err
}

// копирование папки
func (c *Lib) CopyFolder(source string, dest string) (err error) {

	sourceinfo, err := os.Stat(source)
	if err != nil {
		return err
	}

	err = os.MkdirAll(dest, sourceinfo.Mode())
	if err != nil {
		return err
	}

	directory, _ := os.Open(source)

	objects, err := directory.Readdir(-1)

	for _, obj := range objects {

		sourcefilepointer := source + "/" + obj.Name()

		destinationfilepointer := dest + "/" + obj.Name()

		if obj.IsDir() {
			err = c.CopyFolder(sourcefilepointer, destinationfilepointer)
			if err != nil {
				fmt.Println(err)
			}
		} else {
			err = c.CopyFile(sourcefilepointer, destinationfilepointer)
			if err != nil {
				fmt.Println(err)
			}
		}

	}
	return
}

// копирование файла
func (c *Lib) CopyFile(source string, dest string) (err error) {
	sourcefile, err := os.Open(source)
	if err != nil {
		return err
	}

	defer sourcefile.Close()

	destfile, err := os.Create(dest)
	if err != nil {
		return err
	}

	defer destfile.Close()

	_, err = io.Copy(destfile, sourcefile)
	if err == nil {
		sourceinfo, err := os.Stat(source)
		if err != nil {
			err = os.Chmod(dest, sourceinfo.Mode())
		}
	}

	return
}

// создание папки
func (c *Lib) CreateDir(path string, mode os.FileMode) (err error) {
	if mode == 0 {
		mode = 0711
	}
	err = os.MkdirAll(path, mode)
	if err != nil {
		c.Logger.Error(err, "Error creating directory")
		return err
	}

	return nil
}

func (c *Lib) DeleteFile(path string) (err error) {
	err = os.Remove(path)
	if err != nil {
		c.Logger.Error(err, "Error deleted file: ", path)
		return
	}

	return nil
}

func (c *Lib) MoveFile(source string, dest string) (err error) {
	err = c.CopyFile(source, dest)
	if err != nil {
		c.Logger.Error(err, "Error file transfer (error copies)")
		return
	}
	err = c.DeleteFile(source)
	if err != nil {
		c.Logger.Error(err, "Error file transfer (error deleted)")
		return
	}

	return nil
}

// zip("/tmp/documents", "/tmp/backup.zip")
func (c *Lib) Zip(source, target string) error {
	zipfile, err := os.Create(target)
	if err != nil {
		return err
	}
	defer zipfile.Close()

	archive := zip.NewWriter(zipfile)
	defer archive.Close()

	info, err := os.Stat(source)
	if err != nil {
		return nil
	}

	var baseDir string
	if info.IsDir() {
		baseDir = filepath.Base(source)
	}

	filepath.Walk(source, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		header, err := zip.FileInfoHeader(info)
		if err != nil {
			return err
		}

		if baseDir != "" {
			header.Name = filepath.Join(baseDir, strings.TrimPrefix(path, source))
		}

		if info.IsDir() {
			header.Name += "/"
		} else {
			header.Method = zip.Deflate
		}

		writer, err := archive.CreateHeader(header)
		if err != nil {
			return err
		}

		if info.IsDir() {
			return nil
		}

		file, err := os.Open(path)
		if err != nil {
			return err
		}
		defer file.Close()
		_, err = io.Copy(writer, file)
		return err
	})

	return err
}

// unzip("/tmp/report-2015.zip", "/tmp/reports/")
func (c *Lib) Unzip(archive, target string) error {
	reader, err := zip.OpenReader(archive)
	if err != nil {
		return err
	}

	if err := os.MkdirAll(target, 0755); err != nil {
		return err
	}

	for _, file := range reader.File {
		path := filepath.Join(target, file.Name)
		if file.FileInfo().IsDir() {
			os.MkdirAll(path, file.Mode())
			continue
		}

		fileReader, err := file.Open()
		if err != nil {
			return err
		}
		defer fileReader.Close()

		targetFile, err := os.OpenFile(path, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, file.Mode())
		if err != nil {
			return err
		}
		defer targetFile.Close()

		if _, err := io.Copy(targetFile, fileReader); err != nil {
			return err
		}
	}

	return nil
}
