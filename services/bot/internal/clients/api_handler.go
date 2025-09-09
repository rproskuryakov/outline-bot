package clients

package clients

import (
	"crypto/tls"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
    "io/ioutil"
    "net/http"
    "net/url"
	"strconv"
	"bytes"
	"strings"
	"time"
)

// OutlineVPN represents connection source to manage Outline VPN server
type RestAPIClient struct {
	apiURL  string
	client *http.Client
}

// Set default timeout to 5 seconds
var defaultTimeout = time.Second * 5


// NewOutlineVPN creates a new Outline VPN management connection source.
func NewAPIClient(apiURL string, certSha256 string) (*RestAPIClient, error) {
	// todo
	/*if certSha256 == "" {
		return nil, fmt.Errorf("no certificate SHA256 provided. Running without certificate is no longer supported")
	}*/
	// Creating a client
	tr := &http.Transport{
	    TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
    }
	client := &http.Client{Transport: tr}

	// Create OutlineVPN instance with configured TLS client
	return &RestAPIClient{
		apiURL:  apiURL,
		client: client,
	}, nil
}

func (client *RestAPIClient) GetKeys() ([]OutlineKey, error) {

    req, err := http.NewRequest("GET", fmt.Sprintf("%s/access-keys/", client.apiURL), nil)
    if err != nil {
        panic(err)
    }
    resp, err := client.client.Do(req)
    if err != nil {
        panic(err)
    }
    defer resp.Body.Close()

	// If keys is gathered, status code always must be 200
	if resp.StatusCode != http.StatusOK {
		return nil, errors.New("unable to retrieve keys")
	}

	var result struct {
		AccessKeys []OutlineKey `json:"accessKeys"`
	}

    body, err := ioutil.ReadAll(resp.Body)

    if err != nil {
        fmt.Println("Error reading response body:", err)
        return result.AccessKeys, err
    }
	// Trying to unmarshal response body as `ServerInfo`
	if err := json.Unmarshal(body, &result); err != nil {
		return result.AccessKeys, err
	}

	return result.AccessKeys, nil
}

func (vpn *OutlineVPN) GetKey(id string) (*OutlineKey, error) {

    req, err := http.NewRequest("GET", fmt.Sprintf("%s/access-keys/%s", vpn.apiURL, id), nil)
    if err != nil {
        panic(err)
    }
    resp, err := vpn.client.Do(req)
    if err != nil {
        panic(err)
    }
    defer resp.Body.Close()
    return
// // 	var result OutlineKey
//
// 	// If key is added, status code always must be 200
// 	if resp.StatusCode != http.StatusOK {
// 		return &result, errors.New("unable to retrieve key data")
// 	}
//
// 	// Trying unmarshal response body as `OutlineKey`
//     body, err := ioutil.ReadAll(resp.Body)
//
//     if err != nil {
//         fmt.Println("Error reading response body:", err)
//         panic(err)
//     }
// 	// Trying to unmarshal response body as `ServerInfo`
// 	if err := json.Unmarshal(body, &result); err != nil {
// 		return &result, err
// 	}
//
// 	return &result, nil
}
