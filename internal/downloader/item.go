package downloader

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"sync"

	"github.com/matejeliash/medownloader/internal/dto"
)

type DownloadItem struct {
	sync.Mutex // embed mutex
	Id         int64
	Url        string
	Filename   string
	Filepath   string
	Active     bool
	Completed  bool
	Downloaded int64
	Size       int64

	Ctx    context.Context    // ctx for signaling  goroutine
	Cancel context.CancelFunc // run on cancel

	Err error
}

// get "snapshot" of downloads slice
func (d *DownloadItem) getData() dto.DownloadItemDto {
	d.Lock()
	defer d.Unlock()

	// convert error to string
	errStr := ""
	if d.Err != nil {
		errStr = d.Err.Error()
	}

	dto := dto.DownloadItemDto{
		Id:         d.Id,
		Url:        d.Url,
		Filename:   d.Filename,
		Filepath:   d.Filepath,
		Active:     d.Active,
		Completed:  d.Completed,
		Downloaded: d.Downloaded,
		Size:       d.Size,
		Err:        errStr,
	}
	return dto

}

// change ctx, used when we want to start downloading again
func (d *DownloadItem) changeCtx(ctx context.Context, cancel context.CancelFunc) {
	d.Lock()
	defer d.Unlock()
	d.Ctx = ctx
	d.Cancel = cancel
}

func (d *DownloadItem) setDone() {
	d.Lock()
	d.Completed = true
	d.Active = false
	d.Unlock()
}

func (d *DownloadItem) setStopped() {
	d.Lock()
	d.Active = false
	d.Unlock()
}

// set error and also match booleans to error
func (d *DownloadItem) setError(err error) {
	d.Lock()
	d.Completed = false
	d.Active = false
	d.Err = err
	d.Unlock()

}

func (d *DownloadItem) download() {

	//used for resuming
	var resumeByte int64 = 0

	if info, err := os.Stat(d.Filepath); err == nil {
		resumeByte = info.Size()
	}

	req, err := http.NewRequestWithContext(d.Ctx, http.MethodGet, d.Url, nil)
	if err != nil {
		d.setError(err)
		return
	}

	if resumeByte > 0 {
		req.Header.Set("Range", fmt.Sprintf("bytes=%d-", resumeByte))
	}

	httpCient := &http.Client{
		// later add timeout or stall timeout

	}

	// send request
	resp, err := httpCient.Do(req)
	if err != nil {
		d.setError(err)
		return
	}

	defer resp.Body.Close()

	fmt.Printf("creating file: %s\n", d.Filepath)

	// keep if exists, otherwise create
	file, err := os.OpenFile(d.Filepath, os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		d.setError(err)
		return
	}

	defer file.Close()

	// seek to the resume position
	if resumeByte > 0 {
		_, err := file.Seek(resumeByte, 0)
		if err != nil {
			d.setError(err)
			return
		}
	}

	// set flags, so they represent actively downloading
	d.Lock()
	d.Size = resp.ContentLength + resumeByte
	d.Downloaded = resumeByte
	d.Active = true
	d.Err = nil
	d.Completed = false
	d.Unlock()

	//32k buffer
	buf := make([]byte, 1024*32)

	for {
		num, err := resp.Body.Read(buf)

		if num > 0 {
			_, writeErr := file.Write(buf[:num])
			if writeErr != nil {
				d.setError(writeErr)
				return
			}
			// update with mutex
			d.Lock()
			d.Downloaded += int64(num)
			d.Unlock()
		}

		if err != nil {
			// download finished
			if err == io.EOF {
				d.setDone()
				return
			}
			// ctx used to slop download
			if d.Ctx.Err() != nil {
				if errors.Is(err, context.Canceled) {
					// normal stop, do not set error
					d.setStopped()
					return
				}

				d.setError(d.Ctx.Err())
				return
			}

			d.setError(err)
			return
		}
	}

}
