package internal

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
type OutlineVPN struct {
	apiURL  string
	client *http.Client
}

type OutlineVPNClients struct {
    clients map[string]*OutlineVPN
}

// OutlineKey represents access key parameters for Outline server
type OutlineKey struct {
	ID        string `json:"id"`
	Name      string `json:"name"`
	Password  string `json:"password"`
	Port      int64  `json:"port"`
	Method    string `json:"method"`
	AccessURL string `json:"accessUrl"`
}

// OutlineConnectionSource represents connection data given by Outline server
// https://www.reddit.com/r/outlinevpn/wiki/index/dynamic_access_keys/
type OutlineConnectionSource struct {
	Server     string `json:"server"`
	ServerPort int    `json:"server_port"`
	Password   string `json:"password"`
	Method     string `json:"method"`
}

// ServerInfo represents Outline server info
type ServerInfo struct {
	Name               string `json:"name"`
	ServerID           string `json:"serverId"`
	MetricsEnabled     bool   `json:"metricsEnabled"`
	CreatedTimestampMs int64  `json:"createdTimestampMs"`
	Version            string `json:"version"`
	AccessKeyDataLimit struct {
		Bytes int64 `json:"bytes"`
	} `json:"accessKeyDataLimit"`
	PortForNewAccessKeys  int    `json:"portForNewAccessKeys"`
	HostnameForAccessKeys string `json:"hostnameForAccessKeys"`
}

// BytesTransferred represents transferred bytes by client when using Outline VPN
type BytesTransferred struct {
	BytesTransferredByUserId map[string]int64 `json:"bytesTransferredByUserId"`
}

// Set default timeout to 5 seconds
var defaultTimeout = time.Second * 5


// NewOutlineVPN creates a new Outline VPN management connection source.
func NewOutlineVPN(apiURL string, certSha256 string) (*OutlineVPN, error) {
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
	return &OutlineVPN{
		apiURL:  apiURL,
		client: client,
	}, nil
}

// NewOutlineConnection creates a new Outline client connection source.
func NewOutlineConnection(server string, port int, password string, method string) *OutlineConnectionSource {
	return &OutlineConnectionSource{
		Server:     server,
		ServerPort: port,
		Password:   password,
		Method:     method,
	}
}

func NewOutlineKey() *OutlineKey {
	return &OutlineKey{}
}

func (vpn *OutlineVPN) GetKeys() ([]OutlineKey, error) {

    req, err := http.NewRequest("GET", fmt.Sprintf("%s/access-keys/", vpn.apiURL), nil)
    if err != nil {
        panic(err)
    }
    resp, err := vpn.client.Do(req)
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

	var result OutlineKey

	// If key is added, status code always must be 200
	if resp.StatusCode != http.StatusOK {
		return &result, errors.New("unable to retrieve key data")
	}

	// Trying unmarshal response body as `OutlineKey`
    body, err := ioutil.ReadAll(resp.Body)

    if err != nil {
        fmt.Println("Error reading response body:", err)
        panic(err)
    }
	// Trying to unmarshal response body as `ServerInfo`
	if err := json.Unmarshal(body, &result); err != nil {
		return &result, err
	}

	return &result, nil
}

func (vpn *OutlineVPN) KeyExists(id string) bool {
	key, err := vpn.GetKey(id)
	return err == nil && key.IsInitialized()
}

func (vpn *OutlineVPN) GetOrCreateKey(id string) (*OutlineKey, error) {
	if vpn.KeyExists(id) {
		return vpn.GetKey(id)
	}

	key := NewOutlineKey()
	key.ID = id
	return vpn.AddKey(key)
}

func (vpn *OutlineVPN) AddKey(key *OutlineKey) (*OutlineKey, error) {

    jsonData, err := json.Marshal(key)
    if err != nil {
        fmt.Println("Error marshalling JSON:", err)
        panic(err)
    }
    var req *http.Request
    var requestErr error
    if key.ID == "" {
        req, requestErr = http.NewRequest("POST", fmt.Sprintf("%s/access-keys", vpn.apiURL), bytes.NewBuffer(jsonData))
    } else {
        req, requestErr = http.NewRequest("PUT", fmt.Sprintf("%s/access-keys/%s", vpn.apiURL, key.ID), nil)
    }
    if requestErr != nil {
        panic(err)
    }
    resp, err := vpn.client.Do(req)
    if err != nil {
        panic(err)
    }

	if resp.StatusCode != http.StatusCreated {
		return key, errors.New("response error while adding new key")
	}
    body, err := ioutil.ReadAll(resp.Body)
	// Trying to unmarshal response body as `OutlineKey`
	if err := json.Unmarshal(body, &key); err != nil {
		return key, err
	}

	return key, nil
}

func (vpn *OutlineVPN) DeleteKey(key *OutlineKey) error {
	return vpn.DeleteKeyByID(key.ID)
}

