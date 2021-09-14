# START: begin
.PHONY: test
test:
#: START: begin
	go test -race ./... -v
# END: auth