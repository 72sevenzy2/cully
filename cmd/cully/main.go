// Command cully helps users guide coding agents and improve their workflow.
package main

import (
	"flag"
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/mcp-runtime/cully/internal/cully"
)

var version = "dev"

const help = `Cully — your companion for better work and everyday life.

Usage:
  cully agent setup [claude|codex|cursor|all] [--mcp-url URL] [--oauth]
  cully setup [--agent claude|codex|cursor] [--oauth] [--prepare]
  cully uninstall [--purge-data]           Stop local stack and remove managed agent setup
  cully uninstall [claude|codex|cursor|all] Remove one or all agent integrations
  cully status [directory]                Show session and agent status
  cully suggestions                       Review suggested improvements
  cully apply <n> [--dry-run] [--yes] [--cwd DIR]
  cully mcp add --url URL [--agent claude|codex|cursor] [--oauth]
  cully version

Shared personal and project memory is available through Cully's MCP tools.
Docs: https://docs.cully.net
`

func main() {
	if err := run(os.Args[1:]); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func run(args []string) error {
	if len(args) == 0 {
		fmt.Print(help)
		return nil
	}
	switch args[0] {
	case "help", "--help", "-h":
		fmt.Print(help)
	case "setup":
		return runSetup(args[1:])
	case "agent":
		if len(args) < 2 || args[1] != "setup" {
			return fmt.Errorf("usage: cully agent setup [claude|codex|cursor|all] [--mcp-url URL] [--oauth]")
		}
		return runAgentSetup(args[2:])
	case "uninstall":
		return runUninstall(args[1:])
	case "status":
		if len(args) > 2 {
			return fmt.Errorf("usage: cully status [directory]")
		}
		cwd, err := os.Getwd()
		if err != nil {
			return err
		}
		if len(args) == 2 {
			cwd = args[1]
		}
		cully.RunStatus(os.Stdout, cwd)
	case "suggestions":
		cully.RunList(os.Stdout)
	case "apply":
		n, yes, dryRun, cwd, err := parseApply(args[1:])
		if err != nil {
			return err
		}
		return cully.RunApply(n, cwd, yes, dryRun)
	case "version", "--version", "-v":
		fmt.Println("cully", version)
	case "mcp":
		if len(args) < 2 || args[1] != "add" {
			return fmt.Errorf("usage: cully mcp add --url URL [--agent claude|codex|cursor] [--oauth]")
		}
		fs := flag.NewFlagSet("mcp add", flag.ContinueOnError)
		agent := fs.String("agent", "", "coding agent; default: detect configured agents")
		endpoint := fs.String("url", "", "MCP URL of your deployment (required)")
		oauth := fs.Bool("oauth", false, "print sign-in instructions for an OAuth-protected server")
		if err := fs.Parse(args[2:]); err != nil {
			return err
		}
		if len(fs.Args()) != 0 {
			return fmt.Errorf("mcp add: unexpected arguments")
		}
		if *endpoint == "" {
			return fmt.Errorf("mcp add: --url is required")
		}
		return cully.AddMCP(os.Stdout, *agent, *endpoint, *oauth)
	case "_internal":
		return runInternal(args[1:])
	default:
		return fmt.Errorf("unknown command %q; run cully help", args[0])
	}
	return nil
}

func runSetup(args []string) error {
	target, oauth, prepare, err := parseSetup(args)
	if err != nil {
		return err
	}
	return runSelfHost(target, oauth, prepare)
}

func parseSetup(args []string) (target string, oauth, prepare bool, err error) {
	legacyTarget := ""
	if len(args) > 0 && !strings.HasPrefix(args[0], "-") {
		legacyTarget, args = args[0], args[1:]
	}
	fs := flag.NewFlagSet("setup", flag.ContinueOnError)
	selected := fs.String("agent", "", "coding agent to connect: claude, codex, or cursor")
	fs.BoolVar(&oauth, "oauth", false, "server uses OAuth; print sign-in instructions")
	fs.BoolVar(&prepare, "prepare", false, "download the self-hosted stack and create editable configuration without starting services")
	if err = fs.Parse(args); err != nil {
		return "", false, false, err
	}
	usage := fmt.Errorf("usage: cully setup [--agent claude|codex|cursor] [--oauth] [--prepare]")
	if len(fs.Args()) > 1 || (legacyTarget != "" && len(fs.Args()) != 0) {
		return "", false, false, usage
	}
	if legacyTarget != "" {
		target = legacyTarget
	}
	if len(fs.Args()) == 1 {
		target = fs.Args()[0]
	}
	agentFlagSet := false
	fs.Visit(func(option *flag.Flag) {
		if option.Name == "agent" {
			agentFlagSet = true
		}
	})
	if agentFlagSet {
		if *selected == "" || target != "" {
			return "", false, false, usage
		}
		target = *selected
	}
	if target == "all" {
		return "", false, false, fmt.Errorf("server setup accepts one agent at a time")
	}
	if target != "" && target != "claude" && target != "codex" && target != "cursor" {
		return "", false, false, fmt.Errorf("choose --agent claude, codex, or cursor")
	}
	return target, oauth, prepare, nil
}

func runAgentSetup(args []string) error {
	target := ""
	if len(args) > 0 && !strings.HasPrefix(args[0], "-") {
		target, args = args[0], args[1:]
	}
	fs := flag.NewFlagSet("agent setup", flag.ContinueOnError)
	endpoint := fs.String("mcp-url", "", "register this MCP URL while installing the agent integration")
	oauth := fs.Bool("oauth", false, "server uses OAuth; print sign-in instructions")
	if err := fs.Parse(args); err != nil {
		return err
	}
	if len(fs.Args()) > 1 || (target != "" && len(fs.Args()) != 0) {
		return fmt.Errorf("usage: cully agent setup [claude|codex|cursor|all] [--mcp-url URL] [--oauth]")
	}
	if target == "" && len(fs.Args()) == 1 {
		target = fs.Args()[0]
	}
	if *oauth && *endpoint == "" {
		return fmt.Errorf("--oauth requires --mcp-url for agent setup")
	}
	if *endpoint != "" {
		if target != "" {
			return cully.InstallWithMCP(*endpoint, *oauth, target)
		}
		return cully.InstallWithMCP(*endpoint, *oauth)
	}
	if target != "" {
		return cully.Install(target)
	}
	return cully.Install()
}

func parseApply(args []string) (n int, yes, dryRun bool, cwd string, err error) {
	if len(args) == 0 {
		err = fmt.Errorf("usage: cully apply <n> [--dry-run] [--yes] [--cwd DIR]")
		return
	}
	n, err = strconv.Atoi(args[0])
	if err != nil || n < 1 {
		err = fmt.Errorf("apply: suggestion number must be a positive integer")
		return
	}
	fs := flag.NewFlagSet("apply", flag.ContinueOnError)
	fs.BoolVar(&yes, "yes", false, "apply without a prompt")
	fs.BoolVar(&dryRun, "dry-run", false, "show plan only")
	fs.StringVar(&cwd, "cwd", "", "project directory")
	err = fs.Parse(args[1:])
	if err == nil && len(fs.Args()) != 0 {
		err = fmt.Errorf("apply: unexpected arguments")
	}
	return
}

// Internal entry points are installed by Cully; they are not user commands.
func runInternal(args []string) error {
	if len(args) == 0 {
		return fmt.Errorf("missing internal action")
	}
	switch args[0] {
	case "statusline":
		cully.RunStatusline(os.Stdin, os.Stdout)
	case "analyze":
		cully.RunAnalyze(os.Stdin)
	case "cleanup":
		cully.RunCleanup(os.Stdin)
	case "continuity":
		if len(args) != 3 {
			return fmt.Errorf("internal continuity requires agent and event")
		}
		cully.RunContinuityHook(args[1], args[2], os.Stdin, os.Stdout)
	case "daemon":
		cully.RunDaemon()
	case "worker":
		if len(args) != 4 {
			return fmt.Errorf("internal worker requires signals, session and directory")
		}
		cully.RunWorker(args[1], args[2], args[3])
	case "self-hosted-credentials":
		return runSelfHostCredentials(args[1:])
	default:
		return fmt.Errorf("unknown internal action %q", args[0])
	}
	return nil
}
