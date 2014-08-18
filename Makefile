build:
	script/build.sh

clean:
	rm -rf emp
	rm -rf pkg

clobber: clean
	rm -rf src/github.com
	rm -rf src/code.google.com
