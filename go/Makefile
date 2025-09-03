.PHONY: init doc clean all

doc:
	@echo "make arkiv-format     Compile the executable file."
	@echo "make clean            Remove compiled executable file."
	@echo "make all              Remove and then compile the executable file."
	@echo "make init             Initialize Go dependencies"

arkiv-format:
	go build -o arkiv-format ./cmd/arkiv-format

clean:
	rm -f arkiv-format

init:
	go mod tidy

all: clean agent


