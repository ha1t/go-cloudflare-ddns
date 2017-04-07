package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"strings"

	"github.com/ha1t/go-php-function"
	"github.com/mattn/go-jsonpointer"
	"github.com/naoina/toml"
)

// URLで使うから全部string
type Record struct {
	RecId       string
	Name        string
	DisplayName string
	Ip_addr     string
	ServiceMode string
}

type tomlConfig struct {
	GlobalApiKey string
	Email        string
	Domain       string
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

func (record *Record) Update(after_ip_addr string) bool {

	record.Ip_addr = after_ip_addr

	config := loadConfig("config.toml")

	url := "https://www.cloudflare.com/api_json.html?a=rec_edit&"
	url += "tkn=" + config.GlobalApiKey
	url += "&id=" + record.RecId
	url += "&email=" + config.Email
	url += "&z=" + config.Domain + "&type=A&name=" + record.DisplayName + "&content=" + record.Ip_addr + "&service_mode=" + record.ServiceMode + "&ttl=1"

	result := php.File_get_contents(url)

	fmt.Printf("%v", result)

	return true
}

func main() {

	ip_addr := get_ip()

	// 取得したIPアドレスが前回と同じなら何もしない
	if ip_addr == pop_log() {
		//os.Exit(0)
	}

	push_log(ip_addr)

	config := loadConfig("config.toml")
	url := "https://www.cloudflare.com/api_json.html?a=rec_load_all&tkn=" + config.GlobalApiKey + "&email=" + config.Email + "&z=" + config.Domain

	records := get_dnslist(url)

	for _, record := range records {
		if record.Name == "direct.earth.project-p.jp" {
			fmt.Printf("%v", record)
			record.Update(ip_addr)
		}
	}
}

func get_dnslist(url string) []Record {
	test_json := php.File_get_contents(url)
	//fmt.Printf("%s", test_json)

	var obj interface{}
	json.Unmarshal([]byte(test_json), &obj)
	result, _ := jsonpointer.Get(obj, "/response/recs/objs")
	//fmt.Printf("%v\n", len(result.([]interface{})))
	//fmt.Printf("%v\n", result.([]interface{}))

	var records []Record

	for key, day := range result.([]interface{}) {
		_ = key

		if day == nil {
			continue
		}

		if day.(map[string]interface{})["type"] != "A" {
			continue
		}

		record := Record{}
		record.RecId = day.(map[string]interface{})["rec_id"].(string)
		record.Ip_addr = day.(map[string]interface{})["content"].(string)
		record.Name = day.(map[string]interface{})["name"].(string)
		record.ServiceMode = day.(map[string]interface{})["service_mode"].(string)
		record.DisplayName = day.(map[string]interface{})["display_name"].(string)

		records = append(records, record)
	}

	return records
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
