package cli

import (
	"fmt"
	"io"
	"strings"

	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
)

func newAICmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "ai",
		Short: "Print AI-friendly context about this CLI",
		Long: `Print a structured description of all commands, flags, and workflows
that an AI agent can use to understand and operate this CLI.

Usage with AI tools:

  Pipe directly into an AI prompt:
    dnsimple ai | pbcopy

  Use as a bootstrapping step in an AI agent:
    context=$(dnsimple ai)`,
		RunE: func(cmd *cobra.Command, args []string) error {
			printAIContext(cmd.OutOrStdout(), cmd.Root())
			return nil
		},
	}

	return cmd
}

func printAIContext(w io.Writer, root *cobra.Command) {
	fmt.Fprint(w, preamble)
	printGlobalFlags(w, root)
	fmt.Fprint(w, outputSection)
	fmt.Fprintln(w, "## Commands")
	fmt.Fprintln(w)
	printCommandTree(w, root, "dnsimple")
	fmt.Fprint(w, workflowExamples)
}

func printGlobalFlags(w io.Writer, root *cobra.Command) {
	fmt.Fprintln(w, "## Global Flags")
	fmt.Fprintln(w)
	fmt.Fprintln(w, "These flags work with any command:")
	fmt.Fprintln(w)
	root.PersistentFlags().VisitAll(func(f *pflag.Flag) {
		if f.Hidden {
			return
		}
		name := "--" + f.Name
		if f.Shorthand != "" {
			name = "-" + f.Shorthand + ", " + name
		}
		def := ""
		if f.DefValue != "" && f.DefValue != "false" && f.DefValue != "0" {
			def = fmt.Sprintf(" (default: %s)", f.DefValue)
		}
		fmt.Fprintf(w, "- `%s`: %s%s\n", name, f.Usage, def)
	})
	fmt.Fprintln(w)
}

func printCommandTree(w io.Writer, cmd *cobra.Command, prefix string) {
	for _, sub := range cmd.Commands() {
		if sub.Hidden || sub.Name() == "help" || sub.Name() == "ai" || sub.Name() == "completion" {
			continue
		}

		fullCmd := prefix + " " + sub.Name()

		if !sub.HasSubCommands() || sub.Runnable() {
			printLeafCommand(w, sub, fullCmd)
		}

		if sub.HasSubCommands() {
			if !sub.Runnable() {
				fmt.Fprintf(w, "### %s\n\n", fullCmd)
				if sub.Short != "" {
					fmt.Fprintf(w, "%s\n\n", sub.Short)
				}
				if len(sub.Aliases) > 0 {
					fmt.Fprintf(w, "Aliases: %s\n\n", strings.Join(sub.Aliases, ", "))
				}
			}
			printCommandTree(w, sub, fullCmd)
		}
	}
}

func printLeafCommand(w io.Writer, cmd *cobra.Command, fullCmd string) {
	usage := fullCmd
	if cmd.Use != "" {
		parts := strings.SplitN(cmd.Use, " ", 2)
		if len(parts) > 1 {
			usage = fullCmd + " " + parts[1]
		}
	}

	fmt.Fprintf(w, "#### `%s`\n\n", usage)
	if cmd.Short != "" {
		fmt.Fprintf(w, "%s\n\n", cmd.Short)
	}
	if len(cmd.Aliases) > 0 {
		fmt.Fprintf(w, "Aliases: %s\n\n", strings.Join(cmd.Aliases, ", "))
	}

	hasFlags := false
	cmd.LocalFlags().VisitAll(func(f *pflag.Flag) {
		if !f.Hidden {
			hasFlags = true
		}
	})

	if hasFlags {
		fmt.Fprintln(w, "Flags:")
		fmt.Fprintln(w)
		cmd.LocalFlags().VisitAll(func(f *pflag.Flag) {
			if f.Hidden {
				return
			}
			name := "--" + f.Name
			if f.Shorthand != "" {
				name = "-" + f.Shorthand + ", " + name
			}

			annotations := ""
			if isFlagRequired(cmd, f.Name) {
				annotations = " **(required)**"
			}
			def := ""
			if f.DefValue != "" && f.DefValue != "false" && f.DefValue != "0" && f.DefValue != "[]" {
				def = fmt.Sprintf(" (default: %s)", f.DefValue)
			}
			fmt.Fprintf(w, "- `%s`: %s%s%s\n", name, f.Usage, def, annotations)
		})
		fmt.Fprintln(w)
	}
}

func isFlagRequired(cmd *cobra.Command, name string) bool {
	annotations := cmd.LocalFlags().Lookup(name).Annotations
	if annotations == nil {
		return false
	}
	_, ok := annotations[cobra.BashCompOneRequiredFlag]
	return ok
}

const preamble = `# DNSimple CLI — AI Context

You are interacting with the DNSimple CLI (` + "`dnsimple`" + `), a command-line tool for managing domains, DNS records, certificates, and other DNSimple services via the DNSimple API v2.

## Authentication

The CLI requires a DNSimple API token. The user must authenticate before you can use the CLI.

If the CLI is not authenticated, ask the user to run ` + "`dnsimple auth login`" + ` themselves. Do NOT pass tokens via ` + "`--token`" + ` or environment variables — never handle API tokens directly.

You can check whether the CLI is already authenticated by running ` + "`dnsimple auth status`" + `.

To use the sandbox environment for testing, add the ` + "`--sandbox`" + ` flag.

Most commands require an account ID. The CLI resolves it automatically from the token, but you can override it with ` + "`-a <account-id>`" + ` or ` + "`--account <account-id>`" + `.

`

const outputSection = `## Output Formats

- **Table** (default): Human-readable tabular output
- **JSON**: Machine-readable output with ` + "`--json`" + `
- **Custom**: Go template formatting with ` + "`--format <template>`" + `

When scripting or parsing output programmatically, always use ` + "`--json`" + `.

`

const workflowExamples = `## Common Workflows

### List all DNS records for a zone

` + "```" + `
dnsimple zones records list example.com --all --json
` + "```" + `

### Create an A record

` + "```" + `
dnsimple zones records create example.com --type A --name www --content 1.2.3.4 --ttl 3600
` + "```" + `

### Update an existing record

` + "```" + `
dnsimple zones records update example.com 12345 --content 5.6.7.8
` + "```" + `

### Delete a record

` + "```" + `
dnsimple zones records delete example.com 12345
` + "```" + `

### Check domain availability and register

` + "```" + `
dnsimple registrar check example.com
dnsimple registrar prices example.com
dnsimple registrar register example.com --registrant-id 1234
` + "```" + `

### Manage Let's Encrypt certificates

` + "```" + `
dnsimple certificates letsencrypt purchase example.com
dnsimple certificates letsencrypt issue example.com 98765
` + "```" + `

### Apply a one-click service

` + "```" + `
dnsimple services list --json
dnsimple services apply github-pages example.com
` + "```" + `

## Tips

- Use ` + "`--all`" + ` on list commands to fetch every page of results automatically.
- Use ` + "`--json`" + ` when you need to parse output or chain commands.
- Use ` + "`-q`" + ` (quiet) to suppress confirmation messages in scripts.
- The ` + "`zones`" + ` and ` + "`zones records`" + ` commands are the primary way to manage DNS. The top-level ` + "`records`" + ` command is a shortcut alias.
- Domain names (e.g. example.com) are used as zone identifiers — you don't need zone IDs.
- Record IDs are numeric. Use ` + "`zones records list`" + ` with ` + "`--json`" + ` to find them.
`
