package client

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/0x0FACED/rapid/internal/lan/mdnss"
	"github.com/0x0FACED/rapid/internal/model"
)

// LANClient позволяет находить серверы и загружать файлы
type LANClient struct {
	httpClient *http.Client

	mu    sync.Mutex
	mdnss *mdnss.MDNSScanner
}

// New создает новый клиент
func New(mdnss *mdnss.MDNSScanner) *LANClient {
	return &LANClient{
		httpClient: &http.Client{Timeout: 5 * time.Second},
		mdnss:      mdnss,
	}
}

// DiscoverServers ищет сервера в сети
func (c *LANClient) DiscoverServers(ctx context.Context, ch chan model.ServiceInstance) {
	c.mdnss.DiscoverPeers(ctx, ch)
}

// GetFiles получает список файлов с первого найденного сервера
func (c *LANClient) GetFiles(addr string) ([]model.File, error) {
	c.mu.Lock()
	defer c.mu.Unlock()

	fmt.Println("GET FILES ADDR: ", addr)
	url := fmt.Sprintf("http://%s:8070/api/files", addr)

	resp, err := c.httpClient.Get(url)
	if err != nil {
		log.Println("Err getting files with ip:", addr, err)
	}
	defer resp.Body.Close()

	var files []model.File
	if err := json.NewDecoder(resp.Body).Decode(&files); err != nil {
		return nil, err
	}

	return files, nil

}

// DownloadFile скачивает файл по ID
func (c *LANClient) DownloadFile(addr, fileID, filename, savePath string) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	url := fmt.Sprintf("http://%s:8070/api/download/%s", addr, fileID)
	resp, err := c.httpClient.Get(url)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	out, err := os.Create(filepath.Join(savePath, filename))
	if err != nil {
		return err
	}
	defer out.Close()

	_, err = io.Copy(out, resp.Body)
	if err != nil {
		return err
	}

	return nil
}
