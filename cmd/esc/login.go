// Copyright 2023, Pulumi Corporation.

package main

import (
	"context"
	"errors"
	"fmt"
	"net/url"
	"os"
	"path"
	"strings"

	"github.com/spf13/cobra"

	"github.com/pulumi/esc/cmd/esc/internal/client"
	"github.com/pulumi/esc/cmd/esc/internal/workspace"
	"github.com/pulumi/pulumi/pkg/v3/backend/display"
	"github.com/pulumi/pulumi/pkg/v3/backend/filestate"
	"github.com/pulumi/pulumi/pkg/v3/backend/httpstate"
	"github.com/pulumi/pulumi/sdk/v3/go/common/util/cmdutil"
)

func newLoginCmd(esc *escCommand) *cobra.Command {
	var backendURL string
	var defaultOrg string
	var insecure bool
	var shared bool

	cmd := &cobra.Command{
		Use:   "login [<url>]",
		Short: "Log in to the Pulumi Cloud",
		Long: "Log in to the Pulumi Cloud.\n" +
			"\n" +
			"The Pulumi Cloud manages your Pulumi ESC environments. Simply run\n" +
			"\n" +
			"    $ pulumi login\n" +
			"\n" +
			"and this command will prompt you for an access token, including a way to launch your web browser to\n" +
			"easily obtain one. You can script by using `PULUMI_ACCESS_TOKEN` environment variable.\n" +
			"\n" +
			"By default, this will log in to the managed Pulumi Cloud backend.\n" +
			"If you prefer to log in to a self-hosted Pulumi Cloud backend, specify a URL. For example, run\n" +
			"\n" +
			"    $ pulumi login https://api.pulumi.acmecorp.com\n" +
			"\n" +
			"to log in to a self-hosted Pulumi Cloud running at the api.pulumi.acmecorp.com domain.\n" +
			"\n" +
			"For `https://` URLs, the CLI will speak REST to a Pulumi Cloud that manages state and concurrency control.\n" +
			"You can specify a default org to use when logging into the Pulumi Cloud backend or a " +
			"self-hosted Pulumi Cloud.\n",
		Args: cmdutil.MaximumNArgs(1),
		Run: cmdutil.RunFunc(func(cmd *cobra.Command, args []string) error {
			ctx := context.Background()

			// If a <cloud> was specified as an argument, use it.
			if len(args) > 0 {
				if backendURL != "" {
					return errors.New("only one of --cloud-url or argument URL may be specified, not both")
				}
				backendURL = args[0]
			}

			if shared && backendURL != "" {
				return errors.New("a cloud URL may not be specified with --shared")
			}

			if backendURL == "" {
				account, isShared, err := workspace.GetCurrentAccount(shared)
				if err != nil {
					return fmt.Errorf("could not determine current cloud: %w", err)
				}
				backendURL, shared = account.BackendURL, isShared
			}

			switch {
			case isInvalidSelfHostedURL(backendURL):
				return fmt.Errorf("%s is not a valid self-hosted backend, "+
					"use `pulumi login` without arguments to log into the Pulumi Cloud backend", backendURL)
			case filestate.IsFileStateBackendURL(backendURL):
				return fmt.Errorf("%s does not support Pulumi ESC.", backendURL)
			}

			account, err := httpstate.NewLoginManager().Login(
				ctx,
				backendURL,
				insecure,
				"esc",
				"Pulumi ESC environments",
				nil,
				false,
				display.Options{Color: esc.colors},
			)
			// if the user has specified a default org to associate with the backend
			if defaultOrg != "" {
				if err := workspace.SetBackendConfigDefaultOrg(backendURL, defaultOrg); err != nil {
					return err
				}
			}
			if err != nil {
				return fmt.Errorf("problem logging in: %w", err)
			}
			esc.account = workspace.Account{
				Account:    *account,
				BackendURL: backendURL,
				DefaultOrg: defaultOrg,
			}

			if err := workspace.SetCurrentAccount(esc.account, shared); err != nil {
				return fmt.Errorf("setting current account: %w", err)
			}

			backendName := backendURL
			if backendURL == pulumiCloudURL {
				backendName = "pulumi.com"
			}

			client := client.NewClient(backendURL, account.AccessToken, account.Insecure)
			if currentUser, _, _, err := client.GetPulumiAccountDetails(ctx); err == nil {
				// TODO should we print the token information here? (via team MyTeam token MyToken)
				consoleURL := cloudConsoleURL(backendURL, currentUser)
				fmt.Printf("Logged in to %s as %s (%s)\n", backendName, currentUser, consoleURL)
			} else {
				fmt.Printf("Logged in to %s (%s)\n", backendName, cloudConsoleURL(backendURL))
			}

			return nil
		}),
	}

	cmd.Flags().StringVarP(&backendURL, "cloud-url", "c", "", "A cloud URL to log in to")
	cmd.Flags().StringVar(&defaultOrg, "default-org", "", "A default org to associate with the login.")
	cmd.Flags().BoolVar(&insecure, "insecure", false, "Allow insecure server connections when using SSL")
	cmd.Flags().BoolVar(&shared, "shared", false, "Log in to the account in use by the `pulumi` CLI")

	return cmd
}

