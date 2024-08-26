//
// Copyright (c) 2024 Markku Rossi
//
// All rights reserved.
//

package main

import (
	"flag"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"os"
	"slices"
	"strings"

	"github.com/clbanning/mxj"
)

var client = &http.Client{}

func main() {
	method := flag.String("method", "GET", "HTTP request method")
	printURLs := flag.Bool("print-urls", false, "print request URLs")
	flag.Parse()

	for _, arg := range flag.Args() {
		requests, err := NewRequests(arg)
		if err != nil {
			log.Fatalf("failed to parse requests file '%s': %s", arg, err)
		}
		for idx, url := range requests.URLs {
			if *printURLs {
				printURL(url)
			} else {
				headers := requests.Headers[*method]
				if headers == nil {
					log.Fatalf("no headers for method '%s'", *method)
				}
				err := do(*method, url, fmt.Sprintf("%s-%d", arg, idx), headers)
				if err != nil {
					log.Fatal(err)
				}
			}
		}
	}
}

func do(method, u, basename string, headers map[string]string) error {
	req, err := http.NewRequest(method, u, nil)
	if err != nil {
		return err
	}
	for k, v := range headers {
		req.Header.Add(k, v)
	}

	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	out, err := os.Create(fmt.Sprintf("%s-headers.txt", basename))
	if err != nil {
		return err
	}
	fmt.Fprintf(out, "%s %s\n", method, u)
	fmt.Fprintf(out, "=> %s\n", resp.Status)
	for k, values := range resp.Header {
		for _, v := range values {
			fmt.Fprintf(out, "%v: %v\n", k, v)
		}
	}
	out.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}
	suffix := "txt"
	ct := resp.Header.Get("Content-Type")
	if strings.HasSuffix(ct, "/xml") {
		suffix = "xml"
		body, err = mxj.BeautifyXml(body, "", "  ")
		if err != nil {
			return err
		}
	} else if strings.HasSuffix(ct, "/html") {
		suffix = "html"
	}

	resp.Body.Close()
	err = os.WriteFile(fmt.Sprintf("%s-body.%s", basename, suffix), body, 0666)
	if err != nil {
		return err
	}
	return nil
}

func printURL(raw string) {
	fmt.Printf("%s:\n", raw)
	u, err := url.Parse(raw)
	if err != nil {
		log.Fatal(err)
	}
	q := u.Query()
	var keys []string
	for k := range q {
		keys = append(keys, k)
	}
	slices.Sort(keys)
	for _, k := range keys {
		fmt.Printf(" - %s=%v\n", k, q[k])
	}
}
