package main

import (
	"encoding/json"
	"log"
	"log/slog"
	"os"
	"os/exec"
	"os/signal"
	"path/filepath"
	"sync"
	"syscall"

	"github.com/joho/godotenv"
)

var (
	runningCmds   []*exec.Cmd
	runningCmdsMu sync.Mutex
)

func registerCmd(cmd *exec.Cmd) {
	runningCmdsMu.Lock()
	runningCmds = append(runningCmds, cmd)
	runningCmdsMu.Unlock()
}

func unregisterCmd(cmd *exec.Cmd) {
	runningCmdsMu.Lock()
	for i, c := range runningCmds {
		if c == cmd {
			runningCmds = append(runningCmds[:i], runningCmds[i+1:]...)
			break
		}
	}
	runningCmdsMu.Unlock()
}

func killAllCmds() {
	runningCmdsMu.Lock()
	defer runningCmdsMu.Unlock()
	for _, cmd := range runningCmds {
		if cmd.Process != nil {
			// Kill the entire process group to catch grandchildren (e.g. the binary spawned by `go run`)
			_ = syscall.Kill(-cmd.Process.Pid, syscall.SIGKILL)
		}
	}
}

type Command string

const (
	CmdBuild   Command = "build"
	CmdStart   Command = "start"
	CmdAdd     Command = "add"
	CmdMigrate Command = "migrate"
)

type Service struct {
	Dir        string `json:"dir"`
	Entrypoint string `json:"entrypoint"`
	Type       string `json:"type"` // "go" or "bun"
}

// rootDir returns the project root (the parent of the cli/ directory).
// The CLI module lives in cli/, so when invoked with `go run main.go` from
// within cli/, os.Getwd() returns …/Rivon/cli — we go one level up to get
// the actual project root that holds services.json, Server/, Engine/, etc.
func rootDir() string {
	cwd, err := os.Getwd()
	if err != nil {
		log.Fatalf("UNABLE TO DETERMINE WORKING DIRECTORY :: %v", err)
	}
	return filepath.Dir(cwd)
}

func servicesPath() string {
	return filepath.Join(rootDir(), "services.json")
}

func loadServices() (map[string]Service, error) {
	data, err := os.ReadFile(servicesPath())
	if err != nil {
		return nil, err
	}
	var services map[string]Service
	err = json.Unmarshal(data, &services)
	return services, err
}

func saveServices(services map[string]Service) error {
	data, err := json.MarshalIndent(services, "", "  ")
	if err != nil {
		log.Fatalf("ERROR IN JSON PARSING :: %s", err.Error())
	}
	return os.WriteFile(servicesPath(), data, 0644)
}

func runCmdInDir(dir, command string, args ...string) error {
	cmd := exec.Command(command, args...)
	cmd.Dir = dir
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Stdin = os.Stdin
	// Put child in its own process group so killing -pgid reaches grandchildren (e.g. binary spawned by `go run`)
	cmd.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}
	if err := cmd.Start(); err != nil {
		return err
	}
	registerCmd(cmd)
	err := cmd.Wait()
	unregisterCmd(cmd)
	return err
}

func migrate(databaseURL string, direction string) {
	root := rootDir()
	migrationsPath := filepath.Join(root, "Server", "internals", "database", "migrations")
	err := runCmdInDir(root, "migrate",
		"-path", migrationsPath,
		"-database", databaseURL,
		direction,
	)
	if err != nil {
		log.Fatalf("UNABLE TO MIGRATE %s :: %v", direction, err)
	}
}

func add(apps ...string) {
	var template = `package main 
	
func main() {}`

	services, err := loadServices()
	if err != nil {
		log.Fatal("UNABLE TO LOAD services.json ", err.Error())
	}

	root := rootDir()

	for _, app := range apps {
		if _, exists := services[app]; exists {
			log.Printf("Service %s already exists", app)
			continue
		}

		relDir := "./Server"
		relEntrypoint := filepath.Join("./cmd", app)
		absPath := filepath.Join(root, "Server", "cmd", app)

		if err := os.MkdirAll(absPath, 0755); err != nil {
			log.Fatalln("UNABLE TO CREATE SERVICE DIR ", err.Error())
		}
		fileLocation := filepath.Join(absPath, "main.go")
		f, err := os.Create(fileLocation)
		if err != nil {
			log.Fatalln("UNABLE TO CREATE main.go IN THAT LOCATION", err.Error())
		}
		_, err = f.WriteString(template)
		if err != nil {
			f.Close()
			log.Fatalf("UNABLE TO UPDATE main.go :: %s", err.Error())
		}
		services[app] = Service{Dir: relDir, Entrypoint: relEntrypoint, Type: "go"}
		f.Close()
	}
	saveServices(services)
}

