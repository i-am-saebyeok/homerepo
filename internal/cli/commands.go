package cli

var commands = map[string]handler{
	"init":              runInit,
	"repo list":         runRepoList,
	"repo path":         runRepoPath,
	"workspace create":  runWorkspaceCreate,
	"workspace list":    runWorkspaceList,
	"workspace release": runWorkspaceRelease,
}