func (vpn *OutlineVPN) DeleteKeyByID(id string) error {
    req, err := http.NewRequest("DELETE", fmt.Sprintf("%s/access-keys/%s", vpn.apiURL, id), nil)
    if err != nil {
        panic(err)
    }
    resp, err := vpn.client.Do(req)
    if err != nil {
        panic(err)
    }
    defer resp.Body.Close()

	// If key is deleted, status code always must be 204
	if resp.StatusCode != http.StatusNoContent {
		return errors.New("response error while adding new key")
	}

	return nil
}

func (vpn *OutlineVPN) RenameKey(key *OutlineKey, name string) error {
	err := vpn.RenameKeyByID(key.ID, name)
	if err != nil {
		return err
	}
	key.Name = name
	return nil
}

type RenameKeyRequestSchema struct {
    Name string `json:"name"`
}

func (vpn *OutlineVPN) RenameKeyByID(id string, name string) error {
    requestBody := RenameKeyRequestSchema{Name: name}

    // Marshal the struct to JSON
    jsonData, err := json.Marshal(&requestBody)
    if err != nil {
        fmt.Println("Error marshalling JSON:", err)
        panic(err)
    }
    req, err := http.NewRequest("PUT", fmt.Sprintf("%s/access-keys/%s/name", vpn.apiURL, id), bytes.NewBuffer(jsonData))
    if err != nil {
        panic(err)
    }
    req.Header.Add("content-type", "application/x-www-form-urlencoded")
    resp, err := vpn.client.Do(req)
    if err != nil {
        panic(err)
    }

    defer resp.Body.Close()

	// If key is renamed, status code always must be 204
	if resp.StatusCode != http.StatusNoContent {
		return errors.New("response error while renaming key")
	}

	return nil
}

type TransferMetricResponseSchema struct {
    Name string `json:"name"`
    Age  int    `json:"age"`
}

func (vpn *OutlineVPN) GetTransferMetrics() (*BytesTransferred, error) {

    req, err := http.NewRequest("GET", fmt.Sprintf("%s/metrics/transfer"), nil)
    if err != nil {
        panic(err)
    }
    resp, err := vpn.client.Do(req)
    if err != nil {
        panic(err)
    }
    defer resp.Body.Close()

    var result BytesTransferred

	// If data is gathered, status code always must be lower than 400
	if resp.StatusCode >= http.StatusBadRequest {
		return &result, errors.New("unable to get metrics for keys")
	}

	// Trying to unmarshal response body as `BytesTransferred`

    body, err := ioutil.ReadAll(resp.Body)

    if err != nil {
        fmt.Println("Error reading response body:", err)
        panic(err)
    }

	if err := json.Unmarshal(body, &result); err != nil {
		return &result, err
	}

	return &result, nil
}


type ServerInfoResponseSchema struct {
    Name string `json:"name"`
    Age  int    `json:"age"`
}

func (vpn *OutlineVPN) GetServerInfo() (*ServerInfo, error) {
    var result ServerInfo

    req, err := http.NewRequest("GET", fmt.Sprintf("%s/server", vpn.apiURL), nil)
    if err != nil {
        panic(err)
    }
    resp, err := vpn.client.Do(req)
    defer resp.Body.Close()

	// If data is gathered, status code always must be lower than 400
	if resp.StatusCode >= http.StatusBadRequest {
		return &result, errors.New("unable to get metrics for keys")
	}

    body, err := ioutil.ReadAll(resp.Body)

    if err != nil {
        fmt.Println("Error reading response body:", err)
        panic(err)
    }
	// Trying to unmarshal response body as `ServerInfo`
	if err := json.Unmarshal(body, &result); err != nil {
		return &result, err
	}

	return &result, nil
}

func (key *OutlineKey) AsSource() (*OutlineConnectionSource, error) {
	if !key.IsInitialized() {
		return nil, errors.New("unable to retrieve key access url")
	}

	// Parse the access url
	u, err := url.Parse(key.AccessURL)
	if err != nil {
		return nil, err
	}

	// Decode user info
	userInfo := strings.TrimPrefix(u.User.String(), ":")
	decoded, err := base64.StdEncoding.DecodeString(userInfo)
	if err != nil {
		return nil, err
	}

	// Define the host
	host := u.Hostname()
	// Trying to convert the port into an integer
	port, err := strconv.Atoi(u.Port())
	if err != nil {
		return nil, err
	}

	// Split decoded data
	data := strings.Split(string(decoded), ":")
	if len(data) != 2 {
		return nil, errors.New("decoded access url doesn't contains password or method")
	}

	return NewOutlineConnection(host, port, data[1], data[0]), nil
}

func (key *OutlineKey) IsInitialized() bool {
	return key.AccessURL != ""
}