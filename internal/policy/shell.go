package policy

import (
	"regexp"
	"strings"
)

type Action string

const (
	ActionAllow Action = "allow"
	ActionAsk   Action = "ask"
	ActionDeny  Action = "deny"
)

type Decision struct {
	Action Action `json:"action"`
	Reason string `json:"reason"`
	RuleID string `json:"ruleId"`
}

type shellRule struct {
	id     string
	re     *regexp.Regexp
	reason string
}

var denyShellRules = []shellRule{
	newRule("sudo", `(^|[;&|()\s])sudo(\s|$)`, "sudo is blocked; use a non-root workspace-local alternative"),
	newRule("su-doas-pkexec", `(^|[;&|()\s])(su|doas|pkexec)(\s|$)`, "privilege escalation is blocked"),
	newRule("rm-root", `\brm\s+-[^\n;]*[rf][^\n;]*\s+(/|/\*)(\s|$)`, "destructive deletion outside the workspace is blocked"),
	newRule("mkfs-wipefs", `\b(mkfs|wipefs)\b`, "filesystem formatting and wiping commands are blocked"),
	newRule("disk-tools", `\b(fdisk|parted|cryptsetup)\b`, "disk partitioning and encryption commands are blocked"),
	newRule("dd-device", `\bdd\b[^\n;]*(of=/dev/|if=/dev/)`, "raw device reads/writes through dd are blocked"),
	newRule("shutdown", `\b(shutdown|reboot|poweroff|halt)\b`, "system shutdown commands are blocked"),
	newRule("systemctl", `\bsystemctl\b`, "system service mutation is blocked by default"),
	newRule("firewall", `\b(iptables|nft|ufw)\b`, "firewall mutation is blocked by default"),
	newRule("chmod-global", `\bchmod\s+-R\s+777\b`, "unsafe recursive chmod is blocked"),
	newRule("chown-recursive", `\bchown\s+-R\b`, "recursive ownership mutation is blocked"),
	newRule("git-clean", `\bgit\s+clean\s+-[a-zA-Z]*[fdx][a-zA-Z]*`, "git clean can destroy untracked work and is blocked"),
	newRule("git-force-push", `\bgit\s+push\b[^\n;]*--force`, "force-push is blocked"),
	newRule("curl-pipe-shell", `\b(curl|wget)\b[^\n;|]*\|\s*(sh|bash)\b`, "piping remote scripts into a shell is blocked"),
	newRule("secret-print", `\b(cat|sed|awk|grep|rg)\b[^\n;]*(\.env|id_rsa|id_ed25519|credentials|secrets?)\b`, "reading likely secret material is blocked"),
	newRule("remote-mutation", `\b(rsync|scp|sftp|ftp|ssh)\b[^\n;]*(--delete|rm\s|mv\s|put\s|mput\s)`, "remote mutation requires an explicit future policy gate"),
}

var allowReadOnlyRules = []shellRule{
	newRule("pwd", `^pwd$`, "read-only workspace inspection"),
	newRule("ls", `^ls(\s|$)`, "read-only listing"),
	newRule("find", `^find(\s|$)`, "read-only file discovery"),
	newRule("rg", `^rg(\s|$)`, "read-only search"),
	newRule("grep", `^grep(\s|$)`, "read-only search"),
	newRule("cat", `^cat(\s|$)`, "read-only file display"),
	newRule("sed-print", `^sed\s+-n(\s|$)`, "read-only file display"),
	newRule("git-status", `^git\s+status(\s|$)`, "read-only git status"),
	newRule("git-diff", `^git\s+diff(\s|$)`, "read-only git diff"),
	newRule("git-log", `^git\s+log(\s|$)`, "read-only git log"),
	newRule("go-test", `^go\s+test(\s|$)`, "project test command"),
	newRule("npm-test", `^npm\s+(run\s+)?test(\s|$)`, "project test command"),
	newRule("pnpm-test", `^pnpm\s+(run\s+)?test(\s|$)`, "project test command"),
	newRule("pytest", `^pytest(\s|$)`, "project test command"),
}

func newRule(id, pattern, reason string) shellRule {
	return shellRule{id: id, re: regexp.MustCompile(pattern), reason: reason}
}

func ClassifyShell(command string) Decision {
	cmd := normalizeCommand(command)
	if cmd == "" {
		return Decision{Action: ActionDeny, RuleID: "empty", Reason: "empty shell command"}
	}
	for _, rule := range denyShellRules {
		if rule.re.MatchString(cmd) {
			return Decision{Action: ActionDeny, RuleID: rule.id, Reason: rule.reason}
		}
	}
	if containsAmbiguousShellSyntax(cmd) {
		return Decision{Action: ActionAsk, RuleID: "compound-shell", Reason: "compound shell syntax needs a narrower policy decision"}
	}
	for _, rule := range allowReadOnlyRules {
		if rule.re.MatchString(cmd) {
			return Decision{Action: ActionAllow, RuleID: rule.id, Reason: rule.reason}
		}
	}
	return Decision{Action: ActionAsk, RuleID: "default", Reason: "no deterministic allow rule matched"}
}

func normalizeCommand(command string) string {
	cmd := strings.TrimSpace(command)
	cmd = strings.Join(strings.Fields(cmd), " ")
	return cmd
}

func containsAmbiguousShellSyntax(command string) bool {
	return strings.Contains(command, "&&") || strings.Contains(command, "||") || strings.Contains(command, ";") || strings.Contains(command, "`") || strings.Contains(command, "$(")
}
