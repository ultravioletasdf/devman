package main

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"os/exec"
	"slices"
	"strings"
	"sync"
	"time"

	"github.com/goccy/go-yaml"
	"github.com/joho/godotenv"
	"wait4x.dev/v3/checker"
	"wait4x.dev/v3/checker/http"
	"wait4x.dev/v3/checker/postgresql"
	"wait4x.dev/v3/checker/rabbitmq"
	"wait4x.dev/v3/checker/tcp"
	"wait4x.dev/v3/waiter"
)

func main() {
	cfg := parseConfig(readConfigFile())
	if cfg.EnvFile == "" {
		godotenv.Load(".env")
	} else if err := godotenv.Load(cfg.EnvFile); err != nil {
		fmt.Printf("Failed to load envfile: %v\n", err)
		os.Exit(1)
	}

	// Convert and sort cfg.Services into a slice so colours are consistent
	names := []string{}
	for k := range cfg.Services {
		names = append(names, k)
	}
	slices.Sort(names)

	var wg sync.WaitGroup
	var i int
	for _, name := range names {
		wg.Add(1)
		go startService(name, cfg.Services[name], &wg, i)
		i++
	}
	wg.Wait()
	fmt.Println("All processes finished")
}

func readConfigFile() []byte {
	path := "dev.yaml"
	file, err := os.Open(path)
	if err != nil {
		fmt.Printf("Failed to read %s: %v\n", path, err)
		os.Exit(1)
	}
	data, err := io.ReadAll(file)
	if err != nil {
		fmt.Printf("Failed to read %s: %v\n", path, err)
		os.Exit(1)
	}
	return data
}

type Service struct {
	Cmd     string
	WaitFor []string `yaml:"wait_for"`
}
type Config struct {
	EnvFile  string `yaml:"env_file"`
	Services map[string]Service
}

func parseConfig(data []byte) (cfg Config) {
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		fmt.Printf("Failed to parse config: %v\n", err)
		os.Exit(1)
	}
	return
}

func startService(name string, service Service, wg *sync.WaitGroup, i int) {
	checkers := []checker.Checker{}
	for _, address := range service.WaitFor {
		if strings.HasPrefix(address, "http://") || strings.HasPrefix(address, "https://") {
			checkers = append(checkers, http.New(address, http.WithTimeout(time.Minute*2)))
		} else if strings.HasPrefix(address, "amqp://") {
			checkers = append(checkers, rabbitmq.New(address, rabbitmq.WithTimeout(time.Minute*2)))
		} else if strings.HasPrefix(address, "postgres://") {
			checkers = append(checkers, postgresql.New(address))
		} else {
			checkers = append(checkers, tcp.New(address))
		}

	}
	if err := waiter.WaitParallel(checkers, waiter.WithTimeout(3*time.Minute)); err != nil {
		fmt.Printf("Dependency of %s didn't start: %v\n", name, err)
		os.Exit(1)
	}

	colour := colours[i%len(colours)]
	fmt.Printf("Starting %s%s\033[0m\n", colour, name)

	run(name, service.Cmd, colour)
	wg.Done()
}

func run(name, cmd string, colour string) {
	prefix := fmt.Sprintf("%s%s | \033[0m", colour, name)

	command := exec.Command("bash", "-c", cmd)

	stdout, err := command.StdoutPipe()
	if err != nil {
		panic(err)
	}
	stderr, err := command.StderrPipe()
	if err != nil {
		panic(err)
	}
	if err := command.Start(); err != nil {
		panic(err)
	}

	scanner := bufio.NewScanner(io.MultiReader(stdout, stderr))
	for scanner.Scan() {
		fmt.Println(prefix + scanner.Text())
	}
	if err := scanner.Err(); err != nil {
		panic(err)
	}

	command.Wait()
	fmt.Printf("%s%s\033[0m stopped running\n", colour, name)
}

var colours = []string{
	"\033[31m", // red
	"\033[32m", // green
	"\033[33m", // yellow
	"\033[34m", // blue
	"\033[35m", // magenta
	"\033[36m", // cyan
}
