package handlers

type SlaDeploymentDescriptor struct {
	SlaVersion   string        `json:"sla_version"`
	CustomerID   string        `json:"customerID"`
	Applications []Application `json:"applications"`
}

type Application struct {
	ApplicationID        string         `json:"applicationID,omitempty"`
	ApplicationName      string         `json:"application_name"`
	ApplicationNamespace string         `json:"application_namespace"`
	ApplicationDesc      string         `json:"application_desc"`
	Microservices        []Microservice `json:"microservices"`
}

type Microservice struct {
	MicroserviceID        string       `json:"microserviceID"`
	MicroserviceName      string       `json:"microservice_name"`
	MicroserviceNamespace string       `json:"microservice_namespace"`
	Virtualization        string       `json:"virtualization"`
	Cmd                   []string     `json:"cmd,omitempty"`
	Vcpu                  int          `json:"vcpu"`
	Vgpu                  int          `json:"vgpu,omitempty"`
	Memory                int          `json:"memory,omitempty"`
	Storage               int          `json:"storage"`
	Bandwidth_in          int          `json:"bandwidth_in,omitempty"`
	Bandwidth_out         int          `json:"bandwidth_out,omitempty"`
	Port                  string       `json:"port,omitempty"`
	Code                  string       `json:"code"`
	Addresses             *Address     `json:"addresses"`
	InstanceList          []Instance   `json:"instance_list"`
	Constraints           []Constraint `json:"constraints,omitempty"`
}

type Address struct {
	RRIP string `json:"rr_ip"`
}

type Instance struct {
	InstanceNumber  int    `json:"instance_number"`
	ClusterID       string `json:"cluster_id"`
	ClusterLocation string `json:"cluster_location"`
	CPU             string `json:"cpu"`
	Memory          string `json:"memory"`
	PublicIP        string `json:"publicip"`
	Status          string `json:"status"`
}

type Constraint struct {
	Type    string `json:"type"`
	Node    string `json:"node"`
	Cluster string `json:"cluster"`
}
