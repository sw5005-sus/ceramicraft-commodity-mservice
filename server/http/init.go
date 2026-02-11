package http

import (
	"fmt"
	"os"

	"github.com/sw5005-sus/ceramicraft-commodity-mservice/server/config"
	"github.com/sw5005-sus/ceramicraft-commodity-mservice/server/http/router"
	"github.com/sw5005-sus/ceramicraft-commodity-mservice/server/log"
)

func Init(exitSig chan os.Signal) {
	r := router.NewRouter()
	log.Logger.Infof("Product HTTP Server is running on %s:%d", config.Config.HttpConfig.Host, config.Config.HttpConfig.Port)
	err := r.Run(fmt.Sprintf("%s:%d", config.Config.HttpConfig.Host, config.Config.HttpConfig.Port))
	if err != nil {
		log.Logger.Fatalf("Failed to run server: %v", err)
		exitSig <- os.Interrupt
	}
}
