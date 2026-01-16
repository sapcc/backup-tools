// SPDX-FileCopyrightText: 2023 SAP SE or an SAP affiliate company
// SPDX-License-Identifier: Apache-2.0

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
	"github.com/sapcc/go-api-declarations/bininfo"
	"github.com/sapcc/go-bits/httpapi"
	"github.com/sapcc/go-bits/httpext"
	"github.com/sapcc/go-bits/logg"
	"github.com/sapcc/go-bits/must"
	"github.com/sapcc/go-bits/osext"

	"github.com/sapcc/backup-tools/internal/api"
	"github.com/sapcc/backup-tools/internal/backup"
	"github.com/sapcc/backup-tools/internal/core"
)

func main() {
	bininfo.HandleVersionArgument()

	logg.ShowDebug = osext.GetenvBool("BACKUP_TOOLS_DEBUG")

	wrap := httpext.WrapTransport(&http.DefaultTransport)
	wrap.SetOverrideUserAgent(bininfo.Component(), bininfo.VersionOr("unknown"))

	ctx := httpext.ContextWithSIGINT(context.Background(), 1*time.Second)
	cfg := must.Return(core.NewConfiguration(ctx))

	// listen to SIGUSR1 (this signal causes a backup to be created immediately, regardless of schedule)
	signalChan := make(chan os.Signal, 1)
	signal.Notify(signalChan, syscall.SIGUSR1)

	// fork off the main loop
	var wg sync.WaitGroup
	wg.Go(func() {
		ticker := time.NewTicker(1 * time.Minute)
		defer ticker.Stop()
		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				err := backup.CreateIfNecessary(ctx, cfg)
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
	})

	// serve Prometheus metrics on another goroutine (this needs to be separate
	// from the rest of the HTTP API because the metrics port is exposed to
	// outside the container)
	go func() {
		mux := http.NewServeMux()
		mux.Handle("/metrics", promhttp.Handler())
		must.Succeed(httpext.ListenAndServeContext(ctx, ":9188", mux))
	}()

	// serve the HTTP API on the main thread
	//
	//NOTE: This API does not do any authentication at all, and that's okay
	// because it listens on 127.0.0.1 only. Therefore you can only access it via
	// `kubectl exec` or `kubectl port-forward`.
	handler := httpapi.Compose(
		api.API{Config: cfg},
		httpapi.HealthCheckAPI{SkipRequestLog: true},
	)
	must.Succeed(httpext.ListenAndServeContext(ctx, "0.0.0.0:8080", handler))

	// on SIGINT/SIGTERM, give the backup main loop a chance to complete a backup
	// that's currently in flight
	wg.Wait()
}
