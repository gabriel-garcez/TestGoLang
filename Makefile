build:
	@rm -rf release
	@mkdir release
	@go build -o release/bm

install: build
	@sudo rm -rf /opt/bookish-memory
	@sudo rm -rf /usr/bin/bookish-memory
	@sudo mkdir /opt/bookish-memory
	@sudo mv release/bm /opt/bookish-memory
	@rm -rf release
	@sudo chmod +x /opt/bookish-memory/bm
	@sudo ln -s /opt/bookish-memory/bm /usr/bin/bookish-memory
	@bookish-memory -h
