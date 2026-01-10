package main

import (
	"context"
	_ "embed"
	"fmt"
	"log"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"github.com/salex-org/hmip-exporter/internal/homematic"
	"github.com/salex-org/hmip-exporter/internal/util"
	"github.com/salex-org/hmip-exporter/internal/webserver"
)

var (
	homematicClient homematic.HomematicClient
	webServer       webserver.Server

	//go:embed assets/ascii.art
	asciiArt string
)

func main() {
	// Startup function
	fmt.Printf("%s\n\n", fmt.Sprintf(asciiArt, util.Version))
	err := startup()
	if err != nil {
		log.Fatalf("Error during startup: %v\n", err)
	}

	// Notification context for reacting on process termination - used by shutdown function
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	// Waiting group used to await finishing the shutdown process when stopping
	var wait sync.WaitGroup

	// Loop function for webserver
	wait.Add(1)
	go func() {
		defer wait.Done()
		fmt.Printf("Web server started\n")
		_ = webServer.Start()
	}()

	// Loop function for event listening
	wait.Add(1)
	go func() {
		defer wait.Done()
		fmt.Printf("HomematicIP client started\n")
		_ = homematicClient.Start()
	}()

	// Shutdown function waiting for the SIGTERM notification to stop event listening
	wait.Add(1)
	go func() {
		defer wait.Done()
		<-ctx.Done()
		fmt.Printf("\n\U0001F6D1 Shutdown down started...\n")
		shutdown()
	}()

	wait.Wait()
	fmt.Printf("\U0001F3C1 Shutdown finished\n")
	os.Exit(0)
}

func startup() error {
	var err error
	time.Local, err = time.LoadLocation("CET")
	if err != nil {
		return fmt.Errorf("error pinning location: %w", err)
	}
	fmt.Printf("Timezone CET loaded\n")

	webServer = webserver.NewServer(healthCheck)
	fmt.Printf("Web server created\n")

	homematicClient, err = homematic.NewHomematicClient()
	if err != nil {
		return fmt.Errorf("error creationg HomematicIP client: %w", err)
	}
	fmt.Printf("HomematicIP client created\n")

	return nil
}

func shutdown() {
	err := homematicClient.Shutdown()
	if err != nil {
		fmt.Printf("Error stopping event listening: %v\n", err)
	} else {
		fmt.Printf("Event listening stopped\n")
	}

	err = webServer.Shutdown()
	if err != nil {
		fmt.Printf("Error stopping web server: %v\n", err)
	} else {
		fmt.Printf("Web server stopped\n")
	}
}

func healthCheck() map[string]error {
	errors := make(map[string]error)
	if err := homematicClient.Health(); err != nil {
		errors["HomematicIP Client"] = err
	}
	return errors
}
