package utils

import (
	"archive/zip"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/dustin/go-humanize"
	"github.com/vbauerster/mpb/v8"
	"github.com/vbauerster/mpb/v8/decor"
)

func CopyFile(src string, dst string, buffersize int, progressbar *mpb.Bar, modTime time.Time) error {
	sourceFileStat, err := os.Stat(src)
	if err != nil {
		return err
	}

	source, err := os.Open(src)
	if err != nil {
		return err
	}
	defer source.Close()

	_, err = os.Stat(dst)
	if err == nil {
		return fmt.Errorf("file %s already exists", dst)
	}

	destination, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer destination.Close()

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
		if err != nil && !errors.Is(err, io.EOF) {
			return err
		}

		if n == 0 {
			break
		}

		_, err = destination.Write(buf[:n])
		if err != nil {
			return err
		}
	}

	return os.Chtimes(dst, modTime, modTime)
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

func (wc *WriteCounter) PrintProgress() {
	fmt.Printf("\r%s", strings.Repeat(" ", 35))
	fmt.Printf("\rDownloading... %s complete", humanize.Bytes(wc.Total))
}

func DownloadFile(filepath string, url string, progressbar *mpb.Bar, mtime *time.Time) error {
	out, err := os.Create(filepath + ".tmp")
	if err != nil {
		return err
	}

	resp, err := Client.Get(url) // #nosec
	if err != nil {
		_ = out.Close()

		return err
	}
	defer resp.Body.Close()

	if progressbar != nil {
		proxyReader := progressbar.ProxyReader(resp.Body)
		defer proxyReader.Close()

		_, err = io.Copy(out, proxyReader)
		if err != nil {
			_ = out.Close()

			return err
		}
	} else {
		counter := &WriteCounter{}

		_, err = io.Copy(out, io.TeeReader(resp.Body, counter))
		if err != nil {
			_ = out.Close()

			return err
		}
	}

	fmt.Print("\n")

	err = out.Close()
	if err != nil {
		return err
	}

	if mtime != nil {
		err := os.Chtimes(filepath+".tmp", time.Time{}, *mtime)
		if err != nil {
			return err
		}
	}

	return os.Rename(filepath+".tmp", filepath)
}

func Unzip(src string, dest string) error {
	r, err := zip.OpenReader(src)
	if err != nil {
		return err
	}
	defer r.Close() //nolint:errcheck // read-only zip reader

	for _, f := range r.File {
		fpath := filepath.Join(dest, f.Name)                                          // #nosec
		if !strings.HasPrefix(fpath, filepath.Clean(dest)+string(os.PathSeparator)) { // #nosec
			return fmt.Errorf("%s: illegal file path", fpath)
		}

		if f.FileInfo().IsDir() {
			err := os.MkdirAll(fpath, os.ModePerm)
			if err != nil {
				return err
			}

			continue
		}

		err = os.MkdirAll(filepath.Dir(fpath), os.ModePerm)
		if err != nil {
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
				if errors.Is(err, io.EOF) {
					break
				}

				return err
			}
		}

		_ = outFile.Close()
		_ = rc.Close()
	}

	return nil
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
		return "", fmt.Errorf("unable to find %s", directory)
	}

	return FindFolderInPath(modified, directory)
}

var DateFormatReplacer = strings.NewReplacer("dd", "02", "mm", "01", "yyyy", "2006")
