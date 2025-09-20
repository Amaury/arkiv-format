.PHONY: init doc clean all

doc:
	@echo "make arkiv-format     Compile the executable file."
	@echo "make clean            Remove compiled executable file."
	@echo "make all              Remove and then compile the executable file."
	@echo "make init             Initialize Go dependencies"

arkiv-format:
	go build -o arkiv-format ./cmd/arkiv-format

dist:
	GOOS=aix GOARCH=ppc64 CGO_ENABLED=0 go build -ldflags="-s -w" -o arkiv-format_aix-ppc_64 ./cmd/arkiv-format
	GOOS=dragonfly GOARCH=amd64 CGO_ENABLED=0 go build -ldflags="-s -w" -o arkiv-format_dragonfly-x86_64 ./cmd/arkiv-format
	GOOS=freebsd GOARCH=arm CGO_ENABLED=0 go build -ldflags="-s -w" -o arkiv-format_freebsd-arm_32 ./cmd/arkiv-format
	GOOS=freebsd GOARCH=arm64 CGO_ENABLED=0 go build -ldflags="-s -w" -o arkiv-format_freebsd-arm_64 ./cmd/arkiv-format
	GOOS=freebsd GOARCH=riscv64 CGO_ENABLED=0 go build -ldflags="-s -w" -o arkiv-format_freebsd-riscv_64 ./cmd/arkiv-format
	GOOS=freebsd GOARCH=386 CGO_ENABLED=0 go build -ldflags="-s -w" -o arkiv-format_freebsd-x86_32 ./cmd/arkiv-format
	GOOS=freebsd GOARCH=amd64 CGO_ENABLED=0 go build -ldflags="-s -w" -o arkiv-format_freebsd-x86_64 ./cmd/arkiv-format
	GOOS=illumos GOARCH=amd64 CGO_ENABLED=0 go build -ldflags="-s -w" -o arkiv-format_illumos-x86_64 ./cmd/arkiv-format
	GOOS=linux GOARCH=arm CGO_ENABLED=0 go build -ldflags="-s -w" -o arkiv-format_linux-arm_32 ./cmd/arkiv-format
	GOOS=linux GOARCH=arm64 CGO_ENABLED=0 go build -ldflags="-s -w" -o arkiv-format_linux-arm_64 ./cmd/arkiv-format
	GOOS=linux GOARCH=loong64 CGO_ENABLED=0 go build -ldflags="-s -w" -o arkiv-format_linux-loong_64 ./cmd/arkiv-format
	GOOS=linux GOARCH=mips CGO_ENABLED=0 go build -ldflags="-s -w" -o arkiv-format_linux-mips_32 ./cmd/arkiv-format
	GOOS=linux GOARCH=mips64 CGO_ENABLED=0 go build -ldflags="-s -w" -o arkiv-format_linux-mips_64 ./cmd/arkiv-format
	GOOS=linux GOARCH=mipsle CGO_ENABLED=0 go build -ldflags="-s -w" -o arkiv-format_linux-mips_32-le ./cmd/arkiv-format
	GOOS=linux GOARCH=mips64le CGO_ENABLED=0 go build -ldflags="-s -w" -o arkiv-format_linux-mips_64-le ./cmd/arkiv-format
	GOOS=linux GOARCH=ppc64 CGO_ENABLED=0 go build -ldflags="-s -w" -o arkiv-format_linux-ppc_64 ./cmd/arkiv-format
	GOOS=linux GOARCH=ppc64le CGO_ENABLED=0 go build -ldflags="-s -w" -o arkiv-format_linux-ppc_64-le ./cmd/arkiv-format
	GOOS=linux GOARCH=riscv64 CGO_ENABLED=0 go build -ldflags="-s -w" -o arkiv-format_linux-riscv_64 ./cmd/arkiv-format
	GOOS=linux GOARCH=s390x CGO_ENABLED=0 go build -ldflags="-s -w" -o arkiv-format_linux-s390x ./cmd/arkiv-format
	GOOS=linux GOARCH=386 CGO_ENABLED=0 go build -ldflags="-s -w" -o arkiv-format_linux-x86_32 ./cmd/arkiv-format
	GOOS=linux GOARCH=amd64 CGO_ENABLED=0 go build -ldflags="-s -w" -o arkiv-format_linux-x86_64 ./cmd/arkiv-format
	GOOS=darwin GOARCH=arm64 CGO_ENABLED=0 go build -ldflags="-s -w" -o arkiv-format_macos-arm_64 ./cmd/arkiv-format
	GOOS=darwin GOARCH=amd64 CGO_ENABLED=0 go build -ldflags="-s -w" -o arkiv-format_macos-x86_64 ./cmd/arkiv-format
	GOOS=netbsd GOARCH=arm CGO_ENABLED=0 go build -ldflags="-s -w" -o arkiv-format_netbsd-arm_32 ./cmd/arkiv-format
	GOOS=netbsd GOARCH=arm64 CGO_ENABLED=0 go build -ldflags="-s -w" -o arkiv-format_netbsd-arm_64 ./cmd/arkiv-format
	GOOS=netbsd GOARCH=386 CGO_ENABLED=0 go build -ldflags="-s -w" -o arkiv-format_netbsd-x86_32 ./cmd/arkiv-format
	GOOS=netbsd GOARCH=amd64 CGO_ENABLED=0 go build -ldflags="-s -w" -o arkiv-format_netbsd-x86_64 ./cmd/arkiv-format
	GOOS=openbsd GOARCH=arm CGO_ENABLED=0 go build -ldflags="-s -w" -o arkiv-format_openbsd-arm_32 ./cmd/arkiv-format
	GOOS=openbsd GOARCH=arm64 CGO_ENABLED=0 go build -ldflags="-s -w" -o arkiv-format_openbsd-arm_64 ./cmd/arkiv-format
	GOOS=openbsd GOARCH=ppc64 CGO_ENABLED=0 go build -ldflags="-s -w" -o arkiv-format_openbsd-ppc_64 ./cmd/arkiv-format
	GOOS=openbsd GOARCH=riscv64 CGO_ENABLED=0 go build -ldflags="-s -w" -o arkiv-format_openbsd-riscv_64 ./cmd/arkiv-format
	GOOS=openbsd GOARCH=386 CGO_ENABLED=0 go build -ldflags="-s -w" -o arkiv-format_openbsd-x86_32 ./cmd/arkiv-format
	GOOS=openbsd GOARCH=amd64 CGO_ENABLED=0 go build -ldflags="-s -w" -o arkiv-format_openbsd-x86_64 ./cmd/arkiv-format
	GOOS=plan9 GOARCH=arm CGO_ENABLED=0 go build -ldflags="-s -w" -o arkiv-format_plan9-arm_32 ./cmd/arkiv-format
	GOOS=plan9 GOARCH=386 CGO_ENABLED=0 go build -ldflags="-s -w" -o arkiv-format_plan9-x86_32 ./cmd/arkiv-format
	GOOS=plan9 GOARCH=amd64 CGO_ENABLED=0 go build -ldflags="-s -w" -o arkiv-format_plan9-x86_64 ./cmd/arkiv-format
	GOOS=solaris GOARCH=amd64 CGO_ENABLED=0 go build -ldflags="-s -w" -o arkiv-format_solaris-x86_64 ./cmd/arkiv-format
	GOOS=windows GOARCH=arm CGO_ENABLED=0 go build -ldflags="-s -w" -o arkiv-format_windows-arm_32.exe ./cmd/arkiv-format
	GOOS=windows GOARCH=arm64 CGO_ENABLED=0 go build -ldflags="-s -w" -o arkiv-format_windows-arm_64.exe ./cmd/arkiv-format
	GOOS=windows GOARCH=386 CGO_ENABLED=0 go build -ldflags="-s -w" -o arkiv-format_windows-x86_32.exe ./cmd/arkiv-format
	GOOS=windows GOARCH=amd64 CGO_ENABLED=0 go build -ldflags="-s -w" -o arkiv-format_windows-x86_64.exe ./cmd/arkiv-format


clean:
	rm -f arkiv-format*

init:
	go mod tidy

all: clean arkiv-format

