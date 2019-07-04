package main

import (
	"bytes"
	"encoding/json"
	"encoding/xml"
	"fmt"
	"io/ioutil"
	"net/url"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/ddliu/go-httpclient"
	"github.com/rs/zerolog"
	"gopkg.in/alecthomas/kingpin.v2"
)

type cdn77response struct {
	Status      string `json:"status"`
	Description string `json:"description"`
}

type cdn77resourcelist struct {
	Status      string `json:"status"`
	Description string `json:"description"`
	CdnResource []struct {
		Id    int    `json:"id"`
		CName string `json:"cname"`
	} `json:"cdnResources"`
}

type Url struct {
	XMLName xml.Name `xml:"url"`
	Loc     string   `xml:"loc"`
}

type UrlSet struct {
	XMLName xml.Name `xml:"urlset"`
	Urls    []Url    `xml:"url"`
}

var log zerolog.Logger

var (
	login   = kingpin.Flag("login", "Your login (email) to CDN77 control panel").Required().String()
	token   = kingpin.Flag("token", "Your API Token, needs to be generated in the profile section on client.CDN77.com").Required().String()
	site    = kingpin.Flag("site", "Your website aka 'CDN Resource' in CDN77").Required().String()
	sitemap = kingpin.Flag("sitemap", "sitemap.xml file OR url begining with http:// or https://").String()
	purge   = kingpin.Flag("purge-all", "remove (purge) existing HTTP content on CDN77").Bool()
	verbose = kingpin.Flag("verbose", "verbose output").Bool()
)

func main() {
	output := zerolog.ConsoleWriter{Out: os.Stdout, TimeFormat: time.RFC3339}
	output.FormatLevel = func(i interface{}) string {
		return strings.ToUpper(fmt.Sprintf("| %-6s|", i))
	}
	log = zerolog.New(output).With().Timestamp().Logger()

	kingpin.Parse()

	params := map[string]string{
		"login":  *login,
		"passwd": *token,
	}

	resourcelist := getResourceList(params)

	params["cdn_id"] = getCdnId(site, resourcelist)

	if *purge {
		purgeAll(params)
	}

	prefetchUrls(urlsFromSitemap(sitemap), params)

	log.Info().Msg("End.")
	os.Exit(0x0)
}

func getResourceList(params map[string]string) cdn77resourcelist {
	var messages bytes.Buffer
	var resourcelist cdn77resourcelist

	messages.WriteString("Reading resource list from CDN77 ... ")

	response, err := httpclient.Get("https://api.cdn77.com/v2.0/cdn-resource/list", params)
	if err != nil {
		messages.WriteString(err.Error())
		log.Error().Msg(messages.String())
		os.Exit(0xe1)
	}

	responseBody, err := response.ReadAll()
	if err != nil {
		messages.WriteString(err.Error())
		log.Error().Msg(messages.String())
		os.Exit(0xe2)
	}

	err = json.Unmarshal(responseBody, &resourcelist)
	if err != nil {
		messages.WriteString(err.Error())
		log.Error().Msg(messages.String())
		os.Exit(0xe3)
	}

	if resourcelist.Status != "ok" {
		messages.WriteString(resourcelist.Status)
		messages.WriteString(": ")
		messages.WriteString(resourcelist.Description)
		log.Error().Msg(messages.String())
		os.Exit(0xf0)
	}

	messages.WriteString("ok")
	log.Info().Msg(messages.String())

	return resourcelist
}

func getCdnId(site *string, resourcelist cdn77resourcelist) string {
	var messages bytes.Buffer
	var cdn_id string

	messages.WriteString("Searching for ")
	messages.WriteString(*site)
	messages.WriteString(" ... ")

	for _, cdnres := range resourcelist.CdnResource {
		if strings.ToLower(cdnres.CName) == strings.ToLower(*site) {
			cdn_id = strconv.Itoa(cdnres.Id)
			break
		}
	}

	if cdn_id == "" {
		messages.WriteString("not found, aborting")
		log.Error().Msg(messages.String())
		os.Exit(0xa0)
	}

	messages.WriteString("ok (")
	messages.WriteString("resource id #")
	messages.WriteString(cdn_id)
	messages.WriteString(")")
	log.Info().Msg(messages.String())

	return cdn_id
}

