package cmd

import (
	"context"

	messagev1 "github.com/go-goim/api/message/v1"
	"github.com/go-goim/core/pkg/cmd"
	"github.com/go-goim/core/pkg/graceful"
	"github.com/go-goim/core/pkg/log"

	"github.com/go-goim/msg-service/internal/app"
	"github.com/go-goim/msg-service/internal/service"
)

func Main() {
	if err := cmd.ParseFlags(); err != nil {
		panic(err)
	}

	application, err := app.InitApplication()
	if err != nil {
		log.Fatal("InitApplication got err", "error", err)
	}

	// register grpc
	messagev1.RegisterOfflineMessageServer(application.GrpcSrv, &service.OfflineMessageService{})

	if err = application.Run(); err != nil {
		log.Error("application run error", "error", err)
	}

	graceful.Register(application.Shutdown)
	if err = graceful.Shutdown(context.TODO()); err != nil {
		log.Error("graceful shutdown error", "error", err)
	}
}
