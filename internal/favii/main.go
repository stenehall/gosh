package favii

import (
	"crypto/tls"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path"
	"path/filepath"
	"strings"
)

// Site is a simple wrapper around the meta information we can extract from a site.
type Site struct {
	Name            string
	URL             string
	Icon            string
	ThemeColor      string
	BackgroundColor string
}

type Favvo struct {
	client *Favii
}

func Setup() *Favvo {
	// Create output path
	currentPath, err := os.Getwd()
	if err != nil {
		log.Println("error msg", err)
	}
	outPath := filepath.Join(currentPath, "favicons")

	// Create dir output using above code
	if _, err := os.Stat(outPath); os.IsNotExist(err) {
		err := os.MkdirAll(outPath, 0755)
		if err != nil {
			log.Println("Favicon directory already existed")
		}
	}
	client := New(true)

	return &Favvo{
		client,
	}
}

func (f *Favvo) UpdateSite(site Site) (Site, error) {
	if site.Icon != "" {
		log.Printf("[%s]: Set in configs, don't try to fetch from site. Icon: %s\n", site.Name, site.Icon)
		return site, nil
	}

	log.Printf("[%s]: Missing icon, trying to fetch", site.Name)

	metaInfo, err := f.client.GetMetaInfo(site.URL)
	if err != nil {
		return Site{}, err
	}

	if metaInfo.ThemeColor != "" {
		site.ThemeColor = metaInfo.ThemeColor
		site.BackgroundColor = metaInfo.ThemeColor
	}
	if metaInfo.BackgroundColor != "" {
		site.BackgroundColor = metaInfo.BackgroundColor
	}

	iconURL := metaInfo.GetFaviconURL()
	if manifest, err := metaInfo.GetManifest(); err == nil {
		log.Printf("[%s]: manifest found\n", site.Name)
		iconURL = manifest.IconURL
		if manifest.ThemeColor != "" {
			site.ThemeColor = manifest.ThemeColor
		}
		if manifest.BackgroundColor != "" {
			site.BackgroundColor = manifest.BackgroundColor
		}
	}

	faviconPath, err := downloadFavicon(site, iconURL)
	if err == nil {
		// Update the config with the new icon
		site.Icon = faviconPath

		log.Printf("[%s] Icon %s\n", site.Name, iconURL)
	}
	if err != nil {
		log.Printf("downloadFavicon error: %v", err)
	}

	return site, nil
}

func downloadFavicon(site Site, iconURL string) (string, error) {
	// Accept self-signed ssl certs
	tr := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}
	// Try to follow all redirect, hopefully we'll be able to get a response and grab an icon.
	client := &http.Client{
		Transport: tr,
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			return nil
		}}
	response, err := client.Get(iconURL)
	if err != nil {
		return "", fmt.Errorf("failed requesing the icon: %w", err)
	}

	// nolint
	defer func(Body io.ReadCloser) {
		err := Body.Close()
		if err != nil {
			err = fmt.Errorf("[%s]: Couldn't close the file download. The file might be damaged, %w", site.Name, err)
		}
	}(response.Body)

	faviconPath := "/favicons/" + strings.ReplaceAll(site.Name, " ", "-") + path.Ext(iconURL)

	file, err := os.Create("." + faviconPath)
	if err != nil {
		return "", fmt.Errorf("failed saving the icon file: %w", err)
	}
	// nolint
	defer func(file *os.File) {
		err := file.Close()
		if err != nil {
			err = fmt.Errorf("[%s]: Couldn't close the file, download/save the favicon, %w", site.Name, err)
		}
	}(file)

	_, err = io.Copy(file, response.Body)
	if err != nil {
		return "", fmt.Errorf("failed to copy the icon data into the file: %w", err)
	}

	return faviconPath, nil
}
