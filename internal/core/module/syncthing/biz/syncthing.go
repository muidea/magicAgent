package biz

import (
	"fmt"
	"net/http"
	"net/url"
	"path"
	"strings"
	"time"

	cd "github.com/muidea/magicCommon/def"
	"github.com/muidea/magicCommon/event"
	"github.com/muidea/magicCommon/foundation/log"
	"github.com/muidea/magicCommon/foundation/net"

	"github.com/muidea/magieAgent/internal/config"
	"github.com/muidea/magieAgent/pkg/common"
)

type DeviceList []any
type FoldersConfig map[string]any

/*
	func (s folderState) String() string {
		switch s {
		case FolderIdle:
			return "idle"
		case FolderScanning:
			return "scanning"
		case FolderScanWaiting:
			return "scan-waiting"
		case FolderSyncWaiting:
			return "sync-waiting"
		case FolderSyncPreparing:
			return "sync-preparing"
		case FolderSyncing:
			return "syncing"
		case FolderCleaning:
			return "cleaning"
		case FolderCleanWaiting:
			return "clean-waiting"
		case FolderError:
			return "error"
		default:
			return "unknown"
		}
	}
*/
type PathStatus struct {
	State        string `json:"state"`
	StateChanged string `json:"stateChanged"`
}

func (s *PathStatus) Final() bool {
	return strings.Index(s.State, "ing") == -1
}

func (s *PathStatus) TimeStamp() (ret time.Time, err error) {
	ret, err = time.ParseInLocation(time.RFC3339, s.StateChanged, time.Local)
	return
}

const devicesTag = "devices"
const idTag = "id"
const labelTag = "label"
const pathTag = "path"

func (s *Syncthing) checkSyncPathConfig(serverIP, pathID, pathName string) (err *cd.Result) {
	headerValues := url.Values{}
	headerValues.Set(common.APIKey, config.GetSyncSecret())

	urlVal, _ := url.ParseRequestURI(fmt.Sprintf("http://%s:%d/rest/config/folders/%s", serverIP, common.DefaultSyncthingPort, pathID))
	client := &http.Client{}
	defer client.CloseIdleConnections()
	_, httpErr := net.HTTPGet(client, urlVal.String(), nil, headerValues)
	if httpErr == nil {
		return
	}

	deviceList := DeviceList{}
	urlVal, _ = url.ParseRequestURI(fmt.Sprintf("http://%s:%d/rest/config/devices", serverIP, common.DefaultSyncthingPort))
	_, httpErr = net.HTTPGet(client, urlVal.String(), &deviceList, headerValues)
	if httpErr != nil {
		err = cd.NewError(cd.UnExpected, httpErr.Error())
		log.Errorf("checkSyncPathConfig failed, fetch syncthing devices error:%s", httpErr.Error())
		return
	}

	folder := FoldersConfig{}
	urlVal, _ = url.ParseRequestURI(fmt.Sprintf("http://%s:%d/rest/config/defaults/folder", serverIP, common.DefaultSyncthingPort))
	_, httpErr = net.HTTPGet(client, urlVal.String(), &folder, headerValues)
	if httpErr != nil {
		err = cd.NewError(cd.UnExpected, httpErr.Error())
		log.Errorf("checkSyncPathConfig failed, fetch syncthing default folder error:%s", httpErr.Error())
		return
	}
	folder[devicesTag] = deviceList
	folder[idTag] = pathID
	folder[labelTag] = pathID
	folder[pathTag] = path.Join(config.GetTmpPath(), pathName)

	urlVal, _ = url.ParseRequestURI(fmt.Sprintf("http://%s:%d/rest/config/folders", serverIP, common.DefaultSyncthingPort))
	_, httpErr = net.HTTPPost(client, urlVal.String(), folder, nil, headerValues)
	if httpErr != nil {
		log.Errorf("checkSyncPathConfig failed, add new syncthing folder error:%s", httpErr.Error())
		return
	}

	return
}

