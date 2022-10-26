package ip_manager

import (
	"encoding/json"
	uns "k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
)

var ManifestsPath = "./bindata/manifests/cni-config"

type sriovNetwork struct {
	subnet       string           `json:"subnet"`
	resourceName string           `json:"resourcename"`
	ipPool       []sriovIpAddress `json:"ippool,omitempty"`
}

type sriovIpAddress struct {
	Name        string `json:"name"`
	Namespace   string `json:"namespace,omitempty"`
	Address     string `json:"address"`
	Gateway     string `json:"gateway"`
	Nameservers string `json:"nameservers,omitempty"`
	Vlan        int    `json:"vlan,omitempty"`
	VlanQoS     int    `json:"vlanQoS,omitempty"`
	SpoofChk    string `json:"spoofChk,omitempty"`
	Trust       string `json:"trust,omitempty"`
	LinkState   string `json:"linkState,omitempty"`
	MinTxRate   *int   `json:"minTxRate,omitempty"`
	MaxTxRate   *int   `json:"maxTxRate,omitempty"`
}

// RenderNetAttDef renders a net-att-def for sriov CNI
func (si *sriovIpAddress) RenderNetAttDef(resourceName string) (*uns.Unstructured, error) {
	logger := log.WithName("renderNetAttDef")
	logger.Info("Start to render SRIOV CNI NetworkAttachementDefinition")
	// render RawCNIConfig manifests
	data := MakeRenderData()
	data.Data["CniType"] = "sriov"
	data.Data["SriovNetworkName"] = si.Name
	data.Data["SriovNetworkNamespace"] = si.Namespace

	data.Data["SriovCniResourceName"] = resourceName
	data.Data["SriovCniVlan"] = si.Vlan

	if si.VlanQoS <= 7 && si.VlanQoS >= 0 {
		data.Data["VlanQoSConfigured"] = true
		data.Data["SriovCniVlanQoS"] = si.VlanQoS
	} else {
		data.Data["VlanQoSConfigured"] = false
	}

	data.Data["SpoofChkConfigured"] = true
	switch si.SpoofChk {
	case "off":
		data.Data["SriovCniSpoofChk"] = "off"
	case "on":
		data.Data["SriovCniSpoofChk"] = "on"
	default:
		data.Data["SpoofChkConfigured"] = false
	}

	data.Data["TrustConfigured"] = true
	switch si.Trust {
	case "on":
		data.Data["SriovCniTrust"] = "on"
	case "off":
		data.Data["SriovCniTrust"] = "off"
	default:
		data.Data["TrustConfigured"] = false
	}

	data.Data["StateConfigured"] = true
	switch si.LinkState {
	case "enable":
		data.Data["SriovCniState"] = "enable"
	case "disable":
		data.Data["SriovCniState"] = "disable"
	case "auto":
		data.Data["SriovCniState"] = "auto"
	default:
		data.Data["StateConfigured"] = false
	}

	data.Data["MinTxRateConfigured"] = false
	if si.MinTxRate != nil {
		if *si.MinTxRate >= 0 {
			data.Data["MinTxRateConfigured"] = true
			data.Data["SriovCniMinTxRate"] = *si.MinTxRate
		}
	}

	data.Data["MaxTxRateConfigured"] = false
	if si.MaxTxRate != nil {
		if *si.MaxTxRate >= 0 {
			data.Data["MaxTxRateConfigured"] = true
			data.Data["SriovCniMaxTxRate"] = *si.MaxTxRate
		}
	}

	data.Data["SriovCniAddress"] = si.Address
	data.Data["SriovCniGateway"] = si.Gateway
	data.Data["SriovCniNameservers"] = si.Nameservers

	objs, err := RenderDir(ManifestsPath, &data)
	if err != nil {
		return nil, err
	}
	for _, obj := range objs {
		raw, _ := json.Marshal(obj)
		logger.Info("render NetworkAttachementDefinition output", "raw", string(raw))
	}
	return objs[0], nil
}
