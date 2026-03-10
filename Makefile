.PHONY: bridge ui dev-ui build

# Build the HTTP bridge
bridge:
	cd bridge && go build -o ../bin/bridge ./...

# Install UI dependencies
ui-install:
	cd ui && npm install

# Run the bridge (connects to slate at localhost:4242, listens on :8080)
run-bridge: bridge
	./bin/bridge

# Run the Vue dev server (proxies /api to localhost:8080)
dev-ui:
	cd ui && npm run dev

# Build the Vue UI for production
build-ui:
	cd ui && npm run build

# Build everything
build: bridge build-ui

clean:
	rm -rf bin/ ui/dist/
