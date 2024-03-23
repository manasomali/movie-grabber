package main

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"syscall"

	"github.com/manasomali/movie-grabber/internal/helpers"
	"github.com/manasomali/movie-grabber/internal/qbt"
	"golang.org/x/term"
)

const URL = "https://topdezfilmes.de"

func main() {
	run()
	input := bufio.NewScanner(os.Stdin)
	input.Scan()

}

func run() {
	fmt.Println("MOVIE GRABBER")
	fmt.Println()
	fmt.Println("Qbittorrent config:")
	fmt.Print("Host (default: http://localhost:8080): ")
	var QBITORRENT_HOST string = "http://localhost:8080"
	fmt.Scanln(&QBITORRENT_HOST)
	fmt.Print("User (default: admin): ")
	var QBITORRENT_USER string = "admin"
	fmt.Scanln(&QBITORRENT_USER)
	var QBITORRENT_PASS string = "adminadmin"
	fmt.Print("Password (default: adminadmin): ")
	password, err := term.ReadPassword(int(syscall.Stdin))
	if err == nil {
		QBITORRENT_PASS = string(password)
	}

	fmt.Println()
	fmt.Println()
	fmt.Print("How many pages do you want to search? ")
	var pagesDesired int = 1
	fmt.Scanln(&pagesDesired)

	pageNumber := 1
	allLinks := make([]string, 0)
	for pageNumber <= pagesDesired {
		links, err := helpers.GetAllMovieLinks(URL+"/page/"+strconv.Itoa(pageNumber), "a")
		if err != nil {
			fmt.Println("Error downloading file: ", err)
			return
		}
		allLinks = append(allLinks, links...)
		pageNumber++
	}

	movieLinks := helpers.FilterMovieLinks(allLinks)
	movieNames := make([]string, 0)
	for _, link := range movieLinks {
		movieName := filepath.Base(strings.TrimSuffix(link, ".zip"))
		movieNames = append(movieNames, movieName)
		fmt.Println(len(movieNames)-1, "-", strings.ReplaceAll(movieName, "-", " "))
	}

	var movieIndex int
	fmt.Print("Choose which movie do you want to download? ")
	fmt.Scanln(&movieIndex)
	movieChoosed := movieNames[movieIndex]
	err = helpers.DownloadFile("temp/"+movieChoosed+".zip", movieLinks[movieIndex])
	if err != nil {
		fmt.Println("Error downloading file: ", err)
		return
	}

	filesList, err := helpers.Unzip("temp/" + movieChoosed + ".zip")
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

	targetDir := helpers.ChooseDirectory()
	if len(srtFiles) == 0 {
		fmt.Println("No str file found")
	} else if len(srtFiles) == 1 {
		helpers.CopyFile(srtFiles[0], targetDir)
	} else {
		for n, srtFile := range srtFiles {
			fmt.Println(strconv.Itoa(n) + " - " + srtFile)
		}
		var srtDesired int
		fmt.Print("Choose the str file: ")
		fmt.Scanln(&srtDesired)
		helpers.CopyFile(srtFiles[srtDesired], targetDir)
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

	sid, err := qbt.LoginQbt(QBITORRENT_HOST, QBITORRENT_USER, QBITORRENT_PASS)
	if err != nil {
		fmt.Println("Error login qbitorrent " + err.Error())
	}
	err = qbt.AddTorrentFromFile(QBITORRENT_HOST, "temp/"+torrentToAdd, sid)
	if err != nil {
		fmt.Println("Error adding torrent file " + err.Error())
	}
	qbt.LogoutQbt(QBITORRENT_HOST, sid)
	if err != nil {
		fmt.Println("Fail to logout qbitorrent " + err.Error())
	}

	helpers.DeleteFolder("temp")

}
