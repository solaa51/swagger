此处为简要示例，用来临时验证框架，详细案例请到 swagger-examples项目

docker pull mysql:latest

docker run -itd --name mysql-test -p 3306:3306 -e MYSQL_ROOT_PASSWORD=123456 mysql
