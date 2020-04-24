package cloudflare

import (
	"bytes"
	"encoding/json"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"strings"
)

const (
	CLOUDFLARE_ENDPOINT = "https://api.cloudflare.com/client/v4"
	CHECK_IP_ENDPOINT   = "http://checkip.amazonaws.com/"
)

type CFZoneResult struct {
	Success bool     `json:"success"`
	Result  []CFZone `json:"result"`
}

type CFZone struct {
	Name    string `json:"name"`
	ID      string `json:"id"`
	Records map[string]CFRecord
}

type CFRecordResult struct {
	Success bool       `json:"success"`
	Result  []CFRecord `json:"result"`
}

type CFUpdateRecordResult struct {
	Success bool     `json:"success"`
	Result  CFRecord `json:"result"`
}

type CFRecord struct {
	Name    string `json:"name"`
	ID      string `json:"id"`
	Type    string `json:"type"`
	Content string `json:"content"`
}

type CFConfig struct {
	CFKEY  string   // Cloudflare's API Key
	CFUSER string   // Cloudflare's User Email
	Zone   CFZone   // Cloudflare's Domain to update
	CFHOST []string // List of subdomains to update
}

type CFClient struct {
	BaseURL *url.URL
	Config  *CFConfig

	httpClient *http.Client
}

func (c *CFClient) newRequest(method, path string, queries map[string]string, body interface{}) (*http.Request, error) {
	rel := &url.URL{Path: c.BaseURL.Path + path}
	u := c.BaseURL.ResolveReference(rel)
	if queries != nil {
		q, _ := url.ParseQuery(u.RawQuery)
		for key, value := range queries {
			q.Add(key, value)
		}
		u.RawQuery = q.Encode()
	}
	var buf io.ReadWriter
	if body != nil {
		buf = new(bytes.Buffer)
		err := json.NewEncoder(buf).Encode(body)
		if err != nil {
			return nil, err
		}
	}
	req, err := http.NewRequest(method, u.String(), buf)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+c.Config.CFKEY)
	return req, nil
}

func (c *CFClient) do(req *http.Request, v interface{}) (*http.Response, error) {
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	// io.Copy(os.Stdout, resp.Body)
	b, err := ioutil.ReadAll(resp.Body)
	// fmt.Println(string(b))
	json.Unmarshal(b, &v)
	return resp, err
}

func NewClient(apiKey string, email string, zoneName string, hosts []string, httpClient *http.Client) *CFClient {
	if httpClient == nil {
		httpClient = http.DefaultClient
	}
	url, err := url.Parse(CLOUDFLARE_ENDPOINT)

	if err != nil {
		log.Fatal("Could not parse endpoint: %v\n", err)
	}

	c := &CFClient{
		BaseURL: url,
		Config: &CFConfig{
			CFKEY:  apiKey,
			CFUSER: email,
			Zone: CFZone{
				Name: zoneName,
			},
			CFHOST: hosts,
		},
		httpClient: httpClient,
	}
	c.UpdateLocalRecords()
	return c
	// c.ChatService = &ChatService{client: c}
	// c.ChannelService = &ChannelService{client: c}
	// c.UserService = &UserService{client: c}
}

func (c *CFClient) GetZoneID(zone string) (*CFZoneResult, error) {
	query := map[string]string{
		"name": zone,
	}
	req, err := c.newRequest("GET", "/zones", query, nil)
	if err != nil {
		return &CFZoneResult{}, err
	}
	var zoneResult CFZoneResult
	_, err = c.do(req, &zoneResult)
	return &zoneResult, err
}

func (c *CFClient) UpdateLocalRecords() {
	z, _ := c.GetZoneID(c.Config.Zone.Name)
	c.Config.Zone = z.Result[0]
	r, _ := c.GetRecords()
	recordMap := map[string]CFRecord{}
	for _, record := range r.Result {
		recordMap[record.Name] = record
	}
	c.Config.Zone.Records = recordMap
}

func (c *CFClient) GetRecords() (*CFRecordResult, error) {
	// query := map[string]string{
	// 	"name": c.Config.Zone.Name,
	// }
	req, err := c.newRequest("GET", "/zones/"+c.Config.Zone.ID+"/dns_records", nil, nil)
	if err != nil {
		return &CFRecordResult{}, err
	}
	var recordResult CFRecordResult
	_, err = c.do(req, &recordResult)
	return &recordResult, err
}

func (c *CFClient) UpdateRecordContent(host string, content string) error {
	record := c.Config.Zone.Records[host]
	record.Content = content
	req, err := c.newRequest("PUT", "/zones/"+c.Config.Zone.ID+"/dns_records/"+record.ID, nil, record)
	if err != nil {
		return err
	}
	var recordResult CFUpdateRecordResult
	_, err = c.do(req, &recordResult)
	// return &recordResult, err
	// fmt.Println(record.Content == recordResult.Result.Content)
	c.Config.Zone.Records[host] = recordResult.Result
	return err
}

func (c *CFClient) CheckAndUpdate() {
	c.UpdateLocalRecords()
	resp, err := http.Get(CHECK_IP_ENDPOINT)
	if err != nil {
		log.Fatal(err)
	}
	b, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Fatal(err)
	}
	ip := strings.TrimSuffix(string(b), "\n")
	for _, record := range c.Config.CFHOST {
		if c.Config.Zone.Records[record].Content != ip {
			err := c.UpdateRecordContent(record, ip)
			if err != nil {
				log.Fatal("Error")
			} else {
				log.Printf("Updated %v with ip %v", record, ip)
			}
		} else {
			log.Printf("No Update required. Host: %v, IP: %v", record, ip)
		}
	}
}
