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

package main

import (
	"context"
	"net/http"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/sapcc/backup-tools/internal/api"
	"github.com/sapcc/backup-tools/internal/backup"
	"github.com/sapcc/backup-tools/internal/core"
	"github.com/sapcc/go-api-declarations/bininfo"
	"github.com/sapcc/go-bits/httpapi"
	"github.com/sapcc/go-bits/httpext"
	"github.com/sapcc/go-bits/logg"
	"github.com/sapcc/go-bits/must"
)

func main() {
	wrap := httpext.WrapTransport(&http.DefaultTransport)
	wrap.SetOverrideUserAgent(bininfo.Component(), bininfo.VersionOr("unknown"))

	ctx := httpext.ContextWithSIGINT(context.Background(), 1*time.Second)
	cfg := must.Return(core.NewConfiguration())

	//listen to SIGUSR1 (this signal causes a backup to be created immediately, regardless of schedule)
	signalChan := make(chan os.Signal, 1)
	signal.Notify(signalChan, syscall.SIGUSR1)

	//fork off the main loop
	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		ticker := time.NewTicker(1 * time.Minute)
		defer ticker.Stop()
		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				err := backup.CreateIfNecessary(cfg)
				if err != nil {
					logg.Error(err.Error())
				}
			case <-signalChan:
				_, err := backup.Create(cfg, "because of user request (SIGUSR1)")
				if err != nil {
					logg.Error(err.Error())
				}
			}
		}
	}()

	//serve Prometheus metrics on another goroutine (this needs to be separate
	//from the rest of the HTTP API because the metrics port is exposed to
	//outside the container)
	go func() {
		mux := http.NewServeMux()
		mux.Handle("/metrics", promhttp.Handler())
		must.Succeed(httpext.ListenAndServeContext(ctx, ":9188", mux))
	}()

	//serve the HTTP API on the main thread
	//
	//NOTE: This API does not do any authentication at all, and that's okay
	//because it listens on 127.0.0.1 only. Therefore you can only access it via
	//`kubectl exec` or `kubectl port-forward`.
	handler := httpapi.Compose(
		api.API{Config: cfg},
		httpapi.HealthCheckAPI{SkipRequestLog: true},
	)
	must.Succeed(httpext.ListenAndServeContext(ctx, "127.0.0.1:8080", handler))

	//on SIGINT/SIGTERM, give the backup main loop a chance to complete a backup
	//that's currently in flight
	wg.Wait()
}
