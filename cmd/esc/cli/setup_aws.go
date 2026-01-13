// Copyright 2026, Pulumi Corporation.

package cli

import (
	"context"
	"crypto/sha1" //nolint:gosec // Required for AWS OIDC provider thumbprint calculation
	"crypto/tls"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/iam"
	iamtypes "github.com/aws/aws-sdk-go-v2/service/iam/types"
	"github.com/aws/aws-sdk-go-v2/service/sts"
	"github.com/spf13/cobra"

	"github.com/pulumi/pulumi/sdk/v3/go/common/apitype"
)

const (
	oidcIssuerURL  = "https://api.pulumi.com/oidc"
	oidcIssuerHost = "api.pulumi.com/oidc"
)

// iamClient defines the IAM operations needed for AWS OIDC setup.
type iamClient interface {
	CreateOpenIDConnectProvider(ctx context.Context, params *iam.CreateOpenIDConnectProviderInput, optFns ...func(*iam.Options)) (*iam.CreateOpenIDConnectProviderOutput, error)
	AddClientIDToOpenIDConnectProvider(ctx context.Context, params *iam.AddClientIDToOpenIDConnectProviderInput, optFns ...func(*iam.Options)) (*iam.AddClientIDToOpenIDConnectProviderOutput, error)
	CreateRole(ctx context.Context, params *iam.CreateRoleInput, optFns ...func(*iam.Options)) (*iam.CreateRoleOutput, error)
	AttachRolePolicy(ctx context.Context, params *iam.AttachRolePolicyInput, optFns ...func(*iam.Options)) (*iam.AttachRolePolicyOutput, error)
}

func newSetupAWSCmd(setup *setupCommand) *cobra.Command {
	var roleName string
	var policy string
	var envName string
	var orgName string

	cmd := &cobra.Command{
		Use:   "aws",
		Short: "Setup AWS OIDC integration for Pulumi ESC",
		Long: "Setup AWS OIDC integration for Pulumi ESC\n" +
			"\n" +
			"This command creates the necessary AWS resources for OIDC authentication:\n" +
			"  - An OIDC identity provider for Pulumi Cloud\n" +
			"  - An IAM role with a trust policy for your organization\n" +
			"  - A policy attachment for the specified policy\n" +
			"\n" +
			"AWS credentials must be configured in your environment (via AWS_ACCESS_KEY_ID,\n" +
			"AWS_SECRET_ACCESS_KEY, or other standard AWS credential methods).\n" +
			"\n" +
			"Example:\n" +
			"  esc setup aws --role-name PulumiESCRole\n" +
			"  esc setup aws --role-name PulumiESCRole --policy ReadOnlyAccess\n" +
			"  esc setup aws --role-name PulumiESCRole --org myorg\n" +
			"  esc setup aws --role-name PulumiESCRole --environment myorg/myproject/aws-dev\n",
		SilenceUsage: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := context.Background()

			if err := setup.esc.getCachedClient(ctx); err != nil {
				return err
			}

			if roleName == "" {
				return fmt.Errorf("--role-name is required")
			}

			if orgName == "" {
				orgName = setup.esc.account.DefaultOrg
			}
			if orgName == "" {
				return fmt.Errorf("could not determine organization; please specify --org or set a default org with 'esc login --default-org <org>'")
			}

			result, err := setupAWSOIDC(ctx, setup.esc, orgName, roleName, policy)
			if err != nil {
				return err
			}

			if envName != "" {
				return createOrUpdateEnvironment(ctx, setup.esc, orgName, envName, result.roleArn)
			}

			printEnvironmentYAML(setup.esc, result.roleArn)
			return nil
		},
	}

	cmd.Flags().StringVar(&roleName, "role-name", "", "The name of the IAM role to create (required)")
	cmd.Flags().StringVar(&policy, "policy", "AdministratorAccess", "The AWS managed policy name to attach to the role")
	cmd.Flags().StringVar(&orgName, "org", "", "The Pulumi organization to configure OIDC for (defaults to current org)")
	cmd.Flags().StringVar(&envName, "environment", "", "Create or update an ESC environment with the OIDC configuration")

	return cmd
}

type awsSetupResult struct {
	accountID        string
	partition        string
	roleArn          string
	oidcProviderArn  string
	oidcProviderNew  bool
	roleNew          bool
	policyAttachment string
}

