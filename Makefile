GOARCH := $(GOARCH)
GOOS := $(GOOS)
FLAGS := -ldflags "-s -w"

all:
	GOARCH=$$GOARCH GOOS=$$GOOS go build $(FLAGS) meshnamed.go
	GOARCH=$$GOARCH GOOS=$$GOOS go build $(FLAGS) meshname.go

clean:
	$(RM) meshnamed meshname meshnamed.exe meshname.exe

.PHONY: all clean
