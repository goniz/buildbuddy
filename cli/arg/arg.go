package arg

import (
	"fmt"
	"strings"
)

// Returns true if the list of args contains desiredArg
func Has(args []string, desiredArg string) bool {
	arg, _ := find(args, desiredArg)
	return arg != ""
}

// Returns the value of the given desiredArg if contained in args
func Get(args []string, desiredArg string) string {
	arg, _ := find(args, desiredArg)
	return arg
}

// Returns the value of the given desiredArg and a slice with that arg removed
func Pop(args []string, desiredArg string) (string, []string) {
	arg, i := find(args, desiredArg)
	if i < 0 {
		return "", args
	}
	return arg, append(args[:i], args[i+1:]...)
}

// Returns a list with argToRemove removed (if it exists)
func Remove(args []string, argToRemove string) []string {
	_, remaining := Pop(args, argToRemove)
	return remaining
}

// Helper method for finding arguments by prefix within a list of arguments
func find(args []string, desiredArg string) (string, int) {
	prefix := fmt.Sprintf("--%s=", desiredArg)
	for i, arg := range args {
		if strings.HasPrefix(arg, prefix) {
			return strings.TrimPrefix(arg, prefix), i
		}
	}
	return "", -1
}

// Returns the first non-option found in the list of args (doesn't begin with "-")
func GetCommand(args []string) string {
	command, _ := GetCommandAndIndex(args)
	return command
}

// Returns the first non-option found in the list of args (doesn't begin with "-") and the index at which it was found
func GetCommandAndIndex(args []string) (string, int) {
	for i, arg := range args {
		if strings.HasPrefix("-", arg) {
			continue
		}
		return arg, i
	}
	return "", -1
}

// Returns any "passthrough" arugments in args (those to the right of the first " -- ", if any)
func GetPassthroughArgs(args []string) []string {
	for i, arg := range args {
		if arg == "--" {
			return append([]string{}, args[i+1:]...)
		}
	}
	return nil
}

// Returns any non "passthrough" arugments in args (those to the left of the first " -- ", if any)
func GetNonPassthroughArgs(args []string) []string {
	splitIndex := len(args)
	for i, arg := range args {
		if arg == "--" {
			splitIndex = i
			break
		}
	}
	return append([]string{}, args[:splitIndex]...)
}

// Splits bazel args and passthrough args into two separate lists. The first
// "--" separator is dropped if it exists.
func SplitPassthroughArgs(args []string) (bazel []string, passthrough []string) {
	return GetNonPassthroughArgs(args), GetPassthroughArgs(args)
}

// JoinPassthroughArgs joins the given args and passthrough args with a "--"
// separator, if the passthrough args are non-empty. Otherwise it returns
// the original non-passthrough args.
func JoinPassthroughArgs(args, passthroughArgs []string) []string {
	out := append([]string{}, args...)
	if len(passthroughArgs) == 0 {
		return out
	}
	out = append(out, "--")
	out = append(out, passthroughArgs...)
	return out
}

// Returns args with any arguments also found in existingArgs removed
func RemoveExistingArgs(args []string, existingArgs []string) []string {
	exists := make(map[string]struct{}, len(existingArgs))
	for _, x := range existingArgs {
		exists[x] = struct{}{}
	}
	var diff []string
	for _, x := range args {
		if _, found := exists[x]; !found {
			diff = append(diff, x)
		}
	}
	return diff
}
