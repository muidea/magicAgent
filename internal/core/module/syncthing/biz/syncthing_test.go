package biz

import (
	"encoding/json"
	"github.com/muidea/magicCommon/foundation/util"
	"testing"
)

var testConfig = `
{
    "version": 37,
    "folders": [
        {
            "id": "folderDemo",
            "label": "folderDemo",
            "filesystemType": "basic",
            "path": "/backup/folderDemo",
            "type": "sendreceive",
            "devices": [
                {
                    "deviceID": "L2UEHLB-GDNZDKU-CB3TZRK-MRN6ALI-HENHGNV-KZQG2KO-2YS7LLQ-WI7HIQY",
                    "introducedBy": "",
                    "encryptionPassword": ""
                },
                {
                    "deviceID": "XJR4T2L-JCQYJ2B-NWXLPLZ-CTUL46Z-I4BYAX2-TVER7BP-3MQ7PW5-KV325QA",
                    "introducedBy": "",
                    "encryptionPassword": ""
                }
            ],
            "rescanIntervalS": 3600,
            "fsWatcherEnabled": true,
            "fsWatcherDelayS": 10,
            "ignorePerms": false,
            "autoNormalize": true,
            "minDiskFree": {
                "value": 1,
                "unit": "%"
            },
            "versioning": {
                "type": "",
                "params": {},
                "cleanupIntervalS": 3600,
                "fsPath": "",
                "fsType": "basic"
            },
            "copiers": 0,
            "pullerMaxPendingKiB": 0,
            "hashers": 0,
            "order": "random",
            "ignoreDelete": false,
            "scanProgressIntervalS": 0,
            "pullerPauseS": 0,
            "maxConflicts": 10,
            "disableSparseFiles": false,
            "disableTempIndexes": false,
            "paused": false,
            "weakHashThresholdPct": 25,
            "markerName": ".stfolder",
            "copyOwnershipFromParent": false,
            "modTimeWindowS": 0,
            "maxConcurrentWrites": 2,
            "disableFsync": false,
            "blockPullOrder": "standard",
            "copyRangeMethod": "standard",
            "caseSensitiveFS": false,
            "junctionsAsDirs": false,
            "syncOwnership": false,
            "sendOwnership": false,
            "syncXattrs": false,
            "sendXattrs": false,
            "xattrFilter": {
                "entries": [],
                "maxSingleEntrySize": 1024,
                "maxTotalSize": 4096
            }
        },
        {
            "id": "database/mariadb100",
            "label": "database/mariadb100",
            "filesystemType": "basic",
            "path": "/backup/database/mariadb100",
            "type": "sendreceive",
            "devices": [
                {
                    "deviceID": "L2UEHLB-GDNZDKU-CB3TZRK-MRN6ALI-HENHGNV-KZQG2KO-2YS7LLQ-WI7HIQY",
                    "introducedBy": "",
                    "encryptionPassword": ""
                },
                {
                    "deviceID": "XJR4T2L-JCQYJ2B-NWXLPLZ-CTUL46Z-I4BYAX2-TVER7BP-3MQ7PW5-KV325QA",
                    "introducedBy": "",
                    "encryptionPassword": ""
                }
            ],
            "rescanIntervalS": 3600,
            "fsWatcherEnabled": true,
            "fsWatcherDelayS": 10,
            "ignorePerms": false,
            "autoNormalize": true,
            "minDiskFree": {
                "value": 1,
                "unit": "%"
            },
            "versioning": {
                "type": "",
                "params": {},
                "cleanupIntervalS": 3600,
                "fsPath": "",
                "fsType": "basic"
            },
            "copiers": 0,
            "pullerMaxPendingKiB": 0,
            "hashers": 0,
            "order": "random",
            "ignoreDelete": false,
            "scanProgressIntervalS": 0,
            "pullerPauseS": 0,
            "maxConflicts": 10,
            "disableSparseFiles": false,
            "disableTempIndexes": false,
            "paused": false,
            "weakHashThresholdPct": 25,
            "markerName": ".stfolder",
            "copyOwnershipFromParent": false,
            "modTimeWindowS": 0,
            "maxConcurrentWrites": 2,
            "disableFsync": false,
            "blockPullOrder": "standard",
            "copyRangeMethod": "standard",
            "caseSensitiveFS": false,
            "junctionsAsDirs": false,
            "syncOwnership": false,
            "sendOwnership": false,
            "syncXattrs": false,
            "sendXattrs": false,
            "xattrFilter": {
                "entries": [],
                "maxSingleEntrySize": 1024,
                "maxTotalSize": 4096
            }
        },
        {
            "id": "database/mariadb001",
            "label": "database/mariadb001",
            "filesystemType": "basic",
            "path": "/backup/database/mariadb001",
            "type": "sendreceive",
            "devices": [
                {
                    "deviceID": "L2UEHLB-GDNZDKU-CB3TZRK-MRN6ALI-HENHGNV-KZQG2KO-2YS7LLQ-WI7HIQY",
                    "introducedBy": "",
                    "encryptionPassword": ""
                },
                {
                    "deviceID": "XJR4T2L-JCQYJ2B-NWXLPLZ-CTUL46Z-I4BYAX2-TVER7BP-3MQ7PW5-KV325QA",
                    "introducedBy": "",
                    "encryptionPassword": ""
                }
            ],
            "rescanIntervalS": 3600,
            "fsWatcherEnabled": true,
            "fsWatcherDelayS": 10,
            "ignorePerms": false,
            "autoNormalize": true,
            "minDiskFree": {
                "value": 1,
                "unit": "%"
            },
            "versioning": {
                "type": "",
                "params": {},
                "cleanupIntervalS": 3600,
                "fsPath": "",
                "fsType": "basic"
            },
            "copiers": 0,
            "pullerMaxPendingKiB": 0,
            "hashers": 0,
            "order": "random",
            "ignoreDelete": false,
            "scanProgressIntervalS": 0,
            "pullerPauseS": 0,
            "maxConflicts": 10,
            "disableSparseFiles": false,
            "disableTempIndexes": false,
            "paused": false,
            "weakHashThresholdPct": 25,
            "markerName": ".stfolder",
            "copyOwnershipFromParent": false,
            "modTimeWindowS": 0,
            "maxConcurrentWrites": 2,
            "disableFsync": false,
            "blockPullOrder": "standard",
            "copyRangeMethod": "standard",
            "caseSensitiveFS": false,
            "junctionsAsDirs": false,
            "syncOwnership": false,
            "sendOwnership": false,
            "syncXattrs": false,
            "sendXattrs": false,
            "xattrFilter": {
                "entries": [],
                "maxSingleEntrySize": 1024,
                "maxTotalSize": 4096
            }
        },
        {
            "id": "database/mariadb002",
            "label": "database/mariadb002",
            "filesystemType": "basic",
            "path": "/backup/database/mariadb002",
            "type": "sendreceive",
            "devices": [
                {
                    "deviceID": "L2UEHLB-GDNZDKU-CB3TZRK-MRN6ALI-HENHGNV-KZQG2KO-2YS7LLQ-WI7HIQY",
                    "introducedBy": "",
                    "encryptionPassword": ""
                },
                {
                    "deviceID": "XJR4T2L-JCQYJ2B-NWXLPLZ-CTUL46Z-I4BYAX2-TVER7BP-3MQ7PW5-KV325QA",
                    "introducedBy": "",
                    "encryptionPassword": ""
                }
            ],
            "rescanIntervalS": 3600,
            "fsWatcherEnabled": true,
            "fsWatcherDelayS": 10,
            "ignorePerms": false,
            "autoNormalize": true,
            "minDiskFree": {
                "value": 1,
                "unit": "%"
            },
            "versioning": {
                "type": "",
                "params": {},
                "cleanupIntervalS": 3600,
                "fsPath": "",
                "fsType": "basic"
            },
            "copiers": 0,
            "pullerMaxPendingKiB": 0,
            "hashers": 0,
            "order": "random",
            "ignoreDelete": false,
            "scanProgressIntervalS": 0,
            "pullerPauseS": 0,
            "maxConflicts": 10,
            "disableSparseFiles": false,
            "disableTempIndexes": false,
            "paused": false,
            "weakHashThresholdPct": 25,
            "markerName": ".stfolder",
            "copyOwnershipFromParent": false,
            "modTimeWindowS": 0,
            "maxConcurrentWrites": 2,
            "disableFsync": false,
            "blockPullOrder": "standard",
            "copyRangeMethod": "standard",
            "caseSensitiveFS": false,
            "junctionsAsDirs": false,
            "syncOwnership": false,
            "sendOwnership": false,
            "syncXattrs": false,
            "sendXattrs": false,
            "xattrFilter": {
                "entries": [],
                "maxSingleEntrySize": 1024,
                "maxTotalSize": 4096
            }
        },
        {
            "id": "database/mariadb003",
            "label": "database/mariadb003",
            "filesystemType": "basic",
            "path": "/backup/database/mariadb003",
            "type": "sendreceive",
            "devices": [
                {
                    "deviceID": "L2UEHLB-GDNZDKU-CB3TZRK-MRN6ALI-HENHGNV-KZQG2KO-2YS7LLQ-WI7HIQY",
                    "introducedBy": "",
                    "encryptionPassword": ""
                },
                {
                    "deviceID": "XJR4T2L-JCQYJ2B-NWXLPLZ-CTUL46Z-I4BYAX2-TVER7BP-3MQ7PW5-KV325QA",
                    "introducedBy": "",
                    "encryptionPassword": ""
                }
            ],
            "rescanIntervalS": 3600,
            "fsWatcherEnabled": true,
            "fsWatcherDelayS": 10,
            "ignorePerms": false,
            "autoNormalize": true,
            "minDiskFree": {
                "value": 1,
                "unit": "%"
            },
            "versioning": {
                "type": "",
                "params": {},
                "cleanupIntervalS": 3600,
                "fsPath": "",
                "fsType": "basic"
            },
            "copiers": 0,
            "pullerMaxPendingKiB": 0,
            "hashers": 0,
            "order": "random",
            "ignoreDelete": false,
            "scanProgressIntervalS": 0,
            "pullerPauseS": 0,
            "maxConflicts": 10,
            "disableSparseFiles": false,
            "disableTempIndexes": false,
            "paused": false,
            "weakHashThresholdPct": 25,
            "markerName": ".stfolder",
            "copyOwnershipFromParent": false,
            "modTimeWindowS": 0,
            "maxConcurrentWrites": 2,
            "disableFsync": false,
            "blockPullOrder": "standard",
            "copyRangeMethod": "standard",
            "caseSensitiveFS": false,
            "junctionsAsDirs": false,
            "syncOwnership": false,
            "sendOwnership": false,
            "syncXattrs": false,
            "sendXattrs": false,
            "xattrFilter": {
                "entries": [],
                "maxSingleEntrySize": 1024,
                "maxTotalSize": 4096
            }
        },
        {
            "id": "default",
            "label": "Default Folder",
            "filesystemType": "basic",
            "path": "/config/Sync",
            "type": "sendreceive",
            "devices": [
                {
                    "deviceID": "XJR4T2L-JCQYJ2B-NWXLPLZ-CTUL46Z-I4BYAX2-TVER7BP-3MQ7PW5-KV325QA",
                    "introducedBy": "",
                    "encryptionPassword": ""
                }
            ],
            "rescanIntervalS": 3600,
            "fsWatcherEnabled": true,
            "fsWatcherDelayS": 10,
            "ignorePerms": false,
            "autoNormalize": true,
            "minDiskFree": {
                "value": 1,
                "unit": "%"
            },
            "versioning": {
                "type": "",
                "params": {},
                "cleanupIntervalS": 3600,
                "fsPath": "",
                "fsType": "basic"
            },
            "copiers": 0,
            "pullerMaxPendingKiB": 0,
            "hashers": 0,
            "order": "random",
            "ignoreDelete": false,
            "scanProgressIntervalS": 0,
            "pullerPauseS": 0,
            "maxConflicts": 10,
            "disableSparseFiles": false,
            "disableTempIndexes": false,
            "paused": false,
            "weakHashThresholdPct": 25,
            "markerName": ".stfolder",
            "copyOwnershipFromParent": false,
            "modTimeWindowS": 0,
            "maxConcurrentWrites": 2,
            "disableFsync": false,
            "blockPullOrder": "standard",
            "copyRangeMethod": "standard",
            "caseSensitiveFS": false,
            "junctionsAsDirs": false,
            "syncOwnership": false,
            "sendOwnership": false,
            "syncXattrs": false,
            "sendXattrs": false,
            "xattrFilter": {
                "entries": [],
                "maxSingleEntrySize": 1024,
                "maxTotalSize": 4096
            }
        }
    ],
    "devices": [
        {
            "deviceID": "L2UEHLB-GDNZDKU-CB3TZRK-MRN6ALI-HENHGNV-KZQG2KO-2YS7LLQ-WI7HIQY",
            "name": "dlake-5",
            "addresses": [
                "dynamic"
            ],
            "compression": "metadata",
            "certName": "",
            "introducer": false,
            "skipIntroductionRemovals": false,
            "introducedBy": "",
            "paused": false,
            "allowedNetworks": [],
            "autoAcceptFolders": true,
            "maxSendKbps": 0,
            "maxRecvKbps": 0,
            "ignoredFolders": [],
            "maxRequestKiB": 0,
            "untrusted": false,
            "remoteGUIPort": 0,
            "numConnections": 0
        },
        {
            "deviceID": "XJR4T2L-JCQYJ2B-NWXLPLZ-CTUL46Z-I4BYAX2-TVER7BP-3MQ7PW5-KV325QA",
            "name": "dlake-4",
            "addresses": [
                "dynamic"
            ],
            "compression": "metadata",
            "certName": "",
            "introducer": false,
            "skipIntroductionRemovals": false,
            "introducedBy": "",
            "paused": false,
            "allowedNetworks": [],
            "autoAcceptFolders": false,
            "maxSendKbps": 0,
            "maxRecvKbps": 0,
            "ignoredFolders": [],
            "maxRequestKiB": 0,
            "untrusted": false,
            "remoteGUIPort": 0,
            "numConnections": 0
        }
    ],
    "gui": {
        "enabled": true,
        "address": "127.0.0.1:8384",
        "unixSocketPermissions": "",
        "user": "",
        "password": "",
        "authMode": "static",
        "useTLS": false,
        "apiKey": "qC6GuA7Hjyzk3CQRQ9eoMGNHCcFFw5QF",
        "insecureAdminAccess": false,
        "theme": "default",
        "debugging": false,
        "insecureSkipHostcheck": false,
        "insecureAllowFrameLoading": false,
        "sendBasicAuthPrompt": false
    },
    "ldap": {
        "address": "",
        "bindDN": "",
        "transport": "plain",
        "insecureSkipVerify": false,
        "searchBaseDN": "",
        "searchFilter": ""
    },
    "options": {
        "listenAddresses": [
            "default"
        ],
        "globalAnnounceServers": [
            "default"
        ],
        "globalAnnounceEnabled": true,
        "localAnnounceEnabled": true,
        "localAnnouncePort": 21027,
        "localAnnounceMCAddr": "[ff12::8384]:21027",
        "maxSendKbps": 0,
        "maxRecvKbps": 0,
        "reconnectionIntervalS": 60,
        "relaysEnabled": true,
        "relayReconnectIntervalM": 10,
        "startBrowser": true,
        "natEnabled": true,
        "natLeaseMinutes": 60,
        "natRenewalMinutes": 30,
        "natTimeoutSeconds": 10,
        "urAccepted": 3,
        "urSeen": 3,
        "urUniqueId": "zTseQeM2",
        "urURL": "https://data.syncthing.net/newdata",
        "urPostInsecurely": false,
        "urInitialDelayS": 1800,
        "autoUpgradeIntervalH": 12,
        "upgradeToPreReleases": false,
        "keepTemporariesH": 24,
        "cacheIgnoredFiles": false,
        "progressUpdateIntervalS": 5,
        "limitBandwidthInLan": false,
        "minHomeDiskFree": {
            "value": 1,
            "unit": "%"
        },
        "releasesURL": "https://upgrades.syncthing.net/meta.json",
        "alwaysLocalNets": [],
        "overwriteRemoteDeviceNamesOnConnect": false,
        "tempIndexMinBlocks": 10,
        "unackedNotificationIDs": [],
        "trafficClass": 0,
        "setLowPriority": true,
        "maxFolderConcurrency": 0,
        "crURL": "https://crash.syncthing.net/newcrash",
        "crashReportingEnabled": true,
        "stunKeepaliveStartS": 180,
        "stunKeepaliveMinS": 20,
        "stunServers": [
            "default"
        ],
        "databaseTuning": "auto",
        "maxConcurrentIncomingRequestKiB": 0,
        "announceLANAddresses": true,
        "sendFullIndexOnUpgrade": false,
        "featureFlags": [],
        "connectionLimitEnough": 0,
        "connectionLimitMax": 0,
        "insecureAllowOldTLSVersions": false,
        "connectionPriorityTcpLan": 10,
        "connectionPriorityQuicLan": 20,
        "connectionPriorityTcpWan": 30,
        "connectionPriorityQuicWan": 40,
        "connectionPriorityRelay": 50,
        "connectionPriorityUpgradeThreshold": 0
    },
    "remoteIgnoredDevices": [],
    "defaults": {
        "folder": {
            "id": "",
            "label": "",
            "filesystemType": "basic",
            "path": "/backup",
            "type": "sendreceive",
            "devices": [
                {
                    "deviceID": "L2UEHLB-GDNZDKU-CB3TZRK-MRN6ALI-HENHGNV-KZQG2KO-2YS7LLQ-WI7HIQY",
                    "introducedBy": "",
                    "encryptionPassword": ""
                },
                {
                    "deviceID": "XJR4T2L-JCQYJ2B-NWXLPLZ-CTUL46Z-I4BYAX2-TVER7BP-3MQ7PW5-KV325QA",
                    "introducedBy": "",
                    "encryptionPassword": ""
                }
            ],
            "rescanIntervalS": 3600,
            "fsWatcherEnabled": true,
            "fsWatcherDelayS": 10,
            "ignorePerms": false,
            "autoNormalize": true,
            "minDiskFree": {
                "value": 1,
                "unit": "%"
            },
            "versioning": {
                "type": "",
                "params": {},
                "cleanupIntervalS": 3600,
                "fsPath": "",
                "fsType": "basic"
            },
            "copiers": 0,
            "pullerMaxPendingKiB": 0,
            "hashers": 0,
            "order": "random",
            "ignoreDelete": false,
            "scanProgressIntervalS": 0,
            "pullerPauseS": 0,
            "maxConflicts": 10,
            "disableSparseFiles": false,
            "disableTempIndexes": false,
            "paused": false,
            "weakHashThresholdPct": 25,
            "markerName": ".stfolder",
            "copyOwnershipFromParent": false,
            "modTimeWindowS": 0,
            "maxConcurrentWrites": 2,
            "disableFsync": false,
            "blockPullOrder": "standard",
            "copyRangeMethod": "standard",
            "caseSensitiveFS": false,
            "junctionsAsDirs": false,
            "syncOwnership": false,
            "sendOwnership": false,
            "syncXattrs": false,
            "sendXattrs": false,
            "xattrFilter": {
                "entries": [],
                "maxSingleEntrySize": 1024,
                "maxTotalSize": 4096
            }
        },
        "device": {
            "deviceID": "",
            "name": "",
            "addresses": [
                "dynamic"
            ],
            "compression": "metadata",
            "certName": "",
            "introducer": false,
            "skipIntroductionRemovals": false,
            "introducedBy": "",
            "paused": false,
            "allowedNetworks": [],
            "autoAcceptFolders": false,
            "maxSendKbps": 0,
            "maxRecvKbps": 0,
            "ignoredFolders": [],
            "maxRequestKiB": 0,
            "untrusted": false,
            "remoteGUIPort": 0,
            "numConnections": 0
        },
        "ignores": {
            "lines": []
        }
    }
}`

