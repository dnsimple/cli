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

func newAICmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "ai",
		Short: "Show how to use dnsimple (useful for AI agents and LLMs).",
		Long: `Print a structured reference of all commands and common workflows. Formatted
for AI agents and LLMs to understand this CLI and perform automated tasks.

Paste the output into an AI prompt, or load it as a bootstrapping step for an
agent.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			printAIContext(cmd.OutOrStdout(), cmd.Root())
			return nil
		},
	}

	return cmd
}

func printAIContext(w io.Writer, root *cobra.Command) {
	flags := strings.TrimRight(renderGlobalFlags(root), "\n")
	tree := strings.TrimRight(renderCommandTree(root, "dnsimple"), "\n")
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

func renderCommandTree(cmd *cobra.Command, prefix string) string {
	var b strings.Builder
	writeCommandTree(&b, cmd, prefix)
	return b.String()
}

func writeCommandTree(w io.Writer, cmd *cobra.Command, prefix string) {
	for _, sub := range cmd.Commands() {
		if sub.Hidden || sub.Name() == "help" || sub.Name() == "ai" || sub.Name() == "completion" {
			continue
		}

		fullCmd := prefix + " " + sub.Name()

		if !sub.HasSubCommands() || sub.Runnable() {
			writeLeafCommand(w, sub, fullCmd)
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
			writeCommandTree(w, sub, fullCmd)
		}
	}
}

func writeLeafCommand(w io.Writer, cmd *cobra.Command, fullCmd string) {
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

