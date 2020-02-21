package cib

import (
	"encoding/xml"
	"os/exec"

	"github.com/pkg/errors"
)

type Parser interface {
	Parse() (Root, error)
}

type cibAdminParser struct {
	cibAdminPath string
}

func (p *cibAdminParser) Parse() (Root, error) {
	var CIB Root
	cibXML, err := exec.Command(p.cibAdminPath, "--query", "--local").Output()
	if err != nil {
		return CIB, errors.Wrap(err, "error while executing cibadmin")
	}

	err = xml.Unmarshal(cibXML, &CIB)
	if err != nil {
		return CIB, errors.Wrap(err, "could not parse cibadmin status from XML")
	}

	return CIB, nil
}

func NewCibAdminParser(cibAdminPath string) *cibAdminParser {
	return &cibAdminParser{cibAdminPath}
}
