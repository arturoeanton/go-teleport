# go-teleport
Teleport


# Run in cloud
```
   AUTH_TOKEN="mi-token-secreto" go run .  -addr1 8081 -addr2 8082
```


# Run in local

```
    docker run --name some-nginx -d -p 8080:80 nginx
    AUTH_TOKEN="mi-token-secreto" go run . -addr1 cloud.com:8082 -addr2 127.0.0.1:8080 -client
```

# Ejemplo Postgres 

## Run in cloud
```
   AUTH_TOKEN="mi-token-secreto" go run .  -addr1 8080 -addr2 8082
```

## Run in local
```
    docker run -d --name pruebadb -p 5432:5432 -e POSTGRES_PASSWORD=qwerty123 postgres
    AUTH_TOKEN="mi-token-secreto" go run . -addr1 64.23.223.85:8082 -addr2 127.0.0.1:5432 -client
```

Luego te conectas a  postgres://postgres:qwerty123@64.23.223.85:8080/postgres

```
Diagrama


┌────────────┐        TCP:8080         ┌──-──────────────┐        TCP:8082         ┌──────────────┐       TCP:5432        ┌────────────┐
│ ClienteSQL │ <---------------------> │ Mirror / Proxy  │ <---------------------> │ ClienteLocal │ <-------------------> │ PostgreSQL │
│            │                         │ 64.23.223.85    │                         │    (Proxy)   │                       │  Localhost │
└────────────┘                         │ (puertos: 8080/ │                         │              │                       └────────────┘
                                       │          8082)  │                         └──────────────┘
                                       └─-───────────────┘

```


# Ejemplo FireBird 

## Run in cloud
```
   AUTH_TOKEN="mi-token-secreto" go run .  -addr1 8080 -addr2 8082
```

## Run in local
```
    docker run -d --name firebird25ss -p 30505:3050 -d -e ISC_PASSWORD=masterkey jacobalberty/firebird
    docker exec -it firebird25ss /bin/bash
    cd /usr/local/firebird/bin/
    ./isql -u SYSDBA -p masterkey /firebird/data/mi_base.fdb
    CREATE DATABASE '/firebird/data/mi_base.fdb' USER 'SYSDBA' PASSWORD 'masterkey';
    exit;
    quit;
    AUTH_TOKEN="mi-token-secreto" go run . -addr1 64.23.223.85:8082 -addr2 127.0.0.1:30505 -client
```

Luego te conectas a  jdbc:firebirdsql://64.23.223.85:8080//firebird/data/mi_base.fdb
USER 'SYSDBA' PASSWORD 'masterkey';


Diagrama


```
┌────────────┐        TCP:8080         ┌──-──────────────┐        TCP:8082         ┌──────────────┐       TCP:30505       ┌────────────-┐
│ ClienteSQL │ <---------------------> │ Mirror / Proxy  │ <---------------------> │ ClienteLocal │ <-------------------> │ FireBirdSQL │
│            │                         │ 64.23.223.85    │                         │    (Proxy)   │                       │  Localhost  │
└────────────┘                         │ (puertos: 8080/ │                         │              │                       └────────────-┘
                                       │          8082)  │                         └──────────────┘
                                       └─-───────────────┘
```




# Ejemplo FireBird  todo local
```
    docker run -d --name firebird25ss -p 30505:3050 -d -e ISC_PASSWORD=masterkey jacobalberty/firebird
    docker exec -it firebird25ss /bin/bash
    cd /usr/local/firebird/bin/
    ./isql -u SYSDBA -p masterkey /firebird/data/mi_base.fdb
    CREATE DATABASE '/firebird/data/mi_base.fdb' USER 'SYSDBA' PASSWORD 'masterkey';
    exit;
    quit;
    AUTH_TOKEN="mi-token-secreto" SHARED_KEY="thisis32byteslongthisis32byteslo" go run .  -addr1 8080 -addr2 8082
    AUTH_TOKEN="mi-token-secreto" SHARED_KEY="thisis32byteslongthisis32byteslo" go run . -addr1 127.0.0.1:8082 -addr2 127.0.0.1:30505 -client
```