func setupAWSOIDC(ctx context.Context, esc *escCommand, orgName, roleName, policyName string) (*awsSetupResult, error) {
	cfg, err := config.LoadDefaultConfig(ctx)
	if err != nil {
		return nil, fmt.Errorf("loading AWS configuration: %w", err)
	}

	stsClient := sts.NewFromConfig(cfg)
	iamClient := iam.NewFromConfig(cfg)

	identity, err := stsClient.GetCallerIdentity(ctx, &sts.GetCallerIdentityInput{})
	if err != nil {
		return nil, fmt.Errorf("getting AWS caller identity: %w", err)
	}

	accountID := *identity.Account
	partition := parsePartitionFromARN(*identity.Arn)

	fmt.Fprintf(esc.stdout, "AWS Account: %s (partition: %s)\n", accountID, partition)
	fmt.Fprintf(esc.stdout, "Organization: %s\n\n", orgName)

	thumbprint, err := getOIDCThumbprint(oidcIssuerHost)
	if err != nil {
		return nil, fmt.Errorf("getting OIDC thumbprint: %w", err)
	}

	audience := fmt.Sprintf("aws:%s", orgName)
	oidcProviderArn := fmt.Sprintf("arn:%s:iam::%s:oidc-provider/%s", partition, accountID, oidcIssuerHost)

	oidcProviderNew, err := createOrUpdateOIDCProvider(ctx, esc, iamClient, oidcProviderArn, audience, thumbprint)
	if err != nil {
		return nil, err
	}

	roleArn := fmt.Sprintf("arn:%s:iam::%s:role/%s", partition, accountID, roleName)
	roleNew, err := createRole(ctx, esc, iamClient, roleName, oidcProviderArn, audience)
	if err != nil {
		return nil, err
	}

	policyArn := fmt.Sprintf("arn:%s:iam::aws:policy/%s", partition, policyName)
	if err := attachRolePolicy(ctx, esc, iamClient, roleName, policyArn); err != nil {
		return nil, err
	}

	return &awsSetupResult{
		accountID:        accountID,
		partition:        partition,
		roleArn:          roleArn,
		oidcProviderArn:  oidcProviderArn,
		oidcProviderNew:  oidcProviderNew,
		roleNew:          roleNew,
		policyAttachment: policyArn,
	}, nil
}

func parsePartitionFromARN(arn string) string {
	parts := strings.Split(arn, ":")
	if len(parts) >= 2 {
		return parts[1]
	}
	return "aws"
}

func getOIDCThumbprint(host string) (string, error) {
	hostWithPort := host
	if !strings.Contains(host, ":") {
		hostParts := strings.Split(host, "/")
		hostWithPort = hostParts[0] + ":443"
	}

	conn, err := tls.Dial("tcp", hostWithPort, &tls.Config{
		MinVersion: tls.VersionTLS12,
	})
	if err != nil {
		return "", fmt.Errorf("connecting to %s: %w", hostWithPort, err)
	}
	defer conn.Close()

	certs := conn.ConnectionState().PeerCertificates
	if len(certs) == 0 {
		return "", fmt.Errorf("no certificates returned from %s", host)
	}

	rootCert := certs[len(certs)-1]
	thumbprint := sha1.Sum(rootCert.Raw) //nolint:gosec // Required for AWS OIDC provider thumbprint calculation

	var thumbprintStr strings.Builder
	for _, b := range thumbprint {
		fmt.Fprintf(&thumbprintStr, "%02x", b)
	}

	return thumbprintStr.String(), nil
}

func createOrUpdateOIDCProvider(
	ctx context.Context,
	esc *escCommand,
	iamCli iamClient,
	oidcProviderArn, audience, thumbprint string,
) (bool, error) {
	_, err := iamCli.CreateOpenIDConnectProvider(ctx, &iam.CreateOpenIDConnectProviderInput{
		Url:            aws.String(oidcIssuerURL),
		ClientIDList:   []string{audience},
		ThumbprintList: []string{thumbprint},
	})

	if err != nil {
		var entityExists *iamtypes.EntityAlreadyExistsException
		if errors.As(err, &entityExists) {
			fmt.Fprintf(esc.stdout, "OIDC Provider: %s (existing)\n", oidcProviderArn)

			_, err := iamCli.AddClientIDToOpenIDConnectProvider(ctx, &iam.AddClientIDToOpenIDConnectProviderInput{
				OpenIDConnectProviderArn: aws.String(oidcProviderArn),
				ClientID:                 aws.String(audience),
			})
			if err != nil {
				var invalidInput *iamtypes.InvalidInputException
				if errors.As(err, &invalidInput) && strings.Contains(err.Error(), "already registered") {
					fmt.Fprintf(esc.stdout, "  Audience '%s' already configured\n", audience)
					return false, nil
				}
				return false, fmt.Errorf("adding audience to OIDC provider: %w", err)
			}

			fmt.Fprintf(esc.stdout, "  Added audience: %s\n", audience)
			return false, nil
		}
		return false, fmt.Errorf("creating OIDC provider: %w", err)
	}

	fmt.Fprintf(esc.stdout, "OIDC Provider: %s (created)\n", oidcProviderArn)
	return true, nil
}

