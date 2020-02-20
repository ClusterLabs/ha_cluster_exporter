package crmmon

import (
	"encoding/xml"
	"os/exec"

	"github.com/pkg/errors"
)

type Parser interface {
	Parse() (Root, error)
}

type crmMonParser struct {
	crmMonPath string
}

func (c *crmMonParser) Parse() (crmMon Root, err error) {
	crmMonXML, err := exec.Command(c.crmMonPath, "-X", "--group-by-node", "--inactive").Output()
	if err != nil {
		return crmMon, errors.Wrap(err, "error while executing crm_mon")
	}

	err = xml.Unmarshal(crmMonXML, &crmMon)
	if err != nil {
		return crmMon, errors.Wrap(err, "error while parsing crm_mon XML output")
	}

	return crmMon, nil
}

func NewCrmMonParser(crmMonPath string) *crmMonParser {
	return &crmMonParser{crmMonPath}
}

