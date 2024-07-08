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
	"crypto/subtle"
	"encoding/json"
	"io"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/sapcc/go-bits/httpapi"
	"github.com/sapcc/go-bits/respondwith"

	"github.com/sapcc/backup-tools/internal/backup"
	"github.com/sapcc/backup-tools/internal/core"
	"github.com/sapcc/backup-tools/internal/restore"
)

// API contains the HTTP request handlers for the backup-server.
type API struct {
	Config *core.Configuration
}

func requireBasicAuth(handler http.HandlerFunc, username, password string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		user, pass, ok := r.BasicAuth()

		if !ok || subtle.ConstantTimeCompare([]byte(user), []byte(username)) != 1 || subtle.ConstantTimeCompare([]byte(pass), []byte(password)) != 1 {
			w.Header().Set("WWW-Authenticate", `Basic realm="Authentication required"`)
			http.Error(w, "Unauthorized.", http.StatusUnauthorized)
			return
		}

		handler(w, r)
	}
}

// AddTo implements the httpapi.API interface.
func (a API) AddTo(r *mux.Router) {
	r.Methods("GET").Path("/v1/status").HandlerFunc(requireBasicAuth(a.handleGetStatus, a.Config.PgUsername, a.Config.PgPassword))
	r.Methods("POST").Path("/v1/backup-now").HandlerFunc(requireBasicAuth(a.handlePostBackupNow, a.Config.PgUsername, a.Config.PgPassword))
	r.Methods("GET").Path("/v1/backups").HandlerFunc(requireBasicAuth(a.handleGetBackups, a.Config.PgUsername, a.Config.PgPassword))
	r.Methods("POST").Path("/v1/restore/{id}").HandlerFunc(requireBasicAuth(a.handlePostRestore, a.Config.PgUsername, a.Config.PgPassword))
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

	lastTime, err := backup.ReadLastBackupTimestamp(r.Context(), a.Config)
	if respondwith.ErrorText(w, err) {
		return
	}

	respondwith.JSON(w, http.StatusOK, statusResponse{
		LastBackup: timestamp{
			Timestamp: lastTime.UTC().Format(backup.TimeFormat),
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
		return
	}

	respondwith.JSON(w, http.StatusOK, timestamp{
		Timestamp: backupTime.UTC().Format(backup.TimeFormat),
		Unix:      backupTime.Unix(),
	})
}

////////////////////////////////////////////////////////////////////////////////
// GET /v1/backups

func (a API) handleGetBackups(w http.ResponseWriter, r *http.Request) {
	httpapi.IdentifyEndpoint(r, "/v1/backups")

	backups, err := restore.ListRestorableBackups(r.Context(), a.Config)
	if respondwith.ErrorText(w, err) {
		return
	}

	if backups == nil {
		backups = []*restore.RestorableBackup{} // ensure that JSON contains "[]" instead of "null"
	}
	respondwith.JSON(w, http.StatusOK, backups)
}

////////////////////////////////////////////////////////////////////////////////
// POST /v1/restore

func (a API) handlePostRestore(w http.ResponseWriter, r *http.Request) {
	httpapi.IdentifyEndpoint(r, "/v1/restore/:id")

	// superuser credentials may be supplied in the request body
	buf, err := io.ReadAll(r.Body)
	if respondwith.ErrorText(w, err) {
		return
	}
	var req struct {
		SuperUser *restore.SuperUserCredentials `json:"superuser"`
	}
	if len(buf) > 0 {
		err = json.Unmarshal(buf, &req)
		if respondwith.ErrorText(w, err) {
			return
		}
	}

	// find backup
	backups, err := restore.ListRestorableBackups(r.Context(), a.Config)
	if respondwith.ErrorText(w, err) {
		return
	}
	bkp := backups.FindByID(mux.Vars(r)["id"])
	if bkp == nil {
		http.Error(w, "no such backup", http.StatusNotFound)
		return
	}

	// run restore
	err = bkp.Restore(r.Context(), a.Config, req.SuperUser)
	if err == nil {
		http.Error(w, "backup restored successfully", http.StatusOK)
	} else {
		http.Error(w, "restore failed (check the pgbackup container log for details. Did you scale down your application?): "+err.Error(), http.StatusInternalServerError)
	}
}
