# generate wire code
wire:
	cd app/admin/service/cmd/service && wire

# generate proto code
buf:
	cd api && rm -rf ./gen && buf generate