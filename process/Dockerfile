FROM python:3.10-slim-buster

ENV PYTHONUNBUFFERED=1

WORKDIR /process
COPY ./requirements.txt /process/requirements.txt
RUN pip install -r requirements.txt

COPY . /process


ENTRYPOINT ["python", "process.py"]
CMD ["/tmp/hls"]
