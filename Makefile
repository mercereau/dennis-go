.PHONY: build build-backend build-frontend run-backend clean

build: build-frontend build-backend

build-backend:
	go build -o dns-filter .

build-frontend:
	cd frontend && npm install && npm run build

run-backend: build-backend
	./dns-filter -db dns.db
