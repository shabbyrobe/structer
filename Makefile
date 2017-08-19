build:
	go build

cover:
	overalls -project=github.com/shabbyrobe/structer

coverhtml: cover
	go tool cover -html=profile.coverprofile

test:
	go test -v .
