package main

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/url"
	"os"
	"regexp"
	"strings"

	"github.com/logrusorgru/aurora"
	flag "github.com/lynxsecurity/pflag"

	emailparser "github.com/mcnijman/go-emailaddress"
	log "github.com/sirupsen/logrus"
	"github.com/weppos/publicsuffix-go/publicsuffix"
)

var Version string

const banner = `                                          
    ██████████    
      ██  ██      
      ██  ██      
      ██  ██      
      ██  ██      
      ██  ██      
    ██  ░░  ██    
  ██          ██  
██        ░░    ██
██  ░░          ██
██░░░░░░░░░░░░░░██
██░░░░░░░░░░░░░░██
  ██░░░░░░░░░░██  
    ██████████    `

type Result struct {
	//Number of times seen in the list
	Seen int      `json:"times_seen"`
	From []string `json:"from"`
	//FQDN
	Domain string `json:"domain"`
}

type Results struct {
	Data              map[string]Result
	IncludeSubdomains bool
}

func main() {
	subdomains := flag.Bool("subdomains", false, "output subdomains instead")
	formatJSON := flag.Bool("json", false, "output as json")
	silent := flag.Bool("silent", false, "show no output")
	version := flag.Bool("version", false, "show version of essence")
	file := flag.String("output", "", "output results to a file")
	help := flag.Bool("help", false, "show help")

	flag.Parse()
	if *version {
		fmt.Println(Version)
		log.Exit(0)
	}

	if *silent {
		log.SetLevel(log.FatalLevel)
	}

	//Use this as an interface
	input, err := detectInput()
	if err != nil {
		log.Error(err)
		log.Fatal("an error occured while detecting input")
	}

	//Ensure file cleanup
	if file, ok := input.(*os.File); ok {
		defer file.Close()
	}

	if *help {
		fmt.Println(aurora.Red(fmt.Sprintf("%s\n", banner)))
		log.Infof("current essence version %s", Version)
		flag.Usage()
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

	// if usingSTDInput() {
	scanner := bufio.NewScanner(input)
	for scanner.Scan() {
		line := scanner.Text()
		parsed := results.parse(strings.TrimSpace(line))
		if parsed != "" {
			result, err := results.upsert(parsed)
			if err != nil {
				continue
			}
			if result != nil {

				if result.Seen == 1 && !*formatJSON {
					output.WriteString(fmt.Sprintf("%s\n", result.Domain))
				}
			}
		}
	}

	if err := scanner.Err(); err != nil {
		log.Error(err)
	}
	if *formatJSON {
		encoder := json.NewEncoder(output)
		for _, result := range results.Data {
			encoder.Encode(&result)
		}
	}
}

func (r *Results) parse(line string) string {

	switch {

	//Matches a protocol regex like https:// ftp://
	case regexp.MustCompile(`^\w+:\/\/`).MatchString(line):
		uri, err := url.Parse(line)
		if err != nil {
			return ""
		}
		return uri.Hostname()
	case strings.Contains(line, "mailto:"):
		email, err := emailparser.Parse(strings.TrimPrefix(line, "mailto:"))
		if err != nil {
			return ""
		}
		if email != nil {
			return email.Domain
		}
		return ""
	// Likely a dns server such as ns1.example.com.
	case strings.HasSuffix(line, "."):
		return strings.TrimSuffix(line, ".")
	case strings.Contains(line, "@"):
		//parse out email address
		email, err := emailparser.Parse(line)
		if err != nil {
			return ""
		}
		if email != nil {
			return email.Domain
		}
		return ""
	default:
		return line
	}
}

func (r *Results) upsert(input string) (*Result, error) {
	var key string

	if r.IncludeSubdomains {
		parsed, err := publicsuffix.Parse(input)
		if err != nil {
			return nil, err
		}
		key = parsed.String()
	} else {
		domain, err := publicsuffix.Domain(input)
		if err != nil {
			return nil, err
		}
		key = domain
	}

	if domain, exists := (r.Data)[key]; exists {
		//Todo check if its already been seen by the current source then discard
		domain.From = append(domain.From, input)
		domain.Seen += 1
		(r.Data)[key] = domain
		return &domain, nil
	} else {
		newResult := Result{
			Domain: key,
			From:   []string{input},
			Seen:   1,
		}
		(r.Data)[key] = newResult
		return &newResult, nil
	}
}

func detectInput() (io.ReadWriter, error) {
	input := bytes.NewBufferString("")
	if usingSTDInput() {
		return os.Stdin, nil
	}
	if len(os.Args) > 1 {
		arg := os.Args[1]
		if _, err := os.Stat(arg); err == nil {
			file, err := os.Open(arg)
			if err != nil {
				return nil, err
			}
			return file, nil
		}

		if strings.Contains(arg, ",") {
			args := strings.Split(arg, ",")
			for _, arg := range args {
				input.Write([]byte(fmt.Sprintf("%s\n", arg)))
				fmt.Println(input)
			}
			return input, nil
		}

		if arg != "" {
			input.Write([]byte(fmt.Sprintf("%s\n", arg)))
		}
	}

	return input, nil
}

func usingSTDInput() bool {
	stat, err := os.Stdin.Stat()
	if err != nil {
		return false
	}
	return (stat.Mode() & os.ModeCharDevice) == 0
}
