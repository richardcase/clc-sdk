package server

import (
	"fmt"
	"regexp"

	"github.com/mikebeyer/clc-sdk/sdk/api"
)

func New(client *api.Client) *Service {
	return &Service{client}
}

type Service struct {
	client *api.Client
}

func (s *Service) Get(name string) (*Response, error) {
	url := fmt.Sprintf("%s/servers/%s/%s", s.client.Config.BaseURL, s.client.Config.Alias, name)
	if regexp.MustCompile("^[0-9a-f]{32}$").MatchString(name) {
		url = fmt.Sprintf("%s?uuid=true", url)
	}
	server := &Response{}
	err := s.client.Get(url, server)
	return server, err
}

func (s *Service) Create(server Server, poll chan bool) (*QueuedResponse, error) {
	if !server.Valid() {
		return nil, fmt.Errorf("server: server missing required field(s). (Name, CPU, MemoryGB, GroupID, SourceServerID, Type)")
	}

	resp := &QueuedResponse{}
	url := fmt.Sprintf("%s/servers/%s", s.client.Config.BaseURL, s.client.Config.Alias)
	err := s.client.Post(url, server, resp)
	if err != nil {
		return nil, err
	}

	return resp, nil
}

type Server struct {
	Name            string         `json:"name"`
	Description     string         `json:"description,omitempty"`
	GroupID         string         `json:"groupId"`
	SourceServerID  string         `json:"sourceServerId"`
	IsManagedOS     bool           `json:"isManagedOS,omitempty"`
	PrimaryDNS      string         `json:"primaryDns,omitempty"`
	SecondaryDNS    string         `json:"secondaryDns,omitempty"`
	NetworkID       string         `json:"networkId,omitempty"`
	IPaddress       string         `json:"ipAddress,omitempty"`
	Password        string         `json:"password,omitempty"`
	CPU             int            `json:"cpu"`
	MemoryGB        int            `json:"memoryGB"`
	Type            string         `json:"type"`
	Storagetype     string         `json:"storageType,omitempty"`
	Customfields    []Customfields `json:"customFields,omitempty"`
	Additionaldisks []struct {
		Path   string `json:"path"`
		SizeGB int    `json:"sizeGB"`
		Type   string `json:"type"`
	} `json:"additionalDisks,omitempty"`
}

func (s *Server) Valid() bool {
	return s.Name != "" && s.CPU != 0 && s.MemoryGB != 0 && s.GroupID != "" && s.SourceServerID != ""
}

type Response struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
	GroupID     string `json:"groupId"`
	IsTemplate  bool   `json:"isTemplate"`
	LocationID  string `json:"locationId"`
	OStype      string `json:"osType"`
	Status      string `json:"status"`
	Details     struct {
		IPaddresses []struct {
			Internal string `json:"internal"`
		} `json:"ipAddresses"`
		AlertPolicies []struct {
			ID    string `json:"id"`
			Name  string `json:"name"`
			Links []Link `json:"links"`
		} `json:"alertPolicies"`
		CPU               int    `json:"cpu"`
		Diskcount         int    `json:"diskCount"`
		Hostname          string `json:"hostName"`
		InMaintenanceMode bool   `json:"inMaintenanceMode"`
		MemoryMB          int    `json:"memoryMB"`
		Powerstate        string `json:"powerState"`
		Storagegb         int    `json:"storageGB"`
		Disks             []struct {
			ID             string        `json:"id"`
			SizeGB         int           `json:"sizeGB"`
			PartitionPaths []interface{} `json:"partitionPaths"`
		} `json:"disks"`
		Partitions []struct {
			SizeGB float64 `json:"sizeGB"`
			Path   string  `json:"path"`
		} `json:"partitions"`
		Snapshots []struct {
			Name  string `json:"name"`
			Links []Link `json:"links"`
		} `json:"snapshots"`
		Customfields []Customfields `json:"customFields,omitempty"`
	} `json:"details"`
	Type        string `json:"type"`
	Storagetype string `json:"storageType"`
	ChangeInfo  struct {
		CreatedDate  string `json:"createdDate"`
		CreatedBy    string `json:"createdBy"`
		ModifiedDate string `json:"modifiedDate"`
		ModifiedBy   string `json:"modifiedBy"`
	} `json:"changeInfo"`
	Links Links `json:"links"`
}

type QueuedResponse struct {
	Server   string `json:"server"`
	IsQueued bool   `json:"isQueued"`
	Links    Links  `json:"links"`
}

type Customfields struct {
	ID           string `json:"id,omitempty"`
	Name         string `json:"name,omitempty"`
	Value        string `json:"value,omitempty"`
	Displayvalue string `json:"displayValue,omitempty"`
}

type Link struct {
	Rel   string   `json:"rel,omitempty"`
	Href  string   `json:"href,omitempty"`
	ID    string   `json:"id,omitempty"`
	Verbs []string `json:"verbs,omitempty"`
}

type Links []Link

func (l Links) GetID(rel string) (bool, string) {
	for _, v := range l {
		if v.Rel == rel {
			return true, v.ID
		}
	}
	return false, ""
}