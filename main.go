package main

import (
	"archive/zip"
	"bytes"
	"errors"
	"fmt"
	"io"
	"log"
	"mime/multipart"
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/ncruces/zenity"
	"golang.org/x/net/html"
)

const URL = "https://topdezfilmes.de"

func main() {
	fmt.Print("Qbitorrent host (default: http://localhost:8080): ")
	var QBITORRENT_HOST string = "http://localhost:8080"
	fmt.Scanln(&QBITORRENT_HOST)
	fmt.Print("Qbitorrent user (default: admin): ")
	var QBITORRENT_USER string = "admin"
	fmt.Scanln(&QBITORRENT_USER)
	fmt.Print("Qbitorrent password (default: admin): ")
	var QBITORRENT_PASS string = "admin"
	fmt.Scanln(&QBITORRENT_PASS)

	fmt.Print("How many pages do you want to search? ")
	var pagesDesired int = 1
	fmt.Scanln(&pagesDesired)

	pageNumber := 1
	allLinks := make([]string, 0)
	for pageNumber <= pagesDesired {
		links := getAllMovieLinks(URL+"/page/"+strconv.Itoa(pageNumber), "a")
		allLinks = append(allLinks, links...)
		pageNumber++
	}

	movieLinks := filterMovieLinks(allLinks)
	movieNames := make([]string, 0)
	for _, link := range movieLinks {
		movieName := filepath.Base(strings.TrimSuffix(link, ".zip"))
		movieNames = append(movieNames, movieName)
		fmt.Println(len(movieNames)-1, "-", strings.ReplaceAll(movieName, "-", " "))
	}

	var movieIndex int
	fmt.Print("Choose which movie you want to download? ")
	fmt.Scanln(&movieIndex)
	movieChoosed := movieNames[movieIndex]
	err := downloadFile("temp/"+movieChoosed+".zip", movieLinks[movieIndex])
	if err != nil {
		fmt.Println("Error downloading file: ", err)
		return
	}

	filesList, err := unzip("temp/" + movieChoosed + ".zip")
	if err != nil {
		fmt.Println("Error unzip file: ", err)
	}
	srtFiles := make([]string, 0)
	torrentFiles := make([]string, 0)
	for _, file := range filesList {
		if strings.HasSuffix(file, ".srt") {
			srtFiles = append(srtFiles, file)
		} else if strings.HasSuffix(file, ".torrent") {
			torrentFiles = append(torrentFiles, file)
		}
	}

	targetDir := chooseDirectory()
	if len(srtFiles) == 0 {
		fmt.Println("No str file found")
	} else if len(srtFiles) == 1 {
		copySrtFile(srtFiles[0], targetDir)
	} else {
		for n, srtFile := range srtFiles {
			fmt.Println(strconv.Itoa(n) + " - " + srtFile)
		}
		var srtDesired int
		fmt.Print("Choose the str file: ")
		fmt.Scanln(&srtDesired)
		copySrtFile(srtFiles[srtDesired], targetDir)
	}
	torrentToAdd := ""
	if len(torrentFiles) == 0 {
		fmt.Print("No torrent file found")
	} else if len(torrentFiles) == 1 {
		torrentToAdd = torrentFiles[0]
	} else {
		for n, torrentFile := range torrentFiles {
			fmt.Println(strconv.Itoa(n) + " - " + torrentFile)
		}
		var torrentDesired int
		fmt.Print("Choose the torrent file: ")
		fmt.Scanln(&torrentDesired)
		torrentToAdd = torrentFiles[torrentDesired]
	}

	sid, err := loginQbt(QBITORRENT_HOST, QBITORRENT_USER, QBITORRENT_PASS)
	if err != nil {
		fmt.Print("Error login qbitorrent " + err.Error())
	}
	err = addTorrentFromFile(QBITORRENT_HOST, "temp/"+torrentToAdd, sid)
	if err != nil {
		fmt.Print("Error adding torrent file " + err.Error())
	}
	logoutQbt(QBITORRENT_HOST, sid)
	if err != nil {
		fmt.Print("Fail to logout qbitorrent " + err.Error())
	}

	deleteFolder("temp")
}

func chooseDirectory() string {
	selectedPath, err := zenity.SelectFile(zenity.Filename(``), zenity.Directory())
	if err != nil {
		fmt.Println("Failed to choose directory: " + err.Error())
	}
	return selectedPath
}

