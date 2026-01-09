package downloader

import (
	"context"
	"fmt"
	"sync"

	"github.com/matejeliash/medownloader/internal/dto"
)

type DownloadManager struct {
	Downloads []*DownloadItem
	idGetter  int64 // variable for setting download id
	sync.Mutex
}

func NewDownloadManager() *DownloadManager {
	return &DownloadManager{
		Downloads: []*DownloadItem{},
		idGetter:  0,
	}
}

// resume download by creating  new ctx
func (d *DownloadManager) ResumeDownload(id int64) {
	d.Lock()
	defer d.Unlock()

	for _, item := range d.Downloads {
		if item.Id == id {
			ctx, cancel := context.WithCancel(context.Background())
			item.changeCtx(ctx, cancel)
			// run in background
			go item.download()

		}
	}
}

// add download to slice !!! not starting just adding
func (d *DownloadManager) AddDownload(url, filepath, filename string) *DownloadItem {
	d.Lock()
	defer d.Unlock()

	ctx, cancel := context.WithCancel(context.Background())

	downloadItem := &DownloadItem{
		Id:       d.idGetter,
		Url:      url,
		Filepath: filepath,
		Filename: filename,
		Ctx:      ctx,
		Cancel:   cancel,
	}
	// !!! must increment
	d.idGetter++

	d.Downloads = append(d.Downloads, downloadItem)
	return downloadItem
}

// stop download by canceling ctx
func (d *DownloadManager) StopDownload(downloadItem *DownloadItem) {
	if downloadItem.Cancel != nil {
		downloadItem.Cancel()
	}

}

// delete download from slice
func (d *DownloadManager) DeleteDownload(id int64) error {
	d.Lock()
	defer d.Unlock()
	for i, item := range d.Downloads {
		if item.Id == id {
			// cancel ctx in still downloading
			if item.Active {
				item.Cancel()
			}

			d.Downloads = append(d.Downloads[:i], d.Downloads[i+1:]...)
			return nil

		}
	}

	return fmt.Errorf("downloadItem with id: %d not found\n", id)
}

// start download in background
func (d *DownloadManager) StartDownload(item *DownloadItem) {
	go item.download()
}

func (d *DownloadManager) GetItemById(id int64) *DownloadItem {
	d.Lock()
	defer d.Unlock()

	for _, item := range d.Downloads {
		if item.Id == id {
			return item
		}
	}

	return nil
}

// get all downloading as snapshot,
func (d *DownloadManager) GetAllDownloads() []dto.DownloadItemDto {
	d.Lock()
	defer d.Unlock()
	dtos := make([]dto.DownloadItemDto, 0, len(d.Downloads))
	for _, item := range d.Downloads {
		// append increases the length automatically
		dtos = append(dtos, item.getData())
	}

	return dtos

}
