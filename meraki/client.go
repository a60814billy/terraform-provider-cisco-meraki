package meraki

import (
	"encoding/json"
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

type Client interface {
	GetOrganizations() ([]Organization, error)
	GetOrganization(orgID string) (*Organization, error)
}

func NewClient(apiToken string) Client {
	return &client{
		token: apiToken,
	}
}

type client struct {
	token string
}

func (c *client) GetOrganizations() ([]Organization, error) {
	url := base_url + "/organizations"

	// use the token to make http get request to the url
	httpClient := &http.Client{}
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Add("Authorization", "Bearer "+c.token)
	resp, err := httpClient.Do(req)
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
	httpClient := &http.Client{}
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Add("Authorization", "Bearer "+c.token)
	resp, err := httpClient.Do(req)
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
