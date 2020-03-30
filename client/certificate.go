package client

import (
	// #nosec
	"crypto/sha1"
	"encoding/pem"
	"errors"
	"fmt"
	"net"
	"net/url"
	"strconv"
	"strings"

	"github.com/digitalautonomy/wahay/tor"
	log "github.com/sirupsen/logrus"
)

const certServerPort = 8181

func (c *client) requestCertificate(address string) error {
	hostname, port, err := extractHostAndPort(address)
	if err != nil {
		return errors.New("invalid certificate url")
	}

	u := &url.URL{
		Scheme: "http",
		Host:   net.JoinHostPort(hostname, strconv.Itoa(certServerPort)),
	}

	content, err := tor.CurrentInstance().HTTPrequest(u.String())
	if err != nil {
		return err
	}

	cert := []byte(content)
	p, _ := strconv.Atoi(port)
	err = c.storeCertificate(hostname, p, cert)
	if err != nil {
		return err
	}

	// TODO: Should we maintain this?
	return c.saveCertificateConfigFile(cert)
}

func extractHostAndPort(address string) (host string, port string, err error) {
	u, err := url.Parse(address)
	if err != nil {
		return
	}

	host, port, err = net.SplitHostPort(u.Host)
	if err != nil {
		return
	}

	return host, port, nil
}

func (c *client) storeCertificate(hostname string, port int, cert []byte) error {
	if c.isTheCertificateInDB(hostname) {
		return nil
	}

	block, _ := pem.Decode(cert)
	if block == nil || block.Type != "CERTIFICATE" {
		return errors.New("invalid certificate")
	}

	digest, err := digestForCertificate(block.Bytes)
	if err != nil {
		return err
	}

	log.WithFields(log.Fields{
		"hostname": hostname,
		"port":     port,
		"digest":   digest,
	}).Info("Storing Mumble client certificate")

	return c.storeCertificateInDB(hostname, port, digest)
}

const (
	defaultHostToReplace   = "ffaaffaabbddaabbddeeaaddccaaffeebbaabbeeddeeaaddbbeeeeff.onion"
	defaultPortToReplace   = 64738
	defaultDigestToReplace = "AAABACADAFBABBBCBDBEBFCACBCCCDCECFDADBDC"
)

func (c *client) storeCertificateInDB(id string, port int, digest string) error {
	db, err := c.db()
	if err != nil {
		return err
	}

	log.WithFields(log.Fields{
		"defaultHost":   defaultHostToReplace,
		"defaultPort":   defaultPortToReplace,
		"defaultDigest": defaultDigestToReplace,
		"newHost":       id,
		"newPort":       port,
		"newDigest":     digest,
	}).Debug("Replacing content in Mumble sqlite database")

	db.replaceString(defaultHostToReplace, id)
	db.replaceString(defaultDigestToReplace, digest)
	db.replaceInteger(uint16(defaultPortToReplace), uint16(port))

	return db.write()
}

func (c *client) isTheCertificateInDB(hostname string) bool {
	d, err := c.db()
	if err != nil {
		return false
	}

	return d.exists(hostname)
}

func digestForCertificate(cert []byte) (string, error) {
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
