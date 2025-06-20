package engine

import (
	"bold/pkg/errors"
	"bold/pkg/logger"
	"fmt"
	"os"
	"os/exec"
	"strings"
)

// Engine adalah interface universal untuk semua engine eksekusi (Tofu, Ansible, dll).
type Engine interface {
	Init() error
	Plan() error
	Apply() error
	PlanDestroy() error
	Destroy() error
}

// OpenTofuEngine adalah implementasi Engine untuk OpenTofu.
type OpenTofuEngine struct {
	WorkDir string
}

func (t *OpenTofuEngine) runCommand(args ...string) error {
	cmd := exec.Command("tofu", args...)
	cmd.Dir = t.WorkDir
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	commandStr := fmt.Sprintf("tofu %s", strings.Join(args, " "))
	logger.Info("Executing OpenTofu command", logger.Fields{
		"command":  commandStr,
		"work_dir": t.WorkDir,
	})

	fmt.Printf("\n==> Menjalankan: %s (di direktori: %s)\n", commandStr, t.WorkDir)

	if err := cmd.Run(); err != nil {
		logger.LogError(err, "OpenTofu command execution", logger.Fields{
			"command":  commandStr,
			"work_dir": t.WorkDir,
		})
		return &errors.ExecutionError{
			Command:  commandStr,
			Output:   err.Error(),
			ExitCode: cmd.ProcessState.ExitCode(),
		}
	}

	logger.Info("OpenTofu command completed successfully", logger.Fields{
		"command": commandStr,
	})

	return nil
}

func (t *OpenTofuEngine) Init() error {
	logger.Info("Initializing OpenTofu workspace", logger.Fields{
		"work_dir": t.WorkDir,
	})

	return t.runCommand("init", "-upgrade")
}

func (t *OpenTofuEngine) Plan() error {
	logger.Info("Creating OpenTofu plan", logger.Fields{
		"work_dir": t.WorkDir,
	})

	return t.runCommand("plan", "-out=tfplan")
}

func (t *OpenTofuEngine) Apply() error {
	logger.Info("Applying OpenTofu plan", logger.Fields{
		"work_dir": t.WorkDir,
	})

	return t.runCommand("apply", "tfplan")
}

func (t *OpenTofuEngine) PlanDestroy() error {
	logger.Info("Creating OpenTofu destroy plan", logger.Fields{
		"work_dir": t.WorkDir,
	})

	return t.runCommand("plan", "-destroy", "-out=tfplan")
}

func (t *OpenTofuEngine) Destroy() error {
	logger.Info("Destroying infrastructure", logger.Fields{
		"work_dir": t.WorkDir,
	})

	return t.runCommand("apply", "tfplan")
}
