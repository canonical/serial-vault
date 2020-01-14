GOCMD=go

migration:
	$(GOCMD) run cmd/serial-vault-admin/main.go database

run-admin:
	@CSRF_SECURE=disable $(GOCMD) run cmd/serial-vault/main.go -config=settings.yaml -mode=admin

run-signing:
	@CSRF_SECURE=disable $(GOCMD) run cmd/serial-vault/main.go -config=settings.yaml -mode=signing

docker:
	docker-compose -f docker-compose/docker-compose.yml up
