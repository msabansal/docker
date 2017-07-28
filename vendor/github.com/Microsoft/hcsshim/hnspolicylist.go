package hcsshim

import (
	"encoding/json"

	"github.com/Sirupsen/logrus"
)

type L2NATPolicy struct {
	LBPolicy
	VIP string `json:"VIP,omitempty"`
}

type ELBPolicy struct {
	LBPolicy
	SourceVIP string   `json:"SourceVIP,omitempty"`
	VIPs      []string `json:"VIPs,omitempty"`
	ILB       bool     `json:"ILB,omitempty"`
}

type LBPolicy struct {
	Policy
	Protocol     uint16 `json:"Protocol,omitempty"`
	InternalPort uint16
	ExternalPort uint16
}

type PolicyList struct {
	Id                 string            `json:"ID,omitempty"`
	EndpointReferences []string          `json:"References,omitempty"`
	Policies           []json.RawMessage `json:"Policies,omitempty"`
}

// HNSPolicyListRequest makes a call into HNS to update/query a single network
func HNSPolicyListRequest(method, path, request string) (*PolicyList, error) {
	var policy PolicyList
	err := hnsCall(method, "/policylists/"+path, request, &policy)
	if err != nil {
		return nil, err
	}

	return &policy, nil
}

func HNSListPolicyListRequest() ([]PolicyList, error) {
	var plist []PolicyList
	err := hnsCall("GET", "/policylists/", "", &plist)
	if err != nil {
		return nil, err
	}

	return plist, nil
}

// PolicyListRequest makes a HNS call to modify/query a network endpoint
func PolicyListRequest(method, path, request string) (*PolicyList, error) {
	policylist := &PolicyList{}
	err := hnsCall(method, "/policylists/"+path, request, &policylist)
	if err != nil {
		logrus.Debugf("Request failed =%s", err)
		return nil, err
	}

	return policylist, nil
}

// Create PolicyList by sending PolicyListRequest to HNS.
func (policylist *PolicyList) Create() (*PolicyList, error) {
	operation := "Create"
	title := "HCSShim::PolicyList::" + operation
	logrus.Debugf(title+" id=%s", policylist.Id)
	jsonString, err := json.Marshal(policylist)
	if err != nil {
		return nil, err
	}
	return PolicyListRequest("POST", "", string(jsonString))
}

// Create PolicyList by sending PolicyListRequest to HNS
func (policylist *PolicyList) Delete() (*PolicyList, error) {
	operation := "Delete"
	title := "HCSShim::PolicyList::" + operation
	logrus.Debugf(title+" id=%s", policylist.Id)

	return PolicyListRequest("DELETE", policylist.Id, "")
}

// Add an endpoint to a Policy List
func (policylist *PolicyList) AddEndpoint(endpoint *string) (*PolicyList, error) {
	operation := "AddEndpoint"
	title := "HCSShim::PolicyList::" + operation
	logrus.Debugf(title+" id=%s, endpointId:%s", policylist.Id, endpoint)

	// FIXME:  Remove this delete, once we support Update
	_, err := policylist.Delete()
	if err != nil {
		return nil, err
	}

	// Add Endpoint to the Existing List
	policylist.EndpointReferences = append(policylist.EndpointReferences, "/endpoints/"+*endpoint)

	return policylist.Create()
}

// Remove an endpoint from the Policy List
func (policylist *PolicyList) RemoveEndpoint(endpoint *string) (*PolicyList, error) {
	operation := "AddEndpoint"
	title := "HCSShim::PolicyList::" + operation
	logrus.Debugf(title+" id=%s, endpointId:%s", policylist.Id, endpoint)

	_, err := policylist.Delete()
	if err != nil {
		return nil, err
	}

	elementToRemove := "/endpoints/" + *endpoint

	var references []string

	for _, endpointReference := range policylist.EndpointReferences {
		if endpointReference == elementToRemove {
			continue
		}
		references = append(references, endpointReference)
	}
	policylist.EndpointReferences = references
	return policylist.Create()
}

// AddL2NAT policy list for the specified endpoints
func AddOutboundNAT(endpoints []string, vip string, protocol uint16, internalPort uint16, externalPort uint16) (*PolicyList, error) {
	operation := "AddOutboundNAT"
	title := "HCSShim::PolicyList::" + operation
	logrus.Debugf(title+" Vip:%s", vip)

	policylist := &PolicyList{}

	l2Policy := &L2NATPolicy{
		VIP: vip,
	}
	l2Policy.Type = OutboundNat
	l2Policy.Protocol = protocol
	l2Policy.InternalPort = internalPort
	l2Policy.ExternalPort = externalPort

	for _, endpoint := range endpoints {
		policylist.EndpointReferences = append(policylist.EndpointReferences, "/endpoints/"+endpoint)
	}

	jsonString, err := json.Marshal(l2Policy)
	if err != nil {
		return nil, err
	}

	policylist.Policies[0] = jsonString
	return policylist.Create()
}

// AddLoadBalancer policy list for the specified endpoints
func AddLoadBalancer(endpoints []string, isILB bool, vip string, elbPolicies []ELBPolicy) (*PolicyList, error) {
	operation := "AddLoadBalancer"
	title := "HCSShim::PolicyList::" + operation
	logrus.Debugf(title+" Vip:%s", vip)

	policylist := &PolicyList{}

	for _, endpoint := range endpoints {
		policylist.EndpointReferences = append(policylist.EndpointReferences, "/endpoints/"+endpoint)
	}

	for _, elbPolicy := range elbPolicies {
		jsonString, err := json.Marshal(elbPolicy)
		if err != nil {
			return nil, err
		}
		policylist.Policies = append(policylist.Policies, jsonString)
	}

	return policylist.Create()
}
