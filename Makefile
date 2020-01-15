GOCMD=go

migration:
	$(GOCMD) run cmd/serial-vault-admin/main.go database

run-admin:
	@CSRF_SECURE=disable $(GOCMD) run cmd/serial-vault/main.go -config=settings.yaml -mode=admin

run-signing:
	@CSRF_SECURE=disable $(GOCMD) run cmd/serial-vault/main.go -config=settings.yaml -mode=signing

docker-run:
	@CSRF_SECURE=disable docker-compose -f docker-compose/docker-compose.yml up

docker-stop:
	docker-compose -f docker-compose/docker-compose.yml kill && docker-compose -f docker-compose/docker-compose.yml rm && docker rmi  docker-compose_web
