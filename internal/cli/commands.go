package cli

var commands = map[string]handler{
	"init":              runInit,
	"repo list":         runRepoList,
	"repo path":         runRepoPath,
	"workspace create":  runWorkspaceCreate,
	"workspace list":    runWorkspaceList,
	"workspace release": runWorkspaceRelease,
	"thread create":     runThreadCreate,
	"thread list":       runThreadList,
	"thread close":      runThreadClose,
}
