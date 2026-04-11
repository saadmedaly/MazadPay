.PHONY: front back web

front:
	flutter run -d chrome

back:
	cd backend
	go run ./cmd/api/main.go

web:
	cd web
	bun run dev

