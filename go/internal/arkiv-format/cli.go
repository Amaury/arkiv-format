package arkivformat

import (
	"errors"
	"fmt"
	"os"
)

// Aliases for the CLI commands for convenience.
var (
	aliasesCreate  = map[string]bool{"c": true, "-c": true, "create": true, "--create": true}
	aliasesList    = map[string]bool{"l": true, "-l": true, "ls": true, "--ls": true}
	aliasesExtract = map[string]bool{"x": true, "-x": true, "extract": true, "--extract": true}
	aliasesHelp    = map[string]bool{"h": true, "-h": true, "help": true, "--help": true}
)

// RunCLI parses os.Args and dispatches to create, list, or extract commands.
// It enforces the environment variable ARKIV_PASS to provide the password.
func RunCLI(argv []string) error {
	if len(argv) < 2 || aliasesHelp[argv[1]] {
		printHelp()
		return nil
	}

	cmd := argv[1]
	switch {
	case aliasesCreate[cmd]:
		if len(argv) < 4 {
			return errors.New("usage: arkiv-format create ARCHIVE.arkiv PATH [PATH ...]")
		}
		archive := argv[2]
		inputs := argv[3:]
		pass := os.Getenv(EnvPass)
		if pass == "" {
			return fmt.Errorf("%s must be set", EnvPass)
		}
		w := NewArchiveWriter(archive, []byte(pass))
		defer w.Close()
		return w.Create(inputs)

	case aliasesList[cmd]:
		if len(argv) < 3 {
			return errors.New("usage: arkiv-format ls ARCHIVE.arkiv [PREFIX ...]")
		}
		archive := argv[2]
		prefixes := argv[3:]
		pass := os.Getenv(EnvPass)
		if pass == "" {
			return fmt.Errorf("%s must be set", EnvPass)
		}
		r := NewArchiveReader(archive, []byte(pass))
		defer r.Close()
		return r.List(prefixes)

	case aliasesExtract[cmd]:
		if len(argv) < 4 {
			return errors.New("usage: arkiv-format extract ARCHIVE.arkiv DEST [PREFIX ...]")
		}
		archive := argv[2]
		dest := argv[3]
		prefixes := argv[4:]
		pass := os.Getenv(EnvPass)
		if pass == "" {
			return fmt.Errorf("%s must be set", EnvPass)
		}
		r := NewArchiveReader(archive, []byte(pass))
		defer r.Close()
		return r.Extract(dest, prefixes)

	default:
		return fmt.Errorf("unknown command %q. Use --help", cmd)
	}
}

// printHelp prints CLI usage, environment, and examples.
func printHelp() {
	fmt.Println(`Arkiv — single binary compatible with the Arkiv format

USAGE:
  arkiv-format (c|-c|create|--create)   ARCHIVE.arkiv  PATH [PATH ...]
  arkiv-format (l|-l|ls|--ls)           ARCHIVE.arkiv  [PREFIX ...]
  arkiv-format (x|-x|extract|--extract) ARCHIVE.arkiv  DEST [PREFIX ...]
  arkiv-format (h|-h|help|--help)

ENV:
  ARKIV_PASS  Password for OpenSSL-compatible AES-256-CBC (PBKDF2 SHA-256, 10000 iter)

DEPENDENCIES:
  - github.com/klauspost/compress/zstd
  - golang.org/x/crypto/pbkdf2

EXAMPLES:
  export ARKIV_PASS=secret
  arkiv-format create backup.arkiv /etc /var/log/syslog
  arkiv-format ls     backup.arkiv
  arkiv-format ls     backup.arkiv /etc/ssh
  arkiv-format extract backup.arkiv /restore /etc/ssh`)
}