func start(target ...string) {
	var wg sync.WaitGroup
	root := rootDir()
	services, err := loadServices()
	if err != nil {
		log.Fatalln("UNABLE TO LOAD SERVICES FROM services.json FOR STARTING WORKSPACE")
	}
	for _, name := range target {
		serviceConfig, ok := services[name]
		if !ok {
			log.Fatalln("UNKNOWN SERVICE, PLEASE PROVIDE A VALID SERVICE NAME:", name)
		}
		wg.Add(1)
		go func(name string, svc Service) {
			defer wg.Done()
			log.Printf("STARTING APPLICATION... :: %s", name)
			dir := filepath.Join(root, svc.Dir)
			var runErr error
			if svc.Type == "go" {
				runErr = runCmdInDir(dir, "go", "run", svc.Entrypoint)
			} else if svc.Type == "bun" {
				runErr = runCmdInDir(dir, "bun", "run", "dev")
			}
			if runErr != nil {
				log.Printf("UNABLE TO START THE APPLICATION :: %s :: %s", name, runErr.Error())
				killAllCmds()
				os.Exit(1)
			}
			log.Printf("APPLICATION FINISHED :: %s", name)
		}(name, serviceConfig)
	}
	wg.Wait()
}

func build(target ...string) {
	var wg sync.WaitGroup
	root := rootDir()
	services, err := loadServices()
	if err != nil {
		log.Fatalln("UNABLE TO LOAD SERVICES FROM services.json FOR BUILDING WORKSPACE")
	}

	for _, name := range target {
		serviceConfig, ok := services[name]
		if !ok {
			log.Fatalln("UNKNOWN SERVICE, PLEASE PROVIDE A VALID SERVICE NAME:", name)
		}
		wg.Add(1)
		go func(name string, svc Service) {
			defer wg.Done()
			log.Printf("BUILDING APPLICATION... :: %s", name)
			dir := filepath.Join(root, svc.Dir)
			outputPath := filepath.Join(root, "bin", name)
			var buildErr error
			if svc.Type == "go" {
				buildErr = runCmdInDir(dir, "go", "build", "-o", outputPath, svc.Entrypoint)
			} else if svc.Type == "bun" {
				buildErr = runCmdInDir(dir, "bun", "run", "build")
			}
			if buildErr != nil {
				log.Printf("UNABLE TO BUILD THIS SERVICE :: %s :: %s", name, buildErr.Error())
				killAllCmds()
				os.Exit(1)
			}
		}(name, serviceConfig)
	}
	wg.Wait()
}

func main() {
	slog.Info("---------CLI TOOL RUNNING FOR RIVON PROJECT--------")

	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, os.Interrupt, syscall.SIGTERM, syscall.SIGHUP)
	go func() {
		<-sigCh
		log.Println("SHUTTING DOWN ALL SERVICES...")
		killAllCmds()
		os.Exit(0)
	}()

	if len(os.Args) < 2 {
		log.Fatalf("usage: cli <start | build | add | migrate> [services...]\n  migrate expects: migrate <up|down>")
	}
	command := Command(os.Args[1])
	args := os.Args[2:]

	services, err := loadServices()
	if err != nil && command != CmdMigrate {
		log.Fatalln("ERROR IN LOADING SERVICES FROM services.json ", err.Error())
	}

	// If no specific services provided, run all (except for add and migrate)
	if len(args) == 0 && command != CmdMigrate {
		if command == CmdAdd {
			log.Fatalln("PLEASE PROVIDE A SERVICE NAME TO BE CREATED")
		}
		for service := range services {
			args = append(args, service)
		}
	}

	switch command {
	case CmdStart:
		start(args...)
	case CmdBuild:
		build(args...)
	case CmdAdd:
		add(args...)
	case CmdMigrate:
		if len(args) != 1 || (args[0] != "up" && args[0] != "down") {
			log.Fatalf("migrate expects exactly one argument: up | down")
		}
		envFile := filepath.Join(rootDir(), "Server", ".env")
		if loadErr := godotenv.Load(envFile); loadErr != nil {
			log.Printf("No Server/.env found, falling back to system environment variables")
		}
		pgURL := os.Getenv("DATABASE_POSTGRES_URL")
		if pgURL == "" {
			log.Fatalf("DATABASE_POSTGRES_URL is not set")
		}
		migrate(pgURL, args[0])
	default:
		log.Fatalf("unknown command: %s", command)
	}
}
