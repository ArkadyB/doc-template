package docTemp

import (
	"bytes"
	"errors"
	"log"
	"path/filepath"
	"text/template"

	"github.com/arkadyb/doc-template/docx"
)

// Document interface is a combintation of methods use for generic data files
type Document interface {
	LoadFileFromBase64(b64 string) error
	ReadFile(string) error
	UpdateContent(string)
	GetContent() string
	WriteToFile(string, string) error
	WriteToBytes(*bytes.Buffer, string) error
	Close() error
}

// DocTemplate struct combines data and methods from both the Document interface
// and golang's templating library
type DocTemplate struct {
	Template *template.Template
	Document Document
}

func GetTemplateFromBase64(b64 string) (*DocTemplate, error) {
	document := new(docx.Docx)
	err := document.LoadFileFromBase64(b64)
	if err != nil {
		return nil, err
	}
	return &DocTemplate{Document: document, Template: template.New("docTemp")}, nil
}

// GetTemplate uses the file extension to determin the correct document struct to use
func GetTemplate(filePath string) (*DocTemplate, error) {
	var document Document
	switch filepath.Ext(filePath) {
	case ".docx":
		document = new(docx.Docx)
	default:
		return nil, errors.New("Unsupported Document Type")
	}

	err := document.ReadFile(filePath)
	if err != nil {
		return nil, err
	}
	return &DocTemplate{Document: document, Template: template.New("docTemp")}, nil
}

// Execute func runs the template and sends the output to the export path
func (docTemplate *DocTemplate) Execute(exportPath string, data interface{}) error {
	buf := new(bytes.Buffer)
	err := docTemplate.Template.Execute(buf, data)
	if err != nil {
		log.Println(err)
		return err
	}
	err = docTemplate.Document.WriteToFile(exportPath, buf.String())
	return err
}

func (docTemplate *DocTemplate) ExecuteToBytes(data interface{}) ([]byte, error) {
	buf := new(bytes.Buffer)
	defer buf.Reset()

	err := docTemplate.Template.Execute(buf, data)
	if err != nil {
		log.Println(err)
		return nil, err
	}
	out := new(bytes.Buffer)
	err = docTemplate.Document.WriteToBytes(out, buf.String())
	return out.Bytes(), err
}

// AddFunctions adds functions to the template
func (docTemplate *DocTemplate) AddFunctions(funcMap template.FuncMap) {
	docTemplate.Template = docTemplate.Template.Funcs(funcMap)
}

// Parse parses the template
func (docTemplate *DocTemplate) Parse() {
	temp, err := docTemplate.Template.Parse(docTemplate.Document.GetContent())
	if err != nil {
		log.Println(err)
	} else {
		docTemplate.Template = temp
	}
}
