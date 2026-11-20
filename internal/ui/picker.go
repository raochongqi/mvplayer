package ui

import (
	"os"
	"path/filepath"
	"sort"
	"strings"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/storage"
	"fyne.io/fyne/v2/widget"
)

// multiPicker is a small directory browser with multi-select, because Fyne's
// stock file dialog only returns one file.
type multiPicker struct {
	win      fyne.Window
	dir      fyne.URI
	entries  []fyne.URI
	selected map[string]bool
	list     *widget.List
	pathLbl  *widget.Label
	onPick   func([]string)
	dlg      *dialog.ConfirmDialog
}

var videoExts = map[string]bool{".mkv": true, ".mp4": true, ".webm": true, ".mov": true, ".avi": true}

func newMultiPicker(win fyne.Window, onPick func([]string)) *multiPicker {
	home, err := os.UserHomeDir()
	if err != nil {
		home = "."
	}
	p := &multiPicker{
		win:      win,
		dir:      storage.NewFileURI(home),
		selected: map[string]bool{},
		onPick:   onPick,
	}
	p.pathLbl = widget.NewLabel("")
	p.pathLbl.Truncation = fyne.TextTruncateEllipsis
	p.list = widget.NewList(
		func() int { return len(p.entries) },
		func() fyne.CanvasObject { return widget.NewLabel("") },
		func(i widget.ListItemID, o fyne.CanvasObject) {
			if i < 0 || i >= len(p.entries) {
				return
			}
			u := p.entries[i]
			label := u.Name()
			if isDir(u) {
				label = "▸ " + label
			} else if p.selected[u.Path()] {
				label = "☑ " + label
			} else {
				label = "☐ " + label
			}
			o.(*widget.Label).SetText(label)
		},
	)
	p.list.OnSelected = func(i widget.ListItemID) {
		p.list.UnselectAll()
		if i < 0 || i >= len(p.entries) {
			return
		}
		u := p.entries[i]
		if isDir(u) {
			p.navigate(u)
			return
		}
		path := u.Path()
		p.selected[path] = !p.selected[path]
		p.list.Refresh()
	}
	up := widget.NewButton("上一级", func() {
		if parent, err := storage.Parent(p.dir); err == nil && parent != nil {
			p.navigate(parent)
		}
	})
	body := container.NewBorder(container.NewBorder(nil, nil, up, nil, p.pathLbl), nil, nil, nil, p.list)
	p.dlg = dialog.NewCustomConfirm("打开视频（可多选）", "打开选中", "取消", body, func(ok bool) {
		if !ok {
			return
		}
		var paths []string
		for path := range p.selected {
			paths = append(paths, path)
		}
		sort.Strings(paths)
		if len(paths) > 0 && p.onPick != nil {
			p.onPick(paths)
		}
	}, win)
	p.dlg.Resize(fyne.NewSize(680, 520))
	p.navigate(p.dir)
	return p
}

func (p *multiPicker) Show() { p.dlg.Show() }

func (p *multiPicker) navigate(dir fyne.URI) {
	lister, err := storage.ListerForURI(dir)
	if err != nil {
		return
	}
	items, err := lister.List()
	if err != nil {
		return
	}
	var dirs, files []fyne.URI
	for _, it := range items {
		if strings.HasPrefix(it.Name(), ".") {
			continue
		}
		if isDir(it) {
			dirs = append(dirs, it)
			continue
		}
		if videoExts[strings.ToLower(filepath.Ext(it.Name()))] {
			files = append(files, it)
		}
	}
	sort.Slice(dirs, func(i, j int) bool { return dirs[i].Name() < dirs[j].Name() })
	sort.Slice(files, func(i, j int) bool { return files[i].Name() < files[j].Name() })
	p.entries = append(dirs, files...)
	p.dir = dir
	p.pathLbl.SetText(dir.Path())
	p.list.Refresh()
}

func isDir(u fyne.URI) bool {
	return u.Scheme() == "file" && u.Extension() == "" && dirExists(u.Path())
}

func dirExists(path string) bool {
	fi, err := os.Stat(path)
	return err == nil && fi.IsDir()
}
