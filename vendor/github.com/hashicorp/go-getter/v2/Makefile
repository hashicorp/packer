start-smb:
	@docker-compose build
	docker-compose up -d samba

smbtests-prepare:
	@docker-compose build
	@docker-compose up -d
	@sleep 60
	@docker exec -it samba bash -c "echo 'Hello' > /mnt/file.txt && mkdir -p /mnt/subdir  && echo 'Hello' > /mnt/subdir/file.txt"

smbtests:
	@docker cp ./ gogetter:/go-getter/
	@docker exec -it gogetter bash -c "env ACC_SMB_TEST=1 go test -v ./... -run=TestSmb_"