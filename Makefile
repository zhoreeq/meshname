GOARCH := $(GOARCH)
GOOS := $(GOOS)
FLAGS := -ldflags "-s -w"

all:
	GOARCH=$$GOARCH GOOS=$$GOOS go build $(FLAGS) ./cmd/meshnamed

clean:
	$(RM) meshnamed meshnamed.exe

test:
	go test pkg/meshname/*_test.go

.PHONY: all clean test
