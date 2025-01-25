.PHONY: all res exe

all:
	python3 scripts/doBuild.py proxyMon all "-X cmd/proxyMon/main.Build=true"

res: 
	python3 scripts/doBuild.py proxyMon res

exe: 
	python3 scripts/doBuild.py proxyMon exe "-X cmd/proxyMon/main.Build=true"
