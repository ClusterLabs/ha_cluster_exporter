package cfgtool

import (
	"regexp"
	"strings"

	"github.com/pkg/errors"
)

type Parser interface {
	Parse(rawStatus []byte) (*Status, error)
}

type defaultParser struct{}

func (p *defaultParser) Parse(rawStatus []byte) (*Status, error) {
	status := &Status{}

	nodeId, err := parseNodeId(rawStatus)
	if err != nil {
		return nil, errors.Wrap(err, "parser error")
	}
	status.NodeId = nodeId

	status.Rings = parseRings(rawStatus)

	return status, nil
}

func parseNodeId(rawStatus []byte) (string, error) {
	nodeRe := regexp.MustCompile(`(?m)Local node ID (.+)`)
	matches := nodeRe.FindSubmatch(rawStatus)
	if matches == nil {
		return "", errors.New("could not find node ID match")
	}
	nodeId := string(matches[1])
	return nodeId, nil
}

func parseRings(rawStatus []byte) []Ring {
	ringsRe := regexp.MustCompile(`(?m)RING ID (?P<id>\d+)\s+id \s*= (?P<address>.+)\s+status \s*= (?P<status>.+)`)
	matches := ringsRe.FindAllSubmatch(rawStatus, -1)
	rings := make([]Ring, len(matches))
	for i, match := range matches {
		matches := make(map[string]string)
		for i, name := range ringsRe.SubexpNames() {
			if i != 0 && name != "" {
				matches[name] = string(match[i])
			}
		}

		rings[i] = Ring{
			Id:      matches["id"],
			Address: matches["address"],
			Faulty:  strings.Contains(matches["status"], "FAULTY"),
		}
	}
	return rings
}

func NewParser() Parser {
	return &defaultParser{}
}
