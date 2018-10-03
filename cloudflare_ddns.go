package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"os"
	"strings"

	"github.com/cloudflare/cloudflare-go"
	"github.com/ha1t/go-php-function"
	"github.com/naoina/toml"
)

type tomlConfig struct {
	GlobalApiKey     string
	Email            string
	Domain           string
	TargetDomainList []string
	UseLineNotify    bool
	LineNotifyToken  string
}

func loadConfig(filename string) tomlConfig {

	f, err := os.Open(filename)
	if err != nil {
		panic(err)

	}
	defer f.Close()
	buf, err := ioutil.ReadAll(f)
	if err != nil {
		panic(err)
	}

	var config tomlConfig
	if err := toml.Unmarshal(buf, &config); err != nil {
		panic(err)
	}

	return config
}

func main() {

	if len(os.Args) != 2 {
		fmt.Println("configファイルが指定されていないか引数の指定が間違っています")
		os.Exit(1)
	}

	ip_addr := get_ip()

	// 取得したIPアドレスが前回と同じなら何もしない
	if ip_addr == pop_log() {
		os.Exit(0)
	}

	config := loadConfig(os.Args[1])

	// Construct a new API object
	api, err := cloudflare.New(config.GlobalApiKey, config.Email)
	if err != nil {
		log.Fatal(err)
	}

	// Fetch the zone ID
	zoneID, err := api.ZoneIDByName(config.Domain) // Assuming example.com exists in your Cloudflare account already
	if err != nil {
		log.Fatal(err)
	}

	// Fetch all DNS records for example.org
	records, err := api.DNSRecords(zoneID, cloudflare.DNSRecord{})
	if err != nil {
		log.Fatal(err)
		return
	}

	// r.Content が IP
	for _, record := range records {
		for _, target_domain := range config.TargetDomainList {
			if record.Name == target_domain {
				record.Content = ip_addr
				api.UpdateDNSRecord(record.ZoneID, record.ID, record)
				if err != nil {
					log.Fatal(err)
					return
				}
				if config.UseLineNotify {
					err := notifyLine(config.LineNotifyToken, "update:"+record.Name)
					if err != nil {
						fmt.Printf("%s", err)
					}
				}
			}
		}
	}

	// 全部成功してから書き込む
	push_log(ip_addr)
}

func get_ip() string {
	url := "http://ipv4.icanhazip.com/"
	return strings.TrimSpace(php.File_get_contents(url))
}

func pop_log() string {
	filename := "ip_addr.log"
	if php.FileExists(filename) == false {
		php.File_put_contents(filename, "")
	}
	return strings.TrimSpace(php.File_get_contents(filename))
}

func push_log(ip_addr string) {
	filename := "ip_addr.log"
	php.File_put_contents(filename, ip_addr)
}

func notifyLine(token string, message string) error {
	api_url := "https://notify-api.line.me/api/notify"

	values := url.Values{}
	values.Add("message", message)

	req, err := http.NewRequest("POST", api_url, strings.NewReader(values.Encode()))
	if err != nil {
		return err
	}

	// Content-Type 設定
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Add("Content-Type", "application/x-www-form-urlencoded")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return err
	}

	defer resp.Body.Close()

	return err
}
