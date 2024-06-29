# 🌱 Essence
<p float="left">
  <img src="./static/essence.jpeg" width="225" height="250" />
  <img src="./static/miasma.jpeg" width="225" height="250" />
  <img src="./static/smoke.jpeg" width="225" height="250" />
</p>

Essence is a CLI tool that returns a list of unique domains or subdomains from an input list of strings, URLs or emails.



Isn't this the same as [anew](https://github.com/tomnomnom/anew) ? 

Yes sort of. Anew is a fantastic tool keeping a list of duplicate-free entries. Essence on the other hand, is an extractive tool meaning it attempts to parse out any possible domain from a list of URIs, email addresses, links, etc. 


## Installation

### Go
go install github.com/codycline/interrogator/cmd/essence@latest

sudo mv go/bin/interrogate /usr/local/bin

### From source
git clone https://github.com/codycline/essence cd interrogator `go build cmd/interrogate

## Examples 
1. Get unique domains from stdin: `cat urls.txt | essence`

1. Extract root domains from a targets nameservers: `dig ns +short example.com | essence`

1. Parse domains from a file explicitly: `essence urls.txt`

1. Include subdomains instead of root domains: `essence urls.txt --subdomains`


| 🎌 flag             | 📖 desc                                           | 📄 example                          | ⚙️ default                                                                                                                                                             |
| ------------------ | ------------------------------------------------ | ---------------------------------- | --------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| 🧾 json             | output as json                                   | `--json`                           | `false`                                                                                                                                                               |                                                   |
| 📜 output           | output results to a file                         | `--output results.txt`             | defaults to stdout  
| 🌐 subdomains           | output subdomains instead                         | `--subdomains`             | `false`  
| 📙 help           | CLI help                         | `--help`             | `false`  