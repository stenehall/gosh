package favii

/*
https://git.dcpri.me/modules/favii
Copyright 2020 Darshil Chanpura
Permission is hereby granted, free of charge, to any person obtaining a copy of this software and associated
documentation files (the "Software"), to deal in the Software without restriction, including without
limitation the rights to use, copy, modify, merge, publish, distribute, sublicense, and/or sell copies of the Software,
and to permit persons to whom the Software is furnished to do so, subject to the following conditions:
The above copyright notice and this permission notice shall be included in all copies or substantial portions of the
Software.
THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR IMPLIED, INCLUDING BUT NOT LIMITED TO
THE WARRANTIES OF MERCHANTABILITY, FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE AUTHORS
OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR
OTHERWISE, ARISING FROM, OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE SOFTWARE.
*/

import (
	"bytes"
	"crypto/tls"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"strconv"
	"strings"

	"golang.org/x/net/html"
)

// Favii is basically a client to use the struct. It has a http.Client for
// doing all HTTP Requests for fetching the HTML Pages.
type Favii struct {
	client   *http.Client
	cache    map[string]*MetaInfo
	useCache bool
}

// MetaInfo with metadata details, this includes the URL used for requesting
// or calling GetMetaInfo() as well.
type MetaInfo struct {
	URL             *url.URL
	Manifest        Link
	ThemeColor      string
	BackgroundColor string
	Metas           []Meta
	Links           []Link
}

// Manifest contains data obtained from the manifest file if found on the site.
type Manifest struct {
	Name            string
	BackgroundColor string
	ThemeColor      string
	IconURL         string
}

type jsonManifest struct {
	Name  string `json:"name"`
	Icons []struct {
		Src   string `json:"src"`
		Sizes string `json:"sizes"`
		Type  string `json:"type"`
	} `json:"icons"`
	ThemeColor      string `json:"theme_color"`
	BackgroundColor string `json:"background_color"`
}

// Meta is a simple struct to keep name and content attributes of an HTML Page
// mostly contains the details about meta tag.
type Meta struct {
	Name    string
	Content string
}

// Link is a simple struct to keep rel and href attributes of an HTML Page
// mostly contains the details about link tag.
type Link struct {
	Rel  string
	Href string
}

// New creates a new Favii struct with http.DefaultClient and empty map, also
// an optional cache map.
func New(useCache bool) *Favii {
	tr := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}

	return &Favii{
		client:   &http.Client{Transport: tr},
		cache:    map[string]*MetaInfo{},
		useCache: useCache,
	}
}

// GetMetaInfo for getting meta information, it is mainly a wrapper around
// unexported method getMetaInfo().
func (f *Favii) GetMetaInfo(url string) (*MetaInfo, error) {
	m, err := f.getMetaInfo(url)
	if err != nil {
		return nil, err
	}
	return m, nil
}

