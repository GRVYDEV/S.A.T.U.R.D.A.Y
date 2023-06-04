PYTHON_PKG := $(shell python-config --prefix)/lib/pkgconfig

freeze:
	@echo "freezing deps to requirements.txt..."
	@pip freeze > requirements.txt

install:
	@echo "installing deps..."
	@pip install -r requirements.txt

run-go:
	@PKG_CONFIG_PATH=$(PYTHON_PKG) go run main.go

run:
	@uvicorn main:app

dev:
	@uvicorn main:app --reload
