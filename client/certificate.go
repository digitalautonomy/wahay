package client

import (
	// #nosec
	"crypto/sha1"
	"encoding/pem"
	"errors"
	"fmt"
	"io/ioutil"
	"net"
	"net/url"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/digitalautonomy/wahay/tor"
	"github.com/mxk/go-sqlite/sqlite3"

	log "github.com/sirupsen/logrus"
)

func (c *client) LoadCertificateFrom(
	serviceID string,
	servicePort int,
	cert []byte,
	webPort int) error {
	found := c.certificateFoundInDatabase(serviceID)
	if found {
		// certificate already in the Mumble database
		return nil
	}

	var err error
	var content string

	if cert == nil {
		u := &url.URL{
			Scheme: "http",
			Host:   net.JoinHostPort(serviceID, strconv.Itoa(webPort)),
		}

		content, err = tor.GetCurrentInstance().HTTPrequest(u.String())
		if err != nil {
			return err
		}

		cert = []byte(content)
	}

	err = c.storeCertificate(serviceID, servicePort, cert)
	if err != nil {
		return err
	}

	certContent := escapeByteString(arrayByteToString(cert))

	// TODO: should we maintain this?
	return c.saveCertificateConfigFile(certContent)
}

func (c *client) getDBConnection() (*sqlite3.Conn, error) {
	sqlFile := filepath.Join(filepath.Dir(c.configFile), ".mumble.sqlite")
	if !fileExists(sqlFile) {
		data := c.databaseProvider()
		err := ioutil.WriteFile(sqlFile, data, 0644)
		if err != nil {
			return nil, err
		}
	}

	conn, err := sqlite3.Open(sqlFile)
	if err != nil {
		return nil, err
	}

	return conn, nil
}

func (c *client) storeCertificate(serviceID string, servicePort int, cert []byte) error {
	block, _ := pem.Decode(cert)
	if block == nil || block.Type != "CERTIFICATE" {
		return errors.New("invalid certificate")
	}

	digest, err := getDigestForCert(block.Bytes)
	if err != nil {
		return err
	}

	log.WithFields(log.Fields{
		"hostname": serviceID,
		"port":     servicePort,
		"digest":   digest,
	}).Info("Loading certificate")

	params := sqlite3.NamedArgs{
		"$hostname": serviceID,
		"$port":     servicePort,
		"$digest":   digest,
	}

	conn, err := c.getDBConnection()
	if err != nil {
		return err
	}
	defer conn.Close()

	err = conn.Exec("REPLACE INTO `cert` (`hostname`,`port`,`digest`) VALUES ($hostname,$port,$digest)", params)
	if err != nil {
		return err
	}

	return nil
}

func (c *client) certificateFoundInDatabase(serviceID string) bool {
	conn, err := c.getDBConnection()
	if err != nil {
		return false
	}
	defer conn.Close()

	var numOfRecords int
	for q, err := conn.Query("SELECT COUNT(*) FROM cert WHERE hostname = ?", serviceID); err == nil; err = q.Next() {
		err = q.Scan(&numOfRecords)
		if err != nil {
			return false
		}
	}

	return numOfRecords > 0
}

func getDigestForCert(cert []byte) (string, error) {
	// #nosec
	h := sha1.New()
	_, err := h.Write(cert)
	if err != nil {
		return "", err
	}

	bs := h.Sum(nil)

	return fmt.Sprintf("%x", bs), nil
}

func arrayByteToString(data []byte) string {
	return fmt.Sprintf("@ByteArray(%s)", data)
}

// This function is inspired in QtSettings
// TODO: Review this code
func escapeByteString(str string) string {
	var result []string
	var needsQuotes bool

	startPos := len(str)
	escapeNextIfDigit := false

	for _, ch := range str {
		if ch == ';' || ch == ',' || ch == '=' {
			needsQuotes = true
		}

		if escapeNextIfDigit &&
			((ch >= '0' && ch <= '9') ||
				(ch >= 'a' && ch <= 'f') ||
				(ch >= 'A' && ch <= 'F')) {
			result = append(result, "\\x", strconv.FormatInt(int64(ch), 16))
			continue
		}

		if string(ch) == "golang\000" {
			result = append(result, "\\0")
			escapeNextIfDigit = true
			continue
		}

		result, escapeNextIfDigit = processUnicodeHelper(ch)
	}

	if needsQuotes ||
		(startPos < len(result) &&
			(result[0] == " " || result[len(result)-1] == " ")) {
		result = append([]string{"\""}, result...)
		result = append(result, "\"")
	}

	return strings.Join(result, "")
}

func processUnicodeHelper(ch rune) (result []string, escapeNextIfDigit bool) {
	switch ch {
	case '\a':
		result = append(result, "\\a")
	case '\b':
		result = append(result, "\\b")
	case '\f':
		result = append(result, "\\f")
	case '\n':
		result = append(result, "\\n")
	case '\r':
		result = append(result, "\\r")
	case '\t':
		result = append(result, "\\t")
	case '\v':
		result = append(result, "\\v")
	case '"':
	case '\\':
		result = append(result, "\\")
		result = append(result, string(ch))
	default:
		if ch <= 0x1F || ch >= 0x7F {
			result = append(result, "\\x", strconv.FormatInt(int64(ch), 16))
			escapeNextIfDigit = true
		} else {
			result = append(result, string(ch))
		}
	}

	return
}
