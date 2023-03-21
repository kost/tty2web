package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"strings"
	"syscall"

	"github.com/urfave/cli/v2"

	"github.com/kost/tty2web/backend/localcommand"
	"github.com/kost/tty2web/pkg/homedir"
	"github.com/kost/tty2web/server"
	"github.com/kost/tty2web/utils"
)

func main() {
	SCEnvAndExecute()
	app := cli.NewApp()
	app.Name = "tty2web"
	app.Version = Version + "+" + CommitID
	app.Usage = "Share your terminal as a web application"
	app.HideHelp = true
	cli.AppHelpTemplate = helpTemplate

	appOptions := &server.Options{}
	if err := utils.ApplyDefaultValues(appOptions); err != nil {
		log.Printf("Error applying default value: %v", err)
		exit(err, 1)
	}
	backendOptions := &localcommand.Options{}
	if err := utils.ApplyDefaultValues(backendOptions); err != nil {
		log.Printf("Error applying backend default value: %v", err)
		exit(err, 1)
	}

	cliFlags, flagMappings, err := utils.GenerateFlags(appOptions, backendOptions)
	if err != nil {
		log.Printf("Error generating flags: %v", err)
		exit(err, 3)
	}

	app.Flags = append(
		cliFlags,
		&cli.StringFlag{
			Name:   "config",
			Value:  "~/.tty2web",
			Usage:  "Config file path",
			EnvVars: []string{"TTY2WEB_CONFIG"},
		},
		&cli.BoolFlag{
			Name:	"help",
			Usage:	"Displays help",
		},
	)

	app.Action = func(c *cli.Context) error {

		configFile := c.String("config")
		_, err := os.Stat(homedir.Expand(configFile))
		if configFile != "~/.tty2web" || !os.IsNotExist(err) {
			if err := utils.ApplyConfigFile(configFile, appOptions, backendOptions); err != nil {
				log.Printf("Error applying config file: %v", err)
				exit(err, 2)
			}
		}

		utils.ApplyFlags(cliFlags, flagMappings, c, appOptions, backendOptions)

		appOptions.EnableBasicAuth = c.IsSet("credential")
		appOptions.EnableTLSClientAuth = c.IsSet("tls-ca-crt")

		err = appOptions.Validate()
		if err != nil {
			log.Printf("Error validating options: %v", err)
			exit(err, 6)
		}

		if appOptions.Listen!="" {
			log.Printf("Listening for reverse connection %s", appOptions.Listen)
			go func() {
			log.Fatal(listenForAgents(appOptions.Verbose, true, appOptions.Listen, appOptions.Server, appOptions.ListenCert, appOptions.Password))
			}()
			wait4Signals()
			return nil
		}

		args := c.Args()
		if args.Len() == 0 {
			msg := "Error: No command given."
			cli.ShowAppHelp(c)
			exit(fmt.Errorf(msg), 1)
		}

		if c.Bool("help") {
			cli.ShowAppHelp(c)
			exit(err, 1)
		}

		factory, err := localcommand.NewFactory(c.Args().First(), c.Args().Tail(), backendOptions)
		if err != nil {
			log.Printf("Error creating local command: %v", err)
			exit(err, 3)
		}

		hostname, _ := os.Hostname()
		appOptions.TitleVariables = map[string]interface{}{
			"command":  c.Args().First(),
			"argv":     c.Args().Tail(),
			"hostname": hostname,
		}

		srv, err := server.New(factory, appOptions)
		if err != nil {
			log.Printf("Error creating new server: %v", err)
			exit(err, 5)
		}


		log.Printf("tty2web is starting with command: %s", strings.Join(args.Slice(), " "))

		ctx, cancel := context.WithCancel(context.Background())
		gCtx, gCancel := context.WithCancel(context.Background())
		errs := make(chan error, 1)

		go func() {
			errs <- srv.Run(ctx, server.WithGracefullContext(gCtx))
		}()
		err = waitSignals(errs, cancel, gCancel)

		if err != nil && err != context.Canceled {
			fmt.Printf("Error: %s\n", err)
			exit(err, 8)
		}
		return nil

	}
	app.Run(os.Args)
}

func exit(err error, code int) {
	if err != nil {
		fmt.Println(err)
	}
	os.Exit(code)
}

func wait4Signals() {
	c := make(chan os.Signal)
	signal.Notify(c, os.Interrupt)
	select {
        case sig := <-c:
            fmt.Printf("Got %s signal. Aborting...\n", sig)
            os.Exit(1)
        }
}

func waitSignals(errs chan error, cancel context.CancelFunc, gracefullCancel context.CancelFunc) error {
	sigChan := make(chan os.Signal, 1)
	signal.Notify(
		sigChan,
		syscall.SIGINT,
		syscall.SIGTERM,
	)

	select {
	case err := <-errs:
		return err

	case s := <-sigChan:
		switch s {
		case syscall.SIGINT:
			gracefullCancel()
			fmt.Println("C-C to force close")
			select {
			case err := <-errs:
				return err
			case <-sigChan:
				fmt.Println("Force closing...")
				cancel()
				return <-errs
			}
		default:
			cancel()
			return <-errs
		}
	}
}
