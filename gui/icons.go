package gui

import (
	"crypto/sha256"
	"encoding/hex"

	"github.com/coyim/gotk3adapter/gdki"
	log "github.com/sirupsen/logrus"
)

type icon struct {
	decoded       []byte
	size          string
	width, height int
	path          string
	name          string
	g             Graphics
	cached        gdki.Pixbuf
}

var wahayICON *icon

func setupApplicationIcon(g Graphics) *icon {
	wahayICON = &icon{
		name:   "wahay-512x512.png",
		path:   "gui/images/wahay-512x512.png",
		size:   "512x512",
		width:  256,
		height: 256,
		g:      g,
	}

	return wahayICON
}

func getApplicationIcon() *icon {
	if wahayICON == nil {
		log.Fatal("The application icon hasn't been initialized")
	}
	return wahayICON
}

func (i *icon) get() []byte {
	if i.decoded == nil {
		i.decoded = i.g.getImage(i.name)
	}
	return i.decoded
}

func (i *icon) getPixbuf() gdki.Pixbuf {
	var err error

	if i.cached == nil {
		i.cached, err = i.createPixBuf()
		if err != nil {
			panic(err)
		}
	}

	return i.cached
}

func (i *icon) fileName() string {
	return i.fileNameNoExt() + pngExtension
}

func (i *icon) fileNameNoExt() string {
	return "wahay-" + i.hash()
}

func (i *icon) hash() string {
	bytes := i.get()
	res := sha256.Sum256(bytes)
	return hex.EncodeToString(res[:])
}

func (i *icon) createPixBufWithSize(width, height int) (gdki.Pixbuf, error) {
	if i.g.gdk == nil {
		log.Error("createPixBufWithSize(): Graphics hasn't been initialized correctly")
	}

	pl, err := i.g.gdk.PixbufLoaderNew()
	if err != nil {
		return nil, err
	}

	complete := make(chan error)

	go func() {
		_, err = pl.Connect("area-prepared", func() {
			complete <- nil
		})
		if err != nil {
			log.WithFields(log.Fields{
				"caller": "pl.Connect(\"area-prepared\")",
			}).Errorf("createPixBufWithSize(): %s", err.Error())
		}

		_, err = pl.Connect("size-prepared", func() {
			pl.SetSize(width, height)
		})

		if err != nil {
			log.WithFields(log.Fields{
				"caller": "pl.Connect(\"size-prepared\")",
			}).Errorf("createPixBufWithSize(): %s", err.Error())
		}

		bytes := i.get()
		if _, err := pl.Write(bytes); err != nil {
			complete <- err
			return
		}

		if err := pl.Close(); err != nil {
			complete <- err
			return
		}
	}()

	err = <-complete

	if err != nil {
		return nil, err
	}

	return pl.GetPixbuf()
}

func (i *icon) createPixBuf() (gdki.Pixbuf, error) {
	return i.createPixBufWithSize(i.width, i.height)
}
