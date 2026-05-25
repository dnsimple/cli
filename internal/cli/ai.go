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
	fmt.Fprintf(&b, "All commands below are invoked as `%s <command>`.\n", prefix)
	if !full {
		fmt.Fprintf(&b, "Each command below lists its summary and any required flags. For the full flag table, run `%s <command> --help`, or re-run `%s ai --full` for the complete reference.\n", prefix, prefix)
	} else {
		fmt.Fprintln(&b, "Full command flag tables are included for commands that define local flags.")
	}
	fmt.Fprintln(&b)
	writeCommandTree(&b, cmd, "", prefix, full)
	return b.String()
}

func writeCommandTree(w io.Writer, cmd *cobra.Command, parentPath string, rootName string, full bool) {
	for _, sub := range cmd.Commands() {
		if skipAICommand(sub) {
			continue
		}

		commandPath := joinCommandPath(parentPath, sub.Name())

		if target := sub.Annotations[aiAliasTargetAnnotation]; target != "" {
			writeAliasCommand(w, commandPath, trimCommandPrefix(target, rootName))
			continue
		}

		if sub.HasSubCommands() && !sub.Runnable() {
			writeCommandGroup(w, sub, commandPath, rootName, full)
			continue
		}

		writeStandaloneCommand(w, sub, commandPath, full)
	}
}

func writeCommandGroup(w io.Writer, cmd *cobra.Command, commandPath string, rootName string, full bool) {
	fmt.Fprintf(w, "### %s\n\n", commandPath)
	if cmd.Short != "" {
		fmt.Fprintf(w, "%s\n\n", cmd.Short)
	}
	if len(cmd.Aliases) > 0 {
		fmt.Fprintf(w, "Aliases: %s\n\n", strings.Join(cmd.Aliases, ", "))
	}

	for _, sub := range cmd.Commands() {
		if skipAICommand(sub) {
			continue
		}
		if sub.HasSubCommands() && !sub.Runnable() {
			continue
		}
		writeLeafCommand(w, sub, full)
	}
	fmt.Fprintln(w)

	for _, sub := range cmd.Commands() {
		if skipAICommand(sub) {
			continue
		}
		subPath := joinCommandPath(commandPath, sub.Name())
		if target := sub.Annotations[aiAliasTargetAnnotation]; target != "" {
			writeAliasCommand(w, subPath, trimCommandPrefix(target, rootName))
			continue
		}
		if sub.HasSubCommands() && !sub.Runnable() {
			writeCommandGroup(w, sub, subPath, rootName, full)
		}
	}
}

func writeStandaloneCommand(w io.Writer, cmd *cobra.Command, commandPath string, full bool) {
	fmt.Fprintf(w, "### %s\n\n", commandPath)
	if cmd.Short != "" {
		fmt.Fprintf(w, "%s\n\n", cmd.Short)
	}
	if len(cmd.Aliases) > 0 {
		fmt.Fprintf(w, "Aliases: %s\n\n", strings.Join(cmd.Aliases, ", "))
	}
	if full {
		writeLeafFlagsFull(w, cmd, "")
	} else {
		writeLeafRequiredFlags(w, cmd)
	}
}

func writeAliasCommand(w io.Writer, commandPath string, target string) {
	fmt.Fprintf(w, "### %s (alias for `%s`)\n\n", commandPath, target)
}

func writeLeafCommand(w io.Writer, cmd *cobra.Command, full bool) {
	usage := leafUsage(cmd)
	fmt.Fprintf(w, "- `%s`", usage)
	if cmd.Short != "" {
		fmt.Fprintf(w, ": %s", cmd.Short)
	}
	if len(cmd.Aliases) > 0 {
		fmt.Fprintf(w, " (aliases: %s)", strings.Join(cmd.Aliases, ", "))
	}

	if full {
		fmt.Fprintln(w)
		writeLeafFlagsFull(w, cmd, "  ")
		return
	}

	required := requiredFlagNames(cmd)
	if len(required) > 0 {
		fmt.Fprintf(w, ". Required flags: %s", strings.Join(required, ", "))
	}
	fmt.Fprintln(w)
}

func leafUsage(cmd *cobra.Command) string {
	usage := cmd.Name()
	if cmd.Use != "" {
		parts := strings.SplitN(cmd.Use, " ", 2)
		if len(parts) > 1 {
			usage = cmd.Name() + " " + parts[1]
		}
	}
	return usage
}

func joinCommandPath(parentPath string, name string) string {
	if parentPath == "" {
		return name
	}
	return parentPath + " " + name
}

func trimCommandPrefix(commandPath string, rootName string) string {
	return strings.TrimPrefix(commandPath, rootName+" ")
}

func skipAICommand(cmd *cobra.Command) bool {
	return cmd.Hidden || cmd.Name() == "help" || cmd.Name() == "ai" || cmd.Name() == "completion"
}

func requiredFlagNames(cmd *cobra.Command) []string {
	var required []string
	cmd.LocalFlags().VisitAll(func(f *pflag.Flag) {
		if f.Hidden {
			return
		}
		if isFlagRequired(cmd, f.Name) {
			required = append(required, "`--"+f.Name+"`")
		}
	})
	return required
}

func writeLeafFlagsFull(w io.Writer, cmd *cobra.Command, indent string) {
	hasFlags := false
	cmd.LocalFlags().VisitAll(func(f *pflag.Flag) {
		if !f.Hidden {
			hasFlags = true
		}
	})
	if !hasFlags {
		return
	}

	fmt.Fprintf(w, "%sFlags:\n", indent)
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
		fmt.Fprintf(w, "%s- `%s`: %s%s%s\n", indent, name, f.Usage, def, annotations)
	})
}

func writeLeafRequiredFlags(w io.Writer, cmd *cobra.Command) {
	required := requiredFlagNames(cmd)
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
