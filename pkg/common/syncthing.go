package common

import "fmt"

const (
	DefaultSyncthingPort = 8384
)

const (
	SyncFilesToRemote = "/files/sync/remote"
	SyncFilesToLocal  = "/files/sync/local"
)

type SyncInfo struct {
	Name   string `json:"name"`
	Local  string `json:"local"`
	Remote string `json:"remote"`
}

func (s *SyncInfo) String() string {
	return fmt.Sprintf("local:%s,remote:%s", s.Local, s.Remote)
}

const SyncthingModule = "/module/syncthing"
