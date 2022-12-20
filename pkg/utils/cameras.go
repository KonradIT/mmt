package utils

import (
	"archive/zip"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/cheggaaa/pb"
	"github.com/dustin/go-humanize"
	mErrors "github.com/konradit/mmt/pkg/errors"
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
		return 10, mErrors.ErrUnsupportedCamera
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

func CopyFile(src string, dst string, buffersize int) error {
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

	bar := pb.StartNew(int(sourceFileStat.Size()/1000) + 1)

	buf := make([]byte, buffersize)
	for {
		n, err := source.Read(buf)
		if err != nil && err != io.EOF {
			return err
		}

		if n == 0 {
			break
		}
		bar.Increment()

		if _, err := destination.Write(buf[:n]); err != nil {
			return err
		}
	}
	bar.Finish()
	return err
}

// MIT Licensed code: https://gist.github.com/r0l1/92462b38df26839a3ca324697c8cba04
func CopyDir(src string, dst string, bufferSize int) (err error) {
	src = filepath.Clean(src)
	dst = filepath.Clean(dst)

	si, err := os.Stat(src)
	if err != nil {
		return err
	}
	if !si.IsDir() {
		return errors.New("source is not a directory")
	}

	_, err = os.Stat(dst)
	if err != nil && !os.IsNotExist(err) {
		return err
	}
	if err == nil {
		return errors.New("destination already exists")
	}

	err = os.MkdirAll(dst, si.Mode())
	if err != nil {
		return err
	}

	entries, err := ioutil.ReadDir(src)
	if err != nil {
		return err
	}

	for _, entry := range entries {
		srcPath := filepath.Join(src, entry.Name())
		dstPath := filepath.Join(dst, entry.Name())

		if entry.IsDir() {
			err = CopyDir(srcPath, dstPath, bufferSize)
			if err != nil {
				return
			}
		} else {
			// Skip symlinks.
			if entry.Mode()&os.ModeSymlink != 0 {
				continue
			}

			err = CopyFile(srcPath, dstPath, bufferSize)
			if err != nil {
				return
			}
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

func DownloadFile(filepath string, url string) error {
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

	// Create our progress reporter and pass it to be used alongside our writer
	counter := &WriteCounter{}
	if _, err = io.Copy(out, io.TeeReader(resp.Body, counter)); err != nil {
		out.Close()
		return err
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
