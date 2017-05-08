package main

import (
	"commands"
	"common"
	"flag"
	"flags"
	"fmt"
	"os"
	"strconv"

	"github.com/fatih/color"
	fastly "github.com/sethvargo/go-fastly"
	"github.com/sirupsen/logrus"
)

// appVersion is the application version
const appVersion = "0.0.1"

// useful colour settings for printing messages
var yellow = color.New(color.FgYellow).SprintFunc()
var red = color.New(color.FgRed).SprintFunc()
var green = color.New(color.FgGreen).SprintFunc()

var logger *logrus.Entry

func init() {
	logrus.SetFormatter(&logrus.JSONFormatter{})
	logrus.SetLevel(logrus.InfoLevel)
	logger = logrus.WithFields(logrus.Fields{
		"package": "main",
	})
}

func main() {
	f := flags.New()

	if *f.Top.Help == true {
		flag.PrintDefaults()
		os.Exit(1)
	}

	if *f.Top.Version == true {
		fmt.Println(appVersion)
		os.Exit(1)
	}

	if *f.Top.Debug == true {
		logrus.SetLevel(logrus.DebugLevel)
	}

	client, err := fastly.NewClient(*f.Top.Token)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	if *f.Top.Activate != "" {
		activateVersion(f, client)
		return
	}

	if *f.Top.Status != "" && *f.Top.Status == "latest" {
		latestVersion, status, err := getLatestServiceVersion(f, client)
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}

		fmt.Printf("\nLatest service version: %s (%s)\n\n", latestVersion, status)
		return
	}

	if *f.Top.Status != "" {
		status, err := getStatusVersion(*f.Top.Service, *f.Top.Status, client)
		if err != nil {
			fmt.Printf("\nThere was a problem getting the status for version %s\n\n%s\n\n", yellow(*f.Top.Status), red(err))
			os.Exit(1)
		}
		fmt.Printf("\nService '%s' version '%s' is '%s'\n\n", yellow(*f.Top.Service), yellow(*f.Top.Status), status)
		return
	}

	if *f.Top.Settings == "latest" {
		printLatestSettings(f, client)
		return
	}

	if *f.Top.Settings != "" {
		printSettingsFor(*f.Top.Settings, f, client)
		return
	}

	logger.Debug("application starting")

	args := os.Args[1:] // strip first arg `fastly`
	arg, counter := flags.Check(args)

	switch arg {
	case "diff":
		f.Top.Diff.Parse(args[counter:])
		commands.Diff(f, client)
	case "upload":
		f.Top.Upload.Parse(args[counter:])
		commands.Upload(f, client)
	default:
		fmt.Printf("%v is not valid command.\n", arg)
		os.Exit(1)
	}
}

func activateVersion(f flags.Flags, client *fastly.Client) {
	v, err := strconv.Atoi(*f.Top.Activate)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	_, err = client.ActivateVersion(&fastly.ActivateVersionInput{
		Service: *f.Top.Service,
		Version: v,
	})
	if err != nil {
		fmt.Printf("\nThere was a problem activating version %s\n\n%s", yellow(*f.Top.Activate), red(err))
		os.Exit(1)
	}
	fmt.Printf("\nService '%s' now has version '%s' activated\n\n", yellow(*f.Top.Service), green(*f.Top.Activate))
}

func getLatestServiceVersion(f flags.Flags, client *fastly.Client) (string, string, error) {
	latestVersion, err := common.GetLatestVCLVersion(*f.Top.Service, client)
	if err != nil {
		return "", "", err
	}

	status, err := getStatusVersion(*f.Top.Service, latestVersion, client)
	if err != nil {
		return "", "", err
	}

	return latestVersion, status, nil
}

func getStatusVersion(serviceVersion, statusVersion string, client *fastly.Client) (string, error) {
	v, err := strconv.Atoi(statusVersion)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	versionStatus, err := client.GetVersion(&fastly.GetVersionInput{
		Service: serviceVersion,
		Version: v,
	})
	if err != nil {
		return "", err
	}

	status := green("not activated")
	if versionStatus.Active {
		status = red("already activated")
	}

	return status, nil
}

func printLatestSettings(f flags.Flags, client *fastly.Client) {
	latestVersion, _, err := getLatestServiceVersion(f, client)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	printSettingsFor(latestVersion, f, client)
}

func printSettingsFor(serviceVersion string, f flags.Flags, client *fastly.Client) {
	v, err := strconv.Atoi(serviceVersion)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	settings, err := client.GetSettings(&fastly.GetSettingsInput{
		Service: *f.Top.Service,
		Version: v,
	})
	if err != nil {
		fmt.Printf("\nThere was a problem getting the settings for version %s\n\n%s", yellow(serviceVersion), red(err))
		os.Exit(1)
	}

	fmt.Printf(
		"\nDefault Host: %s\nDefault TTL: %d (seconds)\n\n",
		settings.DefaultHost,
		settings.DefaultTTL,
	)
}
