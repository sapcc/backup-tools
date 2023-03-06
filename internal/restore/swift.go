/*******************************************************************************
*
* Copyright 2023 SAP SE
*
* Licensed under the Apache License, Version 2.0 (the "License");
* you may not use this file except in compliance with the License.
* You should have received a copy of the License along with this
* program. If not, you may obtain a copy of the License at
*
*     http://www.apache.org/licenses/LICENSE-2.0
*
* Unless required by applicable law or agreed to in writing, software
* distributed under the License is distributed on an "AS IS" BASIS,
* WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
* See the License for the specific language governing permissions and
* limitations under the License.
*
*******************************************************************************/

package restore

import (
	"fmt"
	"regexp"
	"time"

	"github.com/sapcc/containers/internal/backup"
	"github.com/sapcc/containers/internal/core"
)

// RestorableBackup contains information about a backup that we found in Swift.
type RestorableBackup struct {
	ID             string   `json:"id"`
	ReadableTime   string   `json:"readable_time"`
	DatabaseNames  []string `json:"database_names"`
	TotalSizeBytes uint64   `json:"total_size_bytes"`
}

// ListRestorableBackups searches for restorable backups in Swift.
func ListRestorableBackups(cfg *core.Configuration) ([]*RestorableBackup, error) {
	iter := cfg.Container.Objects()
	iter.Prefix = cfg.ObjectNamePrefix
	objInfos, err := iter.CollectDetailed()
	if err != nil {
		return nil, err
	}

	//NOTE: ObjectNamePrefix has a trailing slash
	rx := regexp.MustCompile(fmt.Sprintf(`^%s(\d{12})/backup/pgsql/base/([^.]*)\.sql\.gz$`, regexp.QuoteMeta(cfg.ObjectNamePrefix)))
	var result []*RestorableBackup
	for _, objInfo := range objInfos {
		//skip files not matching the above pattern (this especially skips the "last_backup_timestamp")
		match := rx.FindStringSubmatch(objInfo.Object.Name())
		if match == nil {
			continue
		}
		backupTimeStr, databaseName := match[1], match[2]
		backupTime, err := time.ParseInLocation(backup.TimeFormat, backupTimeStr, time.UTC)
		if err != nil {
			continue //treat malformed timestamp as "no match"
		}

		//do we already have an entry for this backup? (this can happen when a
		//Postgres contains multiple databases, since each database is backed up
		//into a separate Swift object)
		bkp := findBackupInList(result, backupTimeStr)
		if bkp == nil {
			result = append(result, &RestorableBackup{
				ID:             backupTimeStr,
				ReadableTime:   backupTime.Format(time.RFC1123),
				DatabaseNames:  []string{databaseName},
				TotalSizeBytes: objInfo.SizeBytes,
			})
		} else {
			bkp.DatabaseNames = append(bkp.DatabaseNames, databaseName)
			bkp.TotalSizeBytes += objInfo.SizeBytes
		}
	}

	return result, nil
}

func findBackupInList(backups []*RestorableBackup, id string) *RestorableBackup {
	for _, b := range backups {
		if b.ID == id {
			return b
		}
	}
	return nil
}
