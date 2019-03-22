FROM centos

RUN yum install -y python-requests python-gunicorn python-flask python-netaddr

WORKDIR /home/app

COPY . ./

RUN groupadd -r -g 1001 app && adduser -r -g 1001 --uid 1001 app && \
    chown -R app:app /home/app

USER 1001

EXPOSE 8080

CMD ["gunicorn", "-b", "0.0.0.0:8080", "-w", "4", "app:app"]
