.PHONY: *

run:
	-docker compose -f ./docker-compose.yml -p kraicklist down --remove-orphans
	docker compose -f ./docker-compose.yml -p kraicklist up --build -d