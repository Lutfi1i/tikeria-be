FROM python:3.11-alpine

WORKDIR /app/erp

COPY . /app/erp/

RUN mkdir -p /data
ENV ERP_DB_PATH=/data/frappe_erp.db
ENV PORT=8000
ENV FRAPPE_API_KEY=tikeria_admin_key
ENV FRAPPE_API_SECRET=tikeria_admin_secret

EXPOSE 8000

CMD ["python3", "server.py"]