var testStatus = `
{
    "errors": 0,
    "pullErrors": 0,
    "invalid": "",
    "globalFiles": 376,
    "globalDirectories": 4,
    "globalSymlinks": 0,
    "globalDeleted": 0,
    "globalBytes": 231847626,
    "globalTotalItems": 380,
    "localFiles": 376,
    "localDirectories": 4,
    "localSymlinks": 0,
    "localDeleted": 0,
    "localBytes": 231847626,
    "localTotalItems": 380,
    "needFiles": 0,
    "needDirectories": 0,
    "needSymlinks": 0,
    "needDeletes": 0,
    "needBytes": 0,
    "needTotalItems": 0,
    "receiveOnlyChangedFiles": 0,
    "receiveOnlyChangedDirectories": 0,
    "receiveOnlyChangedSymlinks": 0,
    "receiveOnlyChangedDeletes": 0,
    "receiveOnlyChangedBytes": 0,
    "receiveOnlyTotalItems": 0,
    "inSyncFiles": 376,
    "inSyncBytes": 231847626,
    "state": "idle",
    "stateChanged": "2023-12-01T15:44:15+08:00",
    "error": "",
    "version": 588,
    "sequence": 588,
    "ignorePatterns": false,
    "watchError": ""
}
`

func TestConfig(t *testing.T) {
	syncthing := Syncthing{}
	err := syncthing.checkSyncPathConfig("127.0.0.1", "t001")
	if err != nil {
		t.Errorf("syncthing.checkSyncPathConfig failed, error:%s", err.Error())
	}
}

func TestStatus(t *testing.T) {
	statusPtr := &PathStatus{}
	err := json.Unmarshal([]byte(testStatus), statusPtr)
	if err != nil {
		t.Errorf("unmarshal config failed, error:%s", err.Error())
		return
	}

	if !statusPtr.Final() {
		t.Errorf("check final status failed")
		return
	}

	dtVal, dtErr := statusPtr.TimeStamp()
	if dtErr != nil {
		t.Errorf("check timeStamp status failed")
		return
	}

	if dtVal.Format(util.CSTLayout) != "2023-12-01 15:44:15" {
		t.Errorf("check timeStamp status failed")
		return
	}
}
