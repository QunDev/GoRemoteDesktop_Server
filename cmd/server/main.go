package main

import (
	"QunDev/GoRemoteDesktop_Server/internal/app"
	"fmt"
	"log"
	"net/http"
	"os"
)

func main() {
	application, err := app.InitializeApp()
	if err != nil {
		log.Fatal(err)
		os.Exit(1)
	}

	err = http.ListenAndServe(fmt.Sprintf(":%d", application.Config.Server.Port), nil)
	if err != nil {
		log.Fatal("ListenAndServe: ", err)
	}
}
