# tools

## Заметки

1. Сгенерировать сертификат:
    - openssl genrsa -out private.key 2048
    - openssl req -x509 -new -nodes -key private.key -sha256 -days 1024 -out rootCA.pem
2. CryptoPro.
    - При указании само-подписного серта, расшифровать данные не получится, т.к. отсылка будет на приватный ключ,
      который программа не видит.
    - Нужно за ранее сформировать контейнер и задать необходимые туда данные (cryptcp -creatcert -rdn "CN=x").
    - Если шифровалось с определенным КПС (критерий поиска сертификата) -dn x ("CN=Ivan Petrov") и если был задан
      пароль, то при расшифровки с таким же КПС нужно пароль снова указать.
    - Наверно лучше в итоге сделать по минимуму, чтоб программа cryptcp сама автоматически выбирала:
      cryptcp -creatcert -rdn "CN=x"; cryptcp -encr test.json test.msg; cryptcp -decr test.msg test2.json
