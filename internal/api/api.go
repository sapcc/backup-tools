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

package api

import (
	"net/http"

	"github.com/gorilla/mux"
	"github.com/sapcc/containers/internal/backup"
	"github.com/sapcc/containers/internal/core"
	"github.com/sapcc/go-bits/httpapi"
	"github.com/sapcc/go-bits/logg"
	"github.com/sapcc/go-bits/respondwith"
)

// API contains the HTTP request handlers for the backup-server.
type API struct {
	Config *core.Configuration
}

// AddTo implements the httpapi.API interface.
func (a API) AddTo(r *mux.Router) {
	r.Methods("GET").Path("/v1/status").HandlerFunc(a.handleGetStatus)
	r.Methods("POST").Path("/v1/backup-now").HandlerFunc(a.handlePostBackupNow)
}

////////////////////////////////////////////////////////////////////////////////
// GET /v1/status

type timestamp struct {
	Timestamp string `json:"timestamp"`
	Unix      int64  `json:"unix"`
}

type statusResponse struct {
	LastBackup timestamp `json:"last_backup"`
}

func (a API) handleGetStatus(w http.ResponseWriter, r *http.Request) {
	httpapi.IdentifyEndpoint(r, "/v1/status")

	lastTime, err := backup.ReadLastBackupTimestamp(a.Config)
	if respondwith.ErrorText(w, err) {
		return
	}

	respondwith.JSON(w, http.StatusOK, statusResponse{
		LastBackup: timestamp{
			Timestamp: lastTime.Format(backup.TimeFormat),
			Unix:      lastTime.Unix(),
		},
	})
}

////////////////////////////////////////////////////////////////////////////////
// POST /v1/backup-now

func (a API) handlePostBackupNow(w http.ResponseWriter, r *http.Request) {
	httpapi.IdentifyEndpoint(r, "/v1/backup-now")

	backupTime, err := backup.Create(a.Config, "because of user request")
	if respondwith.ErrorText(w, err) {
		logg.Error(err.Error()) //also ensure that the container log is complete
		return
	}

	respondwith.JSON(w, http.StatusOK, timestamp{
		Timestamp: backupTime.Format(backup.TimeFormat),
		Unix:      backupTime.Unix(),
	})
}
