package biz

import (
	"bufio"
	"bytes"
	"strconv"
	"strings"

	cd "github.com/muidea/magicCommon/def"
	"github.com/muidea/magicCommon/event"
	"github.com/muidea/magicCommon/foundation/log"
	"github.com/muidea/magicCommon/task"

	"github.com/muidea/magicAgent/internal/core/base/biz"
	"github.com/muidea/magicAgent/pkg/common"
)

type Mariadb struct {
	biz.Base
}

func New(
	eventHub event.Hub,
	backgroundRoutine task.BackgroundRoutine,
) *Mariadb {
	ptr := &Mariadb{
		Base: biz.New(common.MariadbModule, eventHub, backgroundRoutine),
	}

	ptr.SubscribeFunc(common.QueryStatus, ptr.queryStatus)

	return ptr
}

func (s *Mariadb) queryStatus(ev event.Event, re event.Result) {
	serviceVal := ev.Data()
	if serviceVal == nil {
		return
	}

	statusPtr, statusErr := s.QueryMariadbClusterStatus(serviceVal.(string))
	if re != nil {
		re.Set(statusPtr, statusErr)
	}
}

const wsrepIncomingAddresses = "wsrep_incoming_addresses"
const wsrepClusterSize = "wsrep_cluster_size"
const wsrepClusterStatus = "wsrep_cluster_status"

func (s *Mariadb) QueryMariadbClusterStatus(serviceName string) (ret *common.ClusterStatus, err *cd.Result) {
	param := &common.ServiceParam{
		Service:  serviceName,
		CmdParam: "mysql -uroot -prootkit -e\"show status like '%wsrep%';\"",
	}

	execEvent := event.NewEvent(common.ExecuteCommand, s.ID(), common.DockerModule, nil, param)
	result := s.SendEvent(execEvent)
	execVal, execErr := result.Get()
	if execErr != nil {
		log.Errorf("queryMariadbCluster failed, error:%s", execErr.Error())
		err = execErr
		return
	}

	statusPtr := &common.ClusterStatus{}
	scanner := bufio.NewScanner(bytes.NewReader(execVal.([]byte)))
	for scanner.Scan() {
		items := strings.Split(scanner.Text(), "\t")
		if len(items) != 2 {
			continue
		}

		switch strings.ToLower(items[0]) {
		case wsrepIncomingAddresses:
			statusPtr.Nodes = strings.Split(items[1], ",")
		case wsrepClusterSize:
			statusPtr.NodeSize, _ = strconv.Atoi(items[1])
		case wsrepClusterStatus:
			statusPtr.Status = items[1]
		default:
		}
	}

	ret = statusPtr
	return
}
