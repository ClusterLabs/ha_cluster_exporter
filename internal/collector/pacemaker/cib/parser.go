package cib

import (
	"context"
	"encoding/xml"
	"fmt"
	"os/exec"
	"time"
)

type Parser interface {
	Parse() (Root, error)
}

type cibAdminParser struct {
	cibAdminPath string
	timeout      time.Duration
}

func (p *cibAdminParser) Parse() (Root, error) {
	var CIB Root
	ctx, cancel := context.WithTimeout(context.Background(), p.timeout)
	defer cancel()

	cibXML, err := exec.CommandContext(ctx, p.cibAdminPath, "--query", "--local").Output()
	if err != nil {
		return CIB, fmt.Errorf("error while executing cibadmin: %w", err)
	}

	err = xml.Unmarshal(cibXML, &CIB)
	if err != nil {
		return CIB, fmt.Errorf("could not parse cibadmin status from XML: %w", err)
	}

	return CIB, nil
}

func NewCibAdminParser(cibAdminPath string, timeout time.Duration) *cibAdminParser {
	return &cibAdminParser{cibAdminPath, timeout}
}
