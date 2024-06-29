package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"net/url"
	"os"
	"regexp"
	"strings"

	flag "github.com/lynxsecurity/pflag"

	emailparser "github.com/mcnijman/go-emailaddress"
	log "github.com/sirupsen/logrus"
	"github.com/weppos/publicsuffix-go/publicsuffix"
)

var Version string

type Result struct {
	//Number of times seen in the list
	Seen int      `json:"seen"`
	From []string `json:"from"`
	//FQDN
	Domain string `json:"domain"`
}

type Results struct {
	Data              map[string]Result
	IncludeSubdomains bool
}

func main() {
	var inputs []string

	subdomains := flag.Bool("subdomains", false, "output subdomains instead")
	formatJSON := flag.Bool("json", false, "output as json")
	version := flag.Bool("version", false, "show version of essence")
	file := flag.String("output", "", "output results to a file")
	flag.Parse()
	if *version {
		fmt.Println(Version)
		log.Exit(0)
	}

	if usingSTDInput() {
		lines := detectSTDInput()
		inputs = append(inputs, lines...)
	}

	if len(os.Args) > 1 {
		args := detectArgInput(os.Args[1])
		inputs = append(inputs, args...)
	}

	if len(inputs) == 0 {
		log.Fatalf("no input detected exiting")
	}
	results := Results{
		Data:              map[string]Result{},
		IncludeSubdomains: *subdomains,
	}
	output := os.Stdout
	if *file != "" {
		var err error
		output, err = os.OpenFile(*file, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
		if err != nil {
			log.Fatalf("cannot create file for output at %s", *file)
		}
		defer output.Close()
	}

	for _, line := range inputs {
		results.parse(line)
	}

	for _, result := range results.Data {
		if *formatJSON {
			encoder := json.NewEncoder(output)
			encoder.Encode(&result)
		} else {
			output.WriteString(fmt.Sprintf("%s\n", result.Domain))
		}
	}
}

func detectArgInput(arg string) []string {
	var inputs []string

	if _, err := os.Stat(arg); err == nil {
		file, err := os.Open(arg)
		if err != nil {
			return inputs
		}
		defer file.Close()

		scanner := bufio.NewScanner(file)
		for scanner.Scan() {
			inputs = append(inputs, scanner.Text())
		}

		if err := scanner.Err(); err != nil {
			return inputs
		}

		return inputs
	}

	if strings.Contains(arg, ",") {
		args := strings.Split(arg, ",")
		return append(inputs, args...)
	}

	if arg != "" {
		return append(inputs, arg)
	}

	return inputs
}

func detectSTDInput() []string {
	var lines []string
	scanner := bufio.NewScanner(os.Stdin)
	for scanner.Scan() {
		lines = append(lines, scanner.Text())
	}

	if err := scanner.Err(); err != nil {
		return lines
	}
	return lines
}

func usingSTDInput() bool {
	stat, err := os.Stdin.Stat()
	if err != nil {
		return false
	}
	return (stat.Mode() & os.ModeCharDevice) == 0
}

func (r *Results) parse(line string) error {
	switch {
	//Likely a dns server such as ns1.example.com.
	case strings.HasSuffix(line, "."):
		r.upsert(strings.TrimSuffix(line, "."))
		return nil

	//Matches a protocol regex like https:// ftp://
	case regexp.MustCompile(`^\w+:\/\/`).MatchString(line):
		uri, err := url.Parse(line)
		if err != nil {
			return err
		}
		r.upsert(uri.Hostname())
		return nil
	case strings.Contains(line, "mailto:"):
		email, err := emailparser.Parse(strings.TrimPrefix(line, "mailto:"))
		if err != nil {
			return err
		}
		if email != nil {
			r.upsert(email.Domain)
		}
		return nil
	case strings.Contains(line, "@"):
		//parse out email address
		email, err := emailparser.Parse(line)
		if err != nil {
			return err
		}
		if email != nil {
			r.upsert(email.Domain)
		}
		return nil

	default:
		r.upsert(line)
	}
	return nil
}

func (r *Results) upsert(input string) error {

	var key string

	if r.IncludeSubdomains {
		parsed, err := publicsuffix.Parse(input)
		if err != nil {
			return err
		}
		key = parsed.String()
		// fmt.Println(key)
	} else {
		domain, err := publicsuffix.Domain(input)
		if err != nil {
			return err
		}
		key = domain
	}

	if domain, exists := (r.Data)[key]; exists {
		//Todo check if its already been seen by the current source then discard
		domain.From = append(domain.From, input)
		domain.Seen += 1
		(r.Data)[key] = domain
	} else {
		(r.Data)[key] = Result{
			Domain: key,
			From:   []string{input},
			Seen:   1,
		}
	}
	return nil
}
