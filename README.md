# How to run & test

## Solution using DB

### Logic
- create table to record user api call 
- counting total api call, when reach to limit return error

### How to test

- create file config.yaml, .env
- copy value from ```config.example.yaml, .env.example```
- run ```make run-with-db``` to run solution with DB  
- open ```localhost:8080/api?user=user_a``` for testing by reload many time continuously

## Solution using redis

### Logic
- using redis for caching user api call & increase them, this value will invalid after 
a duration (config sample is 5 second)

### How to test

- create file config.yaml, .env
- copy value from ```config.example.yaml, .env.example```
- run ```make run-with-redis``` to run solution with DB
- open ```localhost:8080/api?user=user_a``` for testing by reload many time continuously
