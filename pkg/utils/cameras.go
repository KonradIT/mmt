package utils

import (
	"archive/zip"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/dustin/go-humanize"
	"github.com/fatih/color"
	mErrors "github.com/konradit/mmt/pkg/errors"
	"github.com/vbauerster/mpb/v8"
	"github.com/vbauerster/mpb/v8/decor"
)

type Camera int

const (
	GoPro Camera = iota
	DJI
	Insta360
	Android
)

func (c Camera) toString() string {
	extensions := [...]string{"gopro", "dji", "insta360", "android"}

	return extensions[c]
}

func CameraGet(s string) (Camera, error) {
	switch s {
	case GoPro.toString():
		return GoPro, nil
	case DJI.toString():
		return DJI, nil
	case Insta360.toString():
		return Insta360, nil
	case Android.toString():
		return Android, nil
	default:
		return 10, mErrors.ErrUnsupportedCamera(s)
	}
}

type Result struct {
	FilesImported    int
	FilesNotImported []string
	Errors           []error
}

type ConnectionType string

const (
	SDCard  ConnectionType = "sd_card"
	Connect ConnectionType = "connect"
)

func CopyFile(src string, dst string, buffersize int, progressbar *mpb.Bar) error {
	sourceFileStat, err := os.Stat(src)
	if err != nil {
		return err
	}

	if !sourceFileStat.Mode().IsRegular() {
		return fmt.Errorf("%s is not a regular file", src)
	}

	source, err := os.Open(src)
	if err != nil {
		return err
	}
	defer source.Close()

	_, err = os.Stat(dst)
	if err == nil {
		return fmt.Errorf("File %s already exists", dst)
	}

	destination, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer destination.Close()

	if err != nil {
		panic(err)
	}

	if progressbar == nil {
		p := mpb.New(
			mpb.WithWidth(60),
			mpb.WithRefreshRate(180*time.Millisecond),
		)

		progressbar = p.New(sourceFileStat.Size(),
			mpb.BarStyle().Rbound("|"),
			mpb.PrependDecorators(
				decor.CountersKibiByte("% .2f / % .2f"),
			),
			mpb.AppendDecorators(
				decor.EwmaETA(decor.ET_STYLE_GO, 90),
				decor.Name(" ] "),
				decor.EwmaSpeed(decor.UnitKiB, "% .2f", 60),
			),
		)
	}

	buf := make([]byte, buffersize)
	proxyReader := progressbar.ProxyReader(source)

	defer proxyReader.Close()
	for {
		n, err := proxyReader.Read(buf)
		if err != nil && err != io.EOF {
			return err
		}

		if n == 0 {
			break
		}

		if _, err := destination.Write(buf[:n]); err != nil {
			return err
		}
	}

	return nil
}

type WriteCounter struct {
	Total uint64
}

func (wc *WriteCounter) Write(p []byte) (int, error) {
	n := len(p)
	wc.Total += uint64(n)
	wc.PrintProgress()
	return n, nil
}

func (wc WriteCounter) PrintProgress() {
	// Clear the line by using a character return to go back to the start and remove
	// the remaining characters by filling it with spaces
	fmt.Printf("\r%s", strings.Repeat(" ", 35))

	// Return again and print current status of download
	// We use the humanize package to print the bytes in a meaningful way (e.g. 10 MB)
	fmt.Printf("\rDownloading... %s complete", humanize.Bytes(wc.Total))
}

func DownloadFile(filepath string, url string, progressbar *mpb.Bar) error {
	// Create the file, but give it a tmp file extension, this means we won't overwrite a
	// file until it's downloaded, but we'll remove the tmp extension once downloaded.
	out, err := os.Create(filepath + ".tmp")
	if err != nil {
		return err
	}

	// Get the data
	resp, err := http.Get(url) // #nosec
	if err != nil {
		out.Close()
		return err
	}
	defer resp.Body.Close()

	if progressbar != nil {
		proxyReader := progressbar.ProxyReader(resp.Body)
		defer proxyReader.Close()

		if _, err = io.Copy(out, proxyReader); err != nil {
			out.Close()
			return err
		}
	} else {
		counter := &WriteCounter{}
		if _, err = io.Copy(out, io.TeeReader(resp.Body, counter)); err != nil {
			out.Close()
			return err
		}
	}
	// The progress use the same line so print a new line once it's finished downloading
	fmt.Print("\n")

	// Close the file without defer so it can happen before Rename()
	out.Close()

	if err = os.Rename(filepath+".tmp", filepath); err != nil {
		return err
	}
	return nil
}

