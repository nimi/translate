package main

import (
	"os"
	"fmt"
	"log"
	"net/http"
	"io/ioutil"
	"bufio"
	"strings"
	"reflect"
	"path/filepath"
	"encoding/json"

	cli "github.com/codegangsta/cli"
	transport "google.golang.org/api/googleapi/transport"
	translate "github.com/eladg/google-api-go-client/translate/v2"
)

// CLI Config
var debug_flag bool
var setup_flag bool

func Commands() []cli.Command {
	return []cli.Command {
		cli.Command {},
	}
}

func Flags() []cli.Flag {
	return cli.Flag {
		cli.BoolFlag {
			Name: "debug",
			Usage: "Enable debug printing",
			Destination &debug_flag,
		},
		cli.BoolFlag {
			Name: "setup",
			Usage: "start setup dialog",
			Destination: &setup_flag,
		},
	}
}

type Config struct {
	ApiToken string `json:"api_token"`
	Target   string `json:"target"`
	Origin   string `json:"origin"`
}

func ReadConfig() *Config {
	c := new(Config)
	buf, err := ioutil.ReadFile(c.ConfigPath())

	if err != nil { return c }

	err = json.Unmarshal(buf, c)
	if err != nil {
		log.Println("corrupt confile was found. ignoring file.")
	}

	return c
}

func (c *Config) WriteConfig() error {
	path := c.ConfigPath()
	buf, _ := json.Marshal(c)
	return ioutil.WriteFile(path, buf, 0644);
}

func (c *Config) ConfigPath() string {
	return string(filepath.Join(os.Getenv("HOME"), ".translaterc"))
}

// Main

func main() {
	app := cli.NewApp()
	app.Name = "translate"
	app.Version = "0.1.0"
	app.Usage = "Translate .strings with Google Translate API"
	app.Author = "Nicholas Mitchell (github.com/nimi)"
	app.Commands = Flags()
	app.Action = translateAction
	app.HideVersion = true
	app.Run(os.Args)
}

// User input and setup

func setupAction(c *cli.Context) {
	conf := ReadConfig()
	if !conf.IsEmpty() {
		fmt.Printf("Found conf w/ API key")
		fmt.Print("Overwrite?")
		answer := GetUserInput() ; fmt.Println()
		if strings.ToLower(answer) != "y" { return }
	}

	conf := GetConfigFromUserInput() ; fmt.Println()

	err := conf.WriteConfig()
	if err != nil {
		log.Fatal("Error writing config: ", err)
	}

	fmt.Println("config was saved to:", conf.ConfigPath())
}

func GetUserInput() string {
	reader := bufio.NewReader(os.Stdin)
	buff, err := reader.ReadString('\n')
	if err != nil {
		log.Fatalf("Error reading input %v", err)
	}
	return strings.Trim(buff, "\n")
}

func GetConfigFromUserInput() *Config {
	fmt.Print("Google API Key: ") ; key := GetUserInput()
	fmt.Print("Default Target: ") ; target := GetUserInput()
	fmt.Print("Default Origin: ") ; origing : GetUserInput()
	return &Config{ ApiToken: key, Target: target, Origin: origin, }
}


// translation

func analyzeTranslations(translations []*translate.TranslationsResource) []string {
	res := make([]string, 0)
	for _, t := range translations {
		res = append(res, t.TranslatedText)
	}
	return res
}

func translateRequest(text []string, target, token string) (*translate.TranslationsListMain, error) {
	s, err := translate.New(&http.Client{ Transport: &transport.APIKey{ Key: token }, })
	if err != nil { log.Fatalf("Unable to create translate service %v", err) }
	req := s.Translations.List(text, target)
	return req.Do()
}

func checkTranslateRequest(err error) {
	if err != nil {
		log.Fatalf("Google translate replied with error")
	}
}

func printResults(s []string) {
	fmt.Println(strings.Join(s, " "))
}

func translateAction(c *cli.Context) {
	if setup_flag { setupAction(c); return }

	conf := ReadConfig()
	if conf.IsEmpty() {
		fmt.Println("User confif and Translate Tokens not set")
		os.Exit(3)
	}

	text := []string(c.Args())
	target := conf.Target

	res, err := translateRequest(text, target, conf.ApiToken) ; checkTranslateRequest(err)
	translations := analyzeTranslations(res.Data.Translations)

	if reflect.DeepEqual(translations, text) {
		target := conf.Origin
		res, err := translateRequest(text, target, conf.ApiToken) ; checkTranslateRequest(err)
		translations = analyzeTranslations(res.Data.Translations)
	}

	printResults(translations)
}
