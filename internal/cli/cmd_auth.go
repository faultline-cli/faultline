package cli

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"github.com/spf13/cobra"
	"golang.org/x/term"

	"faultline/internal/teams"
)

func newAuthCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "auth",
		Short: "Manage Faultline Teams authentication",
		Long: joinLines(
			"Authenticate with Faultline Teams to enable sync and organizational features.",
			"",
			"Use 'faultline auth login' to sign in interactively.",
			"Use 'faultline auth token set' to configure a token directly in CI environments.",
		),
	}
	cmd.AddCommand(newAuthLoginCommand())
	cmd.AddCommand(newAuthLogoutCommand())
	cmd.AddCommand(newAuthStatusCommand())
	cmd.AddCommand(newAuthTokenCommand())
	return cmd
}

func newAuthLoginCommand() *cobra.Command {
	var (
		apiURL    string
		teamSlug  string
		email     string
		tokenName string
	)

	cmd := &cobra.Command{
		Use:   "login",
		Short: "Sign in to Faultline Teams",
		Long: joinLines(
			"Authenticate with Faultline Teams using your email and password.",
			"",
			"A persistent API token (ft_…) is created and stored at",
			"~/.config/faultline/credentials. This token is used automatically by",
			"'faultline sync' and other Teams commands.",
		),
		Example: joinLines(
			"  faultline auth login",
			"  faultline auth login --team my-org",
		),
		Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			out := cmd.OutOrStdout()
			scanner := bufio.NewScanner(os.Stdin)

			if email == "" {
				fmt.Fprint(out, "Email: ")
				if !scanner.Scan() {
					return fmt.Errorf("read email: unexpected EOF")
				}
				email = strings.TrimSpace(scanner.Text())
			}

			fmt.Fprint(out, "Password: ")
			pwBytes, err := term.ReadPassword(int(os.Stdin.Fd()))
			fmt.Fprintln(out)
			if err != nil {
				return fmt.Errorf("read password: %w", err)
			}
			password := string(pwBytes)

			if teamSlug == "" {
				fmt.Fprint(out, "Team slug: ")
				if !scanner.Scan() {
					return fmt.Errorf("read team slug: unexpected EOF")
				}
				teamSlug = strings.TrimSpace(scanner.Text())
			}

			if email == "" || teamSlug == "" {
				return fmt.Errorf("email and team slug are required")
			}

			resolvedURL := firstNonEmpty(apiURL, os.Getenv("FAULTLINE_API_URL"), teams.DefaultAPIURL)
			client := teams.NewClient(resolvedURL)

			fmt.Fprintln(out, "Authenticating…")
			token, userEmail, err := client.Login(email, password, teamSlug, tokenName)
			if err != nil {
				return err
			}

			creds := &teams.Credentials{
				APIURL:   resolvedURL,
				Token:    token,
				TeamSlug: teamSlug,
				Email:    userEmail,
			}
			if err := teams.SaveCredentials(creds); err != nil {
				return fmt.Errorf("save credentials: %w", err)
			}
			fmt.Fprintf(out, "Logged in as %s (team: %s)\n", userEmail, teamSlug)
			return nil
		},
	}

	cmd.Flags().StringVar(&apiURL, "api-url", "", "Teams API base URL (overrides $FAULTLINE_API_URL)")
	cmd.Flags().StringVar(&teamSlug, "team", "", "team slug")
	cmd.Flags().StringVar(&email, "email", "", "email address")
	cmd.Flags().StringVar(&tokenName, "name", "faultline-cli", "name for the generated API token")
	return cmd
}

func newAuthLogoutCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "logout",
		Short: "Remove stored Faultline Teams credentials",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := teams.ClearCredentials(); err != nil {
				return err
			}
			fmt.Fprintln(cmd.OutOrStdout(), "Logged out.")
			return nil
		},
	}
}

func newAuthStatusCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "status",
		Short: "Show current Faultline Teams authentication status",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			out := cmd.OutOrStdout()
			creds, err := teams.LoadCredentials()
			if err != nil {
				return err
			}
			if creds == nil || creds.Token == "" {
				fmt.Fprintln(out, "Not logged in. Run 'faultline auth login' to authenticate.")
				return nil
			}
			client := teams.NewClient(creds.APIURL)
			email, err := client.VerifyToken(creds.Token, creds.TeamSlug)
			if err != nil {
				fmt.Fprintf(out, "Token invalid or expired: %v\n", err)
				fmt.Fprintln(out, "Run 'faultline auth login' to re-authenticate.")
				return nil
			}
			fmt.Fprintf(out, "Logged in as %s (team: %s, api: %s)\n", email, creds.TeamSlug, creds.APIURL)
			return nil
		},
	}
}

func newAuthTokenCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "token",
		Short: "Manage the stored API token",
	}
	cmd.AddCommand(newAuthTokenSetCommand())
	return cmd
}

func newAuthTokenSetCommand() *cobra.Command {
	var (
		apiURL   string
		teamSlug string
	)

	cmd := &cobra.Command{
		Use:   "set <token>",
		Short: "Store an API token for use by faultline commands",
		Long: joinLines(
			"Store an API token (ft_…) directly. Useful in CI environments where",
			"interactive login is not possible.",
			"",
			"Set FAULTLINE_TOKEN in the environment as an alternative to storing the",
			"token on disk.",
		),
		Example: joinLines(
			"  faultline auth token set ft_abc123 --team my-org",
			"  FAULTLINE_TOKEN=ft_abc123 faultline sync ...",
		),
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			token := args[0]
			if !strings.HasPrefix(token, "ft_") {
				return fmt.Errorf("invalid token: must start with 'ft_'")
			}

			resolvedURL := firstNonEmpty(apiURL, os.Getenv("FAULTLINE_API_URL"), teams.DefaultAPIURL)
			creds := &teams.Credentials{
				APIURL:   resolvedURL,
				Token:    token,
				TeamSlug: teamSlug,
			}
			if err := teams.SaveCredentials(creds); err != nil {
				return fmt.Errorf("save credentials: %w", err)
			}
			fmt.Fprintln(cmd.OutOrStdout(), "Token stored.")
			return nil
		},
	}

	cmd.Flags().StringVar(&apiURL, "api-url", "", "Teams API base URL (overrides $FAULTLINE_API_URL)")
	cmd.Flags().StringVar(&teamSlug, "team", "", "team slug")
	return cmd
}
