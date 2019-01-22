default:
	@echo "Options include test, test-clean"
test:
	mkdir -p tests/keys
	ssh-keygen -t rsa -b 4096 -f tests/keys/ssh-key -N pass123 <<< 'y'
	docker build -f tests/Dockerfile.build.test -t sample-project-test .
	docker-compose -f tests/docker-compose-tests.yml up --abort-on-container-exit --remove-orphans
	docker-compose -f tests/docker-compose-tests.yml down --rmi local --remove-orphans
	rm -rf tests/keys
	docker system prune -f
test-clean:
	docker system prune -f