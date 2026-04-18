.PHONY: mobile back web build run

mobile:
	cd MazadPay && flutter run -d chrome

back:
	cd backend && go run ./cmd/server/main.go

web:
	cd web && bun run dev

build:
	cd backend && go build -o /bin/server.exe ./cmd/server/main.go

run:
	cd backend && ./bin/server.exe