func Unzip(src string, dest string) error {
	r, err := zip.OpenReader(src)
	if err != nil {
		return err
	}
	defer r.Close()

	for _, f := range r.File {
		// Store filename/path for returning and using later on
		fpath := filepath.Join(dest, f.Name)                                          // #nosec
		if !strings.HasPrefix(fpath, filepath.Clean(dest)+string(os.PathSeparator)) { // #nosec
			return fmt.Errorf("%s: illegal file path", fpath)
		}

		if f.FileInfo().IsDir() {
			// Make Folder
			if err = os.MkdirAll(fpath, os.ModePerm); err != nil {
				return err
			}
			continue
		}

		// Make File
		if err = os.MkdirAll(filepath.Dir(fpath), os.ModePerm); err != nil {
			return err
		}

		outFile, err := os.OpenFile(fpath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, f.Mode())
		if err != nil {
			return err
		}

		rc, err := f.Open()
		if err != nil {
			return err
		}

		for {
			_, err := io.CopyN(outFile, rc, 1024)
			if err != nil {
				if err == io.EOF {
					break
				}
				return err
			}
		}
		// Close the file without defer to close before next iteration of loop
		outFile.Close()
		rc.Close()

		if err != nil {
			return err
		}
	}
	return nil
}

func ParseCliOptions(cameraOptions map[string]interface{}) SortOptions {
	byCamera := false
	byLocation := false
	skipAux := false
	skipAuxOption, found := cameraOptions["skip_aux"]
	if found {
		skipAux = skipAuxOption.(bool)
	}
	sortByOptions, found := cameraOptions["sort_by"]
	if found {
		for _, sortop := range sortByOptions.([]string) {
			if sortop == "camera" {
				byCamera = true
			}
			if sortop == "location" {
				byLocation = true
			}
		}
	}

	return SortOptions{
		ByCamera:           byCamera,
		ByLocation:         byLocation,
		SkipAuxiliaryFiles: skipAux,
	}
}

func FindFolderInPath(entirePath, directory string) (string, error) {
	modified := filepath.Dir(entirePath)
	if filepath.Base(modified) == directory {
		return modified, nil
	}
	if filepath.Base(entirePath) == directory {
		return entirePath, nil
	}
	if entirePath == "." || modified == entirePath {
		return "", mErrors.ErrNotFound(directory)
	}
	return FindFolderInPath(modified, directory)
}

type ResultCounter struct {
	mu               sync.Mutex
	CameraName       string
	Errors           []error
	FilesNotImported []string
	FilesImported    int
}

func (rc *ResultCounter) SetFailure(err error, file string) {
	rc.mu.Lock()
	rc.Errors = append(rc.Errors, err)
	rc.FilesNotImported = append(rc.FilesNotImported, file)
	rc.mu.Unlock()
}

func (rc *ResultCounter) SetCameraName(camName string) {
	rc.mu.Lock()
	rc.CameraName = camName
	rc.mu.Unlock()
}

func (rc *ResultCounter) GetCameraName() string {
	rc.mu.Lock()
	defer rc.mu.Unlock()
	return rc.CameraName
}

func (rc *ResultCounter) SetSuccess() {
	rc.mu.Lock()
	rc.FilesImported++
	rc.mu.Unlock()
}

func (rc *ResultCounter) Get() Result {
	rc.mu.Lock()
	defer rc.mu.Unlock()
	return Result{
		FilesImported:    rc.FilesImported,
		FilesNotImported: rc.FilesNotImported,
		Errors:           rc.Errors,
	}
}

type BarType int

const (
	IoTX BarType = iota
	Percentage
)

func GetNewBar(progressBar *mpb.Progress, total int64, filename string, barType BarType) *mpb.Bar {
	decorator := decor.CountersKiloByte("% .2f / % .2f")
	if barType == Percentage {
		decorator = decor.Percentage(decor.WCSyncSpace)
	}
	return progressBar.AddBar(total,
		mpb.PrependDecorators(
			decor.Name(color.CyanString(fmt.Sprintf("%s: ", filename))),
			decorator,
		),
		mpb.AppendDecorators(
			decor.OnComplete(
				decor.EwmaETA(decor.ET_STYLE_GO, 60, decor.WCSyncWidth), "✔️",
			),
		),
	)
}
