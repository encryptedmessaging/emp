build:
	script/build.sh
start:
	script/start.sh
stop:
	script/stop.sh
clean: stop
	rm -rf bin
	rm -rf pkg
	rm -rf ~/.config/emp/log/*

clobber: clean
	rm -rf src/github.com
	rm -rf src/code.google.com
