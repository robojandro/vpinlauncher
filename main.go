package main

import (
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/widget"
	"github.com/pkg/errors"
)

func main() {
	a := app.New()
	w := a.NewWindow("Visual Pinball Launcher")
	w.Resize(fyne.NewSize(900, 1000))

	appPath := "/home/marco/VPinball/VPinballX_GL"
	tablesPath := "/home/marco/VPinball/tables/"

	popupImgErrs := false

	tables, err := scanTables(tablesPath)
	if err != nil {
		errD := dialog.NewError(err, w)
		errD.Show()
	}

	listView := widget.NewList(func() int {
		return len(tables)
	}, func() fyne.CanvasObject {
		return widget.NewLabel("template")
	}, func(id widget.ListItemID, object fyne.CanvasObject) {
		object.(*widget.Label).SetText(formatFileName(tables[id]))
	})

	img, err := loadImage(tables[0])
	if err != nil {
		if popupImgErrs {
			errD := dialog.NewError(err, w)
			errD.Show()
		} else {
			fmt.Println(err)
		}
	}

	currentFileName := tables[0]

	clicked := func() {
		log.Printf("clicked on file: %s\n", currentFileName)
		cmd := exec.Command(appPath, "-play", fmt.Sprintf("%s/%s", tablesPath, currentFileName))
		stdout, err := cmd.Output()
		if err != nil {
			fmt.Println(err.Error())
			os.Exit(1)
		}
		if strings.Contains(string(stdout), "Player closed.") {
			log.Printf("table '%s' was closed\n", currentFileName)
			// os.Exit(0)
		}
	}

	playBtn := widget.NewButton("play", clicked)

	btnContainer := container.NewStack(playBtn)
	imgContainer := container.NewStack(img)

	listView.OnSelected = func(id widget.ListItemID) {
		currentFileName = tables[id]
		fmt.Println(tables[id])
		img, err := loadImage(tables[id])
		if err != nil {
			if popupImgErrs {
				errD := dialog.NewError(err, w)
				errD.Show()
			} else {
				fmt.Println(err)
			}
		}
		imgContainer.Objects[0] = img
	}

	vSplit := container.NewVSplit(imgContainer, btnContainer)
	vSplit.Offset = .8

	hSplit := container.NewHSplit(listView, vSplit)
	hSplit.Offset = .24

	w.SetContent(hSplit)

	w.ShowAndRun()
}

func loadImage(fileName string) (*canvas.Image, error) {
	img := &canvas.Image{}
	if fileName == "" {
		return img, errors.New("fileName was empty")
	}

	frmtd := normalizeFileName(fileName)
	// filePath := "../table_snapshots/" + frmtd + ".png"
	filePath := "table_snapshots/" + frmtd + ".png"

	if _, err := os.Stat(filePath); err != nil {
		return img, fmt.Errorf("did not find table image file: %s\n", err)
	}

	img = canvas.NewImageFromFile(filePath)
	img.FillMode = canvas.ImageFillContain
	return img, nil
}

// Removes the extension, parens, and replaces with underscores
func normalizeFileName(fileName string) string {
	woSuffix := strings.TrimSuffix(fileName, filepath.Ext(fileName))
	var rep string
	if strings.Contains(woSuffix, "(") {
		rep = strings.Replace(woSuffix, "(", "", -1)
		rep = strings.Replace(rep, ")", "", -1)
	} else {
		rep = woSuffix
	}
	rep = strings.Replace(rep, " ", "_", -1)
	rep = strings.ToLower(rep)
	return rep
}

func formatFileName(fileName string) string {
	woSuffix := strings.TrimSuffix(fileName, filepath.Ext(fileName))
	var parts []string
	if strings.Contains(woSuffix, "(") {
		parts = strings.Split(woSuffix, ")")
		parts[0] += ")"
	} else {
		parts = append(parts, woSuffix)
	}
	return parts[0]
}

func scanTables(tablesPath string) ([]string, error) {
	var tables []string

	if tablesPath == "" {
		return nil, errors.New("tablesPath was empty")
	}

	dirContents, err := os.ReadDir(tablesPath)
	if err != nil {
		return nil, err
	}
	for _, found := range dirContents {
		// fmt.Printf("found: %v\n", found.Name())
		if !strings.HasSuffix(found.Name(), ".vpx") {
			continue
		}

		if found.IsDir() {
			continue
		}
		tables = append(tables, found.Name())
	}
	return tables, nil
}
