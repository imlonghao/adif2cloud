package adif

import (
	"bytes"
	"strings"
	"text/template"

	"github.com/Matir/adifparser"
)

func Parse(adif string) map[string]string {
	result := make(map[string]string)
	result["raw"] = adif
	adifReader := adifparser.NewADIFReader(strings.NewReader(adif))
	record, err := adifReader.ReadRecord()
	if err != nil {
		return result
	}
	fields := record.GetFields()
	for _, field := range fields {
		value, err := record.GetValue(field)
		if err != nil {
			continue
		}
		result[field] = value
	}
	return result
}

func FillTemplate(t string, data map[string]string) (string, error) {
	tmpl, err := template.New("body").Parse(t)
	if err != nil {
		return "", err
	}
	var bodyBuffer bytes.Buffer
	if err := tmpl.Execute(&bodyBuffer, data); err != nil {
		return "", err
	}
	return bodyBuffer.String(), nil
}
