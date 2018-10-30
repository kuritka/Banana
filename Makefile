update:
	dep ensure --update
	rm -rf vendor

build:
	docker build . -t edtcontainerregistry.azurecr.io/edt-sftp --build-arg GH_USER=$$GH_USER --build-arg GH_TOKEN=$$GH_TOKEN

push:
	docker push edtcontainerregistry.azurecr.io/edt-sftp