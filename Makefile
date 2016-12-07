all: prefer


prefer: cmds/prefer.go
	go build $<


.PHONY: all
