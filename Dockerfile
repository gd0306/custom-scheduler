FROM gd306.cn/base/centos7
ENV TZ Asia/Shanghai
COPY ./custom-scheduler /bin/custom-scheduler
CMD ["/bin/custom-scheduler"]