// GetManifest from the MetaInfo, using link tags.
func (m *MetaInfo) GetManifest() (Manifest, error) {
	if m == nil {
		return Manifest{}, fmt.Errorf("no manifest found")
	}
	if m.Manifest.Href == "" {
		return Manifest{}, fmt.Errorf("empty manifest url")
	}
	var jsonManifest jsonManifest

	var err error
	if strings.HasPrefix(m.Manifest.Href, "data:application/json;base64,") {
		sDec, err := base64.StdEncoding.DecodeString(strings.ReplaceAll(m.Manifest.Href, "data:application/json;base64,", ""))
		if err != nil {
			return Manifest{}, fmt.Errorf("couldn't decode manifest base64 string: %w", err)
		}
		if err := json.Unmarshal(sDec, &jsonManifest); err != nil {
			return Manifest{}, fmt.Errorf("failed unmarshal base64 manifest: %w", err)
		}
	} else {
		// @TODO HTTP get manifest file
		manifestURL := createURL(m.Manifest.Href, *m.URL)
		response, err := http.Get(manifestURL)
		if err != nil {
			return Manifest{}, fmt.Errorf("failed to fetch the manifest: %w", err)
		}
		// nolint
		defer func(Body io.ReadCloser) {
			err := Body.Close()
			if err != nil {
				err = fmt.Errorf("couldn't close the file download. The file might be damaged: %w", err)
			}
		}(response.Body)
		if err := json.NewDecoder(response.Body).Decode(&jsonManifest); err != nil {
			return Manifest{}, fmt.Errorf("failed to unmarshal the fetched manifest: %w", err)
		}
	}

	if err != nil {
		return Manifest{}, fmt.Errorf("couldn't parse manifest file %w", err)
	}

	// Find the largest favicon
	iconMaxIndex := 0
	iconMaxSize := 0
	for iconIndex, icon := range jsonManifest.Icons {
		sizeParts := strings.Split(icon.Sizes, "x")
		iconSize, err := strconv.Atoi(sizeParts[0])
		// If we fail to handle one icon it's ok, we just try the next
		if err != nil {
			continue
		}

		if iconSize > iconMaxSize {
			iconMaxSize = iconSize
			iconMaxIndex = iconIndex
		}
	}

	manifestIconURL := jsonManifest.Icons[iconMaxIndex].Src
	baseURL := m.URL

	// @TODO clean up handling of "./" in manifest icon urls.
	//		 write some tests for this, since it's an interesting case.
	if strings.HasPrefix(manifestIconURL, "./") {
		baseURL, _ = url.Parse(m.Manifest.Href)
		manifestIconURL = strings.TrimSuffix(baseURL.Path, "/manifest.json") + strings.TrimPrefix(manifestIconURL, ".")
	}

	return Manifest{
		jsonManifest.Name,
		jsonManifest.BackgroundColor,
		jsonManifest.ThemeColor,
		createURL(manifestIconURL, *baseURL),
	}, nil
}

// GetFaviconURL for getting favicon URL from the MetaInfo, using link tags,
// or use default /favicon.ico.
func (m *MetaInfo) GetFaviconURL() string {
	faviconURLs := [2]string{"", ""}

	if m == nil && m.Links == nil {
		return ""
	}

	for index := range m.Links {
		// we've found a (most likely) high-res image stop iterating
		if strings.HasPrefix(m.Links[index].Rel, "apple-touch-icon") {
			faviconURLs[0] = m.Links[index].Href
			break
		}
		if strings.HasPrefix(m.Links[index].Rel, "apple-touch-icon-precomposed") {
			faviconURLs[0] = m.Links[index].Href
			break
		}
		// this is likely a low-res image so keep track of it but continue to look
		if m.Links[index].Rel == "icon" || m.Links[index].Rel == "shortcut icon" {
			faviconURLs[1] = m.Links[index].Href
		}
	}

	var filteredLink string
	if faviconURLs[1] != "" {
		filteredLink = faviconURLs[1]
	}

	if faviconURLs[0] != "" {
		filteredLink = faviconURLs[0]
	}

	// we found an image, use it to create the actual url
	if filteredLink != "" {
		return createURL(filteredLink, *m.URL)
	}

	// in case if nothing is available go for the default one.
	return m.URL.Scheme + "://" + m.URL.Hostname() + "/favicon.ico"
}

func (f *Favii) getMetaInfo(u string) (*MetaInfo, error) {
	parsedURL, err := url.Parse(u)
	if err != nil {
		return nil, fmt.Errorf("failed to parse the url: %w", err)
	}
	if m, ok := f.cache[parsedURL.Hostname()]; ok && f.useCache { // skip this if useCache is false
		return m, nil
	}
	m := &MetaInfo{
		Metas: []Meta{},
		Links: []Link{},
		URL:   parsedURL,
	}
	defer func(m *MetaInfo) {
		f.cache[m.URL.Hostname()] = m
	}(m)

	response, err := f.client.Get(u)
	if err != nil {
		return nil, fmt.Errorf("failed to request the meta information: %w", err)
	}
	defer response.Body.Close()

	m.URL = response.Request.URL
	bodyBytes, _ := ioutil.ReadAll(response.Body)
	t := html.NewTokenizer(response.Body)

	if metas, err := getMeta(*t); err == nil {
		m.Metas = metas
	}

	// http://blog.manugarri.com/how-to-reuse-http-response-body-in-golang/
	// reset the response body to the original unread state
	response.Body = ioutil.NopCloser(bytes.NewBuffer(bodyBytes))
	t = html.NewTokenizer(response.Body)

	if links, err := getLinks(*t); err == nil {
		m.Links = links
	}

	for _, lk := range m.Links {
		if lk.Rel == "manifest" {
			m.Manifest = lk
		} else {
			m.Links = append(m.Links, lk)
		}
	}

	for _, mt := range m.Metas {
		if mt.Name == "theme-color" {
			m.ThemeColor = mt.Content
		}
		if mt.Name == "background-color" {
			m.BackgroundColor = mt.Content
		}
	}

	return m, nil
}