func downloadFile(filepath string, url string) error {
	resp, err := http.Get(url)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	os.MkdirAll("temp", 0755)
	out, err := os.Create(filepath)
	if err != nil {
		return err
	}
	defer out.Close()

	_, err = io.Copy(out, resp.Body)
	return err
}

func unzip(src string) ([]string, error) {
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

func loginQbt(baseUrl string, username string, password string) (string, error) {
	posturl := baseUrl + "/api/v2/"
	data := "username=" + username + "&password=" + password
	client := &http.Client{}
	req, err := http.NewRequest("POST", posturl+"auth/login", strings.NewReader(data))
	if err != nil {
		return "", err
	}

	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("Referer", baseUrl)

	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	var SID string
	for _, cookie := range resp.Cookies() {
		if cookie.Name == "SID" {
			SID = cookie.Value
			break
		}
	}

	if SID != "" {
		return SID, nil
	} else {
		return "", errors.New("SID not found in response cookies")
	}
}

func getAllMovieLinks(url string, element string) []string {
	resp, err := http.Get(url)
	if err != nil {
		log.Fatalf("Error fetching URL: %v", err)
	}
	defer resp.Body.Close()

	doc, err := html.Parse(resp.Body)
	if err != nil {
		log.Fatalf("Error parsing HTML: %v", err)
	}
	links := make([]string, 0)
	var findLinks func(*html.Node)
	findLinks = func(n *html.Node) {
		if n.Type == html.ElementNode && n.Data == "a" {
			for _, attr := range n.Attr {
				if attr.Key == "href" {
					links = append(links, attr.Val)
				}
			}
		}
		for c := n.FirstChild; c != nil; c = c.NextSibling {
			findLinks(c)
		}
	}
	findLinks(doc)
	return links

}

func filterMovieLinks(links []string) []string {
	filteredLinks := []string{}
	for _, link := range links {
		if len(link) >= 4 && link[len(link)-4:] == ".zip" {
			filteredLinks = append(filteredLinks, link)
		}
	}
	return filteredLinks
}

func addTorrentFromFile(baseURL string, filePath string, sid string) error {
	url := baseURL + "/api/v2/torrents/add"
	file, err := os.Open(filePath)
	defer file.Close()
	if err != nil {
		return err
	}
	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)

	part, err := writer.CreateFormFile("torrents", file.Name())
	if err != nil {
		return err
	}
	_, err = io.Copy(part, file)
	if err != nil {
		return err
	}
	err = writer.Close()
	if err != nil {
		return err
	}

	req, err := http.NewRequest("POST", url, body)
	if err != nil {
		return err
	}

	req.Header.Set("Content-Type", writer.FormDataContentType())
	req.AddCookie(&http.Cookie{Name: "SID", Value: sid})
	client := &http.Client{}
	resp, err := client.Do(req)
	defer resp.Body.Close()
	if err != nil {
		return err
	}
	if resp.StatusCode != http.StatusOK {
		return nil
	}
	return errors.New("Fail")
}

func logoutQbt(baseUrl string, sid string) error {
	posturl := baseUrl + "/api/v2/"

	cookie := &http.Cookie{
		Name:  "SID",
		Value: sid,
	}

	jar, _ := cookiejar.New(nil)
	jar.SetCookies(&url.URL{Scheme: "http", Host: strings.TrimPrefix(baseUrl, "http://")}, []*http.Cookie{cookie})
	client := &http.Client{
		Jar: jar,
	}

	req, err := http.NewRequest("POST", posturl+"auth/logout", nil)
	if err != nil {
		return err
	}

	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	return nil
}
func copySrtFile(srtFile string, destinyDir string) {
	inputFile, err := os.Open("temp/" + srtFile)
	if err != nil {
		fmt.Println("Couldn't open source file", err)
	}
	defer inputFile.Close()

	outputFile, err := os.Create(destinyDir + "/" + srtFile)
	if err != nil {
		fmt.Println("Couldn't open dest file", err)
	}
	defer outputFile.Close()

	_, err = io.Copy(outputFile, inputFile)
	if err != nil {
		fmt.Println("Couldn't copy to dest from source", err)
	}
}

func deleteFolder(folderName string) {
	err := os.RemoveAll(folderName)
	if err != nil {
		fmt.Println("Failed to delete folder", err)
	}
}
