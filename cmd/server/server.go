package main

import (
	"bufio"
	"context"
	"flag"
	"fmt"
	"log"
	"net"
	"net/http"
	"os"
	"runtime"
	"runtime/pprof"
	"strconv"

	"github.com/atyronesmith/sa-benchmark/pkg/inetserver"
	"github.com/atyronesmith/sa-benchmark/pkg/unixserver"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

const UNIX_SOCKET_PATH string = "/tmp/smartgateway"

func main() {
	if os.Getenv("DEBUG") != "" {
		runtime.SetBlockProfileRate(20)
		runtime.SetMutexProfileFraction(20)
	}

	inetCommand := flag.NewFlagSet("inet", flag.ExitOnError)
	unixCommand := flag.NewFlagSet("unix", flag.ExitOnError)

	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "usage: %s [options] <command> [options]\n", os.Args[0])
		flag.PrintDefaults()
		fmt.Fprintf(os.Stderr, "\nusage: %s [options] inet [options]\n", os.Args[0])
		inetCommand.PrintDefaults()
		fmt.Fprintf(os.Stderr, "\nusage: %s [options] unix [options]\n\n", os.Args[0])
		unixCommand.PrintDefaults()
	}

	promport := flag.Int("promport", 8081, "Prometheus scrape port.")
	cpuprofile := flag.String("cpuprofile", "", "write cpu profile to file")
	capture := flag.Bool("capture", false, "Catpure json output.")

	// Add Flags for net command
	// parse command line option
	ipAddress := inetCommand.String("ip", "127.0.0.1", "Listening IP address")
	port := inetCommand.Int("port", 0, "Port to use, otherwise OS will choose")

	// Add Flags for shared command
	socketPath := unixCommand.String("path", UNIX_SOCKET_PATH, "Path/file for the shared memeory socket")

	flag.Parse()

	commandArgs := flag.Args()

	// Verify that a subcommand has been provided
	// os.Arg[0] is the main command
	// os.Arg[1] will be the subcommand
	if len(commandArgs) < 1 {
		fmt.Println("inet or unix subcommand is required")
		flag.Usage()
		os.Exit(1)
	}

	// Switch on the subcommand
	// Parse the flags for appropriate FlagSet
	// FlagSet.Parse() requires a set of arguments to parse as input
	// os.Args[2:] will be all arguments starting after the subcommand at os.Args[1]
	switch commandArgs[0] {
	case "inet":
		err := inetCommand.Parse(commandArgs[1:])
		if err != nil {
			panic(err)
		}
	case "unix":
		err := unixCommand.Parse(commandArgs[1:])
		if err != nil {
			panic(err)
		}
	default:
		flag.Usage()
		os.Exit(1)
	}

	var w *bufio.Writer
	var err error

	if *capture {
		var fo *os.File
		// open output file
		fo, err = os.Create("cd-capture.txt")
		if err != nil {
			panic(err)
		}
		// close fo on exit and check for its returned error
		defer func() {
			if err := fo.Close(); err != nil {
				panic(err)
			}
		}()
		// make a write buffer
		w = bufio.NewWriter(fo)
	}

	go func() {
		err := http.ListenAndServe(":"+strconv.Itoa(*promport), promhttp.Handler())
		if err != nil {
			fmt.Printf("http server failed!...")
			fmt.Printf("%+v\n", err)
		}
	}()

	ctx := context.Background()

	if *cpuprofile != "" {
		f, err := os.Create(*cpuprofile)
		if err != nil {
			log.Fatal(err)
		}
		defer f.Close()
		if err := pprof.StartCPUProfile(f); err != nil {
			log.Fatal("could not start CPU profile: ", err)
		}
		defer pprof.StopCPUProfile()
	}

	if inetCommand.Parsed() {
		ip := net.ParseIP(*ipAddress)
		if ip == nil {
			fmt.Fprintf(os.Stderr, "Invalid target IP addres %s...", *ipAddress)
			flag.Usage()
			os.Exit(1)
		}
		err = inetserver.Listen(ctx, ip.String()+":"+strconv.Itoa(*port), w)
		if err != nil {
			fmt.Printf("Error occurred")
		}
	} else if unixCommand.Parsed() {
		err = unixserver.Listen(ctx, *socketPath, w)
		if err != nil {
			fmt.Printf("Error occurred")
		}
	}

	if *capture {
		if err = w.Flush(); err != nil {
			panic(err)
		}
	}
}
