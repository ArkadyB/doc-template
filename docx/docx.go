package docx

import (
	"archive/zip"
	"bufio"
	"encoding/base64"
	"errors"
	"io"
	"log"
	"os"

	"bytes"
)

// Docx struct that contains data from a docx
type Docx struct {
	zipReader *zip.ReadCloser
	content   string
}

func (d *Docx) LoadFileFromBase64(b64 string) error {
	dec, err := base64.StdEncoding.DecodeString(b64)
	if err != nil {
		return err
	}

	r := bytes.NewReader(dec)
	reader, err := zip.NewReader(r, int64(len(dec)))
	if err != nil {
		log.Println(err.Error())
		return errors.New("Cannot Open File")
	}
	content, err := readText(reader.File)
	if err != nil {
		log.Println(err.Error())
		return errors.New("Cannot Read File")
	}

	closer := new(zip.ReadCloser)
	closer.File = reader.File
	d.zipReader = closer

	if content == "" {
		return errors.New("File has no content")
	}
	d.content = cleanText(content)
	return nil
}

// ReadFile func reads a docx file
func (d *Docx) ReadFile(path string) error {
	reader, err := zip.OpenReader(path)
	if err != nil {
		log.Println(err.Error())
		return errors.New("Cannot Open File")
	}
	content, err := readText(reader.File)
	if err != nil {
		log.Println(err.Error())
		return errors.New("Cannot Read File")
	}
	d.zipReader = reader
	if content == "" {
		return errors.New("File has no content")
	}
	d.content = cleanText(content)
	log.Printf("Read File `%s`", path)
	return nil
}

// UpdateContent updates the content string
func (d *Docx) UpdateContent(newContent string) {
	d.content = newContent
}

// GetContent returns the string content
func (d *Docx) GetContent() string {
	return d.content
}

// WriteToFile writes the changes to a new file
func (d *Docx) WriteToFile(path string, data string) error {
	var target *os.File
	target, err := os.Create(path)
	if err != nil {
		return err
	}
	defer target.Close()
	err = d.write(target, data)
	if err != nil {
		return err
	}
	return nil
}

func (d *Docx) WriteToBytes(buf *bytes.Buffer, data string) error {
	target := bufio.NewWriter(buf)
	err := d.write(target, data)
	if err != nil {
		return err
	}
	return nil
}

// Close the document
func (d *Docx) Close() error {
	return d.zipReader.Close()
}

func (d *Docx) write(ioWriter io.Writer, data string) error {
	var err error
	// Reformat string, for some reason the first char is converted to &lt;
	w := zip.NewWriter(ioWriter)
	for _, file := range d.zipReader.File {
		var writer io.Writer
		var readCloser io.ReadCloser
		writer, err := w.Create(file.Name)
		if err != nil {
			return err
		}
		readCloser, err = file.Open()
		if err != nil {
			return err
		}
		if file.Name == "word/document.xml" {
			writer.Write([]byte(data))
		} else {
			writer.Write(streamToByte(readCloser))
		}
	}
	w.Close()
	return err
}
