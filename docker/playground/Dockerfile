FROM alpine:latest
LABEL maintainer="navetacandra"

RUN apk update --no-cache && apk upgrade --available --no-cache
RUN apk add sudo nodejs npm openssh sqlite php82 apache2 apache2-utils python3 curl php82-common php82-pdo php82-apache2 php82-dom php82-mbstring php82-tokenizer php82-xml php82-xmlwriter php82-fileinfo php82-session php82-openssl php82-sqlite3 php82-curl php82-zip php82-phar php82-iconv --no-cache

RUN mkdir /projects
ENV HOME /projects
RUN rm -r /root && ln -s /projects /root
RUN echo root:123456 | chpasswd

RUN mkdir -p /var/run/sshd
RUN echo 'PasswordAuthentication yes' >> /etc/ssh/sshd_config
RUN sed -i 's/#PermitRootLogin prohibit-password/PermitRootLogin yes/' /etc/ssh/sshd_config

COPY profile /.utils
COPY entrypoint.sh /

RUN echo "source /.utils" >> /projects/.profile
RUN echo "PATH=\"\$PATH:/projects/.local/bin\"" >> /projects/.profile
RUN source /projects/.profile

RUN sed -i -e "s/DirectoryIndex index.html/DirectoryIndex index.php index.html/g" /etc/apache2/httpd.conf
RUN sed -i -e "s/#ServerName www.example.com:80/ServerName 0.0.0.0:80/g" /etc/apache2/httpd.conf
RUN mv /var/www/localhost/htdocs/index.html /var/www/localhost/htdocs/index.php

RUN curl https://getcomposer.org/download/latest-stable/composer.phar -o /usr/bin/composer 
RUN chmod +x /usr/bin/composer

RUN curl -sSL https://install.python-poetry.org | python3 -

ENTRYPOINT ["/entrypoint.sh"]


