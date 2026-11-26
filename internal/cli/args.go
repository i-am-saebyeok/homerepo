package cli

import (
	"flag"
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

func newFlags(name string) *flag.FlagSet {
	flags := flag.NewFlagSet(name, flag.ContinueOnError)
	flags.SetOutput(io.Discard)
	return flags
}

func parse(flags *flag.FlagSet, args []string, positionals int) ([]string, error) {
	var found []string
	for {
		if err := flags.Parse(args); err != nil {
			return nil, usagef("%s: %s", flags.Name(), err)
		}
		if flags.NArg() == 0 {
			break
		}
		found = append(found, flags.Arg(0))
		args = flags.Args()[1:]
	}
	if len(found) != positionals {
		return nil, usagef("%s: expected %d argument(s), got %d", flags.Name(), positionals, len(found))
	}
	return found, nil
}
