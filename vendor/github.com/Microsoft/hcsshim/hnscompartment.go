package hcsshim

import (
	"encoding/json"

	"github.com/sirupsen/logrus"
)

// CompartmentResourceType of Request Support in CompartmentResource
type CompartmentResourceType string

// RequestType const
const (
	Endpoint CompartmentResourceType = "Endpoint"
)

// CompartmentResource is a structure defining schema for Route based Policy
type CompartmentResource struct {
	Type CompartmentResourceType `json:",omitempty"`
	Data json.RawMessage         `json:",omitempty"`
}

// CompartmentResourceEndpoint is a structure defining schema for ELB LoadBalancing based Policy
type CompartmentResourceEndpoint struct {
	Id string `json:",omitempty"`
}

// Compartment is a structure defining schema for Policy list request
type Compartment struct {
	ID            string `json:"ID,omitempty"`
	Name          string
	CompartmentId uint32                `json:",omitempty"`
	ResourceList  []CompartmentResource `json:",omitempty"`
}

// HNSCompartmentRequest makes a call into HNS to update/query a single network
func HNSCompartmentRequest(method, path, request string) (*Compartment, error) {
	var policy Compartment
	err := hnsCall(method, "/compartments/"+path, request, &policy)
	if err != nil {
		return nil, err
	}

	return &policy, nil
}

// HNSListCompartmentRequest gets all the compartments list
func HNSListCompartmentRequest() ([]Compartment, error) {
	var plist []Compartment
	err := hnsCall("GET", "/compartments/", "", &plist)
	if err != nil {
		return nil, err
	}

	return plist, nil
}

// CompartmentRequest makes a HNS call to modify/query a network policy list
func CompartmentRequest(method, path, request string) (*Compartment, error) {
	compartment := &Compartment{}
	err := hnsCall(method, "/compartments/"+path, request, &compartment)
	if err != nil {
		return nil, err
	}

	return compartment, nil
}

// GetCompartmentByID get the policy list by ID
func GetCompartmentByID(comparmtnetID string) (*Compartment, error) {
	return CompartmentRequest("GET", comparmtnetID, "")
}

// Create Compartment by sending CompartmentRequest to HNS.
func (compartment *Compartment) Create() (*Compartment, error) {
	operation := "Create"
	title := "HCSShim::Compartment::" + operation
	logrus.Debugf(title+" id=%s", compartment.ID)
	jsonString, err := json.Marshal(compartment)
	if err != nil {
		return nil, err
	}
	return CompartmentRequest("POST", "", string(jsonString))
}

// Delete deletes Compartment
func (compartment *Compartment) Delete() (*Compartment, error) {
	operation := "Delete"
	title := "HCSShim::Compartment::" + operation
	logrus.Debugf(title+" id=%s", compartment.ID)

	return CompartmentRequest("DELETE", compartment.ID, "")
}

// AddEndpoint add an endpoint to a Compartment
func (compartment *Compartment) AddEndpoint(endpoint string) (*Compartment, error) {
	operation := "AddEndpoint"
	title := "HCSShim::Compartment::" + operation
	logrus.Debugf(title+" id=%s, endpointId:%s", compartment.ID, endpoint)

	resourceData := CompartmentResourceEndpoint{
		Id: endpoint,
	}

	jsonString, err := json.Marshal(resourceData)
	if err != nil {
		return nil, err
	}

	resource := &CompartmentResource{
		Type: Endpoint,
		Data: jsonString,
	}

	jsonString, err = json.Marshal(resource)
	if err != nil {
		return nil, err
	}

	return CompartmentRequest("POST", compartment.ID+"/addresource", string(jsonString))
}

// RemoveEndpoint removes an endpoint from the Policy List
func (compartment *Compartment) RemoveEndpoint(endpoint string) (*Compartment, error) {
	operation := "RemoveEndpoint"
	title := "HCSShim::Compartment::" + operation
	logrus.Debugf(title+" id=%s, endpointId:%s", compartment.ID, endpoint)

	resourceData := CompartmentResourceEndpoint{
		Id: endpoint,
	}

	jsonString, err := json.Marshal(resourceData)
	if err != nil {
		return nil, err
	}

	resource := &CompartmentResource{
		Type: Endpoint,
		Data: jsonString,
	}

	jsonString, err = json.Marshal(resource)
	if err != nil {
		return nil, err
	}

	return CompartmentRequest("POST", compartment.ID+"/removeresource", string(jsonString))
}
