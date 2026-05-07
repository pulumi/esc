// Copyright 2023, Pulumi Corporation.

package cli

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"gopkg.in/yaml.v3"
	"mvdan.cc/sh/v3/shell"
)

// dotESC is the parsed shape of a .esc.yaml file. See `esc env --help` for the user-facing schema.
type dotESC struct {
	Environment *dotESCEnv `yaml:"environment,omitempty"`
}

// dotESCEnv decodes from a string, an organization+imports object, or a command object. The
// UnmarshalYAML method picks the form based on the shape of the value; exactly one of the three
// fields will be set.
type dotESCEnv struct {
	Environment string
	Imports     *dotESCImports
	Command     *dotESCCommand
}

type dotESCImports struct {
	Organization string   `yaml:"organization"`
	Imports      []string `yaml:"imports"`
}

type dotESCCommand struct {
	Command string `yaml:"command"`
}

func (e *dotESCEnv) UnmarshalYAML(n *yaml.Node) error {
	stringErr := n.Decode(&e.Environment)
	if stringErr == nil {
		return nil
	}

	var command dotESCCommand
	commandErr := n.Decode(&command)
	if commandErr == nil && command.Command != "" {
		e.Command = &command
		return nil
	}

	var imports dotESCImports
	importsErr := n.Decode(&imports)
	if importsErr == nil && (imports.Organization != "" || len(imports.Imports) != 0) {
		e.Imports = &imports
		return nil
	}

	return errors.Join(stringErr, commandErr, importsErr)
}

// inferCommandEnv runs command and decodes its stdout as a JSON-encoded environment reference or
// import list.
func (cmd *envCommand) inferCommandEnv(ctx context.Context, command string) (environmentDesc, error) {
	fields, err := shell.Fields(command, nil)
	if err != nil {
		return nil, fmt.Errorf("parsing default environment command: %w", err)
	}
	if len(fields) == 0 {
		return nil, fmt.Errorf("default environment command is empty")
	}

	// The command comes from a .esc.yaml file under the user's control, just like the editor in
	// `esc env edit`. Running it is the whole point.
	//nolint:gosec
	out, err := cmd.esc.exec.Output(exec.CommandContext(ctx, fields[0], fields[1:]...))
	if err != nil {
		return nil, fmt.Errorf("running %q: %w", command, err)
	}

	var environment string
	if err := json.Unmarshal(out, &environment); err == nil {
		return cmd.parseRef(environment), nil
	}

	var imports dotESCImports
	if err := json.Unmarshal(out, &imports); err == nil {
		return importList{orgName: imports.Organization, imports: imports.Imports}, nil
	}

	return nil, fmt.Errorf(
		"parsing default environment command output: must be string | "+
			"{organization: string, imports: []string}, got %q", string(out))
}

// inferFSEnv walks up from startDir looking for a .esc.yaml that names a default environment.
func (cmd *envCommand) inferFSEnv(ctx context.Context, startDir string) (environmentDesc, error) {
	dotESC, err := cmd.findDotESC(startDir)
	if err != nil || dotESC == nil || dotESC.Environment == nil {
		return nil, err
	}

	switch {
	case dotESC.Environment.Environment != "":
		return cmd.parseRef(dotESC.Environment.Environment), nil
	case dotESC.Environment.Imports != nil:
		return importList{
			orgName: dotESC.Environment.Imports.Organization,
			imports: dotESC.Environment.Imports.Imports,
		}, nil
	case dotESC.Environment.Command != nil:
		return cmd.inferCommandEnv(ctx, dotESC.Environment.Command.Command)
	default:
		return nil, nil
	}
}

// findDotESC walks up from startDir and decodes the first .esc.yaml it finds.
func (cmd *envCommand) findDotESC(startDir string) (*dotESC, error) {
	dir := startDir
	var dotESCFile fs.File
	for {
		path := filepath.Join(dir, ".esc.yaml")
		f, err := cmd.esc.fs.Open(path)
		if err == nil {
			dotESCFile = f
			break
		}
		if !errors.Is(err, fs.ErrNotExist) {
			return nil, fmt.Errorf("opening %v: %w", path, err)
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			return nil, nil
		}
		dir = parent
	}
	defer dotESCFile.Close()

	var dotESC dotESC
	if err := yaml.NewDecoder(dotESCFile).Decode(&dotESC); err != nil {
		return nil, fmt.Errorf("decoding .esc.yaml: %w", err)
	}
	return &dotESC, nil
}

// inferPulumiIaCEnv resolves the imports of the currently-selected Pulumi IaC stack as an
// anonymous environment.
func (cmd *envCommand) inferPulumiIaCEnv(ctx context.Context) (environmentDesc, error) {
	// Imported environments come from the selected stack's config.
	out, err := cmd.esc.exec.Output(exec.CommandContext(ctx, "pulumi", "config", "env", "ls", "-j"))
	if err != nil {
		return nil, fmt.Errorf(`running "pulumi config env ls -j": %w`, err)
	}
	var imports []string
	if err := json.Unmarshal(out, &imports); err != nil {
		return nil, fmt.Errorf("unmarshaling environment list: %w", err)
	}
	if len(imports) == 0 {
		return nil, nil
	}

	// The organization comes from the selected stack's name.
	out, err = cmd.esc.exec.Output(exec.CommandContext(ctx, "pulumi", "stack", "ls", "-Q", "-j"))
	if err != nil {
		return nil, fmt.Errorf(`running "pulumi stack ls -Q -j": %w`, err)
	}

	var stacks []struct {
		Name    string `json:"name"`
		Current bool   `json:"current"`
	}
	if err := json.Unmarshal(out, &stacks); err != nil {
		return nil, fmt.Errorf("unmarshaling stack list: %w", err)
	}

	var stackName string
	for _, s := range stacks {
		if s.Current {
			stackName = s.Name
			break
		}
	}
	if stackName == "" {
		return nil, fmt.Errorf("no stack selected")
	}
	orgName, _, ok := strings.Cut(stackName, "/")
	if !ok {
		return nil, fmt.Errorf("could not determine organization for stack %q", stackName)
	}
	return importList{orgName: orgName, imports: imports}, nil
}

// inferDefaultEnv resolves the default environment from a .esc.yaml in the working directory or
// any parent, falling back to the imports of the currently-selected Pulumi IaC stack.
func (cmd *envCommand) inferDefaultEnv(ctx context.Context) (environmentDesc, error) {
	dir := cmd.esc.cwd
	if dir == "" {
		var err error
		dir, err = os.Getwd()
		if err != nil {
			return nil, fmt.Errorf("getting working directory: %w", err)
		}
	}

	env, err := cmd.inferFSEnv(ctx, dir)
	if err != nil || env != nil {
		return env, err
	}

	// The Pulumi IaC fallback is best-effort: pulumi may not be installed and the working
	// directory may not be a Pulumi project.
	env, _ = cmd.inferPulumiIaCEnv(ctx)
	return env, nil
}
