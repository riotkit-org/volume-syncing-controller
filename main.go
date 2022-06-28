package main

import (
	"github.com/riotkit-org/volume-syncing-operator/cmd"
	"github.com/riotkit-org/volume-syncing-operator/pkg/helpers"
	"os"
	"strings"
)

func main() {
	command := cmd.Main()
	args := prepareArgs(os.Args)

	if args != nil && args[1] != "serve" {
		args = args[1:]
		command.SetArgs(args)
	}

	err := command.Execute()
	if err != nil {
		os.Exit(1)
	}
}

// prepareArgs Collects all REMOTE_xxx, ENCRYPTED_xxx environment variables and changes to "-p xxx=yyy" and "-e xxx=yyy"
//             That makes easier to use in Kubernetes and in Docker
//
//             e.g. REMOTE_ACCESS_KEY_ID=123456 - will convert to access_key_id=123456 under section [remote]
func prepareArgs(args []string) []string {
	for _, env := range os.Environ() {
		pair := strings.SplitN(env, "=", 2)
		envName := pair[0]
		value := helpers.GetEnvOrDefault(pair[0], "")

		if strings.HasPrefix(envName, "REMOTE_") {
			args = appendVar(args, "REMOTE_", "-p", envName, value.(string))
		} else if strings.HasPrefix(envName, "ENCRYPTED_") {
			args = appendVar(args, "ENCRYPTED_", "-e", envName, value.(string))
		}
	}
	return args
}

func appendVar(args []string, prefix string, switchName string, envName string, value string) []string {
	namePair := strings.SplitN(envName, prefix, 2)
	name := strings.ToLower(namePair[1])

	args = append(args, switchName)
	args = append(args, name+" = "+value)

	return args
}
