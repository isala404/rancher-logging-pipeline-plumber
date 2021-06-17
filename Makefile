run:
	docker build -f server/Dockerfile.dev --tag rancher-logging-explorer server/
	docker run --rm -it -p 8000:8000 -v `pwd`/server:/go/src/app -v `pwd`/ui/build:/go/src/app/build rancher-logging-explorer