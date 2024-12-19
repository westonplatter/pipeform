package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"slices"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/x/term"
	"github.com/magodo/pipeform/internal/log"
	"github.com/magodo/pipeform/internal/reader"
	"github.com/magodo/pipeform/internal/ui"
	"github.com/urfave/cli/v3"
)

type FlagSet struct {
	LogLevel string
	LogPath  string
	TeePath  string
}

var fset FlagSet

func main() {
	cmd := &cli.Command{
		Name:  "pipeform",
		Usage: "Terraform UI by running like: `terraform ... -json | pipeform`",
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:        "log-level",
				Usage:       "The log level",
				Sources:     cli.EnvVars("PF_LOG"),
				Value:       string(log.LevelDebug),
				Destination: &fset.LogLevel,
				Validator: func(input string) error {
					if !slices.Contains(log.PossibleLevels(), log.Level(strings.ToLower(input))) {
						return fmt.Errorf("invalid log level: %s", input)
					}
					return nil
				},
			},
			&cli.StringFlag{
				Name:        "log-path",
				Usage:       "The log path",
				Sources:     cli.EnvVars("PF_LOG_PATH"),
				Destination: &fset.LogPath,
			},
			&cli.StringFlag{
				Name:        "tee",
				Usage:       `Equivalent to "terraform ... -json | tee <value> | pipeform"`,
				Sources:     cli.EnvVars("PF_TEE"),
				Destination: &fset.TeePath,
			},
		},
		Before: func(ctx context.Context, c *cli.Command) (context.Context, error) {
			// If this program starts in standalone, its stdin is the same as the terminal.
			// bubbletea will change the terminal into raw mode and read ansi events from it,
			// which conflicts with the stdin reading for terraform JSON streams.
			// In this case, user's input (e.g. ctrl-c keypress) will most likely be accidently read by
			// the stream reader, instead of the ansi read loop (by bubbletea), causing a lost of event.
			if term.IsTerminal(os.Stdin.Fd()) {
				return ctx, errors.New("Must be followed by a pipe")
			}
			return ctx, nil
		},
		Action: func(context.Context, *cli.Command) error {
			logger, err := log.NewLogger(log.Level(fset.LogLevel), fset.LogPath)
			if err != nil {
				fmt.Println(err)
				os.Exit(1)
			}
			defer logger.Close()

			teeWriter := io.Discard
			if path := fset.TeePath; path != "" {
				f, err := os.OpenFile(path, os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0600)
				if err != nil {
					return fmt.Errorf("open for tee: %v", err)
				}
				teeWriter = f
				defer f.Close()
			}

			reader := reader.NewReader(os.Stdin, teeWriter)
			m := ui.NewRuntimeModel(logger, reader)
			tm, err := tea.NewProgram(m, tea.WithInputTTY(), tea.WithAltScreen()).Run()
			if err != nil {
				return fmt.Errorf("Error running program: %v\n", err)
			}

			m = tm.(ui.UIModel)

			if !m.IsEOF() {
				fmt.Fprintln(os.Stderr, "Interrupted!")
				printDiagErrs(m)
				os.Exit(1)
			}

			printDiagErrs(m)

			return nil
		},
	}

	if err := cmd.Run(context.Background(), os.Args); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func printDiagErrs(m ui.UIModel) {
	if diags := m.Diags(); len(diags.Errs) != 0 {
		for _, diag := range diags.Errs {
			if b, err := json.MarshalIndent(diag, "", "  "); err == nil {
				fmt.Fprintln(os.Stderr, string(b))
			}
		}
	}
}
