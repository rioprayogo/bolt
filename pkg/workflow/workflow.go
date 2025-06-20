package workflow

import (
	"bold/pkg/compiler"
	"bold/pkg/config"
	"bold/pkg/engine"
	"bold/pkg/errors"
	"bold/pkg/logger"
	"bold/pkg/parser"
	"bufio"
	"fmt"
	"os"
	"strings"
)

// Run menjalankan alur kerja standar: Parse -> Compile -> Execute.
func Run(manifestFile string, action string) error {
	cfg, err := config.LoadConfig("")
	if err != nil {
		return fmt.Errorf("failed to load configuration: %w", err)
	}

	if err := logger.Init(cfg.Logging.Level, cfg.Logging.Format, cfg.Logging.Output); err != nil {
		return fmt.Errorf("failed to initialize logger: %w", err)
	}

	defer logger.LogOperation("workflow", logger.Fields{
		"manifest_file": manifestFile,
		"action":        action,
	})()

	logger.Info("Starting manifest parsing", logger.Fields{
		"manifest_file": manifestFile,
	})

	manifest, err := parser.ParseManifest(manifestFile)
	if err != nil {
		logger.LogError(err, "manifest parsing", logger.Fields{
			"manifest_file": manifestFile,
		})
		return &errors.CompilationError{
			Message: "failed to parse manifest",
			Details: err.Error(),
		}
	}

	logger.Info("Manifest parsed successfully", logger.Fields{
		"service_name": manifest.Metadata.Name,
		"owner":        manifest.Metadata.Owner,
		"providers":    len(manifest.Providers),
	})

	compileDir := "./bolt_build"
	logger.Info("Starting compilation", logger.Fields{
		"compile_dir": compileDir,
	})

	err = compiler.CompileToTofu(manifest, compileDir)
	if err != nil {
		logger.LogError(err, "compilation", logger.Fields{
			"compile_dir": compileDir,
		})
		return &errors.CompilationError{
			Message: "failed to compile to OpenTofu",
			Details: err.Error(),
		}
	}

	logger.Info("Compilation completed successfully", logger.Fields{
		"compile_dir": compileDir,
	})

	logger.Info("Starting OpenTofu execution", logger.Fields{
		"action": action,
	})

	tofuEngine := &engine.OpenTofuEngine{WorkDir: compileDir}

	if err := tofuEngine.Init(); err != nil {
		logger.LogError(err, "OpenTofu initialization", logger.Fields{
			"work_dir": compileDir,
		})
		return &errors.ExecutionError{
			Command:  "tofu init",
			Output:   err.Error(),
			ExitCode: 1,
		}
	}

	if action == "plan" {
		if err := tofuEngine.Plan(); err != nil {
			logger.LogError(err, "OpenTofu plan", logger.Fields{
				"work_dir": compileDir,
			})
			return &errors.ExecutionError{
				Command:  "tofu plan",
				Output:   err.Error(),
				ExitCode: 1,
			}
		}
	} else if action == "apply" {
		if cfg.Security.RequireConfirmation {
			if !confirmAction("apply") {
				logger.Info("Apply cancelled by user", logger.Fields{})
				return nil
			}
		}

		if err := tofuEngine.Plan(); err != nil {
			logger.LogError(err, "OpenTofu plan before apply", logger.Fields{
				"work_dir": compileDir,
			})
			return &errors.ExecutionError{
				Command:  "tofu plan",
				Output:   err.Error(),
				ExitCode: 1,
			}
		}

		if err := tofuEngine.Apply(); err != nil {
			logger.LogError(err, "OpenTofu apply", logger.Fields{
				"work_dir": compileDir,
			})
			return &errors.ExecutionError{
				Command:  "tofu apply",
				Output:   err.Error(),
				ExitCode: 1,
			}
		}
	} else if action == "destroy" {
		if cfg.Security.RequireConfirmation {
			if !confirmAction("destroy") {
				logger.Info("Destroy cancelled by user", logger.Fields{})
				return nil
			}
		}

		if err := tofuEngine.PlanDestroy(); err != nil {
			logger.LogError(err, "OpenTofu plan destroy", logger.Fields{
				"work_dir": compileDir,
			})
			return &errors.ExecutionError{
				Command:  "tofu plan -destroy",
				Output:   err.Error(),
				ExitCode: 1,
			}
		}

		if err := tofuEngine.Destroy(); err != nil {
			logger.LogError(err, "OpenTofu destroy", logger.Fields{
				"work_dir": compileDir,
			})
			return &errors.ExecutionError{
				Command:  "tofu destroy",
				Output:   err.Error(),
				ExitCode: 1,
			}
		}
	}

	logger.Info("Workflow completed successfully", logger.Fields{
		"action": action,
	})

	return nil
}

func confirmAction(action string) bool {
	fmt.Printf("\n⚠️  Are you sure you want to %s the infrastructure? (yes/no): ", action)
	reader := bufio.NewReader(os.Stdin)
	response, _ := reader.ReadString('\n')
	response = strings.TrimSpace(strings.ToLower(response))
	return response == "yes" || response == "y"
}
