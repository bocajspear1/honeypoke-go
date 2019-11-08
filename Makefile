staticstrip_flags = -ldflags="-s -w"

all:
	go build $(staticstrip_flags) -o honeypoke cmd/honeypokego/main.go 