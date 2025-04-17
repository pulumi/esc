package main

import (
	"bufio"
	"bytes"
	"context"
	"errors"
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"slices"
	"strconv"
	"strings"
	"time"

	"github.com/pulumi/esc/cmd/esc/cli/client"
)

func main() {
	ctx := context.Background()
	if len(os.Args) < 3 {
		log.Fatal("Usage: git-remote-esc <remote> esc://<org>/<proj>/<env>")
	}
	if os.Getenv("GIT_DIR") == "" {
		log.Fatal("GIT_DIR must be set")
	}

	// decode url into env ref
	envRef := strings.Split(strings.TrimPrefix(os.Args[2], "esc://"), "/")
	orgName := envRef[0]
	projectName := envRef[1]
	envName := envRef[2]

	// esc client
	apiURL := os.Getenv("PULUMI_BACKEND_URL")
	if apiURL == "" {
		apiURL = "https://api.pulumi.com"
	}
	apiToken := os.Getenv("PULUMI_API_TOKEN")
	if apiToken == "" {
		log.Fatal("PULUMI_API_TOKEN must be set")
	}
	esc := client.New(fmt.Sprintf("git-remote-esc/1 (%s; %s)", "<wip>", runtime.GOOS), apiURL, apiToken, false)

	// remote-helper command loop
	scanner := bufio.NewScanner(os.Stdin)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" {
			continue
		}
		parts := strings.Split(line, " ")
		command := parts[0]
		args := parts[1:]
		//log.Println(parts)

		switch command {
		case "capabilities":
			fmt.Println("fetch")
			fmt.Println("push")

		case "list":
			// we need to synthesize commits from revisions during `list` to even be able to return commit hashes
			revMap := getRevisionToCommitMap()
			latestLocal := len(revMap)
			latestUpstream, err := esc.GetRevisionNumber(ctx, orgName, projectName, envName, "latest")
			if err != nil {
				log.Fatal(err)
			}

			toFetch := latestUpstream - latestLocal
			//log.Println(revMap, latestLocal, latestUpstream, toFetch)
			if toFetch > 0 {
				revisions, err := esc.ListEnvironmentRevisions(ctx, orgName, projectName, envName, client.ListEnvironmentRevisionsOptions{Count: &toFetch})
				if err != nil {
					log.Fatal(err)
				}
				slices.Reverse(revisions)

				parentCommit := revMap[latestLocal]
				for _, revision := range revisions {
					log.Println(revision)
					// create a commit for this revision
					yaml, _, _, err := esc.GetEnvironment(ctx, orgName, projectName, envName, strconv.Itoa(revision.Number), false)
					if err != nil {
						log.Fatal(err)
					}
					blobHash := createBlob(yaml)
					treeHash := createTree(map[string]string{"environment.yaml": blobHash})
					commitHash := createCommit(treeHash, parentCommit, revision)

					// keep track of rev num -> commit mappings
					revMap[revision.Number] = commitHash
					parentCommit = commitHash
				}
			}

			forPush := len(args) > 0 && args[0] == "for-push"
			if forPush {
				// for push, we're only interested updating the main branch, so stop here
				fmt.Printf("%s refs/heads/latest\n", revMap[latestUpstream])
				break
			}

			// create tags mapping revision number -> commit
			for revNum, commitHash := range revMap {
				fmt.Printf("%s refs/tags/%d\n", commitHash, revNum)
			}

			// map revision tags to branches so they can float without git complaining
			// see https://git-scm.com/docs/git-tag#_on_re_tagging
			tags, _ := esc.ListEnvironmentRevisionTags(ctx, orgName, projectName, envName, client.ListEnvironmentRevisionTagsOptions{})
			for _, tag := range tags {
				fmt.Printf("%s refs/heads/%s\n", revMap[tag.Revision], tag.Name)
			}

			// use latest as default branch
			fmt.Println("@refs/heads/latest HEAD")

		case "fetch":
			// the real fetching happens during list so fetch is really just a formality to update the local refs
			commitHash := args[0]
			refName := args[1]
			gitCommand("update-ref", refName, commitHash)

		case "push":
			for _, pushSpec := range args {
				src, dst, _ := strings.Cut(pushSpec, ":")
				if dst != "refs/heads/latest" {
					// todo: moving revision tags
					fmt.Printf("error %s not supported for push\n", pushSpec)
					continue
				}

				// we'll always just squash all pending changes into a single revision
				// so just create a revision using the latest commit content
				content := getEnvironmentDefinitionFromCommit(src)
				// todo: make sure etag matches to avoid clobbering
				diag, _, err := esc.UpdateEnvironmentWithRevision(ctx, orgName, projectName, envName, []byte(content), "")
				if err != nil {
					fmt.Println("error", dst, diag)
					break
				}
				if diag != nil {
					fmt.Println("error", dst, "diags:", diag)
					break
				}
				fmt.Println("ok", dst)

				// (at this point upstream and local commits will have diverged based on differing timestamps, so you'll need to pull to remediate)
			}

		default:
			log.Printf("Unsupported command: %s", command)
			fmt.Println("unsupported")
		}
		fmt.Println()
	}
}

