package cmd

import (
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"runtime"

	"github.com/spf13/cobra"
	"gopkg.in/yaml.v2"
)

type ConfigOptions struct {
	Current  string
	Show     bool
	Generate bool
	List     bool

	ConfigFileLocation string
}

var configOptions ConfigOptions

func init() {
	rootCmd.AddCommand(configCmd)
	configCmd.PersistentFlags().StringVarP(&configOptions.Current, "current", "c", "", "Set the current Jenkins")
	configCmd.PersistentFlags().BoolVarP(&configOptions.Show, "show", "s", false, "Show the current Jenkins")
	configCmd.PersistentFlags().BoolVarP(&configOptions.Generate, "generate", "g", false, "Generate a sample config file for you")
	configCmd.PersistentFlags().BoolVarP(&configOptions.List, "list", "l", false, "Display all your Jenkins configs")
}

var configCmd = &cobra.Command{
	Use:   "config",
	Short: "Manage the config of jcli",
	Long:  `Manage the config of jcli`,
	Run: func(cmd *cobra.Command, args []string) {
		if configOptions.Show {
			current := getCurrentJenkins()
			fmt.Printf("Current Jenkins's name is %s, url is %s\n", current.Name, current.URL)
		}

		if configOptions.List {
			fmt.Println("number-name\turl")
			for i, jenkins := range getConfig().JenkinsServers {
				fmt.Printf("%d-%s\t%s\n", i, jenkins.Name, jenkins.URL)
			}
		}

		if configOptions.Generate {
			if data, err := generateSampleConfig(); err == nil {
				fmt.Print(string(data))
			} else {
				log.Fatal(err)
			}
		}

		if configOptions.Current != "" {
			found := false
			for _, jenkins := range getConfig().JenkinsServers {
				if jenkins.Name == configOptions.Current {
					found = true
					break
				}
			}

			if found {
				config.Current = configOptions.Current
				if err := saveConfig(); err != nil {
					log.Fatal(err)
				}
			} else {
				log.Fatalf("Cannot found Jenkins by name %s", configOptions.Current)
			}
		}
	},
}

// JenkinsServer holds the configuration of your Jenkins
type JenkinsServer struct {
	Name      string `yaml:"name"`
	URL       string `yaml:"url"`
	UserName  string `yaml:"username"`
	Token     string `yaml:"token"`
	Proxy     string `yaml:"proxy"`
	ProxyAuth string `yaml:"proxyAuth"`
}

type Config struct {
	Current        string          `yaml:"current"`
	JenkinsServers []JenkinsServer `yaml:"jenkins_servers"`
}

func generateSampleConfig() ([]byte, error) {
	sampleConfig := Config{
		Current: "yourServer",
		JenkinsServers: []JenkinsServer{
			{
				Name:     "yourServer",
				URL:      "http://localhost:8080/jenkins",
				UserName: "admin",
				Token:    "111e3a2f0231198855dceaff96f20540a9",
			},
		},
	}
	return yaml.Marshal(&sampleConfig)
}

var config Config

func getConfig() Config {
	return config
}

func getCurrentJenkins() (jenkinsServer *JenkinsServer) {
	config := getConfig()
	current := config.Current
	jenkinsServer = findJenkinsByName(current)

	return
}

func findJenkinsByName(name string) (jenkinsServer *JenkinsServer) {
	for _, cfg := range config.JenkinsServers {
		if cfg.Name == name {
			jenkinsServer = &cfg
			break
		}
	}
	return
}

func loadDefaultConfig() {
	userHome := userHomeDir()
	if err := loadConfig(fmt.Sprintf("%s/.jenkins-cli.yaml", userHome)); err != nil {
		log.Fatalf("error: %v", err)
	}
}

func loadConfig(path string) (err error) {
	configOptions.ConfigFileLocation = path

	var content []byte
	if content, err = ioutil.ReadFile(path); err == nil {
		err = yaml.Unmarshal([]byte(content), &config)
	}
	return
}

func saveConfig() (err error) {
	var data []byte
	config := getConfig()

	if data, err = yaml.Marshal(&config); err == nil {
		err = ioutil.WriteFile(configOptions.ConfigFileLocation, data, 0644)
	}
	return
}

func userHomeDir() string {
	if runtime.GOOS == "windows" {
		home := os.Getenv("HOMEDRIVE") + os.Getenv("HOMEPATH")
		if home == "" {
			home = os.Getenv("USERPROFILE")
		}
		return home
	} else if runtime.GOOS == "linux" {
		home := os.Getenv("XDG_CONFIG_HOME")
		if home != "" {
			return home
		}
	}
	return os.Getenv("HOME")
}
