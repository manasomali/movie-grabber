package helpers

import (
	"archive/zip"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/ncruces/zenity"
)

func ChooseDirectory() string {
	selectedPath, err := zenity.SelectFile(zenity.Filename(``), zenity.Directory())
	if err != nil {
		fmt.Println("Failed to choose directory: " + err.Error())
	}
	return selectedPath
}

func Unzip(src string) ([]string, error) {
	filesList := make([]string, 0)
	r, err := zip.OpenReader(src)
	if err != nil {
		return filesList, err
	}
	defer func() {
		if err := r.Close(); err != nil {
			fmt.Println(err)
		}
	}()
	os.MkdirAll("temp", 0755)
	extractAndWriteFile := func(f *zip.File) error {
		rc, err := f.Open()
		if err != nil {
			return err
		}
		defer func() {
			if err := rc.Close(); err != nil {
				fmt.Println(err)
			}
		}()

		path := filepath.Join("temp", f.Name)

		if !strings.HasPrefix(path, filepath.Clean("temp")+string(os.PathSeparator)) {
			return fmt.Errorf("illegal file path: %s", path)
		}

		if f.FileInfo().IsDir() {
			os.MkdirAll(path, f.Mode())
		} else {
			os.MkdirAll(filepath.Dir(path), f.Mode())
			f, err := os.OpenFile(path, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, f.Mode())
			if err != nil {
				return err
			}
			defer func() {
				if err := f.Close(); err != nil {
					fmt.Println(err)
				}
			}()

			_, err = io.Copy(f, rc)
			if err != nil {
				return err
			}
		}
		return nil
	}

	for _, f := range r.File {
		filesList = append(filesList, f.Name)
		err := extractAndWriteFile(f)
		if err != nil {
			return filesList, err
		}
	}
	return filesList, nil
}
func CopyFile(file string, destinyDir string) {
	inputFile, err := os.Open("temp/" + file)
	if err != nil {
		fmt.Println("Couldn't open source file", err)
	}
	defer inputFile.Close()

	outputFile, err := os.Create(destinyDir + "/" + file)
	if err != nil {
		fmt.Println("Couldn't open dest file", err)
	}
	defer outputFile.Close()

	_, err = io.Copy(outputFile, inputFile)
	if err != nil {
		fmt.Println("Couldn't copy to dest from source", err)
	}
}

func DeleteFolder(folderName string) {
	err := os.RemoveAll(folderName)
	if err != nil {
		fmt.Println("Failed to delete folder", err)
	}
}