func (s *Syncthing) checkSyncPathStatus(serverIP, pathID string) (err *cd.Result) {
	headerValues := url.Values{}
	headerValues.Set(common.APIKey, config.GetSyncSecret())

	client := &http.Client{}
	defer client.CloseIdleConnections()
	queryValues := url.Values{}
	queryValues.Set("folder", pathID)

	result := &PathStatus{}
	urlVal, _ := url.ParseRequestURI(fmt.Sprintf("http://%s:%d/rest/db/status", serverIP, common.DefaultSyncthingPort))
	urlVal.RawQuery = queryValues.Encode()
	_, httpErr := net.HTTPGet(client, urlVal.String(), result, headerValues)
	if httpErr != nil {
		err = cd.NewError(cd.UnExpected, httpErr.Error())
		log.Errorf("checkSyncPathStatus failed, check %s previous status error:%s", pathID, httpErr.Error())
		return
	}

	previousTime, previousErr := result.TimeStamp()
	if previousErr != nil {
		err = cd.NewError(cd.UnExpected, httpErr.Error())
		log.Errorf("checkSyncPathStatus failed, check %s previous status error:%s", pathID, previousErr.Error())
		return
	}

	urlVal, _ = url.ParseRequestURI(fmt.Sprintf("http://%s:%d/rest/db/scan", serverIP, common.DefaultSyncthingPort))
	urlVal.RawQuery = queryValues.Encode()
	_, httpErr = net.HTTPPost(client, urlVal.String(), nil, nil, headerValues)
	if httpErr != nil {
		err = cd.NewError(cd.UnExpected, httpErr.Error())
		log.Errorf("checkSyncPathStatus failed, scan path %s error:%s", pathID, httpErr.Error())
		return
	}

	urlVal, _ = url.ParseRequestURI(fmt.Sprintf("http://%s:%d/rest/db/status", serverIP, common.DefaultSyncthingPort))
	urlVal.RawQuery = queryValues.Encode()
	for {
		log.Warnf("checkSyncPathStatus, path:%s", pathID)
		_, httpErr = net.HTTPGet(client, urlVal.String(), result, headerValues)
		if httpErr != nil {
			err = cd.NewError(cd.UnExpected, httpErr.Error())
			log.Errorf("checkSyncPathStatus failed, check path status error:%s", httpErr.Error())
			return
		}

		statusTime, statusErr := result.TimeStamp()
		if statusErr != nil {
			err = cd.NewError(cd.UnExpected, httpErr.Error())
			log.Errorf("checkSyncPathStatus failed, check path status error:%s", statusErr.Error())
			return
		}

		elapse := statusTime.Sub(previousTime)
		log.Infof("previousTime:%v, statusTime:%v, elapse:%v", previousTime, statusTime, elapse)
		if elapse < time.Second {
			time.Sleep(time.Second * 5)
			continue
		}

		if result.Final() {
			break
		}

		if elapse > time.Hour {
			err = cd.NewError(cd.UnExpected, "sync data timeout")
			log.Errorf("checkSyncPathStatus failed, check path status error:%s", err.Error())
			break
		}
	}

	return
}

func (s *Syncthing) SyncFilesToRemote(ev event.Event, re event.Result) {
	param := ev.Data()
	if param == nil {
		log.Warnf("SyncFilesToRemote failed, nil param")
		return
	}

	syncInfoPtr, syncInfoOK := param.(*common.SyncInfo)
	if !syncInfoOK || syncInfoPtr == nil {
		log.Warnf("SyncFilesToRemote failed, illegal param")
		return
	}

	log.Infof("SyncFilesToRemote, syncInfo:%v", syncInfoPtr)

	err := s.checkSyncPathConfig(config.GetRemoteHost(), syncInfoPtr.Name, syncInfoPtr.Remote)
	if err != nil {
		log.Errorf("SyncFilesToRemote failed, s.checkSyncPathConfig error:%s", err.Error())
		if re != nil {
			re.Set(nil, err)
		}
		return
	}

	err = s.checkSyncPathConfig(config.GetLocalHost(), syncInfoPtr.Name, syncInfoPtr.Local)
	if err != nil {
		log.Errorf("SyncFilesToRemote failed, s.checkSyncPathConfig error:%s", err.Error())
		if re != nil {
			re.Set(nil, err)
		}
		return
	}

	err = s.checkSyncPathStatus(config.GetRemoteHost(), syncInfoPtr.Name)
	if err != nil {
		log.Errorf("SyncFilesToRemote failed, s.checkSyncPathStatus error:%s", err.Error())
	}

	if re != nil {
		re.Set(nil, err)
	}
	return
}

func (s *Syncthing) SyncFilesToLocal(ev event.Event, re event.Result) {
	param := ev.Data()
	if param == nil {
		log.Warnf("SyncFilesToLocal failed, nil param")
		return
	}

	syncInfoPtr, syncInfoOK := param.(*common.SyncInfo)
	if !syncInfoOK || syncInfoPtr == nil {
		log.Warnf("SyncFilesToLocal failed, illegal param")
		return
	}

	log.Infof("SyncFilesToLocal, syncInfo:%v", syncInfoPtr)

	err := s.checkSyncPathConfig(config.GetLocalHost(), syncInfoPtr.Name, syncInfoPtr.Local)
	if err != nil {
		log.Errorf("SyncFilesToLocal failed, s.checkSyncPathConfig error:%s", err.Error())
		if re != nil {
			re.Set(nil, err)
		}
		return
	}

	err = s.checkSyncPathConfig(config.GetRemoteHost(), syncInfoPtr.Name, syncInfoPtr.Remote)
	if err != nil {
		log.Errorf("SyncFilesToLocal failed, s.checkSyncPathConfig error:%s", err.Error())
		if re != nil {
			re.Set(nil, err)
		}
		return
	}

	err = s.checkSyncPathStatus(config.GetLocalHost(), syncInfoPtr.Name)
	if err != nil {
		log.Errorf("SyncFilesToLocal failed, s.checkSyncPathStatus error:%s", err.Error())
	}

	if re != nil {
		re.Set(nil, err)
	}
	return
}
