package util

import (
	"bytes"
	"context"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/emicklei/go-restful"
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

func DeleteDirs(filePath string) error {
	return os.RemoveAll(filePath)
}

func DeleteFile(filePath string) error {
	return os.Remove(filePath)
}

func CommonRequest(requestUrl, httpMethod, nameServer string, postBody json.RawMessage, header map[string]string, skipTlsCheck, disableKeepAlive bool, timeout time.Duration) ([]byte, http.Header, int, error) {
	var req *http.Request
	var reqErr error

	req, reqErr = http.NewRequest(httpMethod, requestUrl, bytes.NewReader(postBody))
	if reqErr != nil {
		return []byte{}, nil, http.StatusInternalServerError, reqErr
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
		return []byte{}, nil, http.StatusInternalServerError, respErr
	}
	defer resp.Body.Close()
	body, readBodyErr := io.ReadAll(resp.Body)
	if readBodyErr != nil {
		return []byte{}, nil, http.StatusInternalServerError, readBodyErr
	}
	return body, resp.Header, resp.StatusCode, nil
}

func StartMockServer(port int, handlerLoader func(*restful.WebService), stopChan chan struct{}) error {

	svcContainer := restful.NewContainer()
	ws2 := new(restful.WebService)
	if handlerLoader != nil {
		handlerLoader(ws2)
	}

	svcContainer.Add(ws2)

	httpServer := http.Server{
		Handler: svcContainer,
	}

	httpServer.Addr = ":" + strconv.Itoa(port)
	err := httpServer.ListenAndServe()
	if err != nil {
		return err
	}
	<-stopChan
	httpServer.Close()
	return nil
}

func GetStringValueOrDefault(targetDescription, targetValue, defaultValue string) string {
	slog.Debug(fmt.Sprintf("'%s' is not set fall backup to default value: %s", targetDescription, defaultValue))
	if targetValue != "" {
		return targetValue
	}
	return defaultValue
}
