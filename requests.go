//
// Copyright (c) 2024 Markku Rossi
//
// All rights reserved.
//

package main

import (
	"bufio"
	"fmt"
	"net/url"
	"os"
	"regexp"
	"strings"
)

var (
	reSection = regexp.MustCompilePOSIX(`^(\*+)[[:space:]]+(.*)`)
	reHeader  = regexp.MustCompilePOSIX(`^([^:]+):[[:space:]]*(.*)`)
)

// Requests defines HTTP request URLs and options.
type Requests struct {
	scanner   *bufio.Scanner
	unget     bool
	ungetData string

	URLs    []string
	Headers map[string]map[string]string
}

// NewRequests reads the HTTP request information from the file.
func NewRequests(file string) (*Requests, error) {
	reqs := &Requests{
		Headers: make(map[string]map[string]string),
	}
	err := reqs.parse(file)
	if err != nil {
		return nil, err
	}
	return reqs, nil
}

func (r *Requests) getLine() (string, bool) {
	if r.unget {
		r.unget = false
		return r.ungetData, true
	}
	if r.scanner.Scan() {
		return strings.TrimSpace(r.scanner.Text()), true
	}
	return "", false
}

func (r *Requests) ungetLine(line string) {
	r.unget = true
	r.ungetData = line
}

func (r *Requests) parse(file string) error {
	in, err := os.Open(file)
	if err != nil {
		return err
	}
	defer in.Close()

	r.scanner = bufio.NewScanner(in)
	for {
		line, ok := r.getLine()
		if !ok {
			break
		}
		if len(line) == 0 {
			continue
		}
		if strings.HasPrefix(line, "-*-") {
			continue
		}
		if strings.HasPrefix(line, "* Headers") {
			err = r.parseHeaders()
			if err != nil {
				return err
			}
		} else if strings.HasPrefix(line, "* URLs") {
			err = r.parseURLs()
			if err != nil {
				return err
			}
		} else {
			return fmt.Errorf("invalid line: '%s'", line)
		}
	}

	return nil
}

func (r *Requests) parseHeaders() error {
	method := "GET"

	for {
		line, ok := r.getLine()
		if !ok {
			return nil
		}
		if len(line) == 0 {
			continue
		}
		m := reSection.FindStringSubmatch(line)
		if m != nil {
			if len(m[1]) == 1 {
				r.ungetLine(line)
				return nil
			}
			fmt.Printf("%s %s\n", m[1], m[2])
			method = m[2]
			continue
		}
		m = reHeader.FindStringSubmatch(line)
		if m != nil {
			fmt.Printf("%s: %s\n", m[1], m[2])
			c, ok := r.Headers[method]
			if !ok {
				c = make(map[string]string)
				r.Headers[method] = c
			}
			c[m[1]] = m[2]
			continue
		}
		return fmt.Errorf("invalid line: %s", line)
	}
}

func (r *Requests) parseURLs() error {
	for {
		line, ok := r.getLine()
		if !ok {
			return nil
		}
		if len(line) == 0 {
			continue
		}
		_, err := url.Parse(line)
		if err != nil {
			return err
		}
		r.URLs = append(r.URLs, line)
	}
}
