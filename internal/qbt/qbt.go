package qbt

import (
	"bytes"
	"errors"
	"io"
	"mime/multipart"
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"os"
	"strconv"
	"strings"
)

func LoginQbt(baseUrl string, username string, password string) (string, error) {
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

func AddTorrentFromFile(baseURL string, filePath string, sid string) error {
	url := baseURL + "/api/v2/torrents/add"
	file, err := os.Open(filePath)
	if err != nil {
		return err
	}
	defer file.Close()
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
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode == http.StatusOK {
		return errors.New("Fail, status code: " + strconv.Itoa(resp.StatusCode))
	}
	return nil
}

func LogoutQbt(baseUrl string, sid string) error {
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
