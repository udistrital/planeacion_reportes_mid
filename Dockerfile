FROM python:3
RUN pip3 install awscli --no-build-isolation
WORKDIR /
COPY entrypoint.sh entrypoint.sh
COPY main main
COPY conf/app.conf conf/app.conf
COPY static static
RUN chmod +x main entrypoint.sh
ENTRYPOINT ["/entrypoint.sh"]