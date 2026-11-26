package cli

import (
	"errors"
	"flag"
	"fmt"
	"io"
	"strings"
)

type repeated []string

func (r *repeated) String() string {
	return strings.Join(*r, ",")
}

func (r *repeated) Set(value string) error {
	*r = append(*r, value)
	return nil
}

type usageError struct {
	message string
	flags   *flag.FlagSet
}

func (e usageError) Error() string {
	return e.message
}

type helpRequested struct {
	flags *flag.FlagSet
}

func (helpRequested) Error() string {
	return "help requested"
}

func usageOfFlags(flags *flag.FlagSet, format string, values ...any) error {
	return usageError{message: flags.Name() + ": " + fmt.Sprintf(format, values...), flags: flags}
}

func newFlags(name string) *flag.FlagSet {
	flags := flag.NewFlagSet(name, flag.ContinueOnError)
	flags.SetOutput(io.Discard)
	return flags
}

func parse(flags *flag.FlagSet, args []string, positionals int) ([]string, error) {
	var found []string
	for {
		err := flags.Parse(args)
		if errors.Is(err, flag.ErrHelp) {
			return nil, helpRequested{flags: flags}
		}
		if err != nil {
			return nil, usageOfFlags(flags, "%s", err)
		}
		if flags.NArg() == 0 {
			break
		}
		found = append(found, flags.Arg(0))
		args = flags.Args()[1:]
	}
	if len(found) != positionals {
		return nil, usageOfFlags(flags, "expected %d argument(s), got %d", positionals, len(found))
	}
	return found, nil
}
