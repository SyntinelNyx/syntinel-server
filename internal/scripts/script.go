package scripts

// Note for christain please make a new table and query to store the script information in the database.

import (
	"fmt"
	"log"
)

type ScriptInfo struct {
	Name        string
	Path        string
	Description string
	Executable 	string
}

var scriptsMap = map[string]ScriptInfo{
	"test": {Name: "test.sh", Path: "./data/scripts/test.sh", Description: "This is a test script", Executable: "bash"},
}

func AddScript(name, path, description, executable string) error {
	if _, exists := scriptsMap[name]; exists {
		return fmt.Errorf("script with name '%s' already exists", name)
	}

	fullPath := "./data/scripts/" + path

	scriptsMap[name] = ScriptInfo{
		Name:        name,
		Path:        fullPath,
		Description: description,
		Executable: executable,
	}

	log.Printf("Script '%s' added successfully", name)
	return nil
}

func GetScript(name string) (ScriptInfo, error) {
	script, exists := scriptsMap[name]
	if !exists {
		return ScriptInfo{}, fmt.Errorf("script with name '%s' does not exist", name)
	}

	return script, nil
}
