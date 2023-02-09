.DEFAULT_GOAL := build

build:
	go build -ldflags="-s -w" -o go-airflow-simple-discord-monitor main.go

clean:
	rm -f go-airflow-simple-discord-monitor
