package explorer

import (
	"encoding/json"
	"math"

	"github.com/patrickmn/go-cache"
	"github.com/pkg/errors"
	"github.com/threefoldtech/grid_proxy_server/internal/explorer/db"
	"github.com/threefoldtech/zos/pkg/gridtypes"
	"github.com/threefoldtech/zos/pkg/rmb"
)

// ErrNodeNotFound creates new error type to define node existence or server problem
var (
	ErrNodeNotFound = errors.New("node not found")
)

// ErrBadGateway creates new error type to define node existence or server problem
var (
	ErrBadGateway = errors.New("bad gateway")
	ErrBadRequest = errors.New("bad request")
)

// App is the main app objects
type App struct {
	db             db.Database
	rmb            rmb.Client
	lruCache       *cache.Cache
	releaseVersion string
}

// CapacityResult is the NodeData capacity results to unmarshal json in it
type capacityResult struct {
	Total gridtypes.Capacity `json:"total_resources"`
	Used  gridtypes.Capacity `json:"used_resources"`
}

// NodeInfo is node specific info, queried directly from the node
type NodeInfo struct {
	Capacity   capacityResult `json:"capacity"`
	Hypervisor string         `json:"hypervisor"`
	ZosVersion string         `json:"zosVersion"`
}

// Serialize is the serializer for node info struct
func (n *NodeInfo) Serialize() (json.RawMessage, error) {
	bytes, err := json.Marshal(n)
	if err != nil {
		return json.RawMessage{}, errors.Wrap(err, "failed to serialize json data for node info struct")
	}
	return json.RawMessage(bytes), nil
}

// Deserialize is the deserializer for node info struct
func (n *NodeInfo) Deserialize(data []byte) error {
	err := json.Unmarshal(data, n)
	if err != nil {
		return errors.Wrap(err, "failed to deserialize json data for node info struct")
	}
	return nil
}

// NodeStatus is used for status endpoint to decode json in
type NodeStatus struct {
	Status string `json:"status"`
}

// Serialize is the serializer for node status struct
func (n *NodeStatus) Serialize() (json.RawMessage, error) {
	bytes, err := json.Marshal(n)
	if err != nil {
		return json.RawMessage{}, errors.Wrap(err, "failed to serialize json data for node status struct")
	}
	return json.RawMessage(bytes), nil
}

// Deserialize is the deserializer for node status struct
func (n *NodeStatus) Deserialize(data []byte) error {
	err := json.Unmarshal(data, n)
	if err != nil {
		return errors.Wrap(err, "failed to deserialize json data for node status struct")
	}
	return nil
}

type location struct {
	Country string `json:"country"`
	City    string `json:"city"`
}

func roundTotalMemory(cap *gridtypes.Capacity) gridtypes.Capacity {
	return gridtypes.Capacity{
		CRU:   cap.CRU,
		SRU:   cap.SRU,
		HRU:   cap.HRU,
		MRU:   gridtypes.Unit(math.Floor(float64(cap.MRU)/float64(gridtypes.Gigabyte))) * gridtypes.Gigabyte,
		IPV4U: cap.IPV4U,
	}
}

// Node is a struct holding the data for a node for the nodes view
type node struct {
	Version           int                `json:"version"`
	ID                string             `json:"id"`
	NodeID            int                `json:"nodeId"`
	FarmID            int                `json:"farmId"`
	TwinID            int                `json:"twinId"`
	Country           string             `json:"country"`
	GridVersion       int                `json:"gridVersion"`
	City              string             `json:"city"`
	Uptime            int64              `json:"uptime"`
	Created           int64              `json:"created"`
	FarmingPolicyID   int                `json:"farmingPolicyId"`
	UpdatedAt         string             `json:"updatedAt"`
	TotalResources    gridtypes.Capacity `json:"total_resources"`
	UsedResources     gridtypes.Capacity `json:"used_resources"`
	Location          location           `json:"location"`
	PublicConfig      db.PublicConfig    `json:"publicConfig"`
	Status            string             `json:"status"` // added node status field for up or down
	CertificationType string             `json:"certificationType"`
	Hypervisor        string             `json:"hypervisor"`
	ZosVersion        string             `json:"zosVersion"`
	ProxyUpdatedAt    uint64             `json:"proxyUpdatedAt"`
}

