# install cli
installCli:
	# protoc tools
	@go install github.com/go-kratos/kratos/cmd/protoc-gen-go-http/v2@latest
	@go install github.com/go-kratos/kratos/cmd/protoc-gen-go-errors/v2@latest
	@go install github.com/google/gnostic/cmd/protoc-gen-openapi@latest
	# cli tools
	@go get entgo.io/ent/cmd/ent
	@go install github.com/go-kratos/kratos/cmd/kratos/v2@latest
	@go install github.com/google/wire/cmd/wire

# generate wire code
wire:
	@cd app/admin/service/cmd/service && wire

# generate proto code
buf:
	@cd api && rm -rf ./gen && buf dep update && buf generate

# generate ent code
ent:
	@echo "🛠️  Generating Ent code..."
	@cd app/admin/service/internal/data && \
    	go generate ./ent
	@echo "✅ Ent code generated successfully!"
