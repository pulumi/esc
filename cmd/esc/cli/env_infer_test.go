// Copyright 2023, Pulumi Corporation.

package cli

import (
	"bytes"
	"context"
	"errors"
	"io"
	"os/exec"
	"path/filepath"
	"testing"
	"testing/fstest"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/pulumi/esc/cmd/esc/cli/client"
	"github.com/pulumi/esc/cmd/esc/cli/workspace"
)

// inferFixture wires up an envCommand with the minimum dependencies needed by the inference code:
// a virtual filesystem, an account (for parseRef defaults), and a stub exec for command-form
// inference and the Pulumi IaC fallback.
type inferFixture struct {
	cmd *envCommand
	fs  testFS
	exe *inferExec
}

func newInferFixture() *inferFixture {
	fs := testFS{MapFS: fstest.MapFS{}}
	exe := &inferExec{
		results: map[string]inferExecResult{},
	}
	esc := &escCommand{
		fs:      fs,
		exec:    exe,
		account: workspace.Account{DefaultOrg: "default-org"},
	}
	return &inferFixture{
		cmd: &envCommand{esc: esc},
		fs:  fs,
		exe: exe,
	}
}

// writeFile adds a file to the fixture's virtual filesystem at the given (slash-separated, no
// leading slash) path.
func (f *inferFixture) writeFile(path, contents string) {
	f.fs.MapFS[path] = &fstest.MapFile{Data: []byte(contents), Mode: 0o600}
}

// inferExecResult is the canned result for a single Output() invocation, keyed by the joined
// command line in inferExec.results.
type inferExecResult struct {
	stdout string
	err    error
}

type inferExec struct {
	results map[string]inferExecResult
	calls   []string
}

func (e *inferExec) LookPath(command string) (string, error) {
	return command, nil
}

func (e *inferExec) Run(cmd *exec.Cmd) error {
	_, err := e.Output(cmd)
	return err
}

func (e *inferExec) Output(cmd *exec.Cmd) ([]byte, error) {
	key := commandKey(cmd)
	e.calls = append(e.calls, key)
	r, ok := e.results[key]
	if !ok {
		return nil, errors.New("command not found: " + key)
	}
	if cmd.Stdout != nil {
		_, _ = io.Copy(cmd.Stdout, bytes.NewReader([]byte(r.stdout)))
	}
	return []byte(r.stdout), r.err
}

func commandKey(cmd *exec.Cmd) string {
	args := append([]string{filepath.Base(cmd.Path)}, cmd.Args[1:]...)
	return shellJoin(args)
}

func shellJoin(args []string) string {
	var b bytes.Buffer
	for i, a := range args {
		if i > 0 {
			b.WriteByte(' ')
		}
		b.WriteString(a)
	}
	return b.String()
}

func TestInferFSEnv_StringRef(t *testing.T) {
	f := newInferFixture()
	f.writeFile("home/user/project/.esc.yaml", "environment: my-org/my-project/my-env@v1\n")

	desc, err := f.cmd.inferFSEnv(context.Background(), "home/user/project")
	require.NoError(t, err)

	ref, ok := desc.(environmentRef)
	require.True(t, ok, "expected environmentRef, got %T", desc)
	assert.Equal(t, "my-org", ref.orgName)
	assert.Equal(t, "my-project", ref.projectName)
	assert.Equal(t, "my-env", ref.envName)
	assert.Equal(t, "v1", ref.version)
}

func TestInferFSEnv_LegacyOneIdentifier(t *testing.T) {
	f := newInferFixture()
	f.writeFile("home/user/.esc.yaml", "environment: my-env\n")

	desc, err := f.cmd.inferFSEnv(context.Background(), "home/user")
	require.NoError(t, err)

	ref, ok := desc.(environmentRef)
	require.True(t, ok)
	assert.Equal(t, "default-org", ref.orgName)
	assert.Equal(t, client.DefaultProject, ref.projectName)
	assert.Equal(t, "my-env", ref.envName)
}

func TestInferFSEnv_Imports(t *testing.T) {
	f := newInferFixture()
	f.writeFile("home/user/.esc.yaml", `environment:
  organization: my-org
  imports:
    - my-project/my-env
    - other-project/other-env
`)

	desc, err := f.cmd.inferFSEnv(context.Background(), "home/user")
	require.NoError(t, err)

	list, ok := desc.(importList)
	require.True(t, ok, "expected importList, got %T", desc)
	assert.Equal(t, "my-org", list.orgName)
	assert.Equal(t, []string{"my-project/my-env", "other-project/other-env"}, list.imports)
}