func nodeFromDBNode(info db.AllNodeData) node {
	total := roundTotalMemory(&info.NodeData.TotalResources)
	return node{
		Version:         info.NodeData.Version,
		ID:              info.NodeData.ID,
		NodeID:          info.NodeID,
		FarmID:          info.NodeData.FarmID,
		TwinID:          info.NodeData.TwinID,
		Country:         info.NodeData.Country,
		GridVersion:     info.NodeData.GridVersion,
		City:            info.NodeData.City,
		Uptime:          info.NodeData.Uptime,
		Created:         info.NodeData.Created,
		FarmingPolicyID: info.NodeData.FarmingPolicyID,
		UpdatedAt:       info.NodeData.UpdatedAt,
		TotalResources:  total,
		UsedResources: gridtypes.Capacity{
			CRU:   info.PulledNodeData.Resources.UsedCRU,
			SRU:   2*total.SRU - info.PulledNodeData.Resources.FreeSRU,
			HRU:   total.HRU - info.PulledNodeData.Resources.FreeHRU,
			MRU:   total.MRU - info.PulledNodeData.Resources.FreeMRU,
			IPV4U: info.PulledNodeData.Resources.UsedIPV4U,
		},
		Location: location{
			Country: info.NodeData.Country,
			City:    info.NodeData.City,
		},
		PublicConfig:      info.NodeData.PublicConfig,
		Status:            info.PulledNodeData.Status,
		CertificationType: info.NodeData.CertificationType,
		ZosVersion:        info.PulledNodeData.ZosVersion,
		Hypervisor:        info.PulledNodeData.Hypervisor,
		ProxyUpdatedAt:    info.ProxyUpdatedAt,
	}

}

// Node to be compatible with old view
type nodeWithNestedCapacity struct {
	Version           int             `json:"version"`
	ID                string          `json:"id"`
	NodeID            int             `json:"nodeId"`
	FarmID            int             `json:"farmId"`
	TwinID            int             `json:"twinId"`
	Country           string          `json:"country"`
	GridVersion       int             `json:"gridVersion"`
	City              string          `json:"city"`
	Uptime            int64           `json:"uptime"`
	Created           int64           `json:"created"`
	FarmingPolicyID   int             `json:"farmingPolicyId"`
	UpdatedAt         string          `json:"updatedAt"`
	Capacity          capacityResult  `json:"capacity"`
	Location          location        `json:"location"`
	PublicConfig      db.PublicConfig `json:"publicConfig"`
	Status            string          `json:"status"` // added node status field for up or down
	CertificationType string          `json:"certificationType"`
	Hypervisor        string          `json:"hypervisor"`
	ZosVersion        string          `json:"zosVersion"`
	ProxyUpdatedAt    uint64          `json:"proxyUpdatedAt"`
}

func nodeWithNestedCapacityFromDBNode(info db.AllNodeData) nodeWithNestedCapacity {
	total := roundTotalMemory(&info.NodeData.TotalResources)
	return nodeWithNestedCapacity{
		Version:         info.NodeData.Version,
		ID:              info.NodeData.ID,
		NodeID:          info.NodeID,
		FarmID:          info.NodeData.FarmID,
		TwinID:          info.NodeData.TwinID,
		Country:         info.NodeData.Country,
		GridVersion:     info.NodeData.GridVersion,
		City:            info.NodeData.City,
		Uptime:          info.NodeData.Uptime,
		Created:         info.NodeData.Created,
		FarmingPolicyID: info.NodeData.FarmingPolicyID,
		UpdatedAt:       info.NodeData.UpdatedAt,
		Capacity: capacityResult{
			Total: total,
			Used: gridtypes.Capacity{
				CRU:   info.PulledNodeData.Resources.UsedCRU,
				SRU:   2*total.SRU - info.PulledNodeData.Resources.FreeSRU,
				HRU:   total.HRU - info.PulledNodeData.Resources.FreeHRU,
				MRU:   total.MRU - info.PulledNodeData.Resources.FreeMRU,
				IPV4U: info.PulledNodeData.Resources.UsedIPV4U,
			},
		},
		Location: location{
			Country: info.NodeData.Country,
			City:    info.NodeData.City,
		},
		PublicConfig:      info.NodeData.PublicConfig,
		Status:            info.PulledNodeData.Status,
		CertificationType: info.NodeData.CertificationType,
		ZosVersion:        info.PulledNodeData.ZosVersion,
		Hypervisor:        info.PulledNodeData.Hypervisor,
		ProxyUpdatedAt:    info.ProxyUpdatedAt,
	}

}

type farmData struct {
	Farms []db.Farm `json:"farms"`
}

// FarmResult is to unmarshal json in it
type FarmResult struct {
	Data farmData `json:"data"`
}
