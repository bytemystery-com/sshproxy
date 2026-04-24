// Copyright (c) 2026 Reiner Pröls
//
// Permission is hereby granted, free of charge, to any person obtaining a copy
// of this software and associated documentation files (the "Software"), to deal
// in the Software without restriction, including without limitation the rights
// to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
// copies of the Software, and to permit persons to whom the Software is
// furnished to do so, subject to the following conditions:
//
// The above copyright notice and this permission notice shall be included in
// all copies or substantial portions of the Software.
//
// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
// FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
// AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
// LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
// OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
// SOFTWARE.
//
// SPDX-License-Identifier: MIT
//
// Author: Reiner Pröls

package util

import (
	"encoding/json"
	"fmt"
	"image/color"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/storage"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
)

func NewFiller(width, height float32) *canvas.Rectangle {
	filler := canvas.NewRectangle(color.Transparent)
	filler.SetMinSize(fyne.NewSize(width, height))
	filler.Refresh()
	return filler
}

func NewHFiller(f float32) *canvas.Rectangle {
	filler := canvas.NewRectangle(color.Transparent)
	filler.SetMinSize(fyne.NewSize(GetDefaultTextWidth("X")*f, 0))
	filler.Refresh()
	return filler
}

func NewVFiller(f float32) *canvas.Rectangle {
	filler := canvas.NewRectangle(color.Transparent)
	filler.SetMinSize(fyne.NewSize(0, GetDefaultTextHeight("X")*f))
	filler.Refresh()
	return filler
}

func GetNumberFilter(w *widget.Entry, f func(string)) func(string) {
	return func(s string) {
		filtered := ""
		for _, r := range s {
			if r >= '0' && r <= '9' {
				filtered += string(r)
			}
		}
		if filtered != s {
			w.SetText(filtered)
		}
		if f != nil {
			f(w.Text)
		}
	}
}

func GetNumberFilterPlusMinus(w *widget.Entry, f func(string)) func(string) {
	return func(s string) {
		filtered := ""
		first := true
		for _, r := range s {
			if (r >= '0' && r <= '9') || (first && (r == '-' || r == '+')) {
				filtered += string(r)
			}
			first = false
		}
		if filtered != s {
			w.SetText(filtered)
		}
		if f != nil {
			f(w.Text)
		}
	}
}

func GetNoEditFilter(w *widget.Entry, f func(string), defText string) func(string) {
	return func(s string) {
		w.SetText(defText)
		if f != nil {
			f(w.Text)
		}
	}
}

func GetDefaultTextWidth(s string) float32 {
	return fyne.MeasureText(
		s,
		theme.TextSize(),
		fyne.TextStyle{},
	).Width
}

func GetDefaultTextHeight(s string) float32 {
	return fyne.MeasureText(
		s,
		theme.TextSize(),
		fyne.TextStyle{},
	).Width
}

func GetDefaultTextSize(s string) fyne.Size {
	return fyne.MeasureText(
		s,
		theme.TextSize(),
		fyne.TextStyle{},
	)
}

type WriterFunc func(p []byte) (int, error)

func (f WriterFunc) Write(p []byte) (int, error) {
	return f(p)
}

func GetFilename(path string) string {
	path = strings.ReplaceAll(path, "\\", "/")
	i := strings.LastIndex(path, "/")
	if i == -1 {
		return path
	}
	return path[i+1:]
}

type TruncateType int

const (
	None TruncateType = iota
	Begin
	End
)

func TruncateText(s string, maxWidth float32, text *canvas.Text, truncate TruncateType) string {
	if truncate == None {
		return s
	}
	maxWidth -= theme.Padding() * 2
	ellipsis := "…"
	ellW := fyne.MeasureText(ellipsis, text.TextSize, text.TextStyle).Width

	r := []rune(s)
	if fyne.MeasureText(s, text.TextSize, text.TextStyle).Width <= maxWidth {
		return s
	}

	for len(r) > 0 {
		switch truncate {
		case End:
			r = r[:len(r)-1]
		case Begin:
			r = r[1:]
		}

		if fyne.MeasureText(string(r), text.TextSize, text.TextStyle).Width+ellW <= maxWidth {
			switch truncate {
			case End:
				return string(r) + ellipsis
			case Begin:
				return ellipsis + string(r)
			}
		}
	}
	return ellipsis
}

func DebugContainer(obj fyne.CanvasObject, col color.Color) fyne.CanvasObject {
	if col == nil {
		col = color.RGBA{255, 0, 0, 65}
	}
	bg := canvas.NewRectangle(col)
	bg.SetMinSize(obj.MinSize())

	return container.NewStack(bg, obj)
}

func WriteFileToStorage(writer fyne.URIWriteCloser, data []byte) error {
	for len(data) > 0 {
		n, err := writer.Write(data)
		if err != nil {
			return err
		}
		data = data[n:]
	}
	return nil
}

func WriteFile(file string, data []byte) error {
	// Windows: needed ???
	if filepath.VolumeName(file) != "" {
		return os.WriteFile(file, data, 0o664)
	}

	// URI -> use Fyne storage
	uri, err := storage.ParseURI(file)
	if err == nil && uri.Scheme() != "" {
		// URI -> Fyne storage will be used
		w, err := storage.Writer(uri)
		if err != nil {
			return err
		}
		defer w.Close()

		err = WriteFileToStorage(w, data)
		return err
	}

	// Fallback
	return os.WriteFile(file, data, 0o664)
}

func RenameFile(src, target string) error {
	_, err := os.Stat(target)
	if err == nil {
		err = os.Remove(target)
		if err != nil {
			return err
		}
	}
	err = os.Rename(src, target)
	return err
}

func FormatDateTime(ts time.Time, long bool) string {
	if long {
		return ts.In(time.Local).Format("02.01.2006 - 15:04:05")
	} else {
		return ts.In(time.Local).Format("02.01.06 - 15:04")
	}
}

type GitHubRelease struct {
	TagName string `json:"tag_name"`
	HTMLURL string `json:"html_url"`
}

func CheckForUpdate() (string, string, error) {
	url := "https://api.github.com/repos/bytemystery-com/sshproxy/releases/latest"
	resp, err := http.Get(url)
	if err != nil {
		return "", "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return "", "", fmt.Errorf("HTTP %d", resp.StatusCode)
	}

	var release GitHubRelease
	err = json.NewDecoder(resp.Body).Decode(&release)
	if err != nil {
		return "", "", err
	}
	return release.HTMLURL, release.TagName, nil
}

func ReadKeyFile(file string) ([]byte, error) {
	// Windows: needed ???
	if filepath.VolumeName(file) != "" {
		data, err := os.ReadFile(file)
		return data, err
	}

	// URI -> use Fyne storage
	uri, err := storage.ParseURI(file)
	if err == nil && uri.Scheme() != "" {
		// URI -> Fyne storage will be used
		r, err := storage.Reader(uri)
		if err != nil {
			return nil, err
		}
		defer r.Close()
		data, err := io.ReadAll(r)
		if err != nil {
			return nil, err
		}
		return data, nil
	}

	// Fallback
	data, err := os.ReadFile(file)
	return data, err
}
