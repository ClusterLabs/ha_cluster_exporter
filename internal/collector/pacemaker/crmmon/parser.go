package crmmon

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

type crmMonParser struct {
	crmMonPath string
	timeout    time.Duration
}

func (c *crmMonParser) Parse() (crmMon Root, err error) {
	ctx, cancel := context.WithTimeout(context.Background(), c.timeout)
	defer cancel()

	crmMonXML, err := exec.CommandContext(ctx, c.crmMonPath, "-X", "--inactive").Output()
	if err != nil {
		return crmMon, fmt.Errorf("error while executing crm_mon: %w", err)
	}

	err = xml.Unmarshal(crmMonXML, &crmMon)
	if err != nil {
		return crmMon, fmt.Errorf("error while parsing crm_mon XML output: %w", err)
	}

	return crmMon, nil
}

func NewCrmMonParser(crmMonPath string, timeout time.Duration) *crmMonParser {
	return &crmMonParser{crmMonPath, timeout}
}
