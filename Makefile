all: check_docker_file nginx

check_docker_file:
	docker build -t yametech/checkdocker:v0.1.0 -f ./docker/Dockerfile.checkdocker .

render_nginx:
	docker build -t yametech/nginx-render:v0.0.1 -f docker/Dockerfile.nginx .