FROM centos

RUN yum install -y python-requests python-gunicorn python-flask

COPY . ./

CMD ["gunicorn", "-b", "0.0.0.0:8080", "-w", "4", "app:app"]
