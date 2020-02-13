package gui

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"

	"github.com/digitalautonomy/wahay/config"
	"github.com/kardianos/osext"
	log "github.com/sirupsen/logrus"
)

type installation struct {
	g            Graphics
	iconProvider *icon
	dataHome     string
}

func (u *gtkUI) ensureInstallation() {
	if u.g.gdk == nil {
		log.Error("ensureInstallation(): UI Graphics hasn't been initialized correctly.")
		return
	}

	dataHome := config.XdgDataHome()
	if len(dataHome) == 0 && !fileExists(dataHome) {
		log.Fatal("No data home directory is available")
	}

	i := &installation{
		g:        u.g,
		dataHome: dataHome,
	}

	i.ensureApplicationIcons()
	i.ensureApplicationDesktop()
}

var iconSizes = []int{16, 32, 48, 128, 256}

func (i *installation) ensureApplicationIcons() {
	i.iconProvider = setupApplicationIcon(i.g)

	for _, size := range iconSizes {
		icon, err := i.iconProvider.createPixBufWithSize(size, size)
		if err != nil {
			log.WithFields(log.Fields{
				"size": size,
			}).Debug(fmt.Sprintf("saveIconToFolder(): %s", err.Error()))
			continue
		}

		sizeFolder := fmt.Sprintf("%dx%d", size, size)
		dest := filepath.Join(i.dataHome, "icons/hicolor/"+sizeFolder+"/apps")
		err = os.MkdirAll(dest, 0700)
		if err != nil {
			log.WithFields(log.Fields{
				"directory": dest,
			}).Debug(fmt.Sprintf("saveIconToFolder(): %s", err.Error()))
			continue
		}

		log.Debugf("Saving the application icon with size=%dx%d to folder=%s", size, size, dest)
		fileName := filepath.Join(dest, i.iconProvider.fileName())
		err = icon.SavePNG(fileName, 9)
		if err != nil {
			log.WithFields(log.Fields{
				"fileName": fileName,
			}).Debug(fmt.Sprintf("saveIconToFolder(): %s", err.Error()))
			continue
		}
	}
}

func (i *installation) ensureApplicationDesktop() {
	dir := filepath.Join(i.dataHome, "applications")
	err := os.MkdirAll(dir, 0700)
	if err != nil {
		log.WithFields(log.Fields{
			"desktopFileDir": dir,
		}).Errorf("ensureApplicationDesktop(): %s", err.Error())
	}

	fileName := filepath.Join(i.dataHome, "applications", "wahay.desktop")
	content := i.generateDesktopFile()

	err = ioutil.WriteFile(fileName, []byte(content), 0600)
	if err != nil {
		log.WithFields(log.Fields{
			"desktopFileName": fileName,
		}).Errorf("ensureApplicationDesktop(): %s", err.Error())
	}
}

func (i *installation) generateDesktopFile() string {
	path, _ := osext.Executable()
	icon := i.iconProvider.fileNameNoExt()

	return fmt.Sprintf(`
#!/usr/bin/env xdg-open
[Desktop Entry]
Type=Application
Version=1.0
Encoding=UTF-8
Name=%s
Comment=Secure and Decentralized Conference Call Application
Exec=%s
Icon=%s
Terminal=false
Type=Application
Categories=Network`, programName, path, icon)
}

func fileExists(filename string) bool {
	_, err := os.Stat(filename)
	return err == nil
}
