test:
	go test ./... -v --cover

load-test:
	bash tests/load/run.sh

load-test-clean:
	cd tests/load && docker compose down --volumes --remove-orphans --rmi local

