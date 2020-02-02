GOARCH := $(GOARCH)
GOOS := $(GOOS)
FLAGS := -ldflags "-s -w"

all:
	GOARCH=$$GOARCH GOOS=$$GOOS go build $(FLAGS) ./cmd/meshnamed
	GOARCH=$$GOARCH GOOS=$$GOOS go build $(FLAGS) ./cmd/meshname

clean:
	$(RM) meshnamed meshname meshnamed.exe meshname.exe

.PHONY: all clean
