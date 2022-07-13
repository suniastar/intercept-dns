VERSION=0.1.0

bin:
	go build

dep:
	go get

clean:
	rm -rf dist

gox:
	CGO_ENABLED=0 gox -ldflags="-s -w" -osarch '!darwin/386' -output="dist/{{.Dir}}_{{.OS}}_{{.Arch}}"

draft:
	ghr -draft v$(VERSION) dist/