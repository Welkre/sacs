package sac

import (
	"fmt"
	"io"
	"os"

	"gopkg.in/yaml.v3"
)

type User struct {
	// The user struct houses breaks down the user configuration
	name  string            `yaml:"name"`
	email string            `yaml:"email"`
	racks map[string]string `yaml:"racks"` // This is the sac remote references, similar to git remotes
	bags  map[string]string `yaml:"bags"`  // This is the sac local references, similar to git branches
}

type Config struct {
	// The config houses the user configurations and
	// authentication information for the current repo.
	// It also refers to the global configurations as well.
	// the configuration keys and values are stored in a map
	// defined in yaml format.
	user     map[string]string `yaml:"user"`
	global   map[string]string `yaml:"global"`
	override map[string]string `yaml:"override"`
}

func (c *Config) ParseCurrentDir() {
	// Parses the current directory for a .sac.yaml file.
	// If the file exists, it will be parsed and the values
	// will be stored in the override map.
	currentDir, err := os.Getwd()
	if err != nil {
		panic(err)
	}
	// Check if the .sac.yaml file exists in the current directory
	if _, err := os.Stat(currentDir + "/.sac.yaml"); err == nil {
		// Parse the .sac.yaml file and store the values in the override map
		file, err := os.Open(currentDir + "/.sac.yaml")
		if err != nil {
			panic(err)
		}
		defer file.Close()
		// Read the file content into a byte slice
		content, err := io.ReadAll(file)
		if err != nil {
			panic(err)
		}
		// Unmarshal the content into the override map
		err = yaml.Unmarshal(content, &c.override)
		if err != nil {
			panic(err)
		}
	}

}

func (c *Config) ParseGlobal() {
	// Parses the global configuration file for the user.
	// The global configuration file is located in the home directory
	// and is named .sac.yaml. If the file exists, it will be parsed
	// and the values will be stored in the global map.
	homeDir, err := os.UserHomeDir()
	if err != nil {
		panic(err)
	}
	// Check if the .sac.yaml file exists in the home directory
	if _, err := os.Stat(homeDir + "/.sac.yaml"); err == nil {
		// Parse the .sac.yaml file and store the values in the global map
		file, err := os.Open(homeDir + "/.sac.yaml")
		if err != nil {
			panic(err)
		}
		defer file.Close()
		content, err := io.ReadAll(file)
		if err != nil {
			panic(err)
		}
		err = yaml.Unmarshal(content, &c.global)
		if err != nil {
			panic(err)
		}
	}
}

func (c *Config) BuildUser() {
	// Builds the users configuration context using the golbal and override maps.
	// The user map is built by merging the global and override maps.
	// The override map takes precedence over the global map.
	// The user map is used to store the final configuration values for the user.
	c.user = make(map[string]string)

	for key, value := range c.global {
		c.user[key] = value
	}
	for key, value := range c.override {
		c.user[key] = value
	}
	for key, value := range c.user {
		if value == "" {
			delete(c.user, key)
		}
	}

}

func handleConfigCommand(config *Config, args []string, startIndex int) error {
	if len(args) <= startIndex+1 {
		return fmt.Errorf("config command requires a subcommand")
	}

	switch args[startIndex+1] {
	case "set":
		if len(args) <= startIndex+3 {
			return fmt.Errorf("set command requires a key and value")
		}
		key := args[startIndex+2]
		value := args[startIndex+3]
		config.user[key] = value
		return nil

	case "get":
		if len(args) <= startIndex+2 {
			return fmt.Errorf("get command requires a key")
		}
		key := args[startIndex+2]
		value, exists := config.user[key]
		if !exists {
			return fmt.Errorf("no value found for key '%s'", key)
		}
		fmt.Printf("%s: %s\n", key, value)
		return nil

	case "delete":
		if len(args) <= startIndex+2 {
			return fmt.Errorf("delete command requires a key")
		}
		key := args[startIndex+2]
		delete(config.user, key)
		fmt.Printf("Deleted key '%s' from user configuration.\n", key)
		return nil

	case "list":
		if len(config.user) == 0 {
			fmt.Println("No configuration values found.")
			return nil
		}
		for key, value := range config.user {
			fmt.Printf("%s: %s\n", key, value)
		}
		return nil

	default:
		return fmt.Errorf("unknown config subcommand: %s", args[startIndex+1])
	}
}

func CommandSetup(config *Config, args []string) {
	// This function sets up the configuration for the sac command line tool.
	// The User is built using the user map.
	user := User{
		name:  config.user["name"],
		email: config.user["email"],
		racks: make(map[string]string),
		bags:  make(map[string]string),
	}

	// The racks and bags are built using the user map.
	for key, value := range config.user {
		if key == "racks" {
			user.racks = make(map[string]string)
			err := yaml.Unmarshal([]byte(value), &user.racks)
			if err != nil {
				panic(err)
			}
		} else if key == "bags" {
			user.bags = make(map[string]string)
			err := yaml.Unmarshal([]byte(value), &user.bags)
			if err != nil {
				panic(err)
			}
		}
	}
	// switch on the args to determine what setup steps to take
	for i, arg := range args {
		switch arg {
		case "init":
			// Initialize the configuration for the user
			// This will create a new .sac.yaml file in the current directory
			// and populate it with the user configuration values.
			config.ParseCurrentDir()
			config.BuildUser()
		case "config":
			if err := handleConfigCommand(config, args, i); err != nil {
				fmt.Printf("Error: %v\n", err)
			}
			return
		}
	}
}
