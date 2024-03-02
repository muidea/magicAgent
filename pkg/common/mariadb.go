package common

import cd "github.com/muidea/magicCommon/def"

const (
	QueryStatus = "/status/query"
)

/*
Primary: 表示当前节点是集群的主节点（Primary）。这是一个活跃的写入节点，处理所有写入操作，并将更改复制到其他节点。

Non-Primary: 表示当前节点是集群的非主节点（Non-Primary）。这是一个只读节点，用于处理读取请求，无法处理写入操作。

Joining: 表示当前节点正在加入集群。这可能发生在节点刚刚启动或重新加入集群时。

Joined: 表示当前节点已成功加入集群。

Synced: 表示当前节点已与集群同步，并可以参与写入和读取操作。

Donor/Desynced: 表示当前节点处于脱离集群状态或作为捐赠者（Donor）节点。脱离集群状态可能是由于节点启动、节点故障或手动操作引起的。

Undefined: 表示当前节点的集群状态未定义或无法确定。
*/
const (
	Primary    = "Primary"
	NonPrimary = "Non-Primary"
	Joining    = "Joining"
	Joined     = "Joined"
	Synced     = "Synced"
	Donor      = "Donor"
	Desynced   = "Desynced"
	Undefined  = "Undefined"
)

const MariadbModule = "/module/mariadb"

type ClusterStatus struct {
	Nodes    []string `json:"nodes"`
	NodeSize int      `json:"nodeSize"`
	Status   string   `json:"status"`
}

func (s *ClusterStatus) IsNormal() bool {
	switch s.Status {
	case Primary, Joining, Joined, Synced, Donor, Desynced:
		return true
	}

	return false
}

type QueryClusterStatusResult struct {
	cd.Result
	Status *ClusterStatus `json:"status"`
}