func TestInferFSEnv_Command(t *testing.T) {
	f := newInferFixture()
	f.writeFile("home/user/.esc.yaml", `environment:
  command: my-tool resolve-env
`)
	f.exe.results["my-tool resolve-env"] = inferExecResult{stdout: `"my-org/my-project/my-env"`}

	desc, err := f.cmd.inferFSEnv(context.Background(), "home/user")
	require.NoError(t, err)

	ref, ok := desc.(environmentRef)
	require.True(t, ok)
	assert.Equal(t, "my-org", ref.orgName)
	assert.Equal(t, "my-project", ref.projectName)
	assert.Equal(t, "my-env", ref.envName)
}

func TestInferFSEnv_WalksUp(t *testing.T) {
	f := newInferFixture()
	f.writeFile("home/user/.esc.yaml", "environment: my-env\n")

	desc, err := f.cmd.inferFSEnv(context.Background(), "home/user/project/sub/dir")
	require.NoError(t, err)

	ref, ok := desc.(environmentRef)
	require.True(t, ok)
	assert.Equal(t, "my-env", ref.envName)
}

func TestInferFSEnv_NoFile(t *testing.T) {
	f := newInferFixture()

	desc, err := f.cmd.inferFSEnv(context.Background(), "home/user/project")
	require.NoError(t, err)
	assert.Nil(t, desc)
}

func TestInferFSEnv_EmptyEnvironmentField(t *testing.T) {
	f := newInferFixture()
	f.writeFile("home/user/.esc.yaml", "environment:\n")

	desc, err := f.cmd.inferFSEnv(context.Background(), "home/user")
	require.NoError(t, err)
	assert.Nil(t, desc)
}

func TestInferFSEnv_InvalidYAML(t *testing.T) {
	f := newInferFixture()
	f.writeFile("home/user/.esc.yaml", "environment: { unclosed\n")

	_, err := f.cmd.inferFSEnv(context.Background(), "home/user")
	assert.Error(t, err)
}

func TestInferFSEnv_NearestWins(t *testing.T) {
	// A .esc.yaml in a closer ancestor takes precedence over one further up the tree.
	f := newInferFixture()
	f.writeFile("home/.esc.yaml", "environment: outer-env\n")
	f.writeFile("home/user/.esc.yaml", "environment: inner-env\n")

	desc, err := f.cmd.inferFSEnv(context.Background(), "home/user/project")
	require.NoError(t, err)

	ref, ok := desc.(environmentRef)
	require.True(t, ok)
	assert.Equal(t, "inner-env", ref.envName)
}

func TestInferCommandEnv_StringOutput(t *testing.T) {
	f := newInferFixture()
	f.exe.results["my-tool"] = inferExecResult{stdout: `"my-org/my-project/my-env"`}

	desc, err := f.cmd.inferCommandEnv(context.Background(), "my-tool")
	require.NoError(t, err)

	ref, ok := desc.(environmentRef)
	require.True(t, ok)
	assert.Equal(t, "my-env", ref.envName)
}

func TestInferCommandEnv_ImportsOutput(t *testing.T) {
	f := newInferFixture()
	f.exe.results["my-tool resolve"] = inferExecResult{
		stdout: `{"organization":"my-org","imports":["a/b","c/d"]}`,
	}

	desc, err := f.cmd.inferCommandEnv(context.Background(), "my-tool resolve")
	require.NoError(t, err)

	list, ok := desc.(importList)
	require.True(t, ok)
	assert.Equal(t, "my-org", list.orgName)
	assert.Equal(t, []string{"a/b", "c/d"}, list.imports)
}

func TestInferCommandEnv_UnparseableOutput(t *testing.T) {
	f := newInferFixture()
	f.exe.results["my-tool"] = inferExecResult{stdout: `not json`}

	_, err := f.cmd.inferCommandEnv(context.Background(), "my-tool")
	assert.ErrorContains(t, err, "parsing default environment command output")
}

func TestInferCommandEnv_EmptyCommand(t *testing.T) {
	f := newInferFixture()

	_, err := f.cmd.inferCommandEnv(context.Background(), "   ")
	assert.ErrorContains(t, err, "default environment command is empty")
}