func createRole(
	ctx context.Context,
	esc *escCommand,
	iamCli iamClient,
	roleName, oidcProviderArn, audience string,
) (bool, error) {
	trustPolicy := map[string]interface{}{
		"Version": "2012-10-17",
		"Statement": []map[string]interface{}{
			{
				"Effect": "Allow",
				"Principal": map[string]interface{}{
					"Federated": oidcProviderArn,
				},
				"Action": "sts:AssumeRoleWithWebIdentity",
				"Condition": map[string]interface{}{
					"StringEquals": map[string]interface{}{
						fmt.Sprintf("%s:aud", oidcIssuerHost): audience,
					},
				},
			},
		},
	}

	trustPolicyJSON, err := json.Marshal(trustPolicy)
	if err != nil {
		return false, fmt.Errorf("marshaling trust policy: %w", err)
	}

	_, err = iamCli.CreateRole(ctx, &iam.CreateRoleInput{
		RoleName:                 aws.String(roleName),
		AssumeRolePolicyDocument: aws.String(string(trustPolicyJSON)),
	})

	if err != nil {
		var entityExists *iamtypes.EntityAlreadyExistsException
		if errors.As(err, &entityExists) {
			fmt.Fprintf(esc.stdout, "IAM Role: %s (existing)\n", roleName)
			return false, nil
		}
		return false, fmt.Errorf("creating IAM role: %w", err)
	}

	fmt.Fprintf(esc.stdout, "IAM Role: %s (created)\n", roleName)
	return true, nil
}

func attachRolePolicy(
	ctx context.Context,
	esc *escCommand,
	iamCli iamClient,
	roleName, policyArn string,
) error {
	_, err := iamCli.AttachRolePolicy(ctx, &iam.AttachRolePolicyInput{
		RoleName:  aws.String(roleName),
		PolicyArn: aws.String(policyArn),
	})
	if err != nil {
		return fmt.Errorf("attaching policy to role: %w", err)
	}

	fmt.Fprintf(esc.stdout, "Policy Attached: %s\n", policyArn)
	return nil
}

func generateEnvironmentYAML(roleArn string) string {
	return fmt.Sprintf(`values:
  aws:
    login:
      fn::open::aws-login:
        oidc:
          duration: 1h
          roleArn: %s
          sessionName: pulumi-esc-session
  environmentVariables:
    AWS_ACCESS_KEY_ID: ${aws.login.accessKeyId}
    AWS_SECRET_ACCESS_KEY: ${aws.login.secretAccessKey}
    AWS_SESSION_TOKEN: ${aws.login.sessionToken}
`, roleArn)
}

func printEnvironmentYAML(esc *escCommand, roleArn string) {
	fmt.Fprintf(esc.stdout, "\nAdd the following to your ESC environment:\n\n")
	fmt.Fprint(esc.stdout, generateEnvironmentYAML(roleArn))
}

func createOrUpdateEnvironment(ctx context.Context, esc *escCommand, defaultOrg, envName, roleArn string) error {
	yaml := generateEnvironmentYAML(roleArn)

	ref := parseEnvRef(envName, defaultOrg)

	exists := true
	_, err := esc.client.EnvironmentExists(ctx, ref.orgName, ref.projectName, ref.envName)
	if err != nil {
		var errResp *apitype.ErrorResponse
		if errors.As(err, &errResp) && errResp.Code == http.StatusNotFound {
			exists = false
		} else {
			return fmt.Errorf("checking environment existence: %w", err)
		}
	}

	if !exists {
		if err := esc.client.CreateEnvironmentWithProject(ctx, ref.orgName, ref.projectName, ref.envName); err != nil {
			return fmt.Errorf("creating environment: %w", err)
		}
		fmt.Fprintf(esc.stdout, "\nEnvironment created: %s\n", ref.String())
	}

	diags, err := esc.client.UpdateEnvironmentWithProject(ctx, ref.orgName, ref.projectName, ref.envName, []byte(yaml), "")
	if err != nil {
		return fmt.Errorf("updating environment: %w", err)
	}

	if len(diags) != 0 {
		fmt.Fprintf(esc.stderr, "Warning: environment has diagnostics:\n")
		for _, d := range diags {
			fmt.Fprintf(esc.stderr, "  - %s\n", d.Summary)
		}
	}

	if exists {
		fmt.Fprintf(esc.stdout, "\nEnvironment updated: %s\n", ref.String())
	}

	return nil
}