// createBlob creates a blob with the given content and returns its hash
func createBlob(content []byte) string {
	cmd := exec.Command("git", "hash-object", "-w", "--stdin")
	cmd.Stdin = bytes.NewReader(content)
	blobHash, err := cmd.Output()
	if err != nil {
		fmt.Println("error", err)
		return ""
	}
	return strings.TrimSpace(string(blobHash))
}

// createTree create a git tree with the given blobs and returns its hash
// blobs is a map of paths to blob hashes
func createTree(blobs map[string]string) string {
	// create an empty index
	gitDir := os.Getenv("GIT_DIR")
	indexFile := filepath.Join(gitDir, "index.git-remote-esc.tmp")
	defer os.Remove(indexFile)
	env := append(os.Environ(), "GIT_INDEX_FILE="+indexFile)

	cmd := exec.Command("git", "read-tree", "--empty")
	cmd.Env = env
	if err := cmd.Run(); err != nil {
		log.Fatal(err)
	}

	// add blobs to index
	for path, blobHash := range blobs {
		cmd = exec.Command("git", "update-index", "--add", "--cacheinfo", fmt.Sprintf("100644,%s,%s", blobHash, path))
		cmd.Env = env
		if err := cmd.Run(); err != nil {
			log.Fatal(err)
		}
	}

	// write tree
	cmd = exec.Command("git", "write-tree")
	cmd.Env = env
	treeHash, err := cmd.Output()
	if err != nil {
		log.Fatal(err)
	}
	return strings.TrimSpace(string(treeHash))
}

// createCommit creates a git commit and returns its hash
func createCommit(treeHash, parentCommitHash string, revision client.EnvironmentRevision) string {
	email := revision.CreatorLogin
	name := revision.CreatorName
	if name == "" {
		name = email
	}
	date := revision.Created.Format(time.RFC3339)
	message := ""

	env := os.Environ()
	env = append(env, "GIT_AUTHOR_NAME="+name)
	env = append(env, "GIT_AUTHOR_EMAIL="+email)
	env = append(env, "GIT_COMMITTER_NAME="+name)
	env = append(env, "GIT_COMMITTER_EMAIL="+email)
	env = append(env, "GIT_AUTHOR_DATE="+date)
	env = append(env, "GIT_COMMITTER_DATE="+date)

	cmd := exec.Command("git", "commit-tree", treeHash, "-m", message)
	if parentCommitHash != "" {
		cmd.Args = append(cmd.Args, "-p", parentCommitHash)
	}
	cmd.Env = env
	commitHash, err := cmd.Output()
	if err != nil {
		var exitErr *exec.ExitError
		if errors.As(err, &exitErr) {
			log.Fatal(exitErr, string(exitErr.Stderr))
		}
		log.Fatal(err)
	}
	return strings.TrimSpace(string(commitHash))
}

// gitCommand runs a git command and returns its output
func gitCommand(args ...string) string {
	result, err := exec.Command("git", args...).Output()
	if err != nil {
		var exitErr *exec.ExitError
		if errors.As(err, &exitErr) {
			log.Fatal(exitErr, string(exitErr.Stderr))
		}
		log.Fatal(err)
	}
	return strings.TrimSpace(string(result))
}

// getEnvironmentDefinitionFromCommit parses a commit and returns its the environment definition
func getEnvironmentDefinitionFromCommit(commitHash string) string {
	// get tree hash from commit
	treeHash := gitCommand("show", "--no-patch", "--format=%T", commitHash)

	// list tree contents
	treeContent := gitCommand("ls-tree", treeHash)

	// find the environment file
	for _, line := range strings.Split(treeContent, "\n") {
		var mode, objType, objHash, name string
		if _, err := fmt.Sscanln(line, &mode, &objType, &objHash, &name); err != nil {
			log.Fatal(err)
		}

		if name == "environment.yaml" {
			return gitCommand("cat-file", "blob", objHash)
		}
	}

	log.Fatal("unable to find environment definition")
	return ""
}

// getRevisionToCommitMap recovers the esc revision -> git commit mapping based on tags in the local repo
func getRevisionToCommitMap() map[int]string {
	revMapping := map[int]string{}

	cmd := exec.Command("git", "show-ref", "--tags")
	tags, err := cmd.Output()
	if err != nil {
		// show-ref exits with code 1 if there are no matches, so just return the empty map
		return revMapping
	}

	for _, line := range strings.Split(string(tags), "\n") {
		commitHash, refName, ok := strings.Cut(line, " ")
		if !ok {
			continue
		}

		revNum, err := strconv.Atoi(strings.TrimPrefix(refName, "refs/tags/"))
		if err != nil {
			continue
		}
		revMapping[revNum] = commitHash
	}
	return revMapping
}