func TestInferCommandEnv_CommandFails(t *testing.T) {
	f := newInferFixture()
	f.exe.results["my-tool"] = inferExecResult{err: errors.New("boom")}

	_, err := f.cmd.inferCommandEnv(context.Background(), "my-tool")
	assert.ErrorContains(t, err, "running")
	assert.ErrorContains(t, err, "boom")
}

func TestInferPulumiIaCEnv_NoImports(t *testing.T) {
	f := newInferFixture()
	f.exe.results["pulumi config env ls -j"] = inferExecResult{stdout: `[]`}

	desc, err := f.cmd.inferPulumiIaCEnv(context.Background())
	require.NoError(t, err)
	assert.Nil(t, desc)
}

func TestInferPulumiIaCEnv_OK(t *testing.T) {
	f := newInferFixture()
	f.exe.results["pulumi config env ls -j"] = inferExecResult{stdout: `["my-project/my-env"]`}
	f.exe.results["pulumi stack ls -Q -j"] = inferExecResult{
		stdout: `[{"name":"other-org/other-project/dev","current":false},
			    {"name":"my-org/my-project/prod","current":true}]`,
	}

	desc, err := f.cmd.inferPulumiIaCEnv(context.Background())
	require.NoError(t, err)

	list, ok := desc.(importList)
	require.True(t, ok)
	assert.Equal(t, "my-org", list.orgName)
	assert.Equal(t, []string{"my-project/my-env"}, list.imports)
}

func TestInferPulumiIaCEnv_NoCurrentStack(t *testing.T) {
	f := newInferFixture()
	f.exe.results["pulumi config env ls -j"] = inferExecResult{stdout: `["my-project/my-env"]`}
	f.exe.results["pulumi stack ls -Q -j"] = inferExecResult{stdout: `[]`}

	_, err := f.cmd.inferPulumiIaCEnv(context.Background())
	assert.ErrorContains(t, err, "no stack selected")
}

func TestInferPulumiIaCEnv_StackNameWithoutOrg(t *testing.T) {
	f := newInferFixture()
	f.exe.results["pulumi config env ls -j"] = inferExecResult{stdout: `["my-env"]`}
	f.exe.results["pulumi stack ls -Q -j"] = inferExecResult{
		stdout: `[{"name":"dev","current":true}]`,
	}

	_, err := f.cmd.inferPulumiIaCEnv(context.Background())
	assert.ErrorContains(t, err, "could not determine organization")
}

func TestInferPulumiIaCEnv_PulumiNotFound(t *testing.T) {
	f := newInferFixture()
	// No pulumi command registered: Output returns an error.

	_, err := f.cmd.inferPulumiIaCEnv(context.Background())
	assert.Error(t, err)
}

func TestInferDefaultEnv_FSWins(t *testing.T) {
	// When both .esc.yaml and Pulumi context are available, .esc.yaml wins.
	f := newInferFixture()
	f.cmd.esc.cwd = "home/user"
	f.writeFile("home/user/.esc.yaml", "environment: fs-env\n")
	f.exe.results["pulumi config env ls -j"] = inferExecResult{stdout: `["a/b"]`}
	f.exe.results["pulumi stack ls -Q -j"] = inferExecResult{
		stdout: `[{"name":"my-org/my-project/dev","current":true}]`,
	}

	desc, err := f.cmd.inferDefaultEnv(context.Background())
	require.NoError(t, err)

	ref, ok := desc.(environmentRef)
	require.True(t, ok)
	assert.Equal(t, "fs-env", ref.envName)
}

func TestInferDefaultEnv_FallsBackToPulumi(t *testing.T) {
	f := newInferFixture()
	f.cmd.esc.cwd = "home/user"
	f.exe.results["pulumi config env ls -j"] = inferExecResult{stdout: `["a/b"]`}
	f.exe.results["pulumi stack ls -Q -j"] = inferExecResult{
		stdout: `[{"name":"my-org/my-project/dev","current":true}]`,
	}

	desc, err := f.cmd.inferDefaultEnv(context.Background())
	require.NoError(t, err)

	list, ok := desc.(importList)
	require.True(t, ok)
	assert.Equal(t, "my-org", list.orgName)
}

func TestInferDefaultEnv_NoSourceReturnsNil(t *testing.T) {
	f := newInferFixture()
	f.cmd.esc.cwd = "home/user"
	// No .esc.yaml; pulumi command not registered, error is suppressed.

	desc, err := f.cmd.inferDefaultEnv(context.Background())
	require.NoError(t, err)
	assert.Nil(t, desc)
}
