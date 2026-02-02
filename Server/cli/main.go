package main

import (
	"encoding/json"
	"log"
	"log/slog"
	"os"
	"os/exec"
	"path/filepath"
	"sync"

	"github.com/raiashpanda007/rivon/internals/config"
)

type Command string

const (
	CmdBuild   Command = "build"
	CmdStart   Command = "start"
	CmdAdd     Command = "add"
	CmdMigrate Command = "migrate"
)

func loadServices() (map[string]string, error) {
	data, err := os.ReadFile("services.json")
	if err != nil {
		return nil, err
	}
	var services map[string]string
	err = json.Unmarshal(data, &services)
	return services, err
}

func saveServices(services map[string]string) error {
	data, err := json.MarshalIndent(services, "", " ")
	if err != nil {
		log.Fatalf("ERROR IN JSON PARSING :: %s", err.Error())
	}
	return os.WriteFile("services.json", data, 0644)
}

func run(command string, args ...string) error {
	cmd := exec.Command(command, args...)

	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Stdin = os.Stdin

	return cmd.Run()
}

func migrate(databaseURL string, args ...string) {
	if len(args) != 1 {
		log.Fatalln("migrate expects exactly one argument: up | down")
	}

	direction := args[0]
	if direction != "up" && direction != "down" {
		log.Fatalln("invalid migrate command, use: up or down")
	}

	err := run(
		"migrate",
		"-path", "./internals/database/migrations",
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
		log.Fatal("UNABLE TO LOAD services.json", err.Error())
	}

	for _, app := range apps {
		if _, exists := services[app]; exists {
			log.Printf("Service %s already exists", app)
			continue
		}

		finalPath := filepath.Join("./cmd/", app)
		if err := os.Mkdir(finalPath, 0755); err != nil {
			log.Fatalln("UNABLE TO CREATE SERVICE DIR ", err.Error())
		}
		fileLocation := filepath.Join(finalPath, "main.go")
		f, err := os.Create(fileLocation)
		if err != nil {
			log.Fatalln("UNABLE TO CREATE main.go IN THAT LOCATION", err.Error())
		}
		_, err = f.WriteString(template)
		if err != nil {
			f.Close()
			log.Fatalf("UNABLE TO UPDATE THE ./cli/main.go PLEASE UPDATE THAT MANUALLY :: %s", err.Error())
		}
		services[app] = "./" + finalPath
		f.Close()

	}

	saveServices(services)

}

func start(target ...string) {
	var wg sync.WaitGroup
	services, err := loadServices()
	if err != nil {
		log.Fatalln("UNABLE TO LOAD SERVICES FROM services.json FOR STARTING OUR WORKSPACE")
	}
	for _, name := range target {
		servicePath, ok := services[name]
		if !ok {
			log.Fatalln("UNKNOWN SERVICES PLEASE PROVIDE A VALID SERVICE NAME ")
		}
		wg.Add(1)
		go func(name, servicePath string) {
			log.Printf("STARTING APPLICATION... :: %s", name)
			err := run("go", "run", servicePath)
			defer wg.Done()
			if err != nil {
				log.Fatalf("UNABLE TO START THE APPILCAITON :: %s :: %s", name, err.Error())
			}
			log.Printf("APPLICATION FINISHED :: %s", name)

		}(name, servicePath)

	}
	wg.Wait()
}

func build(target ...string) {
	services, err := loadServices()
	var wg sync.WaitGroup
	if err != nil {
		log.Fatalln("UNABLE TO LOAD SERVICES FROM services.json FOR BUILDING OUR WORKSPACE")
	}

	for _, name := range target {
		pathForBuild, ok := services[name]
		if !ok {
			log.Fatalln("UNKNOWN SERVICES PLEASE PROVIDE A VALID SERVICE NAME ")
		}
		wg.Add(1)
		go func(name, pathForBuild string) {
			log.Printf("BUILDING APPLICATION... :: %s", name)
			err := run("go", "build", "-o", filepath.Join("bin", name), pathForBuild)
			if err != nil {
				log.Fatalf("UNABLE TO BUILD THIS SERVICE :: %s :: %s", name, err.Error())
			}
			defer wg.Done()
		}(name, pathForBuild)

	}

	wg.Wait()

}

func main() {
	slog.Info("---------CLI TOOL RUNNING FOR SERVING RIVON PROJECT--------")
	cfg := config.MustLoad()

	if len(os.Args) < 2 {
		log.Fatalf("usage: dev <start |build | add> [services...]")
	}
	command := Command(os.Args[1])
	args := os.Args[2:]
	services, err := loadServices()
	if err != nil {
		log.Fatalln("ERROR IN LOADING SERVICES FROM services.json ", err.Error())
	}

	if len(args) == 0 {
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
		migrate(cfg.Db.PgURL, args...)
	default:
		log.Fatalf("unknown command: %s", command)
	}
}
