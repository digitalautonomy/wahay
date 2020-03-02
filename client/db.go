package client

import (
	"bufio"
	"bytes"
	b "encoding/binary"
	"io/ioutil"
	"os"

	log "github.com/sirupsen/logrus"
)

type dbData struct {
	filename string
	content  []byte
}

func (d *dbData) exists(k string) bool {
	return bytes.Contains(d.content, []byte(k))
}

func (d *dbData) replaceString(find, replace string) {
	newContent := bytes.Replace(d.content, []byte(find), []byte(replace), -1)
	d.content = newContent
}

func (d *dbData) replaceInteger(find, replace int) {
	d.replaceString(
		intToStringToSearch(find),
		intToStringToSearch(replace),
	)
}

func (d *dbData) write() error {
	err := ioutil.WriteFile(d.filename, d.content, 0)
	if err != nil {
		return err
	}
	return nil
}

func intToStringToSearch(n int) string {
	n1 := n >> 8
	n2 := n % 256

	buf := new(bytes.Buffer)
	_ = b.Write(buf, b.LittleEndian, uint8(n1))
	_ = b.Write(buf, b.LittleEndian, uint8(n2))

	return buf.String()
}

func loadDBFromFile(filename string) (*dbData, error) {
	log.WithFields(log.Fields{
		"filepath": filename,
	}).Debug("Loading Mumble sqlite database")

	content, err := readBinaryContent(filename)
	if err != nil {
		return nil, err
	}

	d := &dbData{
		filename: filename,
		content:  content,
	}

	return d, nil
}

func readBinaryContent(filename string) ([]byte, error) {
	file, err := os.Open(filename)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	info, err := file.Stat()
	if err != nil {
		return nil, err
	}

	size := info.Size()
	bytes := make([]byte, size)

	buffer := bufio.NewReader(file)
	_, err = buffer.Read(bytes)
	if err != nil {
		return nil, err
	}

	return bytes, nil
}
