// Command domain-scout runs discovery directly on the machine where it is
// executed. It emits the same discovery.Result JSON payload as POST /api/v1/discoveries.
package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"domain-scout/internal/discovery"
	"domain-scout/internal/inspectors"
	"domain-scout/internal/inspectors/apache"
	"domain-scout/internal/inspectors/caddy"
	"domain-scout/internal/inspectors/docker"
	"domain-scout/internal/inspectors/nginx"
	"domain-scout/internal/inspectors/system"
	"domain-scout/internal/ssh"
)

const version = "1.1.0"

func main() {
	if len(os.Args) < 2 {
		usage()
		os.Exit(2)
	}

	switch os.Args[1] {
	case "version", "--version", "-v":
		fmt.Printf("domain-scout %s\n", version)
	case "help", "--help", "-h":
		usage()
	case "scan":
		if err := runScan(os.Args[2:]); err != nil {
			writeError("discovery_failed", err.Error())
			os.Exit(1)
		}
	default:
		fmt.Fprintf(os.Stderr, "unknown command %q\n\n", os.Args[1])
		usage()
		os.Exit(2)
	}
}

func runScan(args []string) error {
	flags := flag.NewFlagSet("scan", flag.ContinueOnError)
	flags.SetOutput(os.Stderr)
	pretty := flags.Bool("pretty", true, "indent JSON output (default true)")
	jsonOutput := flags.Bool("json", false, "emit JSON output (default true)")
	compact := flags.Bool("compact", false, "emit compact JSON")
	timeout := flags.Duration("timeout", 2*time.Minute, "maximum scan duration")
	if err := flags.Parse(args); err != nil {
		return err
	}
	if *timeout <= 0 {
		return fmt.Errorf("timeout must be greater than zero")
	}
	if *jsonOutput {
		*pretty = true
	}
	if *compact {
		*pretty = false
	}

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()
	ctx, cancel := context.WithTimeout(ctx, *timeout)
	defer cancel()

	registry := []inspectors.Inspector{
		nginx.New(), apache.New(), caddy.New(), docker.New(), system.New(),
	}
	service := discovery.NewService(registry)
	result, err := service.Discover(ctx, ssh.NewLocalExecutor())
	if err != nil {
		return err
	}

	var output []byte
	if *pretty {
		output, err = json.MarshalIndent(result, "", "  ")
	} else {
		output, err = json.Marshal(result)
	}
	if err != nil {
		return fmt.Errorf("encode discovery result: %w", err)
	}
	_, err = os.Stdout.Write(append(output, '\n'))
	return err
}

func writeError(code, message string) {
	message = strings.TrimSpace(message)
	payload := struct {
		Error struct {
			Code    string `json:"code"`
			Message string `json:"message"`
		} `json:"error"`
	}{}
	payload.Error.Code = code
	payload.Error.Message = message
	_ = json.NewEncoder(os.Stdout).Encode(payload)
}

func usage() {
	fmt.Fprintln(os.Stderr, `DomainScout - discover domains configured on this server

Usage:
  domain-scout scan [--json] [--pretty|--compact] [--timeout 2m]
  domain-scout version

Examples:
  sudo ./domain-scout scan --json
  sudo ./domain-scout scan --compact > domains.json
  ./domain-scout version

The scan command prints the same discovery JSON payload as the API endpoint.`)
}
