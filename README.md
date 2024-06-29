# 🌱 Essence

Essence is a CLI tool that returns a list of unique domains or subdomains from a list of strings, URLs or emails.



Isn't this the same as anew ? 

Yes sort of.



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