func purgeAll(params map[string]string) {
	var messages bytes.Buffer

	messages.WriteString("Starting 'purge-all' ... ")
	r, err := httpclient.Post("https://api.cdn77.com/v2.0/data/purge-all", params)
	if err != nil {
		messages.WriteString(err.Error())
		messages.WriteString(", aborting")
		log.Error().Msg(messages.String())
		os.Exit(0xb0)
	}
	b, _ := r.ReadAll()
	var cdn77r cdn77response
	if err = json.Unmarshal(b, &cdn77r); err != nil {
		messages.WriteString(err.Error())
		messages.WriteString(", aborting")
		log.Error().Msg(messages.String())
		os.Exit(0xb1)
	}
	if cdn77r.Status != "ok" {
		messages.WriteString(cdn77r.Status)
		messages.WriteString(": ")
		messages.WriteString(cdn77r.Description)
		messages.WriteString(", aborting")
		log.Error().Msg(messages.String())
		os.Exit(0xb2)
	}

	messages.WriteString("ok")
	if *verbose {
		messages.WriteString(" (")
		messages.WriteString(cdn77r.Description)
		messages.WriteString(")")
	}
	log.Info().Msg(messages.String())
}

func urlsFromSitemap(filename *string) []string {
	var messages bytes.Buffer
	var byteValue []byte
	var err error

	messages.WriteString("Reading '")
	messages.WriteString(*filename)
	messages.WriteString("' ... ")

	if strings.HasPrefix(strings.TrimSpace(strings.ToLower(*filename)), "http://") ||
		strings.HasPrefix(strings.TrimSpace(strings.ToLower(*filename)), "https://") {
		r, err := httpclient.Get(strings.TrimSpace(*filename))
		if err != nil {
			messages.WriteString(err.Error())
			messages.WriteString(", aborting")
			log.Error().Msg(messages.String())
			os.Exit(0xc0)
		}

		byteValue, err = r.ReadAll()

	} else {
		f, err := os.Open(*filename)
		if err != nil {
			messages.WriteString(err.Error())
			messages.WriteString(", aborting")
			log.Error().Msg(messages.String())
			os.Exit(0xc0)
		}

		byteValue, err = ioutil.ReadAll(f)
	}
	if err != nil {
		messages.WriteString(err.Error())
		messages.WriteString(", aborting")
		log.Error().Msg(messages.String())
		os.Exit(0xc1)
	}

	var urlset UrlSet
	err = xml.Unmarshal(byteValue, &urlset)
	if err != nil {
		messages.WriteString(err.Error())
		messages.WriteString(", aborting")
		log.Error().Msg(messages.String())
		os.Exit(0xc2)
	}

	messages.WriteString("ok")
	log.Info().Msg(messages.String())

	s := []string{}
	for _, u := range urlset.Urls {
		s = append(s, u.Loc)
	}

	return s
}

func prefetchUrls(urls []string, params map[string]string) {
	var messages bytes.Buffer
	var cdn77r cdn77response

	messages.WriteString("Prefetching ")

	if *verbose {
		for i, u := range urls {
			if i > 0 {
				messages.WriteString(", ")
			}
			messages.WriteString("'")
			messages.WriteString(u)
			messages.WriteString("'")
		}
	}
	messages.WriteString(" ... ")

	response, err := httpclient.Post("https://api.cdn77.com/v2.0/data/prefetch", url.Values{
		"cdn_id": []string{params["cdn_id"]},
		"login":  []string{params["login"]},
		"passwd": []string{params["passwd"]},
		"url[]":  urls,
	})
	if err != nil {
		messages.WriteString(err.Error())
		log.Error().Msg(messages.String())
		os.Exit(0xd0)
	}
	responseBody, _ := response.ReadAll()

	err = json.Unmarshal(responseBody, &cdn77r)
	if err != nil {
		messages.WriteString(err.Error())
		log.Error().Msg(messages.String())
		os.Exit(0xd1)
	}
	if cdn77r.Status != "ok" {
		messages.WriteString(cdn77r.Status)
		messages.WriteString(" (")
		messages.WriteString(cdn77r.Description)
		messages.WriteString(")")
		log.Error().Msg(messages.String())
		os.Exit(0xd2)
	}

	messages.WriteString("ok")
	if *verbose {
		messages.WriteString(" (")
		messages.WriteString(cdn77r.Description)
		messages.WriteString(")")
	}
	log.Info().Msg(messages.String())

	messages.Reset()
}
