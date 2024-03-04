package util

import (
	"bytes"
	"context"
	"crypto/tls"
	"encoding/json"
	"io"
	"net"
	"net/http"
	"os"
	"time"

	"github.com/jedib0t/go-pretty/v6/table"
)

func ReadTxtFile(url string) ([]byte, error) {
	file, err := os.Open(url)
	if err != nil {
		return []byte(""), err
	}
	defer file.Close()
	stringData, err := io.ReadAll(file)
	if err != nil {
		return []byte(""), err
	}
	return stringData, nil
}

// write file
func WriteFileWithNosec(pathName string, data []byte) error {
	// #nosec G306, Expect WriteFile permissions to be 0600 or less
	return os.WriteFile(pathName, data, 0644)
}

func CreateDirIfNotExists(dirLocation string) error {
	if _, err := os.Stat(dirLocation); os.IsNotExist(err) {
		return os.MkdirAll(dirLocation, os.ModeDir|0755)
	}
	return nil
}

func PathExist(name string) bool {
	_, err := os.Stat(name)
	return err == nil
}

func DeleteDirs(filePath string) error {
	return os.RemoveAll(filePath)
}

// list all dirs in a give dir
func ListDirs(dirPath string) ([]string, error) {
	var dirs []string
	files, err := os.ReadDir(dirPath)
	if err != nil {
		return dirs, err
	}
	for _, file := range files {
		if file.IsDir() {
			dirs = append(dirs, file.Name())
		}
	}
	return dirs, nil
}

func DeleteFile(filePath string) error {
	return os.Remove(filePath)
}

func CommonRequest(requestUrl, httpMethod, nameServer string, postBody json.RawMessage, header map[string]string, skipTlsCheck, disableKeepAlive bool, timeout time.Duration) ([]byte, int, error) {
	var req *http.Request
	var reqErr error

	req, reqErr = http.NewRequest(httpMethod, requestUrl, bytes.NewReader(postBody))
	if reqErr != nil {
		return []byte{}, http.StatusInternalServerError, reqErr
	}
	req.Header.Set("Content-Type", "application/json; charset=UTF-8")
	for key, val := range header {
		req.Header.Set(key, val)
	}
	client := &http.Client{
		Timeout: timeout,
	}
	tr := &http.Transport{
		DisableKeepAlives: disableKeepAlive,
	}
	if skipTlsCheck {
		tr.TLSClientConfig = &tls.Config{InsecureSkipVerify: true}
	}

	if nameServer != "" {
		r := &net.Resolver{
			PreferGo: true,
			Dial: func(ctx context.Context, network, address string) (net.Conn, error) {
				d := net.Dialer{Timeout: 1 * time.Second}
				return d.DialContext(ctx, "udp", nameServer)
			},
		}
		tr.DialContext = (&net.Dialer{
			Resolver: r,
		}).DialContext
	}
	client.Transport = tr
	resp, respErr := client.Do(req)
	if respErr != nil {
		return []byte{}, http.StatusInternalServerError, respErr
	}
	defer resp.Body.Close()
	body, readBodyErr := io.ReadAll(resp.Body)
	if readBodyErr != nil {
		return []byte{}, http.StatusInternalServerError, readBodyErr
	}
	return body, resp.StatusCode, nil
}

func PrintTable(header table.Row, rows []table.Row) {
	t := table.NewWriter()
	t.SetOutputMirror(os.Stdout)
	t.AppendHeader(header)
	for _, row := range rows {
		t.AppendRow(row)
		t.AppendSeparator()
	}
	t.SetStyle(table.StyleColoredBlackOnBlueWhite)
	t.SetAutoIndex(false)
	t.Render()
}