func getMeta(t html.Tokenizer) ([]Meta, error) {
	var metas []Meta

	for {
		tt := t.Next()
		if tt == html.ErrorToken {
			if errors.Is(t.Err(), io.EOF) {
				break
			}
			log.Printf("Error: %v", t.Err())
			break
		}

		// If the tag is a self-closing tag or the start tag we will search it.
		// Otherwise continue to the next string. There's no point in looking at other tags
		if tt != html.SelfClosingTagToken && tt != html.TextToken && tt != html.StartTagToken {
			// fmt.Println("Skipping:", string(tagName))
			continue
		}

		// The tag have no attributes, so they're of no interest for us, for example a <style> or <script> tag
		tagName, hasAttr := t.TagName()

		if !hasAttr {
			continue
		}

		if string(tagName) != "meta" {
			continue
		}
		mt := Meta{
			Name:    "",
			Content: "",
		}
		for {
			tagAttrKey, tagAttrVal, hasMore := t.TagAttr()
			if string(tagAttrKey) == "name" {
				mt.Name = string(tagAttrVal)
			}
			if string(tagAttrKey) == "content" {
				mt.Content = string(tagAttrVal)
			}
			if !hasMore {
				break
			}
		}
		metas = append(metas, mt)
	}
	return metas, nil
}

func getLinks(t html.Tokenizer) ([]Link, error) {
	var links []Link

	for {
		tt := t.Next()
		if tt == html.ErrorToken {
			if t.Err() == io.EOF {
				break
			}
			log.Printf("Error: %v", t.Err())
			break
		}

		// If the tag is a self-closing tag or the start tag we will search it.
		// Otherwise, continue to the next string. There's no point in looking at other tags
		if tt != html.SelfClosingTagToken && tt != html.TextToken && tt != html.StartTagToken {
			// fmt.Println("Skipping:", string(tagName))
			continue
		}

		// The tag have no attributes, so they're of no interest. For example a <style> or <script> tag
		tagName, hasAttr := t.TagName()

		if !hasAttr {
			continue
		}

		if string(tagName) != "link" {
			continue
		}

		lk := Link{
			Rel:  "",
			Href: "",
		}
		for {
			tagAttrKey, tagAttrVal, hasMore := t.TagAttr()
			if string(tagAttrKey) == "rel" {
				lk.Rel = string(tagAttrVal)
			}
			if string(tagAttrKey) == "href" {
				lk.Href = string(tagAttrVal)
			}
			if !hasMore {
				break
			}
		}
		links = append(links, lk)
	}

	return links, nil
}

func createURL(input string, base url.URL) string {
	if strings.HasPrefix(input, "/") {
		return cleanURL(base.Scheme + "://" + base.Hostname() + input)
	}
	if strings.HasPrefix(input, "http") {
		return cleanURL(input)
	}
	baseUrl := base.Scheme + "://" + base.Hostname()
	parts := strings.Split(strings.TrimSpace(base.Path), "/")
	updatedParts := make([]string, 0)

	for _, part := range parts {
		if !strings.Contains(part, ".") {
			updatedParts = append(updatedParts, part)
		}
	}
	partsStr := strings.Join(updatedParts, "/")

	if base.Path != "" && len(parts) > 2 {
		baseUrl = baseUrl + partsStr + "/"
	} else {
		baseUrl = baseUrl + "/"
	}

	return cleanURL(baseUrl + input)
}

func cleanURL(urlStr string) string {
	parsedURL, err := url.Parse(urlStr)
	if err != nil {
		log.Printf("Failed to clean up parsedUrl with error: %v\n", err)
		return urlStr
	}

	return parsedURL.Scheme + "://" + parsedURL.Host + parsedURL.Path
}
