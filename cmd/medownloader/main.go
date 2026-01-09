package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"strconv"
	"time"

	"github.com/matejeliash/medownloader/internal/downloader"
	"github.com/matejeliash/medownloader/internal/server"
)

func parsePort(flagPort int) (string, error) {
	portEnv := os.Getenv("ME_PORT")

	// use port flag if env. var not set
	if portEnv == "" {
		return fmt.Sprintf(":%d", flagPort), nil
	}

	// parse env var
	port, err := strconv.Atoi(portEnv)

	if err != nil {
		return "", fmt.Errorf("provided port number [%s] is invalid", portEnv)
	}

	if port < 1024 || port > 65535 {
		return "", fmt.Errorf("port number [%s] is not in valid range", portEnv)

	}
	return fmt.Sprintf(":%d", port), nil

}

func parsePassword() {

	password := os.Getenv("ME_PASSWORD")

	if password == "" {
		fmt.Println("password not set, using default password [password]")
		os.Setenv("ME_PASSWORD", "password")
	}

}

func parseSessionDuration(flagSessionDuration int) (time.Duration, error) {

	minsEnv := os.Getenv("ME_SESSION_DURATION")
	// use default session duration if env. var not set
	if minsEnv == "" {
		validity := time.Duration(flagSessionDuration) * time.Minute
		fmt.Println("session validity duration not set, using default 30 minutes")
		return validity, nil
	}
	//
	mins, err := strconv.Atoi(minsEnv)
	if err != nil {
		return 0, fmt.Errorf("[%s] is not valid session validity duration in minutes\n", minsEnv)
	}

	validity := time.Duration(mins) * time.Minute
	return validity, nil
}

func main() {

	portFlag := flag.Int("port", 8080, "server port, same as env. variable ME_PORT")
	sessionDurationFlag := flag.Int("sessionDuration", 30, "session duration in minutes, same as env. variable ME_SESSION_DURATION")

	flag.Usage = func() {
		fmt.Println("Medownloader is simple downloader app and server written in golang.")
		fmt.Println(`Default password is "password", but can be changed with env. variable ME_PASSWORD`)
		fmt.Println()
		fmt.Println(`Usage of medownloader:`)

		flag.PrintDefaults()
	}

	flag.Parse()

	parsedPort, err := parsePort(*portFlag)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	validity, err := parseSessionDuration(*sessionDurationFlag)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	parsePassword()
	// fmt.Println("***********")
	// fmt.Println(validity)
	// fmt.Println(parsedPort)
	// fmt.Println(os.Getenv("ME_PASSWORD"))
	// fmt.Println("***********")

	dm := downloader.NewDownloadManager()
	sm := server.NewSessionManager(validity)
	s := server.New(dm, sm, parsedPort)
	log.Println("running server on port " + parsedPort)
	if err := s.Run(); err != nil {
		log.Fatal(err)
	}

}
