GOARCH := $(GOARCH)
GOOS := $(GOOS)
FLAGS := -ldflags "-s -w"

all:
	GOARCH=$$GOARCH GOOS=$$GOOS go build $(FLAGS) ./cmd/meshnamed

clean:
	$(RM) meshnamed meshnamed.exe

.PHONY: all clean
