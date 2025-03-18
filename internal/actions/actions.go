package actions

// Note for christain please make a new table and query to store the script information in the database.

import (
	"fmt"
	"log"
)

type ActionsInfo struct {
	Name        string
	Path        string
	Description string
	Executable 	string
}

var scriptsMap = map[string]ActionsInfo{
	"test": {Name: "test.sh", Path: "./data/actions/test.sh", Description: "This is a test script", Executable: "bash"},
}

func AddScript(name, path, description, executable string) error {
	if _, exists := scriptsMap[name]; exists {
		return fmt.Errorf("script with name '%s' already exists", name)
	}

	fullPath := "./data/actions/" + path

	scriptsMap[name] = ActionsInfo{
		Name:        name,
		Path:        fullPath,
		Description: description,
		Executable: executable,
	}

	log.Printf("Script '%s' added successfully", name)
	return nil
}

func GetScript(name string) (ActionsInfo, error) {
	action, exists := scriptsMap[name]
	if !exists {
		return ActionsInfo{}, fmt.Errorf("script with name '%s' does not exist", name)
	}

	return action, nil
}
