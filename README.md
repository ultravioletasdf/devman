# Devman 

`v0.0.1`

A simple command tool to run multiple services, allowing you to define services to wait for before to starting.

It also allows you to load a .env file to the environment

## Config

Devman requires a config file to run. This must be called dev.yaml in the CWD.

```yaml
env_file = ".env" # This file is loaded by defeault

services:
    name:
        cmd: "echo 'hi'"
    some_server:
        cmd: "go run ." # imaginary server that runs on port 3000
    some_other_server: 
        cmd: "cd otherthing && go run ." # some other imaginary server that waits for some_server to start
        wait_for:
            - "localhost:3000" # wait for a tcp connection
			- "amqp://dev:dev@127.0.0.1:5672" # or a rabbitmq connection
			- "postgres://postgres:dev@localhost:5432/?sslmode=disable" # or postgres
			- "http://localhost:3001" # or a http connection
	some_reload:
		cmd: "air" # You can use air to rebuild your app whenever there is a change
```

