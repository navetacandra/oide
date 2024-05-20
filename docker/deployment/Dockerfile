FROM alpine:latest
LABEL maintainer="navetacandra"

RUN apk update --no-cache && apk upgrade --available --no-cache
RUN apk add sudo nodejs npm sqlite php82 apache2 apache2-utils python3 curl php82-common php82-pdo php82-apache2 php82-dom php82-mbstring php82-tokenizer php82-xml php82-xmlwriter php82-fileinfo php82-session php82-openssl php82-sqlite3 php82-curl php82-zip php82-phar php82-iconv --no-cache

RUN mkdir /app
COPY entrypoint.sh /

RUN sed -i -e "s/DirectoryIndex index.html/DirectoryIndex index.php index.html/g" /etc/apache2/httpd.conf
RUN sed -i -e "s/#ServerName www.example.com:80/ServerName 0.0.0.0:80/g" /etc/apache2/httpd.conf
RUN mv /var/www/localhost/htdocs/index.html /var/www/localhost/htdocs/index.php

RUN curl https://getcomposer.org/download/latest-stable/composer.phar -o /usr/bin/composer 
RUN chmod +x /usr/bin/composer

RUN curl -sSL https://install.python-poetry.org | python3 -

ENTRYPOINT ["/entrypoint.sh"]