func isInvalidSelfHostedURL(url string) bool {
	url = strings.TrimPrefix(strings.TrimPrefix(url, "https://"), "http://")
	return strings.HasPrefix(url, "app.pulumi.com/") || strings.HasPrefix(url, "pulumi.com")
}

func (esc *escCommand) getCachedClient(ctx context.Context) error {
	account, _, err := workspace.GetCurrentAccount(false)
	if err != nil {
		return fmt.Errorf("could not determine current cloud: %w", err)
	}

	ok, err := esc.getCachedCredentials(ctx, account.BackendURL, account.Insecure)
	if err != nil {
		return fmt.Errorf("getting credentials: %w", err)
	}
	if !ok {
		return fmt.Errorf("no credentials. Please run `esc login` to log in.")
	}

	esc.client = client.NewClient(account.BackendURL, account.AccessToken, account.Insecure)
	return nil
}

func (esc *escCommand) getCachedCredentials(ctx context.Context, backendURL string, insecure bool) (bool, error) {
	account, err := httpstate.NewLoginManager().Current(ctx, backendURL, insecure, false)
	if err != nil {
		return false, err
	}

	defaultOrg, err := workspace.GetBackendConfigDefaultOrg(backendURL, account.Username)
	if err != nil {
		return false, err
	}

	esc.account = workspace.Account{
		Account:    *account,
		BackendURL: backendURL,
		DefaultOrg: defaultOrg,
	}
	return true, nil
}

const (
	// consoleDomainEnvVar overrides the way we infer the domain we assume the Pulumi Console will
	// be served from, and instead just use this value. e.g. so links to the stack update go to
	// https://pulumi.example.com/org/project/stack/updates/2 instead.
	consoleDomainEnvVar = "PULUMI_CONSOLE_DOMAIN"

	// pulumiCloudURL is the Cloud URL used if no environment or explicit cloud is chosen.
	pulumiCloudURL = "https://" + defaultAPIDomainPrefix + "pulumi.com"

	// defaultAPIDomainPrefix is the assumed Cloud URL prefix for typical Pulumi Cloud API endpoints.
	defaultAPIDomainPrefix = "api."
	// defaultConsoleDomainPrefix is the assumed Cloud URL prefix typically used for the Pulumi Console.
	defaultConsoleDomainPrefix = "app."
)

// cloudConsoleURL returns a URL to the Pulumi Cloud Console, rooted at cloudURL. If there is
// an error, returns "".
func cloudConsoleURL(cloudURL string, paths ...string) string {
	u, err := url.Parse(cloudURL)
	if err != nil {
		return ""
	}

	switch {
	case os.Getenv(consoleDomainEnvVar) != "":
		// Honor a PULUMI_CONSOLE_DOMAIN environment variable to override the
		// default behavior. Since we identify a backend by a single URI, we
		// cannot know what the Pulumi Console is hosted at...
		u.Host = os.Getenv(consoleDomainEnvVar)
	case strings.HasPrefix(u.Host, defaultAPIDomainPrefix):
		// ... but if the cloudURL (API domain) is "api.", then we assume the
		// console is hosted at "app.".
		u.Host = defaultConsoleDomainPrefix + u.Host[len(defaultAPIDomainPrefix):]
	case u.Host == "localhost:8080":
		// ... or when running locally, on port 3000.
		u.Host = "localhost:3000"
	default:
		// We couldn't figure out how to convert the api hostname into a console hostname.
		// We return "" so that the caller can know to omit the URL rather than just
		// return an incorrect one.
		return ""
	}

	u.Path = path.Join(paths...)
	return u.String()
}
