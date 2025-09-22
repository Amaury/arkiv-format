.PHONY: doc clean init dev dist

doc:
	@echo "$$(tput setab 4)$$(tput bold) Arkiv-format Makefile $$(tput sgr0)"
	@echo "make dev     $$(tput dim)Compile the executable file for the current platform.$$(tput sgr0)"
	@echo "make dist    $$(tput dim)Compile executable files for all supported platforms.$$(tput sgr0)"
	@echo "make clean   $$(tput dim)Remove compiled executable file.$$(tput sgr0)"
	@echo "make init    $$(tput dim)Initialize Go dependencies.$$(tput sgr0)"

clean:
	rm -f arkiv-format*

init:
	go mod tidy

dev:
	go build -o arkiv-format ./cmd/arkiv-format

dist:
	GOOS=aix GOARCH=ppc64 CGO_ENABLED=0 go build -ldflags="-s -w" -o arkiv-format_aix-ppc_64 ./cmd/arkiv-format
	
	GOOS=dragonfly GOARCH=amd64 CGO_ENABLED=0 go build -ldflags="-s -w" -o arkiv-format_dragonfly-x86_64 ./cmd/arkiv-format
	strip --strip-all arkiv-format_dragonfly-x86_64
	if command -v upx > /dev/null 2>&1; then upx -qq arkiv-format_dragonfly-x86_64; fi
	
	GOOS=freebsd GOARCH=arm CGO_ENABLED=0 go build -ldflags="-s -w" -o arkiv-format_freebsd-arm_32 ./cmd/arkiv-format
	GOOS=freebsd GOARCH=arm64 CGO_ENABLED=0 go build -ldflags="-s -w" -o arkiv-format_freebsd-arm_64 ./cmd/arkiv-format
	GOOS=freebsd GOARCH=riscv64 CGO_ENABLED=0 go build -ldflags="-s -w" -o arkiv-format_freebsd-riscv_64 ./cmd/arkiv-format
	GOOS=freebsd GOARCH=386 CGO_ENABLED=0 go build -ldflags="-s -w" -o arkiv-format_freebsd-x86_32 ./cmd/arkiv-format
	GOOS=freebsd GOARCH=amd64 CGO_ENABLED=0 go build -ldflags="-s -w" -o arkiv-format_freebsd-x86_64 ./cmd/arkiv-format
	
	GOOS=illumos GOARCH=amd64 CGO_ENABLED=0 go build -ldflags="-s -w" -o arkiv-format_illumos-x86_64 ./cmd/arkiv-format
	if command -v upx > /dev/null 2>&1; then upx -qq arkiv-format_illumos-x86_64; fi
	
	GOOS=linux GOARCH=arm CGO_ENABLED=0 go build -ldflags="-s -w" -o arkiv-format_linux-arm_32 ./cmd/arkiv-format
	if command -v upx > /dev/null 2>&1; then upx -qq arkiv-format_linux-arm_32; fi
	GOOS=linux GOARCH=arm64 CGO_ENABLED=0 go build -ldflags="-s -w" -o arkiv-format_linux-arm_64 ./cmd/arkiv-format
	if command -v upx > /dev/null 2>&1; then upx -qq arkiv-format_linux-arm_64; fi
	GOOS=linux GOARCH=loong64 CGO_ENABLED=0 go build -ldflags="-s -w" -o arkiv-format_linux-loong_64 ./cmd/arkiv-format
	GOOS=linux GOARCH=mips CGO_ENABLED=0 go build -ldflags="-s -w" -o arkiv-format_linux-mips_32 ./cmd/arkiv-format
	if command -v upx > /dev/null 2>&1; then upx -qq arkiv-format_linux-mips_32; fi
	GOOS=linux GOARCH=mips64 CGO_ENABLED=0 go build -ldflags="-s -w" -o arkiv-format_linux-mips_64 ./cmd/arkiv-format
	GOOS=linux GOARCH=mipsle CGO_ENABLED=0 go build -ldflags="-s -w" -o arkiv-format_linux-mips_32-le ./cmd/arkiv-format
	if command -v upx > /dev/null 2>&1; then upx -qq arkiv-format_linux-mips_32-le; fi
	GOOS=linux GOARCH=mips64le CGO_ENABLED=0 go build -ldflags="-s -w" -o arkiv-format_linux-mips_64-le ./cmd/arkiv-format
	GOOS=linux GOARCH=ppc64 CGO_ENABLED=0 go build -ldflags="-s -w" -o arkiv-format_linux-ppc_64 ./cmd/arkiv-format
	if command -v upx > /dev/null 2>&1; then upx -qq arkiv-format_linux-ppc_64; fi
	GOOS=linux GOARCH=ppc64le CGO_ENABLED=0 go build -ldflags="-s -w" -o arkiv-format_linux-ppc_64-le ./cmd/arkiv-format
	if command -v upx > /dev/null 2>&1; then upx -qq arkiv-format_linux-ppc_64-le; fi
	GOOS=linux GOARCH=riscv64 CGO_ENABLED=0 go build -ldflags="-s -w" -o arkiv-format_linux-riscv_64 ./cmd/arkiv-format
	GOOS=linux GOARCH=s390x CGO_ENABLED=0 go build -ldflags="-s -w" -o arkiv-format_linux-s390x ./cmd/arkiv-format
	GOOS=linux GOARCH=386 CGO_ENABLED=0 go build -ldflags="-s -w" -o arkiv-format_linux-x86_32 ./cmd/arkiv-format
	strip --strip-all arkiv-format_linux-x86_32
	if command -v upx > /dev/null 2>&1; then upx -qq arkiv-format_linux-x86_32; fi
	GOOS=linux GOARCH=amd64 CGO_ENABLED=0 go build -ldflags="-s -w" -o arkiv-format_linux-x86_64 ./cmd/arkiv-format
	strip --strip-all arkiv-format_linux-x86_64
	if command -v upx > /dev/null 2>&1; then upx -qq arkiv-format_linux-x86_64; fi
	
	GOOS=darwin GOARCH=arm64 CGO_ENABLED=0 go build -ldflags="-s -w" -o arkiv-format_macos-arm_64 ./cmd/arkiv-format
	if command -v upx > /dev/null 2>&1; then upx -qq arkiv-format_macos-arm_64; fi
	GOOS=darwin GOARCH=amd64 CGO_ENABLED=0 go build -ldflags="-s -w" -o arkiv-format_macos-x86_64 ./cmd/arkiv-format
	if command -v upx > /dev/null 2>&1; then upx -qq arkiv-format_macos-x86_64; fi
	
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
	strip --strip-all arkiv-format_solaris-x86_64
	if command -v upx > /dev/null 2>&1; then upx -qq arkiv-format_solaris-x86_64; fi
	
	GOOS=windows GOARCH=arm CGO_ENABLED=0 go build -ldflags="-s -w" -o arkiv-format_windows-arm_32.exe ./cmd/arkiv-format
	GOOS=windows GOARCH=arm64 CGO_ENABLED=0 go build -ldflags="-s -w" -o arkiv-format_windows-arm_64.exe ./cmd/arkiv-format
	GOOS=windows GOARCH=386 CGO_ENABLED=0 go build -ldflags="-s -w" -o arkiv-format_windows-x86_32.exe ./cmd/arkiv-format
	GOOS=windows GOARCH=amd64 CGO_ENABLED=0 go build -ldflags="-s -w" -o arkiv-format_windows-x86_64.exe ./cmd/arkiv-format

