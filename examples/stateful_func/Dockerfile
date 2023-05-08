#FROM grussorusso/serverledge-python310 
FROM python:3.10-alpine3.16
RUN pip3 install redis

COPY executor.py /
COPY function.py /
WORKDIR /

CMD python3 executor.py