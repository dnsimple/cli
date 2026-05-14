package cli

import (
	_ "embed"
	"fmt"
	"io"
	"strings"

	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
)

//go:embed ai_context.md
var aiContext string

const aiAliasTargetAnnotation = "dnsimple.ai/alias-target"

func newAICmd() *cobra.Command {
	var full bool

	cmd := &cobra.Command{
		Use:   "ai",
		Short: "Show how to use dnsimple (useful for AI agents and LLMs).",
		Long: `Print a structured reference of all commands and common workflows. Formatted
for AI agents and LLMs to understand this CLI and perform automated tasks.

By default, per-command flag tables are omitted to keep token usage low; only
required flags are listed. Pass --full to include the complete flag tables.
Agents can also run any command with --help to discover its flags on demand.

Paste the output into an AI prompt, or load it as a bootstrapping step for an
agent.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			printAIContext(cmd.OutOrStdout(), cmd.Root(), full)
			return nil
		},
	}

	cmd.Flags().BoolVar(&full, "full", false, "Include the full flag table for each command")

	return cmd
}

func printAIContext(w io.Writer, root *cobra.Command, full bool) {
	flags := strings.TrimRight(renderGlobalFlags(root), "\n")
	tree := strings.TrimRight(renderCommandTree(root, "dnsimple", full), "\n")
	fmt.Fprintf(w, aiContext, flags, tree)
}

func renderGlobalFlags(root *cobra.Command) string {
	var b strings.Builder
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
		fmt.Fprintf(&b, "- `%s`: %s%s\n", name, f.Usage, def)
	})
	return b.String()
}

func renderCommandTree(cmd *cobra.Command, prefix string, full bool) string {
	var b strings.Builder
	if !full {
		fmt.Fprintln(&b, "Each command below lists its summary and any required flags. For the full flag table, run `dnsimple <command> --help`, or re-run `dnsimple ai --full` for the complete reference.")
		fmt.Fprintln(&b)
	}
	writeCommandTree(&b, cmd, prefix, full)
	return b.String()
}

func writeCommandTree(w io.Writer, cmd *cobra.Command, prefix string, full bool) {
	for _, sub := range cmd.Commands() {
		if sub.Hidden || sub.Name() == "help" || sub.Name() == "ai" || sub.Name() == "completion" {
			continue
		}

		fullCmd := prefix + " " + sub.Name()

		if target := sub.Annotations[aiAliasTargetAnnotation]; target != "" {
			writeAliasCommand(w, fullCmd, target)
			continue
		}

		if !sub.HasSubCommands() || sub.Runnable() {
			writeLeafCommand(w, sub, fullCmd, full)
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
			writeCommandTree(w, sub, fullCmd, full)
		}
	}
}

func writeAliasCommand(w io.Writer, fullCmd string, target string) {
	fmt.Fprintf(w, "### %s (alias for `%s`)\n\n", fullCmd, target)
}

func writeLeafCommand(w io.Writer, cmd *cobra.Command, fullCmd string, full bool) {
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

	if full {
		writeLeafFlagsFull(w, cmd)
	} else {
		writeLeafRequiredFlags(w, cmd)
	}
}

func writeLeafFlagsFull(w io.Writer, cmd *cobra.Command) {
	hasFlags := false
	cmd.LocalFlags().VisitAll(func(f *pflag.Flag) {
		if !f.Hidden {
			hasFlags = true
		}
	})
	if !hasFlags {
		return
	}

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

func writeLeafRequiredFlags(w io.Writer, cmd *cobra.Command) {
	var required []string
	cmd.LocalFlags().VisitAll(func(f *pflag.Flag) {
		if f.Hidden {
			return
		}
		if isFlagRequired(cmd, f.Name) {
			required = append(required, "`--"+f.Name+"`")
		}
	})
	if len(required) == 0 {
		return
	}
	fmt.Fprintf(w, "Required flags: %s\n\n", strings.Join(required, ", "))
}

func isFlagRequired(cmd *cobra.Command, name string) bool {
	annotations := cmd.LocalFlags().Lookup(name).Annotations
	if annotations == nil {
		return false
	}
	_, ok := annotations[cobra.BashCompOneRequiredFlag]
	return ok
}
