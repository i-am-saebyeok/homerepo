package cli

var commands = map[string]handler{
	"init":      runInit,
	"repo list": runRepoList,
	"repo path": runRepoPath,
}
