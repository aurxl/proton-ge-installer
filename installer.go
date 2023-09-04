// This little installer is trying to follow the exact steps as the install script

package main

import (
	"archive/tar"
	"compress/gzip"
	"crypto/sha512"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"os"
	"strings"
)

const (
	GHapi_url       = "https://api.github.com/"
	GHreq           = "repos/GloriousEggroll/proton-ge-custom/releases/"
	GHtags          = "tags/"
	versionUsage    = "GE Version (release) to install"
	steam_rootUsage = "steam root dir"
)

var (
	version    string
	steam_root string
	user_home  string
	dwl_index  int
	csm_index  int
)

type releaseInfos struct {
	Url    string
	Assets []struct {
		Name                 string
		Browser_download_url string
	}
}

func getValidRelease(version string) (string, *releaseInfos, error) {
	var rls releaseInfos
	url, _ := url.ParseRequestURI(GHapi_url)

	if version == "latest" {
		url.Path = GHreq + version
	} else {
		if !strings.HasPrefix(version, "GE-Proton") {
			version = "GE-Proton" + version
		}
		url.Path = GHreq + GHtags + version
	}

	url_str := url.String()
	resp, err := http.Get(url_str)
	if err != nil {
		log.Fatal(err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusOK {
		err = json.NewDecoder(resp.Body).Decode(&rls)
		if err != nil {
			return "", nil, err
		}
		return version, &rls, nil
	} else {
		return "", nil, errors.New("invalid Release version: " + version + " {" + url_str + "}")
	}
}

func downloadRelease(name string, url string) (string, error) {
	out, err := os.Create(name)
	if err != nil {
		return "", err
	}
	defer out.Close()

	resp, err := http.Get(url)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", errors.New("bad status:" + resp.Status)
	}

	_, err = io.Copy(out, resp.Body)
	if err != nil {
		return "", err
	}
	return name, nil
}

func getSHA512SumFromUrl(name string, tarname string, url string) (string, error) {
	resp, err := http.Get(url)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", errors.New(":" + resp.Status)
	}

	body, _ := io.ReadAll(resp.Body)
	return strings.Fields(string(body))[0], nil
}

func calcSHA512Sum(filename string) (string, error) {
	file, err := os.Open(filename)
	if err != nil {
		return "", err
	}
	defer file.Close()

	hash := sha512.New()
	_, err = io.Copy(hash, file)
	if err != nil {
		return "", err
	}

	hashInBytes := hash.Sum(nil)
	return fmt.Sprintf("%x", hashInBytes), nil
}

func unpackTarGz(filename string) error {
	gzipStream, err := os.Open(filename)
	if err != nil {
		return err
	}

	uncompressedStream, err := gzip.NewReader(gzipStream)
	if err != nil {
		return err
	}

	log.Println("Extract archive")
	// https://codereview.stackexchange.com/a/272554
	tarReader := tar.NewReader(uncompressedStream)
	var header *tar.Header
	for header, err = tarReader.Next(); err == nil; header, err = tarReader.Next() {
		switch header.Typeflag {
		case tar.TypeDir:
			if err := os.Mkdir(header.Name, 0755); err != nil {
				return err
			}
		default:
			outFile, err := os.Create(header.Name)
			if err != nil {
				return err
			}

			if _, err := io.Copy(outFile, tarReader); err != nil {
				outFile.Close()
				return err
			}
			if err := outFile.Close(); err != nil {
				return err
			}
			//default:
			//	return  fmt.Errorf("ExtractTarGz: uknown type: %b in %s", header.Typeflag, header.Name)
		}
	}
	if err != io.EOF {
		return err
	}
	return nil
}

func init() {
	log.SetPrefix("Proton GE Installer: ")
	log.SetFlags(0)

	user_home, err := os.UserHomeDir()
	if err != nil {
		log.Fatal(err)
	}

	flag.StringVar(&version, "version", "latest", versionUsage)
	flag.StringVar(&version, "v", "latest", versionUsage+" - shorthand")
	flag.StringVar(&steam_root, "steam_dir", user_home+"/.steam/root/", steam_rootUsage)
	flag.StringVar(&steam_root, "d", user_home+"/.steam/root/", steam_rootUsage+" - shorthand")

	flag.Parse()

	if flag.Arg(0) != "" {
		version = flag.Args()[0]
	}

}

func main() {
	// checking if input is a valid release and get needed urls
	validVersion, urls, err := getValidRelease(version)
	if err != nil {
		log.Fatal(err)
	} else {
		log.Printf("found release: %s", validVersion)
	}

	/*
		// createing temp dir /tmp/proton-ge-custom
		tmp_dir, err := os.MkdirTemp("", "proton-ge-custom")
		if err != nil {
			log.Fatal(err)
		} else {
			log.Printf("created dir: %s", tmp_dir)
		}
	*/

	//cd to temp dir
	err = os.Chdir(steam_root + "compatibilitytools.d")
	if err != nil {
		log.Fatal(err)
	}

	for i, url := range urls.Assets {
		if !strings.HasSuffix(url.Name, "sha512sum") {
			dwl_index = i
		} else {
			csm_index = i
		}
	}

	// download tarball
	downloadedFile, err := downloadRelease(urls.Assets[dwl_index].Name, urls.Assets[dwl_index].Browser_download_url)
	if err != nil {
		log.Fatal(err)
	} else {
		log.Printf("Successfully downloaded %s", urls.Assets[dwl_index].Name)
	}
	defer os.Remove(downloadedFile)

	// get checksum
	tarsumBefore, err := getSHA512SumFromUrl(urls.Assets[csm_index].Name, urls.Assets[dwl_index].Name, urls.Assets[csm_index].Browser_download_url)
	if err != nil {
		log.Fatal(err)
	}

	// generating sha512 hash of tarball
	tarsumAfter, err := calcSHA512Sum(urls.Assets[dwl_index].Name)
	if err != nil {
		log.Fatal(err)
	}

	// Comparing checksums
	if tarsumBefore == tarsumAfter {
		log.Println("Checksums matching!")
	} else {
		log.Fatal(errors.New("checksums not matching"))
	}

	// Create dir compatabilitytools.d
	err = os.Mkdir(steam_root+"compatibilitytools.d", 0755)
	if err != nil {
		log.Println(err)
	}

	// extract tar ball
	err = unpackTarGz(urls.Assets[dwl_index].Name)
	if err != nil {
		log.Fatal(err)
	}

	log.Println("Done!")
}
