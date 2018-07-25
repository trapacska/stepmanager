package cmd

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/bitrise-io/stepman/models"
	"github.com/iancoleman/strcase"
	yaml "gopkg.in/yaml.v2"

	"github.com/bitrise-io/go-utils/fileutil"

	"github.com/bitrise-io/go-utils/command"
	"github.com/bitrise-io/go-utils/log"
	"github.com/bitrise-io/go-utils/pathutil"
	"github.com/spf13/cobra"
)

func failf(format string, v ...interface{}) {
	log.Errorf(format, v...)
	os.Exit(1)
}

var rootCmd = &cobra.Command{
	Use:   "stepmanager",
	Short: "Step development manager",
	Run: func(cmd *cobra.Command, args []string) {
		generateConfig()
	},
}

func makeFirstLowerCase(s string) string {

	if len(s) < 2 {
		return strings.ToLower(s)
	}

	bts := []byte(s)

	lc := bytes.ToLower([]byte{bts[0]})
	rest := bts[1:]

	return string(bytes.Join([][]byte{lc, rest}, nil))
}

func generateConfig() {
	log.Infof("Generating config: " + Dir)
	sec := ""
	filecont := "package main\n\nimport \"os\"\n\nvar inputs = struct {\n"
	if cont, err := fileutil.ReadBytesFromFile(filepath.Join(Dir, "step.yml")); err != nil {
		failf("Failed to read string: %s", err)
	} else {
		var step models.StepModel
		if err := yaml.Unmarshal(cont, &step); err != nil {
			failf("Failed to unmarshal: %s", err)
		}
		for _, input := range step.Inputs {
			key, _, err := input.GetKeyValuePair()
			if err != nil {
				failf("Failed to get input key: %s", err)
			}
			filecont += makeFirstLowerCase(strcase.ToCamel(key)) + " string\n"
			sec += makeFirstLowerCase(strcase.ToCamel(key)) + ": os.Getenv(\"" + key + "\"),\n"

		}
	}

	filecont += "}{\n" + sec + "}"
	if err := fileutil.WriteStringToFile(filepath.Join(Dir, "step_config.go"), filecont); err != nil {
		failf("Failed to write string: %s", err)
	}
}

type conf struct {
	Match   string `json:"match"`
	IsAsync bool   `json:"isAsync"`
	Cmd     string `json:"cmd"`
}

var initCmd = &cobra.Command{
	Use:   "init",
	Short: "Prepare visual studio and step repo for step input/output handling",
	Run: func(cmd *cobra.Command, args []string) {
		if b, err := pathutil.IsPathExists("./step.yml"); err != nil {
			failf("Failed to check if step.yml exists: %s", err)
		} else if !b {
			failf("No step.yml in current directory")
		}
		log.Infof("Installing code extension")
		extensionsCmd := command.New("code", "--list-extensions")
		if out, err := extensionsCmd.RunAndReturnTrimmedCombinedOutput(); err != nil {
			failf("Failed to run command: $ %s, output: %s, error: %s", extensionsCmd.PrintableCommandArgs(), out, err)
		} else if !strings.Contains(strings.ToLower(out), "emeraldwalk.runonsave") {
			log.Warnf("No installed yet, installing...")
			installCmd := command.New("code", "--install-extension", "emeraldwalk.runonsave")
			if out, err := installCmd.RunAndReturnTrimmedCombinedOutput(); err != nil {
				failf("Failed to run command: $ %s, output: %s, error: %s", installCmd.PrintableCommandArgs(), out, err)
			}
			log.Donef("- Done")
		} else {
			log.Printf("Already installed, skipping...")
			log.Donef("- Done")
		}
		fmt.Println()
		log.Infof("Configure code settings")

		if cont, err := fileutil.ReadBytesFromFile(filepath.Join(os.Getenv("HOME"), "Library/Application Support/Code/User/settings.json")); err != nil {
			failf("Failed to read string: %s", err)
		} else {
			var data map[string]interface{}
			if err := json.Unmarshal(cont, &data); err != nil {
				failf("Failed to unmarshal: %s", err)
			}
			data["emeraldwalk.runonsave"] = map[string][]conf{
				"commands": []conf{conf{Match: "step.yml", IsAsync: true, Cmd: "stepmanager -d ${fileDirname}"}},
			}
			if b, err := json.MarshalIndent(data, "", " "); err != nil {
				failf("Failed to marshal: %s", err)
			} else {
				if err := fileutil.WriteBytesToFile(filepath.Join(os.Getenv("HOME"), "Library/Application Support/Code/User/settings.json"), b); err != nil {
					failf("Failed to write string: %s", err)
				}
			}
		}
		fmt.Println()

		generateConfig()
	},
}

var Dir string

// Execute ...
func Execute() {
	rootCmd.Flags().StringVarP(&Dir, "dir", "d", "", "work dir")
	rootCmd.MarkFlagRequired("dir")
	rootCmd.AddCommand(initCmd)

	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
