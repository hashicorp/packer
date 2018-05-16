package profitbricks

import (
	"encoding/json"
	"net/http"
	"strconv"
)

type ContractResources struct {
	Id         string                      `json:"id,omitempty"`
	Type_      string                      `json:"type,omitempty"`
	Href       string                      `json:"href,omitempty"`
	Properties ContractResourcesProperties `json:"properties,omitempty"`
	Response   string                      `json:"Response,omitempty"`
	Headers    *http.Header                `json:"headers,omitempty"`
	StatusCode int                         `json:"headers,omitempty"`
}

type ContractResourcesProperties struct {
	PBContractNumber string           `json:"PB-Contract-Number,omitempty"`
	Owner            string           `json:"owner,omitempty"`
	Status           string           `json:"status,omitempty"`
	ResourceLimits   *ResourcesLimits `json:"resourceLimits,omitempty"`
}

type ResourcesLimits struct {
	CoresPerServer        int32 `json:"coresPerServer,omitempty"`
	CoresPerContract      int32 `json:"coresPerContract,omitempty"`
	CoresProvisioned      int32 `json:"coresProvisioned,omitempty"`
	RamPerServer          int32 `json:"ramPerServer,omitempty"`
	RamPerContract        int32 `json:"ramPerContract,omitempty"`
	RamProvisioned        int32 `json:"ramProvisioned,omitempty"`
	HddLimitPerVolume     int64 `json:"hddLimitPerVolume,omitempty"`
	HddLimitPerContract   int64 `json:"hddLimitPerContract,omitempty"`
	HddVolumeProvisioned  int64 `json:"hddVolumeProvisioned,omitempty"`
	SsdLimitPerVolume     int64 `json:"ssdLimitPerVolume,omitempty"`
	SsdLimitPerContract   int64 `json:"ssdLimitPerContract,omitempty"`
	SsdVolumeProvisioned  int64 `json:"ssdVolumeProvisioned,omitempty"`
	ReservableIps         int32 `json:"reservableIps,omitempty"`
	ReservedIpsOnContract int32 `json:"reservedIpsOnContract,omitempty"`
	ReservedIpsInUse      int32 `json:"reservedIpsInUse,omitempty"`
}

func GetContractResources() ContractResources {
	path := contract_resource_path()
	url := mk_url(path) + `?depth=` + Depth + `&pretty=` + strconv.FormatBool(Pretty)
	req, _ := http.NewRequest("GET", url, nil)
	req.Header.Add("Content-Type", FullHeader)
	resp := do(req)
	return toContractResources(resp)
}

func toContractResources(resp Resp) ContractResources {
	var col ContractResources
	json.Unmarshal(resp.Body, &col)
	col.Response = string(resp.Body)
	col.Headers = &resp.Headers
	col.StatusCode = resp.StatusCode
	return col
}
