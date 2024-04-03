package meraki

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"io"
	"net/http"
)

var (
	base_url = "https://api.meraki.com/api/v1"
)

type Organization struct {
	ID   string `json:"id"`
	Name string `json:"name"`
	URL  string `json:"url"`
	API  struct {
		Enabled bool `json:"enabled"`
	} `json:"api"`
	Licensing struct {
		Model string `json:"model"`
	} `json:"licensing"`
	Cloud struct {
		Region struct {
			Name string `json:"name"`
		} `json:"region"`
	} `json:"cloud"`
	Management struct {
		Details []string `json:"details"`
	} `json:"management"`
}

type Network struct {
	ID                      string   `json:"id"`
	OrgID                   string   `json:"organizationId"`
	Name                    string   `json:"name"`
	ProductTypes            []string `json:"productTypes"`
	TimeZone                string   `json:"timeZone"`
	Tags                    []string `json:"tags"`
	EnrollmentString        string   `json:"enrollmentString"`
	URL                     string   `json:"url"`
	Notes                   string   `json:"notes"`
	IsBoundToConfigTemplate bool     `json:"isBoundToConfigTemplate"`
}

type NetworkCreateRequest struct {
	Name         string   `json:"name"`
	ProductTypes []string `json:"productTypes"`
	TimeZone     string   `json:"timeZone"`
	Tags         []string `json:"tags"`
}

type NetworkUpdateRequest struct {
	Name     string `json:"name,omitempty"`
	TimeZone string `json:"timeZone,omitempty"`
}

type Client interface {
	// Organizations
	GetOrganizations() ([]Organization, error)
	GetOrganization(orgID string) (*Organization, error)

	// Networks
	CreateNetwork(orgID string, network *NetworkCreateRequest) (*Network, error)
	GetNetwork(id string) (*Network, error)
	UpdateNetwork(id string, network *NetworkUpdateRequest) (*Network, error)
	DeleteNetwork(id string) error
}

func NewClient(apiToken string) Client {
	return &client{
		token:      apiToken,
		httpClient: &http.Client{},
	}
}

type client struct {
	token      string
	httpClient *http.Client
}

func (c *client) GetOrganizations() ([]Organization, error) {
	url := base_url + "/organizations"

	// use the token to make http get request to the url
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Add("Authorization", "Bearer "+c.token)
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	// parse the response body into a organization slice
	var orgs []Organization
	err = json.NewDecoder(resp.Body).Decode(&orgs)
	if err != nil {
		return nil, err
	}
	return orgs, nil
}

func (c *client) GetOrganization(orgID string) (*Organization, error) {
	url := base_url + "/organizations/" + orgID

	// use the token to make http get request to the url
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Add("Authorization", "Bearer "+c.token)
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	// parse the response body into a organization slice
	var org Organization
	err = json.NewDecoder(resp.Body).Decode(&org)
	if err != nil {
		return nil, err
	}
	return &org, nil
}

func (c *client) CreateNetwork(orgID string, network *NetworkCreateRequest) (*Network, error) {
	url := base_url + "/organizations/" + orgID + "/networks"

	rb, err := json.Marshal(network)
	if err != nil {
		return nil, err
	}

	tflog.Info(context.Background(), "Creating network with request body: "+string(rb)+"\n")

	// use the token to make http get request to the url

	req, err := http.NewRequest("POST", url, bytes.NewBuffer(rb))
	req.Header.Add("Content-Type", "application/json")

	if err != nil {
		return nil, err
	}
	req.Header.Add("Authorization", "Bearer "+c.token)
	resp, err := c.httpClient.Do(req)

	if resp.StatusCode != http.StatusCreated {
		results, _ := io.ReadAll(resp.Body)
		return nil, errors.New("Failed to create network: " + url + "," + string(rb) + ", " + string(results))
	}

	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	// parse the response body into a organization slice
	var net Network
	err = json.NewDecoder(resp.Body).Decode(&net)
	if err != nil {
		return nil, err
	}
	return &net, nil
}

func (c *client) GetNetwork(id string) (*Network, error) {
	url := base_url + "/networks/" + id
	// use the token to make http get request to the url
	req, err := http.NewRequest("GET", url, nil)

	if err != nil {
		return nil, err
	}
	req.Header.Add("Authorization", "Bearer "+c.token)
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	// parse the response body into a organization slice
	var net Network
	err = json.NewDecoder(resp.Body).Decode(&net)
	if err != nil {
		return nil, err
	}
	return &net, nil
}

func (c *client) UpdateNetwork(id string, network *NetworkUpdateRequest) (*Network, error) {
	url := base_url + "/networks/" + id

	rb, err := json.Marshal(network)
	if err != nil {
		return nil, err
	}

	tflog.Info(context.Background(), "Creating network with request body: "+string(rb)+"\n")

	// use the token to make http get request to the url

	req, err := http.NewRequest("PUT", url, bytes.NewBuffer(rb))
	req.Header.Add("Content-Type", "application/json")

	if err != nil {
		return nil, err
	}
	req.Header.Add("Authorization", "Bearer "+c.token)
	resp, err := c.httpClient.Do(req)

	if resp.StatusCode != http.StatusOK {
		results, _ := io.ReadAll(resp.Body)
		return nil, errors.New("Failed to update network: " + url + "," + string(rb) + ", " + string(results))
	}

	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	// parse the response body into a organization slice
	var net Network
	err = json.NewDecoder(resp.Body).Decode(&net)
	if err != nil {
		return nil, err
	}
	return &net, nil
}

func (c *client) DeleteNetwork(id string) error {
	url := base_url + "/networks/" + id

	req, err := http.NewRequest("DELETE", url, nil)
	if err != nil {
		return err
	}
	req.Header.Add("Authorization", "Bearer "+c.token)
	resp, err := c.httpClient.Do(req)

	if err != nil {
		return err
	}

	if resp.StatusCode != http.StatusNoContent {
		results, _ := io.ReadAll(resp.Body)
		return errors.New("Failed to update network: " + url + ", " + string(results))
	}
	return nil
}
