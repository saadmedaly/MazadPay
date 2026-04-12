.PHONY: mobile back web

mobile:
	cd MazadPay && flutter run -d chrome

back:
	cd backend && go run ./cmd/api/main.go

web:
	cd web && bun run dev

