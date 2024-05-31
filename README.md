# lab-02-temperatureSystemByCEP

## Pré-Requisitos

- [Composer](https://getcomposer.org);
- [Docker](https://www.docker.com);




## Para Testar

* Para testar, executar `docker compose up --build`

* Executar `curl -X POST -d '{"cep":"04942000"}' http://localhost:8080/cep -H "Content-Type: application/json"`

* Verificar os traces `http://localhost:9411/zipkin/`
