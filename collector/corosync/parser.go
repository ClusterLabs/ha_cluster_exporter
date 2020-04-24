package corosync

import (
	"regexp"
	"strconv"
	"strings"

	"github.com/pkg/errors"
)

type Parser interface {
	Parse(cfgToolOutput []byte, quorumToolOutput []byte) (*Status, error)
}

type Status struct {
	NodeId string
	RingId string
	Seq uint64
	Rings []Ring
	QuorumVotes QuorumVotes
	Quorate bool
}

type QuorumVotes struct {
	ExpectedVotes uint64
	HighestExpected uint64
	TotalVotes uint64
	Quorum uint64
}

type Ring struct {
	Number  string
	Address string
	Faulty  bool
}

func NewParser() Parser {
	return &defaultParser{}
}

type defaultParser struct{}

func (p *defaultParser) Parse(cfgToolOutput []byte, quorumToolOutput []byte) (*Status, error) {
	status := &Status{}
	var err error

	status.NodeId, err = parseNodeId(quorumToolOutput)
	if err != nil {
		return nil, errors.Wrap(err, "could not parse node id in corosync-quorumtool output")
	}

	status.RingId, status.Seq, err = parseRingIdAndSeq(quorumToolOutput)
	if err != nil {
		return nil, errors.Wrap(err, "could not parse ring id and seq number in corosync-quorumtool output")
	}

	status.Quorate, err = parseQuorate(quorumToolOutput)
	if err != nil {
		return nil, errors.Wrap(err, "could not parse quorate in corosync-quorumtool output")
	}

	status.QuorumVotes, err = parseQuoromVotes(quorumToolOutput)
	if err != nil {
		return nil, errors.Wrap(err, "could not parse quorum votes in corosync-quorumtool output")
	}

	status.Rings = parseRings(cfgToolOutput)

	return status, nil
}

func parseNodeId(cfgToolOutput []byte) (string, error) {
	nodeRe := regexp.MustCompile(`(?m)Node ID:\s+(\w+)`)
	matches := nodeRe.FindSubmatch(cfgToolOutput)
	if matches == nil {
		return "", errors.New("could not find Node ID line")
	}

	return string(matches[1]), nil
}

func parseRingIdAndSeq(cfgToolOutput []byte) (string, uint64, error) {
	nodeRe := regexp.MustCompile(`(?m)Ring ID:\s+(\w+)/(\d+)`)
	matches := nodeRe.FindSubmatch(cfgToolOutput)
	if matches == nil {
		return "", 0, errors.New("could not find Ring ID line")
	}

	seq, err := strconv.Atoi(string(matches[2]))
	if err != nil {
		return "", 0, errors.Wrap(err, "could not parse seq number to uint64")
	}

	return string(matches[1]), uint64(seq), nil
}

func parseQuorate(quorumToolOutput []byte) (bool, error) {
	re := regexp.MustCompile(`(?m)Quorate:\s+(Yes|No)`)
	matches := re.FindSubmatch(quorumToolOutput)
	if matches == nil {
		return false, errors.New("could not find Quorate line")
	}

	if string(matches[1]) == "Yes" {
		return true, nil
	}

	return false, nil
}

func parseRings(cfgToolOutput []byte) []Ring {
	re := regexp.MustCompile(`(?m)RING ID (?P<id>\d+)\s+id \s*= (?P<address>.+)\s+status \s*= (?P<status>.+)`)
	matches := re.FindAllSubmatch(cfgToolOutput, -1)
	rings := make([]Ring, len(matches))
	for i, match := range matches {
		namedMatches := extractRENamedCaptureGroups(re, match)

		rings[i] = Ring{
			Number:  namedMatches["id"],
			Address: namedMatches["address"],
			Faulty:  strings.Contains(namedMatches["status"], "FAULTY"),
		}
	}
	return rings
}

func parseQuoromVotes(quorumToolOutput []byte) (quorumVotes QuorumVotes, err error) {
	re := regexp.MustCompile(`(?m)Expected votes:\s+(\d+)\s+Highest expected:\s+(\d+)\s+Total votes:\s+(\d+)\s+Quorum:\s+(\d+)`)

	matches := re.FindSubmatch(quorumToolOutput)
	if matches == nil {
		return quorumVotes, errors.New("could not find quorum votes numbers")
	}

	quorumVotes.ExpectedVotes, err = strconv.ParseUint(string(matches[1]), 10, 64)
	if err != nil {
		return quorumVotes, errors.Wrap(err, "could not parse vote number to uint64")
	}

	quorumVotes.HighestExpected, err = strconv.ParseUint(string(matches[2]), 10, 64)
	if err != nil {
		return quorumVotes, errors.Wrap(err, "could not parse vote number to uint64")
	}

	quorumVotes.TotalVotes, err = strconv.ParseUint(string(matches[3]), 10, 64)
	if err != nil {
		return quorumVotes, errors.Wrap(err, "could not parse vote number to uint64")
	}

	quorumVotes.Quorum, err = strconv.ParseUint(string(matches[4]), 10, 64)
	if err != nil {
		return quorumVotes, errors.Wrap(err, "could not parse vote number to uint64")
	}

	return quorumVotes, nil
}

// extracts (?P<name>) RegEx capture groups from a match, to avoid numerical index lookups
func extractRENamedCaptureGroups(ringsRe *regexp.Regexp, match [][]byte) map[string]string {
	namedMatches := make(map[string]string)
	for i, name := range ringsRe.SubexpNames() {
		if i != 0 && name != "" {
			namedMatches[name] = string(match[i])
		}
	}
	return namedMatches
}
